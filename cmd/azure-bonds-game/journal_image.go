package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/game"
)

// 手札的地圖與插圖從 `assets/journal/` 讀，對應表在 gamepack。這裡不硬寫任何
// 檔名或條目編號——哪一則有圖是資料決定的。
const journalImageDir = "assets/journal"

// 彈窗的內容區。整張畫面是 640x480，留邊之後給圖的空間。
const (
	journalPanelX      = 24
	journalPanelY      = 40
	journalPanelWidth  = 592
	journalPanelHeight = 400
	journalImageTop    = journalPanelY + 34
	journalImageBottom = journalPanelY + journalPanelHeight - 30
	journalPanStep     = 48
)

// currentJournalImage 回傳目前這一頁手札的圖；沒有圖回 nil。
// 找不到檔案時也回 nil 並記負向快取，不讓缺檔變成每一幀都嘗試開檔。
func (a *app) currentJournalImage() *ebiten.Image {
	messageID := a.currentJournalMessageID()
	if messageID == "" {
		return nil
	}
	if a.journalImages == nil {
		a.journalImages = map[string]*ebiten.Image{}
	}
	if cached, seen := a.journalImages[messageID]; seen {
		return cached
	}
	record, ok := gamepack.JournalImageFor(messageID)
	if !ok {
		a.journalImages[messageID] = nil
		return nil
	}
	loaded, err := loadJournalImage(filepath.Join(journalImageDir, record.File))
	if err != nil {
		// 缺圖不該讓遊戲停下來：記成「這一則沒有圖」，玩家照樣讀得到文字。
		fmt.Fprintf(os.Stderr, "journal image %s: %v\n", record.File, err)
		a.journalImages[messageID] = nil
		return nil
	}
	a.journalImages[messageID] = loaded
	return loaded
}

func loadJournalImage(path string) (*ebiten.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoded, err := png.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return ebiten.NewImageFromImage(decoded), nil
}

// journalImageCaption 從 game pack 取彈窗標題，跟著目前語系走。
func (a *app) journalImageCaption() string {
	record, ok := gamepack.JournalImageFor(a.state.JournalMessageID())
	if !ok || a.gamePack == nil {
		return ""
	}
	caption, _ := a.gamePack.Text(record.CaptionID, a.state.LocaleLanguage())
	return caption
}

func (a *app) updateJournalImage() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyI) {
		a.journalImageOpen = false
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		a.journalImageZoom = !a.journalImageZoom
		a.journalImageOffsetX, a.journalImageOffsetY = 0, 0
		return nil
	}
	// 縮到整張可見時沒有東西可以平移，方向鍵留給翻頁不會有歧義。
	if !a.journalImageZoom {
		return nil
	}
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyLeft):
		a.journalImageOffsetX -= journalPanStep
	case inpututil.IsKeyJustPressed(ebiten.KeyRight):
		a.journalImageOffsetX += journalPanStep
	case inpututil.IsKeyJustPressed(ebiten.KeyUp):
		a.journalImageOffsetY -= journalPanStep
	case inpututil.IsKeyJustPressed(ebiten.KeyDown):
		a.journalImageOffsetY += journalPanStep
	}
	a.clampJournalPan()
	return nil
}

// clampJournalPan 把平移量夾在「圖還看得到」的範圍內。圖比視窗小的那一軸
// 直接歸零，否則玩家會把圖推出畫面外找不回來。
func (a *app) clampJournalPan() {
	picture := a.currentJournalImage()
	if picture == nil {
		return
	}
	width, height := picture.Bounds().Dx(), picture.Bounds().Dy()
	viewWidth := journalPanelWidth - 16
	viewHeight := journalImageBottom - journalImageTop
	a.journalImageOffsetX = clampPan(a.journalImageOffsetX, width, viewWidth)
	a.journalImageOffsetY = clampPan(a.journalImageOffsetY, height, viewHeight)
}

func clampPan(offset, content, view int) int {
	limit := (content - view) / 2
	if limit <= 0 {
		return 0
	}
	if offset > limit {
		return limit
	}
	if offset < -limit {
		return -limit
	}
	return offset
}

func (a *app) drawJournalImage(screen *ebiten.Image, white, cyan color.Color) {
	picture := a.currentJournalImage()
	if picture == nil {
		return
	}
	// 底色鋪滿整個面板：手札文字在底下，不遮住會看不清圖。
	ebitenutil.DrawRect(screen, journalPanelX, journalPanelY,
		journalPanelWidth, journalPanelHeight, color.RGBA{A: 255})
	ebitenutil.DrawRect(screen, journalPanelX, journalPanelY, journalPanelWidth, 1, white)
	ebitenutil.DrawRect(screen, journalPanelX, journalPanelY+journalPanelHeight-1,
		journalPanelWidth, 1, white)
	ebitenutil.DrawRect(screen, journalPanelX, journalPanelY, 1, journalPanelHeight, white)
	ebitenutil.DrawRect(screen, journalPanelX+journalPanelWidth-1, journalPanelY,
		1, journalPanelHeight, white)
	drawFittedText(screen, a.journalImageCaption(), a.face,
		journalPanelX+8, journalPanelY+26, journalPanelWidth-16, cyan)
	drawFittedText(screen, a.state.PlayerUILabel(game.PlayerUILabelJournalImageControls), a.face,
		journalPanelX+8, journalPanelY+journalPanelHeight-12, journalPanelWidth-16, cyan)

	viewWidth := journalPanelWidth - 16
	viewHeight := journalImageBottom - journalImageTop
	scale := journalImageScale(picture.Bounds().Dx(), picture.Bounds().Dy(),
		viewWidth, viewHeight, a.journalImageZoom)
	drawWidth := float64(picture.Bounds().Dx()) * scale
	drawHeight := float64(picture.Bounds().Dy()) * scale

	options := &ebiten.DrawImageOptions{}
	options.Filter = ebiten.FilterLinear
	options.GeoM.Scale(scale, scale)
	options.GeoM.Translate(
		float64(journalPanelX+8)+(float64(viewWidth)-drawWidth)/2-float64(a.journalImageOffsetX),
		float64(journalImageTop)+(float64(viewHeight)-drawHeight)/2-float64(a.journalImageOffsetY),
	)
	view := screen.SubImage(image.Rect(
		journalPanelX+8, journalImageTop,
		journalPanelX+8+viewWidth, journalImageBottom)).(*ebiten.Image)
	view.DrawImage(picture, options)
}

// journalImageScale 決定顯示倍率。預設把整張縮進視窗；按 Z 之後改成 1:1，
// 讓地圖上的房間名讀得清楚，超出的部分靠方向鍵平移。
//
// 縮的時候不放大（上限 1）：把 216x194 的小插圖拉滿整個面板只會糊掉。
func journalImageScale(width, height, viewWidth, viewHeight int, zoom bool) float64 {
	if zoom {
		return 1
	}
	scale := float64(viewWidth) / float64(width)
	if fit := float64(viewHeight) / float64(height); fit < scale {
		scale = fit
	}
	if scale > 1 {
		scale = 1
	}
	return scale
}

// currentJournalMessageID 由**前端自己的顯示頁碼**換算出目前看的是第幾則手札。
//
// ⚠ 不要用 `state.JournalMessageID()`：那一支讀的是 `State.JournalPage`，
// 而前端翻頁走的是 `a.journalDisplayPage`，**從來沒有推進過 `State.JournalPage`**。
// 用它的話不論翻到哪一頁，按 `I` 跳出來的都是第一則的圖（spec 1189）。
func (a *app) currentJournalMessageID() string {
	_, sources := journalDisplayPagesWithSources(
		a.state.JournalPages, a.state.JournalText, a.face, 22*faceCellWidth(a.face), 7)
	if a.journalDisplayPage < 0 || a.journalDisplayPage >= len(sources) {
		return ""
	}
	return a.state.JournalMessageIDAt(sources[a.journalDisplayPage])
}
