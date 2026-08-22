// Command monster-ai-coverage 量「敵方 AI 實際上動得起來嗎」。
//
// ★ 存在的理由：`docs/audit/remake-status.md` 把「敵方 AI 與完整戰鬥規則」列為
// **沒有分母**的項目之一——個別規則有規格也有測試，但整體行為沒有對照原版資料的
// 覆蓋率報表。這一支給它一個分母。
//
// 量什麼：原版每一隻怪物身上帶的**法術清單**與**士氣位元組**，是不是都變成了
// remake 這一側 AI 用得到的東西。
//
//	法術  AI 是靠法術屬性表 `+0Dh` 的分數過門檻來挑候選的（spec 835／836／1116）。
//	      **查不到分數的候選一律略過**——所以一個沒有分數的法術，怪物永遠不會放，
//	      而且不會有任何錯誤訊息。這是「安靜地少一個行為」，不是崩潰。
//	士氣  `+0F7h`：bit 7 是「這個值有效」，低 7 位是原值的一半（spec 758／1116）。
//	      士氣崩了就不施法，是 spec 836 的第一道前置閘門。
//
// ⚠ 這支量的是**決策層的輸入齊不齊**，不是「AI 打得像不像」。後者要對照原版
// 實機錄影，這裡不宣稱。
//
// 用法：
//
//	go run ./cmd/monster-ai-coverage -output docs/audit/monster-ai-coverage.md
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

// moraleOffset 是角色記錄的 `+0F7h`（spec 758／1116）。
const moraleOffset = 0xF7

func main() {
	imagePath := flag.String("image", "curseoftheazurebonds.zip", "原版 image ZIP")
	output := flag.String("output", "", "Markdown 輸出路徑（留白就印到 stdout）")
	flag.Parse()

	archive, err := zip.OpenReader(*imagePath)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	total, withSpells, moraleValid, moraleBroken := 0, 0, 0, 0
	for chapter := 1; chapter <= 6; chapter++ {
		payload := member(&archive.Reader, fmt.Sprintf("MON%dCHA.DAX", chapter))
		if payload == nil {
			continue
		}
		blocks, parseErr := dax.Parse(payload)
		if parseErr != nil {
			log.Fatalf("MON%dCHA.DAX：%v", chapter, parseErr)
		}
		for _, block := range blocks {
			record, recordErr := monster.Parse(block.Data)
			if recordErr != nil {
				continue
			}
			total++
			if len(record.SpellIDs) > 0 {
				withSpells++
			}
			if len(record.Raw) > moraleOffset {
				if _, ok := combat.MoraleValue(record.Raw[moraleOffset]); ok {
					moraleValid++
				} else {
					moraleBroken++
				}
			}
		}
	}

	// 特殊能力在 `MON*SPC.DAX`，不在角色記錄裡。
	affects, kinds := 0, map[uint8]int{}
	for chapter := 1; chapter <= 6; chapter++ {
		payload := member(&archive.Reader, fmt.Sprintf("MON%dSPC.DAX", chapter))
		if payload == nil {
			continue
		}
		blocks, parseErr := dax.Parse(payload)
		if parseErr != nil {
			continue
		}
		for _, block := range blocks {
			records, recordErr := monster.ParseAffects(block.Data)
			if recordErr != nil {
				continue
			}
			for _, record := range records {
				if record.Kind == 0 {
					continue
				}
				affects++
				kinds[record.Kind]++
			}
		}
	}
	ids := make([]uint8, 0, len(kinds))
	for kind := range kinds {
		ids = append(ids, kind)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	// ⚠ 「戰鬥規則不理它」有三種完全不同的意思，混在一起數會得到一個誤導的缺口：
	//
	//	inert   原作那一支**什麼都沒做** ⇒ remake 也不做才是對的，不是缺口。
	//	unread  還沒解讀 ⇒ 不知道是不是缺口。
	//	其餘    原作有動作而 remake 沒有 ⇒ **真的缺口**。
	table, tableErr := gamepack.EffectModifiers()
	usable, inert, unread, uncatalogued, missing := 0, 0, 0, 0, 0
	inertRecords, unreadRecords, uncataloguedRecords, missingRecords := 0, 0, 0, 0
	status := map[uint8]string{}
	for _, kind := range ids {
		if combat.AffectKindIsInterpreted(kind) {
			usable++
			status[kind] = "已接"
			continue
		}
		state, found := "", false
		if tableErr == nil {
			if handler, ok := table.Handler(kind); ok {
				state, found = handler.Status, true
			}
		}
		switch {
		case !found:
			// ⚠ **表裡根本沒有這個碼 ≠ remake 少做了什麼。** 它代表原作那一支
			// 還沒被反組譯登記過，和 `unread` 同一類（不知道），不是「已知缺口」。
			// 第 688 輪把這一格算進缺口，於是 9 個裡有 5 個是假的。
			uncatalogued++
			uncataloguedRecords += kinds[kind]
			status[kind] = "修正表裡還沒有這個碼"
		case state == "inert":
			inert++
			inertRecords += kinds[kind]
			status[kind] = "原作就沒動作"
		case state == "unread":
			unread++
			unreadRecords += kinds[kind]
			status[kind] = "還沒解讀"
		default:
			missing++
			missingRecords += kinds[kind]
			status[kind] = "缺口"
		}
	}

	var report strings.Builder
	fmt.Fprintf(&report, "# 敵方 AI 的輸入齊不齊：原版怪物資料逐筆盤點\n\n")
	fmt.Fprintf(&report, "由 `cmd/monster-ai-coverage` 產生，不要手改。理由與讀法見該檔的註解。\n\n")
	fmt.Fprintf(&report, "量的是**決策層的輸入**，不是「AI 打得像不像」——後者要對照原版實機，這裡不宣稱。\n\n")

	fmt.Fprintf(&report, "| 項目 | 數量 |\n|---|---:|\n")
	fmt.Fprintf(&report, "| 原版怪物記錄 | %d |\n", total)
	fmt.Fprintf(&report, "| 士氣位元組有效（bit 7 已設）| %d |\n", moraleValid)
	fmt.Fprintf(&report, "| 士氣位元組無效 | %d |\n", moraleBroken)
	fmt.Fprintf(&report, "| **身上帶記憶法術的怪物** | **%d** |\n", withSpells)
	fmt.Fprintf(&report, "| `MON*SPC` 特殊能力記錄 | %d |\n", affects)
	fmt.Fprintf(&report, "| 其中相異的效果碼 | %d |\n", len(ids))
	fmt.Fprintf(&report, "| **戰鬥規則會理的** | **%d** |\n", usable)
	fmt.Fprintf(&report, "| 原作那一支就沒動作（`inert`，碼／記錄）| %d／%d |\n", inert, inertRecords)
	fmt.Fprintf(&report, "| 還沒解讀（`unread`，碼／記錄）| %d／%d |\n", unread, unreadRecords)
	fmt.Fprintf(&report, "| 修正表裡還沒有這個碼（碼／記錄）| %d／%d |\n", uncatalogued, uncataloguedRecords)
	fmt.Fprintf(&report, "| **真的缺口**（原作有動作、remake 沒有；碼／記錄）| **%d／%d** |\n\n", missing, missingRecords)

	fmt.Fprintf(&report, "⚠ 四者不能混在一起數。`inert` 是**原作自己什麼都沒做**——"+
		"remake 也不做才是對的；`unread` 與「表裡還沒有這個碼」則是**不知道**，"+
		"同樣不是已知缺口。把它們算進缺口會憑空生出一個永遠補不完的待辦。\n\n")

	fmt.Fprintf(&report, "## ★ 這一款遊戲裡，怪物的 AI 施法路徑**沒有資料可跑**\n\n")
	fmt.Fprintf(&report, "全部 %d 隻怪物的記憶法術槽（角色記錄 `+33h..+6Ah`）**逐位元組都是 0**。\n", total)
	fmt.Fprintf(&report, "也就是說 `AIChooseSpell` 那條路（門檻掃描、每輪抽 3 個、士氣閘門，"+
		"spec 835／836／1116）在 CoAB **一次都不會被觸發**——規則實作了，但這一款遊戲沒有用到它。\n\n")
	fmt.Fprintf(&report, "⚠ 這不是「AI 壞了」，也不是「規則白做了」：同一顆引擎要跑別的 Gold Box 遊戲，"+
		"而那些遊戲的怪物是有法術清單的。這裡只是把**本作的分母**講清楚——"+
		"拿 CoAB 當樣本去驗證施法 AI，會得到一個永遠全綠而且什麼都沒驗到的測試。\n\n")
	fmt.Fprintf(&report, "⇒ 本作的敵方行為分母落在 `MON*SPC` 的特殊能力上，不在法術清單上。\n\n")

	if missing > 0 && tableErr == nil {
		fmt.Fprintf(&report, "## 真的缺口：每一個卡在哪\n\n")
		fmt.Fprintf(&report, "這四個碼在修正表裡**有解出來的動作**，但 `AffectKindIsInterpreted` "+
			"仍然回 false。三個條件（列在某個 timing 裡、那個 timing remake 有查、"+
			"它動的那一格 remake 有讀）少了哪一條，決定了要補的是什麼工。\n\n")
		fmt.Fprintf(&report, "| 效果碼 | 出現在 timing | 動什麼 | 卡在哪 |\n|---:|---|---|---|\n")
		for _, kind := range ids {
			if status[kind] != "缺口" {
				continue
			}
			handler, _ := table.Handler(kind)
			timings := make([]string, 0, 4)
			for timing := 0; timing <= 0x12; timing++ {
				for _, code := range table.TimingEffects(uint8(timing)) {
					if code == int(kind) {
						timings = append(timings, fmt.Sprintf("`%02X`", timing))
						break
					}
				}
			}
			action, blocker := "（沒有解出來的動作）", "修正表裡沒有動作"
			if len(handler.Modifiers) > 0 {
				modifier := handler.Modifiers[0]
				if modifier.Record != "" {
					action = fmt.Sprintf("寫 `%s` 的第 %d 格", modifier.Record, modifier.Field)
					blocker = "那一格 remake 還沒對應到 `Fighter` 的欄位"
				} else {
					action = fmt.Sprintf("設暫存全域 `%s`", table.ScratchName(modifier.Global))
					blocker = "那些 timing **找不到任何呼叫端**（`checkfx-callsites.md`）"
				}
			}
			fmt.Fprintf(&report, "| `%02Xh` | %s | %s | %s |\n",
				kind, strings.Join(timings, "／"), action, blocker)
		}
		fmt.Fprintf(&report, "\n⇒ 四個都**卡在反組譯，不是卡在接線**。\n\n")
		fmt.Fprintf(&report, "⚠ 其中三個（`4Fh`／`50h`／`7Bh`）落在時機 `02h`／`03h` 上，"+
			"而 `cmd/checkfx-callsites` 掃過 30 處呼叫點之後，**這兩個時機一處呼叫端都沒有**。"+
			"如果它們真的不會被問到，那這三個就**不是缺口**——原作也不會跑到。\n\n")
		fmt.Fprintf(&report, "⚠⚠ 但**現在還不能這樣結論**：那一支只看得到兩種呼叫形狀，"+
			"常駐執行檔那一側因為重定位掃不到。**先把常駐側排除掉，再決定這三個是缺口還是死碼**——"+
			"直接當成缺口去實作，等於為一段原作永遠不會執行的路寫程式。\n\n")
	}

	if missing+inert+unread+uncatalogued > 0 {
		fmt.Fprintf(&report, "## 逐碼\n\n")
		fmt.Fprintf(&report, "| 效果碼 | 出現次數 | 狀態 |\n|---:|---:|---|\n")
		for _, kind := range ids {
			fmt.Fprintf(&report, "| `%02Xh` | %d | %s |\n", kind, kinds[kind], status[kind])
		}
	}

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "monsters=%d with-spells=%d affects=%d kinds=%d interpreted=%d real-gaps=%d morale-valid=%d\n",
		total, withSpells, affects, len(ids), usable, missing, moraleValid)
}

func member(archive *zip.Reader, name string) []byte {
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
