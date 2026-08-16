package ecl

import "testing"

// spellProbeBlock 照原作 `ECL4.DAX/0x22 +079Fh` 的形狀組一段 ECL：
//
//	SPELL 22, [7F79h], [7F7Ah]
//	COMPARE [7F79h], imm FFh
//	IF <>            ; 條件不成立時整條下一個指令被跳過（spec 1106）
//	GOTO found
//	SAVE imm 1 -> [7F00h]   ; 沒找到
//	EXIT
//
// found:
//
//	SAVE imm 2 -> [7F00h]   ; 找到了
//	EXIT
//
// ★ 驗的是**呼叫端看得到的效果**（走哪一條路），不是 `SpellSearch` 的欄位值。
// 欄位對了但沒寫回記憶體，ECL 照樣會走錯分支——那才是玩家會遇到的錯。
func spellProbeBlock() []byte {
	return []byte{
		0x00, 0x00, // block 前綴
		0x3B, 0x00, 0x16, 0x01, 0x79, 0x7F, 0x01, 0x7A, 0x7F, // +0000 SPELL 22,[7F79h],[7F7Ah]
		0x03, 0x01, 0x79, 0x7F, 0x00, 0xFF, // +0009 COMPARE [7F79h], FFh
		0x17,                   // +000F IF <>
		0x01, 0x01, 0x1B, 0x80, // +0010 GOTO +001Bh
		0x09, 0x00, 0x01, 0x01, 0x00, 0x7F, // +0014 SAVE 1 -> [7F00h]（沒找到）
		0x00,                               // +001A EXIT
		0x09, 0x00, 0x02, 0x01, 0x00, 0x7F, // +001B SAVE 2 -> [7F00h]（找到了）
		0x00, // +0021 EXIT
	}
}

func spellProbeOutcome(t *testing.T, result RunResult) uint16 {
	t.Helper()
	for _, write := range result.SaveWrites {
		if write.Address == 0x7F00 {
			return write.Value
		}
	}
	t.Fatalf("兩條路都沒有寫 7F00h：%+v", result.SaveWrites)
	return 0
}

func TestSpellSearchLetsTheCallerTakeTheFoundBranch(t *testing.T) {
	party := PartyContext{Members: []PartyMemberContext{
		{Name: "A", SpellSlots: []uint8{1, 2}},
		{Name: "B", SpellSlots: []uint8{9, 22, 3}},
	}}
	result, err := RunSubsetInteractiveSeedWithPartyContext(spellProbeBlock(), 0, 64, nil, 1, party)
	if err != nil {
		t.Fatal(err)
	}
	if got := spellProbeOutcome(t, result); got != 2 {
		t.Fatalf("走到了「沒找到」那一條路（7F00h=%d）；SPELL 沒有把 slot 寫回去", got)
	}
	search := result.SpellSearches[0]
	if !search.Found || search.MemberIndex != 1 || search.SlotIndex != 1 {
		t.Fatalf("search=%+v，want 第 1 名隊員的第 1 個法術槽", search)
	}
}

func TestSpellSearchLetsTheCallerTakeTheNotFoundBranch(t *testing.T) {
	party := PartyContext{Members: []PartyMemberContext{{Name: "A", SpellSlots: []uint8{1, 2}}}}
	result, err := RunSubsetInteractiveSeedWithPartyContext(spellProbeBlock(), 0, 64, nil, 1, party)
	if err != nil {
		t.Fatal(err)
	}
	if got := spellProbeOutcome(t, result); got != 1 {
		t.Fatalf("沒人會這個法術，卻走到「找到了」那一條路（7F00h=%d）", got)
	}
	if result.SpellSearches[0].Found {
		t.Fatal("沒人會這個法術，search 卻回報找到")
	}
}

// 沒有隊伍資料時**不寫回**：ECL 那一側分不出「找不到」與「不知道」。
// 寫一個 FFh 進去會讓劇情走進「沒人會這個法術」的分支，那是猜出來的結果。
func TestSpellSearchDoesNotDecideWithoutPartyContext(t *testing.T) {
	result, err := RunSubset(spellProbeBlock(), 0, 64)
	if err != nil {
		t.Fatal(err)
	}
	if result.SpellSearches[0].Resolved {
		t.Fatal("沒有 party context 卻宣稱查得出來")
	}
	// 沒寫回 ⇒ 比較用的那一格是 0，`COMPARE 0, FFh` 不相等 ⇒ 走 found 分支。
	// 這個結果本身不對，但**它是可見的**：不寫回就是把決定權留給上層，
	// 而不是讓 VM 猜一個。真正的修法是永遠帶著 party context 進來。
	if got := spellProbeOutcome(t, result); got != 2 {
		t.Fatalf("7F00h=%d，與「未寫回」的推論不符", got)
	}
}
