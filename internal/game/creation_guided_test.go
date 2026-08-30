package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func TestNormalCreationEntryUsesOriginalFlowAndReturnsToPartyAssembly(t *testing.T) {
	state := NewState(testCatalog())
	state.SetCharacterLibraryPath(filepath.Join(t.TempDir(), "characters.json"))
	if err := state.OpenCharacterCreation(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCharacterCreation || state.GuidedActive || state.CreationOuterStep != CreationOuterMenu {
		t.Fatalf("mode=%v active=%v outer=%d", state.Mode, state.GuidedActive, state.CreationOuterStep)
	}
	if err := state.BeginGuidedCreation(); err != nil {
		t.Fatal(err)
	}
	if err := state.CancelGuidedCreation(); err != nil {
		t.Fatal(err)
	}
	if state.Mode != ModeCharacterCreation || state.GuidedActive || state.GuidedStep != CreationStepDone {
		t.Fatalf("cancel should return to party assembly: mode=%v active=%v step=%d", state.Mode, state.GuidedActive, state.GuidedStep)
	}
}

// 原版建角是種族 → 性別 → 職業 → 陣營四段（spec 1093 §一），
// 而職業那一段的選項由種族決定（spec 1099 §五）。
func TestGuidedCreationFollowsReferenceFourMenus(t *testing.T) {
	state := NewState(testCatalog())
	libraryPath := filepath.Join(t.TempDir(), "characters.json")
	state.SetCharacterLibraryPath(libraryPath)
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
	// 六個種族：原作的顯示迴圈沒有分支收半獸人（6），所以建角取不到它；
	// 其餘仍用原作的 +74h 編號，不重新連號（spec 1102 §一）。
	if len(races) != 6 || races[0].Value != 1 || races[5].Value != 7 {
		t.Fatalf("種族選單=%v", races)
	}
	for _, option := range races {
		if option.Value == 6 {
			t.Fatalf("半獸人不應該出現在建角選單：%v", races)
		}
	}

	// 選人類（最後一項，編號 7）。
	if err := state.SelectGuidedOption(5); err != nil {
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
	// 聖騎士只有守序善一項（spec 1102 §二 的 DS:41D8h）。
	if len(alignments) != 1 || alignments[0].Value != 0 {
		t.Fatalf("聖騎士的陣營選單=%v", alignments)
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

	// HP 在重擲迴圈裡就算好了（spec 1101）。
	// 單職聖騎士：1d10 擲兩次取大 ⇒ 1..10，再加體質加值。
	if state.GuidedDraft.MaxHitPoints < 1 ||
		state.GuidedDraft.HitPoints != state.GuidedDraft.MaxHitPoints {
		t.Fatalf("HP=%d max=%d", state.GuidedDraft.HitPoints, state.GuidedDraft.MaxHitPoints)
	}

	// `Reroll stats? ` 答 'N' ⇒ 進名字輸入，年齡在這一步決定。
	if err := state.AcceptGuidedAbilities(1234); err != nil {
		t.Fatal(err)
	}
	if state.GuidedStep != CreationStepName {
		t.Fatalf("step=%d, want name", state.GuidedStep)
	}
	if state.GuidedDraft.Age <= 0 {
		t.Fatalf("年齡=%d", state.GuidedDraft.Age)
	}

	// 名字要支援中文，而且空名字要被擋下來（原作直接重問）。
	if err := state.CommitGuidedName(); err == nil {
		t.Fatal("空名字應該被擋下")
	}
	if err := state.AppendGuidedName([]rune("凱蘭德爾")); err != nil {
		t.Fatal(err)
	}
	if err := state.BackspaceGuidedName(); err != nil {
		t.Fatal(err)
	}
	if state.GuidedName != "凱蘭德" {
		t.Fatalf("退格是按字元不是位元組：%q", state.GuidedName)
	}
	if err := state.CommitGuidedName(); err != nil {
		t.Fatal(err)
	}
	if state.GuidedStep != CreationStepIcon {
		t.Fatalf("step=%d, want icon", state.GuidedStep)
	}
	if err := state.AdjustGuidedIcon(-1); err != nil || state.GuidedDraft.IconHeadBlock != 13 {
		t.Fatalf("頭部向前環繞失敗：head=%d err=%v", state.GuidedDraft.IconHeadBlock, err)
	}
	if err := state.MoveGuidedIconCursor(1); err != nil {
		t.Fatal(err)
	}
	if err := state.AdjustGuidedIcon(-1); err != nil || state.GuidedDraft.IconWeaponBlock != 31 {
		t.Fatalf("武器向前環繞失敗：weapon=%d err=%v", state.GuidedDraft.IconWeaponBlock, err)
	}
	if err := state.MoveGuidedIconCursor(1); err != nil {
		t.Fatal(err)
	}
	if err := state.AdjustGuidedIcon(-1); err != nil || state.GuidedDraft.IconSize != 1 {
		t.Fatalf("體型環繞失敗：size=%d err=%v", state.GuidedDraft.IconSize, err)
	}
	if state.GuidedDraft.IconColors != party.DefaultIconColors {
		t.Fatalf("圖示顏色預設值錯誤：% X", state.GuidedDraft.IconColors)
	}
	if err := state.MoveGuidedIconCursor(1); err != nil { // size -> body colour 1
		t.Fatal(err)
	}
	if err := state.AdjustGuidedIcon(-2); err != nil || state.GuidedDraft.IconColors[0] != 0x9F {
		t.Fatalf("身體第一色未依 0..F 環繞：%02X err=%v", state.GuidedDraft.IconColors[0], err)
	}
	if err := state.ConfirmGuidedIcon(false); err != nil {
		t.Fatal(err)
	}
	if state.GuidedDraft.IconHeadBlock != 0 || state.GuidedDraft.IconWeaponBlock != 0 || state.GuidedDraft.IconSize != 2 || state.GuidedDraft.IconColors != party.DefaultIconColors {
		t.Fatalf("Exit 沒有還原圖示：head=%d weapon=%d size=%d",
			state.GuidedDraft.IconHeadBlock, state.GuidedDraft.IconWeaponBlock, state.GuidedDraft.IconSize)
	}
	// 基準值複製在名字之後（spec 1093 §六之二）。
	for index := 0; index < 5; index++ {
		current, err := state.GuidedDraft.Abilities.CurrentValue(index + 1)
		if err != nil {
			t.Fatal(err)
		}
		base, err := state.GuidedDraft.Abilities.Value(index + 1)
		if err != nil {
			t.Fatal(err)
		}
		if current != base {
			t.Fatalf("屬性 %d 的基準值沒抄過去：目前=%d 基準=%d", index+1, current, base)
		}
	}

	// `Save <名字>? ` 答 'N' 之外都存檔。
	if err := state.ConfirmGuidedSave(true); err != nil {
		t.Fatal(err)
	}
	if len(state.CreationRoster) != 0 {
		t.Fatalf("存角色不能自動入隊：%v", state.CreationRoster)
	}
	if len(state.CreationLibrary) != 1 || state.CreationLibrary[0].Name != "凱蘭德" {
		t.Fatalf("角色庫=%v", state.CreationLibrary)
	}
	if state.GuidedActive {
		t.Fatal("存完應該離開四段流程")
	}
	reloaded := NewState(testCatalog())
	reloaded.SetCharacterLibraryPath(libraryPath)
	if err := reloaded.OpenCharacterCreation(); err != nil {
		t.Fatal(err)
	}
	if err := reloaded.OpenCreationCharacterList(); err != nil {
		t.Fatal(err)
	}
	if len(reloaded.CreationLibrary) != 1 || reloaded.CreationLibrary[0].Name != "凱蘭德" {
		t.Fatalf("重載角色庫=%v", reloaded.CreationLibrary)
	}
	if err := reloaded.AddSavedCreationCharacter(0); err != nil {
		t.Fatal(err)
	}
	if err := reloaded.AddSavedCreationCharacter(0); err == nil {
		t.Fatal("同一角色不得重複加入")
	}
}

// 答 'N' 的角色不留下來（原作直接 FreeMem 離開）。
func TestGuidedCreationDiscardsOnNo(t *testing.T) {
	state := NewState(testCatalog())
	if err := state.BeginGuidedCreation(); err != nil {
		t.Fatal(err)
	}
	if err := state.SelectGuidedOption(5); err != nil { // 人類
		t.Fatal(err)
	}
	if err := state.SelectGuidedOption(0); err != nil { // 男性
		t.Fatal(err)
	}
	if err := state.SelectGuidedOption(0); err != nil { // 第一個可選職業
		t.Fatal(err)
	}
	if err := state.SelectGuidedOption(0); err != nil { // 第一個可選陣營
		t.Fatal(err)
	}
	if err := state.RollGuidedAbilities(99); err != nil {
		t.Fatal(err)
	}
	if err := state.AcceptGuidedAbilities(99); err != nil {
		t.Fatal(err)
	}
	if err := state.AppendGuidedName([]rune("測試")); err != nil {
		t.Fatal(err)
	}
	if err := state.CommitGuidedName(); err != nil {
		t.Fatal(err)
	}
	if err := state.ConfirmGuidedIcon(false); err != nil {
		t.Fatal(err)
	}
	if err := state.ConfirmGuidedSave(false); err != nil {
		t.Fatal(err)
	}
	if len(state.CreationRoster) != 0 {
		t.Fatalf("放棄的角色不應該進名冊：%v", state.CreationRoster)
	}
	if state.GuidedActive {
		t.Fatal("放棄後應該離開四段流程")
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

// 四段選單用到的每個 locale key 都必須在 assets/locale/zh-TW.json 裡有譯文。
// 這是防漏詞條的檢查——測試用的 catalog 只有幾條字串，所以直接驗證正式檔案。
func TestGuidedCreationLocaleKeysExist(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "assets", "locale", "zh-TW.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := locale.Load(data)
	if err != nil {
		t.Fatal(err)
	}
	keys := []string{
		"creation_pick_race", "creation_pick_gender", "creation_pick_class",
		"creation_pick_alignment", "creation_ability_prompt", "creation_reroll_prompt",
		"creation_guided_menu_hint", "creation_party_empty", "creation_party_member",
		"creation_party_count", "creation_party_help",
		"gender_male", "gender_female",
	}
	for _, race := range guidedRaces {
		keys = append(keys, race.LocaleKey)
	}
	for _, alignment := range guidedAlignments {
		keys = append(keys, alignment.LocaleKey)
	}
	for _, key := range keys {
		if text := catalog.Text(key, ""); text == "" {
			t.Fatalf("locale 缺少建角詞條 %q", key)
		}
	}

	// 職業組合的顯示名在 game pack（作品內容），十五個組合都要有。
	pack, err := gamepack.Default()
	if err != nil {
		t.Fatal(err)
	}
	for combo, key := range classComboKeys {
		text, ok := pack.Text(key, "zh-TW")
		if !ok || text == "" {
			t.Fatalf("game pack 缺少職業組合 %d 的顯示名 %q", combo, key)
		}
	}
}
