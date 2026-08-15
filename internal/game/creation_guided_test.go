package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// 原版建角是種族 → 性別 → 職業 → 陣營四段（spec 1093 §一），
// 而職業那一段的選項由種族決定（spec 1099 §五）。
func TestGuidedCreationFollowsReferenceFourMenus(t *testing.T) {
	state := NewState(testCatalog())
	if err := state.BeginGuidedCreation(); err != nil {
		t.Fatal(err)
	}
	if state.GuidedStep != CreationStepRace {
		t.Fatalf("第一段應該是種族，得到 %d", state.GuidedStep)
	}
	races, err := state.GuidedCreationOptions()
	if err != nil {
		t.Fatal(err)
	}
	// 七個種族，值就是原作的 +74h 編號 1..7。
	if len(races) != 7 || races[0].Value != 1 || races[6].Value != 7 {
		t.Fatalf("種族選單=%v", races)
	}

	// 選人類（索引 6 ＝ 編號 7）。
	if err := state.SelectGuidedOption(6); err != nil {
		t.Fatal(err)
	}
	if state.GuidedDraft.Race != party.RaceHuman || state.GuidedStep != CreationStepGender {
		t.Fatalf("race=%v step=%d", state.GuidedDraft.Race, state.GuidedStep)
	}

	if err := state.SelectGuidedOption(1); err != nil { // 女性
		t.Fatal(err)
	}
	if state.GuidedDraft.Gender != party.GenderFemale || state.GuidedStep != CreationStepClass {
		t.Fatalf("gender=%v step=%d", state.GuidedDraft.Gender, state.GuidedStep)
	}

	classes, err := state.GuidedCreationOptions()
	if err != nil {
		t.Fatal(err)
	}
	// 人類六個單職，含聖騎士(3)與遊俠(4)——只有人類有這兩個。
	if len(classes) != 6 {
		t.Fatalf("人類應有 6 個可選職業，得到 %d", len(classes))
	}
	var hasPaladin, hasRanger bool
	for _, option := range classes {
		if option.Value == 3 {
			hasPaladin = true
		}
		if option.Value == 4 {
			hasRanger = true
		}
	}
	if !hasPaladin || !hasRanger {
		t.Fatalf("人類的職業選單缺聖騎士或遊俠：%v", classes)
	}

	// 選聖騎士：原作寫進 +75h 的是職業組合編號，不是選單索引。
	paladinIndex := -1
	for index, option := range classes {
		if option.Value == 3 {
			paladinIndex = index
		}
	}
	if err := state.SelectGuidedOption(paladinIndex); err != nil {
		t.Fatal(err)
	}
	if state.GuidedDraft.RawClassID != 3 || state.GuidedDraft.Class != party.ClassPaladin {
		t.Fatalf("raw=%d class=%v", state.GuidedDraft.RawClassID, state.GuidedDraft.Class)
	}
	if state.GuidedStep != CreationStepAlignment {
		t.Fatalf("step=%d, want alignment", state.GuidedStep)
	}

	alignments, err := state.GuidedCreationOptions()
	if err != nil {
		t.Fatal(err)
	}
	if len(alignments) != 9 {
		t.Fatalf("陣營應有 9 個，得到 %d", len(alignments))
	}
	if err := state.SelectGuidedOption(0); err != nil { // 守序善良
		t.Fatal(err)
	}
	if state.GuidedDraft.Alignment != 0 || !state.GuidedDraft.AlignmentKnown {
		t.Fatalf("alignment=%d known=%v", state.GuidedDraft.Alignment, state.GuidedDraft.AlignmentKnown)
	}
	if state.GuidedStep != CreationStepAbilities {
		t.Fatalf("step=%d, want abilities", state.GuidedStep)
	}

	// 擲屬性：結果必須落在種族與職業允許的範圍內。
	if err := state.RollGuidedAbilities(1234); err != nil {
		t.Fatal(err)
	}
	abilities := state.GuidedDraft.Abilities
	// 女性人類：力量 3–18；聖騎士要求魅力 ≥17、力量 ≥12、睿智 ≥13、體質 ≥9。
	if abilities.Charisma < 17 {
		t.Fatalf("聖騎士魅力應被夾到 ≥17，得到 %d", abilities.Charisma)
	}
	if abilities.Strength < 12 || abilities.Wisdom < 13 || abilities.Constitution < 9 {
		t.Fatalf("聖騎士屬性下限未套用：%+v", abilities)
	}
}

// 矮人不能當聖騎士——可選職業表只給戰士／盜賊／戰士-盜賊（spec 1099 §五）。
func TestGuidedCreationDwarfHasNoPaladin(t *testing.T) {
	state := NewState(testCatalog())
	if err := state.BeginGuidedCreation(); err != nil {
		t.Fatal(err)
	}
	if err := state.SelectGuidedOption(0); err != nil { // 矮人 ＝ 編號 1
		t.Fatal(err)
	}
	if err := state.SelectGuidedOption(0); err != nil { // 男性
		t.Fatal(err)
	}
	classes, err := state.GuidedCreationOptions()
	if err != nil {
		t.Fatal(err)
	}
	if len(classes) != 3 {
		t.Fatalf("矮人應有 3 個可選職業，得到 %d：%v", len(classes), classes)
	}
	for _, option := range classes {
		if option.Value == 3 || option.Value == 4 {
			t.Fatalf("矮人不該能當聖騎士或遊俠：%v", classes)
		}
	}
}
