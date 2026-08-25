// Command combatant-name-audit 把原版 `MON*CHA` 每一種戰鬥員的名字，
// 與 game pack 的 `combatant_name_rules` ＋ 兩份語系檔對起來。
//
// ★ 存在的理由：沒宣告的名字**不會有錯誤訊息**——`LocalizeCombatantName`
// 找不到就原樣回傳，玩家在戰鬥選單看到的是 `THRI-KREEN`。這一支把
// 「還有幾種是英文」變成一個可以引用的數字。
//
// 用法：
//
//	go run ./cmd/combatant-name-audit -output docs/audit/combatant-names.md
package main

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

// nameRule 是 `00-core.json` 的 `combatant_name_rules`：把原始名字對到一個
// message id，兩份語系檔再各自翻譯那個 id。
type nameRule struct {
	ID        string `json:"id"`
	Source    string `json:"source"`
	MessageID string `json:"message_id"`
}

type localeFile struct {
	Locales map[string]map[string]string `json:"locales"`
}

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", tooltext.Text("h.90f916a91fdc"))
	core := flag.String("core", "gamepack/pack/00-core.json", "game pack core JSON")
	zhFile := flag.String("zh", "gamepack/pack/20-locale.zh-TW.json", tooltext.Text("h.01d3b4025afa"))
	enFile := flag.String("en", "gamepack/pack/20-locale.en.json", tooltext.Text("h.88ad2e9a2d36"))
	output := flag.String("output", "", tooltext.Text("h.78eb014c7900"))
	flag.Parse()

	roster, err := CollectRoster(*image)
	if err != nil {
		log.Fatal(err)
	}
	rules, err := LoadNameRules(*core)
	if err != nil {
		log.Fatal(err)
	}
	zh, err := LoadLocale(*zhFile, "zh-TW")
	if err != nil {
		log.Fatal(err)
	}
	en, err := LoadLocale(*enFile, "en")
	if err != nil {
		log.Fatal(err)
	}

	bySource := make(map[string]nameRule, len(rules))
	for _, rule := range rules {
		bySource[rule.Source] = rule
	}

	names := make([]string, 0, len(roster))
	for name := range roster {
		names = append(names, name)
	}
	sort.Strings(names)

	undeclared, untranslated := 0, 0
	var report strings.Builder
	fmt.Fprint(&report, tooltext.Format("h.d3cdca52641f"))
	fmt.Fprint(&report, tooltext.Format("h.8ebf9c663851"))
	var body strings.Builder
	fmt.Fprint(&body, tooltext.Format("h.a46881497473"))
	for _, name := range names {
		rule, declared := bySource[name]
		chinese, english := "—", "—"
		if !declared {
			undeclared++
		} else {
			chinese, english = zh[rule.MessageID], en[rule.MessageID]
			if chinese == "" || english == "" {
				untranslated++
				if chinese == "" {
					chinese = "—"
				}
				if english == "" {
					english = "—"
				}
			}
		}
		fmt.Fprintf(&body, "| `%s` | %s | %s | %s |\n",
			name, strings.Join(roster[name], " "), chinese, english)
	}

	fmt.Fprint(&report, tooltext.Format("h.13c83a8a875e"))
	fmt.Fprint(&report, tooltext.Format("h.6389882aa39a", len(names)))
	fmt.Fprint(&report, tooltext.Format("h.7ab35f1c602c", undeclared))
	fmt.Fprint(&report, tooltext.Format("h.8cc53c8b7eca", untranslated))
	report.WriteString(body.String())

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "combatants=%d undeclared=%d untranslated=%d\n",
		len(names), undeclared, untranslated)
}

// CollectRoster 讀出六章 `MON*CHA.DAX` 的每一種戰鬥員，回傳名字 → 出現處。
func CollectRoster(image string) (map[string][]string, error) {
	archive, err := zip.OpenReader(image)
	if err != nil {
		return nil, err
	}
	defer archive.Close()
	roster := map[string][]string{}
	for chapter := 1; chapter <= 6; chapter++ {
		payload := member(archive, fmt.Sprintf("MON%dCHA.DAX", chapter))
		if payload == nil {
			continue
		}
		parsed, parseErr := dax.Parse(payload)
		if parseErr != nil {
			return nil, fmt.Errorf("MON%dCHA.DAX: %w", chapter, parseErr)
		}
		for _, block := range parsed {
			record, recordErr := monster.Parse(block.Data)
			if recordErr != nil {
				continue
			}
			roster[record.Name] = append(roster[record.Name],
				fmt.Sprintf("%d:%d", chapter, block.Entry.ID))
		}
	}
	return roster, nil
}

func LoadNameRules(path string) ([]nameRule, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var core struct {
		Rules []nameRule `json:"combatant_name_rules"`
	}
	if err := json.Unmarshal(payload, &core); err != nil {
		return nil, err
	}
	return core.Rules, nil
}

func LoadLocale(path, language string) (map[string]string, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file localeFile
	if err := json.Unmarshal(payload, &file); err != nil {
		return nil, err
	}
	table, ok := file.Locales[language]
	if !ok {
		return nil, tooltext.Errorf("h.639ff7bf1d3a", path, language)
	}
	return table, nil
}

func member(archive *zip.ReadCloser, name string) []byte {
	for _, file := range archive.File {
		if !strings.EqualFold(file.Name, name) {
			continue
		}
		handle, err := file.Open()
		if err != nil {
			log.Fatal(err)
		}
		defer handle.Close()
		payload, readErr := io.ReadAll(handle)
		if readErr != nil {
			log.Fatal(readErr)
		}
		return payload
	}
	return nil
}
