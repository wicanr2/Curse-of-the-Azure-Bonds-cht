package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"image/color"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/game"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/gfx"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

const (
	logicalWidth  = 640
	logicalHeight = 400
)

type app struct {
	state        game.State
	face         font.Face
	choiceCursor int
	partyPath    string
	tilePreview  bool
	tileImages   []*ebiten.Image
}

func (a *app) Update() error {
	if a.tilePreview {
		if inpututil.IsKeyJustPressed(ebiten.KeyT) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			a.tilePreview = false
		}
		return nil
	}
	if a.state.Mode == game.ModeCharacterCreation {
		if a.state.CreationEditing {
			if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
				return a.state.CancelCreationName()
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
				return a.state.BackspaceCreationName()
			}
			if chars := ebiten.InputChars(); len(chars) > 0 {
				if err := a.state.AppendCreationName(chars); err != nil {
					a.state.CreationMessage = err.Error()
				}
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
				return a.state.CommitCreationName()
			}
			return nil
		}
		if a.state.CreationEditingAbilities {
			if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
				return a.state.ToggleCreationAbilities()
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
				return a.state.MoveCreationAbility(-1)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
				return a.state.MoveCreationAbility(1)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
				return a.state.AdjustCreationAbility(1)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
				return a.state.AdjustCreationAbility(-1)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
				return a.state.AddCreationCharacter(a.state.CreationCursor)
			}
			return nil
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			return a.state.CancelCharacterCreation()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyN) {
			return a.state.BeginCreationName()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyA) {
			return a.state.ToggleCreationAbilities()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			return a.state.RerollCreationAbilities(time.Now().UnixNano())
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyD) {
			return a.state.FinishCharacterCreation()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			return a.state.AddCreationCharacter(a.state.CreationCursor)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyDown) {
			if a.state.CreationCursor+1 < len(a.state.CreationOptions) {
				a.state.CreationCursor++
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyUp) {
			if a.state.CreationCursor > 0 {
				a.state.CreationCursor--
			}
		}
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		return a.state.OpenCharacterCreation()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyT) {
		a.tilePreview = true
		return nil
	}
	if a.state.Mode == game.ModeJournal {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyJ) {
			return a.state.CloseJournal()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyDown) {
			return a.state.NextJournalPage()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyUp) {
			return a.state.PreviousJournalPage()
		}
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyJ) {
		return a.state.OpenJournal()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF5) {
		if err := a.state.SavePartyFile(a.partyPath); err != nil {
			a.state.Message = "儲存失敗：" + err.Error()
		} else {
			a.state.Message = "隊伍已儲存：" + a.partyPath
		}
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF9) {
		if err := a.state.LoadPartyFile(a.partyPath); err != nil {
			a.state.Message = "載入失敗：" + err.Error()
		} else {
			a.state.Message = "隊伍已載入：" + a.partyPath
			a.choiceCursor = 0
		}
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		switch a.state.Mode {
		case game.ModeTitle:
			return a.state.Apply(game.ActionStart)
		case game.ModeWilderness:
			err := a.state.Select(a.choiceCursor)
			if a.state.Mode == game.ModeWilderness {
				a.choiceCursor = 0
			}
			return err
		case game.ModeMap:
			return a.state.EnterPlaces()
		case game.ModePlace:
			err := a.state.Select(a.choiceCursor)
			if a.state.Mode == game.ModePlace {
				a.choiceCursor = 0
			}
			return err
		case game.ModeEvent:
			return a.state.Continue()
		case game.ModeCombat:
			return a.state.CombatAct()
		}
	}
	if a.state.Mode == game.ModeCombat {
		if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
			return a.state.CombatSelectTarget(1)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
			return a.state.CombatSelectTarget(-1)
		}
	}
	if a.state.Mode == game.ModeMap {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			return a.state.LeaveMap()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
			return a.state.Move(0, -1)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
			return a.state.Move(0, 1)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
			return a.state.Move(-1, 0)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
			return a.state.Move(1, 0)
		}
	}
	if a.state.Mode == game.ModeWilderness || a.state.Mode == game.ModePlace {
		if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyRight) {
			if a.choiceCursor+1 < len(a.state.Choices) {
				a.choiceCursor++
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
			if a.choiceCursor > 0 {
				a.choiceCursor--
			}
		}
	}
	return nil
}

func (a *app) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{12, 18, 42, 255})
	white := color.RGBA{232, 238, 255, 255}
	cyan := color.RGBA{92, 220, 255, 255}
	if a.tilePreview {
		a.drawTilePreview(screen, white, cyan)
		return
	}
	if a.state.Mode == game.ModeCharacterCreation {
		a.drawCreation(screen, white, cyan)
		return
	}
	if a.state.Mode == game.ModeJournal {
		text.Draw(screen, a.state.JournalTitle, a.face, 32, 52, cyan)
		line := 100
		for _, paragraph := range strings.Split(a.state.JournalText, "\n") {
			text.Draw(screen, paragraph, a.face, 32, line, white)
			line += 36
		}
		text.Draw(screen, "第 "+strconv.Itoa(a.state.JournalPage+1)+" / "+strconv.Itoa(len(a.state.JournalPages))+" 頁　左右：翻頁", a.face, 32, 320, white)
		text.Draw(screen, a.state.JournalCloseText, a.face, 32, 350, cyan)
		return
	}
	text.Draw(screen, a.state.Title, a.face, 32, 52, cyan)
	text.Draw(screen, a.state.LocationName, a.face, 32, 90, cyan)
	text.Draw(screen, a.state.Prompt, a.face, 32, 130, white)
	if a.state.Mode == game.ModeWilderness || a.state.Mode == game.ModePlace {
		for index, choice := range a.state.Choices {
			prefix := "  "
			if index == a.choiceCursor {
				prefix = "> "
			}
			text.Draw(screen, prefix+choice, a.face, 56, 220+index*40, white)
		}
		text.Draw(screen, "Enter：選擇", a.face, 56, 330, cyan)
		text.Draw(screen, "F5：儲存隊伍　F9：載入隊伍", a.face, 56, 370, white)
	}
	if a.state.Mode == game.ModeEvent {
		text.Draw(screen, a.state.Message, a.face, 56, 220, cyan)
		text.Draw(screen, "Enter：繼續", a.face, 56, 330, white)
	}
	if a.state.Mode == game.ModePlace {
		for index, choice := range a.state.Choices {
			prefix := "  "
			if index == a.choiceCursor {
				prefix = "> "
			}
			text.Draw(screen, prefix+choice, a.face, 56, 220+index*34, white)
		}
		text.Draw(screen, "Enter：選擇", a.face, 56, 350, cyan)
	}
	if a.state.Mode == game.ModeMap {
		text.Draw(screen, "暗影谷荒野", a.face, 56, 220, cyan)
		text.Draw(screen, "位置：("+strconv.Itoa(a.state.MapX)+", "+strconv.Itoa(a.state.MapY)+")", a.face, 56, 260, white)
		text.Draw(screen, "Enter：場所　方向鍵：移動　Esc：離開", a.face, 56, 330, white)
	}
	if a.state.Mode == game.ModeCombat {
		a.drawCombat(screen, white, cyan)
		return
	}
}

func (a *app) drawTilePreview(screen *ebiten.Image, white, cyan color.Color) {
	text.Draw(screen, "原始圖塊預覽（T／Esc：返回）", a.face, 24, 28, cyan)
	for index, tile := range a.tileImages {
		column := index % 8
		row := index / 8
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(2, 2)
		op.GeoM.Translate(float64(24+column*76), float64(44+row*56))
		screen.DrawImage(tile, op)
		text.Draw(screen, strconv.Itoa(index), a.face, 24+column*76, 44+row*56+50, white)
	}
}

func (a *app) drawCreation(screen *ebiten.Image, white, cyan color.Color) {
	text.Draw(screen, "建立冒險隊伍", a.face, 32, 52, cyan)
	text.Draw(screen, a.state.CreationMessage, a.face, 32, 90, white)
	if a.state.CreationEditing {
		text.Draw(screen, "姓名："+a.state.CreationName+"_", a.face, 48, 140, white)
		text.Draw(screen, "輸入文字　Enter：確定　Esc：取消", a.face, 48, 190, cyan)
		return
	}
	if a.state.CreationEditingAbilities {
		character := a.state.CreationOptions[a.state.CreationCursor]
		text.Draw(screen, "編輯："+character.Name+"（左右選能力，上下調整）", a.face, 32, 135, white)
		for index, name := range []string{"力量", "智力", "智慧", "敏捷", "體質", "魅力"} {
			value, _ := character.Abilities.Value(index)
			prefix := "  "
			if index == a.state.CreationAbility {
				prefix = "> "
			}
			text.Draw(screen, prefix+name+"："+strconv.Itoa(value), a.face, 64, 175+index*25, white)
		}
		text.Draw(screen, "A／Esc：返回　Enter：加入隊伍", a.face, 48, 350, cyan)
		return
	}
	for index, character := range a.state.CreationOptions {
		prefix := "  "
		if index == a.state.CreationCursor {
			prefix = "> "
		}
		label := prefix + character.Name + "（" + raceName(character.Race) + "／" + className(character.Class) + "）"
		text.Draw(screen, label, a.face, 48, 150+index*38, white)
	}
	text.Draw(screen, "已加入："+strconv.Itoa(len(a.state.CreationRoster))+" 人", a.face, 48, 285, cyan)
	text.Draw(screen, "N：改名　A：能力值　R：重擲　Enter：加入　D：完成　Esc：取消", a.face, 48, 340, white)
}

func raceName(r party.Race) string {
	return map[party.Race]string{party.RaceDwarf: "矮人", party.RaceElf: "精靈", party.RaceGnome: "侏儒", party.RaceHalfElf: "半精靈", party.RaceHalfling: "半身人", party.RaceHuman: "人類"}[r]
}

func className(c party.Class) string {
	return map[party.Class]string{party.ClassCleric: "牧師", party.ClassFighter: "戰士", party.ClassRanger: "遊俠", party.ClassPaladin: "聖武士", party.ClassMagicUser: "魔法師", party.ClassThief: "盜賊"}[c]
}

func (a *app) drawCombat(screen *ebiten.Image, white, cyan color.Color) {
	text.Draw(screen, "戰鬥", a.face, 32, 52, cyan)
	text.Draw(screen, a.state.CombatMessage(), a.face, 32, 90, white)
	line := 130
	targets := a.state.CombatTargets()
	for _, fighter := range a.state.CombatFighters() {
		if fighter.Side == 0 {
			text.Draw(screen, fighter.Name+" 生命 "+strconv.Itoa(fighter.HitPoints)+"/"+strconv.Itoa(fighter.MaxHitPoints), a.face, 32, line, white)
			line += 24
		}
	}
	line += 12
	for _, fighter := range a.state.CombatFighters() {
		if fighter.Side != 1 {
			continue
		}
		prefix := "  "
		if len(targets) > 0 && a.state.CombatTargetIndex() < len(targets) && targets[a.state.CombatTargetIndex()].ID == fighter.ID {
			prefix = "> "
		}
		text.Draw(screen, prefix+fighter.Name+" 生命 "+strconv.Itoa(fighter.HitPoints)+"/"+strconv.Itoa(fighter.MaxHitPoints), a.face, 32, line, white)
		line += 24
	}
	text.Draw(screen, "左右：選擇目標　Enter：攻擊", a.face, 32, 350, cyan)
}

func (a *app) Layout(_, _ int) (int, int) { return logicalWidth, logicalHeight }

func loadFace(path string) font.Face {
	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if parsed, err := opentype.Parse(data); err == nil {
				if face, err := opentype.NewFace(parsed, &opentype.FaceOptions{Size: 24, DPI: 72, Hinting: font.HintingFull}); err == nil {
					return face
				}
			}
		}
	}
	return basicfont.Face7x13
}

func main() {
	fontPath := flag.String("font", "", "TrueType/OpenType font path; required for Chinese glyphs")
	localePath := flag.String("locale", "assets/locale/zh-TW.json", "locale JSON path")
	imagePath := flag.String("image", "curseoftheazurebonds.zip", "original DOS image ZIP")
	encounter := flag.Bool("encounter", false, "start the observed ECL1 encounter directly")
	encounterBlock := flag.Int("encounter-block", 81, "ECL block for -encounter")
	encounterStart := flag.Int("encounter-start", 0x1293, "payload offset for -encounter")
	partyPath := flag.String("party-save", "party.json", "versioned remake party save path")
	partyLoadPath := flag.String("party-load", "", "load a versioned remake party save before starting")
	flag.Parse()
	data, err := os.ReadFile(*localePath)
	if err != nil {
		log.Fatal(err)
	}
	catalog, err := locale.Load(data)
	if err != nil {
		log.Fatal(err)
	}
	eclBlocks, initialECL, err := loadECLBlocks(*imagePath)
	if err != nil {
		log.Fatal(err)
	}
	state := game.NewStateFromECLBlocks(catalog, eclBlocks, initialECL)
	monsterData, err := zipMember(*imagePath, "MON1CHA.DAX")
	if err != nil {
		log.Fatal(err)
	}
	monsterRecords, err := loadMonsterRecords(monsterData)
	if err != nil {
		log.Fatal(err)
	}
	state.SetMonsterRecords(monsterRecords)
	if *partyLoadPath != "" {
		if err := state.LoadPartyFile(*partyLoadPath); err != nil {
			log.Fatal(err)
		}
	}
	tileImages, err := loadTileImages(*imagePath)
	if err != nil {
		log.Fatal(err)
	}
	if *encounter {
		block, ok := eclBlocks[uint8(*encounterBlock)]
		if !ok {
			log.Fatalf("ECL block 0x%02X is unavailable", *encounterBlock)
		}
		result, runErr := ecl.RunSubset(block, *encounterStart, 128)
		if runErr != nil {
			log.Fatal(runErr)
		}
		if err := state.StartEncounter(result, monsterRecords, demoParty(), 37); err != nil {
			log.Fatal(err)
		}
	}
	ebiten.SetWindowSize(logicalWidth, logicalHeight)
	ebiten.SetWindowTitle(catalog.Text("title", "Curse of the Azure Bonds"))
	if err := ebiten.RunGame(&app{state: state, face: loadFace(*fontPath), partyPath: *partyPath, tileImages: tileImages}); err != nil {
		log.Fatal(err)
	}
}

func loadTileImages(imagePath string) ([]*ebiten.Image, error) {
	data, err := zipMember(imagePath, "TILES.DAX")
	if err != nil {
		return nil, err
	}
	blocks, err := dax.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse TILES.DAX: %w", err)
	}
	images := make([]*ebiten.Image, 0)
	for _, block := range blocks {
		picture, err := gfx.ParsePicture(block.Data, false, 0)
		if err != nil {
			return nil, fmt.Errorf("TILES.DAX block 0x%02X: %w", block.Entry.ID, err)
		}
		for item := 0; item < int(picture.ItemCount); item++ {
			rgba, err := picture.RGBA(item, gfx.EGA16)
			if err != nil {
				return nil, fmt.Errorf("TILES.DAX block 0x%02X item %d: %w", block.Entry.ID, item, err)
			}
			images = append(images, ebiten.NewImageFromImage(rgba))
		}
	}
	return images, nil
}

// loadECLBlocks builds one session namespace from all six original ECL DAX
// members. NEWECL operands are global block IDs (for example ECL4 can jump to
// ECL1 block 0x50), so loading only ECL1 would make real transitions fail at
// the loader boundary.
func loadECLBlocks(imagePath string) (map[uint8][]byte, uint8, error) {
	all := make(map[uint8][]byte)
	var initial uint8
	for index := 1; index <= 6; index++ {
		member := fmt.Sprintf("ECL%d.DAX", index)
		data, err := zipMember(imagePath, member)
		if err != nil {
			return nil, 0, fmt.Errorf("load %s: %w", member, err)
		}
		blocks, err := dax.Parse(data)
		if err != nil || len(blocks) == 0 {
			if err == nil {
				err = fmt.Errorf("contains no blocks")
			}
			return nil, 0, fmt.Errorf("parse %s: %w", member, err)
		}
		if index == 1 {
			initial = blocks[0].Entry.ID
		}
		for _, block := range blocks {
			if _, exists := all[block.Entry.ID]; exists {
				return nil, 0, fmt.Errorf("duplicate ECL block ID 0x%02X while loading %s", block.Entry.ID, member)
			}
			all[block.Entry.ID] = block.Data
		}
	}
	return all, initial, nil
}

func loadMonsterRecords(data []byte) (map[uint8]monster.Record, error) {
	blocks, err := dax.Parse(data)
	if err != nil {
		return nil, err
	}
	records := make(map[uint8]monster.Record, len(blocks))
	for _, block := range blocks {
		record, err := monster.Parse(block.Data)
		if err != nil {
			return nil, fmt.Errorf("MON1CHA block 0x%02X: %w", block.Entry.ID, err)
		}
		records[block.Entry.ID] = record
	}
	return records, nil
}

// demoParty is deliberately an explicit debug roster for -encounter. The
// original party save/creation data is a separate reverse-engineering task;
// normal startup still uses the opening state and does not silently invent it.
func demoParty() []combat.Fighter {
	return []combat.Fighter{
		{ID: "party-1", Name: "戰士", Side: combat.SideParty, HitPoints: 42, MaxHitPoints: 42, ArmorClass: 4, AttackBonus: 16, DamageDiceCount: 1, DamageDiceSides: 8, InitiativeBonus: 1},
		{ID: "party-2", Name: "遊俠", Side: combat.SideParty, HitPoints: 34, MaxHitPoints: 34, ArmorClass: 5, AttackBonus: 15, DamageDiceCount: 1, DamageDiceSides: 8, InitiativeBonus: 2},
		{ID: "party-3", Name: "牧師", Side: combat.SideParty, HitPoints: 30, MaxHitPoints: 30, ArmorClass: 6, AttackBonus: 12, DamageDiceCount: 1, DamageDiceSides: 6, InitiativeBonus: 0},
		{ID: "party-4", Name: "法師", Side: combat.SideParty, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 8, AttackBonus: 10, DamageDiceCount: 1, DamageDiceSides: 4, InitiativeBonus: 1},
	}
}

func zipMember(path, member string) ([]byte, error) {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer archive.Close()
	for _, file := range archive.File {
		if file.Name != member {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		return io.ReadAll(reader)
	}
	return nil, os.ErrNotExist
}
