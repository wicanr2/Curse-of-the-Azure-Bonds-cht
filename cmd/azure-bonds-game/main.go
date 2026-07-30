package main

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"math"
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

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/area"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dungeon"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/etenfont"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/game"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/gfx"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/mapdata"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/sound"
	enginearea "github.com/wicanr2/golden-box-remake-engine/areamap"
	goldenengine "github.com/wicanr2/golden-box-remake-engine/engine"
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
	areaMapPreview      bool
	areaMapSymbols      []*ebiten.Image
	skyImages           [3]*ebiten.Image
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
	combatTerrain       map[string][]*ebiten.Image
	combatTerrainMode   string
	gamePack            *goldenengine.Pack
	combatFrame         *ebiten.Image
	adventureFrame      *ebiten.Image
	combatAnimations    map[string][]combatAnimation
	animationStart      time.Time
	deathOverlayStarted map[string]time.Time
	combatVisualSerial  uint64
	combatVisualStarted time.Time
	combatVisualElapsed time.Duration
	messageSnapshot     string
	messageStart        time.Time
	soundPlayer         *sound.Player
	pc98MusicDriver     []byte
	currentMusicTrack   string
	screenshotPath      string
	screenshotDone      bool
	screenshotFrames    int
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

func wallStampNativePosition(row, column int) (x, y int, ok bool) {
	if row < 0 || row > 10 || column < 0 || column > 10 {
		return 0, 0, false
	}
	return (column + 3) * 8, (row + 3) * 8, true
}

func (a *app) combatMove(dx, dy int) error {
	mode := selectCombatTerrainName(a.state.Area.InDungeon, a.combatTerrainMode)
	referenceCoordinates := a.state.CombatUsesReferenceCoordinates()
	terrain := func(x, y int) (int, bool) {
		entry, ok := combatMovementTerrainEntry(
			mode,
			a.dungeonFloor,
			a.state.WildernessFloor,
			a.state.MapX,
			a.state.MapY,
			referenceCoordinates,
			x,
			y,
		)
		if !ok || !entry.Passable() || entry.MoveCost == 0 {
			return 0, false
		}
		return int(entry.MoveCost), true
	}
	return a.state.CombatMoveWithTerrain(dx, dy, terrain)
}

func (a *app) combatLineTerrain() combat.LineTerrain {
	mode := selectCombatTerrainName(a.state.Area.InDungeon, a.combatTerrainMode)
	referenceCoordinates := a.state.CombatUsesReferenceCoordinates()
	return func(x, y int) combat.LineCell {
		entry, ok := combatMovementTerrainEntry(
			mode,
			a.dungeonFloor,
			a.state.WildernessFloor,
			a.state.MapX,
			a.state.MapY,
			referenceCoordinates,
			x,
			y,
		)
		return combat.LineCell{
			Valid:   ok,
			Reflect: ok && mode == "DUNGCOM" && entry.MoveCost == 0xFF,
		}
	}
}

func combatMovementTerrainEntry(mode string, dungeon *mapdata.DungeonFloor, wilderness mapdata.WildernessFloor, mapX, mapY int, referenceCoordinates bool, x, y int) (mapdata.BackgroundTile, bool) {
	switch mode {
	case "DUNGCOM":
		if dungeon == nil {
			return mapdata.BackgroundTile{}, false
		}
		if !referenceCoordinates {
			x, y = 18+x, 7+y
		}
		return dungeon.Entry(x, y)
	case "WILDCOM":
		if !referenceCoordinates {
			x, y = mapX+x-3, mapY+y-3
		}
		return wilderness.Entry(x, y)
	default:
		return mapdata.BackgroundTile{}, false
	}
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

func (a *app) playSoundEvent(event game.SoundEvent) {
	if a.soundPlayer != nil {
		a.soundPlayer.PlayEvent(sound.Event(event))
	}
}

func (a *app) syncSoundEvents() {
	for _, event := range a.state.ConsumeSoundEvents() {
		a.playSoundEvent(event)
	}
	for _, event := range a.state.ConsumeMusicEvents() {
		if event.Action == "play" {
			a.currentMusicTrack = event.TrackID
			if len(a.pc98MusicDriver) != 0 && a.soundPlayer != nil {
				track, found := a.gamePack.FindMusicTrack(event.TrackID)
				if !found {
					log.Printf("music track %q is not in the game pack", event.TrackID)
					continue
				}
				if err := a.soundPlayer.PlayPC98Track(
					a.pc98MusicDriver,
					int(track.ReferenceSelector),
				); err != nil {
					log.Printf("music track %q disabled: %v", event.TrackID, err)
				}
			}
		}
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
	if a.areaMapPreview {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
			a.areaMapPreview = false
		}
		return nil
	}
	a.syncSoundEvents()
	a.syncGeoMapRequest()
	a.syncLoadPiecesRequest()
	a.syncECLCallRequests()
	if event, ok := a.state.CombatVisualEvent(); ok {
		if a.combatVisualSerial != event.Serial {
			a.combatVisualSerial = event.Serial
			a.combatVisualStarted = time.Now()
		}
		if a.screenshotPath != "" {
			return nil
		}
		if err := a.state.AdvanceCombatVisual(time.Since(a.combatVisualStarted)); err != nil {
			return err
		}
		a.syncSoundEvents()
		if a.state.CombatVisualPending() {
			return nil
		}
	}
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
			if inpututil.IsKeyJustPressed(ebiten.KeyA) {
				a.areaMapPreview = true
				return nil
			}
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
	if a.state.ECLStringEditing() {
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
			return a.state.BackspaceECLString()
		}
		if chars := ebiten.InputChars(); len(chars) > 0 {
			if err := a.state.AppendECLString(chars); err != nil {
				a.state.Message = err.Error()
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			return a.state.SubmitECLString()
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
				spellID := a.state.CombatCastingSpell()
				if spellID == game.LightningBoltSpellID || spellID == game.StinkingCloudSpellID || spellID == game.CloudkillSpellID {
					return a.combatAction(func() error {
						return a.state.CombatCastWithTerrain(spellID, a.combatLineTerrain())
					})
				}
				return a.combatAction(func() error { return a.state.CombatCast(spellID) })
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
				return a.combatAction(func() error { return a.combatMove(0, -1) })
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
				return a.combatAction(func() error { return a.combatMove(1, 0) })
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
				return a.combatAction(func() error { return a.combatMove(0, 1) })
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
				return a.combatAction(func() error { return a.combatMove(-1, 0) })
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) && a.state.CombatCastingSpell() != 0 {
			a.state.CancelCombatCast()
			return nil
		}
		if a.state.CombatCastingSpell() == game.StinkingCloudSpellID ||
			a.state.CombatCastingSpell() == game.CloudkillSpellID ||
			a.state.CombatCastingSpell() == game.FireballSpellID ||
			a.state.CombatCastingSpell() == game.LightningBoltSpellID {
			if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
				return a.combatAction(func() error { return a.state.CombatMoveSpellTarget(0, -1) })
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
				return a.combatAction(func() error { return a.state.CombatMoveSpellTarget(1, 0) })
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
				return a.combatAction(func() error { return a.state.CombatMoveSpellTarget(0, 1) })
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
				return a.combatAction(func() error { return a.state.CombatMoveSpellTarget(-1, 0) })
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyS) && a.state.CombatCanCastMagicMissile() {
			return a.combatAction(func() error { return a.state.BeginCombatCast(game.MagicMissileSpellID) })
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyF) && a.state.CombatCanCastFireball() {
			return a.combatAction(func() error { return a.state.BeginCombatCast(game.FireballSpellID) })
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyL) && a.state.CombatCanCastLightningBolt() {
			return a.combatAction(func() error { return a.state.BeginCombatCast(game.LightningBoltSpellID) })
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyN) && a.state.CombatCanCastStinkingCloud() {
			return a.combatAction(func() error { return a.state.BeginCombatCast(game.StinkingCloudSpellID) })
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyK) && a.state.CombatCanCastCloudkill() {
			return a.combatAction(func() error { return a.state.BeginCombatCast(game.CloudkillSpellID) })
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
	wallFile, symbolFile := "", ""
	if pack, packErr := gamepack.Default(); packErr == nil {
		if definition, found := pack.FindMap(a.state.GeoMapSet, a.state.GeoMapBlock); found {
			wallFile, symbolFile = definition.WallFile, definition.SymbolFile
		}
	}
	sets, err := loadMapPieceSets(a.imagePath, a.state.GeoMapSet, wallFile, symbolFile, selectors)
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
	a.playSoundEvent(game.SoundStep)
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
			if _, _, ok := wallStampNativePosition(stamp.Row, stamp.Column); !ok {
				continue
			}
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
	if a.screenshotPath != "" && !a.screenshotDone {
		a.screenshotFrames++
		if a.screenshotFrames >= 3 {
			defer func() {
				a.screenshotDone = true
				output, err := os.Create(a.screenshotPath)
				if err != nil {
					log.Printf("create screenshot: %v", err)
					return
				}
				pixels := make([]byte, logicalWidth*logicalHeight*4)
				screen.ReadPixels(pixels)
				captured := &image.RGBA{
					Pix:    pixels,
					Stride: logicalWidth * 4,
					Rect:   image.Rect(0, 0, logicalWidth, logicalHeight),
				}
				if err := png.Encode(output, captured); err != nil {
					log.Printf("encode screenshot: %v", err)
				}
				if err := output.Close(); err != nil {
					log.Printf("close screenshot: %v", err)
				}
				os.Exit(0)
			}()
		}
	}
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
	if a.areaMapPreview {
		a.drawAreaMap(screen, white, cyan)
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
	if a.state.ECLStringEditing() {
		if a.state.Area.InDungeon {
			a.drawDungeonGame(screen, white, cyan)
			ebitenutil.DrawRect(screen, 8, 290, 624, 190, color.RGBA{0, 0, 0, 255})
			drawWrappedText(screen, a.state.Message, a.compactFace, 16, 316, 36, 20, 2, cyan)
			text.Draw(screen, "回答："+a.state.ECLStringValue()+"_", a.compactFace, 16, 360, white)
			text.Draw(screen, fmt.Sprintf("Enter：確認　Backspace：刪除　最多 %d 字", a.state.ECLStringMaxLength()), a.compactFace, 16, 438, cyan)
			a.drawOriginalAdventureFrame(screen)
			return
		}
		text.Draw(screen, a.state.Title, a.face, 32, 52, cyan)
		text.Draw(screen, a.state.LocationName, a.face, 32, 90, cyan)
		text.Draw(screen, a.state.Prompt, a.face, 32, 130, white)
		drawWrappedText(screen, a.state.Message, a.face, 56, 190, 22, 32, 4, cyan)
		text.Draw(screen, "回答："+a.state.ECLStringValue()+"_", a.face, 56, 340, white)
		text.Draw(screen, fmt.Sprintf("Enter：確認　Backspace：刪除　最多 %d 字", a.state.ECLStringMaxLength()), a.face, 56, 410, cyan)
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
		a.drawAreaMap(screen, white, cyan)
		return
	}
	if a.state.Mode == game.ModeWilderness && a.state.Message == "" && a.drawOverlandMap(screen, white, cyan) {
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

func (a *app) drawOverlandMap(screen *ebiten.Image, white, cyan color.Color) bool {
	pack, err := gamepack.Default()
	if err != nil {
		return false
	}
	definition, ok := pack.FindMapByKind("overland")
	if !ok {
		return false
	}
	stem := strings.ToLower(strings.TrimSuffix(definition.ImageFile, filepath.Ext(definition.ImageFile)))
	key := fmt.Sprintf("%s-block-%02X-item-00.png", stem, definition.GeometryBlock)
	mapImage := a.combatSprites[key]
	if mapImage == nil {
		return false
	}
	drawPanelFrame(screen, 8, 8, 624, 256)
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Scale(float64(definition.Scale), float64(definition.Scale))
	op.GeoM.Translate(16, 16)
	screen.DrawImage(mapImage, op)

	currentValue := a.state.Area.CurrentCity
	currentName := a.state.LocationName
	for _, point := range definition.Locations {
		if point.Value != currentValue {
			continue
		}
		x := 16 + point.X*definition.Scale
		y := 16 + point.Y*definition.Scale
		ink := color.RGBA{255, 255, 82, 255}
		ebitenutil.DrawRect(screen, float64(x-7), float64(y-7), 15, 3, ink)
		ebitenutil.DrawRect(screen, float64(x-7), float64(y+5), 15, 3, ink)
		ebitenutil.DrawRect(screen, float64(x-7), float64(y-7), 3, 15, ink)
		ebitenutil.DrawRect(screen, float64(x+5), float64(y-7), 3, 15, ink)
		if localized, found := pack.Text(point.MessageID, "zh-TW"); found {
			currentName = localized
		}
		break
	}

	drawPanelFrame(screen, 8, 264, 624, 184)
	drawPanelFrame(screen, 8, 448, 624, 28)
	text.Draw(screen, "月海諸地・世界地圖", a.face, 24, 296, cyan)
	text.Draw(screen, "目前位置："+currentName, a.face, 24, 328, white)
	timeLabel := strings.Split(a.state.GameTimeText(), "　")[0]
	text.Draw(screen, timeLabel, a.compactFace, 468, 294, cyan)
	for index, choice := range a.state.Choices {
		prefix := "  "
		if index == a.choiceCursor {
			prefix = "> "
		}
		text.Draw(screen, prefix+choice, a.face, 40, 366+index*30, white)
	}
	text.Draw(screen, "方向鍵選擇　Enter 確認", a.compactFace, 344, 470, cyan)
	return true
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
	if a.screenshotPath != "" {
		return a.state.Message
	}
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
			drawImageCover(screen, sprite, image.Rect(16, 16, 256, 256))
			a.drawOriginalAdventureFrame(screen)
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
	a.drawAdventureChrome(screen)
	drawImageCover(screen, frame.image, image.Rect(16, 16, 256, 256))
	a.drawOriginalAdventureFrame(screen)
	a.drawPictureMessage(screen)
	text.Draw(screen, "Enter：繼續", a.compactFace, 24, 470, color.RGBA{255, 255, 255, 255})
}

func drawImageCover(screen, source *ebiten.Image, destination image.Rectangle) {
	if source == nil || destination.Empty() {
		return
	}
	sourceWidth, sourceHeight := source.Bounds().Dx(), source.Bounds().Dy()
	scale := max(
		float64(destination.Dx())/float64(sourceWidth),
		float64(destination.Dy())/float64(sourceHeight),
	)
	target := screen.SubImage(destination).(*ebiten.Image)
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(
		float64(destination.Dx()-int(float64(sourceWidth)*scale))/2,
		float64(destination.Dy()-int(float64(sourceHeight)*scale))/2,
	)
	target.DrawImage(source, op)
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

func (a *app) drawCombatStoneFrame(screen *ebiten.Image) {
	if a.combatFrame == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Scale(2, 2)
	screen.DrawImage(a.combatFrame, op)
}

func (a *app) drawOriginalAdventureFrame(screen *ebiten.Image) {
	if a.adventureFrame == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Scale(2, 2)
	screen.DrawImage(a.adventureFrame, op)
}

func (a *app) drawAdventureChrome(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, 0, 0, 640, 480, color.RGBA{A: 255})
	a.drawOriginalAdventureFrame(screen)
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

func (a *app) drawAreaMap(screen *ebiten.Image, white, cyan color.Color) {
	drawPanelFrame(screen, 8, 8, 352, 352)
	drawPanelFrame(screen, 360, 8, 272, 352)
	drawPanelFrame(screen, 8, 360, 624, 88)
	drawPanelFrame(screen, 8, 448, 624, 28)
	if a.geoGrid == nil {
		text.Draw(screen, "尚未載入 GEO 區域資料", a.face, 24, 48, white)
		return
	}
	x, y, direction := a.state.DungeonGeometryView()
	if !a.areaMapPreview {
		x, y = a.state.MapX, a.state.MapY
	}
	view, err := enginearea.BuildOriginal(*a.geoGrid, x, y, int(direction))
	if err != nil || len(a.areaMapSymbols) < 20 {
		text.Draw(screen, "AREA 原始符號資料無法使用", a.compactFace, 24, 48, white)
		return
	}
	const tileSize, originX, originY = 16, 32, 32
	drawSymbol := func(item, screenX, screenY int) {
		if item < 0 || item >= len(a.areaMapSymbols) {
			return
		}
		options := &ebiten.DrawImageOptions{}
		options.GeoM.Scale(2, 2)
		options.GeoM.Translate(float64(originX+screenX*tileSize), float64(originY+screenY*tileSize))
		screen.DrawImage(a.areaMapSymbols[item], options)
	}
	for _, tile := range view.Tiles {
		drawSymbol(tile.Item, tile.ScreenX, tile.ScreenY)
	}
	drawSymbol(view.PartyItem, view.PartyScreenX, view.PartyScreenY)

	text.Draw(screen, "區域地圖", a.compactFace, 376, 38, cyan)
	text.Draw(screen, fmt.Sprintf("GEO%d／%02X　8X8D／CA", a.geoSet, a.geoBlock), a.compactFace, 376, 66, white)
	text.Draw(screen, fmt.Sprintf("位置 (%d,%d)", x, y), a.compactFace, 376, 102, white)
	text.Draw(screen, "方向 "+dungeonDirectionName(direction), a.compactFace, 376, 130, white)
	text.Draw(screen, "原版 11×11 AREA 視窗", a.compactFace, 376, 182, cyan)
	text.Draw(screen, "方向標記：隊伍位置", a.compactFace, 376, 210, color.RGBA{255, 255, 82, 255})
	text.Draw(screen, a.state.LocationName, a.compactFace, 24, 390, cyan)
	text.Draw(screen, "AREA 顯示目前區域的牆面與隊伍方向。", a.compactFace, 24, 422, white)
	text.Draw(screen, "A／Esc：返回探索", a.compactFace, 24, 470, cyan)
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
	if production {
		a.drawDungeonGame(screen, white, cyan)
		return
	}
	title := "地城結構預覽"
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
	controls := "方向鍵：移動　Q／R：轉向　P：撬鎖　K：敲擊術　B：撞門　D／Esc：返回"
	drawWrappedText(screen, controls, a.compactFace, 24, 434, 36, 20, 2, cyan)
}

func (a *app) drawDungeonGame(screen *ebiten.Image, white, cyan color.Color) {
	// DOS oracle: the native 320px top row is split at x=128, not near the
	// centre. At 2x the first-person panel is 256px and the roster is 384px.
	// The extra 80 vertical remake pixels enlarge only the text region.
	ebitenutil.DrawRect(screen, 0, 0, 640, 480, color.RGBA{0, 0, 0, 255})
	_, _, direction := a.state.DungeonGeometryView()
	background, backgroundErr := gfx.BuildBackground(
		a.state.Area.OutdoorSkyColor,
		a.state.Area.IndoorSkyColor,
		a.state.DungeonWallRoof,
		int(a.state.GameTimeDisplay().Hour),
		direction,
	)
	if backgroundErr == nil {
		for _, rectangle := range background.Rects {
			ebitenutil.DrawRect(screen,
				float64(rectangle.X*2), float64(rectangle.Y*2),
				float64(rectangle.Width*2), float64(rectangle.Height*2),
				gfx.EGA16[rectangle.PaletteIndex])
		}
		for _, overlay := range background.Overlays {
			index := int(overlay.Kind)
			if index < 0 || index >= len(a.skyImages) || a.skyImages[index] == nil {
				continue
			}
			options := &ebiten.DrawImageOptions{}
			options.Filter = ebiten.FilterNearest
			options.GeoM.Scale(2, 2)
			options.GeoM.Translate(float64((overlay.Column+1)*16), float64((overlay.Row+1)*16))
			screen.DrawImage(a.skyImages[index], options)
		}
	}

	for _, stamp := range a.wallPreview {
		nativeX, nativeY, ok := wallStampNativePosition(stamp.row, stamp.column)
		if !ok {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest
		op.GeoM.Scale(2, 2)
		op.GeoM.Translate(float64(nativeX*2), float64(nativeY*2))
		screen.DrawImage(stamp.image, op)
	}

	text.Draw(screen, "姓名", a.compactFace, 272, 38, white)
	text.Draw(screen, "AC", a.compactFace, 468, 38, white)
	text.Draw(screen, "HP", a.compactFace, 586, 38, white)
	for index, fighter := range a.state.PartyFighters() {
		if index >= 8 {
			break
		}
		text.Draw(screen, fighter.Name, a.compactFace, 272, 68+index*20, cyan)
		text.Draw(screen, strconv.Itoa(fighter.ArmorClass), a.compactFace, 472, 68+index*20, cyan)
		text.Draw(screen, strconv.Itoa(fighter.HitPoints), a.compactFace, 588, 68+index*20, cyan)
	}

	status := fmt.Sprintf("(%d,%d) %s %02d:%02d", a.dungeonX, a.dungeonY,
		dungeonDirectionName(direction), a.state.GameTimeDisplay().Hour, a.state.GameTimeDisplay().Minute)
	text.Draw(screen, status, a.compactFace, 272, 254, cyan)
	if a.state.Message != "" {
		drawWrappedText(screen, a.state.Message, a.face, 8, 302, 38, 24, 5, white)
	} else {
		text.Draw(screen, a.state.LocationName, a.face, 8, 302, cyan)
	}
	if a.dungeonDoorMenu {
		text.Draw(screen, "上鎖的門：B 撞門　P 撬鎖　N 敲擊　Esc 離開", a.compactFace, 8, 430, color.RGBA{255, 255, 82, 255})
	}
	text.Draw(screen, "↑前進　K/M轉向　S搜索　E紮營　P撬鎖　N敲擊　B撞門", a.compactFace, 8, 472, cyan)
	a.drawOriginalAdventureFrame(screen)
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
	// Preserve the original 320x184 upper combat layout at exactly 2x:
	// 8px frame + 7*24px battlefield + 8px divider + 128px status + 8px
	// frame. The extra 80 logical pixels belong to the CJK combat log.
	const (
		battlefieldX    = 16
		battlefieldY    = 16
		battlefieldSize = 336
		combatLogY      = 368
		combatFooterY   = 448
	)
	battlefield := ebiten.NewImage(battlefieldSize, battlefieldSize)
	battlefield.Fill(color.RGBA{R: 82, G: 82, B: 82, A: 255})
	terrainName := selectCombatTerrainName(a.state.Area.InDungeon, a.combatTerrainMode)
	if len(a.combatTerrain[terrainName]) > 0 {
		for row := 0; row < 7; row++ {
			for column := 0; column < 7; column++ {
				entry, ok := combatTerrainEntry(
					terrainName,
					a.dungeonFloor,
					a.state.WildernessFloor,
					a.state.MapX,
					a.state.MapY,
					column,
					row,
				)
				if !ok {
					continue
				}
				for _, layer := range combatTerrainLayers(terrainName, entry) {
					tiles := a.combatTerrain[layer.Atlas]
					if layer.Index < 0 || layer.Index >= len(tiles) {
						continue
					}
					op := &ebiten.DrawImageOptions{}
					op.Filter = ebiten.FilterNearest
					op.GeoM.Scale(2, 2)
					op.GeoM.Translate(float64(column*48), float64(row*48))
					battlefield.DrawImage(tiles[layer.Index], op)
				}
			}
		}
	}
	battlefieldOp := &ebiten.DrawImageOptions{}
	battlefieldOp.GeoM.Translate(battlefieldX, battlefieldY)
	screen.DrawImage(battlefield, battlefieldOp)
	ebitenutil.DrawRect(screen, 368, 16, 256, 336, color.RGBA{R: 82, G: 82, B: 82, A: 255})
	a.drawCombatStoneFrame(screen)
	ebitenutil.DrawRect(screen, 0, combatLogY, 640, 80, color.RGBA{A: 255})
	ebitenutil.DrawRect(screen, 0, combatFooterY, 640, 32, color.RGBA{A: 255})
	combatMessage := a.state.CombatMessage()
	if event, ok := a.state.CombatVisualEvent(); ok {
		combatMessage = a.combatVisualMessage(event, a.combatVisualFrame(event), combatMessage)
	}
	drawWrappedText(screen, combatMessage, a.compactFace, 8, 392, 39, 20, 3, white)
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
			footprint := combat.FootprintForSize(fighter.CombatSize)
			minX, maxX, minY, maxY = fighter.CombatX, fighter.CombatX+footprint.Width-1, fighter.CombatY, fighter.CombatY+footprint.Height-1
			havePosition = true
		} else {
			footprint := combat.FootprintForSize(fighter.CombatSize)
			minX, maxX = min(minX, fighter.CombatX), max(maxX, fighter.CombatX)
			maxX = max(maxX, fighter.CombatX+footprint.Width-1)
			minY, maxY = min(minY, fighter.CombatY), max(maxY, fighter.CombatY+footprint.Height-1)
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
	a.drawCombatPersistentAreas(battlefield, camera)
	partyIndex, enemyIndex := 0, 0
	targets := a.state.CombatTargets()
	spellTargets := a.state.CombatSpellTargets()
	for _, fighter := range a.state.CombatFighters() {
		if event, ok := a.state.CombatVisualEvent(); ok {
			frame := a.combatVisualFrame(event)
			if impact, preserve := combatVisualPreservedImpact(event, frame, fighter.ID); preserve {
				// Battle rules have already resolved HP. Preserve the target's
				// pre-impact map presence until the visual commit reaches death.
				fighter.HitPoints = 1
				fighter.DeathOverlay = false
				fighter.DownedCorpse = false
				fighter.HasCombatPosition = true
				fighter.CombatX, fighter.CombatY = impact.To.X, impact.To.Y
			}
		}
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
			tile = mirroredCombatAnchor(tile)
			x, y := tile.X*48, tile.Y*48
			if !fighter.DownedCorpse && !fighter.DeathOverlay {
				if event, ok := a.state.CombatVisualEvent(); ok {
					frame := a.combatVisualFrame(event)
					fighter.IconAttack = event.ActorID == fighter.ID && frame.Phase == combat.VisualWindup
				}
				a.drawFighterSprite(battlefield, fighter, partyIndex, x, y)
			}
			a.drawFighterDeathOverlay(battlefield, fighter, x, y)
			selected := false
			if (a.state.CombatCastingSpell() == game.CureLightWoundsSpellID || a.state.CombatCastingSpell() == game.ProtectionFromEvilSpellID || (a.state.CombatCastingSpell() == game.ProtectionFromGoodSpellID && !a.state.CombatSpellTargetsEnemy())) && a.state.CombatSpellTargetIndex() < len(spellTargets) && spellTargets[a.state.CombatSpellTargetIndex()].ID == fighter.ID {
				selected = true
			}
			isActive := activeOK && (active.ID == fighter.ID ||
				(active.Side == fighter.Side && active.Name == fighter.Name &&
					active.CombatX == fighter.CombatX && active.CombatY == fighter.CombatY))
			a.drawCombatSpriteMarker(battlefield, fighter, isActive, selected, x, y)
			partyIndex++
			continue
		}
		tile := combat.FormationTile(fighter.Side, enemyIndex)
		if fighter.HasCombatPosition || fighter.DeathOverlay || fighter.DownedCorpse {
			tile = combat.TilePoint{X: fighter.CombatX, Y: fighter.CombatY}
		}
		tile = camera.Apply(tile)
		tile = mirroredCombatAnchor(tile)
		x, y := tile.X*48, tile.Y*48
		if !fighter.DownedCorpse && !fighter.DeathOverlay {
			if event, ok := a.state.CombatVisualEvent(); ok {
				frame := a.combatVisualFrame(event)
				fighter.IconAttack = event.ActorID == fighter.ID && frame.Phase == combat.VisualWindup
			}
			a.drawFighterSprite(battlefield, fighter, enemyIndex, x, y)
		}
		a.drawFighterDeathOverlay(battlefield, fighter, x, y)
		selected := false
		if (a.state.CombatCastingSpell() == 0 || a.state.CombatSpellTargetsEnemy()) && len(targets) > 0 && a.state.CombatTargetIndex() < len(targets) && targets[a.state.CombatTargetIndex()].ID == fighter.ID {
			selected = true
		}
		isActive := activeOK && (active.ID == fighter.ID ||
			(active.Side == fighter.Side && active.Name == fighter.Name &&
				active.CombatX == fighter.CombatX && active.CombatY == fighter.CombatY))
		a.drawCombatSpriteMarker(battlefield, fighter, isActive, selected, x, y)
		enemyIndex++
	}
	if point, ok := a.state.CombatSpellTargetPoint(); ok {
		tile := mirroredCombatAnchor(camera.Apply(point))
		x, y := tile.X*48, tile.Y*48
		marker := color.RGBA{R: 255, G: 255, B: 255, A: 255}
		ebitenutil.DrawRect(battlefield, float64(x+2), float64(y+2), 44, 2, marker)
		ebitenutil.DrawRect(battlefield, float64(x+2), float64(y+44), 44, 2, marker)
		ebitenutil.DrawRect(battlefield, float64(x+2), float64(y+2), 2, 44, marker)
		ebitenutil.DrawRect(battlefield, float64(x+44), float64(y+2), 2, 44, marker)
	}
	if event, ok := a.state.CombatVisualEvent(); ok {
		a.drawCombatVisual(battlefield, event, a.combatVisualFrame(event), camera)
	}
	screen.DrawImage(battlefield, battlefieldOp)
	a.drawCombatStoneFrame(screen)
	if activeOK {
		statusGreen := color.RGBA{R: 92, G: 255, B: 92, A: 255}
		statusYellow := color.RGBA{R: 255, G: 255, B: 82, A: 255}
		text.Draw(screen, active.Name, a.compactFace, 370, 30, cyan)
		text.Draw(screen, "生命值", a.compactFace, 370, 62, statusGreen)
		text.Draw(screen, strconv.Itoa(active.HitPoints), a.compactFace, 450, 62, statusYellow)
		text.Draw(screen, "防護等級", a.compactFace, 370, 94, statusGreen)
		text.Draw(screen, strconv.Itoa(active.ArmorClass), a.compactFace, 466, 94, statusYellow)
	}
	spellHints := make([]string, 0, 7)
	if a.state.CombatCastingSpell() != 0 {
		if a.state.CombatCastingSpell() == game.BlessSpellID {
			text.Draw(screen, "確認施法：Enter　取消：Esc", a.face, 32, 350, cyan)
			return
		}
		if a.state.CombatCastingSpell() == game.StinkingCloudSpellID ||
			a.state.CombatCastingSpell() == game.CloudkillSpellID ||
			a.state.CombatCastingSpell() == game.FireballSpellID ||
			a.state.CombatCastingSpell() == game.LightningBoltSpellID {
			prompt := "選擇火球中心"
			if a.state.CombatCastingSpell() == game.LightningBoltSpellID {
				prompt = "選擇閃電方向格"
			} else if a.state.CombatCastingSpell() == game.StinkingCloudSpellID {
				prompt = "選擇惡臭雲霧西北角"
			} else if a.state.CombatCastingSpell() == game.CloudkillSpellID {
				prompt = "選擇致命毒雲中心"
			}
			text.Draw(screen, prompt+"：方向鍵移動　Enter：確認　Esc：取消", a.face, 32, 350, cyan)
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
	if a.state.CombatCanCastFireball() {
		spellHints = append(spellHints, "F火球")
	}
	if a.state.CombatCanCastLightningBolt() {
		spellHints = append(spellHints, "L閃電")
	}
	if a.state.CombatCanCastStinkingCloud() {
		spellHints = append(spellHints, "N臭雲")
	}
	if a.state.CombatCanCastCloudkill() {
		spellHints = append(spellHints, "K毒雲")
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
	footerStatus := ""
	if len(targets) > 0 && a.state.CombatTargetIndex() < len(targets) {
		footerStatus = "目標：" + targets[a.state.CombatTargetIndex()].Name
	}
	text.Draw(screen, footerStatus, a.compactFace, 8, 462, color.RGBA{R: 255, G: 82, B: 255, A: 255})
	text.Draw(screen, "移動　查看　瞄準　使用　施法　快速　結束", a.compactFace, 8, 478, cyan)
	if len(spellHints) > 0 {
		text.Draw(screen, "快捷："+strings.Join(spellHints, "　"), a.compactFace, 378, 340, cyan)
	}
}

func (a *app) drawCombatPersistentAreas(screen *ebiten.Image, camera combat.CombatCamera) {
	hiddenAreaID := uint64(0)
	if event, ok := a.state.CombatVisualEvent(); ok && event.PersistentAreaID != 0 {
		frame := a.combatVisualFrame(event)
		if frame.Phase == combat.VisualWindup || frame.Phase == combat.VisualTravel {
			hiddenAreaID = event.PersistentAreaID
		}
	}
	for _, area := range a.state.CombatPersistentAreas() {
		if area.ID == hiddenAreaID {
			continue
		}
		trigger := ""
		switch area.Kind {
		case combat.PersistentAreaStinkingCloud:
			trigger = "stinking_cloud"
		case combat.PersistentAreaCloudkill:
			trigger = "cloudkill"
		}
		definition, found := a.findCombatVisual(trigger, "area")
		if !found || len(definition.Frames) == 0 {
			continue
		}
		frame := definition.Frames[0]
		atlas := strings.ToUpper(strings.TrimSuffix(frame.SourceFile, filepath.Ext(frame.SourceFile)))
		tiles := a.combatTerrain[atlas]
		if int(frame.Block) >= len(tiles) || tiles[frame.Block] == nil {
			continue
		}
		scale := definition.Scale
		if scale < 1 {
			scale = 1
		}
		for _, cell := range area.Cells {
			tile := mirroredCombatAnchor(camera.Apply(combat.TilePoint{X: cell.X, Y: cell.Y}))
			op := &ebiten.DrawImageOptions{}
			op.Filter = ebiten.FilterNearest
			op.GeoM.Scale(float64(scale), float64(scale))
			op.GeoM.Translate(float64(tile.X*48), float64(tile.Y*48))
			screen.DrawImage(tiles[frame.Block], op)
		}
	}
}

func selectCombatTerrainName(inDungeon bool, override string) string {
	if override == "DUNGCOM" || override == "WILDCOM" || override == "RANDCOM" {
		return override
	}
	if inDungeon {
		return "DUNGCOM"
	}
	return "WILDCOM"
}

// mirroredCombatAnchor converts the original combat column to the remake's
// mirrored view. CPIC coordinates remain the upper-left drawing anchor even
// for multi-cell monsters; their footprint expands right and down from it.
func mirroredCombatAnchor(tile combat.TilePoint) combat.TilePoint {
	return combat.TilePoint{X: 6 - tile.X, Y: tile.Y}
}

type combatTerrainLayer struct {
	Atlas string
	Index int
}

func combatTerrainLayers(mode string, entry mapdata.BackgroundTile) []combatTerrainLayer {
	index := int(entry.TileIndex)
	switch mode {
	case "DUNGCOM":
		if index >= 0 && index < 25 {
			return []combatTerrainLayer{{Atlas: "DUNGCOM", Index: index}}
		}
		// The reference BackgroundTiles table uses one global graphic
		// namespace. IDs 0x22..0x27 select RANDCOM items 0..5. They are
		// transparent furniture/obstacle overlays placed only after the
		// dungeon generator has found an open DUNGCOM floor tile (0x16).
		if index >= 0x22 && index <= 0x27 {
			return []combatTerrainLayer{
				{Atlas: "DUNGCOM", Index: 0x16},
				{Atlas: "RANDCOM", Index: index - 0x22},
			}
		}
	case "WILDCOM":
		if index >= 0 && index < 34 {
			return []combatTerrainLayer{{Atlas: "WILDCOM", Index: index}}
		}
	}
	return nil
}

func combatTerrainEntry(mode string, dungeon *mapdata.DungeonFloor, wilderness mapdata.WildernessFloor, mapX, mapY, column, row int) (mapdata.BackgroundTile, bool) {
	switch mode {
	case "DUNGCOM":
		if dungeon == nil {
			return mapdata.BackgroundTile{}, false
		}
		return dungeon.Entry(18+column, 7+row)
	case "WILDCOM":
		return wilderness.Entry(mapX+column-3, mapY+row-3)
	default:
		// RANDCOM is an overlay/decor atlas, not a complete floor. Its
		// placement routine remains a separate reverse-engineering boundary.
		return mapdata.BackgroundTile{}, false
	}
}

func (a *app) drawCombatSpriteMarker(screen *ebiten.Image, fighter combat.Fighter, active, selected bool, x, y int) {
	if !active && !selected {
		return
	}
	// The reference renderer identifies the cursor with a square tile-sized
	// frame. It does not draw red/blue team bars across every combatant.
	marker := color.RGBA{R: 232, G: 232, B: 224, A: 255}
	if selected {
		marker = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	footprint := combat.FootprintForSize(fighter.CombatSize)
	width, height := footprint.Width*48, footprint.Height*48
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(width), 2, marker)
	ebitenutil.DrawRect(screen, float64(x), float64(y+height-2), float64(width), 2, marker)
	ebitenutil.DrawRect(screen, float64(x), float64(y), 2, float64(height), marker)
	ebitenutil.DrawRect(screen, float64(x+width-2), float64(y), 2, float64(height), marker)
}

func (a *app) drawCombatVisual(screen *ebiten.Image, event combat.VisualEvent, frame combat.VisualFrame, camera combat.CombatCamera) {
	fromX, fromY, toX, toY, x, y := combatVisualPoint(event, frame, camera)
	switch frame.Phase {
	case combat.VisualTravel:
		switch event.Kind {
		case combat.VisualMissile:
			definition, _ := a.findCombatVisual("missile", "travel")
			key, flip := combatArrowSprite(definition, event, camera)
			if !a.drawCombatProjectileSprite(screen, key, flip, definition.Scale, x, y) {
				ebitenutil.DrawLine(screen, fromX, fromY, x, y, color.RGBA{R: 255, G: 244, B: 180, A: 255})
			}
		case combat.VisualMagicMissile, combat.VisualAreaSpell, combat.VisualLineSpell:
			trigger := "magic_missile"
			if event.Effect != "" {
				trigger = event.Effect
			}
			definition, _ := a.findCombatVisual(trigger, "travel")
			key, flip := combatMagicMissileSprite(definition, event, frame)
			if !a.drawCombatProjectileSprite(screen, key, flip, definition.Scale, x, y) {
				ebitenutil.DrawRect(screen, x-4, y-4, 8, 8, color.RGBA{R: 126, G: 205, B: 255, A: 255})
			}
		}
	case combat.VisualSegmentTravel:
		if event.Kind == combat.VisualLineSpell && frame.SegmentIndex >= 0 &&
			frame.SegmentIndex < len(event.Segments) {
			definition, _ := a.findCombatVisual(event.Effect, "line")
			segment := event.Segments[frame.SegmentIndex]
			key, flip := combatPathSequenceSprite(definition, segment.From, segment.To, frame.Progress)
			a.drawCombatProjectileSprite(screen, key, flip, definition.Scale, x, y)
		}
	case combat.VisualImpact:
		if (event.Kind == combat.VisualMagicMissile || event.Kind == combat.VisualAreaSpell ||
			event.Kind == combat.VisualLineSpell) && event.Hit {
			trigger := "magic_missile"
			if event.Effect != "" {
				trigger = event.Effect
			}
			definition, _ := a.findCombatVisual(trigger, "impact")
			key, flip := combatMagicImpactSprite(definition, frame)
			a.drawCombatProjectileSprite(screen, key, flip, definition.Scale, toX, toY)
		}
	case combat.VisualDeath:
		iconKey := "comspr-block-8B-item-00.png"
		if int(frame.Progress*9)%2 == 1 {
			iconKey = "comspr-block-19-item-00.png"
		}
		if icon := a.combatSprites[iconKey]; icon != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(2, 2)
			op.GeoM.Translate(toX-24, toY-24)
			screen.DrawImage(icon, op)
		}
	}
}

func (a *app) findCombatVisual(trigger, phase string) (goldenengine.CombatVisualDefinition, bool) {
	if a.gamePack == nil {
		return goldenengine.CombatVisualDefinition{}, false
	}
	return a.gamePack.FindCombatVisual(trigger, phase)
}

func (a *app) drawCombatProjectileSprite(screen *ebiten.Image, key string, flip bool, scale int, x, y float64) bool {
	sprite := a.combatSprites[key]
	if sprite == nil {
		return false
	}
	if scale < 1 {
		scale = 1
	}
	width := float64(sprite.Bounds().Dx() * scale)
	height := float64(sprite.Bounds().Dy() * scale)
	op := &ebiten.DrawImageOptions{}
	if flip {
		op.GeoM.Scale(float64(-scale), float64(scale))
		op.GeoM.Translate(x+width/2, y-height/2)
	} else {
		op.GeoM.Scale(float64(scale), float64(scale))
		op.GeoM.Translate(x-width/2, y-height/2)
	}
	screen.DrawImage(sprite, op)
	return true
}

func combatArrowSprite(definition goldenengine.CombatVisualDefinition, event combat.VisualEvent, camera combat.CombatCamera) (key string, flip bool) {
	from := mirroredCombatAnchor(camera.Apply(event.From))
	to := mirroredCombatAnchor(camera.Apply(event.To))
	frame, ok := definition.FrameForDirection(uint8(combatProjectileDirection(to.X-from.X, to.Y-from.Y)))
	if !ok {
		return "", false
	}
	return combatVisualSpriteKey(frame)
}

// combatProjectileDirection matches the original clockwise numbering:
// north=0, north-east=1, east=2 ... north-west=7.
func combatProjectileDirection(dx, dy int) int {
	if dx == 0 && dy == 0 {
		return 0
	}
	angle := math.Atan2(float64(dx), float64(-dy))
	if angle < 0 {
		angle += 2 * math.Pi
	}
	return int(math.Floor(angle/(math.Pi/4)+0.5)) % 8
}

func combatMagicMissileSprite(definition goldenengine.CombatVisualDefinition, event combat.VisualEvent, frame combat.VisualFrame) (key string, flip bool) {
	steps := max(absInt(event.To.X-event.From.X), absInt(event.To.Y-event.From.Y)) * 3
	if steps < 1 {
		steps = 1
	}
	if len(definition.Frames) == 0 {
		return "", false
	}
	return combatVisualSpriteKey(definition.Frames[int(frame.Progress*float64(steps))%len(definition.Frames)])
}

func combatMagicImpactSprite(definition goldenengine.CombatVisualDefinition, frame combat.VisualFrame) (key string, flip bool) {
	if len(definition.Frames) == 0 {
		return "", false
	}
	index := min(int(frame.Progress*float64(len(definition.Frames))), len(definition.Frames)-1)
	return combatVisualSpriteKey(definition.Frames[index])
}

func combatPathSequenceSprite(definition goldenengine.CombatVisualDefinition, from, to combat.TilePoint, progress float64) (key string, flip bool) {
	if len(definition.Frames) == 0 {
		return "", false
	}
	steps := max(absInt(to.X-from.X), absInt(to.Y-from.Y)) * 3
	if steps < 1 {
		steps = 1
	}
	return combatVisualSpriteKey(definition.Frames[int(progress*float64(steps))%len(definition.Frames)])
}

func combatVisualSpriteKey(frame goldenengine.CombatVisualFrame) (key string, flip bool) {
	extension := filepath.Ext(frame.SourceFile)
	stem := strings.ToLower(strings.TrimSuffix(frame.SourceFile, extension))
	return fmt.Sprintf("%s-block-%02X-item-00.png", stem, frame.Block), frame.FlipX
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func (a *app) combatVisualFrame(event combat.VisualEvent) combat.VisualFrame {
	if a.screenshotPath != "" {
		return event.FrameAt(a.combatVisualElapsed)
	}
	return event.FrameAt(time.Since(a.combatVisualStarted))
}

func combatVisualPoint(event combat.VisualEvent, frame combat.VisualFrame, camera combat.CombatCamera) (fromX, fromY, toX, toY, x, y float64) {
	source := event.From
	target := event.To
	if frame.Phase == combat.VisualSegmentTravel && frame.SegmentIndex >= 0 &&
		frame.SegmentIndex < len(event.Segments) {
		source = event.Segments[frame.SegmentIndex].From
		target = event.Segments[frame.SegmentIndex].To
	} else if impact, ok := event.Impact(frame); ok && frame.Phase != combat.VisualTravel {
		target = impact.To
	}
	from := mirroredCombatAnchor(camera.Apply(source))
	to := mirroredCombatAnchor(camera.Apply(target))
	fromX, fromY = float64(from.X*48+24), float64(from.Y*48+24)
	toX, toY = float64(to.X*48+24), float64(to.Y*48+24)
	x = fromX + (toX-fromX)*frame.Progress
	y = fromY + (toY-fromY)*frame.Progress
	return
}

func combatVisualPreservedImpact(event combat.VisualEvent, frame combat.VisualFrame, targetID string) (combat.VisualImpactTarget, bool) {
	for index := 0; index < event.ImpactCount(); index++ {
		impact, _ := event.ImpactAt(index)
		if impact.TargetID != targetID || !impact.Killed {
			continue
		}
		switch {
		case frame.Phase == combat.VisualHandoff || frame.Done:
			return combat.VisualImpactTarget{}, false
		case index < frame.ResolvedImpacts:
			return combat.VisualImpactTarget{}, false
		case frame.ImpactIndex == index && frame.Phase == combat.VisualDeath:
			return combat.VisualImpactTarget{}, false
		default:
			return impact, true
		}
	}
	return combat.VisualImpactTarget{}, false
}

func (a *app) combatVisualMessage(event combat.VisualEvent, frame combat.VisualFrame, fallback string) string {
	if event.Effect == "cloudkill" {
		impact, ok := event.Impact(frame)
		if !ok || (frame.Phase != combat.VisualImpact && frame.Phase != combat.VisualCommit && frame.Phase != combat.VisualDeath) {
			return fallback
		}
		name := impact.TargetID
		for _, fighter := range a.state.CombatFighters() {
			if fighter.ID == impact.TargetID {
				name = fighter.Name
				break
			}
		}
		if impact.Saved {
			return name + " 抵抗了致命毒氣。"
		}
		return name + " 中毒身亡。"
	}
	if event.Effect == "stinking_cloud" {
		impact, ok := event.Impact(frame)
		if !ok || (frame.Phase != combat.VisualImpact && frame.Phase != combat.VisualCommit) {
			return fallback
		}
		name := impact.TargetID
		for _, fighter := range a.state.CombatFighters() {
			if fighter.ID == impact.TargetID {
				name = fighter.Name
				break
			}
		}
		if impact.Saved {
			return name + " 開始咳嗽。"
		}
		return fmt.Sprintf("%s 因噁心而窒息乾嘔，將有 %d 回合無法行動。", name, impact.Damage)
	}
	if event.Kind != combat.VisualLineSpell {
		return fallback
	}
	impact, ok := event.Impact(frame)
	if !ok {
		return fallback
	}
	name := impact.TargetID
	for _, fighter := range a.state.CombatFighters() {
		if fighter.ID == impact.TargetID {
			name = fighter.Name
			break
		}
	}
	switch frame.Phase {
	case combat.VisualImpact:
		return fmt.Sprintf("%s 受到 %d 點電擊傷害。", name, impact.Damage)
	case combat.VisualCommit:
		if impact.Saved {
			return name + " 的法術豁免成功，傷害減半。"
		}
		return name + " 的法術豁免失敗。"
	default:
		return fallback
	}
}

func prepareCombatVisualDemo(state *game.State, kind string) (time.Duration, error) {
	fireballDemo := strings.HasPrefix(kind, "fireball-")
	lightningDemo := strings.HasPrefix(kind, "lightning-")
	stinkingCloudDemo := strings.HasPrefix(kind, "stinking-cloud-")
	cloudkillDemo := strings.HasPrefix(kind, "cloudkill-")
	hero := combat.Fighter{
		ID: "demo-hero", Name: "弓手艾琳", Side: combat.SideParty,
		HitPoints: 30, MaxHitPoints: 30, ArmorClass: 0,
		AttackBonus: 30, DamageDiceCount: 1, DamageDiceSides: 4,
		InitiativeBonus: 30, HasCombatPosition: true, CombatX: 1, CombatY: 2,
		SavingThrows: []uint8{10, 10, 10, 10, 10},
		HasPartyIcon: true, PartyIconSize: 2,
	}
	enemy := combat.Fighter{
		ID: "demo-orc", Name: "半獸人", Side: combat.SideEnemy,
		HitPoints: 30, MaxHitPoints: 30, ArmorClass: 10,
		AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 4,
		InitiativeBonus: -20, HasCombatPosition: true, CombatX: 5, CombatY: 2,
		SavingThrows: []uint8{10, 10, 10, 10, 10},
		SpriteSet:    1, SpriteBlock: 1,
	}
	switch kind {
	case "melee":
		hero.Name = "戰士艾琳"
	case "bow":
		hero.MissileWeapon = true
	case "kill":
		hero.Name = "戰士艾琳"
		enemy.HitPoints, enemy.MaxHitPoints = 1, 1
	case "magic", "magic-impact":
		hero.InitiativeBonus = -20
		enemy.Name = "散提爾法師"
		enemy.InitiativeBonus = 30
		enemy.MonsterSpellUses[0] = 1
		enemy.MonsterSpellIDs = []uint8{combat.MonsterMagicMissileSpellID}
	case "fireball-travel", "fireball-impact-1", "fireball-impact-2":
		hero.Name = "法師艾琳"
	case "lightning-target-hit", "lightning-line-continue", "lightning-reflect":
		hero.Name = "法師艾琳"
	case "stinking-cloud-travel", "stinking-cloud-persistent":
		hero.Name = "法師艾琳"
	case "cloudkill-travel", "cloudkill-persistent":
		hero.Name = "法師艾琳"
		hero.HitDice = 7
		enemy.HitDice = 4
	default:
		return 0, fmt.Errorf("unknown combat visual demo %q", kind)
	}
	heroes := []combat.Fighter{hero}
	enemies := []combat.Fighter{enemy}
	if fireballDemo || lightningDemo || stinkingCloudDemo || cloudkillDemo {
		ally := combat.Fighter{
			ID: "demo-ally", Name: "戰士布蘭", Side: combat.SideParty,
			HitPoints: 100, MaxHitPoints: 100, ArmorClass: 2,
			InitiativeBonus: -10, HasCombatPosition: true, CombatX: 2, CombatY: 4,
			SavingThrows: []uint8{10, 10, 10, 10, 10},
			HasPartyIcon: true, PartyHeadBlock: 1, PartyBodyBlock: 1, PartyIconSize: 2,
		}
		enemy.HitPoints, enemy.MaxHitPoints = 100, 100
		enemy.ID = "demo-orc-a"
		enemy.CombatX, enemy.CombatY = 3, 2
		secondEnemy := enemy
		secondEnemy.ID, secondEnemy.Name = "demo-orc-b", "半獸人隊長"
		if lightningDemo {
			secondEnemy.CombatX, secondEnemy.CombatY = 5, 2
		} else {
			secondEnemy.CombatX, secondEnemy.CombatY = 4, 3
		}
		heroes = append(heroes, ally)
		enemies = append(enemies, secondEnemy)
		abilities := party.Abilities{Strength: 12, Intelligence: 16, Wisdom: 12, Dexterity: 14, Constitution: 12, Charisma: 12}
		spellID := game.FireballSpellID
		if lightningDemo {
			spellID = game.LightningBoltSpellID
		} else if stinkingCloudDemo {
			spellID = game.StinkingCloudSpellID
		} else if cloudkillDemo {
			spellID = game.CloudkillSpellID
		}
		if err := state.SetPartyRoster(party.Roster{
			{ID: hero.ID, Name: hero.Name, Race: party.RaceHuman, Class: party.ClassMagicUser,
				Abilities: abilities, Level: 5, HitPoints: 30, MaxHitPoints: 30,
				SpellSlots: []uint8{spellID}, SavingThrows: hero.SavingThrows},
			{ID: ally.ID, Name: ally.Name, Race: party.RaceHuman, Class: party.ClassFighter,
				Abilities: abilities, Level: 5, HitPoints: 100, MaxHitPoints: 100,
				SavingThrows: ally.SavingThrows},
		}); err != nil {
			return 0, err
		}
	}
	if err := state.StartCombat(heroes, enemies, 37); err != nil {
		return 0, err
	}
	if fireballDemo {
		if err := state.BeginCombatCast(game.FireballSpellID); err != nil {
			return 0, err
		}
		if err := state.CombatCast(game.FireballSpellID); err != nil {
			return 0, err
		}
	} else if lightningDemo {
		if err := state.BeginCombatCast(game.LightningBoltSpellID); err != nil {
			return 0, err
		}
		terrain := func(x, y int) combat.LineCell {
			return combat.LineCell{
				Valid:   x >= 0 && x < 8 && y >= 0 && y < 7,
				Reflect: x == 6,
			}
		}
		if err := state.CombatCastWithTerrain(game.LightningBoltSpellID, terrain); err != nil {
			return 0, err
		}
	} else if stinkingCloudDemo {
		if err := state.BeginCombatCast(game.StinkingCloudSpellID); err != nil {
			return 0, err
		}
		terrain := func(x, y int) combat.LineCell {
			return combat.LineCell{Valid: x >= 0 && x < 8 && y >= 0 && y < 7}
		}
		if err := state.CombatCastWithTerrain(game.StinkingCloudSpellID, terrain); err != nil {
			return 0, err
		}
	} else if cloudkillDemo {
		if err := state.BeginCombatCast(game.CloudkillSpellID); err != nil {
			return 0, err
		}
		terrain := func(x, y int) combat.LineCell {
			return combat.LineCell{Valid: x >= 0 && x < 8 && y >= 0 && y < 7}
		}
		if err := state.CombatCastWithTerrain(game.CloudkillSpellID, terrain); err != nil {
			return 0, err
		}
	} else if kind != "magic" && kind != "magic-impact" {
		if err := state.CombatAct(); err != nil {
			return 0, err
		}
	}
	event, ok := state.CombatVisualEvent()
	if !ok {
		return 0, fmt.Errorf("combat visual demo %q did not queue an event", kind)
	}
	switch kind {
	case "melee":
		return combat.VisualWindupDuration / 2, nil
	case "bow", "magic":
		return combat.VisualWindupDuration + combat.VisualTravelDuration/2, nil
	case "magic-impact":
		return combat.VisualWindupDuration + combat.VisualTravelDuration +
			3*combat.VisualImpactDuration/4, nil
	case "fireball-travel":
		return combat.VisualWindupDuration + combat.VisualTravelDuration/2, nil
	case "stinking-cloud-travel":
		return combat.VisualWindupDuration + combat.VisualTravelDuration/2, nil
	case "stinking-cloud-persistent":
		return combat.VisualWindupDuration + combat.VisualTravelDuration +
			combat.VisualImpactDuration/2, nil
	case "cloudkill-travel":
		return combat.VisualWindupDuration + combat.VisualTravelDuration/2, nil
	case "cloudkill-persistent":
		return combat.VisualWindupDuration + combat.VisualTravelDuration +
			combat.VisualImpactDuration/2, nil
	case "fireball-impact-1":
		return combat.VisualWindupDuration + combat.VisualTravelDuration +
			3*combat.VisualImpactDuration/4, nil
	case "fireball-impact-2":
		return combat.VisualWindupDuration + combat.VisualTravelDuration +
			combat.VisualImpactDuration + combat.VisualCommitDuration +
			3*combat.VisualImpactDuration/4, nil
	case "lightning-target-hit":
		return combat.VisualWindupDuration + combat.VisualTravelDuration +
			3*combat.VisualImpactDuration/4, nil
	case "lightning-line-continue":
		return combat.VisualWindupDuration + combat.VisualTravelDuration +
			combat.VisualImpactDuration + combat.VisualCommitDuration +
			combat.VisualTravelDuration/2, nil
	case "lightning-reflect":
		return combat.VisualWindupDuration + combat.VisualTravelDuration +
			combat.VisualImpactDuration + combat.VisualCommitDuration +
			combat.VisualTravelDuration +
			combat.VisualImpactDuration + combat.VisualCommitDuration +
			combat.VisualTravelDuration +
			combat.VisualTravelDuration/2, nil
	case "kill":
		return combat.VisualWindupDuration + combat.VisualTravelDuration +
			combat.VisualImpactDuration + combat.VisualCommitDuration +
			2*combat.DeathOverlayPhaseDuration, nil
	}
	return event.Duration() / 2, nil
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
	etenFontPath := flag.String("eten-font", "", "ETen STDFONT.15 path; uses bold 16x15 Chinese glyphs")
	etenSymbolPath := flag.String("eten-symbol-font", "", "optional ETen SPCFONT.15 path for full-width punctuation")
	localePath := flag.String("locale", "assets/locale/zh-TW.json", "locale JSON path")
	imagePath := flag.String("image", "curseoftheazurebonds.zip", "original DOS image ZIP")
	geoSet := flag.Int("geo-set", 2, "GEO DAX set/chapter (2..6) used by the map preview")
	geoBlock := flag.Int("geo-block", 1, "original GEO block ID used by the map preview")
	dungeonXOverride := flag.Int("dungeon-x", -1, "override dungeon X (0..15) for deterministic visual verification")
	dungeonYOverride := flag.Int("dungeon-y", -1, "override dungeon Y (0..15) for deterministic visual verification")
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
	burialRedWeb := flag.Bool("burial-red-web", false, "show the Burial Glen red-web INPUT STRING checkpoint")
	burialRedWebBattle := flag.Bool("burial-red-web-battle", false, "show the first Burial Glen red-web spider battle")
	worldMapPreview := flag.Bool("world-map", false, "show the original BIGPIC overland map for deterministic visual verification")
	areaMapPreview := flag.Bool("area-map", false, "show the GEO overhead AREA map for deterministic visual verification")
	encounterBlock := flag.Int("encounter-block", 81, "ECL block for -encounter")
	encounterStart := flag.Int("encounter-start", 0x1293, "payload offset for -encounter")
	encounterMonsterMember := flag.String("encounter-monster-member", "MON1CHA.DAX", "MON*CHA member for -encounter")
	encounterArea := flag.Int("encounter-area", 1, "original graphics area used by -encounter sprites (1..6)")
	combatTerrainMode := flag.String("combat-terrain", "", "override combat terrain atlas for visual verification: DUNGCOM, WILDCOM, or RANDCOM")
	combatVisualDemo := flag.String("combat-visual-demo", "", "deterministic visual oracle: melee, bow, magic, fireball, lightning, stinking-cloud, cloudkill, or kill checkpoints")
	partyPath := flag.String("party-save", "party.json", "versioned remake party save path")
	soundDir := flag.String("sound-dir", "assets/audio", "reference WAV asset directory; missing assets disable sound")
	pc98MusicDriverPath := flag.String("pc98-music-driver", "", "local extracted MSCDRV.EXE used to synthesize the PC-98 soundtrack")
	pc98SFXGamePath := flag.String("pc98-sfx-game", "", "local exact PC-98 GAME.EXE used to reconstruct software-speaker effects")
	pc98SFXClock := flag.Uint64("pc98-sfx-clock", 8_000_000, "PC-98 CPU clock for reconstructed software-speaker timing")
	partyLoadPath := flag.String("party-load", "", "load a versioned remake party save before starting")
	savgamDir := flag.String("savgam-dir", "", "directory containing reference savgam?.dat and CHRDAT player bundles")
	savgamSlot := flag.String("savgam-slot", "", "reference SAVGAM slot key A..J to load and save with -savgam-dir")
	dosCharacterID := flag.String("dos-character-id", "dos-character", "ID for a direct DOS character import")
	dosCharacterRecord := flag.String("dos-character-record", "", "DOS .SAV/.GUY path to load directly into the remake")
	dosCharacterEffects := flag.String("dos-character-effects", "", "optional DOS .FX path for direct character import")
	dosCharacterInventory := flag.String("dos-character-inventory", "", "optional DOS .SWG path for direct character import")
	screenshotPath := flag.String("screenshot", "", "write one deterministic 640x480 frame to PNG and exit")
	flag.Parse()
	*combatTerrainMode = strings.ToUpper(*combatTerrainMode)
	if *combatTerrainMode != "" && *combatTerrainMode != "DUNGCOM" && *combatTerrainMode != "WILDCOM" && *combatTerrainMode != "RANDCOM" {
		log.Fatal("-combat-terrain must be DUNGCOM, WILDCOM, RANDCOM, or empty for automatic selection")
	}
	*combatVisualDemo = strings.ToLower(strings.TrimSpace(*combatVisualDemo))
	if *combatVisualDemo != "" && *combatVisualDemo != "melee" && *combatVisualDemo != "bow" &&
		*combatVisualDemo != "magic" && *combatVisualDemo != "magic-impact" &&
		*combatVisualDemo != "fireball-travel" && *combatVisualDemo != "fireball-impact-1" &&
		*combatVisualDemo != "fireball-impact-2" &&
		*combatVisualDemo != "lightning-target-hit" &&
		*combatVisualDemo != "lightning-line-continue" &&
		*combatVisualDemo != "lightning-reflect" &&
		*combatVisualDemo != "stinking-cloud-travel" &&
		*combatVisualDemo != "stinking-cloud-persistent" &&
		*combatVisualDemo != "cloudkill-travel" &&
		*combatVisualDemo != "cloudkill-persistent" &&
		*combatVisualDemo != "kill" {
		log.Fatal("-combat-visual-demo has an unknown value")
	}
	if (*dungeonXOverride == -1) != (*dungeonYOverride == -1) || *dungeonXOverride < -1 || *dungeonXOverride >= geo.Width || *dungeonYOverride < -1 || *dungeonYOverride >= geo.Height {
		log.Fatal("-dungeon-x and -dungeon-y must both be omitted or both be 0..15")
	}
	if *burialRedWeb || *burialRedWebBattle {
		*geoSet = 6
		*geoBlock = 0x40
	}
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
	if *worldMapPreview {
		state.Mode = game.ModeWilderness
		state.Area.CurrentCity = 4
		state.Location = game.LocationStandingStone
		state.LocationName = catalog.Text("standing_stone", "立石")
		state.Choices = []string{
			catalog.Text("enter_city", "進入城市"),
			catalog.Text("journey_on", "繼續旅程"),
			catalog.Text("camp", "紮營"),
		}
		state.Prompt = catalog.Text("press_button", "請選擇行動")
	}
	if *dungeonXOverride >= 0 {
		state.DungeonX, state.DungeonY = *dungeonXOverride, *dungeonYOverride
	}
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
		log.Printf("some sound effects are unavailable: %v", soundErr)
	}
	if *pc98SFXGamePath != "" {
		pc98SFXGame, readErr := os.ReadFile(*pc98SFXGamePath)
		if readErr != nil {
			log.Fatal(readErr)
		}
		if loadErr := soundPlayer.LoadPC98Effects(pc98SFXGame, *pc98SFXClock); loadErr != nil {
			log.Fatal(loadErr)
		}
		log.Printf(
			"PC-98 software-speaker effects enabled with reconstructed V30 timing at %d Hz",
			*pc98SFXClock,
		)
	}
	var pc98MusicDriver []byte
	if *pc98MusicDriverPath != "" {
		pc98MusicDriver, err = os.ReadFile(*pc98MusicDriverPath)
		if err != nil {
			log.Fatal(err)
		}
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
	combatTerrain, err := loadCombatTerrainImages(*imagePath)
	if err != nil {
		log.Fatal(err)
	}
	pack, err := gamepack.Default()
	if err != nil {
		log.Fatal(err)
	}
	areaMapDefinition, ok := pack.FindMapByKind("area")
	if !ok {
		log.Fatal("game pack has no AREA map definition")
	}
	areaMapSymbols, err := loadAreaMapSymbols(*imagePath, areaMapDefinition.SymbolFile, areaMapDefinition.SymbolBlock)
	if err != nil {
		log.Fatal(err)
	}
	firstPersonDefinition, exactFirstPerson := pack.FindMapByKindLocation(
		"first_person", uint8(*geoSet), uint8(*geoBlock),
	)
	ok = exactFirstPerson
	if !ok {
		firstPersonDefinition, ok = pack.FindMapByKind("first_person")
	}
	if !ok {
		log.Fatal("game pack has no first-person map definition")
	}
	if exactFirstPerson {
		if firstPersonDefinition.OutdoorSkyColor != nil {
			state.Area.OutdoorSkyColor = uint16(*firstPersonDefinition.OutdoorSkyColor)
		}
		if firstPersonDefinition.IndoorSkyColor != nil {
			state.Area.IndoorSkyColor = uint16(*firstPersonDefinition.IndoorSkyColor)
		}
	}
	skyImages, err := loadSkyImages(*imagePath, firstPersonDefinition.SkyFile, firstPersonDefinition.SkyBlocks)
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
		if *combatTerrainMode == "WILDCOM" {
			state.SetInDungeon(false)
			state.MapX, state.MapY = 25, 12
			state.WildernessFloor = mapdata.GenerateWilderness(0x20, 1)
		} else if *combatTerrainMode == "DUNGCOM" {
			state.SetInDungeon(true)
		}
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
	} else if *burialRedWeb || *burialRedWebBattle {
		if err := state.OpenCharacterCreation(); err != nil {
			log.Fatal(err)
		}
		if err := state.AddCreationCharacter(0); err != nil {
			log.Fatal(err)
		}
		if err := state.FinishCharacterCreation(); err != nil {
			log.Fatal(err)
		}
		if err := state.StartDungeonStoryPreview(0x40, 0x50, 6); err != nil {
			log.Fatal(err)
		}
		if err := state.Continue(); err != nil {
			log.Fatal(err)
		}
		state.SetDungeonGeometryView(6, 14, 2)
		state.DungeonWallRoof = 0x82
		if err := state.RunDungeonLifecycle(); err != nil {
			log.Fatal(err)
		}
		if *burialRedWebBattle {
			if err := state.Select(0); err != nil {
				log.Fatal(err)
			}
		} else {
			if err := state.Select(1); err != nil {
				log.Fatal(err)
			}
			if err := state.AppendECLString([]rune("Krrkik")); err != nil {
				log.Fatal(err)
			}
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
	regularFace := loadFace(*fontPath, 24)
	compactFace := loadFace(*fontPath, 16)
	if *etenFontPath != "" {
		etenFace, err := etenfont.Load(*etenFontPath, *etenSymbolPath, compactFace, true)
		if err != nil {
			log.Fatal(err)
		}
		regularFace = etenFace
		compactFace = etenFace
	}
	state.EnableCombatVisualTimeline(true)
	visualSerial := uint64(0)
	visualStarted := time.Time{}
	if *combatVisualDemo != "" {
		offset, err := prepareCombatVisualDemo(&state, *combatVisualDemo)
		if err != nil {
			log.Fatal(err)
		}
		event, _ := state.CombatVisualEvent()
		visualSerial = event.Serial
		visualStarted = time.Now().Add(-offset)
	}
	if err := ebiten.RunGame(&app{state: state, imagePath: *imagePath, face: regularFace, compactFace: compactFace, partyPath: *partyPath, savgamDir: *savgamDir, savgamSlot: loadedSAVGAMSlot, savgamSlotSave: loadedSAVGAMSlot != 0, soundPlayer: soundPlayer, pc98MusicDriver: pc98MusicDriver, tileImages: tileImages, areaMapSymbols: areaMapSymbols, skyImages: skyImages, geoGrid: geoGrid, areaMapPreview: *areaMapPreview, dungeonFloor: dungeonFloor, dungeonX: dungeonX, dungeonY: dungeonY, geoLabel: geoLabel, geoCatalog: geoCatalog, geoSet: geoRef.Set, geoBlock: geoRef.BlockID, pieceSets: make(map[uint8]gfx.PieceSet), combatSprites: combatSprites, combatSpriteIDs: combatSpriteIDs, combatTerrain: combatTerrain, combatTerrainMode: *combatTerrainMode, gamePack: pack, combatFrame: ebiten.NewImageFromImage(gfx.CombatFrame()), adventureFrame: ebiten.NewImageFromImage(gfx.AdventureFrame()), combatAnimations: combatAnimations, animationStart: time.Now(), combatVisualSerial: visualSerial, combatVisualStarted: visualStarted, combatVisualElapsed: time.Since(visualStarted), screenshotPath: *screenshotPath}); err != nil {
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
		if strings.HasPrefix(name, "cpic") || strings.HasPrefix(name, "comspr") ||
			strings.HasPrefix(name, "party") || strings.HasPrefix(name, "chead") ||
			strings.HasPrefix(name, "cbody") {
			decoded = chromaKeyTopLeft(decoded)
		}
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
		decoded = chromaKeyTopLeft(decoded)
		frameMarker := strings.Index(record.Name, "-frame-")
		if frameMarker < 0 {
			return nil, nil, nil, fmt.Errorf("animation asset %q has no frame marker", record.Name)
		}
		key := record.Name[:frameMarker]
		animations[key] = append(animations[key], combatAnimation{image: ebiten.NewImageFromImage(decoded), delay: record.Delay, x: record.X, y: record.Y})
	}
	return images, ids, animations, nil
}

// chromaKeyTopLeft applies the indexed-picture transparent color used by
// combat sprites. Derived PNGs preserve the EGA RGB value but not the original
// masked blit operation, so loading them as opaque rectangles is incorrect.
func chromaKeyTopLeft(source image.Image) image.Image {
	bounds := source.Bounds()
	output := image.NewNRGBA(bounds)
	key := color.NRGBAModel.Convert(source.At(bounds.Min.X, bounds.Min.Y)).(color.NRGBA)
	if key.A == 0 {
		return source
	}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := color.NRGBAModel.Convert(source.At(x, y)).(color.NRGBA)
			if pixel.R == key.R && pixel.G == key.G && pixel.B == key.B {
				pixel.A = 0
			}
			output.SetNRGBA(x, y, pixel)
		}
	}
	return output
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

func loadCombatTerrainImages(imagePath string) (map[string][]*ebiten.Image, error) {
	result := make(map[string][]*ebiten.Image)
	for _, source := range []string{"DUNGCOM", "WILDCOM", "RANDCOM"} {
		data, err := zipMember(imagePath, source+".DAX")
		if err != nil {
			return nil, err
		}
		blocks, err := dax.Parse(data)
		if err != nil || len(blocks) != 1 {
			return nil, fmt.Errorf("parse %s.DAX: blocks=%d err=%v", source, len(blocks), err)
		}
		set, err := gfx.ParseCombatTiles(blocks[0].Data)
		if err != nil {
			return nil, fmt.Errorf("parse %s.DAX combat tiles: %w", source, err)
		}
		images := make([]*ebiten.Image, 0, len(set.Tiles))
		for _, tile := range set.Tiles {
			rgba, err := tile.RGBA(0, gfx.EGA16)
			if err != nil {
				return nil, err
			}
			images = append(images, ebiten.NewImageFromImage(rgba))
		}
		result[source] = images
	}
	return result, nil
}

func loadAreaMapSymbols(imagePath, symbolFile string, blockID uint8) ([]*ebiten.Image, error) {
	if symbolFile == "" {
		return nil, fmt.Errorf("AREA symbol file is empty")
	}
	data, err := zipMember(imagePath, symbolFile)
	if err != nil {
		return nil, err
	}
	blocks, err := dax.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", symbolFile, err)
	}
	for _, block := range blocks {
		if block.Entry.ID != blockID {
			continue
		}
		picture, err := gfx.ParsePicture(block.Data, true, 13)
		if err != nil {
			return nil, fmt.Errorf("%s block 0x%02X: %w", symbolFile, blockID, err)
		}
		if picture.Width() != 8 || picture.Height() != 8 || picture.ItemCount < 20 {
			return nil, fmt.Errorf("%s block 0x%02X is %dx%d items=%d, want 8x8 and at least 20 items", symbolFile, blockID, picture.Width(), picture.Height(), picture.ItemCount)
		}
		images := make([]*ebiten.Image, 0, picture.ItemCount)
		for item := 0; item < int(picture.ItemCount); item++ {
			rgba, err := picture.RGBA(item, gfx.EGA16)
			if err != nil {
				return nil, err
			}
			images = append(images, ebiten.NewImageFromImage(rgba))
		}
		return images, nil
	}
	return nil, fmt.Errorf("%s has no AREA symbol block 0x%02X", symbolFile, blockID)
}

func loadSkyImages(imagePath, skyFile string, blockIDs [3]uint8) ([3]*ebiten.Image, error) {
	var result [3]*ebiten.Image
	if skyFile == "" {
		return result, fmt.Errorf("first-person SKY file is empty")
	}
	data, err := zipMember(imagePath, skyFile)
	if err != nil {
		return result, err
	}
	blocks, err := dax.Parse(data)
	if err != nil {
		return result, fmt.Errorf("parse %s: %w", skyFile, err)
	}
	byID := make(map[uint8]dax.Block, len(blocks))
	for _, block := range blocks {
		byID[block.Entry.ID] = block
	}
	for index, blockID := range blockIDs {
		block, ok := byID[blockID]
		if !ok {
			return result, fmt.Errorf("%s has no SKY block 0x%02X", skyFile, blockID)
		}
		picture, err := gfx.ParsePicture(block.Data, true, 13)
		if err != nil {
			return result, fmt.Errorf("%s block 0x%02X: %w", skyFile, blockID, err)
		}
		rgba, err := picture.RGBA(0, gfx.EGA16)
		if err != nil {
			return result, err
		}
		result[index] = ebiten.NewImageFromImage(rgba)
	}
	return result, nil
}

// loadMapPieceSets mirrors the verified reference LoadWalldef mapping while
// keeping selector interpretation out of the ECL VM and State packages.
func loadMapPieceSets(imagePath string, areaID uint8, wallFile, symbolFile string, selectors [3]uint16) (map[uint8]gfx.PieceSet, error) {
	if areaID < 1 || areaID > 6 {
		return nil, fmt.Errorf("map piece area %d is outside original range 1..6", areaID)
	}
	if wallFile == "" {
		wallFile = fmt.Sprintf("WALLDEF%d.DAX", areaID)
	}
	if symbolFile == "" {
		symbolFile = fmt.Sprintf("8X8D%d.DAX", areaID)
	}
	wallData, err := zipMember(imagePath, wallFile)
	if err != nil {
		return nil, err
	}
	symbolData, err := zipMember(imagePath, symbolFile)
	if err != nil {
		return nil, err
	}
	wallBlocks, err := dax.Parse(wallData)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", wallFile, err)
	}
	symbolBlocks, err := dax.Parse(symbolData)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", symbolFile, err)
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
