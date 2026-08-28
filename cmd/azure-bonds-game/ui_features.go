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
	return uiSettings{Schema: "coab-ui-settings/1", Theme: "modern-a6", Width: 640, Height: 480, Explored: map[string][]string{}}
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
		a.state.Message = "無法保存介面設定：" + err.Error()
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
		switch a.state.Mode {
		case game.ModeCombat:
			drawA6CombatFrame(screen, w, h)
		case game.ModeWilderness, game.ModeEvent, game.ModeMap, game.ModePlace, game.ModeJournal, game.ModeDungeon:
			drawA6Frame(screen, w, h)
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
	drawA6OuterFrame(screen, width, height)
	sx, sy := float64(width)/640, float64(height)/480
	stone, shadow := color.RGBA{218, 211, 190, 255}, color.RGBA{88, 78, 61, 255}
	ebitenutil.DrawRect(screen, 360*sx, 8*sy, 6*sx, 352*sy, stone)
	ebitenutil.DrawRect(screen, 365*sx, 8*sy, sx, 352*sy, shadow)
	ebitenutil.DrawRect(screen, 2*sx, 360*sy, 636*sx, 6*sy, stone)
	ebitenutil.DrawRect(screen, 2*sx, 448*sy, 636*sx, 5*sy, stone)
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
	a.drawOverlayPanel(screen, width, height, "Help／按鍵說明")
	text := "F1  關閉 Help\nF2  切換原版／A6 theme\nF3  目前地圖攻略\nF4  切換 640×480／1024×768／1280×960\nF5  存檔　F9  讀檔\nF10 自動存檔後離開\nEsc 取消／返回\nCtrl+S 音效　Ctrl+O 音樂"
	drawWrappedText(screen, text, a.compactFace, int(float64(width)*0.10), int(float64(height)*0.24), 42, 24, 10, color.RGBA{232, 238, 255, 255})
}

func (a *app) drawGuideOverlay(screen *ebiten.Image, width, height int) {
	mode := "探索資訊"
	if a.ui.guideFull {
		mode = "完整攻略（含劇透）"
	}
	a.drawOverlayPanel(screen, width, height, "目前地圖攻略｜"+mode)
	if a.state.Mode != game.ModeDungeon && a.state.Mode != game.ModeMap {
		drawWrappedText(screen, "目前畫面沒有可用的地圖攻略。\n進入地城或區域地圖後再按 F3。", a.compactFace, int(float64(width)*0.12), int(float64(height)*0.32), 36, 28, 6, color.RGBA{232, 238, 255, 255})
		drawFittedText(screen, "F3／Esc：關閉", a.compactFace, int(float64(width)*0.10), int(float64(height)*0.88), int(float64(width)*0.80), color.RGBA{232, 238, 255, 255})
		return
	}
	x, y, direction := a.state.DungeonGeometryView()
	mapKey := guideMapKey(a.state.GeoMapSet, a.state.GeoMapBlock)
	definition, _ := a.guide.Maps[mapKey]
	mapTitle := definition.Title
	if mapTitle == "" {
		mapTitle = fmt.Sprintf("GEO%d／0x%02X", a.state.GeoMapSet, a.state.GeoMapBlock)
	}
	drawFittedText(screen, fmt.Sprintf("%s　位置 (%d,%d) 朝向 %d", mapTitle, x, y, direction), a.compactFace, int(float64(width)*0.10), int(float64(height)*0.22), int(float64(width)*0.80), color.RGBA{92, 220, 255, 255})
	a.drawGuideGrid(screen, width, height, definition)
	drawFittedText(screen, "V：探索／完整　F3／Esc：關閉", a.compactFace, int(float64(width)*0.10), int(float64(height)*0.88), int(float64(width)*0.80), color.RGBA{232, 238, 255, 255})
}

func (a *app) drawSpoilerWarning(screen *ebiten.Image, width, height int) {
	a.drawOverlayPanel(screen, width, height, "劇透警告")
	drawWrappedText(screen, "完整攻略會顯示尚未觸發的事件、出口與說明。\n\nEnter／Y：顯示完整攻略\nEsc：維持探索模式", a.compactFace, int(float64(width)*0.12), int(float64(height)*0.30), 38, 28, 8, color.RGBA{255, 220, 92, 255})
}
