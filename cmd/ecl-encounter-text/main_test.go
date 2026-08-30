package main

import (
	"archive/zip"
	"fmt"
	"os"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/golden-box-remake-engine/dax"
)

const imagePath = "../../curseoftheazurebonds.zip"

// `29h ENCOUNTER MENU` 的旁白，玩家真的會看到的那一句必須接得上譯文。
//
// ★ 這條擋的是一個**別處量不到**的缺口：`cmd/ecl-text-coverage` 的分母裡沒有這個
// opcode，所以那份報告「未接上 0 群」的時候，這 20 處旁白連被數到都沒有。第一次
// 跑出來是 9 句演得到卻沒有規則——玩家會看到整句英文。
//
// ⚠ 一句一句比對，不是把三句合起來比：`MatchText` 收到的是 joined 字串，三句
// 一起餵進去會讓第一句的規則把另外兩句蓋掉，缺口就消失了。
func TestShownEncounterPromptsAreLocalized(t *testing.T) {
	if _, err := os.Stat(imagePath); err != nil {
		t.Skip("找不到遊戲 image，跳過")
	}
	pack, err := gamepack.Default()
	if err != nil {
		t.Fatal(err)
	}
	archive, err := zip.OpenReader(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()

	prompts := []prompt(nil)
	for member := 1; member <= 6; member++ {
		payload := memberPayload(archive, fmt.Sprintf("ECL%d.DAX", member))
		if payload == nil {
			t.Fatalf("image 裡沒有 ECL%d.DAX", member)
		}
		blocks, parseErr := dax.Parse(payload)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		for _, block := range blocks {
			found, scanErr := scanBlock(pack, "zh-TW", member, block.Entry.ID, block.Data)
			if scanErr != nil {
				t.Fatalf("ECL%d/0x%02X: %v", member, block.Entry.ID, scanErr)
			}
			prompts = append(prompts, found...)
		}
	}

	shown, hidden := 0, 0
	for _, item := range prompts {
		if !item.shown {
			hidden++
			continue
		}
		shown++
		if item.ruleID == "" {
			t.Errorf("`ECL%d/0x%02X` `%#04x` 的旁白沒有譯文規則：%q",
				item.member, item.block, item.offset, item.text)
			continue
		}
		if item.message == "" {
			t.Errorf("`ECL%d/0x%02X` `%#04x` 的規則 %s 在 zh-TW 沒有訊息",
				item.member, item.block, item.offset, item.ruleID)
		}
	}
	// ⚠ 非空閘門：掃描壞掉會變成「一句都沒有 ⇒ 一句都沒缺」。
	if shown < 20 {
		t.Errorf("只掃到 %d 句演得到的旁白，應該至少 20 句", shown)
	}
	// remake 目前只演第一句，另外兩句演不到。這個數字塌成 0 代表旁白的解讀壞了，
	// 不代表還原度補上了——真的補上時要連同這條一起改。
	if hidden < 20 {
		t.Errorf("只掃到 %d 句 remake 演不到的旁白，應該至少 20 句", hidden)
	}
	t.Logf("旁白 %d 句：演得到 %d、演不到 %d", len(prompts), shown, hidden)
}
