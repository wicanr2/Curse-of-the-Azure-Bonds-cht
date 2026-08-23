package gamepack

import (
	"fmt"
	"strings"
	"testing"
)

// tavernTaleCount 是 Adventurer's Journal 收的酒館傳聞則數（1..62）。
const tavernTaleCount = 62

// TestEveryTavernTaleResolvesForBothPhrasings 擋住兩種會安靜顯示錯字的情形。
//
// ★ 為什麼需要這一條。 `MatchText` 是**第一個命中的規則就贏**，而規則的比對是
// **子字串**——`YOU OVERHEAR TAVERN TALE 3` 是 `... TALE 30` 的前綴，所以只要
// 第 3 則的規則排在第 30 則前面，玩家看到第 30 則時讀到的會是第 3 則的內容。
// 那不是錯誤訊息，是**一段讀起來完全正常、但講錯事情的中文**。實測第 30、32..39
// 共九則原本就是這樣被搶走的。
//
// ⚠ 原作有兩種說法（`YOU OVERHEAR TAVERN TALE <n>` 與 `HE TELLS TAVERN TALE # <n>`），
// 兩種都要對到同一則。
func TestEveryTavernTaleResolvesForBothPhrasings(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	for number := 1; number <= tavernTaleCount; number++ {
		overheard := pack.MatchText([]string{
			fmt.Sprintf("YOU OVERHEAR TAVERN TALE %d .", number)}, "zh-TW")
		// 第三種說法：編號放在**句尾**，後面什麼都沒有。
		suffix := pack.MatchText([]string{fmt.Sprintf(
			"AS YOU CONSUME THE LOCAL EXCUSE FOR FOOD AND DRINK, YOU OVERHEAR TAVERN TALE %d",
			number)}, "zh-TW")
		if !suffix.Matched || suffix.Message != overheard.Message {
			t.Errorf("第 %d 則傳聞：編號在句尾那一種說法對到 rule=%q，與 rule=%q 不同",
				number, suffix.RuleID, overheard.RuleID)
		}
		if !overheard.Matched || strings.TrimSpace(overheard.Message) == "" {
			t.Errorf("第 %d 則傳聞（聽見）對不到譯文：rule=%q", number, overheard.RuleID)
			continue
		}
		told := pack.MatchText([]string{
			fmt.Sprintf("HE TELLS TAVERN TALE # %d", number)}, "zh-TW")
		if !told.Matched || strings.TrimSpace(told.Message) == "" {
			t.Errorf("第 %d 則傳聞（他說）對不到譯文：rule=%q", number, told.RuleID)
			continue
		}
		if told.Message != overheard.Message {
			t.Errorf("第 %d 則傳聞兩種說法對到不同內容：聽見 rule=%q／他說 rule=%q",
				number, overheard.RuleID, told.RuleID)
		}
	}
}

// TestTavernTalesAreDistinct 擋住「兩則傳聞共用同一段譯文」——那是前綴被搶走
// 之後最常見的形狀，而且逐則看沒有一則看起來是錯的。
func TestTavernTalesAreDistinct(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]int{}
	for number := 1; number <= tavernTaleCount; number++ {
		message := pack.MatchText([]string{
			fmt.Sprintf("YOU OVERHEAR TAVERN TALE %d .", number)}, "zh-TW").Message
		if previous, ok := seen[message]; ok {
			t.Errorf("第 %d 則與第 %d 則的譯文一模一樣：%.40q", number, previous, message)
			continue
		}
		seen[message] = number
	}
}
