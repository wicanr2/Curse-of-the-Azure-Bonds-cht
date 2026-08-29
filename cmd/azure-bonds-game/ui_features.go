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
	Schema          string              `json:"schema"`
	Theme           string              `json:"theme"`
	Language        string              `json:"language"`
	Width           int                 `json:"width"`
	Height          int                 `json:"height"`
	SpoilerWarning  bool                `json:"spoiler_warning_acknowledged"`
	Explored        map[string][]string `json:"explored,omitempty"`
	FrameStyle      string              `json:"frame_style,omitempty"`
	OuterBorderPX   int                 `json:"outer_border_px,omitempty"`
	InnerBorderPX   int                 `json:"inner_border_px,omitempty"`
	ReadingTextPX   int                 `json:"reading_text_px,omitempty"`
	InterfaceTextPX int                 `json:"interface_text_px,omitempty"`
	ReadingBold     *bool               `json:"reading_bold,omitempty"`
	InterfaceBold   *bool               `json:"interface_bold,omitempty"`
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
	settingsOpen      bool
	settingsRow       int
	settingsDraft     uiSettings
}

func defaultUISettingsPath() string {
	root, err := os.UserConfigDir()
	if err != nil || root == "" {
		return "ui-settings.json"
	}
	return filepath.Join(root, "curse-of-the-azure-bonds-remake", "ui-settings.json")
}

func defaultUISettings() uiSettings {
	readingBold, interfaceBold := true, true
	return uiSettings{Schema: "coab-ui-settings/1", Theme: "modern-a6", Language: "zh-TW", Width: 640, Height: 480, Explored: map[string][]string{}, FrameStyle: "A", OuterBorderPX: 10, InnerBorderPX: 8, ReadingTextPX: 24, InterfaceTextPX: 16, ReadingBold: &readingBold, InterfaceBold: &interfaceBold}
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
	normalizeAppearanceSettings(&settings)
	return settings
}

func normalizeAppearanceSettings(settings *uiSettings) {
	defaults := defaultUISettings()
	if settings.FrameStyle != "A" && settings.FrameStyle != "B" && settings.FrameStyle != "C" {
		settings.FrameStyle = defaults.FrameStyle
	}
	if settings.OuterBorderPX < 4 || settings.OuterBorderPX > 20 || settings.OuterBorderPX%2 != 0 {
		settings.OuterBorderPX = defaults.OuterBorderPX
	}
	if settings.InnerBorderPX < 1 || settings.InnerBorderPX > 12 {
		settings.InnerBorderPX = defaults.InnerBorderPX
	}
	if settings.ReadingTextPX < 12 || settings.ReadingTextPX > 36 {
		settings.ReadingTextPX = defaults.ReadingTextPX
	}
	if settings.InterfaceTextPX < 12 || settings.InterfaceTextPX > 36 {
		settings.InterfaceTextPX = defaults.InterfaceTextPX
	}
	if settings.ReadingBold == nil {
		settings.ReadingBold = defaults.ReadingBold
	}
	if settings.InterfaceBold == nil {
		settings.InterfaceBold = defaults.InterfaceBold
	}
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
		case a.ui.settingsOpen:
			a.ui.settingsOpen = false
			return true, nil
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
	if a.ui.settingsOpen {
		if a.justPressed(ebiten.KeyF7) {
			a.ui.settingsOpen = false
			return true, nil
		}
		if a.justPressed(ebiten.KeyArrowUp) {
			a.ui.settingsRow = (a.ui.settingsRow + 6) % 7
		}
		if a.justPressed(ebiten.KeyArrowDown) {
			a.ui.settingsRow = (a.ui.settingsRow + 1) % 7
		}
		if a.justPressed(ebiten.KeyArrowLeft) {
			adjustAppearanceDraft(&a.ui.settingsDraft, a.ui.settingsRow, -1)
		}
		if a.justPressed(ebiten.KeyArrowRight) {
			adjustAppearanceDraft(&a.ui.settingsDraft, a.ui.settingsRow, 1)
		}
		if a.justPressed(ebiten.KeyR) {
			resetAppearanceDraft(&a.ui.settingsDraft)
		}
		if a.justPressed(ebiten.KeyEnter) {
			a.ui.settings = a.ui.settingsDraft
			a.applyTypographySettings()
			a.persistUISettings()
			a.ui.settingsOpen = false
		}
		return true, nil
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
	if a.justPressed(ebiten.KeyF7) {
		a.ui.settingsDraft = a.ui.settings
		a.ui.settingsRow = 0
		a.ui.settingsOpen = true
		a.ui.helpOpen = false
		a.ui.guideOpen = false
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

func adjustAppearanceDraft(settings *uiSettings, row, direction int) {
	switch row {
	case 0:
		styles := []string{"A", "B", "C"}
		index := 0
		for i, style := range styles {
			if style == settings.FrameStyle {
				index = i
				break
			}
		}
		settings.FrameStyle = styles[(index+direction+len(styles))%len(styles)]
	case 1:
		settings.OuterBorderPX = clampInt(settings.OuterBorderPX+direction*2, 4, 20)
	case 2:
		settings.InnerBorderPX = clampInt(settings.InnerBorderPX+direction, 1, 12)
	case 3:
		settings.ReadingTextPX = clampInt(settings.ReadingTextPX+direction, 12, 36)
	case 4:
		value := settings.ReadingBold == nil || !*settings.ReadingBold
		settings.ReadingBold = &value
	case 5:
		settings.InterfaceTextPX = clampInt(settings.InterfaceTextPX+direction, 12, 36)
	case 6:
		value := settings.InterfaceBold == nil || !*settings.InterfaceBold
		settings.InterfaceBold = &value
	}
}

func clampInt(value, minimum, maximum int) int {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func resetAppearanceDraft(settings *uiSettings) {
	defaults := defaultUISettings()
	settings.FrameStyle, settings.OuterBorderPX, settings.InnerBorderPX = defaults.FrameStyle, defaults.OuterBorderPX, defaults.InnerBorderPX
	settings.ReadingTextPX, settings.InterfaceTextPX = defaults.ReadingTextPX, defaults.InterfaceTextPX
	readingBold, interfaceBold := true, true
	settings.ReadingBold, settings.InterfaceBold = &readingBold, &interfaceBold
}

func (a *app) drawGlobalUI(screen *ebiten.Image) {
	w, h := a.ui.logicalSize()
	if a.ui.settings.Theme == "modern-a6" {
		a.drawModernA6Content(screen, w, h)
		customFrame := a.ui.settings.OuterBorderPX != 10 || a.ui.settings.InnerBorderPX != 8
		switch a.state.Mode {
		case game.ModeCombat:
			if customFrame {
				drawConfigurableCombatFrame(screen, w, h, a.ui.settings)
			} else if !a.drawModernA6CombatFrame(screen, w, h) {
				drawA6CombatFrame(screen, w, h)
			}
		case game.ModeWilderness, game.ModeEvent, game.ModeMap, game.ModePlace, game.ModeJournal, game.ModeDungeon:
			// 世界地圖的上半部是單一 608×240 BIGPIC；不得疊入冒險畫面的
			// scene 內框與左右分隔。下方選單仍由 drawOverlandMap 自己分區。
			if a.state.Mode == game.ModeWilderness && a.state.Message == "" {
				if customFrame {
					drawConfigurableOuterFrame(screen, w, h, a.ui.settings)
				} else if !a.drawModernA6OuterFrame(screen, w, h) {
					drawA6OuterFrame(screen, w, h)
				}
			} else if customFrame {
				drawConfigurableAdventureFrame(screen, w, h, a.ui.settings)
			} else if !a.drawModernA6AdventureFrame(screen, w, h) {
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
	if a.ui.settingsOpen {
		a.drawSettingsOverlay(screen, w, h)
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
	// 使用者收斂後的 10px 精修石框；內緣是較細的雙層亮金雕線。
	line(0, 0, 640, 10, stone)
	line(0, 470, 640, 10, shadow)
	line(0, 0, 10, 480, stone)
	line(630, 0, 10, 480, shadow)
	line(2, 2, 636, 2, light)
	line(10, 10, 620, 2, goldLight)
	line(10, 12, 620, 1, gold)
	line(10, 13, 620, 1, goldShadow)
	line(10, 466, 620, 1, goldLight)
	line(10, 467, 620, 2, gold)
	line(10, 469, 620, 1, goldShadow)
	line(10, 10, 2, 460, goldLight)
	line(12, 10, 1, 460, gold)
	line(13, 10, 1, 460, goldShadow)
	line(626, 10, 1, 460, goldLight)
	line(627, 10, 2, 460, gold)
	line(629, 10, 1, 460, goldShadow)
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
	for _, point := range [][2]float64{{10, 10}, {626, 10}, {10, 466}, {626, 466}, {358, 358}} {
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

const guideOverlayBorder = 8

func (a *app) drawGuideOverlayFrame(screen *ebiten.Image, width, height int) {
	x, y := float64(width)*0.06, float64(height)*0.08
	w, h := float64(width)*0.88, float64(height)*0.84
	border := guideOverlayBorder
	style := "A"
	if a.ui.settings.Theme == "modern-a6" {
		border, style = a.ui.settings.InnerBorderPX, a.ui.settings.FrameStyle
	}
	bx, by := float64(width)*float64(border)/640, float64(height)*float64(border)/480
	palette := paletteForFrameStyle(style)
	stone, light := palette.stone, palette.light
	gold := color.RGBA{255, 213, 45, 255}
	shadow := palette.shadow
	ebitenutil.DrawRect(screen, x, y, w, by, light)
	ebitenutil.DrawRect(screen, x, y+h-by, w, by, shadow)
	ebitenutil.DrawRect(screen, x, y, bx, h, stone)
	ebitenutil.DrawRect(screen, x+w-bx, y, bx, h, shadow)
	ebitenutil.DrawRect(screen, x+bx, y+by-1, w-2*bx, 1, gold)
	ebitenutil.DrawRect(screen, x+bx, y+h-by, w-2*bx, 1, gold)
	ebitenutil.DrawRect(screen, x+bx-1, y+by, 1, h-2*by, gold)
	ebitenutil.DrawRect(screen, x+w-bx, y+by, 1, h-2*by, gold)
}

func (a *app) drawHelpOverlay(screen *ebiten.Image, width, height int) {
	a.drawOverlayPanel(screen, width, height, a.state.LocalizedText("ui_help_title", "Help / Controls"))
	text := a.state.LocalizedText("ui_help_body", "F1 Help\nF2 Theme\nF3 Guide Map\nF4 Resolution\nF5 Save  F9 Load\nF6 Language\nF7 Settings\nF10 Save and quit\nEsc Cancel / Back")
	drawWrappedText(screen, text, a.compactFace, int(float64(width)*0.10), int(float64(height)*0.24), 42, 24, 10, color.RGBA{232, 238, 255, 255})
}

func (a *app) drawSettingsOverlay(screen *ebiten.Image, width, height int) {
	a.drawOverlayPanel(screen, width, height, a.state.LocalizedText("ui_settings_title", "Appearance Settings"))
	draft := a.ui.settingsDraft
	boolText := func(value *bool) string {
		if value != nil && *value {
			return a.state.LocalizedText("ui_setting_on", "On")
		}
		return a.state.LocalizedText("ui_setting_off", "Off")
	}
	rows := []string{
		fmt.Sprintf(a.state.LocalizedText("ui_setting_frame_style", "Hand-drawn frame: %s"), draft.FrameStyle),
		fmt.Sprintf(a.state.LocalizedText("ui_setting_outer_border", "Outer border: %dpx"), draft.OuterBorderPX),
		fmt.Sprintf(a.state.LocalizedText("ui_setting_inner_border", "Inner border: %dpx"), draft.InnerBorderPX),
		fmt.Sprintf(a.state.LocalizedText("ui_setting_reading_size", "Reading text: %dpx"), draft.ReadingTextPX),
		fmt.Sprintf(a.state.LocalizedText("ui_setting_reading_bold", "Reading bold: %s"), boolText(draft.ReadingBold)),
		fmt.Sprintf(a.state.LocalizedText("ui_setting_interface_size", "Interface text: %dpx"), draft.InterfaceTextPX),
		fmt.Sprintf(a.state.LocalizedText("ui_setting_interface_bold", "Interface bold: %s"), boolText(draft.InterfaceBold)),
	}
	for index, row := range rows {
		prefix := "  "
		ink := color.RGBA{232, 238, 255, 255}
		if index == a.ui.settingsRow {
			prefix, ink = "> ", color.RGBA{92, 220, 255, 255}
		}
		drawFittedText(screen, prefix+row, a.compactFace, int(float64(width)*0.12), int(float64(height)*0.25)+index*34, int(float64(width)*0.76), ink)
	}
	footer := a.state.LocalizedText("ui_settings_footer", "Arrows: Select / Adjust   Enter: Apply   R: Defaults   Esc: Cancel")
	drawFittedText(screen, footer, a.compactFace, int(float64(width)*0.09), int(float64(height)*0.88), int(float64(width)*0.82), color.RGBA{255, 220, 92, 255})
}

func (a *app) drawGuideOverlay(screen *ebiten.Image, width, height int) {
	mode := a.state.LocalizedText("ui_guide_explore", "Exploration Information")
	if a.ui.guideFull {
		mode = a.state.LocalizedText("ui_guide_full", "Full Guide (spoilers)")
	}
	a.drawOverlayPanel(screen, width, height, fmt.Sprintf(a.state.LocalizedText("ui_guide_title", "Current Guide Map | %s"), mode))
	a.drawGuideOverlayFrame(screen, width, height)
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
