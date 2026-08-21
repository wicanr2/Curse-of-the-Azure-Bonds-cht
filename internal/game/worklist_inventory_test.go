package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// `WORKLIST.md` 的「還開著的工作」是路由文件：後面每一輪照它決定做什麼。
// 它的數字是**手抄**自各支稽核工具的，工具重跑之後沒人回頭改，就會出現
// 「工作早就做完了、清單還寫著開著」——比沒有清單更糟，因為它會把人帶錯方向。
//
// ★ 這一條把三個最容易過期的數字釘住：只要工具的產出或資料變了而 WORKLIST 沒跟，
// 測試會指出是哪一個。
func TestWorklistInventoryNumbersAreCurrent(t *testing.T) {
	root := filepath.Join("..", "..")
	worklist := readRepoFile(t, filepath.Join(root, "WORKLIST.md"))

	// 1) ECL 副作用：分母與 partial 指令數取自產生的稽核檔。
	effects := readRepoFile(t, filepath.Join(root, "docs", "audit", "ecl-effect-coverage.md"))
	partial := auditNumber(t, effects, `\|\s*`+"`partial`"+`\s*\|\s*([\d,]+)\s*\|`)
	requireWorklistMentions(t, worklist, "partial", "ECL 副作用的 `partial` 指令數", partial)

	// 2) 存檔：角色記錄還有幾個位元組是 unknown。
	// ⚠ 正本是 `dos-save-field-coverage.md`（`cmd/save-field-coverage` 的用法
	// 註解與 spec 1115 都指這個名字）。曾經同時存在一份 `save-field-coverage.md`
	// 的副本，兩份各自過期。
	save := readRepoFile(t, filepath.Join(root, "docs", "audit", "dos-save-field-coverage.md"))
	unknown := auditNumber(t, save, `\|\s*`+"`unknown`"+`\s*\|\s*([\d,]+)\s*\|`)
	requireWorklistMentions(t, worklist, "unknown", "角色記錄的 `unknown` 位元組數", unknown)

	// 3) 手札：locale 宣告幾則、其中幾則有 producer（`journal_message_ids`）。
	declared, wired := journalProducerCoverage(t, root)
	if declared == 0 {
		t.Fatal("locale 裡沒有 journal 條目，量測本身壞了")
	}
	if missing := declared - wired; missing > 0 {
		requireWorklistMentions(t, worklist, "手札", "沒有 producer 的手札則數",
			fmt.Sprintf("%d", missing))
	} else if strings.Contains(worklist, "則還沒有觸發來源") ||
		strings.Contains(worklist, "則尚缺 producer 接線") {
		t.Errorf("手札 %d 則全部都有 producer，WORKLIST 卻還寫著「還沒有觸發來源」", declared)
	}
}

func readRepoFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("讀不到 %s：%v", path, err)
	}
	return string(data)
}

// auditNumber 從產生的稽核 markdown 抓一個數字。抓不到就讓測試紅——**抓不到不能
// 當成 0 跳過**，那會讓這道閘門安靜地失效。
func auditNumber(t *testing.T, document, pattern string) string {
	t.Helper()
	match := regexp.MustCompile(pattern).FindStringSubmatch(document)
	if match == nil {
		t.Fatalf("稽核檔裡找不到 %q 這個欄位，閘門的解析要跟著改", pattern)
	}
	return strings.ReplaceAll(match[1], ",", "")
}

// requireWorklistMentions 要求那個數字出現在**關鍵字後面一小段窗口內**。
//
// ⚠ 不能只比對「整份文件裡有沒有這個數字」：`24` 這種小數字在一份長清單裡一定
// 找得到。也不能只比對「同一行」——WORKLIST 的一列就是一整行，同一行提到
// `unknown` 與別處的「28」就會誤過。**踩過一次：手札那一列讓存檔的檢查安靜地變綠。**
func requireWorklistMentions(t *testing.T, worklist, keyword, label, number string) {
	t.Helper()
	// 逗號分位與不分位兩種寫法都算數。
	variants := []string{number}
	if len(number) > 3 {
		variants = append(variants, number[:len(number)-3]+","+number[len(number)-3:])
	}
	occurrences := 0
	for offset := 0; ; {
		index := strings.Index(worklist[offset:], keyword)
		if index < 0 {
			break
		}
		occurrences++
		start := offset + index
		end := start + len(keyword) + worklistNumberWindow
		if end > len(worklist) {
			end = len(worklist)
		}
		window := worklist[start:end]
		for _, variant := range variants {
			if strings.Contains(window, variant) {
				return
			}
		}
		offset = start + len(keyword)
	}
	if occurrences == 0 {
		t.Errorf("WORKLIST.md 裡沒有一處提到 %q，%s 的閘門失去著力點", keyword, label)
		return
	}
	t.Errorf("WORKLIST.md 提到 %q 的 %d 處後面都沒有%s（現在是 %s）——工具重跑過而清單沒跟",
		keyword, occurrences, label, number)
}

// worklistNumberWindow 是「關鍵字後面幾個位元組內要看到那個數字」。
// 夠短才擋得住同一列裡別的數字，夠長才容得下「還有 **24**」這種排版。
const worklistNumberWindow = 40

// journalProducerCoverage 回傳「locale 宣告的手札則數」與「其中有 ECL producer 的
// 則數」。producer ＝ 內容規則上的 `journal_message_ids`。
func journalProducerCoverage(t *testing.T, root string) (int, int) {
	t.Helper()
	declared := map[string]bool{}
	collectJournalIDs(decodePackFile(t, filepath.Join(root, "gamepack", "pack", "20-locale.zh-TW.json")), declared)
	wired := map[string]bool{}
	for _, name := range []string{"10-content.json", "00-core.json"} {
		collectJournalProducers(decodePackFile(t, filepath.Join(root, "gamepack", "pack", name)), wired)
	}
	count := 0
	for id := range declared {
		if wired[id] {
			count++
		}
	}
	return len(declared), count
}

func decodePackFile(t *testing.T, path string) any {
	t.Helper()
	var value any
	if err := json.Unmarshal([]byte(readRepoFile(t, path)), &value); err != nil {
		t.Fatalf("%s 不是合法 JSON：%v", path, err)
	}
	return value
}

func collectJournalIDs(node any, out map[string]bool) {
	switch value := node.(type) {
	case map[string]any:
		for key, child := range value {
			if text, ok := child.(string); ok && strings.HasPrefix(key, "journal.") && text != "" {
				out[key] = true
				continue
			}
			collectJournalIDs(child, out)
		}
	case []any:
		for _, child := range value {
			collectJournalIDs(child, out)
		}
	}
}

func collectJournalProducers(node any, out map[string]bool) {
	switch value := node.(type) {
	case map[string]any:
		for key, child := range value {
			if key == "journal_message_ids" {
				if list, ok := child.([]any); ok {
					for _, entry := range list {
						if id, ok := entry.(string); ok {
							out[id] = true
						}
					}
				}
				continue
			}
			collectJournalProducers(child, out)
		}
	case []any:
		for _, child := range value {
			collectJournalProducers(child, out)
		}
	}
}
