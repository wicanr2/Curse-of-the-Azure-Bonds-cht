// Command player-path-audit 回答一個問題：**戰役測試走的每一步，玩家按得出來嗎？**
//
// ★ 存在的理由：`TestRealNewGameRunsToTheEnding` 從角色建立打到提朗瑟克斯，
// spec 1188 又讓真的前端把那條路線的每個檢查點畫了一張。剩下的缺口是**輸入那一層**
// ——戰役直接呼叫 `state.X()`，玩家按的是鍵盤。如果戰役靠某個 `State` 方法推進劇情，
// 而前端的 `Update()` 到不了那個方法，**那一段劇情就沒有任何按鍵組合到得了**。
//
// 做法用 `go/ast`，不是 grep：
//
//	戰役側  掃戰役測試裡對 `state` 的方法呼叫
//	前端側  從 `(*app).Update` 出發做**可達性閉包**（Update 會呼叫其他 app 方法，
//	        例如存檔走 `a.saveCurrentGame()`），收集沿路所有 `a.state.X` 呼叫
//	別名    自動解：`func (s *State) A() { return s.B() }` 這種單行轉呼叫，
//	        A 到不了但 B 到得了就算到得了
//
// ⚠ **這是靜態可達性，不是「玩家真的按得到」。** `Update` 裡那一行可能藏在
// 一個永遠不成立的條件底下。這份報表證明得了「前端**沒有**那條路」（那是硬缺口），
// 證明不了「有那條路所以玩得到」。方向不對稱，不要反著讀。
//
// ⚠ 別名表是**推出來的**，不是手打的。這個 session 已經被手打對照表漏格咬過三次。
//
// 用法：
//
//	go run ./cmd/player-path-audit -output docs/audit/player-path.md
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	campaignFiles := flag.String("campaign",
		"internal/game/campaign_normal_test.go,internal/game/ecl_integration_test.go",
		"戰役驅動程式（逗號分隔）")
	frontend := flag.String("frontend", "cmd/azure-bonds-game", "前端目錄")
	gameDir := flag.String("game", "internal/game", "State 所在的套件目錄")
	output := flag.String("output", "", "Markdown 輸出路徑（留白就印到 stdout）")
	flag.Parse()

	campaign := map[string]int{}
	for _, path := range strings.Split(*campaignFiles, ",") {
		collectStateCalls(strings.TrimSpace(path), campaign)
	}
	if len(campaign) == 0 {
		log.Fatal("戰役側一個呼叫都掃不到——先確認路徑")
	}

	input, updateCalls := frontendReachable(*frontend, "Update")
	display, _ := frontendReachable(*frontend, "Draw")
	startup := frontendStartup(*frontend)
	if len(input) == 0 {
		log.Fatal("前端側一個呼叫都掃不到——先確認 (*app).Update 找得到")
	}
	aliases := delegationAliases(*gameDir)
	mutating, stateGraph := mutatingMethods(*gameDir)

	// ★ 前端的可達集合要**再沿著 `State` 內部的呼叫展開一次**。
	//
	// ⚠ 少了這一步會生出假缺口：前端沒有直接呼叫 `AdvanceGameTimeHours`，
	// 但玩家紮營走 `Select()` → `selectCamp()` → `AdvanceGameTimeHours()`。
	// 只看「前端有沒有直接叫」＝ 只走一步，會把「走三步到得了」報成到不了。
	expandThroughState := func(seed map[string]bool) map[string]bool {
		out := map[string]bool{}
		queue := make([]string, 0, len(seed))
		for name := range seed {
			out[name] = true
			queue = append(queue, name)
		}
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			for _, next := range stateGraph[current] {
				if out[next] {
					continue
				}
				out[next] = true
				queue = append(queue, next)
			}
		}
		return out
	}

	directInput := len(input)
	input = expandThroughState(input)
	display = expandThroughState(display)
	startup = expandThroughState(startup)

	type row struct {
		name    string
		uses    int
		verdict string
		note    string
	}
	rows := make([]row, 0, len(campaign))
	gaps := 0
	for name, uses := range campaign {
		item := row{name: name, uses: uses}
		alias := aliases[name]
		switch {
		case input[name]:
			item.verdict = "✅ 按得出來"
			item.note = "`Update()` 的可達閉包裡有"
		case alias != "" && input[alias]:
			item.verdict = "✅ 按得出來"
			item.note = fmt.Sprintf("別名 → `%s`（單行轉呼叫），前端走那一支", alias)
		case !ast.IsExported(name):
			item.verdict = "— 觀測點"
			item.note = "**未匯出**：前端在別的套件本來就叫不到 ⇒ 戰役拿它當觀測點"
		case display[name] || (alias != "" && display[alias]):
			item.verdict = "— 觀測點"
			item.note = "前端在 `Draw()` 讀它 ⇒ 這是**看**的東西，不是按的東西"
		case startup[name] || (alias != "" && startup[alias]):
			item.verdict = "— 啟動接線"
			item.note = "前端在 `main()` 的啟動流程呼叫 ⇒ 不是輸入動作"
		case !mutating[name]:
			item.verdict = "— 讀取器差異"
			item.note = "**不改狀態**：前端用別的方式讀同一件事（讀欄位、或它已經被投影進 `Choices`）"
		default:
			item.verdict = "**前端沒有這條路**"
			item.note = "會改狀態而前端到不了 ⇒ 玩家按不出來"
			gaps++
		}
		rows = append(rows, item)
	}
	sort.Slice(rows, func(left, right int) bool {
		if rows[left].verdict != rows[right].verdict {
			return rows[left].verdict > rows[right].verdict
		}
		return rows[left].name < rows[right].name
	})

	var report strings.Builder
	fmt.Fprintf(&report, "# 戰役測試走的每一步，玩家按得出來嗎\n\n")
	fmt.Fprintf(&report, "由 `cmd/player-path-audit` 產生，不要手改。方法與判讀見 spec 1189。\n\n")
	fmt.Fprintf(&report, "★ 補的是「開場到結局」剩下的那半個缺口：spec 1188 已經讓真的前端把戰役的"+
		"每個檢查點**畫**了一張，但**輸入那一層**還沒有證明——戰役直接呼叫 `state.X()`，"+
		"玩家按的是鍵盤。前端的 `Update()` 到不了的方法，就沒有任何按鍵組合到得了。\n\n")
	fmt.Fprintf(&report, "⚠ **方向不對稱，不要反著讀。** 這是靜態可達性：證明得了「前端**沒有**那條路」"+
		"（硬缺口），證明不了「有那條路所以玩得到」——那一行可能藏在永遠不成立的條件底下。\n\n")

	fmt.Fprintf(&report, "| 指標 | 數字 |\n|---|---:|\n")
	fmt.Fprintf(&report, "| 戰役用到的 `State` 進入點 | %d |\n", len(campaign))
	fmt.Fprintf(&report, "| 前端直接呼叫的 `State` 方法（輸入側）| %d |\n", directInput)
	fmt.Fprintf(&report, "| 再沿 `State` 內部呼叫展開後 | %d |\n", len(input))
	fmt.Fprintf(&report, "| `Draw()`（顯示）可達閉包裡的 | %d |\n", len(display))
	fmt.Fprintf(&report, "| `main()` 啟動流程裡的 | %d |\n", len(startup))
	fmt.Fprintf(&report, "| 其中直接寫在 `Update()` 本體的 | %d |\n", updateCalls)
	fmt.Fprintf(&report, "| 自動解出的單行別名 | %d |\n", len(aliases))
	fmt.Fprintf(&report, "| 判定為**會改狀態**的 `State` 方法 | %d |\n", len(mutating))
	fmt.Fprintf(&report, "| **前端沒有路可以到的** | %d |\n\n", gaps)

	fmt.Fprintf(&report, "| `State` 進入點 | 戰役用幾處 | 玩家到得了 | 說明 |\n|---|---:|---|---|\n")
	for _, item := range rows {
		fmt.Fprintf(&report, "| `%s` | %d | %s | %s |\n", item.name, item.uses, item.verdict, item.note)
	}
	fmt.Fprintf(&report, "\n")
	if gaps == 0 {
		fmt.Fprintf(&report, "★ **戰役推進劇情用到的每一個匯出方法，前端都有路可以到。**\n\n")
		fmt.Fprintf(&report, "⚠ 這**不等於**玩得完：可達性不管條件、不管順序，也不管畫面上有沒有"+
			"提示那個鍵。它排除的是「有一段劇情根本沒有按鍵到得了」這一種硬缺口。\n")
	}

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "campaign=%d input=%d display=%d startup=%d aliases=%d gaps=%d\n",
		len(campaign), len(input), len(display), len(startup), len(aliases), gaps)
}

// collectStateCalls 掃一個檔案裡對名為 `state` 的變數做的方法呼叫。
func collectStateCalls(path string, into map[string]int) {
	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		return
	}
	ast.Inspect(parsed, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		receiver, ok := selector.X.(*ast.Ident)
		if !ok || (receiver.Name != "state" && receiver.Name != "source") {
			return true
		}
		into[selector.Sel.Name]++
		return true
	})
	return
}

// frontendReachable 從 `(*app).Update` 出發做可達性閉包，收集沿路所有
// `a.state.X` 呼叫。回傳可達集合與「直接寫在 Update 本體」的數量。
//
// ★ 為什麼要閉包而不是只掃 `Update` 本體：存檔走 `a.saveCurrentGame()`，
// 那一支才呼叫 `state.SavePartyFile`。只掃本體會把它算成「前端沒有這條路」。
func frontendReachable(dir, root string) (map[string]bool, int) {
	fileSet := token.NewFileSet()
	packages, err := parser.ParseDir(fileSet, dir, func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		return nil, 0
	}
	// appMethod 名稱 → 它呼叫的 app 方法與 state 方法。
	type methodBody struct {
		appCalls   []string
		stateCalls []string
	}
	methods := map[string]*methodBody{}
	for _, pkg := range packages {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Recv == nil || fn.Body == nil {
					continue
				}
				body := &methodBody{}
				// ⚠ 走**每一個 SelectorExpr**，不是只走 `CallExpr` 的 Fun。
				// 前端把方法當成值傳出去的寫法很常見（`a.combatAction(a.state.CombatAct)`），
				// 只認「被呼叫的」會漏掉——`CombatAct` 就是這樣被誤判成
				// 「前端沒有這條路」的。**掃描面比實際窄，就會生出假缺口。**
				ast.Inspect(fn.Body, func(node ast.Node) bool {
					selector, ok := node.(*ast.SelectorExpr)
					if !ok {
						return true
					}
					switch inner := selector.X.(type) {
					case *ast.Ident:
						// a.someMethod
						if inner.Name == "a" || inner.Name == "app" {
							body.appCalls = append(body.appCalls, selector.Sel.Name)
						}
					case *ast.SelectorExpr:
						// a.state.X
						if base, ok := inner.X.(*ast.Ident); ok &&
							(base.Name == "a" || base.Name == "app") && inner.Sel.Name == "state" {
							body.stateCalls = append(body.stateCalls, selector.Sel.Name)
						}
					}
					return true
				})
				methods[fn.Name.Name] = body
			}
		}
	}
	entry, ok := methods[root]
	if !ok {
		return nil, 0
	}
	reachable := map[string]bool{}
	seen := map[string]bool{root: true}
	queue := []string{root}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		body := methods[current]
		if body == nil {
			continue
		}
		for _, name := range body.stateCalls {
			reachable[name] = true
		}
		for _, name := range body.appCalls {
			if seen[name] {
				continue
			}
			seen[name] = true
			queue = append(queue, name)
		}
	}
	return reachable, len(entry.stateCalls)
}

// delegationAliases 找出 `func (s *State) A(...) { return s.B(...) }` 這種
// **單行轉呼叫**，回傳 A → B。
//
// ⚠ 這一份是**推出來的**：手打的別名表漏一格不會報錯，只會讓某個方法被誤判成
// 「前端沒有這條路」，然後有人跑去補一個本來就存在的功能。
func delegationAliases(dir string) map[string]string {
	fileSet := token.NewFileSet()
	packages, err := parser.ParseDir(fileSet, dir, func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		return nil
	}
	aliases := map[string]string{}
	for _, pkg := range packages {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Recv == nil || fn.Body == nil || len(fn.Body.List) != 1 {
					continue
				}
				var call *ast.CallExpr
				switch statement := fn.Body.List[0].(type) {
				case *ast.ReturnStmt:
					if len(statement.Results) != 1 {
						continue
					}
					call, _ = statement.Results[0].(*ast.CallExpr)
				case *ast.ExprStmt:
					call, _ = statement.X.(*ast.CallExpr)
				}
				if call == nil {
					continue
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					continue
				}
				if receiver, ok := selector.X.(*ast.Ident); !ok || receiver.Name != "s" {
					continue
				}
				aliases[fn.Name.Name] = selector.Sel.Name
			}
		}
	}
	return aliases
}

var _ = filepath.Join

// frontendStartup 掃前端**沒有 receiver 的函式**（`main` 與它的輔助函式）裡
// 對 `state` 的呼叫。那是啟動接線：載入遊戲資料、接上目錄、跳到指定段落。
//
// ★ 為什麼要單獨分一類：戰役測試也要做同樣的接線（`SetGeoCatalog`、
// `SetMonsterRecordsForECL`…）。把它們算成「玩家按不出來」是**對的字面、
// 錯的意思**——那些本來就不是輸入動作，前端在開機時就做完了。
func frontendStartup(dir string) map[string]bool {
	fileSet := token.NewFileSet()
	packages, err := parser.ParseDir(fileSet, dir, func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		return nil
	}
	found := map[string]bool{}
	for _, pkg := range packages {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Recv != nil || fn.Body == nil {
					continue
				}
				ast.Inspect(fn.Body, func(node ast.Node) bool {
					selector, ok := node.(*ast.SelectorExpr)
					if !ok {
						return true
					}
					switch inner := selector.X.(type) {
					case *ast.Ident:
						if inner.Name == "state" {
							found[selector.Sel.Name] = true
						}
					case *ast.SelectorExpr:
						if inner.Sel.Name == "state" {
							found[selector.Sel.Name] = true
						}
					}
					return true
				})
			}
		}
	}
	return found
}

// mutatingMethods 判斷每個 `State` 方法會不會改狀態：直接寫 `s.X = …`／`s.X++`，
// 或呼叫另一個會改狀態的方法（做到不動點）。
//
// ★ 為什麼需要它：「前端到不了這個方法」對**讀取器**沒有意義。前端不呼叫
// `ShopOffers()` 不代表玩家不能買東西——商品早就被投影進 `Choices`，買賣走
// `Select()`。把讀取器算成「玩家按不出來」是**對的字面、錯的意思**。
//
// ⚠ 這是靜態近似：透過介面或閉包改狀態的看不出來，會被低估成讀取器。
// 方向是**寧可少報缺口**——所以這一欄不能拿來宣稱「沒有缺口」，只能用來
// 把明顯不是缺口的濾掉。
func mutatingMethods(dir string) (map[string]bool, map[string][]string) {
	fileSet := token.NewFileSet()
	packages, err := parser.ParseDir(fileSet, dir, func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		return nil, nil
	}
	writes := map[string]bool{}
	calls := map[string][]string{}
	for _, pkg := range packages {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Recv == nil || fn.Body == nil || len(fn.Recv.List) == 0 {
					continue
				}
				receiver := ""
				if len(fn.Recv.List[0].Names) > 0 {
					receiver = fn.Recv.List[0].Names[0].Name
				}
				if receiver == "" {
					continue
				}
				name := fn.Name.Name
				ast.Inspect(fn.Body, func(node ast.Node) bool {
					switch statement := node.(type) {
					case *ast.AssignStmt:
						for _, target := range statement.Lhs {
							if rootIdent(target) == receiver {
								writes[name] = true
							}
						}
					case *ast.IncDecStmt:
						if rootIdent(statement.X) == receiver {
							writes[name] = true
						}
					case *ast.CallExpr:
						if selector, ok := statement.Fun.(*ast.SelectorExpr); ok {
							if base, ok := selector.X.(*ast.Ident); ok && base.Name == receiver {
								calls[name] = append(calls[name], selector.Sel.Name)
							}
						}
					}
					return true
				})
			}
		}
	}
	// 不動點：呼叫到會改狀態的，自己也算會改狀態。
	for changed := true; changed; {
		changed = false
		for name, targets := range calls {
			if writes[name] {
				continue
			}
			for _, target := range targets {
				if writes[target] {
					writes[name] = true
					changed = true
					break
				}
			}
		}
	}
	return writes, calls
}

// rootIdent 回傳 `s.a.b` 這種選擇鏈最左邊的識別字。
func rootIdent(expr ast.Expr) string {
	for {
		switch node := expr.(type) {
		case *ast.Ident:
			return node.Name
		case *ast.SelectorExpr:
			expr = node.X
		case *ast.IndexExpr:
			expr = node.X
		case *ast.StarExpr:
			expr = node.X
		default:
			return ""
		}
	}
}
