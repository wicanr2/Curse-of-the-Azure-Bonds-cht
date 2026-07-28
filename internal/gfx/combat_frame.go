package gfx

import (
	"image"
	"image/color"
)

const (
	CombatFrameWidth  = 320
	CombatFrameHeight = 184
)

var (
	frameWhite     = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	frameLightGray = color.RGBA{R: 173, G: 173, B: 173, A: 255}
	frameDarkGray  = color.RGBA{R: 85, G: 85, B: 85, A: 255}
	frameBlack     = color.RGBA{A: 255}
)

// CombatFrame reconstructs the fixed 320x184 DOS combat chrome as native
// pixels. The transparent panel interiors are filled by the battlefield and
// status renderer before this image is overlaid.
func CombatFrame() *image.RGBA {
	frame := image.NewRGBA(image.Rect(0, 0, CombatFrameWidth, CombatFrameHeight))
	fillFrameRect(frame, image.Rect(0, 0, 320, 8))
	fillFrameRect(frame, image.Rect(0, 0, 8, 184))
	fillFrameRect(frame, image.Rect(176, 0, 184, 184))
	fillFrameRect(frame, image.Rect(312, 0, 320, 184))
	fillFrameRect(frame, image.Rect(0, 176, 320, 184))

	drawHorizontalInnerEdge(frame, 7, 8, 175, -1)
	drawHorizontalInnerEdge(frame, 7, 184, 311, -1)
	drawHorizontalInnerEdge(frame, 176, 8, 175, 1)
	drawHorizontalInnerEdge(frame, 176, 184, 311, 1)
	drawVerticalInnerEdge(frame, 7, 8, 175, -1)
	drawVerticalInnerEdge(frame, 176, 8, 175, 1)
	drawVerticalInnerEdge(frame, 183, 8, 175, -1)
	drawVerticalInnerEdge(frame, 312, 8, 175, 1)

	for _, x := range []int{14, 69, 108, 143, 269} {
		drawTopCrack(frame, x)
	}
	for _, x := range []int{14, 94, 143, 269} {
		drawBottomCrack(frame, x)
	}
	for _, y := range []int{17, 81, 145} {
		drawLeftCrack(frame, y)
		drawRightCrack(frame, y)
	}
	for _, y := range []int{17, 81, 145} {
		drawDividerCrack(frame, y)
	}
	return frame
}

func fillFrameRect(frame *image.RGBA, rect image.Rectangle) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			frame.SetRGBA(x, y, frameWhite)
		}
	}
}

func drawHorizontalInnerEdge(frame *image.RGBA, y, start, end, direction int) {
	for x := start; x <= end; x++ {
		frame.SetRGBA(x, y+direction*3, frameLightGray)
		frame.SetRGBA(x, y+direction*2, frameDarkGray)
		frame.SetRGBA(x, y+direction, frameBlack)
		if x%2 == 0 {
			frame.SetRGBA(x, y, frameLightGray)
		} else {
			frame.SetRGBA(x, y, frameBlack)
		}
	}
}

func drawVerticalInnerEdge(frame *image.RGBA, x, start, end, direction int) {
	for y := start; y <= end; y++ {
		frame.SetRGBA(x+direction*3, y, frameLightGray)
		frame.SetRGBA(x+direction*2, y, frameDarkGray)
		frame.SetRGBA(x+direction, y, frameBlack)
		frame.SetRGBA(x, y, frameBlack)
		if y%2 == 0 {
			frame.SetRGBA(x, y, frameLightGray)
		}
	}
}

func drawTopCrack(frame *image.RGBA, x int) {
	for index, point := range [][2]int{{0, 0}, {1, 1}, {1, 2}, {2, 3}, {2, 4}, {3, 5}} {
		ink := frameBlack
		if index == 5 {
			ink = frameDarkGray
		}
		frame.SetRGBA(x+point[0], point[1], ink)
	}
}

func drawBottomCrack(frame *image.RGBA, x int) {
	for _, point := range [][2]int{{0, 183}, {1, 182}, {1, 181}, {2, 180}, {3, 179}, {3, 178}} {
		frame.SetRGBA(x+point[0], point[1], frameBlack)
	}
}

func drawLeftCrack(frame *image.RGBA, y int) {
	for _, point := range [][2]int{{0, 0}, {1, 1}, {1, 2}, {2, 3}, {2, 4}, {3, 5}} {
		frame.SetRGBA(point[0], y+point[1], frameBlack)
	}
}

func drawRightCrack(frame *image.RGBA, y int) {
	for _, point := range [][2]int{{319, 0}, {318, 1}, {318, 2}, {317, 3}, {317, 4}, {316, 5}} {
		frame.SetRGBA(point[0], y+point[1], frameBlack)
	}
}

func drawDividerCrack(frame *image.RGBA, y int) {
	for _, point := range [][2]int{{176, 0}, {177, 1}, {177, 2}, {178, 3}, {177, 4}, {176, 5}} {
		frame.SetRGBA(point[0], y+point[1], frameBlack)
	}
}
