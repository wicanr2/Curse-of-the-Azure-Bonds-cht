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
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
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
	// symbol 是原版的符號名（`SOUNDHALT`），不帶中文標籤——`judgedGaps` 用它當鍵。
	symbol   string
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
		pairs = append(pairs, soundPair{symbol: info.Symbol, original: name, remake: constantName(info.Event)})
	}
	sort.SliceStable(pairs, func(left, right int) bool { return pairs[left].original < pairs[right].original })
	return pairs
}

// soundEffectLabels 只給中文說明；名字與位址都在 `pc98sfx` 那一份。
var soundEffectLabels = map[string]string{
	"SOUNDHALT": tooltext.Text("h.ca4d973c0b00"), "SOUNDOFF": tooltext.Text("h.94ffcbc4a678"), "SOUNDON": tooltext.Text("h.320d1ab4f80f"),
	"CASTFX": tooltext.Text("h.e377c6152522"), "MISSFX": tooltext.Text("h.d7aa5b20173e"), "SPELLHITFX": tooltext.Text("h.0118c5026c62"),
	"DEADFX": tooltext.Text("h.82d3130fa582"), "WHISTLEFX": tooltext.Text("h.d0928210330c"), "HITFX": tooltext.Text("h.393df9bb13ea"),
	"LIGHTNINGFX": tooltext.Text("h.48693c8fd90b"), "SWISHFX": tooltext.Text("h.692fbefa6024"), "PADFX": tooltext.Text("h.9f18c28a4e63"),
	"FIREBALLFX": tooltext.Text("h.010568a02690"), "ARROWFX": tooltext.Text("h.674650b36c70"), "OVERTUREFX": tooltext.Text("h.3df5903dbdcd"),
	"COMBATFX": tooltext.Text("h.625dd417c2c3"), "CRASHFX": tooltext.Text("h.b15e09c78fd2"),
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

// judgedGaps 是「原版有、remake 沒有，但已經判定不是缺漏」的音效，值是理由。
//
// ★ 存在的理由：這一欄的數字如果只是「1」，下一個人（或下一個 session）會把它
// 當成待辦重新查一次。判過的結論要留在**產生報表的程式碼裡**，不是留在某一輪的
// 敘述裡——敘述會被讀成歷史，表才會跟著每次重生的報表一起出現。
var judgedGaps = map[string]string{
	"SOUNDHALT": tooltext.Text("h.3d8de43a7173") +
		tooltext.Text("h.12c9e18e12c6") +
		tooltext.Text("h.1784b7da1d5a"),
}

func main() {
	report := flag.String("triggers", "docs/audit/pc98-music-triggers.md", tooltext.Text("h.abacf712d389"))
	source := flag.String("source", "internal", tooltext.Text("h.90a2533b98a1"))
	output := flag.String("output", "", tooltext.Text("h.78eb014c7900"))
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
		log.Fatal(tooltext.Format("h.16741c936afb", strings.Join(uncovered, "、")))
	}

	var out strings.Builder
	fmt.Fprint(&out, tooltext.Format("h.4503f49c11aa"))
	fmt.Fprint(&out, tooltext.Format("h.b8421d3efe4c"))
	fmt.Fprint(&out, tooltext.Text("h.f975cc008699")+
		tooltext.Text("h.a8a5957a920d")+
		tooltext.Text("h.66cf4dc87dd2"))
	fmt.Fprint(&out, tooltext.Text("h.7be774eba0d4")+
		tooltext.Text("h.ce265e423e6e")+
		tooltext.Text("h.3a0883da1285"))
	fmt.Fprint(&out, tooltext.Text("h.a917245539cf")+
		tooltext.Text("h.3212e2c7cc17")+
		tooltext.Text("h.27610e57ffc9"))
	fmt.Fprint(&out, tooltext.Text("h.5b9cf8590d39")+
		tooltext.Text("h.d2e76307e732"))

	fmt.Fprint(&out, tooltext.Format("h.9226baf238d5"))
	fmt.Fprintf(&out, "|---|---:|---|---:|---|\n")
	missing, extra, judged := 0, 0, 0
	for _, pair := range pairing {
		left := original[pair.original]
		right := remake[pair.remake]
		note := ""
		switch {
		case left > 0 && right == 0:
			if reason, ok := judgedGaps[pair.symbol]; ok {
				note = tooltext.Text("h.75d39151abed") + reason
				judged++
				break
			}
			note = tooltext.Text("h.03decb8a605d")
			missing++
		case left == 0 && right > 0:
			note = tooltext.Text("h.fdd4335812bd")
			extra++
		}
		fmt.Fprintf(&out, "| %s | %d | `%s` | %d | %s |\n",
			pair.original, left, pair.remake, right, note)
	}
	fmt.Fprintf(&out, "\n")

	fmt.Fprint(&out, tooltext.Format("h.548e4db4f8a3"))
	fmt.Fprint(&out, tooltext.Format("h.38cf3e7ea3ea", len(pairing)))
	fmt.Fprint(&out, tooltext.Format("h.9fcd94b333d0", missing))
	fmt.Fprint(&out, tooltext.Format("h.38a5a46c1b4e", judged))
	fmt.Fprint(&out, tooltext.Format("h.a155cc48aae8", extra))
	fmt.Fprint(&out, tooltext.Text("h.78854eacdbfb")+
		tooltext.Text("h.f88cd37d4ef3")+
		tooltext.Text("h.64ed5b694cde")+
		tooltext.Text("h.e2412c278435"))
	fmt.Fprint(&out, tooltext.Text("h.68fc313006e8")+
		tooltext.Text("h.00cc8867ca74"))
	fmt.Fprint(&out, tooltext.Text("h.8d9661b0c176")+
		tooltext.Text("h.5226489b68a3"))

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

// scanRemake 數 remake 這一側的發送點。
//
// 兩種形狀都要數：
//
//	requestSound(SoundXxx)   直接送出
//	return SoundXxx          由挑選器回傳，再由呼叫端送出
//
// ⚠ **只數第一種會產生假零**：`SoundWhistle` 是由 `missileImpactSound` 依武器
// 類別挑出來的，一處 `requestSound` 都沒有，於是報表會說「remake 從沒發過」
// ——而它其實每次用投石索攻擊都會響。這正是本輪在原版那一側踩到的同一種錯
// （掃描面比實際窄），只是換到 remake 這一側。
func scanRemake(root string) map[string]int {
	counts := map[string]int{}
	call := regexp.MustCompile(`(?:requestSound\(|return )(Sound[A-Za-z]+)`)
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
			// `return SoundEvent` 是型別不是事件。
			if match[1] == "SoundEvent" {
				continue
			}
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
