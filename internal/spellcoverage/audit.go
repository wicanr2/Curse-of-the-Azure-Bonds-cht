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
	"sort"
	"strings"
	"unicode"

	goldenbox "github.com/wicanr2/golden-box-remake-engine/engine"
)

type Report struct {
	PackID       string        `json:"pack_id"`
	SourceFile   string        `json:"source_file"`
	SpellCount   int           `json:"spell_count"`
	CoveredCount int           `json:"covered_count"`
	MissingCount int           `json:"missing_count"`
	Spells       []SpellReport `json:"spells"`
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
			if visual.Trigger != definition.Behavior {
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
	return report, nil
}

func scanRuntime(path string) (map[string]runtimeEvidence, map[string]bool, bool, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, false, err
	}
	file, err := parser.ParseFile(token.NewFileSet(), path, source, 0)
	if err != nil {
		return nil, nil, false, err
	}
	evidence := make(map[string]runtimeEvidence)
	dispatch := make(map[string]bool)
	dispatchFunctions := make(map[string][]string)
	sharedVisualSoundEvents := false
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Body == nil {
			continue
		}
		name := function.Name.Name
		if name == "BeginCombatCast" || name == "CombatCast" || name == "CombatCastWithTerrain" {
			ast.Inspect(function.Body, func(node ast.Node) bool {
				switch value := node.(type) {
				case *ast.CaseClause:
					for _, expression := range value.List {
						if literal, ok := expression.(*ast.BasicLit); ok && literal.Kind.String() == "STRING" {
							behavior := strings.Trim(literal.Value, "\"")
							dispatch[behavior] = true
							dispatchFunctions[behavior] = append(dispatchFunctions[behavior], name)
						}
					}
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
		ast.Inspect(function.Body, func(node ast.Node) bool {
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
		})
		evidence[behavior] = item
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
