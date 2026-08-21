package party

import (
	"reflect"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

// ★★ 這一條擋的是一整類 bug，不是一個欄位。
//
// `FighterWithEquipment` 把裝備投影成戰鬥員的欄位；自動換裝之後由
// `Battle.ReplaceFighterEquipment` 把新的投影搬回戰鬥中的那一隻。**兩份名單必須
// 一致**，而它們分別寫在兩個套件裡——先前 `LargeDamage*` 與 `HasSlotZeroWeapon`
// 加進投影卻沒加進搬運，結果是「換裝之後打大型目標用的還是上一把武器的骰」
// （spec 1175）。
//
// 名單不用手寫：**拿同一個角色帶武器與不帶武器各投影一次，有差的欄位就是
// 裝備衍生的**。這樣新增欄位不必記得更新任何清單。
func TestReplaceFighterEquipmentCarriesEveryEquipmentDerivedField(t *testing.T) {
	catalog := weaponSlotCatalog()
	base := Character{
		ID: "p1", Name: "戰士", Race: RaceHuman, Class: ClassFighter, Level: 3,
		Abilities: Abilities{Strength: 12, Intelligence: 10, Wisdom: 10,
			Dexterity: 10, Constitution: 10, Charisma: 10},
	}
	armed := base
	armed.Equipment = []monster.ItemRecord{{Name: "Sword", Type: 0x10, Readied: true}}

	bare, err := base.FighterWithEquipment(catalog)
	if err != nil {
		t.Fatal(err)
	}
	projected, err := armed.FighterWithEquipment(catalog)
	if err != nil {
		t.Fatal(err)
	}

	derived := differingFieldNames(t, bare, projected)
	if len(derived) == 0 {
		t.Fatal("帶武器與不帶武器投影出來一模一樣——fixture 沒有測到任何東西")
	}

	battle, err := combat.NewBattle([]combat.Fighter{bare, {
		ID: "e1", Side: combat.SideEnemy, HitPoints: 5, MaxHitPoints: 5,
	}}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := battle.ReplaceFighterEquipment(bare.ID, projected); err != nil {
		t.Fatal(err)
	}
	var stored combat.Fighter
	for _, fighter := range battle.Fighters() {
		if fighter.ID == bare.ID {
			stored = fighter
		}
	}

	storedValue := reflect.ValueOf(stored)
	wantValue := reflect.ValueOf(projected)
	for _, name := range derived {
		got := storedValue.FieldByName(name).Interface()
		want := wantValue.FieldByName(name).Interface()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("`%s` 是裝備衍生的，但換裝重投影沒有把它搬過去：%v，want %v",
				name, got, want)
		}
	}
}

// differingFieldNames 回報兩個戰鬥員之間值不同的匯出欄位名。
func differingFieldNames(t *testing.T, left, right combat.Fighter) []string {
	t.Helper()
	leftValue, rightValue := reflect.ValueOf(left), reflect.ValueOf(right)
	structType := leftValue.Type()
	names := make([]string, 0, structType.NumField())
	for index := 0; index < structType.NumField(); index++ {
		field := structType.Field(index)
		if field.PkgPath != "" {
			continue
		}
		if reflect.DeepEqual(leftValue.Field(index).Interface(), rightValue.Field(index).Interface()) {
			continue
		}
		names = append(names, field.Name)
	}
	return names
}
