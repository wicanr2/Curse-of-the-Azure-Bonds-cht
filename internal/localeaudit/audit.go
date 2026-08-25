// Package localeaudit audits the boundaries between the CoAB UI catalog and
// the game-pack catalog. It reports orphan data, but only fails on broken
// references or malformed stable-ID bindings.
package localeaudit

import (
	"encoding/json"
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Issue struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Path     string `json:"path"`
	Detail   string `json:"detail"`
}

type Report struct {
	AssetsLocale   CatalogReport `json:"assets_locale"`
	GamepackLocale CatalogReport `json:"gamepack_locale"`
	ProductUsage   UsageReport   `json:"product_usage"`
	Issues         []Issue       `json:"issues"`
}

type CatalogReport struct {
	Path            string   `json:"path"`
	KeyCount        int      `json:"key_count"`
	MissingKeys     []string `json:"missing_keys,omitempty"`
	EqualEnglish    []string `json:"equal_english,omitempty"`
	ReferencedCount int      `json:"referenced_count"`
	OrphanKeys      []string `json:"orphan_keys,omitempty"`
	ReferencedIDs   []string `json:"referenced_ids,omitempty"`
}

type UsageReport struct {
	ProductFiles      int      `json:"product_files"`
	ReferencedKeys    []string `json:"referenced_keys,omitempty"`
	MissingAssetsKeys []string `json:"missing_assets_keys,omitempty"`
	DynamicTextCalls  int      `json:"dynamic_text_calls"`
}

type rawPack struct {
	Locales     map[string]map[string]string `json:"locales"`
	OptionRules []rawOption                  `json:"option_rules"`
	Raw         map[string]any               `json:"-"`
}

type rawOption struct {
	ID        string `json:"id"`
	Source    string `json:"source"`
	MessageID string `json:"message_id"`
}

// Run loads the two intentionally separate locale layers and audits their
// own consumers. It does not require the assets and game-pack key sets to be
// equal: assets/locale is the frontend/runtime UI catalog; gamepack locales
// are content messages, journal pages, and stable event bindings.
func Run(root string) (Report, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return Report{}, err
	}
	assetsPath := filepath.Join(root, "assets", "locale", "zh-TW.json")
	packPath := filepath.Join(root, "gamepack", "events", "pit-of-moander.json")
	assets, err := loadStrings(assetsPath)
	if err != nil {
		return Report{}, fmt.Errorf("load assets locale: %w", err)
	}
	pack, err := loadPack(packPath)
	if err != nil {
		return Report{}, fmt.Errorf("load gamepack: %w", err)
	}
	usage, err := scanProductUsage(root)
	if err != nil {
		return Report{}, fmt.Errorf("scan product usage: %w", err)
	}

	report := Report{
		AssetsLocale:   CatalogReport{Path: filepath.ToSlash(filepath.Join("assets", "locale", "zh-TW.json")), KeyCount: len(assets)},
		GamepackLocale: CatalogReport{Path: filepath.ToSlash(filepath.Join("gamepack", "events", "pit-of-moander.json"))},
		ProductUsage:   usage,
	}
	report.ProductUsage.MissingAssetsKeys = missing(usage.ReferencedKeys, assets)
	report.AssetsLocale.ReferencedCount = len(usage.ReferencedKeys)
	report.AssetsLocale.OrphanKeys = difference(sortedKeys(assets), usage.ReferencedKeys)

	for _, key := range report.ProductUsage.MissingAssetsKeys {
		report.Issues = append(report.Issues, Issue{"violation", "missing-assets-key", report.AssetsLocale.Path, key})
	}

	en := pack.Locales["en"]
	zh := pack.Locales["zh-TW"]
	report.GamepackLocale.KeyCount = len(zh)
	report.GamepackLocale.MissingKeys = difference(sortedKeys(en), sortedKeys(zh))
	for _, key := range difference(sortedKeys(zh), sortedKeys(en)) {
		report.Issues = append(report.Issues, Issue{"violation", "gamepack-en-zh-asymmetry", report.GamepackLocale.Path, "zh-TW-only: " + key})
	}
	for _, key := range report.GamepackLocale.MissingKeys {
		report.Issues = append(report.Issues, Issue{"violation", "gamepack-en-zh-asymmetry", report.GamepackLocale.Path, "en-only: " + key})
	}
	for key := range en {
		if en[key] == zh[key] {
			report.GamepackLocale.EqualEnglish = append(report.GamepackLocale.EqualEnglish, key)
		}
	}
	sort.Strings(report.GamepackLocale.EqualEnglish)

	referencedIDs, issues := referencedMessageIDs(pack)
	report.GamepackLocale.ReferencedIDs = sortedSet(referencedIDs)
	report.GamepackLocale.ReferencedCount = len(referencedIDs)
	report.Issues = append(report.Issues, issues...)
	for _, key := range difference(sortedKeys(en), sortedSet(referencedIDs)) {
		report.GamepackLocale.OrphanKeys = append(report.GamepackLocale.OrphanKeys, key)
	}
	for _, key := range report.GamepackLocale.OrphanKeys {
		report.Issues = append(report.Issues, Issue{"info", "orphan-gamepack-key", report.GamepackLocale.Path, key})
	}
	for _, key := range report.AssetsLocale.OrphanKeys {
		report.Issues = append(report.Issues, Issue{"info", "orphan-assets-key", report.AssetsLocale.Path, key})
	}
	sortIssues(report.Issues)
	return report, nil
}

func loadStrings(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Strings map[string]string `json:"strings"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if len(doc.Strings) == 0 {
		return nil, fmt.Errorf("strings is empty")
	}
	return doc.Strings, nil
}

func loadPack(path string) (rawPack, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return rawPack{}, err
	}
	var pack rawPack
	if err := json.Unmarshal(data, &pack); err != nil {
		return rawPack{}, err
	}
	if err := json.Unmarshal(data, &pack.Raw); err != nil {
		return rawPack{}, err
	}
	for _, language := range []string{"en", "zh-TW"} {
		if len(pack.Locales[language]) == 0 {
			return rawPack{}, fmt.Errorf("locales.%s is empty or missing", language)
		}
	}
	return pack, nil
}

func scanProductUsage(root string) (UsageReport, error) {
	keys := map[string]bool{}
	files := 0
	dynamic := 0
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
			if rel == ".git" || rel == "workplace" || rel == "golden-box-remake-engine" || strings.HasPrefix(rel, "cmd/pc98-") || strings.HasPrefix(rel, "cmd/azure-bonds") && rel != "cmd/azure-bonds-game" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") || !isProductGoPath(rel) {
			return nil
		}
		files++
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return fmt.Errorf("parse %s: %w", rel, err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 || !isTextCall(call.Fun) {
				return true
			}
			literal, ok := call.Args[0].(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				dynamic++
				return true
			}
			key, err := strconv.Unquote(literal.Value)
			if err == nil && key != "" {
				keys[key] = true
			}
			return true
		})
		return nil
	})
	if err != nil {
		return UsageReport{}, err
	}
	return UsageReport{ProductFiles: files, ReferencedKeys: sortedSet(keys), DynamicTextCalls: dynamic}, nil
}

func isProductGoPath(path string) bool {
	return strings.HasPrefix(path, "internal/game/") || strings.HasPrefix(path, "internal/party/") || strings.HasPrefix(path, "internal/monster/") || strings.HasPrefix(path, "cmd/azure-bonds-game/")
}

func isTextCall(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "Text"
}

func referencedMessageIDs(pack rawPack) (map[string]bool, []Issue) {
	ids := map[string]bool{}
	var issues []Issue
	var visit func(any, string)
	visit = func(value any, path string) {
		switch typed := value.(type) {
		case map[string]any:
			for key, child := range typed {
				childPath := path + "." + key
				if key == "message_id" {
					if message, ok := child.(string); ok && message != "" {
						ids[message] = true
					} else {
						issues = append(issues, Issue{"violation", "empty-message-id", "gamepack/events/pit-of-moander.json", childPath})
					}
				}
				visit(child, childPath)
			}
		case []any:
			for index, child := range typed {
				visit(child, fmt.Sprintf("%s[%d]", path, index))
			}
		}
	}
	visit(pack.Raw, "pack")

	seenIDs, seenSources := map[string]string{}, map[string]string{}
	for index, rule := range pack.OptionRules {
		path := fmt.Sprintf("gamepack/events/pit-of-moander.json:option_rules[%d]", index)
		if rule.ID == "" || rule.Source == "" || rule.MessageID == "" {
			issues = append(issues, Issue{"violation", "invalid-option-rule", path, tooltext.Text("h.c7a03000c800")})
		}
		if previous, ok := seenIDs[rule.ID]; ok && rule.ID != "" {
			issues = append(issues, Issue{"violation", "duplicate-option-id", path, rule.ID + " also at " + previous})
		}
		if previous, ok := seenSources[rule.Source]; ok && rule.Source != "" {
			issues = append(issues, Issue{"violation", "duplicate-option-source", path, rule.Source + " also at " + previous})
		}
		seenIDs[rule.ID] = path
		seenSources[rule.Source] = path
		if rule.MessageID != "" {
			ids[rule.MessageID] = true
		}
	}
	for id := range ids {
		if pack.Locales["en"][id] == "" || pack.Locales["zh-TW"][id] == "" {
			issues = append(issues, Issue{"violation", "missing-gamepack-message", "gamepack/events/pit-of-moander.json", id})
		}
	}
	return ids, issues
}

func missing(keys []string, catalog map[string]string) []string {
	result := make([]string, 0)
	for _, key := range keys {
		if catalog[key] == "" {
			result = append(result, key)
		}
	}
	return result
}

func difference(left, right []string) []string {
	set := map[string]bool{}
	for _, key := range right {
		set[key] = true
	}
	result := make([]string, 0)
	for _, key := range left {
		if !set[key] {
			result = append(result, key)
		}
	}
	return result
}

func sortedKeys(values map[string]string) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func sortedSet(values map[string]bool) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func sortIssues(issues []Issue) {
	sort.Slice(issues, func(i, j int) bool {
		if issues[i].Severity != issues[j].Severity {
			return issues[i].Severity < issues[j].Severity
		}
		if issues[i].Code != issues[j].Code {
			return issues[i].Code < issues[j].Code
		}
		return issues[i].Detail < issues[j].Detail
	})
}

func ViolationCount(report Report) int {
	count := 0
	for _, issue := range report.Issues {
		if issue.Severity == "violation" {
			count++
		}
	}
	return count
}
