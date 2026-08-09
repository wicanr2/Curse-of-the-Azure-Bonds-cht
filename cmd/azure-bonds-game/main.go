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
	engineaction "github.com/wicanr2/golden-box-remake-engine/combat/action"
	enginescan "github.com/wicanr2/golden-box-remake-engine/combat/scan"
	goldenengine "github.com/wicanr2/golden-box-remake-engine/engine"
)

const (
	logicalWidth  = 640
	logicalHeight = 480
)

type app struct {
	state                 game.State
	imagePath             string
	face                  font.Face
	compactFace           font.Face
	choiceCursor          int
	partyPath             string
	savgamDir             string
	savgamSlot            byte
	savgamSlotSave        bool
	tilePreview           bool
	tileImages            []*ebiten.Image
	geoPreview            bool
	areaMapPreview        bool
	areaMapSymbols        []*ebiten.Image
	skyImages             [3]*ebiten.Image
	geoGrid               *geo.Grid
	geoX                  int
	geoY                  int
	geoLabel              string
	geoCatalog            geo.Catalog
	geoSet                uint8
	geoBlock              uint8
	dungeonPreview        bool
	dungeonFloor          *mapdata.DungeonFloor
	dungeonX              int
	dungeonY              int
	dungeonDoorMenu       bool
	pieceSets             map[uint8]gfx.PieceSet
	pieceLabel            string
	wallPreview           []wallPreviewStamp
	combatSprites         map[string]*ebiten.Image
	combatSpriteIDs       []string
	combatTerrain         map[string][]*ebiten.Image
	combatTerrainMode     string
	combatPreviewFocus    uint8
	gamePack              *goldenengine.Pack
	combatFrame           *ebiten.Image
	adventureFrame        *ebiten.Image
	characterStageFrame   *ebiten.Image
	firstPersonStageFrame *ebiten.Image
	combatAnimations      map[string][]combatAnimation
	animationStart        time.Time
	deathOverlayStarted   map[string]time.Time
	combatVisualSerial    uint64
	combatVisualStarted   time.Time
	combatVisualBase      time.Duration
	combatVisualElapsed   time.Duration
	combatDoneMenu        bool
	combatSpeedMenu       bool
	messageSnapshot       string
	messageStart          time.Time
	soundPlayer           *sound.Player
	pc98MusicDriver       []byte
	currentMusicTrack     string
	screenshotPath        string
	screenshotDone        bool
	screenshotFrames      int
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

// combatScanTacticalMap projects the title-owned floor buffer and recovered
// TDEF table into the reusable engine's TACTICALMAP contract. Floor bytes are
// already one-based TDEF IDs; they must not be translated through TileIndex,
// which belongs to the separate renderer-atlas namespace.
func (a *app) combatScanTacticalMap() (enginescan.TacticalMap, error) {
	mode := selectCombatTerrainName(a.state.Area.InDungeon, a.combatTerrainMode)
	referenceCoordinates := a.state.CombatUsesReferenceCoordinates()
	width, height := mapdata.WildernessWidth, mapdata.WildernessHeight
	if mode == "DUNGCOM" && !referenceCoordinates {
		// Existing reconstructed placements use local coordinates whose (0,0)
		// corresponds to the reference floor's (18,7). Keep the proven 32x16
		// combat namespace; the remaining floor-buffer rows are not evidence
		// that local combat placement may enter them.
		width = mapdata.WildernessWidth - 18
		height = 16
	}
	if mode != "DUNGCOM" && mode != "WILDCOM" {
		return enginescan.TacticalMap{}, fmt.Errorf("combat mode %q has no recovered TACTICALMAP floor", mode)
	}
	if len(mapdata.BackgroundTiles) < 66 {
		return enginescan.TacticalMap{}, fmt.Errorf("TDEF projection requires 65 records, found %d", len(mapdata.BackgroundTiles)-1)
	}
	definitions := make([]enginescan.TerrainDefinition, 65)
	for index, tile := range mapdata.BackgroundTiles[1:66] {
		definitions[index] = enginescan.TerrainDefinition{
			HT: tile.MoveCost, LOS: tile.Height, SYM: tile.Field, Raw3: tile.TileIndex,
		}
	}
	tiles := make([]uint8, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var tileID uint8
			var ok bool
			switch mode {
			case "DUNGCOM":
				if a.dungeonFloor != nil {
					sourceX, sourceY := x, y
					if !referenceCoordinates {
						sourceX, sourceY = 18+x, 7+y
					}
					tileID, ok = a.dungeonFloor.TileID(sourceX, sourceY)
				}
			case "WILDCOM":
				sourceX, sourceY := x, y
				if !referenceCoordinates {
					sourceX, sourceY = a.state.MapX+x-3, a.state.MapY+y-3
				}
				tileID, ok = a.state.WildernessFloor.TileID(sourceX, sourceY)
			}
			if ok {
				if tileID == 0 || int(tileID) > len(definitions) {
					return enginescan.TacticalMap{}, fmt.Errorf("TACTICALMAP (%d,%d) has invalid one-based TD %d", x, y, tileID)
				}
				tiles[y*width+x] = tileID
			}
		}
	}
	return enginescan.TacticalMap{
		Width: width, Height: height, Tiles: tiles, Definitions: definitions,
	}, nil
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
	if a.soundPlayer != nil {
		snapshot, err := a.soundPlayer.SnapshotPC98Music()
		if err != nil {
			return fmt.Errorf("snapshot PC-98 music: %w", err)
		}
		if err := a.state.SetMusicPlaybackSnapshot(snapshot); err != nil {
			return err
		}
		oneShots, err := a.soundPlayer.SnapshotOneShots()
		if err != nil {
			return fmt.Errorf("snapshot one-shot audio: %w", err)
		}
		if err := a.state.SetOneShotPlaybackSnapshot(&oneShots); err != nil {
			return err
		}
	} else if err := a.state.SetOneShotPlaybackSnapshot(nil); err != nil {
		return err
	}
	return a.state.SavePartyFile(a.partyPath)
}

func (a *app) restoreAudioSnapshot() error {
	if a.soundPlayer != nil {
		if snapshot, ok := a.state.OneShotPlaybackSnapshot(); ok {
			if err := a.soundPlayer.RestoreOneShots(*snapshot); err != nil {
				return fmt.Errorf("restore one-shot audio: %w", err)
			}
		} else {
			// Saves before v9 cannot identify an audible one-shot position. Stop
			// pre-load effects instead of leaking them into the restored world.
			a.soundPlayer.StopOneShots()
		}
	}
	return a.restoreMusicSnapshot()
}

func (a *app) restoreMusicSnapshot() error {
	trackID, snapshot, ok := a.state.MusicPlaybackSnapshot()
	if !ok {
		a.currentMusicTrack = ""
		if a.soundPlayer != nil {
			a.soundPlayer.StopMusic()
		}
		return nil
	}
	a.currentMusicTrack = trackID
	if len(a.pc98MusicDriver) == 0 || a.soundPlayer == nil {
		return nil
	}
	track, found := a.gamePack.FindMusicTrack(trackID)
	if !found {
		return fmt.Errorf("music track %q is not in the game pack", trackID)
	}
	if snapshot == nil {
		return a.soundPlayer.PlayPC98Track(a.pc98MusicDriver, int(track.ReferenceSelector))
	}
	if snapshot.Selector != int(track.ReferenceSelector) {
		return fmt.Errorf("music track %q selector %d does not match snapshot selector %d", trackID, track.ReferenceSelector, snapshot.Selector)
	}
	return a.soundPlayer.RestorePC98Track(a.pc98MusicDriver, *snapshot)
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
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			a.state.CombatManualControl()
		}
		if a.combatVisualSerial != event.Serial {
			a.combatVisualSerial = event.Serial
			a.combatVisualStarted = time.Now()
			a.combatVisualBase = a.state.CombatVisualElapsed()
		}
		if a.screenshotPath != "" {
			return nil
		}
		elapsed := combatVisualResumeElapsed(a.combatVisualBase, time.Since(a.combatVisualStarted), a.state.CombatSpeed())
		if err := a.state.AdvanceCombatVisual(elapsed); err != nil {
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
			a.state.Message = a.state.FileOperationMessage(game.FileOperationSave, game.FileOperationFailed, err.Error())
		} else {
			a.state.Message = a.state.FileOperationMessage(game.FileOperationSave, game.FileOperationSucceeded, a.saveTarget())
		}
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF9) {
		if err := a.state.LoadPartyFile(a.partyPath); err != nil {
			a.state.Message = a.state.FileOperationMessage(game.FileOperationLoad, game.FileOperationFailed, err.Error())
		} else if err := a.restoreAudioSnapshot(); err != nil {
			a.state.Message = a.state.FileOperationMessage(game.FileOperationAudioRestore, game.FileOperationFailed, err.Error())
		} else {
			a.state.Message = a.state.FileOperationMessage(game.FileOperationLoad, game.FileOperationSucceeded, a.partyPath)
			a.choiceCursor = 0
		}
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		(inpututil.IsKeyJustPressed(ebiten.KeySpace) && a.state.Mode != game.ModeCombat) {
		switch a.state.Mode {
		case game.ModeTitle:
			return a.state.Apply(game.ActionStart)
		case game.ModeWilderness:
			err := a.state.Select(a.choiceCursor)
			if a.state.ConsumeSaveRequest() {
				if saveErr := a.saveCurrentGame(); saveErr != nil {
					a.state.Message = a.state.FileOperationMessage(game.FileOperationSave, game.FileOperationFailed, saveErr.Error())
				} else {
					a.state.Message = a.state.FileOperationMessage(game.FileOperationSave, game.FileOperationSucceeded, a.saveTarget())
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
			if a.combatDoneMenu {
				return nil
			}
			if a.state.CombatViewActive() {
				a.state.EndCombatView()
				return nil
			}
			if a.state.CombatCastingSpell() != 0 {
				return a.combatAction(func() error { return a.state.ConfirmCombatCast(a.combatLineTerrain()) })
			}
			return a.combatAction(a.state.CombatAct)
		}
	}
	if a.state.Mode == game.ModeCombat {
		if a.combatSpeedMenu {
			if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyE) {
				a.combatSpeedMenu = false
				return nil
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyS) {
				a.state.CombatSpeedSlower()
				return nil
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyF) {
				a.state.CombatSpeedFaster()
				return nil
			}
			return nil
		}
		if a.combatDoneMenu {
			if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyE) {
				a.combatDoneMenu = false
				return nil
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyD) {
				a.combatDoneMenu = false
				return a.combatAction(a.state.CombatDelay)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
				a.combatDoneMenu = false
				return a.combatAction(a.state.CombatDone)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyG) && a.state.CombatCanGuard() {
				a.combatDoneMenu = false
				return a.combatAction(a.state.CombatGuard)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyB) && a.state.CombatCanBandage() {
				a.combatDoneMenu = false
				return a.combatAction(a.state.CombatBandage)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyS) {
				a.combatSpeedMenu = true
				return nil
			}
			return nil
		}
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
			a.state.CombatCastingSpell() == game.LightningBoltSpellID ||
			a.state.CombatCastingSpell() == game.SleepSpellID {
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
		if inpututil.IsKeyJustPressed(ebiten.KeyZ) && a.state.CombatCanCastSleep() {
			return a.combatAction(func() error { return a.state.BeginCombatCast(game.SleepSpellID) })
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
			if combatAltPressed() {
				return a.combatAction(func() error {
					_, err := a.state.CombatToggleQuickMagic()
					return err
				})
			}
			return a.combatAction(a.state.BeginCombatMove)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
			if combatAltPressed() {
				return a.combatAction(a.state.CombatQuickAll)
			}
			return a.combatAction(a.state.CombatQuick)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			a.state.CombatManualControl()
			return nil
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyD) {
			a.combatDoneMenu = true
			return nil
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

func combatAltPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyAltLeft) || ebiten.IsKeyPressed(ebiten.KeyAltRight)
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
		a.pieceLabel = a.state.PreviewPiecesFailedText(err)
		return
	}
	a.pieceSets = sets
	a.prepareWallPreview()
	a.pieceLabel = a.state.PreviewPiecesLoadedText(selectors[0], selectors[1], selectors[2])
}

func (a *app) syncGeoMapRequest() {
	set, block, ok := a.state.ConsumeGeoMapRequest()
	if !ok {
		return
	}
	grid, found := a.geoCatalog.Lookup(geo.MapRef{Set: set, BlockID: block})
	if !found {
		a.state.Message = a.state.PreviewGeoMapMissingText(set, block)
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
	if a.geoGrid != nil {
		a.state.TurnDungeonWithGrid(*a.geoGrid, delta)
		return
	}
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
	if a.geoGrid == nil {
		return
	}
	geometryX, geometryY, facing := a.state.DungeonGeometryView()
	if a.state.Mode != game.ModeDungeon {
		// Preview maps have their own cursor; mirror it into State so the same
		// movement transaction is exercised without an ECL lifecycle.
		a.state.SetDungeonGeometryView(a.dungeonX, a.dungeonY, facing)
		geometryX, geometryY = a.dungeonX, a.dungeonY
	}
	if !a.geoGrid.CanMoveDungeonWrapped(geometryX, geometryY, direction) {
		if flags, ok := a.dungeonDoorFlags(); ok && (flags == 2 || flags == 3) {
			a.dungeonDoorMenu = true
			a.state.Message = a.state.DungeonMessageText(game.DungeonMessageLockedPrompt)
		}
		return
	}
	if err := a.state.MoveDungeon(*a.geoGrid, dx, dy, direction); err != nil {
		a.state.Message = a.state.DungeonLifecycleErrorText(err)
		return
	}
	a.dungeonX, a.dungeonY, _ = a.state.DungeonGeometryView()
	a.refreshDungeonPreview()
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
		a.state.Message = a.state.DungeonMessageText(game.DungeonMessagePickUnavailable)
		return
	}
	result := a.state.PickDungeonLock()
	_, _, direction := a.state.DungeonGeometryView()
	if result.Opened && a.geoGrid.UnlockDoorWrapped(a.dungeonX, a.dungeonY, int(direction)) {
		a.state.Message = a.state.DungeonMessageText(game.DungeonMessagePickSucceeded)
		a.refreshDungeonPreview()
		return
	}
	a.state.Message = a.state.DungeonMessageText(game.DungeonMessagePickFailed)
}

func (a *app) tryDungeonKnock() {
	flags, ok := a.dungeonDoorFlags()
	if !ok || (flags != 2 && flags != 3) {
		a.state.Message = a.state.DungeonMessageText(game.DungeonMessageKnockSurfaceUnavailable)
		return
	}
	if !a.state.ConsumeDungeonKnockSpell() {
		a.state.Message = a.state.DungeonKnockUnavailableText(dungeon.KnockSpellID)
		return
	}
	_, _, direction := a.state.DungeonGeometryView()
	if a.geoGrid.UnlockDoorWrapped(a.dungeonX, a.dungeonY, int(direction)) {
		a.state.Message = a.state.DungeonMessageText(game.DungeonMessageKnockSucceeded)
		a.refreshDungeonPreview()
	}
}

func (a *app) tryDungeonBash() {
	flags, ok := a.dungeonDoorFlags()
	if !ok || (flags != 2 && flags != 3) {
		a.state.Message = a.state.DungeonMessageText(game.DungeonMessageBashUnavailable)
		return
	}
	result := a.state.BashDungeonDoor(flags)
	_, _, direction := a.state.DungeonGeometryView()
	if result.Opened && a.geoGrid.UnlockDoorWrapped(a.dungeonX, a.dungeonY, int(direction)) {
		a.state.Message = a.state.DungeonMessageText(game.DungeonMessageBashSucceeded)
		a.refreshDungeonPreview()
		return
	}
	a.state.Message = a.state.DungeonMessageText(game.DungeonMessageBashFailed)
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
		text.Draw(screen, a.state.RenameInputText(), a.face, 56, 220, cyan)
		text.Draw(screen, a.state.RenameInputHelp(), a.face, 56, 330, white)
		return
	}
	if a.state.ECLStringEditing() {
		if a.state.Area.InDungeon {
			a.drawDungeonGame(screen, white, cyan)
			ebitenutil.DrawRect(screen, 8, 290, 624, 190, color.RGBA{0, 0, 0, 255})
			drawWrappedText(screen, a.state.Message, a.compactFace, 16, 316, 36, 20, 2, cyan)
			text.Draw(screen, a.state.ECLStringInputText(), a.compactFace, 16, 360, white)
			text.Draw(screen, a.state.ECLStringInputHelp(), a.compactFace, 16, 438, cyan)
			a.drawOriginalAdventureFrame(screen)
			return
		}
		text.Draw(screen, a.state.Title, a.face, 32, 52, cyan)
		text.Draw(screen, a.state.LocationName, a.face, 32, 90, cyan)
		text.Draw(screen, a.state.Prompt, a.face, 32, 130, white)
		drawWrappedText(screen, a.state.Message, a.face, 56, 190, 22, 32, 4, cyan)
		text.Draw(screen, a.state.ECLStringInputText(), a.face, 56, 340, white)
		text.Draw(screen, a.state.ECLStringInputHelp(), a.face, 56, 410, cyan)
		return
	}
	if a.state.Mode == game.ModeCharacterCreation {
		a.drawCreation(screen, white, cyan)
		return
	}
	if a.state.Mode == game.ModeJournal {
		text.Draw(screen, a.state.JournalTitle, a.face, 32, 52, cyan)
		drawWrappedText(screen, a.state.JournalText, a.face, 32, 100, 22, 32, 7, white)
		text.Draw(screen, a.state.JournalPageStatus(), a.face, 32, 350, white)
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
			text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelSelectHelp), a.face, 56, 330, cyan)
			text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelSaveLoadHelp), a.face, 56, 370, white)
		}
	}
	if a.state.Mode == game.ModeEvent {
		drawWrappedText(screen, a.revealedMessage(), a.face, 56, 210, 22, 32, 5, cyan)
		text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelContinueHelp), a.face, 56, 410, white)
	}
	if a.state.Mode == game.ModeMap {
		text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelShadowdaleMapTitle), a.face, 56, 220, cyan)
		text.Draw(screen, a.state.AreaMapPositionText(), a.face, 56, 260, white)
		text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelMapControls), a.face, 56, 330, white)
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
		if localized, found := pack.Text(point.MessageID, a.state.LocaleLanguage()); found {
			currentName = localized
		}
		break
	}

	drawPanelFrame(screen, 8, 264, 624, 184)
	drawPanelFrame(screen, 8, 448, 624, 28)
	text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelOverlandTitle), a.face, 24, 296, cyan)
	text.Draw(screen, a.state.OverlandCurrentLocationText(currentName), a.face, 24, 328, white)
	timeLabel := a.state.OverlandDateText()
	text.Draw(screen, timeLabel, a.compactFace, 468, 294, cyan)
	for index, choice := range a.state.Choices {
		prefix := "  "
		if index == a.choiceCursor {
			prefix = "> "
		}
		text.Draw(screen, prefix+choice, a.face, 40, 366+index*30, white)
	}
	text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelOverlandControls), a.compactFace, 344, 470, cyan)
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
			a.drawSceneCharacter(screen, sprite)
			a.drawOriginalAdventureFrame(screen)
			a.drawPictureMessage(screen)
			text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelContinueHelp), a.compactFace, 24, 468, color.RGBA{255, 255, 255, 255})
			return
		}
		text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelSceneCharacterMissing), a.face, 56, 220, color.RGBA{255, 220, 100, 255})
		text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelContinueHelp), a.face, 56, 330, color.RGBA{255, 255, 255, 255})
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
			text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelBigPictureContinue), a.face, 56, 446, color.RGBA{255, 255, 255, 255})
			return
		}
		text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelBigPictureMissing), a.face, 56, 220, color.RGBA{255, 220, 100, 255})
		text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelContinueHelp), a.face, 56, 330, color.RGBA{255, 255, 255, 255})
		return
	}
	key := fmt.Sprintf("pic%d-block-%02X", a.state.Area.GameArea, a.state.PictureBlock)
	frames := a.combatAnimations[key]
	if len(frames) == 0 {
		text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelPictureMissing), a.face, 56, 220, color.RGBA{255, 220, 100, 255})
		text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelContinueHelp), a.face, 56, 330, color.RGBA{255, 255, 255, 255})
		return
	}
	frame := frames[0]
	// A screenshot checkpoint freezes the source animation on frame zero;
	// container startup time must not choose a different XOR/delta phase.
	if a.state.AnimationsEnabled() && a.screenshotPath == "" {
		frame = animationFrame(frames, time.Since(a.animationStart))
	}
	a.drawAdventureChrome(screen)
	drawImageCover(screen, frame.image, image.Rect(48, 48, 224, 224))
	a.drawFirstPersonStageFrame(screen)
	a.drawOriginalAdventureFrame(screen)
	a.drawPictureMessage(screen)
	text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelContinueHelp), a.compactFace, 24, 468, color.RGBA{255, 255, 255, 255})
}

func (a *app) drawSceneCharacter(screen, sprite *ebiten.Image) {
	if sprite == nil || a.gamePack == nil || a.gamePack.Presentation == nil ||
		a.gamePack.Presentation.SceneCharacter == nil {
		return
	}
	presentation := a.gamePack.Presentation
	layout := presentation.SceneCharacter
	scale := presentation.NativeScale
	clip := image.Rect(
		layout.Clip.X*scale,
		layout.Clip.Y*scale,
		(layout.Clip.X+layout.Clip.Width)*scale,
		(layout.Clip.Y+layout.Clip.Height)*scale,
	)
	target := screen.SubImage(clip).(*ebiten.Image)
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(
		float64(layout.SpriteX*scale-clip.Min.X),
		float64(layout.SpriteY*scale-clip.Min.Y),
	)
	target.DrawImage(sprite, op)
	if a.characterStageFrame != nil {
		frameOptions := &ebiten.DrawImageOptions{}
		frameOptions.Filter = ebiten.FilterNearest
		frameOptions.GeoM.Scale(float64(scale), float64(scale))
		screen.DrawImage(a.characterStageFrame, frameOptions)
	}
}

func (a *app) drawFirstPersonStageFrame(screen *ebiten.Image) {
	if a.firstPersonStageFrame == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Scale(2, 2)
	screen.DrawImage(a.firstPersonStageFrame, op)
}

func drawImageCover(screen, source *ebiten.Image, destination image.Rectangle) {
	if source == nil || destination.Empty() {
		return
	}
	sourceWidth, sourceHeight := source.Bounds().Dx(), source.Bounds().Dy()
	scale, translateX, translateY := imageCoverTransform(sourceWidth, sourceHeight, destination)
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(translateX, translateY)
	screen.DrawImage(source, op)
}

func imageCoverTransform(sourceWidth, sourceHeight int, destination image.Rectangle) (scale, x, y float64) {
	scale = max(
		float64(destination.Dx())/float64(sourceWidth),
		float64(destination.Dy())/float64(sourceHeight),
	)
	x = float64(destination.Min.X) + float64(destination.Dx()-int(float64(sourceWidth)*scale))/2
	y = float64(destination.Min.Y) + float64(destination.Dy()-int(float64(sourceHeight)*scale))/2
	return scale, x, y
}

func (a *app) drawPictureMessage(screen *ebiten.Image) {
	drawWrappedText(
		screen, a.revealedMessage(), a.compactFace,
		24, 282, 36, 24, 6, color.RGBA{92, 220, 255, 255},
	)
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
	text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelCharacterNameHeader), a.compactFace, 280, 38, color.RGBA{232, 238, 255, 255})
	text.Draw(screen, "AC", a.compactFace, 528, 38, color.RGBA{232, 238, 255, 255})
	text.Draw(screen, "HP", a.compactFace, 600, 38, color.RGBA{232, 238, 255, 255})
	for index, fighter := range a.state.PartyFighters() {
		if index >= 8 {
			break
		}
		ink := color.RGBA{R: 92, G: 220, B: 255, A: 255}
		text.Draw(screen, fighter.Name, a.compactFace, 280, 68+index*20, ink)
		text.Draw(screen, strconv.Itoa(fighter.ArmorClass), a.compactFace, 532, 68+index*20, ink)
		text.Draw(screen, strconv.Itoa(fighter.HitPoints), a.compactFace, 604, 68+index*20, ink)
	}
}

func (a *app) drawAreaMap(screen *ebiten.Image, white, cyan color.Color) {
	drawPanelFrame(screen, 8, 8, 352, 352)
	drawPanelFrame(screen, 360, 8, 272, 352)
	drawPanelFrame(screen, 8, 360, 624, 88)
	drawPanelFrame(screen, 8, 448, 624, 28)
	if a.geoGrid == nil {
		text.Draw(screen, a.state.PreviewLabelText(game.PreviewLabelAreaGeoMissing), a.face, 24, 48, white)
		return
	}
	x, y, direction := a.state.DungeonGeometryView()
	if !a.areaMapPreview {
		x, y = a.state.MapX, a.state.MapY
	}
	view, err := enginearea.BuildOriginal(*a.geoGrid, x, y, int(direction))
	if err != nil || len(a.areaMapSymbols) < 20 {
		text.Draw(screen, a.state.PreviewLabelText(game.PreviewLabelAreaSymbolsMissing), a.compactFace, 24, 48, white)
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

	text.Draw(screen, a.state.PreviewLabelText(game.PreviewLabelAreaTitle), a.compactFace, 376, 38, cyan)
	text.Draw(screen, a.state.PreviewAreaSourceText(a.geoSet, a.geoBlock), a.compactFace, 376, 66, white)
	text.Draw(screen, a.state.PreviewAreaPositionText(x, y), a.compactFace, 376, 102, white)
	text.Draw(screen, a.state.PreviewAreaDirectionText(dungeonDirectionName(direction)), a.compactFace, 376, 130, white)
	text.Draw(screen, a.state.PreviewLabelText(game.PreviewLabelAreaOriginalViewport), a.compactFace, 376, 182, cyan)
	text.Draw(screen, a.state.PreviewLabelText(game.PreviewLabelAreaPartyMarker), a.compactFace, 376, 210, color.RGBA{255, 255, 82, 255})
	text.Draw(screen, a.state.LocationName, a.compactFace, 24, 390, cyan)
	text.Draw(screen, a.state.PreviewLabelText(game.PreviewLabelAreaDescription), a.compactFace, 24, 422, white)
	text.Draw(screen, a.state.PreviewLabelText(game.PreviewLabelAreaReturn), a.compactFace, 24, 468, cyan)
}

func (a *app) drawTilePreview(screen *ebiten.Image, white, cyan color.Color) {
	text.Draw(screen, a.state.PreviewLabelText(game.PreviewLabelTileTitle), a.face, 24, 28, cyan)
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
	text.Draw(screen, a.state.PreviewGeoTitleText(a.geoLabel), a.face, 24, 28, cyan)
	if a.geoGrid == nil {
		text.Draw(screen, a.state.PreviewLabelText(game.PreviewLabelGeoMissing), a.face, 24, 70, white)
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
	text.Draw(screen, a.state.PreviewLabelText(game.PreviewLabelGeoCursorHelp), a.face, 24, 382, white)
}

func (a *app) drawDungeonPreview(screen *ebiten.Image, white, cyan color.Color) {
	production := a.state.Mode == game.ModeDungeon
	if production {
		a.drawDungeonGame(screen, white, cyan)
		return
	}
	title := a.state.PreviewLabelText(game.PreviewLabelDungeonTitle)
	text.Draw(screen, title, a.face, 24, 30, cyan)
	if a.dungeonFloor == nil {
		text.Draw(screen, a.state.PreviewLabelText(game.PreviewLabelDungeonFloorMissing), a.face, 24, 70, white)
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
	text.Draw(screen, a.state.PreviewDungeonStatusText(a.dungeonX, a.dungeonY, dungeonDirectionName(geometryDirection), a.geoSet, a.geoBlock), a.face, 24, 242, white)
	text.Draw(screen, a.state.PreviewDungeonWallText(a.state.DungeonWallType, a.state.DungeonWallRoof), a.face, 24, 278, white)
	if a.state.DungeonWallType != 0 && doorFlagsOK {
		text.Draw(screen, a.state.PreviewDungeonDoorStatusText(doorFlags), a.face, 360, 278, white)
	}
	if a.state.Message != "" {
		drawWrappedText(screen, a.state.Message, a.face, 24, 342, 22, 30, 2, white)
	}
	if a.dungeonDoorMenu && doorFlagsOK {
		options := a.state.DungeonDoorMenuOptions(doorFlags)
		text.Draw(screen, a.state.PreviewDungeonDoorHelpText(options.Pick, options.Knock), a.compactFace, 24, 404, color.RGBA{255, 220, 110, 255})
	}
	if a.pieceLabel != "" {
		label := a.pieceLabel
		if production {
			label = a.state.PreviewDungeonPiecesText(a.state.LoadPieces[0], a.state.LoadPieces[1], a.state.LoadPieces[2])
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
	controls := a.state.PreviewLabelText(game.PreviewLabelDungeonControls)
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
	a.drawFirstPersonStageFrame(screen)

	text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelCharacterNameHeader), a.compactFace, 280, 38, white)
	text.Draw(screen, "AC", a.compactFace, 528, 38, white)
	text.Draw(screen, "HP", a.compactFace, 600, 38, white)
	for index, fighter := range a.state.PartyFighters() {
		if index >= 8 {
			break
		}
		text.Draw(screen, fighter.Name, a.compactFace, 280, 68+index*20, cyan)
		text.Draw(screen, strconv.Itoa(fighter.ArmorClass), a.compactFace, 532, 68+index*20, cyan)
		text.Draw(screen, strconv.Itoa(fighter.HitPoints), a.compactFace, 604, 68+index*20, cyan)
	}

	status := fmt.Sprintf("(%d,%d) %s %02d:%02d", a.dungeonX, a.dungeonY,
		dungeonDirectionName(direction), a.state.GameTimeDisplay().Hour, a.state.GameTimeDisplay().Minute)
	text.Draw(screen, status, a.compactFace, 280, 254, cyan)
	if a.state.Message != "" {
		drawWrappedText(screen, a.state.Message, a.compactFace, 24, 282, 36, 24, 6, white)
	} else {
		text.Draw(screen, a.state.LocationName, a.compactFace, 24, 282, cyan)
	}
	if a.dungeonDoorMenu {
		text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelDungeonDoorHelp), a.compactFace, 8, 430, color.RGBA{255, 255, 82, 255})
	}
	text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelDungeonExploreHelp), a.compactFace, 8, 468, cyan)
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
	// Character creation uses the same full-screen adventure chrome as the
	// original DOS menus.  Keep the 16x15 CJK face here: the larger display
	// face is useful for short headings, but makes the original option density
	// impossible to preserve once Traditional Chinese labels are loaded.
	a.drawOriginalAdventureFrame(screen)
	face := a.compactFace
	text.Draw(screen, a.state.LocaleText("creation_title"), face, 32, 52, cyan)
	text.Draw(screen, a.state.CreationMessage, face, 32, 90, white)
	if a.state.CreationEditing {
		text.Draw(screen, fmt.Sprintf(a.state.LocaleText("creation_name_input"), a.state.CreationName), face, 48, 140, white)
		text.Draw(screen, a.state.LocaleText("creation_name_help"), face, 48, 190, cyan)
		return
	}
	if a.state.CreationEditingAbilities {
		character := a.state.CreationOptions[a.state.CreationCursor]
		text.Draw(screen, fmt.Sprintf(a.state.LocaleText("creation_ability_title"), character.Name), face, 32, 135, white)
		abilityKeys := []string{"ability_strength", "ability_intelligence", "ability_wisdom", "ability_dexterity", "ability_constitution", "ability_charisma"}
		for index, key := range abilityKeys {
			value, _ := character.Abilities.Value(index)
			prefix := "  "
			if index == a.state.CreationAbility {
				prefix = "> "
			}
			label := fmt.Sprintf(a.state.LocaleText("creation_ability_row"), a.state.LocaleText(key), value)
			text.Draw(screen, prefix+label, face, 64, 175+index*25, white)
		}
		text.Draw(screen, a.state.LocaleText("creation_ability_help"), face, 48, 350, cyan)
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
		label := prefix + fmt.Sprintf(a.state.LocaleText("creation_option_label"), character.Name,
			a.state.CharacterRaceName(character.Race), a.state.CharacterClassName(character.Class))
		text.Draw(screen, label, face, 48, 150+(index-start)*32, white)
	}
	text.Draw(screen, fmt.Sprintf(a.state.LocaleText("creation_progress"), a.state.CreationCursor+1,
		len(a.state.CreationOptions), len(a.state.CreationRoster)), face, 48, 325, cyan)
	text.Draw(screen, a.state.LocaleText("creation_help"), face, 48, 340, white)
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
		combatMessage = a.state.CombatVisualMessage(event, a.combatVisualFrame(event), combatMessage)
	}
	drawWrappedText(screen, combatMessage, a.compactFace, 8, 392, 39, 20, 3, white)
	if a.state.CombatViewActive() {
		text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelCombatViewTitle), a.face, 64, 145, cyan)
		for index, line := range a.state.CombatViewLines() {
			text.Draw(screen, line, a.face, 64, 185+index*30, white)
		}
		text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelCombatViewReturn), a.face, 64, 350, cyan)
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
		cameraFocus = combatCameraFocus(active, activeOK, minX, maxX, minY, maxY)
	} else if activeOK {
		cameraFocus = combat.TilePoint{X: active.CombatX, Y: active.CombatY}
	}
	if focus, ok := combatPreviewFocus(a.state.CombatFighters(), a.combatPreviewFocus); ok {
		cameraFocus = focus
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
		text.Draw(screen, a.state.CombatHitPointsLabel(), a.compactFace, 370, 62, statusGreen)
		text.Draw(screen, strconv.Itoa(active.HitPoints), a.compactFace, 450, 62, statusYellow)
		text.Draw(screen, a.state.CombatArmorClassLabel(), a.compactFace, 370, 94, statusGreen)
		text.Draw(screen, strconv.Itoa(active.ArmorClass), a.compactFace, 466, 94, statusYellow)
	}
	if prompt, selecting := a.state.CombatSelectionPrompt(); selecting {
		text.Draw(screen, prompt, a.face, 32, 350, cyan)
		return
	}
	spellHints := a.state.CombatQuickSpellHints()
	footerStatus := ""
	if len(targets) > 0 && a.state.CombatTargetIndex() < len(targets) {
		footerStatus = a.state.CombatTargetStatus(targets[a.state.CombatTargetIndex()].Name)
	}
	text.Draw(screen, footerStatus, a.compactFace, 8, 462, color.RGBA{R: 255, G: 82, B: 255, A: 255})
	combatMenu := a.state.CombatMainMenuText()
	if a.combatSpeedMenu {
		combatMenu = a.state.CombatSpeedMenuText()
	} else if a.combatDoneMenu {
		combatMenu = a.state.CombatDoneMenuText()
	}
	text.Draw(screen, combatMenu, a.compactFace, 8, 478, cyan)
	if len(spellHints) > 0 {
		text.Draw(screen, a.state.CombatQuickStatus(spellHints), a.compactFace, 378, 340, cyan)
	}
}

func combatCameraFocus(active combat.Fighter, activeOK bool, minX, maxX, minY, maxY int) combat.TilePoint {
	if activeOK && active.HasCombatPosition {
		return combat.TilePoint{X: active.CombatX, Y: active.CombatY}
	}
	return combat.TilePoint{X: (minX + maxX + 1) / 2, Y: (minY + maxY + 1) / 2}
}

func combatPreviewFocus(fighters []combat.Fighter, spriteBlock uint8) (combat.TilePoint, bool) {
	if spriteBlock == 0 {
		return combat.TilePoint{}, false
	}
	for _, fighter := range fighters {
		if fighter.Side == combat.SideEnemy && fighter.SpriteBlock == spriteBlock && fighter.HasCombatPosition {
			return combat.TilePoint{X: fighter.CombatX, Y: fighter.CombatY}, true
		}
	}
	return combat.TilePoint{}, false
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
		if event.Kind == combat.VisualTwinkle {
			drawCombatTwinkle(screen, toX, toY, frame.Progress)
			break
		}
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

// drawCombatTwinkle reconstructs the PC-98 TWINKLE 24x6 dynamic icon contract.
// Overlay 24 builds four frames at runtime, so there is no DAX block to load.
// The exact source geometry and four-frame/repeat timing are preserved here;
// palette-table pixels remain layout-reconstructed pending a runtime capture.
func drawCombatTwinkle(screen *ebiten.Image, x, y float64, progress float64) {
	frame := min(int(progress*20)%4, 3)
	colors := []color.RGBA{
		{R: 255, G: 255, B: 255, A: 255},
		{R: 96, G: 224, B: 255, A: 255},
		{R: 255, G: 240, B: 96, A: 255},
		{R: 255, G: 255, B: 255, A: 255},
	}
	c := colors[frame]
	halfWidth := float64(6 + frame*2)
	halfHeight := float64(3 + frame)
	ebitenutil.DrawLine(screen, x-halfWidth, y, x+halfWidth, y, c)
	ebitenutil.DrawLine(screen, x, y-halfHeight, x, y+halfHeight, c)
	if frame == 1 || frame == 2 {
		ebitenutil.DrawLine(screen, x-halfHeight, y-halfHeight, x+halfHeight, y+halfHeight, c)
		ebitenutil.DrawLine(screen, x+halfHeight, y-halfHeight, x-halfHeight, y+halfHeight, c)
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
	return event.FrameAt(a.state.CombatVisualElapsed())
}

func combatSpeedElapsed(elapsed time.Duration, speed uint8) time.Duration {
	return engineaction.Speed(speed).ScaleElapsed(elapsed)
}

func combatVisualResumeElapsed(base, clockDelta time.Duration, speed uint8) time.Duration {
	return base + combatSpeedElapsed(clockDelta, speed)
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

func prepareCombatVisualDemo(state *game.State, kind string) (time.Duration, error) {
	fireballDemo := strings.HasPrefix(kind, "fireball-")
	lightningDemo := strings.HasPrefix(kind, "lightning-")
	stinkingCloudDemo := strings.HasPrefix(kind, "stinking-cloud-")
	cloudkillDemo := strings.HasPrefix(kind, "cloudkill-")
	hero := combat.Fighter{
		ID: "demo-hero", Name: state.DemoFighterName(game.DemoNameArcherErin), Side: combat.SideParty,
		HitPoints: 30, MaxHitPoints: 30, ArmorClass: 0,
		AttackBonus: 30, DamageDiceCount: 1, DamageDiceSides: 4,
		InitiativeBonus: 30, HasCombatPosition: true, CombatX: 1, CombatY: 2,
		SavingThrows: []uint8{10, 10, 10, 10, 10},
		HasPartyIcon: true, PartyIconSize: 2,
	}
	enemy := combat.Fighter{
		ID: "demo-orc", Name: state.DemoFighterName(game.DemoNameOrc), Side: combat.SideEnemy,
		HitPoints: 30, MaxHitPoints: 30, ArmorClass: 10,
		AttackBonus: 20, DamageDiceCount: 1, DamageDiceSides: 4,
		InitiativeBonus: -20, HasCombatPosition: true, CombatX: 5, CombatY: 2,
		SavingThrows: []uint8{10, 10, 10, 10, 10},
		SpriteSet:    1, SpriteBlock: 1,
	}
	switch kind {
	case "melee":
		hero.Name = state.DemoFighterName(game.DemoNameFighterErin)
	case "bow":
		hero.MissileWeapon = true
	case "kill":
		hero.Name = state.DemoFighterName(game.DemoNameFighterErin)
		enemy.HitPoints, enemy.MaxHitPoints = 1, 1
	case "magic", "magic-impact":
		hero.InitiativeBonus = -20
		enemy.Name = state.DemoFighterName(game.DemoNameZhentMage)
		enemy.InitiativeBonus = 30
		enemy.MonsterSpellUses[0] = 1
		enemy.MonsterSpellIDs = []uint8{combat.MonsterMagicMissileSpellID}
	case "fireball-travel", "fireball-impact-1", "fireball-impact-2":
		hero.Name = state.DemoFighterName(game.DemoNameMageErin)
	case "lightning-target-hit", "lightning-line-continue", "lightning-reflect":
		hero.Name = state.DemoFighterName(game.DemoNameMageErin)
	case "stinking-cloud-travel", "stinking-cloud-persistent":
		hero.Name = state.DemoFighterName(game.DemoNameMageErin)
	case "cloudkill-travel", "cloudkill-persistent":
		hero.Name = state.DemoFighterName(game.DemoNameMageErin)
		hero.HitDice = 7
		enemy.HitDice = 4
	default:
		return 0, fmt.Errorf("unknown combat visual demo %q", kind)
	}
	heroes := []combat.Fighter{hero}
	enemies := []combat.Fighter{enemy}
	if fireballDemo || lightningDemo || stinkingCloudDemo || cloudkillDemo {
		ally := combat.Fighter{
			ID: "demo-ally", Name: state.DemoFighterName(game.DemoNameFighterBran), Side: combat.SideParty,
			HitPoints: 100, MaxHitPoints: 100, ArmorClass: 2,
			InitiativeBonus: -10, HasCombatPosition: true, CombatX: 2, CombatY: 4,
			SavingThrows: []uint8{10, 10, 10, 10, 10},
			HasPartyIcon: true, PartyHeadBlock: 1, PartyBodyBlock: 1, PartyIconSize: 2,
		}
		enemy.HitPoints, enemy.MaxHitPoints = 100, 100
		enemy.ID = "demo-orc-a"
		enemy.CombatX, enemy.CombatY = 3, 2
		secondEnemy := enemy
		secondEnemy.ID, secondEnemy.Name = "demo-orc-b", state.DemoFighterName(game.DemoNameOrcCaptain)
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
		text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelFallen), a.face, x, y+22, color.RGBA{R: 255, G: 220, B: 180, A: 255})
		return
	}
	frame, active := a.deathOverlayFrame(fighter)
	if !active {
		if fighter.DownedCorpse {
			ebitenutil.DrawRect(screen, float64(x-4), float64(y-4), 36, 44, color.RGBA{R: 58, G: 38, B: 28, A: 220})
			text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelFallen), a.face, x, y+22, color.RGBA{R: 255, G: 220, B: 180, A: 255})
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
	text.Draw(screen, a.state.PlayerUILabel(game.PlayerUILabelFallen), a.face, x, y+22, color.RGBA{R: 255, G: 220, B: 180, A: 255})
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
	tilvertonDungeon := flag.Bool("tilverton-dungeon", false, "enter Tilverton's first-person map through the formal new-game flow")
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
	burialGraveBattle := flag.Bool("burial-grave-battle", false, "show the Burial Glen grave-looter thri-kreen battle")
	burialDaemir := flag.Bool("burial-daemir", false, "show Princess Daemir's Burial Glen blessing choice")
	innerRitual := flag.Bool("inner-ritual", false, "show the Tyranthraxus/Nameless inner-ruins ritual checkpoint")
	innerFinalBattle := flag.Bool("inner-final-battle", false, "show the original 37-enemy Tyranthraxus final-battle boundary")
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
	if *burialRedWeb || *burialRedWebBattle || *burialGraveBattle || *burialDaemir || *innerRitual || *innerFinalBattle {
		*geoSet = 6
		if *innerRitual || *innerFinalBattle {
			*geoBlock = 0x43
		} else {
			*geoBlock = 0x40
		}
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
		state.PrepareWorldMapPreview()
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
	// Deterministic screenshot runs have no audible output and must not depend
	// on a host ALSA device being exposed to the isolated Xvfb container.
	if *screenshotPath != "" {
		soundPlayer = nil
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
		if err := state.StartEncounter(result, monsterRecords, demoParty(&state), 37); err != nil {
			log.Fatal(err)
		}
	} else if *burialRedWeb || *burialRedWebBattle || *burialGraveBattle || *burialDaemir || *innerRitual || *innerFinalBattle {
		if err := state.OpenCharacterCreation(); err != nil {
			log.Fatal(err)
		}
		if err := state.AddCreationCharacter(0); err != nil {
			log.Fatal(err)
		}
		if err := state.FinishCharacterCreation(); err != nil {
			log.Fatal(err)
		}
		previewBlock, previousBlock := uint8(0x40), uint8(0x50)
		if *innerRitual || *innerFinalBattle {
			previewBlock, previousBlock = 0x43, 0x42
		}
		if err := state.StartDungeonStoryPreview(previewBlock, previousBlock, 6); err != nil {
			log.Fatal(err)
		}
		if state.Mode == game.ModeEvent {
			if err := state.Continue(); err != nil {
				log.Fatal(err)
			}
		}
		if *innerFinalBattle {
			if err := prepareInnerFinalBattlePreview(&state, geoGrid); err != nil {
				log.Fatal(err)
			}
		} else if *innerRitual {
			if state.Mode == game.ModeWilderness {
				if err := state.Select(0); err != nil {
					log.Fatal(err)
				}
			}
			if state.Mode != game.ModeDungeon {
				log.Fatalf("-inner-ritual initialization ended in mode %v", state.Mode)
			}
			state.SetECLMemoryValue(0x4C59, 1)
			state.SetECLMemoryValue(0x4C5A, 1)
			state.SetECLMemoryValue(0x4C5B, 0xFF)
			state.SetDungeonGeometryView(7, 11, 4)
			state.DungeonWallRoof = 0x83
			if err := state.RunDungeonLifecycle(); err != nil {
				log.Fatal(err)
			}
			for step := 0; step < 2; step++ {
				if err := state.Select(0); err != nil {
					log.Fatal(err)
				}
			}
			if state.Mode != game.ModeEvent || state.PictureBlock != 0x47 {
				log.Fatalf("-inner-ritual did not reach PICTURE 47: mode=%v picture=%02X",
					state.Mode, state.PictureBlock)
			}
		} else if *burialDaemir {
			state.SetDungeonGeometryView(13, 14, 4)
			state.DungeonWallRoof = 0x03
			if err := state.RunDungeonLifecycle(); err != nil {
				log.Fatal(err)
			}
		} else if *burialGraveBattle {
			state.SetECLSeed(1)
			for attempt := 0; attempt < 8 && !state.CombatActive(); attempt++ {
				for _, y := range []int{13, 12} {
					direction := uint8(4)
					if y == 12 {
						direction = 0
					}
					state.SetDungeonGeometryView(6, y, direction)
					state.DungeonWallRoof = geoGrid.Cells[y][6].Terrain
					if err := state.RunDungeonLifecycle(); err != nil {
						log.Fatal(err)
					}
					if state.Mode == game.ModeWilderness && len(state.Choices) == 4 {
						// A normal ECL random encounter can precede the
						// terrain event; flee and keep walking.
						if err := state.Select(2); err != nil {
							log.Fatal(err)
						}
					} else if state.Mode == game.ModeWilderness && len(state.Choices) == 1 {
						if err := state.Select(0); err != nil {
							log.Fatal(err)
						}
					}
					if state.CombatActive() {
						break
					}
				}
			}
			if !state.CombatActive() {
				log.Fatal("Burial Glen grave battle did not trigger within deterministic preview budget")
			}
		} else {
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
				if index, found := state.OriginalChoiceIndex("WAIT", "ATTACK WIZARD"); found {
					selection = index
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
				if state.MessageContainsGamePackText("wizard-tower.dragons-convinced") {
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
				if index, found := state.OriginalChoiceIndex("WAIT", "PARLAY WITH THE DRAGONS", "PARLAY_SLY"); found {
					selection = index
				}
				if err := state.Select(selection); err != nil {
					log.Fatal(err)
				}
			}
			if !reachedParlayText {
				log.Fatal("-wizard-tower-parlay did not reach the original successful parlay text")
			}
		}
	} else if *opening || *tilvertonDungeon || *inn || *filani || *weaponShop || *temple || *training || *tavern || *highPriest || *carriage || *guildmaster || *sewers {
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
		if *tilvertonDungeon || *inn || *filani || *weaponShop || *temple || *training || *tavern || *highPriest || *carriage || *guildmaster || *sewers {
			if err := state.Select(0); err != nil {
				log.Fatal(err)
			}
			if err := state.Continue(); err != nil {
				log.Fatal(err)
			}
			if err := state.Select(0); err != nil {
				log.Fatal(err)
			}
			if *tilvertonDungeon {
				if state.Mode != game.ModeDungeon {
					log.Fatalf("-tilverton-dungeon normal flow ended in mode %v", state.Mode)
				}
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
			if !*tilvertonDungeon && !*carriage && !*guildmaster && !*sewers {
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
	gameApp := &app{state: state, imagePath: *imagePath, face: regularFace, compactFace: compactFace, partyPath: *partyPath, savgamDir: *savgamDir, savgamSlot: loadedSAVGAMSlot, savgamSlotSave: loadedSAVGAMSlot != 0, soundPlayer: soundPlayer, pc98MusicDriver: pc98MusicDriver, tileImages: tileImages, areaMapSymbols: areaMapSymbols, skyImages: skyImages, geoGrid: geoGrid, areaMapPreview: *areaMapPreview, dungeonFloor: dungeonFloor, dungeonX: dungeonX, dungeonY: dungeonY, geoLabel: geoLabel, geoCatalog: geoCatalog, geoSet: geoRef.Set, geoBlock: geoRef.BlockID, pieceSets: make(map[uint8]gfx.PieceSet), combatSprites: combatSprites, combatSpriteIDs: combatSpriteIDs, combatTerrain: combatTerrain, combatTerrainMode: *combatTerrainMode, gamePack: pack, combatFrame: ebiten.NewImageFromImage(gfx.CombatFrame()), adventureFrame: ebiten.NewImageFromImage(gfx.ExtendedAdventureFrame()), characterStageFrame: ebiten.NewImageFromImage(gfx.CharacterStageFrame()), firstPersonStageFrame: ebiten.NewImageFromImage(gfx.FirstPersonStageFrame()), combatAnimations: combatAnimations, animationStart: time.Now(), combatVisualSerial: visualSerial, combatVisualStarted: visualStarted, combatVisualElapsed: time.Since(visualStarted), screenshotPath: *screenshotPath}
	if *innerFinalBattle && *screenshotPath != "" {
		// Capture-only boss observation camera. Formal play keeps the RuleBook
		// active-fighter camera established above.
		gameApp.combatPreviewFocus = 0x47
	}
	gameApp.state.SetCombatLineTerrain(gameApp.combatLineTerrain())
	gameApp.state.SetCombatScanMapProvider(gameApp.combatScanTacticalMap)
	if *partyLoadPath != "" {
		if err := gameApp.restoreAudioSnapshot(); err != nil {
			log.Fatal(err)
		}
	}
	if err := ebiten.RunGame(gameApp); err != nil {
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

func prepareInnerFinalBattlePreview(state *game.State, grid *geo.Grid) error {
	if state.Mode == game.ModeWilderness {
		if err := state.Select(0); err != nil {
			return fmt.Errorf("enter inner ruins for final-battle preview: %w", err)
		}
	}
	if state.Mode != game.ModeDungeon {
		return fmt.Errorf("final-battle preview initialization ended in mode %v", state.Mode)
	}
	previewParty := state.PartyFighters()
	if len(previewParty) == 0 {
		return fmt.Errorf("final-battle preview has no party")
	}
	for index := range previewParty {
		previewParty[index].HitPoints, previewParty[index].MaxHitPoints = 999, 999
		previewParty[index].ArmorClass = -10
		previewParty[index].AttackBonus = 100
		previewParty[index].DamageDiceCount, previewParty[index].DamageDiceSides = 1, 1
		previewParty[index].DamageBonus = 100
		previewParty[index].AttacksPerTurn = 8
		previewParty[index].InitiativeBonus = 100
	}
	if err := state.SetParty(previewParty); err != nil {
		return fmt.Errorf("install bounded final-battle preview party: %w", err)
	}
	// The direct visual checkpoint starts after the separately verified final
	// ritual. This is preview context only: terrain 9Ah still dispatches the
	// original ECL text, monster records and combat boundary.
	state.SetECLMemoryValue(0x4C00, 1)
	state.SetECLMemoryValue(0x4CBD, 1)
	state.SetECLMemoryValue(0x4CC7, 1)
	// Enter the second floor through terrain 97h so the original staircase
	// transaction establishes its floor-local context. The remaining ten
	// steps are the shortest legal route recorded by READY spec 408.
	state.SetDungeonGeometryView(10, 7, 2)
	state.DungeonWallRoof = grid.CellWrapped(10, 7).Terrain
	if err := state.RunDungeonLifecycle(); err != nil {
		return fmt.Errorf("enter inner-ruins staircase terrain 97: %w", err)
	}
	if state.Mode != game.ModeWilderness || len(state.Choices) == 0 {
		return fmt.Errorf("inner-ruins staircase did not request ascent: mode=%v choices=%v", state.Mode, state.Choices)
	}
	if err := state.Select(0); err != nil {
		return fmt.Errorf("ascend inner-ruins staircase: %w", err)
	}
	route := []struct {
		x, y      int
		direction uint8
	}{
		{2, 4, 0}, {2, 3, 0}, {2, 2, 0}, {2, 1, 0}, {2, 0, 0},
		{3, 0, 2}, {4, 0, 2}, {5, 0, 2}, {6, 0, 2}, {6, 1, 4},
	}
	for index, point := range route {
		state.SetDungeonGeometryView(point.x, point.y, point.direction)
		state.DungeonWallRoof = grid.CellWrapped(point.x, point.y).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			return fmt.Errorf("final route step %d: %w", index, err)
		}
		for boundary := 0; index < len(route)-1 && state.Mode != game.ModeDungeon && boundary < 6; boundary++ {
			if state.CombatActive() {
				for action := 0; action < 256 && state.CombatActive(); action++ {
					if err := state.CombatAct(); err != nil {
						return fmt.Errorf("resolve final route step %d combat: %w", index, err)
					}
					if event, ok := state.CombatVisualEvent(); ok {
						if err := state.AdvanceCombatVisual(event.Duration()); err != nil {
							return fmt.Errorf("advance final route step %d combat visual: %w", index, err)
						}
					}
				}
				if state.CombatActive() || state.CombatStatus() != combat.StatusPartyWon {
					return fmt.Errorf("final route step %d combat status=%v active=%v", index, state.CombatStatus(), state.CombatActive())
				}
				continue
			}
			switch state.Mode {
			case game.ModeEvent:
				if err := state.Continue(); err != nil {
					return fmt.Errorf("continue final route step %d: %w", index, err)
				}
			case game.ModeWilderness:
				if len(state.Choices) == 0 {
					return fmt.Errorf("final route step %d pause has no choice", index)
				}
				if err := state.Select(0); err != nil {
					return fmt.Errorf("advance final route step %d: %w", index, err)
				}
			default:
				return fmt.Errorf("final route step %d stopped in mode %v", index, state.Mode)
			}
		}
		if index < len(route)-1 && state.Mode != game.ModeDungeon {
			return fmt.Errorf("final route step %d did not resume dungeon mode", index)
		}
	}
	for step := 0; step < 8 && !state.CombatActive(); step++ {
		switch state.Mode {
		case game.ModeEvent:
			if err := state.Continue(); err != nil {
				return fmt.Errorf("continue final confrontation: %w", err)
			}
		case game.ModeWilderness:
			if len(state.Choices) == 0 {
				return fmt.Errorf("final confrontation pause has no choice")
			}
			if err := state.Select(0); err != nil {
				return fmt.Errorf("advance final confrontation: %w", err)
			}
		default:
			return fmt.Errorf("final confrontation stalled in mode %v", state.Mode)
		}
	}
	if !state.CombatActive() {
		return fmt.Errorf("final confrontation did not reach combat")
	}
	counts := map[uint8]int{}
	for _, fighter := range state.CombatFighters() {
		if fighter.Side == combat.SideEnemy {
			counts[fighter.SpriteBlock]++
		}
	}
	if counts[0x45] != 28 || counts[0x47] != 1 || counts[0x48] != 8 {
		return fmt.Errorf("final confrontation enemies=%v", counts)
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
func demoParty(state *game.State) []combat.Fighter {
	return []combat.Fighter{
		{ID: "party-1", Name: state.DemoFighterName(game.DemoNamePartyFighter), Side: combat.SideParty, HitPoints: 42, MaxHitPoints: 42, ArmorClass: 4, AttackBonus: 16, DamageDiceCount: 1, DamageDiceSides: 8, InitiativeBonus: 1},
		{ID: "party-2", Name: state.DemoFighterName(game.DemoNamePartyRanger), Side: combat.SideParty, HitPoints: 34, MaxHitPoints: 34, ArmorClass: 5, AttackBonus: 15, DamageDiceCount: 1, DamageDiceSides: 8, InitiativeBonus: 2},
		{ID: "party-3", Name: state.DemoFighterName(game.DemoNamePartyCleric), Side: combat.SideParty, HitPoints: 30, MaxHitPoints: 30, ArmorClass: 6, AttackBonus: 12, DamageDiceCount: 1, DamageDiceSides: 6, InitiativeBonus: 0},
		{ID: "party-4", Name: state.DemoFighterName(game.DemoNamePartyWizard), Side: combat.SideParty, HitPoints: 20, MaxHitPoints: 20, ArmorClass: 8, AttackBonus: 10, DamageDiceCount: 1, DamageDiceSides: 4, InitiativeBonus: 1},
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
