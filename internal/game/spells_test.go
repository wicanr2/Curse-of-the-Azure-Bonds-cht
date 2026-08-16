package game

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
)

// loadShippedCatalog 讀出貨用的繁中字表，不是測試 fixture——這條測試要驗的
// 正是實際出貨的那份鍵。
func loadShippedCatalog() (locale.Catalog, error) {
	data, err := os.ReadFile(filepath.Join("..", "..", "assets", "locale", "zh-TW.json"))
	if err != nil {
		return locale.Catalog{}, err
	}
	return locale.Load(data)
}

// parseSpellKeyID 取 `spell_<職業>_<編號>` 結尾的編號。
func parseSpellKeyID(parts []string) (int, error) {
	if len(parts) < 3 {
		return 0, strconv.ErrSyntax
	}
	return strconv.Atoi(parts[len(parts)-1])
}

// 法術槽裡存的是**原作的全域編號**，所以名字只能由編號決定。
//
// ⚠ 這裡曾經是「編號 ＋ 持有者職業」一起查：法師的 1..13 當成「法師第幾支」、
// 牧師的 1..8 當成全域編號。兩套編號在同一個 locale 前綴下會互相蓋掉
// （法師第 7 支是魔法飛彈，全域第 7 支是防護善良），而雙職角色兩邊都有。
func TestCampSpellLabelComesFromTheOriginalTableNotTheHolder(t *testing.T) {
	catalog := testCatalog()
	catalog.Strings["spell_cleric_7"] = "防護善良"
	catalog.Strings["spell_magic_user_15"] = "魔法飛彈"

	if got := campSpellLabel(catalog, ProtectionFromGoodSpellID); got != "防護善良" {
		t.Fatalf("全域編號 0x%02X 的名字是 %q", ProtectionFromGoodSpellID, got)
	}
	if got := campSpellLabel(catalog, MagicMissileSpellID); got != "魔法飛彈" {
		t.Fatalf("全域編號 0x%02X 的名字是 %q", MagicMissileSpellID, got)
	}
	// 沒有譯名時保留十六進位編號：匯入的存檔可能帶著 remake 還沒翻的編號，
	// 退回一個看起來像法術名的字串會讓那件事查不出來。
	if got := campSpellLabel(catalog, 0x2F); !strings.Contains(got, "0x2F") {
		t.Fatalf("沒有譯名的編號應該保留編號，got %q", got)
	}
	// 表外的編號同理。
	if got := campSpellLabel(catalog, 0xFE); !strings.Contains(got, "0xFE") {
		t.Fatalf("表外編號應該保留編號，got %q", got)
	}
}

// locale 的每一個法術名鍵都必須對得上原作表：職業與編號都要相符。
// 這條擋的是「翻譯照著另一套編號寫」——譯名本身沒有錯，錯的是掛的編號。
func TestSpellNameKeysFollowTheOriginalTable(t *testing.T) {
	catalog, err := loadShippedCatalog()
	if err != nil {
		t.Skipf("shipped catalog is unavailable: %v", err)
	}
	checked := 0
	for key := range catalog.Strings {
		if !strings.HasPrefix(key, "spell_") {
			continue
		}
		parts := strings.Split(key, "_")
		spellID, err := parseSpellKeyID(parts)
		if err != nil {
			continue // spell_unknown 之類的非編號鍵
		}
		want, ok := spellMessageID(uint8(spellID))
		if !ok {
			t.Fatalf("locale 有 %q，但編號 %d 在原作表裡是占位項或不存在", key, spellID)
		}
		if want != key {
			t.Fatalf("locale 的 %q 依原作表應該是 %q（同一支法術只屬於一個施法職業）",
				key, want)
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("一個法術名鍵都沒檢查到，這條測試等於沒驗")
	}
	t.Logf("檢查了 %d 個法術名鍵", checked)
}
