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
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
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

// deadTimings 是**分派表裡有效果碼、但整個執行檔沒有任何呼叫端**的時機。
//
// ★ 出處：`cmd/checkfx-callsites` 掃過三個方向——跨 overlay 的 far call、
// overlay-23 內部的 `E8` 近呼叫、常駐執行檔（有正對照：常駐側確實會用 far call
// 叫 overlay，只是從不叫 `CHECKFX`）——30 處呼叫點的時機全部是立即數，沒有一處
// 來自變數。只落在這些時機底下的效果碼，原作**永遠不會執行**。
//
// ⚠ 所以它們**不是缺口**。remake 不實作它們是對的；算成缺口等於為死碼寫程式。
var deadTimings = map[uint8]bool{0x02: true, 0x03: true}

func main() {
	imagePath := flag.String("image", "curseoftheazurebonds.zip", tooltext.Text("h.79f855c8b433"))
	output := flag.String("output", "", tooltext.Text("h.78eb014c7900"))
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
	usable, inert, unread, uncatalogued, dead, missing := 0, 0, 0, 0, 0, 0
	inertRecords, unreadRecords, uncataloguedRecords, deadRecords, missingRecords := 0, 0, 0, 0, 0
	// 特殊攻擊走 CALLEFFECT 分派（spec 1202），不在 CHECKFX 修正表裡——
	// 用修正表查它們會得到「不知道」，而它們的原作語意其實已經讀完。
	// wiredBy 非空＝remake 已用該機制接上；空＝原作有動作、remake 沒有。
	specialAttacks := map[uint8]struct{ spec, wiredBy string }{
		0x56: {spec: tooltext.Text("monster_ai_coverage.special_56"), wiredBy: "SpecialAttackRules"},
		0x58: {spec: tooltext.Text("monster_ai_coverage.special_58")},
		0x5A: {spec: tooltext.Text("monster_ai_coverage.special_5a"), wiredBy: "SpecialAttackRules"},
		0x7E: {spec: tooltext.Text("monster_ai_coverage.special_7e"), wiredBy: "SpecialAttackRules"},
		0x80: {spec: tooltext.Text("monster_ai_coverage.special_80"), wiredBy: "SpecialAttackRules"},
		0x83: {spec: tooltext.Text("monster_ai_coverage.special_83"), wiredBy: "SpecialAttackRules"},
		0x84: {spec: tooltext.Text("monster_ai_coverage.special_84"), wiredBy: "MonsterSpellRules"},
	}
	status := map[uint8]string{}
	for _, kind := range ids {
		if combat.AffectKindIsInterpreted(kind) {
			usable++
			status[kind] = tooltext.Text("h.f8c349bcf027")
			continue
		}
		if attack, ok := specialAttacks[kind]; ok {
			if attack.wiredBy != "" {
				usable++
				status[kind] = tooltext.Format("monster_ai_coverage.special_attack_wired", attack.wiredBy)
			} else {
				missing++
				missingRecords += kinds[kind]
				status[kind] = tooltext.Format("monster_ai_coverage.special_attack_missing", attack.spec)
			}
			continue
		}
		state, found := "", false
		if tableErr == nil {
			if handler, ok := table.Handler(kind); ok {
				state, found = handler.Status, true
			}
		}
		switch {
		case tableErr == nil && onlyDeadTimings(table, kind):
			dead++
			deadRecords += kinds[kind]
			status[kind] = tooltext.Text("h.496420dcda95")
		case !found:
			// ⚠ **表裡根本沒有這個碼 ≠ remake 少做了什麼。** 它代表原作那一支
			// 還沒被反組譯登記過，和 `unread` 同一類（不知道），不是「已知缺口」。
			// 第 688 輪把這一格算進缺口，於是 9 個裡有 5 個是假的。
			uncatalogued++
			uncataloguedRecords += kinds[kind]
			status[kind] = tooltext.Text("h.31f728323bcd")
		case state == "inert":
			inert++
			inertRecords += kinds[kind]
			status[kind] = tooltext.Text("h.edafa9e8419a")
		case state == "unread":
			unread++
			unreadRecords += kinds[kind]
			status[kind] = tooltext.Text("h.9d429bdc87c5")
		default:
			missing++
			missingRecords += kinds[kind]
			status[kind] = tooltext.Text("h.245cdcd5fad2")
		}
	}

	var report strings.Builder
	fmt.Fprint(&report, tooltext.Format("h.f6ae980276c6"))
	fmt.Fprint(&report, tooltext.Format("h.9cfbb2304e50"))
	fmt.Fprint(&report, tooltext.Format("h.69c6530f87c0"))

	fmt.Fprint(&report, tooltext.Format("h.74b2b2b87498"))
	fmt.Fprint(&report, tooltext.Format("h.331975a22dda", total))
	fmt.Fprint(&report, tooltext.Format("h.db9ec0bd2dcc", moraleValid))
	fmt.Fprint(&report, tooltext.Format("h.c6c414bf7fd8", moraleBroken))
	fmt.Fprint(&report, tooltext.Format("h.6a3327330d4d", withSpells))
	fmt.Fprint(&report, tooltext.Format("h.79222bd6baed", affects))
	fmt.Fprint(&report, tooltext.Format("h.144ee8386c22", len(ids)))
	fmt.Fprint(&report, tooltext.Format("h.d831ff3a546d", usable))
	fmt.Fprint(&report, tooltext.Format("h.b83576d9cf8f", inert, inertRecords))
	fmt.Fprint(&report, tooltext.Format("h.25f3644c9c1e", unread, unreadRecords))
	fmt.Fprint(&report, tooltext.Format("h.a5085942f436", uncatalogued, uncataloguedRecords))
	fmt.Fprint(&report, tooltext.Format("h.dff2889b9a93", dead, deadRecords))
	fmt.Fprint(&report, tooltext.Format("h.380a095840a3", missing, missingRecords))

	fmt.Fprint(&report, tooltext.Text("h.489c17d501d5")+
		tooltext.Text("h.bad98c8bc29f")+
		tooltext.Text("h.e82fcfbb066d"))

	fmt.Fprint(&report, tooltext.Format("h.2595539ad018"))
	fmt.Fprint(&report, tooltext.Format("h.ba03b02dea46", total))
	fmt.Fprint(&report, tooltext.Text("h.14ca182acf81")+
		tooltext.Text("h.2d045b11e6c5"))
	fmt.Fprint(&report, tooltext.Text("h.8ab6d779800d")+
		tooltext.Text("h.26c279e7b3cb")+
		tooltext.Text("h.6c4c9aa11a81"))
	fmt.Fprint(&report, tooltext.Format("h.dc3d349f5b83"))

	if missing > 0 && tableErr == nil {
		fmt.Fprint(&report, tooltext.Format("h.31731a0a46ba"))
		fmt.Fprint(&report, tooltext.Text("h.48dd82e8d59d")+
			tooltext.Text("h.2c736614739a")+
			tooltext.Text("h.365c359598fa"))
		fmt.Fprint(&report, tooltext.Format("h.8e42b2dc44dd"))
		for _, kind := range ids {
			if status[kind] != tooltext.Text("h.245cdcd5fad2") {
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
			action, blocker := tooltext.Text("h.d00fa7b349a1"), tooltext.Text("h.7af3e2e2aecf")
			if len(handler.Modifiers) > 0 {
				modifier := handler.Modifiers[0]
				if modifier.Record != "" {
					action = tooltext.Format("h.27974c2ad2d1", modifier.Record, modifier.Field)
					blocker = tooltext.Text("h.7a1d0120174e")
				} else {
					action = tooltext.Format("h.46ef0e82bf0b", table.ScratchName(modifier.Global))
					blocker = tooltext.Text("h.b0674f1ab885")
				}
			}
			fmt.Fprintf(&report, "| `%02Xh` | %s | %s | %s |\n",
				kind, strings.Join(timings, "／"), action, blocker)
		}
		fmt.Fprint(&report, tooltext.Format("h.8afa668227b3"))
		fmt.Fprint(&report, tooltext.Text("h.048dbbd2fafb")+
			tooltext.Text("h.a1942e6cd228")+
			tooltext.Text("h.a8bdf44b8ca6"))
		fmt.Fprint(&report, tooltext.Text("h.2ac051975dfe")+
			tooltext.Text("h.b0b67acbdc2c")+
			tooltext.Text("h.124f12260a1f"))
	}

	if missing+inert+unread+uncatalogued+dead > 0 {
		fmt.Fprint(&report, tooltext.Format("h.793cd407235d"))
		fmt.Fprint(&report, tooltext.Format("h.f72314607a05"))
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

// onlyDeadTimings 回答「這個效果碼是不是**只**出現在沒有呼叫端的時機底下」。
// 一個都沒出現在任何 timing 清單裡的碼不算（那是另一種未知）。
func onlyDeadTimings(table *gamepack.EffectModifierTable, kind uint8) bool {
	listed := false
	for timing := 0; timing <= 0x16; timing++ {
		for _, code := range table.TimingEffects(uint8(timing)) {
			if code != int(kind) {
				continue
			}
			listed = true
			if !deadTimings[uint8(timing)] {
				return false
			}
		}
	}
	return listed
}
