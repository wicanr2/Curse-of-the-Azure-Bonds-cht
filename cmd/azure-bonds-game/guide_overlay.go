package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/game"
	enginearea "github.com/wicanr2/golden-box-remake-engine/areamap"
)

type guidePoint struct {
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Label   string `json:"label"`
	Summary string `json:"summary"`
	Source  string `json:"source"`
}

type guideMap struct {
	Title  string       `json:"title"`
	Points []guidePoint `json:"points"`
}

type guideCatalog struct {
	Schema string              `json:"schema"`
	Maps   map[string]guideMap `json:"maps"`
}

func loadGuideCatalog(path string) (guideCatalog, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return guideCatalog{}, err
	}
	var catalog guideCatalog
	if err := json.Unmarshal(raw, &catalog); err != nil {
		return guideCatalog{}, err
	}
	if catalog.Schema != "coab-guide-maps/1" {
		return guideCatalog{}, fmt.Errorf("unsupported guide schema %q", catalog.Schema)
	}
	return catalog, nil
}

func guideMapKey(set, block uint8) string { return fmt.Sprintf("%d/%02X", set, block) }
func guideCellKey(x, y int) string        { return fmt.Sprintf("%d,%d", x, y) }

func (a *app) rememberCurrentGuideCell() {
	if a.state == nil || a.state.Mode != game.ModeDungeon || a.state.GeoMapSet == 0 {
		return
	}
	if a.ui.settings.Explored == nil {
		a.ui.settings.Explored = map[string][]string{}
	}
	mapKey := guideMapKey(a.state.GeoMapSet, a.state.GeoMapBlock)
	x, y, _ := a.state.DungeonGeometryView()
	cell := guideCellKey(x, y)
	for _, existing := range a.ui.settings.Explored[mapKey] {
		if existing == cell {
			return
		}
	}
	a.ui.settings.Explored[mapKey] = append(a.ui.settings.Explored[mapKey], cell)
	sort.Strings(a.ui.settings.Explored[mapKey])
	a.ui.settingsDirty = true
	a.ui.exploredSinceSave++
	// 避免每走一格都在 Update 內同步寫磁碟；每八個新格做一次有界 checkpoint，
	// 關閉攻略與 F10 也會沖刷尚未保存的尾段。
	if a.ui.exploredSinceSave >= 8 {
		a.persistUISettings()
	}
}

func (a *app) guideCellExplored(mapKey string, x, y int) bool {
	for _, cell := range a.ui.settings.Explored[mapKey] {
		if cell == guideCellKey(x, y) {
			return true
		}
	}
	return false
}

func (a *app) drawGuideGrid(screen *ebiten.Image, width, height int, definition guideMap) {
	mapKey := guideMapKey(a.state.GeoMapSet, a.state.GeoMapBlock)
	originX, originY := float64(width)*0.10, float64(height)*0.27
	cell := float64(height) / 30
	gridSize := cell * enginearea.OriginalViewSize
	currentX, currentY, direction := a.state.DungeonGeometryView()
	if a.geoGrid == nil {
		drawWrappedText(screen, a.state.LocalizedText("ui_guide_area_unavailable", "The original AREA map is unavailable."), a.compactFace, int(originX), int(originY), 24, 18, 3, color.RGBA{232, 238, 255, 255})
		return
	}
	view, err := enginearea.BuildOriginal(*a.geoGrid, currentX, currentY, int(direction))
	if err != nil || len(a.areaMapSymbols) < 20 {
		drawWrappedText(screen, a.state.LocalizedText("ui_guide_area_unavailable", "The original AREA map is unavailable."), a.compactFace, int(originX), int(originY), 24, 18, 3, color.RGBA{232, 238, 255, 255})
		return
	}
	drawSymbol := func(item, screenX, screenY int) {
		if item < 0 || item >= len(a.areaMapSymbols) {
			return
		}
		options := &ebiten.DrawImageOptions{}
		options.Filter = ebiten.FilterNearest
		options.GeoM.Scale(cell/8, cell/8)
		options.GeoM.Translate(originX+float64(screenX)*cell, originY+float64(screenY)*cell)
		screen.DrawImage(a.areaMapSymbols[item], options)
	}
	for _, tile := range view.Tiles {
		// 探索模式仍保留迷霧，但已探索格直接畫原版 AREA symbol，
		// 不再以 remake 自繪灰色方格替代原作地圖語彙。
		if a.ui.guideFull || a.guideCellExplored(mapKey, tile.MapX, tile.MapY) {
			drawSymbol(tile.Item, tile.ScreenX, tile.ScreenY)
		} else {
			ebitenutil.DrawRect(screen, originX+float64(tile.ScreenX)*cell, originY+float64(tile.ScreenY)*cell, cell, cell, color.RGBA{8, 12, 20, 255})
		}
	}
	drawSymbol(view.PartyItem, view.PartyScreenX, view.PartyScreenY)
	visiblePoints := make([]guidePoint, 0, len(definition.Points))
	for _, point := range definition.Points {
		if !a.ui.guideFull && !a.guideCellExplored(mapKey, point.X, point.Y) {
			continue
		}
		visiblePoints = append(visiblePoints, point)
		screenX, screenY := point.X-view.OffsetX, point.Y-view.OffsetY
		if screenX >= 0 && screenX < enginearea.OriginalViewSize && screenY >= 0 && screenY < enginearea.OriginalViewSize {
			ebitenutil.DrawRect(screen, originX+float64(screenX)*cell+cell*.28, originY+float64(screenY)*cell+cell*.28, cell*.44, cell*.44, color.RGBA{255, 210, 64, 255})
		}
	}
	textX := int(originX + gridSize + float64(width)*0.04)
	textWidth := int(float64(width)*0.84) - textX
	if textWidth < 120 {
		textWidth = 120
	}
	lines := ""
	for index, point := range visiblePoints {
		if index >= 7 {
			lines += "…\n"
			break
		}
		lines += fmt.Sprintf("● (%d,%d) %s\n%s\n", point.X, point.Y, point.Label, point.Summary)
	}
	if lines == "" {
		lines = a.state.LocalizedText("ui_guide_no_points", "No marked event has been explored.\nFull mode shows every guide point.")
	}
	drawWrappedText(screen, lines, a.compactFace, textX, int(originY), maxInt(12, textWidth/faceCellWidth(a.compactFace)), 18, 18, color.RGBA{232, 238, 255, 255})
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
