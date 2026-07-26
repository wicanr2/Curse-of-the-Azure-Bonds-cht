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

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/game"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/gfx"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/mapdata"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

const (
	logicalWidth  = 640
	logicalHeight = 400
)

type app struct {
	state            game.State
	face             font.Face
	choiceCursor     int
	partyPath        string
	tilePreview      bool
	tileImages       []*ebiten.Image
	geoPreview       bool
	geoGrid          *geo.Grid
	geoX             int
	geoY             int
	geoLabel         string
	geoCatalog       geo.Catalog
	geoSet           uint8
	geoBlock         uint8
	dungeonPreview   bool
	dungeonFloor     *mapdata.DungeonFloor
	combatSprites    map[string]*ebiten.Image
	combatSpriteIDs  []string
	combatAnimations map[string][]combatAnimation
	animationStart   time.Time
}

type combatAnimation struct {
	image *ebiten.Image
	delay uint32
	x     int16
	y     int16
}

func (a *app) Update() error {
	a.syncGeoMapRequest()
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
	if a.dungeonPreview {
		if inpututil.IsKeyJustPressed(ebiten.KeyD) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			a.dungeonPreview = false
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
	if inpututil.IsKeyJustPressed(ebiten.KeyG) && a.geoGrid != nil {
		a.geoPreview = true
		a.geoX, a.geoY = 0, 0
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) && a.dungeonFloor != nil {
		a.dungeonPreview = true
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
			if a.state.ConsumeSaveRequest() {
				if saveErr := a.state.SavePartyFile(a.partyPath); saveErr != nil {
					a.state.Message = "儲存失敗：" + saveErr.Error()
				} else {
					a.state.Message = "隊伍已儲存：" + a.partyPath
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
	floor := mapdata.GenerateDungeon(grid, 8, 8)
	a.dungeonFloor = &floor
	a.geoSet, a.geoBlock = set, block
	a.geoLabel = fmt.Sprintf("GEO%d block 0x%02X", set, block)
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
	if a.dungeonPreview {
		a.drawDungeonPreview(screen, white, cyan)
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
	if a.state.Mode == game.ModeMap {
		a.drawWildernessMap(screen, white, cyan)
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
		if a.state.PictureRequested {
			a.drawPictureAnimation(screen)
			return
		}
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

func (a *app) drawPictureAnimation(screen *ebiten.Image) {
	if a.state.SceneCharacterRequested {
		key := fmt.Sprintf("character-area-%d-head-%02X-body-%02X.png", a.state.Area.GameArea, a.state.SceneHeadBlock, a.state.SceneBodyBlock)
		if sprite := a.combatSprites[key]; sprite != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(2, 2)
			op.GeoM.Translate(float64((logicalWidth-sprite.Bounds().Dx()*2)/2), 52)
			screen.DrawImage(sprite, op)
			text.Draw(screen, "人物場景　Enter：繼續", a.face, 56, 350, color.RGBA{255, 255, 255, 255})
			return
		}
		text.Draw(screen, "人物圖層素材尚未載入", a.face, 56, 220, color.RGBA{255, 220, 100, 255})
		text.Draw(screen, "Enter：繼續", a.face, 56, 330, color.RGBA{255, 255, 255, 255})
		return
	}
	if a.state.BigPictureRequested {
		key := fmt.Sprintf("bigpic%d-block-%02X-item-00.png", a.state.Area.GameArea, a.state.PictureBlock)
		if sprite := a.combatSprites[key]; sprite != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64((logicalWidth-sprite.Bounds().Dx())/2), float64((logicalHeight-sprite.Bounds().Dy())/2))
			screen.DrawImage(sprite, op)
			text.Draw(screen, "大幅事件畫面　Enter：繼續", a.face, 56, 380, color.RGBA{255, 255, 255, 255})
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
	op.GeoM.Scale(2, 2)
	op.GeoM.Translate(160+float64(frame.x*2), 76+float64(frame.y*2))
	screen.DrawImage(frame.image, op)
	text.Draw(screen, "事件畫面　Enter：繼續", a.face, 56, 350, color.RGBA{255, 255, 255, 255})
}

func (a *app) drawWildernessMap(screen *ebiten.Image, white, cyan color.Color) {
	text.Draw(screen, "暗影谷荒野（原版 50×25 floor）", a.face, 24, 28, cyan)
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
	text.Draw(screen, "Dungeon floor composition（D／Esc：返回）", a.face, 24, 28, cyan)
	if a.dungeonFloor == nil {
		text.Draw(screen, "沒有載入 dungeon floor", a.face, 24, 70, white)
		return
	}
	const (
		viewWidth  = 13
		viewHeight = 5
		originX    = 24
		originY    = 52
		tileSize   = 24
	)
	startX, startY := 18, 8
	for row := 0; row < viewHeight; row++ {
		for column := 0; column < viewWidth; column++ {
			x, y := startX+column, startY+row
			entry, ok := a.dungeonFloor.Entry(x, y)
			left, top := originX+column*tileSize, originY+row*tileSize
			if ok && int(entry.TileIndex) < len(a.tileImages) {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(left), float64(top))
				screen.DrawImage(a.tileImages[entry.TileIndex], op)
			} else {
				ebitenutil.DrawRect(screen, float64(left), float64(top), tileSize, tileSize, color.RGBA{24, 30, 48, 255})
			}
		}
	}
	text.Draw(screen, "GEO wall/door → 13×5 dungeon background entries → TILES pixel art", a.face, 24, 210, white)
	text.Draw(screen, "目前為 "+a.geoLabel+" map center (8,8) 的可重現 floor slice", a.face, 24, 245, white)
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
	partyIndex, enemyIndex := 0, 0
	targets := a.state.CombatTargets()
	for _, fighter := range a.state.CombatFighters() {
		if fighter.Side == combat.SideParty {
			tile := combat.FormationTile(fighter.Side, partyIndex)
			if fighter.HasCombatPosition {
				tile = combat.TilePoint{X: fighter.CombatX, Y: fighter.CombatY}
			}
			x, y := 28+tile.X*48, 108+tile.Y*56
			a.drawFighterSprite(screen, fighter, partyIndex, x, y)
			text.Draw(screen, fighter.Name, a.face, x, y+66, white)
			text.Draw(screen, strconv.Itoa(fighter.HitPoints)+"/"+strconv.Itoa(fighter.MaxHitPoints), a.face, x, y+66, white)
			partyIndex++
			continue
		}
		tile := combat.FormationTile(fighter.Side, enemyIndex)
		if fighter.HasCombatPosition {
			tile = combat.TilePoint{X: fighter.CombatX, Y: fighter.CombatY}
		}
		x, y := 28+tile.X*48, 108+tile.Y*56
		a.drawFighterSprite(screen, fighter, enemyIndex, x, y)
		prefix := "  "
		if len(targets) > 0 && a.state.CombatTargetIndex() < len(targets) && targets[a.state.CombatTargetIndex()].ID == fighter.ID {
			prefix = "> "
		}
		text.Draw(screen, prefix+fighter.Name, a.face, x, y+66, white)
		text.Draw(screen, strconv.Itoa(fighter.HitPoints)+"/"+strconv.Itoa(fighter.MaxHitPoints), a.face, x, y+84, white)
		enemyIndex++
	}
	text.Draw(screen, "左右：選擇目標　Enter：攻擊", a.face, 32, 350, cyan)
}

func (a *app) drawFighterSprite(screen *ebiten.Image, fighter combat.Fighter, ordinal, x, y int) {
	if len(a.combatSprites) == 0 {
		return
	}
	key := ""
	var sprite *ebiten.Image
	var frameX, frameY int16
	if fighter.Side == combat.SideParty {
		// Party icons are the original CHEAD+CBODY composition. The current
		// roster has deterministic slots; character-specific icon IDs will
		// replace this selector when the player record decoder is complete.
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
	geoSet := flag.Int("geo-set", 2, "GEO DAX set/chapter (2..6) used by the map preview")
	geoBlock := flag.Int("geo-block", 1, "original GEO block ID used by the map preview")
	encounter := flag.Bool("encounter", false, "start the observed ECL1 encounter directly")
	encounterBlock := flag.Int("encounter-block", 81, "ECL block for -encounter")
	encounterStart := flag.Int("encounter-start", 0x1293, "payload offset for -encounter")
	partyPath := flag.String("party-save", "party.json", "versioned remake party save path")
	partyLoadPath := flag.String("party-load", "", "load a versioned remake party save before starting")
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
	dungeonFloor := loadDungeonPreview(geoGrid)
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
	geoLabel := fmt.Sprintf("GEO%d block 0x%02X", *geoSet, *geoBlock)
	combatSprites, combatSpriteIDs, combatAnimations, err := loadCombatSprites()
	if err != nil {
		log.Fatal(err)
	}
	if err := ebiten.RunGame(&app{state: state, face: loadFace(*fontPath), partyPath: *partyPath, tileImages: tileImages, geoGrid: geoGrid, dungeonFloor: dungeonFloor, geoLabel: geoLabel, geoCatalog: geoCatalog, geoSet: geoRef.Set, geoBlock: geoRef.BlockID, combatSprites: combatSprites, combatSpriteIDs: combatSpriteIDs, combatAnimations: combatAnimations, animationStart: time.Now()}); err != nil {
		log.Fatal(err)
	}
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
	floor := mapdata.GenerateDungeon(*grid, 8, 8)
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
