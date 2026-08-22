// Command sound-trigger-compare 把原版的音效觸發點與 remake 的發出點並排。
//
// ★ 存在的理由：`pc98-music-triggers.md` 已經解出原版 36 處 `SOUNDFX` 各放哪一個
// 具名音效，但那只是**原版那一側**。要知道 remake 少發或多發了什麼，得把兩邊放在
// 同一張表上。
//
// 原版：PC-98 far-call 表裡打到 `SOUNDX` 段位移 `0000h`（`SOUNDFX`）的呼叫點，
// 引數是音效描述子變數，名字由 Borland 符號表讀出。
// remake：掃 `internal/` 的 `requestSound(SoundXxx)` 呼叫點。
//
// ⚠ **處數不等於播放次數**，兩邊都是。一處在迴圈裡可以響很多次，而同一個音效在
// 原版可能由一支共用常式依武器種類分歧（`sub_2A72` 就同時放箭／揮擊／哨音）。
// 這張表回答的是「**有沒有這條路**」，不是「響幾次」。
//
// ⚠ 名字的對應是**人工對照**（`SWISHFX` ↔ `SoundSwish`），不是自動推的。
// 對錯了會讓兩邊看起來一致或不一致，所以對照表寫在原始碼裡、逐條可查。
//
// 用法：
//
//	go run ./cmd/sound-trigger-compare -output docs/audit/sound-trigger-compare.md
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98sfx"
)

// pairing 是「原版音效 ↔ remake 事件」的對照，**由 `pc98sfx.Selectors()` 推出來**。
//
// ⚠ 這裡以前是一張手打的對照表，而它**漏了 MISSFX 與 SPELLHITFX 兩列**——
// 正好是 remake 用得最兇的 `SoundSpellHit`（29 處）與 `SoundMiss`（2 處）。
// 漏一列不會報錯：報表看起來完整，只是那兩個事件從來沒被比對過。
// 改成從選擇子表推之後，兩邊的名字都來自同一份資料，漏列不可能發生。
//
// remake 那一側的常數名由事件字串反推（`spell_hit` → `SoundSpellHit`），
// 並且在 `main` 裡拿**原始碼裡實際宣告的常數**驗證過——推錯名字會被抓出來，
// 不會安靜地變成「這個事件 0 處」。
type soundPair struct {
	original string
	remake   string
}

func buildPairing() []soundPair {
	pairs := make([]soundPair, 0, 17)
	for _, info := range pc98sfx.Selectors() {
		if info.Event == "" {
			// SOUNDOFF／SOUNDON 是驅動開關，沒有對應的玩法事件。
			continue
		}
		label := soundEffectLabels[info.Symbol]
		name := info.Symbol
		if label != "" {
			name = fmt.Sprintf("%s（%s）", info.Symbol, label)
		}
		pairs = append(pairs, soundPair{original: name, remake: constantName(info.Event)})
	}
	sort.SliceStable(pairs, func(left, right int) bool { return pairs[left].original < pairs[right].original })
	return pairs
}

// soundEffectLabels 只給中文說明；名字與位址都在 `pc98sfx` 那一份。
var soundEffectLabels = map[string]string{
	"SOUNDHALT": "停止", "SOUNDOFF": "關", "SOUNDON": "開",
	"CASTFX": "施法", "MISSFX": "揮空", "SPELLHITFX": "法術命中",
	"DEADFX": "死亡", "WHISTLEFX": "哨音", "HITFX": "命中",
	"LIGHTNINGFX": "閃電", "SWISHFX": "揮擊", "PADFX": "腳步",
	"FIREBALLFX": "火球", "ARROWFX": "箭", "OVERTUREFX": "序曲",
	"COMBATFX": "戰鬥", "CRASHFX": "撞擊",
}

// constantName 把事件字串轉成 remake 的常數名（`spell_hit` → `SoundSpellHit`）。
func constantName(event string) string {
	var out strings.Builder
	out.WriteString("Sound")
	for _, word := range strings.Split(event, "_") {
		if word == "" {
			continue
		}
		out.WriteString(strings.ToUpper(word[:1]))
		out.WriteString(word[1:])
	}
	return out.String()
}

func main() {
	report := flag.String("triggers", "docs/audit/pc98-music-triggers.md", "原版音效觸發報表")
	source := flag.String("source", "internal", "remake 原始碼目錄")
	output := flag.String("output", "", "Markdown 輸出路徑（留白就印到 stdout）")
	flag.Parse()

	original := parseOriginal(*report)
	remake := scanRemake(*source)
	pairing := buildPairing()

	// ⚠ 護欄：**每一個宣告過的 `SoundEvent` 常數都必須有一列**。
	// 少一列不會報錯，報表照樣看起來完整——上一版就是這樣安靜地掉了
	// `SoundSpellHit` 與 `SoundMiss`。這裡拿原始碼裡實際宣告的常數對一次，
	// 對不上就讓它紅，不要印一份看起來很完整的報表。
	declared := scanDeclaredEvents(*source)
	covered := map[string]bool{}
	for _, pair := range pairing {
		covered[pair.remake] = true
	}
	var uncovered []string
	for _, name := range declared {
		if !covered[name] {
			uncovered = append(uncovered, name)
		}
	}
	if len(uncovered) > 0 {
		log.Fatalf("這些 SoundEvent 常數沒有對照列，報表會漏掉它們：%s",
			strings.Join(uncovered, "、"))
	}

	var out strings.Builder
	fmt.Fprintf(&out, "# 音效觸發點：原版 vs remake\n\n")
	fmt.Fprintf(&out, "由 `cmd/sound-trigger-compare` 產生，不要手改。\n\n")
	fmt.Fprintf(&out, "⚠ **處數不等於播放次數**，兩邊都是：一處在迴圈裡可以響很多次，"+
		"而同一個音效在原版可能由一支共用常式依武器種類分歧（`SHOWARROW` 就同時放箭／揮擊／哨音）。"+
		"這張表回答的是「**有沒有這條路**」，不是「響幾次」。逐處的時機在 spec 1186。\n\n")
	fmt.Fprintf(&out, "⚠ 對照關係由 `internal/pc98sfx` 的選擇子表推出來（`SWISHFX` 帶著 `swish`，"+
		"`swish` 對應 `SoundSwish`），不是另外抄一份。名字**判讀**仍然是人做的——"+
		"`MISSFX` 到底是揮空還是法術沒中，要看呼叫端；那一題的答案在 spec 1186。\n\n")
	fmt.Fprintf(&out, "⚠ 原版處數來自**位元組直掃**（見 `pc98-music-triggers.md`），"+
		"不是 far-call 對照表——表比實際少 12 處，而且會把 `LIGHTNINGFX` 印成 0。\n\n")

	fmt.Fprintf(&out, "| 原版音效 | 原版處數 | remake 事件 | remake 處數 | 落差 |\n")
	fmt.Fprintf(&out, "|---|---:|---|---:|---|\n")
	missing, extra := 0, 0
	for _, pair := range pairing {
		left := original[pair.original]
		right := remake[pair.remake]
		note := ""
		switch {
		case left > 0 && right == 0:
			note = "**原版有、remake 從沒發過**"
			missing++
		case left == 0 && right > 0:
			note = "remake 有、原版這一支沒出現在 `SOUNDFX` 立即呼叫裡"
			extra++
		}
		fmt.Fprintf(&out, "| %s | %d | `%s` | %d | %s |\n",
			pair.original, left, pair.remake, right, note)
	}
	fmt.Fprintf(&out, "\n")

	fmt.Fprintf(&out, "## 結論\n\n")
	fmt.Fprintf(&out, "- 對照的音效種類：%d 種。\n", len(pairing))
	fmt.Fprintf(&out, "- **原版有、remake 從沒發過**：%d 種。\n", missing)
	fmt.Fprintf(&out, "- remake 有、原版那一支沒出現在 `SOUNDFX` 的立即呼叫裡：%d 種。\n\n", extra)
	fmt.Fprintf(&out, "⚠ 第二類**先當成掃描面的問題**：上一版把 `LIGHTNINGFX` 列在這一類，"+
		"而它其實在 `CASTSPELL` 裡——是 far-call 對照表看不到，不是原版沒有。"+
		"現在原版那一側改成位元組直掃並且涵蓋常駐，這一類要是還有東西，"+
		"要先問「是不是又有一個面沒掃到」再問「是不是 remake 多做」。\n\n")
	fmt.Fprintf(&out, "⚠ 第一類才是可以動手的：那幾個 `SoundEvent` 常數**宣告了卻從來沒有人送出**"+
		"——編譯得過、測試全綠、玩起來就是少了那個聲音。\n")

	text := out.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "pairs=%d original-only=%d remake-only=%d\n", len(pairing), missing, extra)
}

// parseOriginal 從音效觸發報表撈「音效 → 處數」。
func parseOriginal(path string) map[string]int {
	counts := map[string]int{}
	raw, err := os.ReadFile(path)
	if err != nil {
		return counts
	}
	row := regexp.MustCompile(`^\|\s*([A-Z]+FX（[^）]*）|SOUND[A-Z]+（[^）]*）)\s*\|\s*(\d+)\s*\|`)
	for _, line := range strings.Split(string(raw), "\n") {
		match := row.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		count := 0
		fmt.Sscanf(match[2], "%d", &count)
		counts[match[1]] += count
	}
	return counts
}

// scanRemake 數 `requestSound(SoundXxx)` 的呼叫點。
func scanRemake(root string) map[string]int {
	counts := map[string]int{}
	call := regexp.MustCompile(`requestSound\((Sound[A-Za-z]+)\)`)
	filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		// ⚠ 測試檔不算：那是驗證用的呼叫，不是遊戲真的會走的路。
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		for _, match := range call.FindAllStringSubmatch(string(raw), -1) {
			counts[match[1]]++
		}
		return nil
	})
	return counts
}


// scanDeclaredEvents 讀出 remake **宣告**了哪些 `SoundEvent` 常數。
//
// ★ 與 `scanRemake` 的差別是「宣告」與「送出」：一個常數可以宣告了卻從來沒有
// 人送出（那正是這份報表要抓的東西），但**不能宣告了卻沒有對照列**——後者是
// 報表自己的洞，會讓那個事件連被檢查的機會都沒有。
func scanDeclaredEvents(root string) []string {
	decl := regexp.MustCompile(`(Sound[A-Za-z]+)\s+SoundEvent\s*=`)
	names := map[string]bool{}
	filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		for _, match := range decl.FindAllStringSubmatch(string(raw), -1) {
			names[match[1]] = true
		}
		return nil
	})
	list := make([]string, 0, len(names))
	for name := range names {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}
