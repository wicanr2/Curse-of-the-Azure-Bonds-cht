package party

import (
	"os"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	engineability "github.com/wicanr2/golden-box-remake-engine/combat/ability"
	enginearmorclass "github.com/wicanr2/golden-box-remake-engine/combat/armorclass"
)

// 原版產生的角色檔把**輸入**（`+73h`、`+124h`、屬性、職業槽等級）與
// **算出來的結果**（`+199h`、`+19Ah`、`+19Bh`、`+1A2h`）同時存了下來，
// 所以它就是整條派生鏈的 oracle：拿輸入重算一次，四個結果要逐格相同。
//
// 這條測試一次釘住四張表：職業命中表（spec 1140）、力量命中與傷害調整
// （spec 694／697）、敏捷防禦調整（spec 694），以及第二個 AC 的算式
// （spec 1000 §七）。任何一張改壞了都會紅。
func TestOriginalCharacterRecordDerivationRoundTrips(t *testing.T) {
	const path = "../../docs/reference/original-dos/save-samples/CHRDATA1.sav"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skip("找不到原版角色樣本，跳過")
	}
	if len(data) < 0x1A6 {
		t.Fatalf("樣本只有 %d bytes", len(data))
	}
	var classLevels [8]uint8
	copy(classLevels[:], data[0x109:0x111])
	tables, err := gamepack.CombatBase()
	if err != nil {
		t.Fatal(err)
	}

	// 一、`+73h` ＝ 八個職業槽查表取最大。
	attackAbility, err := AttackAbilityFrom(tables, classLevels)
	if err != nil {
		t.Fatal(err)
	}
	if attackAbility != data[0x73] {
		t.Fatalf("+73h 算出 %d，記錄裡是 %d", attackAbility, data[0x73])
	}

	// 二、`+199h` ＝ `+73h` ＋ 力量命中調整（第 0 裝備槽是空的）。
	character := Character{
		AttackAbility:      data[0x73],
		BaseArmorClass:     data[0x124],
		AbilityAdjustments: data[0x125],
		Abilities: Abilities{
			StrengthFull:        int(data[0x11]),
			StrengthExceptional: int(data[0x1C]),
			Dexterity:           int(data[0x17]),
		},
	}
	hit, damage, ok := character.StrengthAdjustments()
	if !ok {
		t.Fatal("這個樣本的力量調整應該查得到")
	}
	if got := int(data[0x73]) + hit; got != int(data[0x199]) {
		t.Fatalf("+199h 算出 %d，記錄裡是 %d", got, data[0x199])
	}

	// 三、`+1A2h` ＝ 力量傷害調整（基準值為 0 的樣本）。
	if damage != int(data[0x1A2]) {
		t.Fatalf("+1A2h 算出 %d，記錄裡是 %d", damage, data[0x1A2])
	}

	// 四、`+19Ah` ＝ `+124h` ＋ 敏捷防禦調整（沒有裝備）。
	dexterity := engineability.DexterityDefenceAdjustment(int(data[0x17]))
	armorClass := int(data[0x124]) + dexterity
	if armorClass != int(data[0x19A]) {
		t.Fatalf("+19Ah 算出 %d，記錄裡是 %d", armorClass, data[0x19A])
	}

	// 五、`+19Bh` ＝ 扣掉敏捷與盾牌槽再減 2；這個樣本沒有裝備，盾牌那一項是 0。
	if got := enginearmorclass.Rear(armorClass, dexterity, 0); got != int(data[0x19B]) {
		t.Fatalf("+19Bh 算出 %d，記錄裡是 %d", got, data[0x19B])
	}
}
