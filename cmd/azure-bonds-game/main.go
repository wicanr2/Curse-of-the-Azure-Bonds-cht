package main

import (
	"archive/zip"
	"flag"
	"image/color"
	"io"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/game"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
)

const (
	logicalWidth  = 640
	logicalHeight = 400
)

type app struct {
	state game.State
	face  font.Face
}

func (a *app) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		switch a.state.Mode {
		case game.ModeTitle:
			return a.state.Apply(game.ActionStart)
		case game.ModeWilderness:
			return a.state.Apply(game.ActionEnterCity)
		}
	}
	if a.state.Mode == game.ModeWilderness && inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		return a.state.Apply(game.ActionJourneyOn)
	}
	return nil
}

func (a *app) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{12, 18, 42, 255})
	white := color.RGBA{232, 238, 255, 255}
	cyan := color.RGBA{92, 220, 255, 255}
	text.Draw(screen, a.state.Title, a.face, 32, 52, cyan)
	text.Draw(screen, a.state.Prompt, a.face, 32, 130, white)
	if a.state.Mode == game.ModeWilderness {
		for index, choice := range a.state.Choices {
			text.Draw(screen, choice, a.face, 56, 220+index*40, white)
		}
		text.Draw(screen, "Enter：選擇", a.face, 56, 330, cyan)
	}
	if a.state.Mode == game.ModeEvent {
		text.Draw(screen, a.state.Message, a.face, 56, 220, cyan)
		text.Draw(screen, "Enter：繼續", a.face, 56, 330, white)
	}
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
	flag.Parse()
	data, err := os.ReadFile(*localePath)
	if err != nil {
		log.Fatal(err)
	}
	catalog, err := locale.Load(data)
	if err != nil {
		log.Fatal(err)
	}
	imageData, err := zipMember(*imagePath, "ECL1.DAX")
	if err != nil {
		log.Fatal(err)
	}
	blocks, err := dax.Parse(imageData)
	if err != nil || len(blocks) == 0 {
		log.Fatalf("ECL1.DAX: %v", err)
	}
	ebiten.SetWindowSize(logicalWidth, logicalHeight)
	ebiten.SetWindowTitle(catalog.Text("title", "Curse of the Azure Bonds"))
	if err := ebiten.RunGame(&app{state: game.NewStateFromECL(catalog, blocks[0].Data), face: loadFace(*fontPath)}); err != nil {
		log.Fatal(err)
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
