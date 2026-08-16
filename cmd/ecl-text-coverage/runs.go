package main

import (
	"fmt"
	"hash/fnv"
	"os"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

// runPath 是一次執行累積起來、**一次**交給 `MatchText` 的那份文字，
// 以及貢獻了這份文字的頁（以 payload offset 表示）。
type runPath struct {
	Texts []string
	Pages []string
}

// pageNode 是「已累積的頁」的持久化串列。走訪要一直分岔（每個 `IF` 兩條路、
// 每張 `ON GOTO` 表十幾條），分岔時複製整份 slice 會讓狀態數與文字長度相乘；
// 串列讓分岔只複製一個指標，被共用的前綴不會被改到。
type pageNode struct {
	text string
	page string
	prev *pageNode
	// depth 與 hash 一起攤平，讓「這個狀態走過了嗎」不必重組整份文字。
	depth int
	hash  uint64
}

func (n *pageNode) collect() ([]string, []string) {
	if n == nil {
		return nil, nil
	}
	texts := make([]string, n.depth)
	pages := make([]string, n.depth)
	for at, node := n.depth-1, n; node != nil; at, node = at-1, node.prev {
		texts[at], pages[at] = node.text, node.page
	}
	return texts, pages
}

func hashOf(prev uint64, text string) uint64 {
	digest := fnv.New64a()
	fmt.Fprintf(digest, "%d|%s", prev, text)
	return digest.Sum64()
}

// stackNode 是 GOSUB 的返回位址堆疊，同樣做成持久化串列。
type stackNode struct {
	offset int
	prev   *stackNode
	hash   uint64
}

func pushReturn(top *stackNode, offset int) *stackNode {
	var prevHash uint64
	if top != nil {
		prevHash = top.hash
	}
	return &stackNode{offset: offset, prev: top, hash: hashOf(prevHash, fmt.Sprint(offset))}
}

// walkRuns 跟著控制流走，列出這個 block 在實機可能產生的每一份 run 文字。
//
// ★ 為什麼不能用 offset 順序切。 工具原本把「位址相鄰的頁」當成同一次執行，
// 在只跟循序與 GOTO 的年代還算堪用。把 `25h ON GOTO` 一起追進來之後就不成立了：
// 一張表底下七個互斥的隨機事件在位址上緊鄰，線性模型會把七段文字併成一份，
// 於是**只寫一條規則，另外六段也會變成「已接上」**——待辦憑空消失。
// 反過來，if/else 的合流（`GOTO` 跳過 else 之後繼續印同一頁）又要求不能在
// `GOTO` 處切開。兩個要求在 offset 順序上無法同時滿足，只能跟著控制流走。
//
// 走訪規則對齊 `internal/ecl/runtime.go`：
//
//   - `12h PRINTCLEAR` 開新頁，`11h PRINT` 接在同一頁後面（spec 1104）。
//   - `16h..1Bh IF` 兩條路都要走：條件不成立時**整條下一個指令被跳過**（spec 1106）。
//   - `02h GOSUB` 進去、`13h RETURN` 回來；堆疊空的 `RETURN` 等於結束。
//   - `25h/26h ON GOTO/GOSUB` 每個目的地各走一條路，index 超出範圍則落到表後面。
//   - `15h/2Bh` 選單的選項字串接在指令後面，長度要自己算（`ecl.MenuEnd`）。
//   - 會 `return result` 的指令（選單、戰鬥、寶物、輸入…）把累積的文字交出去，
//     然後清空繼續走——那正是 `MatchText` 被呼叫的時機。
func walkRuns(label string, data []byte) ([]runPath, map[string]string, error) {
	if len(data) < 2 {
		return nil, nil, nil
	}
	payload := data[2:]
	points, _, err := ecl.EntryPoints(data, 5)
	if err != nil {
		return nil, nil, err
	}
	// pageText 是「這一頁玩家看到什麼」。同一頁在不同路徑上可能長短不同
	// （頁中間有 if/else），留最長的那一份——它才涵蓋得到所有片段。
	pageText := map[string]string{}

	type state struct {
		offset int
		texts  *pageNode
		stack  *stackNode
		open   bool
		steps  int
	}
	key := func(s state) uint64 {
		var textHash, stackHash uint64
		if s.texts != nil {
			textHash = s.texts.hash
		}
		if s.stack != nil {
			stackHash = s.stack.hash
		}
		open := uint64(0)
		if s.open {
			open = 1
		}
		return hashOf(hashOf(uint64(s.offset)*2+open, fmt.Sprint(textHash)), fmt.Sprint(stackHash))
	}

	decode := decoder(data)
	var queue []state
	seen := map[uint64]bool{}
	truncated := false
	push := func(s state) {
		if len(seen) >= maxWalkStates {
			truncated = true
			return
		}
		k := key(s)
		if seen[k] {
			return
		}
		seen[k] = true
		queue = append(queue, s)
	}
	for _, point := range points {
		offset := int(point) - ecl.CodeAddressBase
		if offset < 0 || offset >= len(payload) {
			continue
		}
		push(state{offset: offset})
	}

	var paths []runPath
	// 同一份文字可以由不同路徑產生，而且**印在不同的位址上**（共用的敘述會被
	// 複製好幾份）。去重時只留第一條的話，其餘那些頁就沒有任何 run 認領它們，
	// 報告會把它們算成「沒接上」——明明規則早就命中了那份文字。
	// ⇒ 去重的是文字，不是頁：重複出現時把頁併進既有那一條。
	emitted := map[uint64]int{}
	emit := func(s state) {
		if s.texts == nil {
			return
		}
		texts, pages := s.texts.collect()
		if at, ok := emitted[s.texts.hash]; ok {
			known := map[string]bool{}
			for _, page := range paths[at].Pages {
				known[page] = true
			}
			for _, page := range pages {
				if !known[page] {
					paths[at].Pages = append(paths[at].Pages, page)
				}
			}
			return
		}
		emitted[s.texts.hash] = len(paths)
		paths = append(paths, runPath{Texts: texts, Pages: pages})
	}
	appendPage := func(s state, offset int, text string) state {
		at := fmt.Sprintf("0x%04X", offset)
		var prevHash uint64
		if s.texts != nil {
			prevHash = s.texts.hash
		}
		if s.open && s.texts != nil {
			joined := s.texts.text + " " + text
			s.texts = &pageNode{
				text: joined, page: s.texts.page, prev: s.texts.prev,
				depth: s.texts.depth, hash: hashOf(prevHash, text),
			}
			if len(joined) > len(pageText[s.texts.page]) {
				pageText[s.texts.page] = joined
			}
			return s
		}
		depth := 1
		if s.texts != nil {
			depth = s.texts.depth + 1
		}
		s.texts = &pageNode{text: text, page: at, prev: s.texts, depth: depth, hash: hashOf(prevHash, text)}
		s.open = true
		if len(text) > len(pageText[at]) {
			pageText[at] = text
		}
		return s
	}

	for len(queue) > 0 {
		current := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		depth := 0
		if current.texts != nil {
			depth = current.texts.depth
		}
		if current.steps > maxWalkSteps || depth > maxRunPages ||
			current.offset < 0 || current.offset >= len(payload) {
			emit(current)
			continue
		}
		instruction, err := decode(current.offset)
		if err != nil {
			emit(current)
			continue
		}
		opcode := instruction.Command.Opcode
		// ⚠ 這裡**不能**加「同一條路走過同一個位址就停」的迴圈防護。原作每一頁
		// 結尾都 `GOSUB` 同一支翻頁提示子程式，一條長路徑會踏過它十幾次；
		// 「走兩次就停」會在第三次呼叫時砍掉路徑，該子程式**返回之後**的整段劇情
		// 就此消失（`ECL1.DAX/0x50 +11AEh` 的暗影谷客棧曾經這樣不見）。
		// 迴圈由 `seen`（同樣的位址＋堆疊＋已累積文字不再走第二次）與
		// `maxWalkSteps` 擋住，兩者都不會誤殺「同一支子程式被呼叫很多次」。
		next := state{offset: instruction.Next, texts: current.texts,
			stack: current.stack, open: current.open, steps: current.steps + 1}

		switch {
		case opcode == 0x11 || opcode == 0x12:
			if opcode == 0x12 {
				next.open = false
			}
			if text := instructionText(instruction); text != "" {
				next = appendPage(next, instruction.Offset, text)
			}
			push(next)

		case opcode == 0x01: // GOTO
			if target, ok := branchTargetOf(instruction, len(payload)); ok {
				next.offset = target
				push(next)
			} else {
				emit(current)
			}

		case opcode == 0x02: // GOSUB
			target, ok := branchTargetOf(instruction, len(payload))
			if !ok {
				emit(current)
				continue
			}
			next.stack = pushReturn(next.stack, instruction.Next)
			next.offset = target
			push(next)

		case opcode == 0x13: // RETURN
			if current.stack == nil {
				emit(current)
				continue
			}
			next.offset = current.stack.offset
			next.stack = current.stack.prev
			push(next)

		case opcode == 0x00: // EXIT
			emit(current)

		case opcode == 0x25 || opcode == 0x26:
			targets, after, err := ecl.BranchTargets(data, instruction.Offset)
			if err != nil {
				emit(current)
				continue
			}
			for _, target := range targets {
				branch := next
				if opcode == 0x26 {
					branch.stack = pushReturn(branch.stack, after)
				}
				branch.offset = target
				push(branch)
			}
			// index 超出 count 時原作直接落到表後面。
			next.offset = after
			push(next)

		case opcode >= 0x16 && opcode <= 0x1B: // IF
			push(next)
			if skipped, err := decode(instruction.Next); err == nil {
				elseBranch := next
				elseBranch.offset = skipped.Next
				push(elseBranch)
			}

		case endsRun(opcode):
			// 這裡就是 runtime `return result` 的地方：文字交出去，然後繼續。
			emit(current)
			if opcode == 0x15 || opcode == 0x2B {
				end, err := ecl.MenuEnd(data, instruction.Offset)
				if err != nil {
					continue
				}
				next.offset = end
			}
			next.texts, next.open = nil, false
			push(next)

		default:
			push(next)
		}
	}
	if truncated {
		// 沉默的截斷會讓「走完了」與「走到一半就放棄」長得一模一樣。
		fmt.Fprintf(os.Stderr, "walkRuns: %s 狀態數達到上限 %d，沒走完\n", label, maxWalkStates)
	}
	sort.Slice(paths, func(i, j int) bool {
		return strings.Join(paths[i].Texts, " ") < strings.Join(paths[j].Texts, " ")
	})
	return paths, pageText, nil
}

const (
	// maxWalkStates／maxWalkSteps 是 `IF` 分岔造成的組合爆炸的煞車。碰到上限會
	// 少走幾條路——少走的代價是某些頁留在待辦裡，比反過來安全，而且會印一行警告：
	// 沉默的截斷會讓「走完了」與「放棄了」長得一樣。
	maxWalkStates = 4000000
	maxWalkSteps  = 6000
	// maxRunPages 擋住「會印字的迴圈」把同一段文字無限接下去。原作最長的一次
	// run 是開場捲軸的七頁，留很大的餘裕。
	maxRunPages = 64
)

// decoder 對同一個位址只解一次。走訪會反覆踏過同一段程式碼，逐次重解會讓
// 整份 corpus 的走訪從幾秒變成幾分鐘。
func decoder(block []byte) func(int) (ecl.Instruction, error) {
	cache := map[int]ecl.Instruction{}
	failed := map[int]bool{}
	return func(offset int) (ecl.Instruction, error) {
		if instruction, ok := cache[offset]; ok {
			return instruction, nil
		}
		if failed[offset] {
			return ecl.Instruction{}, fmt.Errorf("cannot decode at %d", offset)
		}
		trace, err := ecl.TraceAt(block, offset, 1)
		if err != nil || len(trace) == 0 {
			failed[offset] = true
			return ecl.Instruction{}, fmt.Errorf("cannot decode at %d", offset)
		}
		cache[offset] = trace[0]
		return trace[0], nil
	}
}

func branchTargetOf(instruction ecl.Instruction, payloadLength int) (int, bool) {
	if len(instruction.Operands) != 1 {
		return 0, false
	}
	return ecl.CodeTarget(instruction.Operands[0], payloadLength)
}
