package eclcatalog

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The phase ledger anchors every conclusion to a DOS handler address. That
// address is also published by scripts/ecl_dispatch_table.py in a separate
// generated table. Copying it into Go creates a drift risk, so pin the two
// together instead of trusting that they were updated in the same commit.
func TestPhaseLedgerHandlerAddressesMatchDispatchTable(t *testing.T) {
	dispatch := readDispatchTable(t)
	for _, row := range PhaseRows() {
		want, ok := dispatch[row.Opcode]
		if !ok {
			t.Fatalf("opcode %s is absent from the generated dispatch table", row.Opcode)
		}
		if row.DOSHandler != want {
			t.Fatalf("opcode %s DOS handler %s, dispatch table says %s",
				row.Opcode, row.DOSHandler, want)
		}
	}
}

// readDispatchTable parses the `| `00h` | `0052h` | ... |` rows of the
// generated dispatch audit. Only the opcode and DOS columns are read.
func readDispatchTable(t *testing.T) map[string]string {
	t.Helper()
	path := filepath.Join("..", "..", "docs", "audit", "ecl-opcode-dispatch.md")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	table := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), "|")
		if len(fields) < 4 {
			continue
		}
		opcode := strings.Trim(strings.TrimSpace(fields[1]), "`")
		handler := strings.Trim(strings.TrimSpace(fields[2]), "`")
		if !strings.HasSuffix(opcode, "h") || !strings.HasSuffix(handler, "h") {
			continue
		}
		value, err := parseHexByte(strings.TrimSuffix(opcode, "h"))
		if err != nil {
			continue
		}
		table[value] = handler
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(table) == 0 {
		t.Fatal("dispatch table produced no rows; the generated format changed")
	}
	return table
}

func parseHexByte(text string) (string, error) {
	var value int
	for _, char := range text {
		digit := strings.IndexRune("0123456789ABCDEF", char)
		if digit < 0 {
			return "", os.ErrInvalid
		}
		value = value*16 + digit
	}
	return hexByte(byte(value)), nil
}

// A phase must never be asserted without evidence behind it, and evidence must
// never be recorded without a phase.
func TestPhaseLedgerConfidenceMatchesEvidence(t *testing.T) {
	for _, row := range PhaseRows() {
		known := row.Phase != PhaseUnknown
		switch {
		case known && row.Confidence == "unknown":
			t.Fatalf("opcode %s claims phase %s with unknown confidence", row.Opcode, row.Phase)
		case known && len(row.SpecRefs) == 0:
			t.Fatalf("opcode %s claims phase %s without a spec reference", row.Opcode, row.Phase)
		case !known && len(row.SpecRefs) > 0 && row.Confidence != "unknown":
			t.Fatalf("opcode %s has evidence but no phase", row.Opcode)
		}
		if row.Note == "" {
			t.Fatalf("opcode %s has no note", row.Opcode)
		}
	}
}

// NEWECL is the reason `endsStraightLine` grew a member. Pin the classification
// so a later edit cannot quietly turn it back into a fallthrough opcode.
func TestNewEclIsTerminalAndEndsStraightLine(t *testing.T) {
	row, ok := PhaseFor("0x20")
	if !ok {
		t.Fatal("NEWECL has no phase row")
	}
	if row.Phase != PhaseTerminal {
		t.Fatalf("NEWECL phase %s, want %s", row.Phase, PhaseTerminal)
	}
	if !endsStraightLine(0x20) {
		t.Fatal("NEWECL must end a straight-line candidate region")
	}
}

// The commit point is a single opcode branch. If a second opcode ever claims
// it, the ordered-effect model has changed and the spec must move with it.
func TestExactlyOneCommitPointOpcode(t *testing.T) {
	var found []string
	for _, row := range PhaseRows() {
		if row.Phase == PhaseCommitPoint {
			found = append(found, row.Opcode)
		}
	}
	if len(found) != 1 || found[0] != "0x2D" {
		t.Fatalf("commit-point opcodes %v, want exactly [0x2D]", found)
	}
}
