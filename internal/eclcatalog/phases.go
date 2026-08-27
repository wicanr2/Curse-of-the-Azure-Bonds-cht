package eclcatalog

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
)

// PhaseFormatVersion versions the ordered-effect phase ledger independently of
// the corpus catalog: the corpus is regenerated from the archive, the ledger is
// hand-curated from disassembly and only moves when new bytes are read.
const PhaseFormatVersion = 1

// Commit phases. The first four are the vocabulary RE-01 asked for; the last
// three are additions the DOS bytes forced. See spec 1104 §五 for why
// `resume_only` has no member: every effect that waits for the end of a run
// lives in the lifecycle driver (overlay-02:3691h), not in an opcode handler.
const (
	// PhaseImmediate commits inside the handler, before it returns. The next
	// opcode already sees the result.
	PhaseImmediate = "immediate"
	// PhasePauseBeforeCommit blocks inside the handler for player input or a
	// nested subsystem, then commits. The ECL PC is already past the
	// instruction while it blocks.
	PhasePauseBeforeCommit = "pause_before_commit"
	// PhaseDeferred only records state. The visible effect is committed by a
	// different instruction or by the driver.
	PhaseDeferred = "deferred"
	// PhaseResumeOnly would commit only after the interpreter run ends.
	PhaseResumeOnly = "resume_only"
	// PhaseCommitPoint flushes state that earlier instructions deferred.
	PhaseCommitPoint = "commit_point"
	// PhaseTerminal ends the current interpreter run.
	PhaseTerminal = "terminal"
	// PhaseControlFlow only moves the ECL PC.
	PhaseControlFlow = "control_flow"
	// PhaseUnknown means the handler has not been read yet. It is the default
	// so that an unlisted opcode fails the coverage gate instead of silently
	// inheriting a neighbour's phase.
	PhaseUnknown = "unknown"
)

// How the handler moves the ECL program counter before running its effect.
const (
	// AdvanceDecodeOperands calls overlay-07 entry#2, which consumes the
	// opcode byte plus its operands.
	AdvanceDecodeOperands = "decode_operands"
	// AdvanceInc is the single-byte `inc word ptr ds:4FB4h`.
	AdvanceInc = "inc"
	// AdvanceAssign writes a branch destination into the PC.
	AdvanceAssign = "assign"
)

// How the remake models the boundary. Kept separate from Phase so that a
// remake-side resumable transaction is never mistaken for evidence that the
// original paused there.
const (
	BoundaryInline    = "inline"
	BoundaryResumable = "resumable"
)

// OpcodePhase is one row of the ordered-effect record. Every field is either an
// observation with an address behind it or an explicit `unknown`.
type OpcodePhase struct {
	Opcode     string   `json:"opcode"`
	Name       string   `json:"name"`
	DOSHandler string   `json:"dos_handler"`
	Advance    string   `json:"advance"`
	Phase      string   `json:"phase"`
	Boundary   string   `json:"remake_boundary"`
	Confidence string   `json:"confidence"`
	SpecRefs   []string `json:"spec_refs,omitempty"`
	Note       string   `json:"note"`
}

// PhaseLedger is the serialized ordered-effect record.
type PhaseLedger struct {
	FormatVersion int           `json:"format_version"`
	Generator     string        `json:"generator"`
	Vocabulary    []string      `json:"phase_vocabulary"`
	Limitations   []string      `json:"limitations"`
	Rows          []OpcodePhase `json:"rows"`
}

// phaseRow is the curated table's storage shape. It deliberately has no note
// field: the per-opcode Traditional Chinese evidence summary lives in the tool
// text catalog under `ecl_phases.note.<opcode>`, the same way the sound-BIOS
// argument table does. A missing ID panics, so a new row cannot ship without
// its evidence summary.
type phaseRow struct {
	Opcode     string
	Name       string
	DOSHandler string
	Advance    string
	Phase      string
	Boundary   string
	Confidence string
	SpecRefs   []string
}

// opcodePhases is the curated table. Rows marked `unknown` are the remaining
// work: they are listed on purpose so "how many handlers are still unread" is
// answerable without re-deriving it.
var opcodePhases = []phaseRow{
	{"0x00", "EXIT", "0052h", AdvanceInc, PhaseTerminal, BoundaryInline, "exact",
		[]string{"1104"}},
	{"0x01", "GOTO", "00E8h", AdvanceAssign, PhaseControlFlow, BoundaryInline, "exact",
		[]string{"1104"}},
	{"0x02", "GOSUB", "0107h", AdvanceAssign, PhaseControlFlow, BoundaryInline, "exact",
		[]string{"1104"}},
	{"0x03", "COMPARE", "011Eh", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x04", "ADD", "019Bh", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x05", "SUBTRAT", "019Bh", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x06", "DIVIDE", "019Bh", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x07", "MULTIPLY", "019Bh", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x08", "RANDOM", "0244h", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x09", "SAVE", "02A2h", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x0A", "LOAD CHARACTER", "02F0h", AdvanceDecodeOperands, PhaseImmediate, BoundaryInline, "exact",
		[]string{"1104", "1098"}},
	{"0x0B", "LOAD MONSTER", "0466h", AdvanceDecodeOperands, PhaseImmediate, BoundaryInline, "exact",
		[]string{"1104"}},
	{"0x0C", "SETUP MONSTER", "03CAh", AdvanceDecodeOperands, PhaseImmediate, BoundaryInline, "exact",
		[]string{"1104"}},
	{"0x0E", "PICTURE", "0841h", AdvanceDecodeOperands, PhaseDeferred, BoundaryInline, "exact",
		[]string{"1104"}},
	{"0x0D", "APPROACH", "0801h", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x10", "INPUT STRING", "0972h", AdvanceDecodeOperands, PhaseUnknown, BoundaryResumable, "unknown",
		nil},
	{"0x15", "VERTICAL MENU", "0EBDh", AdvanceDecodeOperands, PhaseUnknown, BoundaryResumable, "unknown",
		nil},
	{"0x28", "ROB", "1F46h", AdvanceDecodeOperands, PhaseUnknown, BoundaryResumable, "unknown",
		nil},
	{"0x2E", "DAMAGE", "2942h", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x39", "WHO", "2D5Eh", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x3C", "PROTECTION", "321Fh", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x3E", "DUMP", "3251h", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x3F", "FIND SPECIAL", "3284h", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x11", "PRINT", "09EAh", AdvanceDecodeOperands, PhaseImmediate, BoundaryInline, "exact",
		[]string{"1104"}},
	{"0x12", "PRINTCLEAR", "09EAh", AdvanceDecodeOperands, PhaseImmediate, BoundaryInline, "exact",
		[]string{"1104"}},
	{"0x13", "RETURN", "0A86h", AdvanceAssign, PhaseControlFlow, BoundaryInline, "exact",
		[]string{"1104"}},
	{"0x14", "COMPARE AND", "0AD6h", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x16", "IF =", "0B3Bh", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x17", "IF <>", "0B3Bh", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x18", "IF <", "0B3Bh", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x19", "IF >", "0B3Bh", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x1A", "IF <=", "0B3Bh", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x1B", "IF >=", "0B3Bh", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x1C", "CLEARMONSTERS", "120Eh", AdvanceInc, PhaseImmediate, BoundaryInline, "exact",
		[]string{"1104"}},
	{"0x1D", "PARTYSTRENGTH", "1271h", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x1E", "CHECKPARTY", "1416h", AdvanceDecodeOperands, PhaseImmediate, BoundaryInline, "exact",
		[]string{"1087"}},
	{"0x20", "NEWECL", "0BBBh", AdvanceDecodeOperands, PhaseTerminal, BoundaryResumable, "exact",
		[]string{"1104"}},
	{"0x21", "LOAD FILES", "0C15h", AdvanceDecodeOperands, PhaseDeferred, BoundaryInline, "exact",
		[]string{"1104"}},
	{"0x24", "COMBAT", "179Ah", AdvanceInc, PhasePauseBeforeCommit, BoundaryResumable, "exact",
		[]string{"1095"}},
	{"0x25", "ON GOTO", "1A9Bh", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x26", "ON GOSUB", "1A9Bh", AdvanceAssign, PhaseUnknown, BoundaryInline, "unknown",
		[]string{"560", "564"}},
	{"0x27", "TREASURE", "1B53h", AdvanceDecodeOperands, PhasePauseBeforeCommit, BoundaryResumable, "exact",
		[]string{"255", "257", "258", "558"}},
	{"0x29", "ENCOUNTER MENU", "2086h", AdvanceDecodeOperands, PhaseUnknown, BoundaryResumable, "unknown",
		[]string{"1083"}},
	{"0x2A", "GETTABLE", "0E13h", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x2B", "HORIZONTAL MENU", "1082h", AdvanceDecodeOperands, PhasePauseBeforeCommit, BoundaryResumable, "exact",
		[]string{"1104"}},
	{"0x2C", "PARLAY", "27A8h", AdvanceDecodeOperands, PhaseUnknown, BoundaryResumable, "unknown",
		[]string{"560", "1083"}},
	{"0x2D", "CALL", "2F02h", AdvanceDecodeOperands, PhaseCommitPoint, BoundaryInline, "exact",
		[]string{"1104"}},
	{"0x2F", "AND", "0DA4h", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x30", "OR", "0DA4h", AdvanceDecodeOperands, PhaseImmediate, BoundaryInline, "exact",
		[]string{"1157"}},
	{"0x31", "SPRITE OFF", "2C8Fh", AdvanceInc, PhaseImmediate, BoundaryInline, "exact",
		[]string{"1104"}},
	{"0x32", "FIND ITEM", "2847h", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x33", "PRINT RETURN", "2CEAh", AdvanceInc, PhaseImmediate, BoundaryInline, "exact",
		[]string{"1104"}},
	{"0x34", "ECL CLOCK", "2CB5h", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		[]string{"241", "560", "1106"}},
	{"0x35", "SAVE TABLE", "0E71h", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		nil},
	{"0x36", "ADD NPC", "2DA9h", AdvanceDecodeOperands, PhaseImmediate, BoundaryInline, "exact",
		[]string{"1104"}},
	{"0x37", "LOAD PIECES", "0C15h", AdvanceDecodeOperands, PhaseDeferred, BoundaryInline, "exact",
		[]string{"1104"}},
	{"0x38", "PROGRAM", "30DDh", AdvanceDecodeOperands, PhaseTerminal, BoundaryResumable, "exact",
		[]string{"1104", "1087"}},
	{"0x3A", "DELAY", "28F3h", AdvanceInc, PhaseImmediate, BoundaryInline, "exact",
		[]string{"1104"}},
	{"0x3B", "SPELL", "2E16h", AdvanceDecodeOperands, PhaseUnknown, BoundaryInline, "unknown",
		[]string{"560", "1083"}},
	{"0x3D", "CLEAR BOX", "2D15h", AdvanceInc, PhaseImmediate, BoundaryInline, "exact",
		[]string{"1104"}},
	{"0x40", "DESTROY ITEMS", "32D8h", AdvanceDecodeOperands, PhaseImmediate, BoundaryInline, "exact",
		[]string{"1104"}},
}

// VerifyPhaseCoverage fails closed: every opcode the corpus actually contains
// must have a curated row, and no row may describe an opcode the corpus does
// not contain. A missing row would otherwise let a newly discovered opcode ship
// with no recorded commit phase; a stale row would let a removed opcode keep
// claiming evidence.
func VerifyPhaseCoverage(catalog Catalog, ledger PhaseLedger) error {
	rows := make(map[string]bool, len(ledger.Rows))
	for _, row := range ledger.Rows {
		if rows[row.Opcode] {
			return fmt.Errorf("ordered-effect phase ledger lists opcode %s twice", row.Opcode)
		}
		rows[row.Opcode] = true
	}
	missing := make([]string, 0)
	for opcode := range catalog.Summary.OpcodeCounts {
		if !rows[opcode] {
			missing = append(missing, opcode)
		}
	}
	stale := make([]string, 0)
	for opcode := range rows {
		if _, ok := catalog.Summary.OpcodeCounts[opcode]; !ok {
			stale = append(stale, opcode)
		}
	}
	sort.Strings(missing)
	sort.Strings(stale)
	if len(missing) > 0 {
		return fmt.Errorf("ordered-effect phase ledger is missing corpus opcodes %s",
			strings.Join(missing, ", "))
	}
	if len(stale) > 0 {
		return fmt.Errorf("ordered-effect phase ledger describes opcodes absent from the corpus %s",
			strings.Join(stale, ", "))
	}
	return nil
}

// PhaseRows returns the curated ordered-effect record with each row's evidence
// summary resolved from the tool text catalog.
func PhaseRows() []OpcodePhase {
	rows := make([]OpcodePhase, 0, len(opcodePhases))
	for _, row := range opcodePhases {
		rows = append(rows, OpcodePhase{
			Opcode:     row.Opcode,
			Name:       row.Name,
			DOSHandler: row.DOSHandler,
			Advance:    row.Advance,
			Phase:      row.Phase,
			Boundary:   row.Boundary,
			Confidence: row.Confidence,
			SpecRefs:   append([]string(nil), row.SpecRefs...),
			Note:       tooltext.Text("ecl_phases.note." + strings.TrimPrefix(row.Opcode, "0x")),
		})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Opcode < rows[j].Opcode })
	return rows
}

// PhaseFor reports the curated row for one opcode.
func PhaseFor(opcode string) (OpcodePhase, bool) {
	for _, row := range PhaseRows() {
		if row.Opcode == opcode {
			return row, true
		}
	}
	return OpcodePhase{}, false
}

// BuildPhaseLedger assembles the serializable ledger.
func BuildPhaseLedger() PhaseLedger {
	return PhaseLedger{
		FormatVersion: PhaseFormatVersion,
		Generator:     "cmd/ecl-event-catalog",
		Vocabulary: []string{
			PhaseImmediate, PhasePauseBeforeCommit, PhaseDeferred, PhaseResumeOnly,
			PhaseCommitPoint, PhaseTerminal, PhaseControlFlow, PhaseUnknown,
		},
		Limitations: []string{
			tooltext.Text("ecl_phases.limitation_dos_only"),
			tooltext.Text("ecl_phases.limitation_unknown"),
			tooltext.Text("ecl_phases.limitation_boundary"),
		},
		Rows: PhaseRows(),
	}
}

// EncodePhaseJSON serializes the ledger deterministically.
func EncodePhaseJSON(ledger PhaseLedger) ([]byte, error) {
	data, err := json.MarshalIndent(ledger, "", " ")
	if err != nil {
		return nil, fmt.Errorf("encode ordered-effect phase ledger: %w", err)
	}
	return append(data, '\n'), nil
}

// EncodePhaseMarkdown renders the human-reviewable companion.
func EncodePhaseMarkdown(ledger PhaseLedger) []byte {
	var output strings.Builder
	fmt.Fprintln(&output, tooltext.Text("ecl_phases.title"))
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, tooltext.Text("ecl_phases.generated_notice"))
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, tooltext.Text("ecl_phases.rule_heading"))
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, tooltext.Text("ecl_phases.rule_body"))
	fmt.Fprintln(&output)
	for _, limitation := range ledger.Limitations {
		fmt.Fprintf(&output, "- %s\n", limitation)
	}
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, tooltext.Text("ecl_phases.table_heading"))
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, tooltext.Text("ecl_phases.table_header"))
	fmt.Fprintln(&output, "|---|---|---|---|---|---|---|---|")
	for _, row := range ledger.Rows {
		refs := make([]string, 0, len(row.SpecRefs))
		for _, ref := range row.SpecRefs {
			refs = append(refs, fmt.Sprintf("`%s`", ref))
		}
		reference := strings.Join(refs, tooltext.Text("ecl_catalog.reference_separator"))
		if reference == "" {
			reference = "—"
		}
		fmt.Fprintf(&output, "| `%s` | %s | `%s` | `%s` | `%s` | `%s` | %s | %s |\n",
			row.Opcode, row.Name, row.DOSHandler, row.Advance, row.Phase,
			row.Confidence, reference, row.Note)
	}
	fmt.Fprintln(&output)
	fmt.Fprint(&output, tooltext.Format("ecl_phases.progress_line", countPhase(ledger.Rows, false), len(ledger.Rows), countPhase(ledger.Rows, true)))
	return []byte(output.String())
}

func countPhase(rows []OpcodePhase, unknown bool) int {
	total := 0
	for _, row := range rows {
		if (row.Phase == PhaseUnknown) == unknown {
			total++
		}
	}
	return total
}
