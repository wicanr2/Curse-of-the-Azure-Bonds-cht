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
	cell := float64(height) * 0.026
	gridSize := cell * 16
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			visible := a.ui.guideFull || a.guideCellExplored(mapKey, x, y)
			fill := color.RGBA{12, 18, 30, 255}
			if visible {
				fill = color.RGBA{42, 54, 66, 255}
			}
			ebitenutil.DrawRect(screen, originX+float64(x)*cell, originY+float64(y)*cell, cell-1, cell-1, fill)
		}
	}
	currentX, currentY, _ := a.state.DungeonGeometryView()
	ebitenutil.DrawRect(screen, originX+float64(currentX)*cell, originY+float64(currentY)*cell, cell-1, cell-1, color.RGBA{92, 220, 255, 255})
	visiblePoints := make([]guidePoint, 0, len(definition.Points))
	for _, point := range definition.Points {
		if !a.ui.guideFull && !a.guideCellExplored(mapKey, point.X, point.Y) {
			continue
		}
		visiblePoints = append(visiblePoints, point)
		ebitenutil.DrawRect(screen, originX+float64(point.X)*cell+cell*.25, originY+float64(point.Y)*cell+cell*.25, cell*.5, cell*.5, color.RGBA{255, 210, 64, 255})
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
		lines = "尚未探索到已標記事件。\n完整模式會顯示所有攻略點。"
	}
	drawWrappedText(screen, lines, a.compactFace, textX, int(originY), maxInt(12, textWidth/faceCellWidth(a.compactFace)), 18, 18, color.RGBA{232, 238, 255, 255})
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
