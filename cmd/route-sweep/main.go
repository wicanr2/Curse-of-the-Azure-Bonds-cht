// route-sweep 盤點世界地圖上「走這一段路會演什麼」。
//
// ★ 存在的理由：`cmd/cell-sweep` 只掃得到地城——它站在格子上跑地城生命週期。
// 世界地圖那兩個 hub（`ECL1/0x50`／`0x51`）不是地城，它們的每段旅途事件走的是
// 第二種分派形狀：
//
//	4C9D ＝ 出發地 × 4 ＋ 選的方向        { 一條有向邊的編號 }
//	GETTABLE <表>, 4C9D → 7F7A
//	ON GOTO  7F7A                         { 14 個旅途場景 }
//
// 所以「哪一段路演哪一場」要用走的：從世界地圖 hub 直接進去，把原作選單的每個
// 選項都選一次，記下演出來的字與選項本身是不是中文。
//
// ⚠ **這一支量到的是下界，不是全集。** 它只跟著原作選單走（`CurrentOriginalChoices`
// 沒有底線的那種），而世界地圖上大部分的地點要有劇情進度才到得了——目前從 hub
// 直入只走得到暗影谷附近那一圈。它的用途是**找得出沒中文化的字**，不是宣稱
// 「世界地圖全部掃過了」。
//
// ⚠ 每一段路都**從頭重走**：旅途事件有 once-only 旗標與旅行時鐘，接續著走會讓
// 後面的路段看起來沒內容。
//
// ⚠ 這支量的是**內容與語系**，不是可達性。隊伍被撐起來以免路上的戰鬥擋住盤點。
//
// 用法：
//
//	go run ./cmd/route-sweep
//	go run ./cmd/route-sweep -out docs/audit/route-sweep.md -depth 6
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"unicode"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/game"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/gamecorpus"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/segment"
)

// leg 是一段路：從哪個地點、選第幾個方向、演了什麼。
type leg struct {
	hub      string
	from     uint8
	choice   int
	option   string
	to       uint8
	text           string
	language       string
	optionLanguage string
	path           []int
}

// key 是一個分支的身分：hub ＋ 所在地點 ＋ 選項文字。
//
// ⚠ 不能只用「地點 ＋ 第幾個選項」：世界地圖上同一個地點會依劇情停在不同的選單
// （去哪裡、遭遇怎麼應對），第幾個選項在兩個選單之間意思不同。
func (l leg) key() string { return fmt.Sprintf("%s/%d/%s", l.hub, l.from, l.option) }

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "game image zip")
	localePath := flag.String("locale", "assets/locale/zh-TW.json", "locale JSON path")
	// ⚠ 預設寫到 workplace/（不進版控）：這一支的覆蓋是**下界**（見檔頭），
	// 把下界當 audit 產物擺進 repo 會讓人誤以為那就是全集。
	out := flag.String("out", "workplace/route-sweep.md", "輸出的 markdown")
	depth := flag.Int("depth", 6, "從開局最多走幾段路")
	flag.Parse()

	data, err := gamecorpus.Load(*image, *localePath)
	if err != nil {
		log.Fatal(err)
	}
	legs, unreached := []leg(nil), 0
	for _, id := range []string{"ECL1/0x50", "ECL1/0x51"} {
		hub, ok := segment.Lookup(id)
		if !ok {
			log.Fatalf("註冊表沒有 %s", id)
		}
		found, missed := walk(data, hub, *depth)
		legs, unreached = append(legs, found...), unreached+missed
	}
	if err := os.WriteFile(*out, []byte(render(legs, unreached, *depth)), 0o644); err != nil {
		log.Fatal(err)
	}
	counts := summarise(legs)
	fmt.Printf("路段=%d 中文=%d 落回原文=%d 不出文字=%d → %s\n",
		len(legs), counts["中文"], counts["原文"], counts["—"], *out)
}

// walk 從開局的世界地圖做廣度優先：每條路徑都從頭重走一次，走到底再換下一條。
func walk(data gamecorpus.Corpus, hub segment.Segment, depth int) ([]leg, int) {
	seen := map[string]bool{}
	legs := make([]leg, 0, 64)
	queue := [][]int{nil}
	unreached := 0
	for len(queue) > 0 && len(legs) < maxLegs {
		path := queue[0]
		queue = queue[1:]
		if len(path) >= depth {
			continue
		}
		state, ok := replay(data, hub, path)
		if !ok {
			unreached++
			continue
		}
		if !originalMenu(&state) {
			// remake 自己的 UI 選單（紮營、法術、商店…）不是原作的世界地圖分支，
			// 往下走只會把走訪帶進 UI 樹裡。
			continue
		}
		for choice := 0; choice < len(state.Choices); choice++ {
			step := append(append([]int(nil), path...), choice)
			result, ok := travel(data, hub, step)
			if !ok {
				unreached++
				continue
			}
			result.hub, result.path = hub.ID, step
			if seen[result.key()] {
				continue
			}
			seen[result.key()] = true
			legs = append(legs, result)
			queue = append(queue, step)
		}
	}
	sort.Slice(legs, func(a, b int) bool {
		if legs[a].from != legs[b].from {
			return legs[a].from < legs[b].from
		}
		return legs[a].choice < legs[b].choice
	})
	return legs, unreached
}

// maxLegs 是路段數的上限，防止走訪失控。世界地圖宣告的地點不到 30 個、
// 每個地點四個方向，正常不會接近它。
const maxLegs = 400

// replay 從全新一局重走一條路徑，回傳停在世界地圖上的狀態。
func replay(data gamecorpus.Corpus, hub segment.Segment, path []int) (game.State, bool) {
	state, err := data.NewParty()
	if err != nil {
		return state, false
	}
	if err := state.EnterSegment(hub); err != nil {
		return state, false
	}
	if err := gamecorpus.BoostParty(&state); err != nil {
		return state, false
	}
	if !settle(&state) {
		return state, false
	}
	for _, choice := range path {
		if choice >= len(state.Choices) {
			return state, false
		}
		if err := state.Select(choice); err != nil {
			return state, false
		}
		if !settle(&state) {
			return state, false
		}
	}
	return state, state.Mode == game.ModeWilderness && actionableMenu(&state)
}

// actionableMenu 判斷世界地圖上停的是不是玩家真的要做選擇的選單。
//
// ⚠ 世界地圖也會停在只有一個「按任意鍵繼續」的畫面上，那不是選單——把它當選單
// 會把整條走訪卡在同一頁。反過來，**遭遇的應對選單也算**：世界地圖上的分支不
// 只有「去哪裡」，把它們排除掉會漏掉一半的旅途內容。
func actionableMenu(state *game.State) bool {
	switch {
	case len(state.Choices) == 0:
		return false
	case len(state.Choices) == 1 && isContinue(state.Choices[0]):
		return false
	}
	return true
}

// originalMenu 判斷這個選單是不是原作的。原作選單的原文是自然英文
// （`ENTER CITY`、`MEET THEM`），remake 自己的 UI 選單用帶底線的識別字
// （`REST_START`、`TEMPLE_HEAL`）。
func originalMenu(state *game.State) bool {
	originals := state.CurrentOriginalChoices()
	if len(originals) == 0 {
		return false
	}
	for _, option := range originals {
		if strings.Contains(option, "_") {
			return false
		}
	}
	return true
}

func isContinue(option string) bool {
	return strings.Contains(option, "繼續") || strings.Contains(option, "CONTINUE")
}

// travel 重走一條路徑的前段，然後走最後一步並把那一步演的字記下來。
func travel(data gamecorpus.Corpus, hub segment.Segment, path []int) (leg, bool) {
	state, ok := replay(data, hub, path[:len(path)-1])
	if !ok {
		return leg{}, false
	}
	choice := path[len(path)-1]
	if choice >= len(state.Choices) {
		return leg{}, false
	}
	result := leg{from: state.Area.CurrentCity, choice: choice, option: state.Choices[choice]}
	result.optionLanguage = languageOf(result.option)
	if err := state.Select(choice); err != nil {
		return leg{}, false
	}
	// 旅途事件在停下來之前就印出來了，所以邊推進邊收字——停下來才讀會讀到
	// 下一個畫面的字。
	result.text = collect(&state)
	result.to = state.Area.CurrentCity
	result.language = languageOf(result.text)
	return result, true
}

// settle 把畫面推進到停在世界地圖或走不動為止。
func settle(state *game.State) bool {
	for step := 0; step < 24; step++ {
		switch {
		case state.CombatActive():
			for turn := 0; turn < 400 && state.CombatActive(); turn++ {
				if err := state.CombatAct(); err != nil {
					return false
				}
			}
		case state.Mode == game.ModeWilderness && actionableMenu(state):
			return true
		default:
			if err := state.Continue(); err != nil {
				if state.Mode == game.ModeWilderness {
					return true
				}
				return false
			}
		}
	}
	return state.Mode == game.ModeWilderness
}

// collect 把一步之內演出來的字收齊：一段旅途可能連印好幾頁。
func collect(state *game.State) string {
	lines := make([]string, 0, 4)
	push := func(text string) {
		text = strings.TrimSpace(text)
		if text == "" {
			return
		}
		for _, existing := range lines {
			if existing == text {
				return
			}
		}
		lines = append(lines, text)
	}
	push(state.Message)
	push(state.Prompt)
	for step := 0; step < 24; step++ {
		if state.CombatActive() {
			for turn := 0; turn < 400 && state.CombatActive(); turn++ {
				if err := state.CombatAct(); err != nil {
					break
				}
			}
			push(state.Message)
			continue
		}
		if state.Mode == game.ModeWilderness && actionableMenu(state) {
			break
		}
		if err := state.Continue(); err != nil {
			break
		}
		push(state.Message)
		push(state.Prompt)
	}
	return strings.Join(lines, " ／ ")
}

// languageOf 判定一段玩家看得到的字是中文、原文還是沒有字。原作文字是英文，
// 所以「沒有漢字但有英文字母」就是落回原文。
func languageOf(text string) string {
	hasHan, hasLatin := false, false
	for _, glyph := range text {
		switch {
		case unicode.Is(unicode.Han, glyph):
			hasHan = true
		case glyph >= 'A' && glyph <= 'Z', glyph >= 'a' && glyph <= 'z':
			hasLatin = true
		}
	}
	switch {
	case hasHan:
		return "中文"
	case hasLatin:
		return "原文"
	}
	return "—"
}

func summarise(legs []leg) map[string]int {
	counts := map[string]int{}
	for _, item := range legs {
		counts[item.language]++
		if item.optionLanguage == "原文" {
			counts["選項原文"]++
		}
	}
	return counts
}

func render(legs []leg, unreached, depth int) string {
	var out strings.Builder
	out.WriteString("# 世界地圖的路段盤點（走這一段路會演什麼）\n\n" +
		"由 `cmd/route-sweep` 產生，不要手改。\n\n" +
		"原作的旅途事件是「出發地 × 4 ＋ 選的方向」查表分派（`4C9D`），"+
		"所以這一份是用走的：從開局的世界地圖出發，把每個地點的每個方向都選一次，"+
		"記下演出來的字。\n\n" +
		"⚠ 每一段路都**從頭重走**，once-only 旗標與旅行時鐘不互相污染。\n" +
		"⚠ 量的是**內容與語系**，不是可達性——隊伍被撐起來以免路上的戰鬥擋住盤點。\n" +
		fmt.Sprintf("⚠ 從開局最多走 %d 段（`-depth`）。走不到的路段不會出現在表裡，"+
			"**這一份是下界不是全集**。\n\n", depth))
	out.WriteString("| hub | 地點 | 選項 | 選項語言 | 到達 | 語言 | 演出來的字 |\n")
	out.WriteString("|---|---:|---|---|---:|---|---|\n")
	for _, item := range legs {
		out.WriteString(fmt.Sprintf("| `%s` | %d | %s | %s | %d | %s | %s |\n",
			item.hub, item.from, item.option, item.optionLanguage, item.to,
			item.language, firstLine(item.text)))
	}
	counts := summarise(legs)
	out.WriteString(fmt.Sprintf("\n## 摘要\n\n| 項目 | 數 |\n|---|---:|\n"+
		"| 走到的路段 | %d |\n| 演出來是中文 | %d |\n| 演出來落回原文 | %d |\n"+
		"| 走這一段不出文字 | %d |\n| 推不動的分支 | %d |\n",
		len(legs), counts["中文"], counts["原文"], counts["—"], unreached))
	return out.String()
}

func firstLine(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "—"
	}
	runes := []rune(strings.ReplaceAll(text, "\n", " "))
	if len(runes) > 44 {
		return string(runes[:44]) + "…"
	}
	return string(runes)
}
