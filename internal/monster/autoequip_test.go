package monster

import (
	"archive/zip"
	"io"
	"os"
	"testing"
)

// originalItemsMember 讀遊戲 image 裡的 `ITEMS` 成員——也就是原作執行時填進
// `DS:5CF6h` 的那張類別表（spec 1120）。找不到 image 就跳過，不假裝通過。
func originalItemsMember(t *testing.T) []byte {
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
	for _, file := range archive.File {
		if file.Name != "ITEMS" {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		defer reader.Close()
		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	t.Fatal("image 裡沒有 ITEMS 成員")
	return nil
}

// 類別表的位移對照——這一組數字是 spec 1120 的算式與 `BaseItem` 之間的橋。
// 對錯了會讓評分讀到完全不同的欄位，而症狀只是「AI 挑的武器怪怪的」。
func autoEquipCatalog(t *testing.T) BaseItemCatalog {
	t.Helper()
	// 索引 0：長劍類，單手、對小型 1d8、無加值、近戰。
	// 索引 1：弓類，雙手、對小型 1d6、射速 4、射程 20、發射 ＋ 需要彈藥槽 A。
	// 索引 2：匕首類，單手、對小型 1d4、可投擲、射程 2。
	// 索引 3：盾牌類，槽 1。
	data := make([]byte, BaseItemHeaderSize+4*BaseItemRecordSize)
	set := func(index int, fields map[int]byte) {
		offset := BaseItemHeaderSize + index*BaseItemRecordSize
		for position, value := range fields {
			data[offset+position] = value
		}
	}
	set(0, map[int]byte{0: 0, 1: 1, 9: 1, 10: 8, 13: 0xFF})
	set(1, map[int]byte{0: 0, 1: 2, 5: 4, 9: 1, 10: 6, 12: 20, 13: 0xFF,
		14: baseItemFlagFired | baseItemFlagAmmoSlotA})
	set(2, map[int]byte{0: 0, 1: 1, 9: 1, 10: 4, 12: 2, 13: 0xFF,
		14: baseItemFlagsThrowReady})
	set(3, map[int]byte{0: 1, 1: 1, 13: 0xFF})
	catalog, err := ParseBaseItems(data)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func TestAutoEquipScoreFollowsTheOriginalFormula(t *testing.T) {
	catalog := autoEquipCatalog(t)
	character := AutoEquipCharacter{ClassUsabilityMask: 0xFF}
	sword, _ := catalog.Lookup(0)
	// 1d8 ＝ 8 分，單手 ＋3 ⇒ 11。
	if got := ScoreWeaponForAutoEquip(ItemRecord{Type: 0}, sword, character); got != 11 {
		t.Fatalf("長劍 %d 分，want 11（1×8 ＋ 單手 3）", got)
	}
	// ＋2 的長劍多 16 分。
	if got := ScoreWeaponForAutoEquip(ItemRecord{Type: 0, Plus: 2}, sword, character); got != 27 {
		t.Fatalf("＋2 長劍 %d 分，want 27（11 ＋ 2×8）", got)
	}
	// 負加值不倒扣（原作是 `if > 0` 才加）。
	if got := ScoreWeaponForAutoEquip(ItemRecord{Type: 0, Plus: -1}, sword, character); got != 11 {
		t.Fatalf("−1 長劍 %d 分，want 11：負加值不倒扣", got)
	}
	bow, _ := catalog.Lookup(1)
	// 1d6 ＝ 6，射速 4 ⇒ ＋(4−1)×2 ＝ 6，雙手不加 3 ⇒ 12。
	if got := ScoreWeaponForAutoEquip(ItemRecord{Type: 1}, bow, character); got != 12 {
		t.Fatalf("弓 %d 分，want 12（6 ＋ 射速 6）", got)
	}
}

// 手不夠就是 0，不論前面加了多少分。
func TestAutoEquipScoreZeroesWhenHandsRunOut(t *testing.T) {
	catalog := autoEquipCatalog(t)
	bow, _ := catalog.Lookup(1)
	character := AutoEquipCharacter{ClassUsabilityMask: 0xFF, HandsInUse: 2}
	if got := ScoreWeaponForAutoEquip(ItemRecord{Type: 1, Plus: 5}, bow, character); got != 0 {
		t.Fatalf("兩隻手已經用掉、弓要兩隻手，卻拿到 %d 分", got)
	}
	// 界線是 `> 3`：已用 1 ＋ 需要 2 ＝ 3，不超過。
	character.HandsInUse = 1
	if got := ScoreWeaponForAutoEquip(ItemRecord{Type: 1}, bow, character); got == 0 {
		t.Fatal("已用 1 ＋ 需要 2 ＝ 3，沒有超過 3，不該歸零")
	}
}

// 三個歸零條件各自獨立。
func TestAutoEquipScoreZeroesOnCurseAlignmentAndRejectAffect(t *testing.T) {
	catalog := autoEquipCatalog(t)
	sword, _ := catalog.Lookup(0)
	character := AutoEquipCharacter{ClassUsabilityMask: 0xFF, Alignment: 3}
	for _, item := range []struct {
		name string
		item ItemRecord
	}{
		{"詛咒", ItemRecord{Type: 0, Plus: 5, Cursed: true}},
		{"陣營不符", ItemRecord{Type: 0, Plus: 5, Affects: [3]uint8{0, 0x05, 0x84}}},
		{"拒絕效果碼", ItemRecord{Type: 0, Plus: 5, Affects: [3]uint8{0, 0x53, 0}}},
	} {
		if got := ScoreWeaponForAutoEquip(item.item, sword, character); got != 0 {
			t.Errorf("%s 的物品拿到 %d 分，該歸零", item.name, got)
		}
	}
	// 陣營相符就不歸零。
	ok := ItemRecord{Type: 0, Plus: 5, Affects: [3]uint8{0, 0x03, 0x84}}
	if got := ScoreWeaponForAutoEquip(ok, sword, character); got == 0 {
		t.Error("陣營相符卻被歸零")
	}
}

// ★ 遠程的門檻是 `分A > 分B ÷ 2`。寫成 `> 分B` 會讓 AI 幾乎不用弓。
func TestRangedWinsWhenItBeatsHalfTheMeleeScore(t *testing.T) {
	catalog := autoEquipCatalog(t)
	items := []ItemRecord{{Type: 0, Plus: 1}, {Type: 1}}
	character := AutoEquipCharacter{ClassUsabilityMask: 0xFF, AmmunitionSlotA: true}
	choice := ChooseAutoEquipWeapon(items, catalog, character)
	// 長劍＋1 ＝ 19、弓 ＝ 12。12 < 19 但 12 > 19÷2 ＝ 9 ⇒ 選弓。
	if choice.MeleeScore != 19 || choice.RangedScore != 12 {
		t.Fatalf("分數 melee=%d ranged=%d，want 19／12", choice.MeleeScore, choice.RangedScore)
	}
	if !choice.ChoseRanged {
		t.Fatal("弓的分數低於長劍但高於一半，原作會選弓")
	}
}

// 沒有彈藥就不能選發射類。
func TestFiredWeaponNeedsAmmunition(t *testing.T) {
	catalog := autoEquipCatalog(t)
	items := []ItemRecord{{Type: 0}, {Type: 1}}
	character := AutoEquipCharacter{ClassUsabilityMask: 0xFF}
	if ChooseAutoEquipWeapon(items, catalog, character).ChoseRanged {
		t.Fatal("彈藥槽是空的還是選了弓")
	}
	character.AmmunitionSlotA = true
	if !ChooseAutoEquipWeapon(items, catalog, character).ChoseRanged {
		t.Fatal("有彈藥卻沒選弓")
	}
}

// 敵人貼身就換近戰；投擲武器例外（射程 > 1 且旗標齊）。
func TestAdjacentEnemyForcesMeleeExceptForThrownWeapons(t *testing.T) {
	catalog := autoEquipCatalog(t)
	character := AutoEquipCharacter{ClassUsabilityMask: 0xFF, AmmunitionSlotA: true, EnemyAdjacent: true}
	if ChooseAutoEquipWeapon([]ItemRecord{{Type: 0}, {Type: 1}}, catalog, character).ChoseRanged {
		t.Fatal("敵人貼身還是拿弓")
	}
	thrown := ChooseAutoEquipWeapon([]ItemRecord{{Type: 0}, {Type: 2, Plus: 3}}, catalog, character)
	if !thrown.ChoseRanged {
		t.Fatalf("投擲武器射程 2、旗標齊，貼身也該照用：%+v", thrown)
	}
}

// 用不了的類別整件跳過——不是 0 分，是根本不進候選。
func TestUnusableClassIsSkippedEntirely(t *testing.T) {
	catalog := autoEquipCatalog(t)
	character := AutoEquipCharacter{ClassUsabilityMask: 0x00}
	choice := ChooseAutoEquipWeapon([]ItemRecord{{Type: 0, Plus: 5}}, catalog, character)
	if choice.Melee != nil || choice.Chosen != nil {
		t.Fatalf("可用性遮罩為 0 卻挑到了 %+v", choice)
	}
}

// 盾牌那一槽只看加值。
func TestShieldSlotScoresPlusOnly(t *testing.T) {
	catalog := autoEquipCatalog(t)
	character := AutoEquipCharacter{ClassUsabilityMask: 0xFF}
	choice := ChooseAutoEquipWeapon([]ItemRecord{{Type: 3, Plus: 2}, {Type: 3, Plus: -1}}, catalog, character)
	if choice.Shield == nil || choice.ShieldScore != 3 {
		t.Fatalf("盾牌 %+v 分數 %d，want ＋2 那面、3 分", choice.Shield, choice.ShieldScore)
	}
}

// 用真的 `ITEMS` 資料釘住「旗標判斷比類別列舉多抓到哪些」。
//
// ★ 這條測試的價值在於**證明修法有差**：如果哪天 `IsMissileWeapon` 又被改回
// 類別列舉，它會紅在「少了 11 個類別」上，而不是安靜地讓 AI 把弓當棍子揮。
func TestMissileFlagFindsMoreThanTheOldTypeRange(t *testing.T) {
	catalog, err := ParseBaseItems(originalItemsMember(t))
	if err != nil {
		t.Fatal(err)
	}
	var flagged, ranged []int
	for index, item := range catalog.Items {
		if item.IsMissileWeapon() {
			flagged = append(flagged, index)
		}
		if index >= 41 && index <= 47 {
			ranged = append(ranged, index)
		}
	}
	if len(flagged) <= len(ranged) {
		t.Fatalf("旗標抓到 %d 個、舊的類別列舉 %d 個——修法沒有差別，先確認讀的是同一份資料",
			len(flagged), len(ranged))
	}
	for _, index := range ranged {
		if !catalog.Items[index].IsMissileWeapon() {
			t.Errorf("類別 %d 原本就被當成遠程，旗標卻沒設——修法弄丟了既有行為", index)
		}
	}
	// 正對照：近戰武器不能被誤判成遠程。長劍（類別 1）沒有射程。
	if catalog.Items[1].IsMissileWeapon() || catalog.Items[1].IsThrownWeapon() {
		t.Error("類別 1 被判成遠程／投擲，旗標位元對錯了")
	}
}
