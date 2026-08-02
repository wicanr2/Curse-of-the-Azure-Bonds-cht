package sourceaudit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const BaselineSchemaVersion = 1

type Finding struct {
	Path        string `json:"path"`
	Function    string `json:"function,omitempty"`
	SHA256      string `json:"sha256"`
	Occurrences int    `json:"occurrences"`
	Category    string `json:"category"`
}

type Baseline struct {
	SchemaVersion int       `json:"schema_version"`
	Findings      []Finding `json:"findings"`
}

func ScanGoHanLiterals(root string) ([]Finding, error) {
	findings := map[string]*aggregate{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if entry.IsDir() {
			if rel == "workplace" || rel == "golden-box-remake-engine" || rel == ".git" || strings.HasPrefix(rel, "workplace/") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fileSet := token.NewFileSet()
		file, err := parser.ParseFile(fileSet, path, nil, 0)
		if err != nil {
			return fmt.Errorf("parse %s: %w", rel, err)
		}
		function := ""
		ast.Inspect(file, func(node ast.Node) bool {
			switch typed := node.(type) {
			case *ast.FuncDecl:
				previous := function
				function = typed.Name.Name
				if typed.Body != nil {
					inspectFunctionLiterals(typed.Body, rel, function, findings)
				}
				function = previous
				return false
			case *ast.BasicLit:
				if typed.Kind == token.STRING {
					addLiteral(rel, function, typed.Value, findings)
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	result := make([]Finding, 0, len(findings))
	for _, finding := range findings {
		result = append(result, finding.Finding)
	}
	sortFindings(result)
	return result, nil
}

func inspectFunctionLiterals(body *ast.BlockStmt, path, function string, findings map[string]*aggregate) {
	ast.Inspect(body, func(node ast.Node) bool {
		literal, ok := node.(*ast.BasicLit)
		if ok && literal.Kind == token.STRING {
			addLiteral(path, function, literal.Value, findings)
		}
		return true
	})
}

type aggregate struct {
	Finding
}

func addLiteral(path, function, quoted string, findings map[string]*aggregate) {
	value, err := strconv.Unquote(quoted)
	if err != nil || !containsHan(value) {
		return
	}
	digest := sha256.Sum256([]byte(value))
	hash := hex.EncodeToString(digest[:])
	key := path + "\x00" + function + "\x00" + hash
	if existing := findings[key]; existing != nil {
		existing.Occurrences++
		return
	}
	findings[key] = &aggregate{Finding: Finding{
		Path: path, Function: function, SHA256: hash, Occurrences: 1,
		Category: category(path, function),
	}}
}

func containsHan(value string) bool {
	for _, r := range value {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

func category(path, function string) string {
	if strings.Contains(strings.ToLower(function), "localize") || function == "unlockJournalEntries" {
		return "localization_debt"
	}
	if strings.HasPrefix(path, "cmd/azure-bonds-game/") {
		return "frontend_ui_debt"
	}
	return "runtime_ui_debt"
}

func LoadBaseline(path string) (Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Baseline{}, err
	}
	var baseline Baseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		return Baseline{}, err
	}
	if baseline.SchemaVersion != BaselineSchemaVersion {
		return Baseline{}, fmt.Errorf("baseline schema_version=%d, want %d", baseline.SchemaVersion, BaselineSchemaVersion)
	}
	sortFindings(baseline.Findings)
	return baseline, nil
}

func EncodeBaseline(findings []Finding) ([]byte, error) {
	copyFindings := append([]Finding(nil), findings...)
	sortFindings(copyFindings)
	return json.MarshalIndent(Baseline{SchemaVersion: BaselineSchemaVersion, Findings: copyFindings}, "", "  ")
}

func Compare(current, baseline []Finding) (added, removed []Finding) {
	currentMap := findingMap(current)
	baselineMap := findingMap(baseline)
	for key, finding := range currentMap {
		if baselineFinding, ok := baselineMap[key]; !ok || baselineFinding.Occurrences != finding.Occurrences || baselineFinding.Category != finding.Category {
			added = append(added, finding)
		}
	}
	for key, finding := range baselineMap {
		if currentFinding, ok := currentMap[key]; !ok || currentFinding.Occurrences != finding.Occurrences || currentFinding.Category != finding.Category {
			removed = append(removed, finding)
		}
	}
	sortFindings(added)
	sortFindings(removed)
	return added, removed
}

func Summary(findings []Finding) map[string]int {
	result := map[string]int{}
	for _, finding := range findings {
		result[finding.Category] += finding.Occurrences
	}
	return result
}

func findingMap(findings []Finding) map[string]Finding {
	result := make(map[string]Finding, len(findings))
	for _, finding := range findings {
		result[finding.Path+"\x00"+finding.Function+"\x00"+finding.SHA256] = finding
	}
	return result
}

func sortFindings(findings []Finding) {
	sort.Slice(findings, func(i, j int) bool {
		left, right := findings[i], findings[j]
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.Function != right.Function {
			return left.Function < right.Function
		}
		return left.SHA256 < right.SHA256
	})
}
