package main

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/area"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dungeon"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/game"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/gfx"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/mapdata"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/sound"
)

const (
	logicalWidth  = 640
	logicalHeight = 480
)

type app struct {
	state               game.State
	imagePath           string
	face                font.Face
	compactFace         font.Face
	choiceCursor        int
	partyPath           string
	savgamDir           string
	savgamSlot          byte
	savgamSlotSave      bool
	tilePreview         bool
	tileImages          []*ebiten.Image
	geoPreview          bool
	geoGrid             *geo.Grid
	geoX                int
	geoY                int
	geoLabel            string
	geoCatalog          geo.Catalog
	geoSet              uint8
	geoBlock            uint8
	dungeonPreview      bool
	dungeonFloor        *mapdata.DungeonFloor
	dungeonX            int
	dungeonY            int
	dungeonDoorMenu     bool
	pieceSets           map[uint8]gfx.PieceSet
	pieceLabel          string
	wallPreview         []wallPreviewStamp
	combatSprites       map[string]*ebiten.Image
	combatSpriteIDs     []string
	combatAnimations    map[string][]combatAnimation
	animationStart      time.Time
	deathOverlayStarted map[string]time.Time
	messageSnapshot     string
	messageStart        time.Time
	soundPlayer         *sound.Player
}

type combatAnimation struct {
	image *ebiten.Image
	delay uint32
	x     int16
	y     int16
}

type wallPreviewStamp struct {
	image  *ebiten.Image
	row    int
	column int
}

func (a *app) combatAction(action func() error) error {
	if err := action(); err != nil {
		a.state.ReportCombatError(err)
	}
	return nil
}

func (a *app) playSound(id sound.ID) {
	if a.soundPlayer != nil {
		a.soundPlayer.Play(id)
	}
}

func (a *app) syncSoundEvents() {
	for _, event := range a.state.ConsumeSoundEvents() {
		a.playSound(sound.ID(event))
	}
}

func (a *app) saveCurrentGame() error {
	if a.savgamSlotSave {
		return a.state.SaveSAVGAMSlot(a.savgamDir, a.savgamSlot)
	}
	return a.state.SavePartyFile(a.partyPath)
}

func (a *app) saveTarget() string {
	if a.savgamSlotSave {
		return filepath.Join(a.savgamDir, fmt.Sprintf("savgam%c.dat", a.savgamSlot+('a'-'A')))
	}
	return a.partyPath
}

func (a *app) Update() error {
	a.syncSoundEvents()
	a.syncGeoMapRequest()
	a.syncLoadPiecesRequest()
	a.syncECLCallRequests()
	if a.tilePreview {
		if inpututil.IsKeyJustPressed(ebiten.KeyT) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			a.tilePreview = false
		}
		return nil
	}
	if a.geoPreview {
		if inpututil.IsKeyJustPressed(ebiten.KeyG) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			a.geoPreview = false
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyUp) && a.geoGrid.CanMove(a.geoX, a.geoY, 0) {
			a.geoY--
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyRight) && a.geoGrid.CanMove(a.geoX, a.geoY, 2) {
			a.geoX++
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyDown) && a.geoGrid.CanMove(a.geoX, a.geoY, 4) {
			a.geoY++
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyLeft) && a.geoGrid.CanMove(a.geoX, a.geoY, 6) {
			a.geoX--
		}
		return nil
	}
	if a.dungeonPreview || a.state.Mode == game.ModeDungeon {
		productionDungeon := a.state.Mode == game.ModeDungeon
		if !productionDungeon && (inpututil.IsKeyJustPressed(ebiten.KeyD) || inpututil.IsKeyJustPressed(ebiten.KeyEscape)) {
			a.dungeonDoorMenu = false
			a.dungeonPreview = false
		}
		if a.dungeonDoorMenu {
			a.updateDungeonDoorMenu()
			return nil
		}
		if productionDungeon {
			if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
				a.moveDungeonForward()
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyE) {
				return a.state.EnterDungeonCamp()
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyS) {
				return a.state.SearchDungeonLocation()
			}
		} else {
			if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
				a.moveDungeonPreview(0, -1, 0)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
				a.moveDungeonPreview(1, 0, 2)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
				a.moveDungeonPreview(0, 1, 4)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
				a.moveDungeonPreview(-1, 0, 6)
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyQ) || inpututil.IsKeyJustPressed(ebiten.KeyK) {
			a.turnDungeonGeometry(-2)
			a.prepareWallPreview()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyR) || inpututil.IsKeyJustPressed(ebiten.KeyM) {
			a.turnDungeonGeometry(2)
			a.prepareWallPreview()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyP) {
			a.tryDungeonPickLock()
		}
		if (productionDungeon && inpututil.IsKeyJustPressed(ebiten.KeyN)) ||
			(!productionDungeon && inpututil.IsKeyJustPressed(ebiten.KeyK)) {
			a.tryDungeonKnock()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyB) {
			a.tryDungeonBash()
		}
		return nil
	}
	if a.state.RenameEditing() {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			return a.state.CancelRename()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
			return a.state.BackspaceRenameName()
		}
		if chars := ebiten.InputChars(); len(chars) > 0 {
			if err := a.state.AppendRenameName(chars); err != nil {
				a.state.Message = err.Error()
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			return a.state.CommitRename()
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
	if inpututil.IsKeyJustPressed(ebiten.KeyC) && a.state.Mode != game.ModeCombat {
		return a.state.OpenCharacterCreation()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyT) {
		a.tilePreview = true
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyG) && a.geoGrid != nil {
		a.geoPreview = true
		a.geoX, a.geoY = 0, 0
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) && a.dungeonFloor != nil {
		a.dungeonPreview = true
		a.refreshDungeonPreview()
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
		if err := a.saveCurrentGame(); err != nil {
			a.state.Message = "儲存失敗：" + err.Error()
		} else {
			a.state.Message = "隊伍已儲存：" + a.saveTarget()
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
			if a.state.ConsumeSaveRequest() {
				if saveErr := a.saveCurrentGame(); saveErr != nil {
					a.state.Message = "儲存失敗：" + saveErr.Error()
				} else {
					a.state.Message = "隊伍已儲存：" + a.saveTarget()
				}
			}
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
			if a.state.CombatViewActive() {
				a.state.EndCombatView()
				return nil
			}
			if a.state.CombatCastingSpell() != 0 {
				return a.combatAction(func() error { return a.state.CombatCast(a.state.CombatCastingSpell()) })
			}
			return a.combatAction(a.state.CombatAct)
		}
	}
	if a.state.Mode == game.ModeCombat {
		if a.state.CombatViewActive() {
			if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
				a.state.EndCombatView()
			}
			return nil
		}
		if a.state.CombatMoveMode() {
			if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
				a.state.CancelCombatMove()
				return nil
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
				return a.combatAction(func() error { return a.state.CombatMove(0, -1) })
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
				return a.combatAction(func() error { return a.state.CombatMove(1, 0) })
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
				return a.combatAction(func() error { return a.state.CombatMove(0, 1) })
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
				return a.combatAction(func() error { return a.state.CombatMove(-1, 0) })
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) && a.state.CombatCastingSpell() != 0 {
			a.state.CancelCombatCast()
			return nil
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyS) && a.state.CombatCanCastMagicMissile() {
			return a.combatAction(func() error { return a.state.BeginCombatCast(game.MagicMissileSpellID) })
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyH) && a.state.CombatCanCastCureLightWounds() {
			return a.combatAction(func() error { return a.state.BeginCombatCast(game.CureLightWoundsSpellID) })
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyB) && a.state.CombatCanCastBless() {
			return a.combatAction(func() error { return a.state.BeginCombatCast(game.BlessSpellID) })
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyC) && a.state.CombatCanCastCurse() {
			return a.combatAction(func() error { return a.state.BeginCombatCast(game.CurseSpellID) })
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyW) && a.state.CombatCanCastCauseLightWounds() {
			return a.combatAction(func() error { return a.state.BeginCombatCast(game.CauseLightWoundsSpellID) })
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyP) && a.state.CombatCanCastProtectionFromEvil() {
			return a.combatAction(func() error { return a.state.BeginCombatCast(game.ProtectionFromEvilSpellID) })
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyG) && a.state.CombatCanCastProtectionFromGood() {
			return a.combatAction(func() error { return a.state.BeginCombatCast(game.ProtectionFromGoodSpellID) })
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyM) {
			return a.combatAction(a.state.BeginCombatMove)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyD) {
			return a.combatAction(a.state.CombatDone)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyV) {
			return a.combatAction(a.state.BeginCombatView)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
			if a.state.CombatCastingSpell() != 0 {
				return a.combatAction(func() error { return a.state.CombatSelectSpellTarget(1) })
			}
			return a.combatAction(func() error { return a.state.CombatSelectTarget(1) })
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
			if a.state.CombatCastingSpell() != 0 {
				return a.combatAction(func() error { return a.state.CombatSelectSpellTarget(-1) })
			}
			return a.combatAction(func() error { return a.state.CombatSelectTarget(-1) })
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

func (a *app) syncLoadPiecesRequest() {
	selectors, ok := a.state.ConsumeLoadPiecesRequest()
	if !ok {
		return
	}
	sets, err := loadMapPieceSets(a.imagePath, a.state.GeoMapSet, selectors)
	if err != nil {
		// Asset diagnostics belong to the renderer label. ECL event text is
		// authoritative story state and must survive a missing optional wall
		// selector (the wizard-tower PICTURE is still independently usable).
		a.pieceLabel = "LOAD PIECES 未載入：" + err.Error()
		return
	}
	a.pieceSets = sets
	a.prepareWallPreview()
	a.pieceLabel = fmt.Sprintf("LOAD PIECES [%d,%d,%d]：WALLDEF／8X8D 已載入", selectors[0], selectors[1], selectors[2])
}

func (a *app) syncGeoMapRequest() {
	set, block, ok := a.state.ConsumeGeoMapRequest()
	if !ok {
		return
	}
	grid, found := a.geoCatalog.Lookup(geo.MapRef{Set: set, BlockID: block})
	if !found {
		a.state.Message = fmt.Sprintf("找不到 GEO%d block 0x%02X", set, block)
		return
	}
	a.geoGrid = &grid
	a.dungeonX, a.dungeonY, _ = a.state.DungeonGeometryView()
	if a.dungeonX < 0 || a.dungeonX >= geo.Width || a.dungeonY < 0 || a.dungeonY >= geo.Height {
		a.dungeonX, a.dungeonY = 7, 13
	}
	a.refreshDungeonPreview()
	a.geoSet, a.geoBlock = set, block
	a.geoLabel = fmt.Sprintf("GEO%d block 0x%02X", set, block)
}

func (a *app) refreshDungeonPreview() {
	if a.geoGrid == nil {
		return
	}
	floor := mapdata.GenerateDungeon(*a.geoGrid, a.dungeonX, a.dungeonY)
	a.dungeonFloor = &floor
	a.prepareWallPreview()
}

func (a *app) turnDungeonGeometry(delta int) {
	x, y, direction := a.state.DungeonGeometryView()
	direction = uint8((int(direction) + delta + 8) % 8)
	a.state.SetDungeonGeometryView(x, y, direction)
}

func (a *app) syncECLCallRequests() {
	for _, address := range a.state.ConsumeECLCallRequests() {
		switch address {
		case 0x2E10, 0xC01E:
			a.dungeonX, a.dungeonY = a.state.DungeonX, a.state.DungeonY
			a.refreshDungeonPreview()
		}
	}
}

func (a *app) moveDungeonPreview(dx, dy, direction int) {
	if a.geoGrid == nil || !a.geoGrid.CanMoveDungeonWrapped(a.dungeonX, a.dungeonY, direction) {
		if flags, ok := a.dungeonDoorFlags(); ok && (flags == 2 || flags == 3) {
			a.dungeonDoorMenu = true
			a.state.Message = "門已上鎖，請選擇 Bash／Pick／Knock／Exit"
		}
		return
	}
	nextX, nextY := a.dungeonX+dx, a.dungeonY+dy
	exitAttempt := nextX < 0 || nextX >= geo.Width || nextY < 0 || nextY >= geo.Height
	a.dungeonX = geo.WrapCoordinate(nextX, geo.Width)
	a.dungeonY = geo.WrapCoordinate(nextY, geo.Height)
	a.state.SetDungeonGeometryView(a.dungeonX, a.dungeonY, uint8(direction))
	a.refreshDungeonPreview()
	a.playSound(sound.Step)
	if a.state.Mode == game.ModeDungeon {
		var err error
		if exitAttempt {
			err = a.state.RunDungeonExitLifecycle()
		} else {
			err = a.state.RunDungeonLifecycle()
		}
		if err != nil {
			a.state.Message = "地城事件執行失敗：" + err.Error()
		}
	}
}

func (a *app) moveDungeonForward() {
	_, _, direction := a.state.DungeonGeometryView()
	switch direction {
	case 0:
		a.moveDungeonPreview(0, -1, 0)
	case 2:
		a.moveDungeonPreview(1, 0, 2)
	case 4:
		a.moveDungeonPreview(0, 1, 4)
	case 6:
		a.moveDungeonPreview(-1, 0, 6)
	}
}

func (a *app) updateDungeonDoorMenu() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		a.dungeonDoorMenu = false
		return
	}
	flags, ok := a.dungeonDoorFlags()
	if !ok {
		a.dungeonDoorMenu = false
		return
	}
	options := a.state.DungeonDoorMenuOptions(flags)
	if inpututil.IsKeyJustPressed(ebiten.KeyB) && options.Bash {
		a.tryDungeonBash()
		a.dungeonDoorMenu = false
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyP) && options.Pick {
		a.tryDungeonPickLock()
		// Pick is one-shot even when it fails; remaining actions may still be
		// selected, matching can_pick_door=false in locked_door.
		if flags, ok := a.dungeonDoorFlags(); !ok || flags != 2 {
			a.dungeonDoorMenu = false
		}
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyK) && options.Knock {
		a.tryDungeonKnock()
		a.dungeonDoorMenu = false
	}
}

func (a *app) prepareWallPreview() {
	a.wallPreview = nil
	if a.geoGrid == nil {
		return
	}
	a.syncDungeonWallState()
	if len(a.pieceSets) == 0 {
		return
	}
	_, _, direction := a.state.DungeonGeometryView()
	view, err := gfx.TraverseWallViewWrapped(*a.geoGrid, direction, a.dungeonX, a.dungeonY)
	if err != nil {
		return
	}
	for _, call := range view.Calls {
		setID := uint8((call.WallType-1)/5 + 1)
		piece, ok := a.pieceSets[setID]
		if !ok {
			continue
		}
		stamps, err := gfx.BuildWallLayout(piece, call.WallType, call.Layout, call.RowStart, call.ColStart)
		if err != nil {
			continue
		}
		for _, stamp := range stamps {
			rgba, err := stamp.Picture.RGBA(stamp.Item, gfx.EGA16)
			if err != nil {
				continue
			}
			a.wallPreview = append(a.wallPreview, wallPreviewStamp{
				image:  ebiten.NewImageFromImage(rgba),
				row:    stamp.Row,
				column: stamp.Column,
			})
		}
	}
}

func (a *app) syncDungeonWallState() {
	if a.geoGrid == nil {
		return
	}
	_, _, direction := a.state.DungeonGeometryView()
	wall, _ := a.geoGrid.WallWrapped(a.dungeonX, a.dungeonY, int(direction))
	cell := a.geoGrid.CellWrapped(a.dungeonX, a.dungeonY)
	a.state.DungeonWallType = wall
	a.state.DungeonWallRoof = cell.Terrain
}

func (a *app) dungeonDoorFlags() (uint8, bool) {
	if a.geoGrid == nil {
		return 0, false
	}
	_, _, direction := a.state.DungeonGeometryView()
	return a.geoGrid.WallDoorFlagsWrapped(a.dungeonX, a.dungeonY, int(direction))
}

func (a *app) tryDungeonPickLock() {
	flags, ok := a.dungeonDoorFlags()
	if !ok || flags != 2 {
		a.state.Message = "目前門面不可撬鎖（只有 detail 2 可撬鎖）"
		return
	}
	result := a.state.PickDungeonLock()
	_, _, direction := a.state.DungeonGeometryView()
	if result.Opened && a.geoGrid.UnlockDoorWrapped(a.dungeonX, a.dungeonY, int(direction)) {
		a.state.Message = "撬鎖成功，門已雙側解鎖"
		a.refreshDungeonPreview()
		return
	}
	a.state.Message = "撬鎖失敗，本次撬鎖機會已消耗"
}

func (a *app) tryDungeonKnock() {
	flags, ok := a.dungeonDoorFlags()
	if !ok || (flags != 2 && flags != 3) {
		a.state.Message = "目前沒有可施放 Knock 的上鎖門面"
		return
	}
	if !a.state.ConsumeDungeonKnockSpell() {
		a.state.Message = fmt.Sprintf("沒有可用的 Knock（0x%02X）", dungeon.KnockSpellID)
		return
	}
	_, _, direction := a.state.DungeonGeometryView()
	if a.geoGrid.UnlockDoorWrapped(a.dungeonX, a.dungeonY, int(direction)) {
		a.state.Message = "Knock 成功，門已雙側解鎖"
		a.refreshDungeonPreview()
	}
}

func (a *app) tryDungeonBash() {
	flags, ok := a.dungeonDoorFlags()
	if !ok || (flags != 2 && flags != 3) {
		a.state.Message = "目前沒有可撞擊的上鎖門面"
		return
	}
	result := a.state.BashDungeonDoor(flags)
	_, _, direction := a.state.DungeonGeometryView()
	if result.Opened && a.geoGrid.UnlockDoorWrapped(a.dungeonX, a.dungeonY, int(direction)) {
		a.state.Message = "撞門成功，門已雙側解鎖"
		a.refreshDungeonPreview()
		return
	}
	a.state.Message = "撞門失敗"
}

func (a *app) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{12, 18, 42, 255})
	white := color.RGBA{232, 238, 255, 255}
	cyan := color.RGBA{92, 220, 255, 255}
	if a.tilePreview {
		a.drawTilePreview(screen, white, cyan)
		return
	}
	if a.geoPreview {
		a.drawGeoPreview(screen, white, cyan)
		return
	}
	if a.dungeonPreview || a.state.Mode == game.ModeDungeon {
		a.drawDungeonPreview(screen, white, cyan)
		return
	}
	if a.state.RenameEditing() {
		text.Draw(screen, a.state.Title, a.face, 32, 52, cyan)
		text.Draw(screen, a.state.Prompt, a.face, 32, 130, white)
		text.Draw(screen, "名稱："+a.state.RenameText()+"_", a.face, 56, 220, cyan)
		text.Draw(screen, "Enter：確認　Backspace：刪除　Esc：取消", a.face, 56, 330, white)
		return
	}
	if a.state.Mode == game.ModeCharacterCreation {
		a.drawCreation(screen, white, cyan)
		return
	}
	if a.state.Mode == game.ModeJournal {
		text.Draw(screen, a.state.JournalTitle, a.face, 32, 52, cyan)
		drawWrappedText(screen, a.state.JournalText, a.face, 32, 100, 22, 32, 7, white)
		text.Draw(screen, "第 "+strconv.Itoa(a.state.JournalPage+1)+" / "+strconv.Itoa(len(a.state.JournalPages))+" 頁　左右：翻頁", a.face, 32, 350, white)
		text.Draw(screen, a.state.JournalCloseText, a.face, 32, 390, cyan)
		return
	}
	if a.state.Mode == game.ModeMap {
		a.drawWildernessMap(screen, white, cyan)
		return
	}
	if a.state.Mode == game.ModeEvent && a.state.PictureRequested {
		a.drawPictureAnimation(screen)
		return
	}
	if a.state.Mode == game.ModeCombat {
		a.drawCombat(screen, white, cyan)
		return
	}
	text.Draw(screen, a.state.Title, a.face, 32, 52, cyan)
	text.Draw(screen, a.state.LocationName, a.face, 32, 90, cyan)
	text.Draw(screen, a.state.Prompt, a.face, 32, 130, white)
	text.Draw(screen, a.state.GameTimeText(), a.face, 32, 170, cyan)
	if a.state.Mode == game.ModeWilderness || a.state.Mode == game.ModePlace {
		choiceTop := 220
		if a.state.Mode == game.ModeWilderness && a.state.Message != "" {
			drawWrappedText(screen, a.revealedMessage(), a.face, 32, 210, 22, 30, 4, cyan)
			choiceTop = 350
		}
		for index, choice := range a.state.Choices {
			prefix := "  "
			if index == a.choiceCursor {
				prefix = "> "
			}
			text.Draw(screen, prefix+choice, a.face, 56, choiceTop+index*40, white)
		}
		if a.state.Message == "" {
			text.Draw(screen, "Enter：選擇", a.face, 56, 330, cyan)
			text.Draw(screen, "F5：儲存隊伍　F9：載入隊伍", a.face, 56, 370, white)
		}
	}
	if a.state.Mode == game.ModeEvent {
		drawWrappedText(screen, a.revealedMessage(), a.face, 56, 210, 22, 32, 5, cyan)
		text.Draw(screen, "Enter：繼續", a.face, 56, 410, white)
	}
	if a.state.Mode == game.ModeMap {
		text.Draw(screen, "暗影谷荒野", a.face, 56, 220, cyan)
		text.Draw(screen, "位置：("+strconv.Itoa(a.state.MapX)+", "+strconv.Itoa(a.state.MapY)+")", a.face, 56, 260, white)
		text.Draw(screen, "Enter：場所　方向鍵：移動　Esc：離開", a.face, 56, 330, white)
	}
}

func drawWrappedText(screen *ebiten.Image, value string, face font.Face, x, y, lineRunes, lineHeight, maxLines int, ink color.Color) {
	lines := wrapTextLines(value, lineRunes, maxLines)
	for index, line := range lines {
		text.Draw(screen, line, face, x, y+index*lineHeight, ink)
	}
}

func wrapTextLines(value string, lineRunes, maxLines int) []string {
	if lineRunes < 1 || maxLines < 1 {
		return nil
	}
	lines := make([]string, 0, maxLines)
	for _, paragraph := range strings.Split(value, "\n") {
		runes := []rune(paragraph)
		if len(runes) == 0 {
			lines = append(lines, "")
			continue
		}
		for len(runes) > 0 {
			count := lineRunes
			if len(runes) < count {
				count = len(runes)
			}
			lines = append(lines, string(runes[:count]))
			runes = runes[count:]
		}
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	return lines
}

func (a *app) revealedMessage() string {
	if a.messageSnapshot != a.state.Message {
		a.messageSnapshot = a.state.Message
		a.messageStart = time.Now()
	}
	runes := []rune(a.state.Message)
	if len(runes) == 0 {
		return ""
	}
	speed := a.state.MessageSpeed()
	if speed < 1 {
		speed = 1
	}
	interval := time.Duration(120/speed) * time.Millisecond
	count := int(time.Since(a.messageStart) / interval)
	if count >= len(runes) {
		return a.state.Message
	}
	return string(runes[:count])
}

func (a *app) drawPictureAnimation(screen *ebiten.Image) {
	if a.state.SceneCharacterRequested {
		key := fmt.Sprintf("character-area-%d-head-%02X-body-%02X.png", a.state.Area.GameArea, a.state.SceneHeadBlock, a.state.SceneBodyBlock)
		if sprite := a.combatSprites[key]; sprite != nil {
			a.drawAdventureChrome(screen)
			const pixelScale = 2
			op := &ebiten.DrawImageOptions{}
			op.Filter = ebiten.FilterNearest
			op.GeoM.Scale(pixelScale, pixelScale)
			op.GeoM.Translate(float64(16+(256-sprite.Bounds().Dx()*pixelScale)/2), float64(16+(256-sprite.Bounds().Dy()*pixelScale)/2))
			screen.DrawImage(sprite, op)
			a.drawPictureMessage(screen)
			text.Draw(screen, "Enter：繼續", a.compactFace, 24, 470, color.RGBA{255, 255, 255, 255})
			return
		}
		text.Draw(screen, "人物圖層素材尚未載入", a.face, 56, 220, color.RGBA{255, 220, 100, 255})
		text.Draw(screen, "Enter：繼續", a.face, 56, 330, color.RGBA{255, 255, 255, 255})
		return
	}
	if a.state.BigPictureRequested {
		key := fmt.Sprintf("bigpic%d-block-%02X-item-00.png", a.state.Area.GameArea, a.state.PictureBlock)
		if sprite := a.combatSprites[key]; sprite != nil {
			const pixelScale = 2
			op := &ebiten.DrawImageOptions{}
			op.Filter = ebiten.FilterNearest
			op.GeoM.Scale(pixelScale, pixelScale)
			op.GeoM.Translate(float64((logicalWidth-sprite.Bounds().Dx()*pixelScale)/2), 44)
			screen.DrawImage(sprite, op)
			a.drawPictureMessage(screen)
			text.Draw(screen, "大幅事件畫面　Enter：繼續", a.face, 56, 446, color.RGBA{255, 255, 255, 255})
			return
		}
		text.Draw(screen, "大幅事件圖片素材尚未載入", a.face, 56, 220, color.RGBA{255, 220, 100, 255})
		text.Draw(screen, "Enter：繼續", a.face, 56, 330, color.RGBA{255, 255, 255, 255})
		return
	}
	key := fmt.Sprintf("pic%d-block-%02X", a.state.Area.GameArea, a.state.PictureBlock)
	frames := a.combatAnimations[key]
	if len(frames) == 0 {
		text.Draw(screen, "事件圖片素材尚未載入", a.face, 56, 220, color.RGBA{255, 220, 100, 255})
		text.Draw(screen, "Enter：繼續", a.face, 56, 330, color.RGBA{255, 255, 255, 255})
		return
	}
	frame := frames[0]
	if a.state.AnimationsEnabled() {
		frame = animationFrame(frames, time.Since(a.animationStart))
	}
	op := &ebiten.DrawImageOptions{}
	a.drawAdventureChrome(screen)
	const pixelScale = 2
	op.Filter = ebiten.FilterNearest
	op.GeoM.Scale(pixelScale, pixelScale)
	op.GeoM.Translate(float64(16+(256-frame.image.Bounds().Dx()*pixelScale)/2)+float64(frame.x*pixelScale), float64(16+(256-frame.image.Bounds().Dy()*pixelScale)/2)+float64(frame.y*pixelScale))
	screen.DrawImage(frame.image, op)
	a.drawPictureMessage(screen)
	text.Draw(screen, "Enter：繼續", a.compactFace, 24, 470, color.RGBA{255, 255, 255, 255})
}

func (a *app) drawPictureMessage(screen *ebiten.Image) {
	runes := []rune(a.revealedMessage())
	// A 24px CJK glyph is much wider than the original 8px Latin cell.
	// Keep picture captions to 22 Unicode code points so both Traditional
	// Chinese and mixed ASCII text stay inside the 640px logical canvas.
	const lineRunes = 24
	for line := 0; line < 4 && len(runes) > 0; line++ {
		count := lineRunes
		if len(runes) < count {
			count = len(runes)
		}
		text.Draw(screen, string(runes[:count]), a.face, 24, 312+line*30, color.RGBA{92, 220, 255, 255})
		runes = runes[count:]
	}
}

func drawPanelFrame(screen *ebiten.Image, x, y, width, height int) {
	edge := color.RGBA{R: 214, G: 216, B: 208, A: 255}
	shadow := color.RGBA{R: 92, G: 98, B: 112, A: 255}
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(width), 3, edge)
	ebitenutil.DrawRect(screen, float64(x), float64(y+height-3), float64(width), 3, edge)
	ebitenutil.DrawRect(screen, float64(x), float64(y), 3, float64(height), edge)
	ebitenutil.DrawRect(screen, float64(x+width-3), float64(y), 3, float64(height), edge)
	ebitenutil.DrawRect(screen, float64(x+4), float64(y+4), float64(width-8), 2, shadow)
}

func (a *app) drawAdventureChrome(screen *ebiten.Image) {
	drawPanelFrame(screen, 8, 8, 264, 272)
	drawPanelFrame(screen, 272, 8, 360, 272)
	drawPanelFrame(screen, 8, 280, 624, 168)
	drawPanelFrame(screen, 8, 448, 624, 28)
	text.Draw(screen, "姓名", a.compactFace, 288, 38, color.RGBA{232, 238, 255, 255})
	text.Draw(screen, "AC", a.compactFace, 520, 38, color.RGBA{232, 238, 255, 255})
	text.Draw(screen, "HP", a.compactFace, 578, 38, color.RGBA{232, 238, 255, 255})
	for index, fighter := range a.state.PartyFighters() {
		if index >= 8 {
			break
		}
		ink := color.RGBA{R: 92, G: 220, B: 255, A: 255}
		text.Draw(screen, fighter.Name, a.compactFace, 288, 68+index*25, ink)
		text.Draw(screen, strconv.Itoa(fighter.ArmorClass), a.compactFace, 526, 68+index*25, ink)
		text.Draw(screen, strconv.Itoa(fighter.HitPoints), a.compactFace, 580, 68+index*25, ink)
	}
}

func (a *app) drawWildernessMap(screen *ebiten.Image, white, cyan color.Color) {
	text.Draw(screen, "暗影谷荒野（原版 50×25 floor）", a.face, 24, 28, cyan)
	text.Draw(screen, a.state.GameTimeText(), a.face, 390, 55, cyan)
	const (
		viewWidth  = 7
		viewHeight = 5
		scale      = 2
		tileSize   = 24
		originX    = 24
		originY    = 48
	)
	for row := 0; row < viewHeight; row++ {
		for column := 0; column < viewWidth; column++ {
			x := a.state.MapX + column - viewWidth/2
			y := a.state.MapY + row - viewHeight/2
			entry, ok := a.state.WildernessFloor.Entry(x, y)
			left := originX + column*tileSize*scale
			top := originY + row*tileSize*scale
			if ok && int(entry.TileIndex) < len(a.tileImages) {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Scale(scale, scale)
				op.GeoM.Translate(float64(left), float64(top))
				screen.DrawImage(a.tileImages[entry.TileIndex], op)
			} else {
				ebitenutil.DrawRect(screen, float64(left), float64(top), tileSize*scale, tileSize*scale, color.RGBA{24, 30, 48, 255})
			}
			if x == a.state.MapX && y == a.state.MapY {
				ebitenutil.DrawRect(screen, float64(left+tileSize*scale/2-4), float64(top+tileSize*scale/2-4), 8, 8, color.RGBA{255, 230, 80, 255})
			}
		}
	}
	text.Draw(screen, "位置：（"+strconv.Itoa(a.state.MapX)+"，"+strconv.Itoa(a.state.MapY)+"）", a.face, 390, 90, white)
	if entry, ok := a.state.WildernessFloor.Entry(a.state.MapX, a.state.MapY); ok {
		text.Draw(screen, "背景 entry："+strconv.Itoa(int(a.state.WildernessFloor.Tiles[a.state.MapY][a.state.MapX]))+"　tile："+strconv.Itoa(int(entry.TileIndex)), a.face, 390, 125, white)
	}
	text.Draw(screen, "方向鍵：移動（依 floor movement cost）", a.face, 390, 180, white)
	text.Draw(screen, "Enter：場所　Esc：離開", a.face, 390, 215, cyan)
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

func (a *app) drawGeoPreview(screen *ebiten.Image, white, cyan color.Color) {
	text.Draw(screen, a.geoLabel+" 原始幾何預覽（G／Esc：返回）", a.face, 24, 28, cyan)
	if a.geoGrid == nil {
		text.Draw(screen, "沒有載入 GEO geometry", a.face, 24, 70, white)
		return
	}
	const cellSize = 20
	const originX, originY = 24, 48
	for y := 0; y < geo.Height; y++ {
		for x := 0; x < geo.Width; x++ {
			cell, _ := a.geoGrid.Cell(x, y)
			left := float64(originX + x*cellSize)
			top := float64(originY + y*cellSize)
			ebitenutil.DrawRect(screen, left+1, top+1, cellSize-2, cellSize-2, color.RGBA{20, 28, 52, 255})
			drawGeoWall := func(direction int, value uint8) {
				if value == 0 {
					return
				}
				wallColor := cyan
				switch direction {
				case 0:
					ebitenutil.DrawLine(screen, left, top, left+cellSize, top, wallColor)
				case 2:
					ebitenutil.DrawLine(screen, left+cellSize, top, left+cellSize, top+cellSize, wallColor)
				case 4:
					ebitenutil.DrawLine(screen, left, top+cellSize, left+cellSize, top+cellSize, wallColor)
				case 6:
					ebitenutil.DrawLine(screen, left, top, left, top+cellSize, wallColor)
				}
			}
			drawGeoWall(0, cell.WallDirections[0])
			drawGeoWall(2, cell.WallDirections[1])
			drawGeoWall(4, cell.WallDirections[2])
			drawGeoWall(6, cell.WallDirections[3])
		}
	}
	ebitenutil.DrawRect(screen, float64(originX+a.geoX*cellSize+5), float64(originY+a.geoY*cellSize+5), cellSize-10, cellSize-10, color.RGBA{255, 255, 82, 255})
	text.Draw(screen, "黃點：GEO wall 可通行游標；方向鍵移動　（不是完整 tile collision）", a.face, 24, 382, white)
}

func (a *app) drawDungeonPreview(screen *ebiten.Image, white, cyan color.Color) {
	production := a.state.Mode == game.ModeDungeon
	title := "地城結構預覽"
	if production {
		title = a.state.LocationName + "・地城探索"
	}
	text.Draw(screen, title, a.face, 24, 30, cyan)
	if a.dungeonFloor == nil {
		text.Draw(screen, "沒有載入 dungeon floor", a.face, 24, 70, white)
		return
	}
	ebitenutil.DrawRect(screen, 350, 64, 290, 145, dungeonSkyColor(a.state.Area, a.state.DungeonWallRoof))
	doorFlags, doorFlagsOK := uint8(0), false
	_, _, geometryDirection := a.state.DungeonGeometryView()
	if a.geoGrid != nil {
		doorFlags, doorFlagsOK = a.geoGrid.WallDoorFlagsWrapped(a.dungeonX, a.dungeonY, int(geometryDirection))
	}
	const (
		viewWidth  = 6
		viewHeight = 3
		originX    = 24
		originY    = 76
		tileSize   = 24
		pixelScale = 2
	)
	startX, startY := 21, 9
	for row := 0; row < viewHeight; row++ {
		for column := 0; column < viewWidth; column++ {
			x, y := startX+column, startY+row
			entry, ok := a.dungeonFloor.Entry(x, y)
			left, top := originX+column*tileSize*pixelScale, originY+row*tileSize*pixelScale
			if ok && int(entry.TileIndex) < len(a.tileImages) {
				op := &ebiten.DrawImageOptions{}
				op.Filter = ebiten.FilterNearest
				op.GeoM.Scale(pixelScale, pixelScale)
				op.GeoM.Translate(float64(left), float64(top))
				screen.DrawImage(a.tileImages[entry.TileIndex], op)
			} else {
				ebitenutil.DrawRect(screen, float64(left), float64(top), tileSize*pixelScale, tileSize*pixelScale, color.RGBA{24, 30, 48, 255})
			}
		}
	}
	text.Draw(screen, fmt.Sprintf("位置：(%d,%d)　方向：%s　地圖：GEO%d/%02X",
		a.dungeonX, a.dungeonY, dungeonDirectionName(geometryDirection), a.geoSet, a.geoBlock), a.face, 24, 242, white)
	text.Draw(screen, fmt.Sprintf("牆面：%02X　地形／屋頂：%02X", a.state.DungeonWallType, a.state.DungeonWallRoof), a.face, 24, 278, white)
	if a.state.DungeonWallType != 0 && doorFlagsOK {
		text.Draw(screen, fmt.Sprintf("門狀態：%d", doorFlags), a.face, 360, 278, white)
	}
	if a.state.Message != "" {
		drawWrappedText(screen, a.state.Message, a.face, 24, 342, 22, 30, 2, white)
	}
	if a.dungeonDoorMenu && doorFlagsOK {
		options := a.state.DungeonDoorMenuOptions(doorFlags)
		items := "B Bash　Esc Exit"
		if options.Pick {
			items = "B Bash　P Pick　"
			if options.Knock {
				items += "K Knock　"
			}
			items += "Esc Exit"
		} else if options.Knock {
			items = "B Bash　K Knock　Esc Exit"
		}
		text.Draw(screen, "上鎖的門："+items, a.compactFace, 24, 404, color.RGBA{255, 220, 110, 255})
	}
	if a.pieceLabel != "" {
		label := a.pieceLabel
		if production {
			label = fmt.Sprintf("地城牆面素材已載入：%d／%d／%d", a.state.LoadPieces[0], a.state.LoadPieces[1], a.state.LoadPieces[2])
		}
		text.Draw(screen, label, a.face, 24, 314, cyan)
	}
	if len(a.wallPreview) > 0 {
		for _, stamp := range a.wallPreview {
			op := &ebiten.DrawImageOptions{}
			op.Filter = ebiten.FilterNearest
			op.GeoM.Scale(2, 2)
			op.GeoM.Translate(float64(360+stamp.column*16), float64(76+stamp.row*16))
			screen.DrawImage(stamp.image, op)
		}
	}
	controls := "↑ 前進　K/M 轉向　E 紮營　P 撬鎖　N 敲擊　B 撞門"
	if !production {
		controls = "方向鍵：移動　Q／R：轉向　P：撬鎖　K：敲擊術　B：撞門　D／Esc：返回"
	}
	drawWrappedText(screen, controls, a.compactFace, 24, 434, 36, 20, 2, cyan)
}

func dungeonSkyColor(areaState area.State, wallRoof uint8) color.RGBA {
	var skyColours = [...]uint8{0x00, 0x0F, 0x04, 0x0B, 0x0D, 0x02, 0x09, 0x0E, 0x00, 0x0F, 0x04, 0x0B, 0x0D, 0x02, 0x09, 0x0E}
	index := areaState.OutdoorSkyColor
	if wallRoof > 0x7F {
		index = areaState.IndoorSkyColor
	}
	return gfx.EGA16[skyColours[index%uint16(len(skyColours))]]
}

func dungeonDirectionName(direction uint8) string {
	names := [...]string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
	if int(direction) >= len(names) {
		return "?"
	}
	return names[direction]
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
	start := a.state.CreationCursor - 3
	if start < 0 {
		start = 0
	}
	if maxStart := len(a.state.CreationOptions) - 5; start > maxStart && maxStart > 0 {
		start = maxStart
	}
	end := start + 5
	if end > len(a.state.CreationOptions) {
		end = len(a.state.CreationOptions)
	}
	for index := start; index < end; index++ {
		character := a.state.CreationOptions[index]
		prefix := "  "
		if index == a.state.CreationCursor {
			prefix = "> "
		}
		label := prefix + character.Name + "（" + raceName(character.Race) + "／" + className(character.Class) + "）"
		text.Draw(screen, label, a.face, 48, 150+(index-start)*32, white)
	}
	text.Draw(screen, "選項 "+strconv.Itoa(a.state.CreationCursor+1)+"／"+strconv.Itoa(len(a.state.CreationOptions))+"　已加入："+strconv.Itoa(len(a.state.CreationRoster))+" 人", a.face, 48, 325, cyan)
	text.Draw(screen, "N：改名　A：能力值　R：重擲　Enter：加入　D：完成　Esc：取消", a.face, 48, 340, white)
}

func raceName(r party.Race) string {
	return map[party.Race]string{party.RaceDwarf: "矮人", party.RaceElf: "精靈", party.RaceGnome: "侏儒", party.RaceHalfElf: "半精靈", party.RaceHalfling: "半身人", party.RaceHuman: "人類", party.RaceHalfOrc: "半獸人"}[r]
}

func className(c party.Class) string {
	return map[party.Class]string{party.ClassCleric: "牧師", party.ClassFighter: "戰士", party.ClassRanger: "遊俠", party.ClassPaladin: "聖武士", party.ClassMagicUser: "魔法師", party.ClassThief: "盜賊"}[c]
}

func (a *app) drawCombat(screen *ebiten.Image, white, cyan color.Color) {
	drawPanelFrame(screen, 8, 8, 352, 376)
	drawPanelFrame(screen, 360, 8, 272, 376)
	drawPanelFrame(screen, 8, 384, 624, 64)
	drawPanelFrame(screen, 8, 448, 624, 28)
	battlefield := ebiten.NewImage(352, 376)
	for row := 0; row < 7; row++ {
		for column := 0; column < 7; column++ {
			cell := color.RGBA{R: 34, G: 46, B: 58, A: 255}
			if (row+column)%2 == 0 {
				cell = color.RGBA{R: 40, G: 54, B: 66, A: 255}
			}
			ebitenutil.DrawRect(battlefield, float64(16+column*48), float64(16+row*48), 46, 46, cell)
		}
	}
	battlefieldOp := &ebiten.DrawImageOptions{}
	battlefieldOp.GeoM.Translate(8, 8)
	screen.DrawImage(battlefield, battlefieldOp)
	drawWrappedText(screen, a.state.CombatMessage(), a.compactFace, 24, 414, 36, 20, 1, white)
	if a.state.CombatViewActive() {
		text.Draw(screen, "角色檢視", a.face, 64, 145, cyan)
		for index, line := range a.state.CombatViewLines() {
			text.Draw(screen, line, a.face, 64, 185+index*30, white)
		}
		text.Draw(screen, "Enter／Esc：返回戰鬥", a.face, 64, 350, cyan)
		return
	}
	active, activeOK := a.state.CombatActiveFighter()
	cameraFocus := combat.TilePoint{}
	minX, maxX, minY, maxY := 0, 0, 0, 0
	havePosition := false
	for _, fighter := range a.state.CombatFighters() {
		if !fighter.HasCombatPosition {
			continue
		}
		if !havePosition {
			minX, maxX, minY, maxY = fighter.CombatX, fighter.CombatX, fighter.CombatY, fighter.CombatY
			havePosition = true
		} else {
			minX, maxX = min(minX, fighter.CombatX), max(maxX, fighter.CombatX)
			minY, maxY = min(minY, fighter.CombatY), max(maxY, fighter.CombatY)
		}
	}
	useCamera := havePosition && (minX < 0 || maxX > 6 || minY < 0 || maxY > 6)
	if useCamera {
		cameraFocus = combat.TilePoint{X: (minX + maxX + 1) / 2, Y: (minY + maxY + 1) / 2}
	} else if activeOK {
		cameraFocus = combat.TilePoint{X: active.CombatX, Y: active.CombatY}
	}
	camera := combat.NewCombatCamera(
		cameraFocus,
		combat.TilePoint{X: 3, Y: 3},
		useCamera,
	)
	partyIndex, enemyIndex := 0, 0
	targets := a.state.CombatTargets()
	spellTargets := a.state.CombatSpellTargets()
	for _, fighter := range a.state.CombatFighters() {
		if fighter.HitPoints <= 0 && !fighter.DownedCorpse {
			if _, active := a.deathOverlayFrame(fighter); !active {
				continue
			}
		}
		if fighter.Side == combat.SideParty {
			tile := combat.FormationTile(fighter.Side, partyIndex)
			if fighter.HasCombatPosition || fighter.DeathOverlay || fighter.DownedCorpse {
				tile = combat.TilePoint{X: fighter.CombatX, Y: fighter.CombatY}
			}
			tile = camera.Apply(tile)
			tile.X = 6 - tile.X
			x, y := 12+tile.X*48, 16+tile.Y*48
			if !fighter.DownedCorpse && !fighter.DeathOverlay {
				a.drawFighterSprite(battlefield, fighter, partyIndex, x, y)
			}
			a.drawFighterDeathOverlay(battlefield, fighter, x, y)
			selected := false
			if (a.state.CombatCastingSpell() == game.CureLightWoundsSpellID || a.state.CombatCastingSpell() == game.ProtectionFromEvilSpellID || (a.state.CombatCastingSpell() == game.ProtectionFromGoodSpellID && !a.state.CombatSpellTargetsEnemy())) && a.state.CombatSpellTargetIndex() < len(spellTargets) && spellTargets[a.state.CombatSpellTargetIndex()].ID == fighter.ID {
				selected = true
			}
			a.drawCombatSpriteMarker(battlefield, fighter, activeOK && active.ID == fighter.ID, selected, x, y)
			partyIndex++
			continue
		}
		tile := combat.FormationTile(fighter.Side, enemyIndex)
		if fighter.HasCombatPosition || fighter.DeathOverlay || fighter.DownedCorpse {
			tile = combat.TilePoint{X: fighter.CombatX, Y: fighter.CombatY}
		}
		tile = camera.Apply(tile)
		tile.X = 6 - tile.X
		x, y := 12+tile.X*48, 16+tile.Y*48
		if !fighter.DownedCorpse && !fighter.DeathOverlay {
			a.drawFighterSprite(battlefield, fighter, enemyIndex, x, y)
		}
		a.drawFighterDeathOverlay(battlefield, fighter, x, y)
		selected := false
		if (a.state.CombatCastingSpell() == 0 || a.state.CombatSpellTargetsEnemy()) && len(targets) > 0 && a.state.CombatTargetIndex() < len(targets) && targets[a.state.CombatTargetIndex()].ID == fighter.ID {
			selected = true
		}
		a.drawCombatSpriteMarker(battlefield, fighter, activeOK && active.ID == fighter.ID, selected, x, y)
		enemyIndex++
	}
	screen.DrawImage(battlefield, battlefieldOp)
	drawPanelFrame(screen, 8, 8, 352, 376)
	if activeOK {
		text.Draw(screen, active.Name, a.face, 378, 44, cyan)
		text.Draw(screen, fmt.Sprintf("HP %d/%d", active.HitPoints, active.MaxHitPoints), a.face, 378, 82, white)
		text.Draw(screen, fmt.Sprintf("AC %d", active.ArmorClass), a.face, 378, 118, white)
	}
	if len(targets) > 0 && a.state.CombatTargetIndex() < len(targets) {
		target := targets[a.state.CombatTargetIndex()]
		text.Draw(screen, "目標："+target.Name, a.face, 378, 190, cyan)
		text.Draw(screen, fmt.Sprintf("HP %d/%d", target.HitPoints, target.MaxHitPoints), a.face, 378, 228, white)
		text.Draw(screen, fmt.Sprintf("AC %d", target.ArmorClass), a.face, 378, 264, white)
	}
	spellHints := make([]string, 0, 7)
	if a.state.CombatCastingSpell() != 0 {
		if a.state.CombatCastingSpell() == game.BlessSpellID {
			text.Draw(screen, "確認施法：Enter　取消：Esc", a.face, 32, 350, cyan)
			return
		}
		text.Draw(screen, "選擇施法目標：左右切換　Enter：確認　Esc：取消", a.face, 32, 350, cyan)
		return
	}
	if a.state.CombatMoveMode() {
		text.Draw(screen, fmt.Sprintf("移動方向：方向鍵　剩餘 %d 格　取消：Esc", a.state.CombatMoveRemaining()), a.face, 32, 350, cyan)
		return
	}
	if a.state.CombatCanCastMagicMissile() {
		spellHints = append(spellHints, "S飛彈")
	}
	if a.state.CombatCanCastCureLightWounds() {
		spellHints = append(spellHints, "H治療")
	}
	if a.state.CombatCanCastBless() {
		spellHints = append(spellHints, "B祝福")
	}
	if a.state.CombatCanCastCurse() {
		spellHints = append(spellHints, "C詛咒")
	}
	if a.state.CombatCanCastCauseLightWounds() {
		spellHints = append(spellHints, "W傷害")
	}
	if a.state.CombatCanCastProtectionFromEvil() {
		spellHints = append(spellHints, "P防邪")
	}
	if a.state.CombatCanCastProtectionFromGood() {
		spellHints = append(spellHints, "G防善")
	}
	text.Draw(screen, "移動　查看　瞄準　使用　施法　快速　結束", a.compactFace, 20, 470, cyan)
	if len(spellHints) > 0 {
		text.Draw(screen, "快捷："+strings.Join(spellHints, "　"), a.compactFace, 378, 340, cyan)
	}
}

func (a *app) drawCombatSpriteMarker(screen *ebiten.Image, fighter combat.Fighter, active, selected bool, x, y int) {
	teamColor := color.RGBA{R: 70, G: 190, B: 255, A: 255}
	if fighter.Side == combat.SideEnemy {
		teamColor = color.RGBA{R: 255, G: 82, B: 82, A: 255}
	}
	ebitenutil.DrawRect(screen, float64(x), float64(y-4), 48, 3, teamColor)
	if !active && !selected {
		return
	}
	marker := color.RGBA{R: 255, G: 235, B: 80, A: 255}
	if selected {
		marker = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	ebitenutil.DrawRect(screen, float64(x-3), float64(y-7), 54, 2, marker)
	ebitenutil.DrawRect(screen, float64(x-3), float64(y+49), 54, 2, marker)
	ebitenutil.DrawRect(screen, float64(x-3), float64(y-7), 2, 58, marker)
	ebitenutil.DrawRect(screen, float64(x+49), float64(y-7), 2, 58, marker)
}

// drawFighterDeathOverlay is the renderer adapter for the combat-core
// DeathOverlay signal. seg001.Init maps combat_icons[24] to COMSPR block 0x0B
// (attack becomes 0x8B) and combat_icons[25] to COMSPR block 0x19 (normal).
// CombatantKilled alternates those two icons while flashing the skull.
func (a *app) drawFighterDeathOverlay(screen *ebiten.Image, fighter combat.Fighter, x, y int) {
	if !fighter.DeathOverlay && !fighter.DownedCorpse {
		if a.deathOverlayStarted != nil {
			delete(a.deathOverlayStarted, fighter.ID)
		}
		return
	}
	if fighter.DownedCorpse && !fighter.DeathOverlay {
		ebitenutil.DrawRect(screen, float64(x-4), float64(y-4), 36, 44, color.RGBA{R: 58, G: 38, B: 28, A: 220})
		text.Draw(screen, "倒下", a.face, x, y+22, color.RGBA{R: 255, G: 220, B: 180, A: 255})
		return
	}
	frame, active := a.deathOverlayFrame(fighter)
	if !active {
		if fighter.DownedCorpse {
			ebitenutil.DrawRect(screen, float64(x-4), float64(y-4), 36, 44, color.RGBA{R: 58, G: 38, B: 28, A: 220})
			text.Draw(screen, "倒下", a.face, x, y+22, color.RGBA{R: 255, G: 220, B: 180, A: 255})
		}
		return
	}
	iconKey := "comspr-block-19-item-00.png"
	if frame == 0 {
		iconKey = "comspr-block-8B-item-00.png"
	}
	if icon := a.combatSprites[iconKey]; icon != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(2, 2)
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(icon, op)
		return
	}
	// Keep a visible diagnostic fallback if a derived sprite was not packaged.
	ebitenutil.DrawRect(screen, float64(x-4), float64(y-4), 36, 44, color.RGBA{R: 80, G: 12, B: 20, A: 220})
	text.Draw(screen, "倒下", a.face, x, y+22, color.RGBA{R: 255, G: 220, B: 180, A: 255})
}

func (a *app) deathOverlayFrame(fighter combat.Fighter) (uint8, bool) {
	if !fighter.DeathOverlay {
		if a.deathOverlayStarted != nil {
			delete(a.deathOverlayStarted, fighter.ID)
		}
		return 0, false
	}
	if a.deathOverlayStarted == nil {
		a.deathOverlayStarted = make(map[string]time.Time)
	}
	started, ok := a.deathOverlayStarted[fighter.ID]
	if !ok {
		started = time.Now()
		a.deathOverlayStarted[fighter.ID] = started
	}
	frame, active := combat.DeathOverlayFrame(time.Since(started))
	if !active && !fighter.DownedCorpse {
		delete(a.deathOverlayStarted, fighter.ID)
	}
	return frame, active
}

func (a *app) drawFighterSprite(screen *ebiten.Image, fighter combat.Fighter, ordinal, x, y int) {
	if len(a.combatSprites) == 0 {
		return
	}
	key := ""
	var sprite *ebiten.Image
	var frameX, frameY int16
	if fighter.Side == combat.SideParty {
		// Party icons are the original CHEAD+CBODY composition. Imported DOS
		// slots are normalized by Character.CombatIconBlocks; the fallback below
		// composes the extracted raw layers on demand.
		headBlock, bodyBlock := uint8(0), uint8(0)
		if fighter.HasPartyIcon {
			headBlock, bodyBlock = fighter.PartyHeadBlock, fighter.PartyBodyBlock
		}
		prefix := "party"
		if fighter.IconAttack {
			prefix = "party-attack"
		}
		key = fmt.Sprintf("%s-head-%02X-body-%02X.png", prefix, headBlock, bodyBlock)
		sprite = a.combatSprites[key]
	}
	if sprite == nil && fighter.HasAnimation {
		key = fmt.Sprintf("sprit%d-block-%02X", fighter.SpriteSet, fighter.AnimationBlock)
		if animation := a.combatAnimations[key]; len(animation) > 0 {
			frame := animation[0]
			if a.state.AnimationsEnabled() {
				frame = animationFrame(animation, time.Since(a.animationStart))
			}
			sprite = frame.image
			frameX, frameY = frame.x, frame.y
		}
	}
	if sprite == nil && fighter.Side == combat.SideParty && fighter.HasPartyIcon {
		headBlock, bodyBlock := fighter.PartyHeadBlock, fighter.PartyBodyBlock
		if fighter.IconAttack {
			headBlock += 0x80
			bodyBlock += 0x80
		}
		headKey := fmt.Sprintf("chead-block-%02X-item-00.png", headBlock)
		bodyKey := fmt.Sprintf("cbody-block-%02X-item-00.png", bodyBlock)
		headImage, headOK := a.combatSprites[headKey]
		bodyImage, bodyOK := a.combatSprites[bodyKey]
		if headOK && bodyOK {
			composite := ebiten.NewImage(24, 24)
			composite.DrawImage(bodyImage, nil)
			composite.DrawImage(headImage, nil)
			sprite = composite
		}
	}
	if sprite == nil && fighter.SpriteBlock != 0 {
		key = fmt.Sprintf("cpic%d-block-%02X-item-00.png", fighter.SpriteSet, fighter.SpriteBlock)
		sprite = a.combatSprites[key]
	}
	if sprite == nil {
		if len(a.combatSpriteIDs) == 0 {
			return
		}
		key = a.combatSpriteIDs[ordinal%len(a.combatSpriteIDs)]
		sprite = a.combatSprites[key]
	}
	op := &ebiten.DrawImageOptions{}
	if fighter.IconDirection > 3 {
		op.GeoM.Scale(-2, 2)
		op.GeoM.Translate(float64(x)+float64(frameX*2)+float64(sprite.Bounds().Dx()*2), float64(y)+float64(frameY*2))
	} else {
		op.GeoM.Scale(2, 2)
		op.GeoM.Translate(float64(x)+float64(frameX*2), float64(y)+float64(frameY*2))
	}
	screen.DrawImage(sprite, op)
}

func (a *app) Layout(_, _ int) (int, int) { return logicalWidth, logicalHeight }

func loadFace(path string, size float64) font.Face {
	candidates := []string{path}
	if path == "" {
		candidates = []string{
			"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
			"/usr/share/fonts/truetype/noto/NotoSansCJK-Regular.ttc",
			"/System/Library/Fonts/PingFang.ttc",
			`C:\Windows\Fonts\msjh.ttc`,
		}
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		data, err := os.ReadFile(candidate)
		if err == nil {
			if parsed, err := opentype.Parse(data); err == nil {
				if face, err := opentype.NewFace(parsed, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull}); err == nil {
					return face
				}
			}
			if collection, err := opentype.ParseCollection(data); err == nil && collection.NumFonts() > 0 {
				if parsed, err := collection.Font(0); err == nil {
					if face, err := opentype.NewFace(parsed, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull}); err == nil {
						return face
					}
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
	geoSet := flag.Int("geo-set", 2, "GEO DAX set/chapter (2..6) used by the map preview")
	geoBlock := flag.Int("geo-block", 1, "original GEO block ID used by the map preview")
	encounter := flag.Bool("encounter", false, "start a decoded ECL encounter directly")
	opening := flag.Bool("opening", false, "start at the formal new-game opening with one generated character")
	inn := flag.Bool("inn", false, "start at the first Windlord's Inn event through the formal new-game flow")
	filani := flag.Bool("filani", false, "start at sage Filani through the formal Tilverton ECL flow")
	weaponShop := flag.Bool("weapon-shop", false, "start at Weaponers of Cormyr through the formal Tilverton ECL flow")
	temple := flag.Bool("temple", false, "start at Gond's altar through the formal Tilverton ECL flow")
	training := flag.Bool("training", false, "start at Tilverton's Hall of Training through the formal ECL flow")
	tavern := flag.Bool("tavern", false, "start at Tilverton's tavern through the formal ECL flow")
	highPriest := flag.Bool("high-priest", false, "start at Tilverton's high priest through the formal ECL flow")
	carriage := flag.Bool("carriage", false, "start at Tilverton's royal-carriage main-story event through the formal ECL flow")
	guildmaster := flag.Bool("guildmaster", false, "start at the Thieves' Guild mixed-team battle through the full ECL story path")
	sewers := flag.Bool("sewers", false, "start at the first Tilverton Sewers checkpoint through the full ECL story path")
	lavaTube := flag.Bool("lava-tube", false, "start at the Hap map route into the ancient lava tube")
	wizardTower := flag.Bool("wizard-tower", false, "start at the ECL5 wizard-tower courtyard and Dracandros story")
	wizardTowerBattle := flag.Bool("wizard-tower-battle", false, "start at Dracandros' original wizard-tower patrol battle")
	wizardTowerParlay := flag.Bool("wizard-tower-parlay", false, "start after successfully parlaying with the wizard-tower dragons")
	wizardTowerExit := flag.Bool("wizard-tower-exit", false, "start at the completed wizard-tower roof exit menu")
	encounterBlock := flag.Int("encounter-block", 81, "ECL block for -encounter")
	encounterStart := flag.Int("encounter-start", 0x1293, "payload offset for -encounter")
	encounterMonsterMember := flag.String("encounter-monster-member", "MON1CHA.DAX", "MON*CHA member for -encounter")
	encounterArea := flag.Int("encounter-area", 1, "original graphics area used by -encounter sprites (1..6)")
	partyPath := flag.String("party-save", "party.json", "versioned remake party save path")
	soundDir := flag.String("sound-dir", "assets/audio", "reference WAV asset directory; missing assets disable sound")
	partyLoadPath := flag.String("party-load", "", "load a versioned remake party save before starting")
	savgamDir := flag.String("savgam-dir", "", "directory containing reference savgam?.dat and CHRDAT player bundles")
	savgamSlot := flag.String("savgam-slot", "", "reference SAVGAM slot key A..J to load and save with -savgam-dir")
	dosCharacterID := flag.String("dos-character-id", "dos-character", "ID for a direct DOS character import")
	dosCharacterRecord := flag.String("dos-character-record", "", "DOS .SAV/.GUY path to load directly into the remake")
	dosCharacterEffects := flag.String("dos-character-effects", "", "optional DOS .FX path for direct character import")
	dosCharacterInventory := flag.String("dos-character-inventory", "", "optional DOS .SWG path for direct character import")
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
	itemData, err := zipMember(*imagePath, "ITEMS")
	if err != nil {
		log.Fatal(err)
	}
	itemCatalog, err := monster.ParseBaseItems(itemData)
	if err != nil {
		log.Fatal(err)
	}
	state.SetItemCatalog(itemCatalog)
	treasureItems, err := loadTreasureItemBlocks(*imagePath)
	if err != nil {
		log.Fatal(err)
	}
	state.SetTreasureItemBlocks(treasureItems)
	monsterData, err := zipMember(*imagePath, *encounterMonsterMember)
	if err != nil {
		log.Fatal(err)
	}
	monsterRecords, err := loadMonsterRecords(monsterData)
	if err != nil {
		log.Fatal(err)
	}
	state.SetMonsterRecords(monsterRecords)
	monsterAffectData, err := zipMember(*imagePath, "MON1SPC.DAX")
	if err != nil {
		log.Fatal(err)
	}
	monsterAffects, err := loadMonsterAffects(monsterAffectData)
	if err != nil {
		log.Fatal(err)
	}
	state.SetMonsterAffects(monsterAffects)
	for chapter := uint8(1); chapter <= 6; chapter++ {
		member := fmt.Sprintf("MON%dCHA.DAX", chapter)
		data, loadErr := zipMember(*imagePath, member)
		if loadErr != nil {
			log.Fatal(loadErr)
		}
		records, parseErr := loadMonsterRecords(data)
		if parseErr != nil {
			log.Fatal(parseErr)
		}
		state.SetMonsterRecordsForECL(chapter, records)
		affectData, affectErr := zipMember(*imagePath, fmt.Sprintf("MON%dSPC.DAX", chapter))
		if affectErr != nil {
			log.Fatal(affectErr)
		}
		affects, parseAffectErr := loadMonsterAffects(affectData)
		if parseAffectErr != nil {
			log.Fatal(parseAffectErr)
		}
		state.SetMonsterAffectsForECL(chapter, affects)
		itemData, itemErr := zipMember(*imagePath, fmt.Sprintf("MON%dITM.DAX", chapter))
		if itemErr != nil {
			log.Fatal(itemErr)
		}
		items, parseItemErr := loadMonsterItems(itemData)
		if parseItemErr != nil {
			log.Fatal(parseItemErr)
		}
		state.SetMonsterItemsForECL(chapter, items)
	}
	soundPlayer, soundErr := sound.Load(*soundDir)
	if soundErr != nil {
		log.Printf("sound disabled: %v", soundErr)
	}
	var loadedSAVGAMSlot byte
	if *savgamDir != "" || *savgamSlot != "" {
		if *savgamDir == "" || len(*savgamSlot) != 1 || *partyLoadPath != "" || *dosCharacterRecord != "" {
			log.Fatal("-savgam-dir/-savgam-slot require exactly one A..J key and cannot be combined with party/player import flags")
		}
		loadedSAVGAMSlot = strings.ToUpper(*savgamSlot)[0]
		if loadedSAVGAMSlot < 'A' || loadedSAVGAMSlot > 'J' {
			log.Fatal("-savgam-slot must be one letter A..J")
		}
		if err := state.LoadSAVGAMSlot(*savgamDir, loadedSAVGAMSlot); err != nil {
			log.Fatal(err)
		}
	} else if *partyLoadPath != "" {
		if *dosCharacterRecord != "" {
			log.Fatal("-party-load and -dos-character-record cannot be used together")
		}
		if err := state.LoadPartyFile(*partyLoadPath); err != nil {
			log.Fatal(err)
		}
	} else if *dosCharacterRecord != "" {
		record, err := os.ReadFile(*dosCharacterRecord)
		if err != nil {
			log.Fatal(err)
		}
		effects, err := readOptional(*dosCharacterEffects)
		if err != nil {
			log.Fatal(err)
		}
		inventory, err := readOptional(*dosCharacterInventory)
		if err != nil {
			log.Fatal(err)
		}
		if err := state.LoadDOSCharacterFiles(*dosCharacterID, party.DOSPlayerFiles{Record: record, Effects: effects, Inventory: inventory}); err != nil {
			log.Fatal(err)
		}
	}
	tileImages, err := loadTileImages(*imagePath)
	if err != nil {
		log.Fatal(err)
	}
	geoCatalog, err := loadGEOCatalog(*imagePath)
	if err != nil {
		log.Fatal(err)
	}
	geoRef := geo.MapRef{Set: uint8(*geoSet), BlockID: uint8(*geoBlock)}
	geoGridValue, ok := geoCatalog.Lookup(geoRef)
	if !ok {
		log.Fatalf("GEO%d block 0x%02X is not in original catalog", geoRef.Set, geoRef.BlockID)
	}
	geoGrid := &geoGridValue
	dungeonFloorValue := mapdata.GenerateDungeon(*geoGrid, state.DungeonX, state.DungeonY)
	dungeonFloor := &dungeonFloorValue
	if *encounter {
		if *encounterArea < 1 || *encounterArea > 6 {
			log.Fatal("-encounter-area must be 1..6")
		}
		state.Area.GameArea = uint8(*encounterArea)
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
	} else if *lavaTube {
		if len(state.PartyFighters()) != 0 {
			log.Fatal("-lava-tube cannot be combined with a loaded party")
		}
		if err := state.OpenCharacterCreation(); err != nil {
			log.Fatal(err)
		}
		if err := state.AddCreationCharacter(0); err != nil {
			log.Fatal(err)
		}
		if err := state.FinishCharacterCreation(); err != nil {
			log.Fatal(err)
		}
		if err := state.StartDungeonStoryPreview(0x32, 0x31, 5); err != nil {
			log.Fatal(err)
		}
	} else if *wizardTower || *wizardTowerBattle || *wizardTowerParlay || *wizardTowerExit {
		if len(state.PartyFighters()) != 0 {
			log.Fatal("wizard-tower previews cannot be combined with a loaded party")
		}
		if err := state.OpenCharacterCreation(); err != nil {
			log.Fatal(err)
		}
		if err := state.AddCreationCharacter(0); err != nil {
			log.Fatal(err)
		}
		if err := state.FinishCharacterCreation(); err != nil {
			log.Fatal(err)
		}
		if err := state.StartDungeonStoryPreview(0x33, 0x32, 5); err != nil {
			log.Fatal(err)
		}
		if *wizardTowerExit {
			state.SetECLMemoryValue(0x4C60, 1)
			state.Mode = game.ModeDungeon
			state.PictureRequested = false
			state.DungeonX, state.DungeonY, state.DungeonDirection = 7, 15, 2
			state.DungeonWallRoof = 1
			if err := state.RunDungeonLifecycle(); err != nil {
				log.Fatal(err)
			}
			if len(state.Choices) != 3 {
				log.Fatal("-wizard-tower-exit did not reach the original three-way roof menu")
			}
		}
		if *wizardTowerBattle {
			for step := 0; step < 32 && !state.CombatActive(); step++ {
				if state.Mode == game.ModeEvent {
					if err := state.Continue(); err != nil {
						log.Fatal(err)
					}
					continue
				}
				selection := 0
				for index, choice := range state.Choices {
					if choice == "等待" || choice == "攻擊法師" {
						selection = index
					}
				}
				if err := state.Select(selection); err != nil {
					log.Fatal(err)
				}
			}
			if !state.CombatActive() {
				log.Fatal("-wizard-tower-battle did not reach the original combat boundary")
			}
		}
		if *wizardTowerParlay {
			reachedParlayText := false
			for step := 0; step < 40; step++ {
				if strings.Contains(state.Message, "沒有對付龍族的陰謀") {
					reachedParlayText = true
					break
				}
				if state.Mode == game.ModeEvent {
					if err := state.Continue(); err != nil {
						log.Fatal(err)
					}
					continue
				}
				selection := 0
				for index, choice := range state.Choices {
					if choice == "等待" || choice == "與龍群交涉" || choice == "狡猾" {
						selection = index
					}
				}
				if err := state.Select(selection); err != nil {
					log.Fatal(err)
				}
			}
			if !reachedParlayText {
				log.Fatal("-wizard-tower-parlay did not reach the original successful parlay text")
			}
		}
	} else if *opening || *inn || *filani || *weaponShop || *temple || *training || *tavern || *highPriest || *carriage || *guildmaster || *sewers {
		if len(state.PartyFighters()) != 0 {
			log.Fatal("story preview flags cannot be combined with a loaded party")
		}
		if err := state.OpenCharacterCreation(); err != nil {
			log.Fatal(err)
		}
		if err := state.AddCreationCharacter(0); err != nil {
			log.Fatal(err)
		}
		if err := state.FinishCharacterCreation(); err != nil {
			log.Fatal(err)
		}
		if *inn || *filani || *weaponShop || *temple || *training || *tavern || *highPriest || *carriage || *guildmaster || *sewers {
			if err := state.Select(0); err != nil {
				log.Fatal(err)
			}
			if err := state.Continue(); err != nil {
				log.Fatal(err)
			}
			if err := state.Select(0); err != nil {
				log.Fatal(err)
			}
			if *carriage || *guildmaster || *sewers {
				if err := prepareCarriagePreview(&state, geoGrid); err != nil {
					log.Fatal(err)
				}
				if *guildmaster || *sewers {
					if err := prepareGuildmasterBattle(&state, geoGrid); err != nil {
						log.Fatal(err)
					}
					if *sewers {
						sewerGrid, ok := geoCatalog.Lookup(geo.MapRef{Set: 2, BlockID: 3})
						if !ok {
							log.Fatal("GEO2 block 3 is unavailable")
						}
						if err := prepareSewerCheckpoint(&state, geoGrid, &sewerGrid); err != nil {
							log.Fatal(err)
						}
					}
				}
			}
			x, y, direction := 6, 13, uint8(6)
			if *filani {
				x, y, direction = 6, 5, 0
			} else if *weaponShop {
				x, y, direction = 2, 12, 0
			} else if *temple {
				x, y, direction = 0, 7, 0
			} else if *training {
				x, y, direction = 5, 2, 0
			} else if *tavern {
				x, y, direction = 6, 10, 0
			} else if *highPriest {
				x, y, direction = 1, 10, 0
			} else if *carriage || *guildmaster || *sewers {
				x, y, direction = state.DungeonX, state.DungeonY, state.DungeonDirection
			}
			if !*carriage && !*guildmaster && !*sewers {
				state.DungeonX, state.DungeonY, state.DungeonDirection = x, y, direction
				state.DungeonWallType, _ = geoGrid.WallWrapped(x, y, int(direction))
				state.DungeonWallRoof = geoGrid.CellWrapped(x, y).Terrain
				if err := state.RunDungeonLifecycle(); err != nil {
					log.Fatal(err)
				}
			}
		}
	}
	ebiten.SetWindowSize(logicalWidth, logicalHeight)
	ebiten.SetWindowTitle(catalog.Text("title", "Curse of the Azure Bonds"))
	geoLabel := fmt.Sprintf("GEO%d block 0x%02X", *geoSet, *geoBlock)
	combatSprites, combatSpriteIDs, combatAnimations, err := loadCombatSprites()
	if err != nil {
		log.Fatal(err)
	}
	dungeonX, dungeonY, _ := state.DungeonGeometryView()
	if err := ebiten.RunGame(&app{state: state, imagePath: *imagePath, face: loadFace(*fontPath, 24), compactFace: loadFace(*fontPath, 16), partyPath: *partyPath, savgamDir: *savgamDir, savgamSlot: loadedSAVGAMSlot, savgamSlotSave: loadedSAVGAMSlot != 0, soundPlayer: soundPlayer, tileImages: tileImages, geoGrid: geoGrid, dungeonFloor: dungeonFloor, dungeonX: dungeonX, dungeonY: dungeonY, geoLabel: geoLabel, geoCatalog: geoCatalog, geoSet: geoRef.Set, geoBlock: geoRef.BlockID, pieceSets: make(map[uint8]gfx.PieceSet), combatSprites: combatSprites, combatSpriteIDs: combatSpriteIDs, combatAnimations: combatAnimations, animationStart: time.Now()}); err != nil {
		log.Fatal(err)
	}
}

func prepareSewerCheckpoint(state *game.State, guildGrid, sewerGrid *geo.Grid) error {
	for turn := 0; turn < 100 && state.CombatActive(); turn++ {
		if err := state.CombatAct(); err != nil {
			return fmt.Errorf("finish guild battle: %w", err)
		}
	}
	if state.CombatStatus() != combat.StatusPartyWon {
		return fmt.Errorf("guild battle ended with status %v", state.CombatStatus())
	}
	if err := state.Select(0); err != nil {
		return err
	}
	state.SetDungeonGeometryView(10, 15, 4)
	state.DungeonWallType, _ = guildGrid.WallWrapped(10, 15, 4)
	state.DungeonWallRoof = guildGrid.CellWrapped(10, 15).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		return fmt.Errorf("inspect sewer door: %w", err)
	}
	if err := state.Continue(); err != nil {
		return err
	}
	if err := state.RunDungeonExitLifecycle(); err != nil {
		return fmt.Errorf("enter sewers: %w", err)
	}
	if err := state.Select(0); err != nil {
		return err
	}
	state.DungeonX, state.DungeonY, state.DungeonDirection = 1, 8, 2
	state.DungeonWallType, _ = sewerGrid.WallWrapped(1, 8, 2)
	state.DungeonWallRoof = sewerGrid.CellWrapped(1, 8).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		return fmt.Errorf("enter Fire Knife checkpoint: %w", err)
	}
	return nil
}

func prepareGuildmasterBattle(state *game.State, grid *geo.Grid) error {
	hero := state.PartyFighters()[0]
	hero.HitPoints, hero.MaxHitPoints = 200, 200
	hero.ArmorClass, hero.InitiativeBonus, hero.AttackBonus = -10, 100, 100
	hero.DamageDiceCount, hero.DamageDiceSides, hero.DamageBonus = 1, 1, 100
	if err := state.SetParty([]combat.Fighter{hero}); err != nil {
		return err
	}
	if err := state.Continue(); err != nil {
		return err
	}
	for range 4 {
		if err := state.Select(0); err != nil {
			return err
		}
	}
	for turn := 0; turn < 10 && state.CombatActive(); turn++ {
		if err := state.CombatAct(); err != nil {
			return err
		}
	}
	for _, selection := range []int{0, 0, 0} {
		if err := state.Select(selection); err != nil {
			return err
		}
	}
	if err := state.Continue(); err != nil {
		return err
	}
	for range 2 {
		if err := state.Select(0); err != nil {
			return err
		}
	}
	x, y, direction := state.DungeonGeometryView()
	state.DungeonWallType, _ = grid.WallWrapped(x, y, int(direction))
	state.DungeonWallRoof = grid.CellWrapped(x, y).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		return err
	}
	if err := state.Continue(); err != nil {
		return err
	}
	if err := state.Select(1); err != nil {
		return err
	}
	for range 4 {
		if err := state.Select(0); err != nil {
			return err
		}
	}
	return nil
}

func prepareCarriagePreview(state *game.State, grid *geo.Grid) error {
	enter := func(x, y int) error {
		state.DungeonX, state.DungeonY, state.DungeonDirection = x, y, 0
		state.DungeonWallType, _ = grid.WallWrapped(x, y, 0)
		state.DungeonWallRoof = grid.CellWrapped(x, y).Terrain
		return state.RunDungeonLifecycle()
	}

	// Visit Weaponers through the real shop service, then leave without buying.
	if err := enter(2, 12); err != nil {
		return fmt.Errorf("prepare carriage Weaponers: %w", err)
	}
	if err := state.Continue(); err != nil {
		return err
	}
	if err := state.Select(0); err != nil {
		return err
	}
	if err := state.Select(8); err != nil {
		return err
	}
	if err := state.Select(0); err != nil {
		return err
	}

	// Tell Filani the truth through the real ROB／Journal 38 path.
	if err := enter(6, 5); err != nil {
		return fmt.Errorf("prepare carriage Filani: %w", err)
	}
	if err := state.Continue(); err != nil {
		return err
	}
	for _, selection := range []int{0, 0, 0, 0} {
		if err := state.Select(selection); err != nil {
			return err
		}
	}

	// The first gate visit records the warning; the second starts the carriage.
	if err := enter(1, 0); err != nil {
		return fmt.Errorf("prepare carriage first gate: %w", err)
	}
	if err := state.Select(0); err != nil {
		return err
	}
	if err := enter(1, 0); err != nil {
		return fmt.Errorf("prepare carriage second gate: %w", err)
	}
	return nil
}

func animationFrame(frames []combatAnimation, elapsed time.Duration) combatAnimation {
	if len(frames) == 0 {
		return combatAnimation{}
	}
	return frames[animationFrameIndex(frames, elapsed)]
}

func animationFrameIndex(frames []combatAnimation, elapsed time.Duration) int {
	if len(frames) == 0 {
		return -1
	}
	delays := make([]uint32, len(frames))
	for index, frame := range frames {
		delays[index] = frame.delay
	}
	return gfx.AnimationFrameIndex(delays, elapsed)
}

func loadCombatSprites() (map[string]*ebiten.Image, []string, map[string][]combatAnimation, error) {
	paths, err := filepath.Glob("assets/sprites/cpic*-block-*-item-00.png")
	if err != nil {
		return nil, nil, nil, err
	}
	partyPaths, err := filepath.Glob("assets/sprites/party*-head-*-body-*.png")
	if err != nil {
		return nil, nil, nil, err
	}
	paths = append(paths, partyPaths...)
	cheadPaths, err := filepath.Glob("assets/sprites/chead-block-*-item-00.png")
	if err != nil {
		return nil, nil, nil, err
	}
	cbodyPaths, err := filepath.Glob("assets/sprites/cbody-block-*-item-00.png")
	if err != nil {
		return nil, nil, nil, err
	}
	paths = append(paths, cheadPaths...)
	paths = append(paths, cbodyPaths...)
	comsprPaths, err := filepath.Glob("assets/sprites/comspr-block-*-item-00.png")
	if err != nil {
		return nil, nil, nil, err
	}
	paths = append(paths, comsprPaths...)
	bigPicturePaths, err := filepath.Glob("assets/sprites/bigpic*-block-*-item-00.png")
	if err != nil {
		return nil, nil, nil, err
	}
	paths = append(paths, bigPicturePaths...)
	sceneCharacterPaths, err := filepath.Glob("assets/sprites/character-area-*.png")
	if err != nil {
		return nil, nil, nil, err
	}
	paths = append(paths, sceneCharacterPaths...)
	sort.Strings(paths)
	images := make(map[string]*ebiten.Image, len(paths))
	ids := make([]string, 0, len(paths))
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			return nil, nil, nil, err
		}
		decoded, err := png.Decode(file)
		file.Close()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("decode combat sprite %s: %w", path, err)
		}
		name := filepath.Base(path)
		images[name] = ebiten.NewImageFromImage(decoded)
		if strings.HasPrefix(name, "cpic") {
			ids = append(ids, name)
		}
	}
	animationData, err := os.ReadFile("assets/sprites/animation.json")
	if err != nil {
		return nil, nil, nil, err
	}
	var records []struct {
		Name  string `json:"name"`
		Delay uint32 `json:"delay"`
		X     int16  `json:"x"`
		Y     int16  `json:"y"`
	}
	if err := json.Unmarshal(animationData, &records); err != nil {
		return nil, nil, nil, fmt.Errorf("parse combat animation manifest: %w", err)
	}
	animations := make(map[string][]combatAnimation)
	for _, record := range records {
		file, err := os.Open(filepath.Join("assets/sprites", record.Name))
		if err != nil {
			return nil, nil, nil, err
		}
		decoded, err := png.Decode(file)
		file.Close()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("decode combat animation %s: %w", record.Name, err)
		}
		frameMarker := strings.Index(record.Name, "-frame-")
		if frameMarker < 0 {
			return nil, nil, nil, fmt.Errorf("animation asset %q has no frame marker", record.Name)
		}
		key := record.Name[:frameMarker]
		animations[key] = append(animations[key], combatAnimation{image: ebiten.NewImageFromImage(decoded), delay: record.Delay, x: record.X, y: record.Y})
	}
	return images, ids, animations, nil
}

func loadDungeonPreview(grid *geo.Grid) *mapdata.DungeonFloor {
	floor := mapdata.GenerateDungeon(*grid, 7, 13)
	return &floor
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

// loadMapPieceSets mirrors the verified reference LoadWalldef mapping while
// keeping selector interpretation out of the ECL VM and State packages.
func loadMapPieceSets(imagePath string, areaID uint8, selectors [3]uint16) (map[uint8]gfx.PieceSet, error) {
	if areaID < 1 || areaID > 6 {
		return nil, fmt.Errorf("map piece area %d is outside original range 1..6", areaID)
	}
	wallData, err := zipMember(imagePath, fmt.Sprintf("WALLDEF%d.DAX", areaID))
	if err != nil {
		return nil, err
	}
	symbolData, err := zipMember(imagePath, fmt.Sprintf("8X8D%d.DAX", areaID))
	if err != nil {
		return nil, err
	}
	wallBlocks, err := dax.Parse(wallData)
	if err != nil {
		return nil, fmt.Errorf("parse WALLDEF%d.DAX: %w", areaID, err)
	}
	symbolBlocks, err := dax.Parse(symbolData)
	if err != nil {
		return nil, fmt.Errorf("parse 8X8D%d.DAX: %w", areaID, err)
	}
	result := make(map[uint8]gfx.PieceSet, 3)
	for index, rawSelector := range selectors {
		if rawSelector == 0xFF {
			continue
		}
		if rawSelector > 0xFF {
			return nil, fmt.Errorf("map piece selector %d overflows byte", rawSelector)
		}
		setID := uint8(index + 1)
		pieceSet, err := gfx.ParsePieceSet(setID, uint8(rawSelector), wallBlocks, symbolBlocks)
		if err != nil {
			return nil, fmt.Errorf("piece set %d: %w", setID, err)
		}
		result[setID] = pieceSet
	}
	return result, nil
}

func loadGEOPreview(imagePath string, set, blockID uint8) (*geo.Grid, error) {
	if set < 2 || set > 6 {
		return nil, fmt.Errorf("GEO set %d is outside original range 2..6", set)
	}
	catalog, err := loadGEOCatalog(imagePath)
	if err != nil {
		return nil, err
	}
	grid, ok := catalog.Lookup(geo.MapRef{Set: set, BlockID: blockID})
	if !ok {
		return nil, fmt.Errorf("GEO%d block 0x%02X is not in original catalog", set, blockID)
	}
	return &grid, nil
}

func loadGEOCatalog(imagePath string) (geo.Catalog, error) {
	catalog := geo.NewCatalog()
	for set := uint8(2); set <= 6; set++ {
		data, err := zipMember(imagePath, fmt.Sprintf("GEO%d.DAX", set))
		if err != nil {
			return catalog, err
		}
		if err := catalog.AddDAX(set, data); err != nil {
			return catalog, err
		}
	}
	return catalog, nil
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

func loadMonsterAffects(data []byte) (map[uint8][]monster.AffectRecord, error) {
	blocks, err := dax.Parse(data)
	if err != nil {
		return nil, err
	}
	affects := make(map[uint8][]monster.AffectRecord, len(blocks))
	for _, block := range blocks {
		records, parseErr := monster.ParseAffects(block.Data)
		if parseErr != nil {
			return nil, fmt.Errorf("MON*SPC block 0x%02X: %w", block.Entry.ID, parseErr)
		}
		affects[block.Entry.ID] = records
	}
	return affects, nil
}

func loadMonsterItems(data []byte) (map[uint8][]monster.ItemRecord, error) {
	blocks, err := dax.Parse(data)
	if err != nil {
		return nil, err
	}
	items := make(map[uint8][]monster.ItemRecord, len(blocks))
	for _, block := range blocks {
		records, parseErr := monster.ParseItems(block.Data)
		if parseErr != nil {
			return nil, fmt.Errorf("MON*ITM block 0x%02X: %w", block.Entry.ID, parseErr)
		}
		items[block.Entry.ID] = records
	}
	return items, nil
}

// loadTreasureItemBlocks decodes the six original ITEM*.DAX containers into
// one global raw block map. TREASURE's final operand is the block ID, so the
// State adapter can resolve it without coupling ECL to ZIP/DAX I/O.
func loadTreasureItemBlocks(imagePath string) (map[uint16][]monster.ItemRecord, error) {
	areaData := make(map[uint8][]byte)
	for area := 1; area <= 6; area++ {
		member := fmt.Sprintf("ITEM%d.DAX", area)
		data, err := zipMember(imagePath, member)
		if err != nil {
			return nil, fmt.Errorf("load %s: %w", member, err)
		}
		areaData[uint8(area)] = data
	}
	return game.ParseTreasureItemBlocks(areaData)
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

func readOptional(path string) ([]byte, error) {
	if strings.TrimSpace(path) == "" {
		return nil, nil
	}
	return os.ReadFile(path)
}
