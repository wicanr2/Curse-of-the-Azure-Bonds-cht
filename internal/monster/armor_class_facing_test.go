package monster

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

// dexterityDefenceAdjustment 是原作 `overlay-24:117Ah`（spec 694 那張表的
// 未取負號版本）：敏捷高就是正值，代表防禦變好。
//
// ⚠ 這裡是**測試本地**的表。它是規則，將來 AC 刻度統一之後要搬進正式程式碼；
// 現在還沒有生產端會用它，先不放進去當死碼。
func dexterityDefenceAdjustment(dexterity int) int {
	switch {
	case dexterity <= 3:
		return -4
	case dexterity == 4:
		return -3
	case dexterity == 5:
		return -2
	case dexterity == 6:
		return -1
	case dexterity <= 14:
		return 0
	case dexterity == 15:
		return 1
	case dexterity == 16:
		return 2
	case dexterity == 17:
		return 3
	case dexterity <= 20:
		return 4
	case dexterity <= 23:
		return 5
	default:
		return 6
	}
}

func originalMonsterRecords(t *testing.T) []([]byte) {
	t.Helper()
	const imagePath = "../../curseoftheazurebonds.zip"
	if _, err := os.Stat(imagePath); err != nil {
		t.Skip("找不到遊戲 image，跳過")
	}
	archive, err := zip.OpenReader(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	records := make([][]byte, 0, 96)
	for chapter := 1; chapter <= 6; chapter++ {
		member := fmt.Sprintf("MON%dCHA.DAX", chapter)
		for _, file := range archive.File {
			if file.Name != member {
				continue
			}
			reader, err := file.Open()
			if err != nil {
				t.Fatal(err)
			}
			payload, err := io.ReadAll(reader)
			reader.Close()
			if err != nil {
				t.Fatal(err)
			}
			blocks, err := dax.Parse(payload)
			if err != nil {
				t.Fatal(err)
			}
			for _, block := range blocks {
				if len(block.Data) >= RecordSize {
					records = append(records, block.Data)
				}
			}
		}
	}
	if len(records) == 0 {
		t.Fatal("image 裡讀不到任何 MON*CHA 記錄")
	}
	return records
}

// 第二個護甲欄位的算式（spec 1000 §七）拿原作資料驗一次：
//
//	+19Bh = +19Ah − 敏捷防禦調整 − 盾牌那一槽 − 2
//
// 盾牌那一槽的加值來自 BSS 裡的物品類別表，靜態讀不到，所以這條測試量的是
// **殘差**：81 筆記錄裡 78 筆殘差為 0，剩下三筆分別是 1、1、2——正好是
// 「身上有 ＋1／＋2 盾牌」會產生的值。殘差落在 0..2 之外就表示算式錯了。
func TestSecondArmorClassMatchesTheRecalculationFormula(t *testing.T) {
	records := originalMonsterRecords(t)
	exact, residuals := 0, map[int]int{}
	for _, data := range records {
		record, err := Parse(data)
		if err != nil {
			continue
		}
		expected := record.ArmorClass - dexterityDefenceAdjustment(int(record.Dexterity)) - 2
		residual := expected - record.ArmorClassFacing
		if residual == 0 {
			exact++
		}
		residuals[residual]++
		if residual < 0 || residual > 2 {
			t.Errorf("%s：+19Ah=%d 敏捷=%d ⇒ 預期 +19Bh=%d，實際 %d（殘差 %d 超出盾牌加值的範圍）",
				record.Name, record.ArmorClass, record.Dexterity, expected, record.ArmorClassFacing, residual)
		}
		if record.ArmorClassFacing == 0 {
			t.Errorf("%s 的 +19Bh 是 0——原作記錄沒有這種值，解析位移可能錯了", record.Name)
		}
	}
	if exact*4 < len(records)*3 {
		t.Fatalf("%d 筆記錄裡只有 %d 筆殘差為 0（殘差分佈 %v），算式可能不對", len(records), exact, residuals)
	}
	t.Logf("記錄數 %d，殘差為 0 的 %d 筆，殘差分佈 %v", len(records), exact, residuals)
}

// 投影到 `Fighter` 的是**差值**：`ResolveAttack` 用 `attackTotal >= AC` 判命中，
// 數字小才好打，所以背後那一格一定要比正面那一格小。
// 直接搬 `+19Bh` 的絕對值會因為 `CombatArmorClass` 的反轉而變成「背後更難打」。
func TestFacingArmorClassProjectsThePenaltyNotTheStoredValue(t *testing.T) {
	for _, data := range originalMonsterRecords(t) {
		record, err := Parse(data)
		if err != nil {
			continue
		}
		fighter := record.Fighter("x", 0)
		if !fighter.ArmorClassFacingKnown {
			t.Fatalf("%s 的兩個護甲欄位都有值，卻沒標成已知", record.Name)
		}
		penalty := record.ArmorClass - record.ArmorClassFacing
		if penalty < 2 {
			t.Fatalf("%s 的減免是 %d——至少要有那個固定的 2", record.Name, penalty)
		}
		if fighter.ArmorClassFacing != fighter.ArmorClass-penalty {
			t.Fatalf("%s：正面 AC %d、背後 AC %d，差值應該是 %d",
				record.Name, fighter.ArmorClass, fighter.ArmorClassFacing, penalty)
		}
		if fighter.ArmorClassFacing >= fighter.ArmorClass {
			t.Fatalf("%s：背後 AC %d 不比正面 AC %d 好打",
				record.Name, fighter.ArmorClassFacing, fighter.ArmorClass)
		}
	}
}
