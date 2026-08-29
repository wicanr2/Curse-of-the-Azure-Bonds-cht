package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/game"
)

type uiSettings struct {
	Schema         string              `json:"schema"`
	Theme          string              `json:"theme"`
	Language       string              `json:"language"`
	Width          int                 `json:"width"`
	Height         int                 `json:"height"`
	SpoilerWarning bool                `json:"spoiler_warning_acknowledged"`
	Explored       map[string][]string `json:"explored,omitempty"`
}

type uiRuntime struct {
	settings          uiSettings
	settingsPath      string
	helpOpen          bool
	guideOpen         bool
	guideFull         bool
	spoilerAsk        bool
	settingsDirty     bool
	exploredSinceSave int
	resizeWindow      func(int, int)
}

func defaultUISettingsPath() string {
	root, err := os.UserConfigDir()
	if err != nil || root == "" {
		return "ui-settings.json"
	}
	return filepath.Join(root, "curse-of-the-azure-bonds-remake", "ui-settings.json")
}

func defaultUISettings() uiSettings {
	return uiSettings{Schema: "coab-ui-settings/1", Theme: "modern-a6", Language: "zh-TW", Width: 640, Height: 480, Explored: map[string][]string{}}
}

func loadUISettings(path string) uiSettings {
	settings := defaultUISettings()
	raw, err := os.ReadFile(path)
	if err != nil || json.Unmarshal(raw, &settings) != nil || settings.Schema != "coab-ui-settings/1" {
		return defaultUISettings()
	}
	if settings.Theme != "original" && settings.Theme != "modern-a6" {
		settings.Theme = "modern-a6"
	}
	if !validLanguage(settings.Language) {
		settings.Language = "zh-TW"
	}
	if !validResolution(settings.Width, settings.Height) {
		settings.Width, settings.Height = 640, 480
	}
	if settings.Explored == nil {
		settings.Explored = map[string][]string{}
	}
	return settings
}

func saveUISettings(path string, settings uiSettings) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}

func newUIRuntime(settings uiSettings, path string) uiRuntime {
	return uiRuntime{settings: settings, settingsPath: path}
}

func validResolution(width, height int) bool {
	return (width == 640 && height == 480) || (width == 1024 && height == 768) || (width == 1280 && height == 960)
}

func validLanguage(language string) bool {
	switch language {
	case "zh-TW", "zh-CN", "ja", "en":
		return true
	default:
		return false
	}
}

func (ui *uiRuntime) logicalSize() (int, int) {
	if !validResolution(ui.settings.Width, ui.settings.Height) {
		return 640, 480
	}
	return ui.settings.Width, ui.settings.Height
}

func (a *app) persistUISettings() {
	if a.ui.settingsPath == "" {
		return
	}
	if err := saveUISettings(a.ui.settingsPath, a.ui.settings); err != nil {
		a.state.Message = "Could not save UI settings: " + err.Error()
		return
	}
	a.ui.settingsDirty = false
	a.ui.exploredSinceSave = 0
}

func (a *app) updateGlobalUI() (bool, error) {
	if a.justPressed(ebiten.KeyEscape) {
		switch {
		case a.ui.spoilerAsk:
			a.ui.spoilerAsk = false
			return true, nil
		case a.ui.helpOpen:
			a.ui.helpOpen = false
			return true, nil
		case a.ui.guideOpen:
			a.ui.guideOpen = false
			a.ui.guideFull = false
			if a.ui.settingsDirty {
				a.persistUISettings()
			}
			return true, nil
		}
	}
	if a.ui.spoilerAsk {
		if a.justPressed(ebiten.KeyY) || a.justPressed(ebiten.KeyEnter) {
			a.ui.settings.SpoilerWarning = true
			a.ui.spoilerAsk = false
			a.ui.guideFull = true
			a.persistUISettings()
		}
		return true, nil
	}
	if a.justPressed(ebiten.KeyF1) {
		a.ui.helpOpen = !a.ui.helpOpen
		a.ui.guideOpen = false
		return true, nil
	}
	if a.justPressed(ebiten.KeyF2) {
		if a.ui.settings.Theme == "modern-a6" {
			a.ui.settings.Theme = "original"
		} else {
			a.ui.settings.Theme = "modern-a6"
		}
		a.persistUISettings()
		return true, nil
	}
	if a.justPressed(ebiten.KeyF3) {
		if a.ui.guideOpen && a.ui.settingsDirty {
			a.persistUISettings()
		}
		a.ui.guideOpen = !a.ui.guideOpen
		a.ui.helpOpen = false
		a.ui.guideFull = false
		return true, nil
	}
	if a.justPressed(ebiten.KeyF4) {
		resolutions := [][2]int{{640, 480}, {1024, 768}, {1280, 960}}
		index := 0
		for i, resolution := range resolutions {
			if resolution[0] == a.ui.settings.Width && resolution[1] == a.ui.settings.Height {
				index = i
				break
			}
		}
		next := resolutions[(index+1)%len(resolutions)]
		a.ui.settings.Width, a.ui.settings.Height = next[0], next[1]
		if a.ui.resizeWindow != nil {
			a.ui.resizeWindow(next[0], next[1])
		}
		a.persistUISettings()
		return true, nil
	}
	if a.justPressed(ebiten.KeyF6) {
		a.cycleLanguage()
		return true, nil
	}
	if a.ui.guideOpen && a.justPressed(ebiten.KeyV) {
		if !a.ui.guideFull && !a.ui.settings.SpoilerWarning {
			a.ui.spoilerAsk = true
		} else {
			a.ui.guideFull = !a.ui.guideFull
		}
		return true, nil
	}
	if a.ui.helpOpen || a.ui.guideOpen {
		return true, nil
	}
	if a.justPressed(ebiten.KeyF10) {
		if a.ui.settingsDirty {
			a.persistUISettings()
		}
		if len(a.state.PartyFighters()) == 0 {
			return true, ebiten.Termination
		}
		if err := a.saveCurrentGame(); err != nil {
			a.state.Message = a.state.FileOperationMessage(game.FileOperationSave, game.FileOperationFailed, err.Error())
			return true, nil
		}
		return true, ebiten.Termination
	}
	return false, nil
}

func (a *app) drawGlobalUI(screen *ebiten.Image) {
	w, h := a.ui.logicalSize()
	if a.ui.settings.Theme == "modern-a6" {
		a.drawModernA6Content(screen, w, h)
		switch a.state.Mode {
		case game.ModeCombat:
			if !a.drawModernA6CombatFrame(screen, w, h) {
				drawA6CombatFrame(screen, w, h)
			}
		case game.ModeWilderness, game.ModeEvent, game.ModeMap, game.ModePlace, game.ModeJournal, game.ModeDungeon:
			if !a.drawModernA6AdventureFrame(screen, w, h) {
				drawA6Frame(screen, w, h)
			}
		default:
			drawA6OuterFrame(screen, w, h)
		}
	}
	if a.ui.helpOpen {
		a.drawHelpOverlay(screen, w, h)
	}
	if a.ui.guideOpen {
		a.drawGuideOverlay(screen, w, h)
	}
	if a.ui.spoilerAsk {
		a.drawSpoilerWarning(screen, w, h)
	}
}

func drawA6OuterFrame(screen *ebiten.Image, width, height int) {
	sx, sy := float64(width)/640, float64(height)/480
	stone, light, shadow := color.RGBA{218, 211, 190, 255}, color.RGBA{247, 239, 211, 255}, color.RGBA{88, 78, 61, 255}
	ebitenutil.DrawRect(screen, 2*sx, 2*sy, 636*sx, 4*sy, stone)
	ebitenutil.DrawRect(screen, 2*sx, 2*sy, 636*sx, sy, light)
	ebitenutil.DrawRect(screen, 2*sx, 474*sy, 636*sx, 4*sy, shadow)
	ebitenutil.DrawRect(screen, 2*sx, 2*sy, 4*sx, 476*sy, stone)
	ebitenutil.DrawRect(screen, 634*sx, 2*sy, 4*sx, 476*sy, shadow)
}

func drawA6CombatFrame(screen *ebiten.Image, width, height int) {
	sx, sy := float64(width)/640, float64(height)/480
	stone, light, shadow := color.RGBA{205, 196, 172, 255}, color.RGBA{248, 238, 204, 255}, color.RGBA{66, 57, 45, 255}
	gold, goldLight, goldShadow := color.RGBA{255, 213, 45, 255}, color.RGBA{255, 249, 171, 255}, color.RGBA{133, 75, 3, 255}
	blue := color.RGBA{35, 137, 219, 255}
	line := func(x, y, w, h float64, c color.Color) { ebitenutil.DrawRect(screen, x*sx, y*sy, w*sx, h*sy, c) }
	// 使用者選定的 18px 精修石框；內緣是較細的雙層亮金雕線。
	line(0, 0, 640, 18, stone)
	line(0, 462, 640, 18, shadow)
	line(0, 0, 18, 480, stone)
	line(622, 0, 18, 480, shadow)
	line(2, 2, 636, 2, light)
	line(14, 14, 612, 2, goldLight)
	line(14, 16, 612, 1, gold)
	line(14, 17, 612, 1, goldShadow)
	line(14, 462, 612, 1, goldLight)
	line(14, 463, 612, 2, gold)
	line(14, 465, 612, 1, goldShadow)
	line(14, 14, 2, 452, goldLight)
	line(16, 14, 1, 452, gold)
	line(17, 14, 1, 452, goldShadow)
	line(622, 14, 1, 452, goldLight)
	line(623, 14, 2, 452, gold)
	line(625, 14, 1, 452, goldShadow)
	// 戰場／狀態與訊息區分隔延續同一金雕語彙，不再退回灰色 DOS 框。
	for _, divider := range [][4]float64{{358, 16, 8, 344}, {16, 358, 608, 8}, {16, 446, 608, 7}} {
		x, y, w, h := divider[0], divider[1], divider[2], divider[3]
		line(x, y, w, h, shadow)
		if w > h {
			line(x, y+1, w, 2, goldLight)
			line(x, y+3, w, 2, gold)
		} else {
			line(x+1, y, 2, h, goldLight)
			line(x+3, y, 2, h, gold)
		}
	}
	for _, point := range [][2]float64{{14, 14}, {622, 14}, {14, 462}, {622, 462}, {358, 358}} {
		line(point[0], point[1], 5, 5, goldShadow)
		line(point[0]+1, point[1]+1, 3, 3, blue)
		line(point[0]+2, point[1]+1, 1, 1, color.RGBA{159, 226, 255, 255})
	}
}

func drawA6Frame(screen *ebiten.Image, width, height int) {
	sx, sy := float64(width)/640, float64(height)/480
	stone := color.RGBA{218, 211, 190, 255}
	stoneLight := color.RGBA{247, 239, 211, 255}
	shadow := color.RGBA{88, 78, 61, 255}
	gold := color.RGBA{255, 218, 61, 255}
	goldLight := color.RGBA{255, 246, 151, 255}
	goldShadow := color.RGBA{145, 87, 5, 255}
	line := func(x, y, w, h float64, c color.Color) { ebitenutil.DrawRect(screen, x*sx, y*sy, w*sx, h*sy, c) }
	// A6 沿用 DOS 冒險畫面的實際分割：左舞台 0..263、右狀態 272..639、
	// 下敘事 272..455。石條刻意只有 4..6 px，避免吃掉內容。
	line(2, 2, 636, 4, stone)
	line(2, 2, 636, 1, stoneLight)
	line(2, 474, 636, 4, shadow)
	line(2, 2, 4, 476, stone)
	line(634, 2, 4, 476, shadow)
	line(264, 2, 6, 254, stone)
	line(269, 2, 1, 254, shadow)
	line(2, 256, 636, 6, stone)
	line(2, 261, 636, 1, shadow)
	line(2, 454, 636, 5, stone)
	line(2, 454, 636, 1, stoneLight)
	// 稀疏鑿刻節奏：外框只留下低密度的菱形投影，不與內層金雕競爭。
	for x := 12.0; x < 630; x += 18 {
		line(x, 3, 5, 1, shadow)
		line(x+2, 4, 5, 1, stoneLight)
		line(x, 475, 5, 1, stoneLight)
		line(x+2, 476, 5, 1, shadow)
	}
	for y := 12.0; y < 446; y += 18 {
		line(3, y, 1, 5, shadow)
		line(4, y+2, 1, 5, stoneLight)
		line(635, y, 1, 5, stoneLight)
		line(636, y+2, 1, 5, shadow)
	}
	// 左上場景／人物的第二層：對準現行 184×184 舞台，使用 2 px 明亮金雕邊。
	line(48, 36, 188, 2, goldLight)
	line(48, 230, 188, 2, goldShadow)
	line(48, 36, 2, 196, goldLight)
	line(234, 36, 2, 196, goldShadow)
	line(50, 38, 184, 1, gold)
	line(50, 229, 184, 1, gold)
	for x := 52.0; x < 232; x += 6 {
		line(x, 36, 3, 1, goldShadow)
		line(x+2, 37, 3, 1, gold)
	}
	for y := 40.0; y < 228; y += 6 {
		line(48, y, 1, 3, goldShadow)
		line(49, y+2, 1, 3, gold)
	}
}

func (a *app) drawOverlayPanel(screen *ebiten.Image, width, height int, title string) {
	ebitenutil.DrawRect(screen, float64(width)*0.06, float64(height)*0.08, float64(width)*0.88, float64(height)*0.84, color.RGBA{2, 5, 12, 246})
	ebitenutil.DrawRect(screen, float64(width)*0.06, float64(height)*0.08, float64(width)*0.88, 2, color.RGBA{255, 210, 64, 255})
	drawFittedText(screen, title, a.face, int(float64(width)*0.09), int(float64(height)*0.15), int(float64(width)*0.82), color.RGBA{255, 220, 92, 255})
}

func (a *app) drawHelpOverlay(screen *ebiten.Image, width, height int) {
	a.drawOverlayPanel(screen, width, height, a.state.LocalizedText("ui_help_title", "Help / Controls"))
	text := a.state.LocalizedText("ui_help_body", "F1 Help\nF2 Theme\nF3 Guide Map\nF4 Resolution\nF5 Save  F9 Load\nF6 Language\nF10 Save and quit\nEsc Cancel / Back")
	drawWrappedText(screen, text, a.compactFace, int(float64(width)*0.10), int(float64(height)*0.24), 42, 24, 10, color.RGBA{232, 238, 255, 255})
}

func (a *app) drawGuideOverlay(screen *ebiten.Image, width, height int) {
	mode := a.state.LocalizedText("ui_guide_explore", "Exploration Information")
	if a.ui.guideFull {
		mode = a.state.LocalizedText("ui_guide_full", "Full Guide (spoilers)")
	}
	a.drawOverlayPanel(screen, width, height, fmt.Sprintf(a.state.LocalizedText("ui_guide_title", "Current Guide Map | %s"), mode))
	if a.state.Mode != game.ModeDungeon && a.state.Mode != game.ModeMap {
		drawWrappedText(screen, a.state.LocalizedText("ui_guide_unavailable", "No guide is available on this screen.\nEnter a dungeon or area map, then press F3."), a.compactFace, int(float64(width)*0.12), int(float64(height)*0.32), 36, 28, 6, color.RGBA{232, 238, 255, 255})
		drawFittedText(screen, a.state.LocalizedText("ui_guide_close", "F3 / Esc: Close"), a.compactFace, int(float64(width)*0.10), int(float64(height)*0.88), int(float64(width)*0.80), color.RGBA{232, 238, 255, 255})
		return
	}
	x, y, direction := a.state.DungeonGeometryView()
	mapKey := guideMapKey(a.state.GeoMapSet, a.state.GeoMapBlock)
	definition, _ := a.guide.Maps[mapKey]
	mapTitle := definition.Title
	if mapTitle == "" {
		mapTitle = fmt.Sprintf("GEO%d／0x%02X", a.state.GeoMapSet, a.state.GeoMapBlock)
	}
	drawFittedText(screen, fmt.Sprintf(a.state.LocalizedText("ui_guide_position", "%s Position (%d,%d) Facing %d"), mapTitle, x, y, direction), a.compactFace, int(float64(width)*0.10), int(float64(height)*0.22), int(float64(width)*0.80), color.RGBA{92, 220, 255, 255})
	a.drawGuideGrid(screen, width, height, definition)
	drawFittedText(screen, a.state.LocalizedText("ui_guide_footer", "V: Explore / Full  F3 / Esc: Close"), a.compactFace, int(float64(width)*0.10), int(float64(height)*0.88), int(float64(width)*0.80), color.RGBA{232, 238, 255, 255})
}

func (a *app) drawSpoilerWarning(screen *ebiten.Image, width, height int) {
	a.drawOverlayPanel(screen, width, height, a.state.LocalizedText("ui_spoiler_title", "Spoiler Warning"))
	drawWrappedText(screen, a.state.LocalizedText("ui_spoiler_body", "The full guide reveals unseen events, exits, and notes.\n\nEnter / Y: Show full guide\nEsc: Keep exploration mode"), a.compactFace, int(float64(width)*0.12), int(float64(height)*0.30), 38, 28, 8, color.RGBA{255, 220, 92, 255})
}
