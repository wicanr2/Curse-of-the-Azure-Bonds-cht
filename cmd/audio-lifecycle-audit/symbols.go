package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// symbolIndex 把 overlay 模組名對到它的公開程序清單，用來回答「這個位移落在誰身上」。
type symbolIndex struct {
	byModule map[string][]struct {
		offset int
		name   string
	}
}

func loadSymbolIndex(path string) symbolIndex {
	index := symbolIndex{byModule: map[string][]struct {
		offset int
		name   string
	}{}}
	raw, err := os.ReadFile(path)
	if err != nil {
		return index
	}
	var table struct {
		Segments []struct {
			Segment int    `json:"segment"`
			Module  string `json:"module"`
		} `json:"overlay_segments"`
		Symbols []struct {
			Name    string `json:"name"`
			Segment int    `json:"segment"`
			Offset  int    `json:"offset"`
		} `json:"symbols"`
	}
	if err := json.Unmarshal(raw, &table); err != nil {
		return index
	}
	moduleOf := map[int]string{}
	for _, segment := range table.Segments {
		moduleOf[segment.Segment] = segment.Module
	}
	for _, symbol := range table.Symbols {
		module, ok := moduleOf[symbol.Segment]
		if !ok {
			continue
		}
		index.byModule[module] = append(index.byModule[module], struct {
			offset int
			name   string
		}{offset: symbol.Offset, name: symbol.Name})
	}
	for module := range index.byModule {
		list := index.byModule[module]
		sort.Slice(list, func(left, right int) bool { return list[left].offset < list[right].offset })
		index.byModule[module] = list
	}
	return index
}

// routineAt 回傳位移所屬的公開程序；不是入口就標 `名字＋n`。
func (index symbolIndex) routineAt(module string, ea int) string {
	list := index.byModule[module]
	best := -1
	for position, symbol := range list {
		if symbol.offset > ea {
			break
		}
		best = position
	}
	if best < 0 {
		return ""
	}
	if delta := ea - list[best].offset; delta > 0 {
		return fmt.Sprintf("%s＋%Xh", list[best].name, delta)
	}
	return list[best].name
}

// scanRemakeActions 掃 remake 規則層**真的會發出來**的音訊動作字串。
//
// ★ 用 `go/ast` 找 `MusicEvent{Action: "..."}` 這種複合字面，不用 grep：
// grep 會把註解與說明文字也算進去，而那正是「看起來接上了其實沒有」的來源。
func scanRemakeActions(dir string) map[string]bool {
	found := map[string]bool{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return found
	}
	fileSet := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		parsed, parseErr := parser.ParseFile(fileSet, filepath.Join(dir, name), nil, 0)
		if parseErr != nil {
			continue
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			composite, ok := node.(*ast.CompositeLit)
			if !ok {
				return true
			}
			typeName, ok := composite.Type.(*ast.Ident)
			if !ok || typeName.Name != "MusicEvent" {
				return true
			}
			for _, element := range composite.Elts {
				pair, ok := element.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				key, ok := pair.Key.(*ast.Ident)
				if !ok || key.Name != "Action" {
					continue
				}
				if literal, ok := pair.Value.(*ast.BasicLit); ok {
					found[strings.Trim(literal.Value, `"`)] = true
				}
			}
			return true
		})
	}
	return found
}
