// Command cht-proofread-audit 給「全文校對」一個分母。
//
// ★ 存在的理由：`remake-status` 把「全文校對」列為沒有數字的項目。文字**覆蓋率**
// 早就有了（`ecl-text-coverage.md`：1,022 頁裡 1,002 頁接上規則、0 頁沒接），
// 但那回答的是「有沒有翻」，不是「翻得對不對、一不一致」。這一支量後者。
//
// 三類機械查得到的問題（不做語感判斷，那不是靜態工具的事）：
//
//	未翻    中譯裡整段都是拉丁字母、一個漢字都沒有。
//	不一致  **同一句英文**在不同地方被翻成不同的中文。
//	標點    中文句子裡混進半形標點（`,` `;` `?` `!` `:`），與全形混用。
//
// ⚠ 三類都會有**合理的例外**：專有名詞可以保留原文、同一句英文在不同語境本來就
// 該有不同譯法、數字與代號旁邊的半形標點是對的。所以這一支報的是**待看清單**，
// 不是缺陷數——把它當缺陷數會逼出「為了讓數字歸零而改壞翻譯」。
//
// 用法：
//
//	go run ./cmd/cht-proofread-audit -output docs/audit/cht-proofread.md
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"log"
	"os"
	"sort"
	"strings"
	"unicode"
)

func main() {
	zhPath := flag.String("zh", "gamepack/pack/20-locale.zh-TW.json", tooltext.Text("h.18a8248f0745"))
	enPath := flag.String("en", "gamepack/pack/20-locale.en.json", tooltext.Text("h.88ad2e9a2d36"))
	output := flag.String("output", "", tooltext.Text("h.78eb014c7900"))
	limit := flag.Int("limit", 20, tooltext.Text("h.996b58182f3c"))
	flag.Parse()

	zh := loadLocale(*zhPath)
	en := loadLocale(*enPath)

	keys := make([]string, 0, len(zh))
	for key := range zh {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	untranslated := make([]string, 0, 16)
	punctuation := make([]string, 0, 16)
	// ⚠ 正對照。 「半形標點 0 筆」有兩種可能：真的沒有，或者檢查根本沒生效。
	// 數一下有多少條目帶**全形**標點——那個數字大，才證明這批文字確實有標點、
	// 而檢查看得到它們。沒有這個對照，一個壞掉的檢查會回報成滿分。
	fullWidth := 0
	// 同一句英文 → 有哪些中譯。
	byEnglish := map[string]map[string]bool{}
	for _, key := range keys {
		text := zh[key]
		if strings.TrimSpace(text) == "" {
			continue
		}
		if !hasHan(text) && hasLatinLetter(text) {
			untranslated = append(untranslated, fmt.Sprintf("`%s` — %s", key, trim(text)))
		}
		if strings.ContainsAny(text, "，。！？；：、") {
			fullWidth++
		}
		if mark, found := halfWidthInHan(text); found {
			punctuation = append(punctuation, tooltext.Format("h.9e3ddaedd64a", key, trim(text), mark))
		}
		source, ok := en[key]
		if !ok || strings.TrimSpace(source) == "" || !hasHan(text) {
			continue
		}
		normalised := strings.ToUpper(strings.TrimSpace(source))
		if len(normalised) < 8 {
			// 太短的英文（`UP`、`EXIT`）在不同語境本來就會有不同譯法，
			// 收進來只會製造雜訊。
			continue
		}
		if byEnglish[normalised] == nil {
			byEnglish[normalised] = map[string]bool{}
		}
		byEnglish[normalised][text] = true
	}

	inconsistent := make([]string, 0, 16)
	for source, translations := range byEnglish {
		if len(translations) < 2 {
			continue
		}
		options := make([]string, 0, len(translations))
		for text := range translations {
			options = append(options, trim(text))
		}
		sort.Strings(options)
		inconsistent = append(inconsistent, fmt.Sprintf("`%s` → %s", trim(source), strings.Join(options, "／")))
	}
	sort.Strings(inconsistent)

	var report strings.Builder
	fmt.Fprint(&report, tooltext.Format("h.fc444386e10e"))
	fmt.Fprint(&report, tooltext.Format("h.d4232137f59c"))
	fmt.Fprint(&report, tooltext.Text("h.5e1d82de481c")+
		tooltext.Text("h.9bb815523620")+
		tooltext.Text("h.d96ad9ceb160"))
	fmt.Fprint(&report, tooltext.Format("h.7cd6773fb760"))

	fmt.Fprint(&report, tooltext.Format("h.74b2b2b87498"))
	fmt.Fprint(&report, tooltext.Format("h.b0b238f92a40", len(zh)))
	fmt.Fprint(&report, tooltext.Format("h.369657f07944", len(untranslated)))
	fmt.Fprint(&report, tooltext.Format("h.3b7dd4ad3e15", len(inconsistent)))
	fmt.Fprint(&report, tooltext.Format("h.ef81b587fc92", len(punctuation)))
	fmt.Fprint(&report, tooltext.Format("h.18efdde42412", fullWidth))
	if len(punctuation) == 0 && fullWidth > len(zh)/4 {
		fmt.Fprintf(&report, tooltext.Text("h.9455efc07579")+
			tooltext.Text("h.2a77e8ea84ce"), fullWidth, len(zh))
	}

	section(&report, tooltext.Text("h.c03e2f87b88f"), untranslated, *limit)
	section(&report, tooltext.Text("h.daada28c10a7"), inconsistent, *limit)
	section(&report, tooltext.Text("h.f3a2b18be509"), punctuation, *limit)

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "entries=%d untranslated=%d inconsistent=%d punctuation=%d full-width-control=%d\n",
		len(zh), len(untranslated), len(inconsistent), len(punctuation), fullWidth)
}

func section(report *strings.Builder, title string, items []string, limit int) {
	if len(items) == 0 {
		fmt.Fprint(report, tooltext.Format("h.6a914315b1be", title))
		return
	}
	fmt.Fprint(report, tooltext.Format("h.cae965c94fb5", title, len(items)))
	for index, item := range items {
		if index >= limit {
			fmt.Fprint(report, tooltext.Format("h.da2dee540ba3", len(items)-limit))
			break
		}
		fmt.Fprintf(report, "- %s\n", item)
	}
	fmt.Fprintf(report, "\n")
}

func loadLocale(path string) map[string]string {
	raw, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	var file struct {
		Locales map[string]map[string]string `json:"locales"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		log.Fatal(err)
	}
	for _, entries := range file.Locales {
		return entries
	}
	return map[string]string{}
}

func hasHan(text string) bool {
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

func hasLatinLetter(text string) bool {
	for _, r := range text {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			return true
		}
	}
	return false
}

// halfWidthInHan 找「中文字旁邊的半形標點」。只有前後其中一邊是漢字才算——
// 這樣 `HP: 12` 那種純數字代號不會被抓進來。
func halfWidthInHan(text string) (string, bool) {
	runes := []rune(text)
	for index, r := range runes {
		if !strings.ContainsRune(",;?!:", r) {
			continue
		}
		before := index > 0 && unicode.Is(unicode.Han, runes[index-1])
		after := index+1 < len(runes) && unicode.Is(unicode.Han, runes[index+1])
		if before || after {
			return string(r), true
		}
	}
	return "", false
}

func trim(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	if len([]rune(text)) > 46 {
		return string([]rune(text)[:46]) + "…"
	}
	return text
}
