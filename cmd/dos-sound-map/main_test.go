package main

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98sfx"
)

func TestDifferenceCountsPerModule(t *testing.T) {
	left := distribution{"overlay-02": 3, "overlay-13": 1}
	right := distribution{"overlay-02": 3, "overlay-13": 1}
	if got := difference(left, right); got != 0 {
		t.Fatalf("相同的分佈差 ＝ %d，want 0", got)
	}
	// ⚠ 差要**逐模組**算：總數相同但分佈不同必須是非零，否則
	// 「overlay-02×4」與「overlay-02×3、overlay-13×1」會被當成同一個東西。
	skewed := distribution{"overlay-02": 4}
	if got := difference(left, skewed); got == 0 {
		t.Fatal("總數相同但模組不同，差不該是 0")
	}
	// 一邊沒有的模組也要算進去。
	if got := difference(left, distribution{}); got != 4 {
		t.Fatalf("對空分佈的差 ＝ %d，want 4", got)
	}
}

// 這一條驗的是整支工具的核心主張：**整張表平移一格之後，位址仍然連續、名字仍然
// 一一對應，光看表看不出偏移；只有呼叫點分佈會整排錯開。**
//
// ★ 沒有這一條的話，`pickBase` 就算永遠回傳「最低的那個描述子」也一樣會通過
// ——而那在「最低那一格剛好沒有 DOS 呼叫點」時會整排偏掉。
func TestPickBaseRejectsAShiftedTable(t *testing.T) {
	const dosBase = 0x25AA
	// 用真實的選擇子表造一份「DOS 側」：每個效果給一個**彼此不同**的分佈。
	pc98ByEffect := map[string]distribution{}
	dosByDescriptor := map[int]distribution{}
	descriptors := make([]int, 0, 17)
	for index, info := range pc98sfx.Selectors() {
		dist := distribution{"overlay-" + string(rune('A'+index)): index + 1}
		pc98ByEffect[info.Symbol] = dist
		address := dosBase + (info.Descriptor - pc98sfx.HaltDescriptor)
		dosByDescriptor[address] = dist
		descriptors = append(descriptors, address)
	}

	base, best, second := pickBase(descriptors, dosByDescriptor, pc98ByEffect)
	if base != dosBase {
		t.Fatalf("挑到的基底 %04Xh，want %04Xh", base, dosBase)
	}
	if best != len(pc98sfx.Selectors()) {
		t.Fatalf("正確基底只對上 %d 格，want %d", best, len(pc98sfx.Selectors()))
	}
	// ⚠ 這一行是重點：平移一格必須**幾乎全錯**。分數咬得很近就代表分佈沒有
	// 鑑別力，那時候這張對照表只是巧合。
	if second != 0 {
		t.Fatalf("平移一格還對上 %d 格——分佈沒有鑑別力，對照不算數", second)
	}
	// 直接驗平移後的分數，不依賴排序結果。
	if shifted := scoreBase(dosBase+2, dosByDescriptor, pc98ByEffect); shifted != 0 {
		t.Fatalf("往後平移一格對上 %d 格，want 0", shifted)
	}
	if shifted := scoreBase(dosBase-2, dosByDescriptor, pc98ByEffect); shifted != 0 {
		t.Fatalf("往前平移一格對上 %d 格，want 0", shifted)
	}
}

// 負對照：DOS 那一側沒有呼叫點的格子不參與計分。少了這條規則，那些格子會在
// **任何**錨點下都「相同」，把所有候選拉成平手。
func TestScoreBaseIgnoresDescriptorsWithNoDOSSites(t *testing.T) {
	pc98ByEffect := map[string]distribution{}
	for _, info := range pc98sfx.Selectors() {
		pc98ByEffect[info.Symbol] = distribution{}
	}
	if got := scoreBase(0x25AA, map[int]distribution{}, pc98ByEffect); got != 0 {
		t.Fatalf("兩邊都沒有呼叫點時分數 ＝ %d，want 0", got)
	}
}
