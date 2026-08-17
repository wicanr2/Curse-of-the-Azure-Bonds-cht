// Package spellcoverage audits the boundary between the title-owned player
// spell pack and the existing CoAB runtime. It deliberately reports missing
// handlers, visuals, and sounds instead of treating a pack declaration as an
// implementation claim.
package spellcoverage

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	goldenbox "github.com/wicanr2/golden-box-remake-engine/engine"
)

type Report struct {
	PackID       string        `json:"pack_id"`
	SourceFile   string        `json:"source_file"`
	SpellCount   int           `json:"spell_count"`
	CoveredCount int           `json:"covered_count"`
	MissingCount int           `json:"missing_count"`
	Spells       []SpellReport `json:"spells"`
	// Original 是**原作那 100 筆**的分母。`SpellCount` 只數 game pack 宣告過的
	// 那幾支，用它當分母會得到「12 支裡 3 支完成」這種與玩家無關的比例。
	Original OriginalCoverage `json:"original_table"`
}

// OriginalCoverage 把 remake 的實作對回原作法術表（`gamepack.Spells`）。
//
// ★ 分母的定義要能講清楚哪一支不算、為什麼：占位那 13 筆玩家取不到，
// 只能紮營施放的 8 支不會出現在戰鬥選單。剩下的才是「戰鬥中玩家可能想施放，
// 但 remake 還沒有」的缺口。
type OriginalCoverage struct {
	// TableSpells 是表的筆數（量測值，spec 815）。
	TableSpells int `json:"table_spells"`
	// Placeholders 是沒有名字的占位項，玩家取不到。
	Placeholders int `json:"placeholders"`
	// Playable ＝ TableSpells − Placeholders。
	Playable int `json:"playable"`
	// CampOnly 是 `+0Bh = 0`（spec 827），不會出現在戰鬥選單。
	CampOnly int `json:"camp_only"`
	// CombatCastable ＝ Playable − CampOnly，戰鬥法術的真正分母。
	CombatCastable int `json:"combat_castable"`
	// Declared 是 game pack 宣告過的戰鬥法術數。
	Declared int `json:"declared"`
	// HandlerObserved 是宣告且 runtime handler 已被觀察到的數量。
	HandlerObserved int `json:"handler_observed"`
	// MissingByLevel 是還沒宣告的戰鬥法術，依環數分。
	MissingByLevel map[int]int `json:"missing_by_level"`
	// MissingNames 是還沒宣告的戰鬥法術名（依編號）。
	MissingNames []string `json:"missing_names"`
	// MissingEffectReady 是還沒宣告、但效果碼 `+0Ah` 已經在 remake 判讀範圍內的
	// 那幾支——它們只差一行 pack 宣告，不需要新的規則程式碼。
	MissingEffectReady int `json:"missing_effect_ready"`
	// MissingEffectReadyNames 同上，列名字，讓下一輪直接照著宣告。
	MissingEffectReadyNames []string `json:"missing_effect_ready_names"`
	// MissingEffectUnhandled 是效果碼有值、但 remake 還看不懂那個碼。
	// **記得上去不等於解讀得了**，這個數字就是那條界線。
	MissingEffectUnhandled int `json:"missing_effect_unhandled"`
	// MissingDamageClass 是 `+0Ah = 0` 的傷害類：骰數不在屬性表裡，
	// 在各自的 overlay-22 handler，所以資料驅動這條路走不到它們。
	MissingDamageClass int `json:"missing_damage_class"`
	// EffectKindsUsed 是全表用到的相異效果碼數（不含 0）。
	EffectKindsUsed int `json:"effect_kinds_used"`
	// EffectKindsInterpreted 是其中 remake 判讀得了的。
	EffectKindsInterpreted int `json:"effect_kinds_interpreted"`
	// UnsupportedClass 是**職業模型還不支援**的法術：德魯伊（remake 沒有這個
	// 職業）與表裡沒有職業的那幾筆。它們不是「還沒做」，是宣告不了——
	// 混在「尚未宣告」裡會讓那個數字永遠降不到 0 而看不出原因。
	UnsupportedClass int `json:"unsupported_class"`
	// Declarable ＝ CombatCastable − UnsupportedClass，才是能收斂到 0 的分母。
	Declarable int `json:"declarable"`
}

type SpellReport struct {
	ID          string        `json:"id"`
	SpellID     uint8         `json:"spell_id"`
	CasterClass string        `json:"caster_class"`
	Behavior    string        `json:"behavior"`
	TargetMode  string        `json:"target_mode"`
	Handler     HandlerReport `json:"handler"`
	Visual      VisualReport  `json:"visual"`
	Sound       SoundReport   `json:"sound"`
	Limitations []string      `json:"limitations"`
}

type HandlerReport struct {
	Function     string `json:"function,omitempty"`
	DispatchCase bool   `json:"dispatch_case"`
	Status       string `json:"status"`
}

type VisualReport struct {
	PackPhases   []string `json:"pack_phases"`
	RuntimeKinds []string `json:"runtime_kinds"`
	Status       string   `json:"status"`
}

type SoundReport struct {
	Events               []string `json:"events"`
	SharedVisualTimeline bool     `json:"shared_visual_timeline"`
	Status               string   `json:"status"`
}

type runtimeEvidence struct {
	Function     string
	DispatchCase bool
	VisualKinds  []string
	SoundEvents  []string
}

// Build parses the existing runtime source and combines its callsite evidence
// with the formal game-pack. It does not execute combat and does not infer
// original AD&D semantics.
func Build(pack *goldenbox.Pack, sourceFile string) (Report, error) {
	evidence, dispatch, sharedVisualSoundEvents, err := scanRuntime(sourceFile)
	if err != nil {
		return Report{}, err
	}
	report := Report{PackID: pack.ID, SourceFile: sourceFile}
	for _, definition := range pack.CombatPlayerSpells {
		item := SpellReport{
			ID: definition.ID, SpellID: definition.SpellID,
			CasterClass: definition.CasterClass, Behavior: definition.Behavior,
			TargetMode: definition.TargetMode,
			Handler:    HandlerReport{DispatchCase: dispatch[definition.Behavior]},
			Visual:     VisualReport{Status: "not_declared"},
			Sound:      SoundReport{Status: "not_observed"},
		}
		if found, ok := evidence[definition.Behavior]; ok {
			item.Handler.Function = found.Function
			item.Handler.DispatchCase = item.Handler.DispatchCase || found.DispatchCase
			item.Visual.RuntimeKinds = append([]string(nil), found.VisualKinds...)
			item.Sound.Events = append([]string(nil), found.SoundEvents...)
		}
		for _, visual := range pack.CombatVisuals {
			// `spell_cast_shared` 是原作那一段**每支法術都會播**的施法投射物
			// （`CASTSPELL` 裡的 `<overlay-24 entry#24>(槽 18)`，spec 1126）。
			// 它不對應任何單一 behavior，所以對每一支宣告過的法術都算數——
			// 這不是放寬標準，是原作真的只有一段共用演出。
			if visual.Trigger != definition.Behavior &&
				visual.Trigger != sharedSpellCastTrigger {
				continue
			}
			item.Visual.PackPhases = append(item.Visual.PackPhases, visual.Phase)
		}
		item.Visual.PackPhases = uniqueSorted(item.Visual.PackPhases)
		item.Visual.RuntimeKinds = uniqueSorted(item.Visual.RuntimeKinds)
		item.Sound.Events = uniqueSorted(item.Sound.Events)
		if len(item.Sound.Events) == 0 && len(item.Visual.RuntimeKinds) != 0 && sharedVisualSoundEvents {
			item.Sound.Events = []string{"shared_visual_timeline"}
			item.Sound.SharedVisualTimeline = true
			item.Sound.Status = "observed_shared"
		}
		switch {
		case item.Handler.Function == "" && !item.Handler.DispatchCase:
			item.Handler.Status = "missing"
		case item.Handler.Function == "" || !item.Handler.DispatchCase:
			item.Handler.Status = "partial"
		default:
			item.Handler.Status = "observed"
		}
		switch {
		case len(item.Visual.PackPhases) == 0 && len(item.Visual.RuntimeKinds) == 0:
			item.Visual.Status = "missing"
		case len(item.Visual.PackPhases) == 0 || len(item.Visual.RuntimeKinds) == 0:
			item.Visual.Status = "partial"
		default:
			item.Visual.Status = "observed"
		}
		switch {
		case len(item.Sound.Events) == 0:
			item.Sound.Status = "missing"
		case item.Sound.Status == "observed_shared":
			// Preserve the more precise shared-timeline status.
		default:
			item.Sound.Status = "observed"
		}
		if item.Handler.Status != "observed" {
			item.Limitations = append(item.Limitations, "runtime handler 尚未由 source callsite 完整觀察")
		}
		if item.Visual.Status != "observed" {
			item.Limitations = append(item.Limitations, "pack／runtime 視覺 binding 不完整；不代表原版動畫已還原")
		}
		if item.Sound.Status != "observed" {
			item.Limitations = append(item.Limitations, "未觀察到該 behavior 的 requestSound intent")
		}
		if len(item.Limitations) == 0 {
			item.Limitations = []string{"僅證明目前 remake callsite 與 binding 存在，不證明完整原版規則、時序或音訊素材"}
		}
		if item.Sound.SharedVisualTimeline {
			item.Limitations = append(item.Limitations, "音效由共用 AdvanceCombatVisual 時間軸觀察，未宣稱每個原版 cue 已 exact")
		}
		report.Spells = append(report.Spells, item)
	}
	report.SpellCount = len(report.Spells)
	for _, item := range report.Spells {
		if item.Handler.Status == "observed" && item.Visual.Status == "observed" && item.Sound.Status == "observed" {
			report.CoveredCount++
		} else {
			report.MissingCount++
		}
	}
	original, err := buildOriginalCoverage(report.Spells)
	if err != nil {
		return Report{}, err
	}
	report.Original = original
	return report, nil
}

// buildOriginalCoverage 把已宣告的法術對回原作表。找不到表就回錯誤——
// 靜靜回一組 0 會讓報告看起來像「原作只有 0 支法術」。
func buildOriginalCoverage(declared []SpellReport) (OriginalCoverage, error) {
	table, err := gamepack.Spells()
	if err != nil {
		return OriginalCoverage{}, err
	}
	handled := map[int]string{}
	for _, item := range declared {
		handled[int(item.SpellID)] = item.Handler.Status
	}
	coverage := OriginalCoverage{
		TableSpells:    len(table.Spells),
		MissingByLevel: map[int]int{},
	}
	effectKinds := map[uint8]bool{}
	for _, spell := range table.Spells {
		if spell.EffectID != 0 {
			effectKinds[uint8(spell.EffectID)] = true
		}
	}
	coverage.EffectKindsUsed = len(effectKinds)
	for kind := range effectKinds {
		if combat.AffectKindIsInterpreted(kind) {
			coverage.EffectKindsInterpreted++
		}
	}
	for _, spell := range table.Spells {
		if spell.Placeholder {
			coverage.Placeholders++
			continue
		}
		coverage.Playable++
		if spell.CampOnly {
			coverage.CampOnly++
			continue
		}
		coverage.CombatCastable++
		if spell.CasterClass != "cleric" && spell.CasterClass != "magic-user" {
			coverage.UnsupportedClass++
			continue
		}
		coverage.Declarable++
		status, ok := handled[spell.SpellID]
		if !ok {
			coverage.MissingByLevel[spell.Level]++
			coverage.MissingNames = append(coverage.MissingNames, spell.Name)
			switch {
			case spell.EffectID == 0:
				coverage.MissingDamageClass++
			case combat.AffectKindIsInterpreted(uint8(spell.EffectID)):
				coverage.MissingEffectReady++
				coverage.MissingEffectReadyNames = append(
					coverage.MissingEffectReadyNames, spell.Name)
			default:
				coverage.MissingEffectUnhandled++
			}
			continue
		}
		coverage.Declared++
		if status == "observed" {
			coverage.HandlerObserved++
		}
	}
	return coverage, nil
}

// runtimeFunctions 讀進要稽核的來源。`path` 是目錄時整包都讀。
//
// ★ 一定要整包讀。施法路後來拆成好幾個檔（`spell_dice.go`、`spell_special.go`、
// `spell_dispel.go`），只讀 `combat_state.go` 會讓那些法術的 visual／sound
// 一律報成沒有——**那是量測的洞，不是覆蓋的洞**，而兩者長得一模一樣。
func runtimeFunctions(path string) ([]*ast.File, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	paths := []string{path}
	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		paths = paths[:0]
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() || !strings.HasSuffix(name, ".go") ||
				strings.HasSuffix(name, "_test.go") {
				continue
			}
			paths = append(paths, filepath.Join(path, name))
		}
		sort.Strings(paths)
	}
	fileSet := token.NewFileSet()
	files := make([]*ast.File, 0, len(paths))
	for _, item := range paths {
		source, err := os.ReadFile(item)
		if err != nil {
			return nil, err
		}
		parsed, err := parser.ParseFile(fileSet, item, source, 0)
		if err != nil {
			return nil, err
		}
		files = append(files, parsed)
	}
	return files, nil
}

// sharedSpellCastTrigger 是共用施法演出的 trigger 名稱（spec 1126）。
const sharedSpellCastTrigger = "spell_cast_shared"

func scanRuntime(path string) (map[string]runtimeEvidence, map[string]bool, bool, error) {
	files, err := runtimeFunctions(path)
	if err != nil {
		return nil, nil, false, err
	}
	declarations := make([]*ast.FuncDecl, 0)
	bodies := make(map[string]*ast.BlockStmt)
	for _, file := range files {
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			declarations = append(declarations, function)
			bodies[function.Name.Name] = function.Body
		}
	}
	evidence := make(map[string]runtimeEvidence)
	dispatch := make(map[string]bool)
	dispatchFunctions := make(map[string][]string)
	behaviorFunctions := make(map[string][]string)
	byFunction := make(map[string]runtimeEvidence)
	sharedVisualSoundEvents := false
	for _, function := range declarations {
		name := function.Name.Name
		if name == "BeginCombatCast" || name == "CombatCast" || name == "CombatCastWithTerrain" {
			ast.Inspect(function.Body, func(node ast.Node) bool {
				clause, ok := node.(*ast.CaseClause)
				if !ok {
					return true
				}
				// 這個 case 實際呼叫哪一支施法函式。名字對不上 behavior 是常態
				// （`effect` 走 `combatCastEffectSpell`、`damage_dice` 走
				// `combatCastSpellDice`），靠名字猜會把那幾支的 visual／sound
				// 全部讀成沒有。
				called := make([]string, 0, 1)
				for _, statement := range clause.Body {
					ast.Inspect(statement, func(child ast.Node) bool {
						call, ok := child.(*ast.CallExpr)
						if !ok {
							return true
						}
						if selector, ok := call.Fun.(*ast.SelectorExpr); ok &&
							strings.HasPrefix(selector.Sel.Name, "combatCast") {
							called = append(called, selector.Sel.Name)
						}
						return true
					})
				}
				for _, expression := range clause.List {
					literal, ok := expression.(*ast.BasicLit)
					if !ok || literal.Kind.String() != "STRING" {
						continue
					}
					behavior := strings.Trim(literal.Value, "\"")
					dispatch[behavior] = true
					dispatchFunctions[behavior] = append(dispatchFunctions[behavior], name)
					behaviorFunctions[behavior] = append(behaviorFunctions[behavior], called...)
				}
				return true
			})
		}
		if name == "AdvanceCombatVisual" {
			ast.Inspect(function.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok || len(call.Args) != 1 {
					return true
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || selector.Sel.Name != "requestSound" {
					return true
				}
				if _, ok := call.Args[0].(*ast.Ident); ok {
					sharedVisualSoundEvents = true
				}
				return true
			})
		}
		if !strings.HasPrefix(name, "combatCast") {
			continue
		}
		behavior := lowerSnake(strings.TrimPrefix(name, "combatCast"))
		if behavior == "" {
			continue
		}
		item := runtimeEvidence{Function: name}
		// 施法收尾抽成共用函式之後，聲音是在那一支發的。只看施法函式本體
		// 會把「有發聲音」讀成「沒有」，所以同套件的呼叫要往下追一層。
		scanned := []*ast.BlockStmt{function.Body}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if body, found := bodies[selector.Sel.Name]; found && body != function.Body {
				scanned = append(scanned, body)
			}
			return true
		})
		inspect := func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			switch selector.Sel.Name {
			case "queueMagicMissileVisual":
				item.VisualKinds = append(item.VisualKinds, "VisualMagicMissile")
			case "requestSound":
				if len(call.Args) == 1 {
					switch value := call.Args[0].(type) {
					case *ast.Ident:
						item.SoundEvents = append(item.SoundEvents, value.Name)
					case *ast.SelectorExpr:
						item.SoundEvents = append(item.SoundEvents, value.Sel.Name)
					}
				}
			case "queueCombatVisual":
				if len(call.Args) == 1 {
					ast.Inspect(call.Args[0], func(child ast.Node) bool {
						if field, ok := child.(*ast.KeyValueExpr); ok && field.Key != nil {
							key, ok := field.Key.(*ast.Ident)
							if !ok || key.Name != "Kind" && key.Name != "Effect" {
								return true
							}
							switch value := field.Value.(type) {
							case *ast.BasicLit:
								if key.Name == "Effect" && value.Kind.String() == "STRING" {
									item.VisualKinds = append(item.VisualKinds, strings.Trim(value.Value, "\""))
								}
							case *ast.SelectorExpr:
								if key.Name == "Kind" {
									item.VisualKinds = append(item.VisualKinds, value.Sel.Name)
								}
							}
						}
						return true
					})
				}
			}
			return true
		}
		for _, body := range scanned {
			ast.Inspect(body, inspect)
		}
		byFunction[name] = item
		evidence[behavior] = item
	}
	// 分派表指到的那一支才是這個 behavior 真正跑的程式。
	for behavior, functions := range behaviorFunctions {
		merged := runtimeEvidence{}
		names := uniqueSorted(functions)
		for _, function := range names {
			item, found := byFunction[function]
			if !found {
				continue
			}
			merged.VisualKinds = append(merged.VisualKinds, item.VisualKinds...)
			merged.SoundEvents = append(merged.SoundEvents, item.SoundEvents...)
		}
		if len(names) == 0 {
			continue
		}
		merged.Function = strings.Join(names, ",")
		evidence[behavior] = merged
	}
	for behavior, functions := range dispatchFunctions {
		if _, found := evidence[behavior]; found {
			continue
		}
		evidence[behavior] = runtimeEvidence{Function: strings.Join(uniqueSorted(functions), ",")}
	}
	if item, found := evidence["magic_missile"]; found {
		item.VisualKinds = append(item.VisualKinds, "VisualMagicMissile")
		evidence["magic_missile"] = item
	}
	return evidence, dispatch, sharedVisualSoundEvents, nil
}

func lowerSnake(value string) string {
	var builder strings.Builder
	for index, runeValue := range value {
		if unicode.IsUpper(runeValue) && index > 0 {
			builder.WriteByte('_')
		}
		builder.WriteRune(unicode.ToLower(runeValue))
	}
	return builder.String()
}

func uniqueSorted(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	sort.Strings(values)
	result := values[:0]
	for _, value := range values {
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return result
}
