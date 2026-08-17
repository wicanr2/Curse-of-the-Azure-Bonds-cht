package monster

// AI 自動換裝：評分挑最佳武器（spec 1120）。
//
// ★ 類別表就是 `ITEMS` 成員。 原作的評分逐欄查 `DS:5CF6h`，而那一段在執行檔的
// 常駐資料段裡是整片 `FF`（BSS）——它是**開機時從 `ITEMS` 載進去的**，也就是
// `BaseItemCatalog` 已經在解析的那 128 筆 16 bytes。欄位一一對得上：
// `+0` 裝備槽、`+1` 佔手數、`+2`／`+3` 對大型的骰、`+9`／`+0Ah` 對小型的骰、
// `+0Bh` 對小型的加值、`+0Ch` 射程、`+0Dh` 職業可用性遮罩、`+0Eh` 旗標。

const (
	// baseItemFlagFired 是「用發射的」（有射速）。`+0Eh` bit 3。
	baseItemFlagFired uint8 = 0x08
	// baseItemFlagThrown 是「用投擲的」。`+0Eh` bit 4。
	baseItemFlagThrown uint8 = 0x10
	// baseItemFlagAmmoSlotA／B 選角色身上兩個彈藥槽（`+17Dh`／`+181h`）中的哪一個。
	baseItemFlagAmmoSlotA uint8 = 0x01
	baseItemFlagAmmoSlotB uint8 = 0x80
	// baseItemFlagsSelfSufficient：`+0Eh` 剛好等於這個值時不需要另外的彈藥。
	// 語意未解讀，照數值比對。
	baseItemFlagsSelfSufficient uint8 = 0x0A
	// baseItemFlagsThrowReady 是「投擲候選要同時具備的兩個位元」（`14h`）。
	// bit 4 是投擲，bit 2 的語意未解讀——所以照原作比整組，不拆開命名。
	baseItemFlagsThrowReady uint8 = 0x14

	// autoEquipWeaponSlot 是武器槽（類別表 `+0`）。評分迴圈只看這一槽。
	autoEquipWeaponSlot uint8 = 0
	// autoEquipShieldSlot 是另一組統計那一槽（類別表 `+0` ＝ 1）。
	autoEquipShieldSlot uint8 = 1

	// autoEquipHandLimit 是 `角色^[185h] + 這件的手數 > 3` 那個界線。
	// ⚠ 是 3 不是 2——換裝主流程另外用 `> 2`／`< 2`，兩個數字各自出現在不同
	// 位置，不要統一。
	autoEquipHandLimit = 3

	// autoEquipAlignmentAffect 是「這件物品有陣營限制」的效果碼（物品 `+3Eh`）。
	autoEquipAlignmentAffect uint8 = 0x84
	// autoEquipRejectAffect 是「一律不能自動裝上」的效果碼（物品 `+3Dh`）。
	autoEquipRejectAffect uint8 = 0x53
	// autoEquipTargetSpecificType 是那個「對某種目標特別有效」的物品類別。
	autoEquipTargetSpecificType uint8 = 0x55
	// autoEquipTargetSpecificScore 是上面那一類命中時的分數。**覆蓋不是累加**。
	autoEquipTargetSpecificScore = 8
)

// AutoEquipCharacter 是評分需要的角色側資料。欄位名沿用已解讀的語意，
// 未解讀的保留原始位移。
type AutoEquipCharacter struct {
	// ClassUsabilityMask 是角色 `+12Bh`，與類別表 `+0Dh` 做 `and`，為 0 就用不了。
	ClassUsabilityMask uint8
	// Alignment 是角色 `+11Bh`。
	Alignment uint8
	// HandsInUse 是 `+185h`：主流程已經把「考慮中的那些」扣掉了。
	HandsInUse int
	// AmmunitionSlotA／B 是 `+17Dh`／`+181h` 兩個彈藥槽有沒有東西。
	AmmunitionSlotA bool
	AmmunitionSlotB bool
	// TargetTakesTargetSpecific 是「目前有目標，而且目標的 `+0E9h` > 0」。
	// `+0E9h` 是什麼沒有解讀。
	TargetTakesTargetSpecific bool
	// EnemyAdjacent 是 `overlay-24 entry#32(角色, 1) <> 0`：半徑 1 內有敵人。
	EnemyAdjacent bool
}

// ScoreWeaponForAutoEquip 重現 `overlay-09:01542h`。
//
// 三個歸零條件寫在最後而且沒有 else，所以任何一個成立就一定是 0，
// 不論前面加了多少分。
func ScoreWeaponForAutoEquip(item ItemRecord, base BaseItem, character AutoEquipCharacter) int {
	score := int(base.SmallDamageDice) * int(base.SmallDamageSides)
	if item.Plus > 0 {
		score += item.Plus * 8
	}
	if base.SmallDamageBonus > 0 {
		score += int(base.SmallDamageBonus) * 2
	}
	if item.Type == autoEquipTargetSpecificType && character.TargetTakesTargetSpecific {
		score = autoEquipTargetSpecificScore
	}
	if base.AmmunitionType&baseItemFlagFired != 0 {
		// 射速只對發射類有意義，所以這一項掛在那個位元下面。
		score += (int(base.RateOfFire) - 1) * 2
	}
	if base.HandsRequired <= 1 {
		score += 3
	}
	if character.HandsInUse+int(base.HandsRequired) > autoEquipHandLimit {
		score = 0
	}
	if item.Affects[2] == autoEquipAlignmentAffect &&
		item.Affects[1]&0x0F != character.Alignment {
		score = 0
	}
	if item.Affects[1] == autoEquipRejectAffect {
		score = 0
	}
	if item.Cursed {
		score = 0
	}
	if score < 0 {
		// byte 運算：原作存回一個位元組，負數在那裡不會出現。
		score = 0
	}
	return score
}

// AutoEquipChoice 是掃完整條物品鏈之後的結果。
type AutoEquipChoice struct {
	// Ranged 是「發射或投擲」那一組的最高分（原作的最佳A）。
	Ranged      *ItemRecord
	RangedScore int
	// Melee 是「不發射」那一組（原作的最佳B）。投擲武器**兩組都進**，
	// 因為它近戰也能用。
	Melee      *ItemRecord
	MeleeScore int
	// Shield 是類別表 `+0` ＝ 1 那一槽的最高分，評分只有 `加值 + 1`。
	Shield      *ItemRecord
	ShieldScore int
	// Chosen 是最後選的那一件（可能為 nil）。
	Chosen *ItemRecord
	// ChoseRanged 記錄走的是遠程那一支。
	ChoseRanged bool
}

// ChooseAutoEquipWeapon 重現 `overlay-09:01681h` 的挑選段（spec 1120 §二／§三）。
//
// ⚠ 遠程勝出的門檻是 **`分A > 分B ÷ 2`**，不是 `分A > 分B`——原作明顯偏好遠程，
// 遠程只要贏過近戰的一半就會被選走。寫成 `>` 會讓 AI 幾乎不用弓。
func ChooseAutoEquipWeapon(items []ItemRecord, catalog BaseItemCatalog,
	character AutoEquipCharacter) AutoEquipChoice {
	choice := AutoEquipChoice{}
	for index := range items {
		item := items[index]
		base, ok := catalog.Lookup(item.Type)
		if !ok {
			continue
		}
		if base.ClassUsabilityMask&character.ClassUsabilityMask == 0 {
			continue
		}
		switch base.Slot {
		case autoEquipWeaponSlot:
			score := ScoreWeaponForAutoEquip(item, base, character)
			if base.AmmunitionType&(baseItemFlagFired|baseItemFlagThrown) != 0 &&
				score > choice.RangedScore {
				choice.Ranged, choice.RangedScore = &items[index], score
			}
			if base.AmmunitionType&baseItemFlagFired == 0 && score > choice.MeleeScore {
				choice.Melee, choice.MeleeScore = &items[index], score
			}
		case autoEquipShieldSlot:
			// 這一槽的評分只看加值，負加值當 0。
			score := 0
			if item.Plus >= 0 {
				score = item.Plus + 1
			}
			if score > choice.ShieldScore {
				choice.Shield, choice.ShieldScore = &items[index], score
			}
		}
	}
	choice.Chosen = choice.Melee
	if choice.Ranged == nil {
		return choice
	}
	rangedBase, _ := catalog.Lookup(choice.Ranged.Type)
	flags := rangedBase.AmmunitionType
	hasAmmunition := flags&baseItemFlagThrown != 0 ||
		(flags&baseItemFlagFired != 0 &&
			((flags&baseItemFlagAmmoSlotA != 0 && character.AmmunitionSlotA) ||
				(flags&baseItemFlagAmmoSlotB != 0 && character.AmmunitionSlotB))) ||
		flags == baseItemFlagsSelfSufficient
	// 投擲武器而且射程 > 1 時，就算敵人貼身也照用——丟得出去。
	throwReady := rangedBase.Range > 1 && flags&baseItemFlagsThrowReady == baseItemFlagsThrowReady
	if choice.RangedScore > choice.MeleeScore/2 && hasAmmunition &&
		(throwReady || !character.EnemyAdjacent) {
		choice.Chosen, choice.ChoseRanged = choice.Ranged, true
	}
	return choice
}
