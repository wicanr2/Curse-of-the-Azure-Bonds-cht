package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// modernA6Assets is a deliberately separate asset group. A theme switch must
// never mutate or overwrite the original-derived PNG catalog.
type modernA6Assets struct {
	pictures       map[string]*ebiten.Image
	sprites        map[string]*ebiten.Image
	tiles          map[string]*ebiten.Image
	animations     map[string][]combatAnimation
	combat         map[string]*ebiten.Image
	symbols        map[string]*ebiten.Image
	sky            map[string]*ebiten.Image
	adventureFrame *ebiten.Image
	combatFrame    *ebiten.Image
}

func loadModernA6Assets(root string) (*modernA6Assets, error) {
	result := &modernA6Assets{
		pictures:   make(map[string]*ebiten.Image),
		sprites:    make(map[string]*ebiten.Image),
		tiles:      make(map[string]*ebiten.Image),
		animations: make(map[string][]combatAnimation),
		combat:     make(map[string]*ebiten.Image),
		symbols:    make(map[string]*ebiten.Image),
		sky:        make(map[string]*ebiten.Image),
	}
	for directory, destination := range map[string]map[string]*ebiten.Image{
		"combat": result.combat, "symbols": result.symbols, "sky": result.sky,
	} {
		paths, err := filepath.Glob(filepath.Join(root, directory, "*.png"))
		if err != nil {
			return nil, err
		}
		for _, path := range paths {
			picture, err := loadModernPNG(path)
			if err != nil {
				return nil, err
			}
			if directory == "symbols" {
				picture, err = loadModernChromaPNG(path)
				if err != nil {
					return nil, err
				}
			}
			destination[filepath.Base(path)] = picture
		}
	}
	for directory, destination := range map[string]map[string]*ebiten.Image{
		"pictures": result.pictures,
		"sprites":  result.sprites,
		"tiles":    result.tiles,
	} {
		paths, err := filepath.Glob(filepath.Join(root, directory, "*.png"))
		if err != nil {
			return nil, err
		}
		for _, path := range paths {
			picture, err := loadModernPNG(path)
			if err != nil {
				return nil, err
			}
			destination[filepath.Base(path)] = picture
		}
	}
	animationData, err := os.ReadFile(filepath.Join("assets", "sprites", "animation.json"))
	if err != nil {
		return nil, err
	}
	var records []struct {
		Name  string `json:"name"`
		Delay uint32 `json:"delay"`
		X     int16  `json:"x"`
		Y     int16  `json:"y"`
	}
	if err := json.Unmarshal(animationData, &records); err != nil {
		return nil, fmt.Errorf("parse modern A6 animation manifest: %w", err)
	}
	for _, record := range records {
		picture := result.sprites[record.Name]
		if picture == nil {
			continue
		}
		marker := strings.Index(record.Name, "-frame-")
		if marker < 0 {
			return nil, fmt.Errorf("modern A6 animation asset %q has no frame marker", record.Name)
		}
		key := record.Name[:marker]
		result.animations[key] = append(result.animations[key], combatAnimation{
			image: picture, delay: record.Delay, x: record.X, y: record.Y,
		})
	}
	frame, err := loadModernPNG(filepath.Join(root, "ui", "adventure-frame.png"))
	if err != nil {
		return nil, err
	}
	result.adventureFrame = frame
	combatFrame, err := loadModernPNG(filepath.Join(root, "ui", "combat-frame.png"))
	if err != nil {
		return nil, err
	}
	result.combatFrame = combatFrame
	return result, nil
}

func loadModernChromaPNG(path string) (*ebiten.Image, error) {
	handle, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	decoded, decodeErr := png.Decode(handle)
	closeErr := handle.Close()
	if decodeErr != nil {
		return nil, fmt.Errorf("decode modern A6 chroma asset %s: %w", path, decodeErr)
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return ebiten.NewImageFromImage(chromaKeyTopLeft(decoded)), nil
}

func loadModernPNG(path string) (*ebiten.Image, error) {
	handle, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	decoded, decodeErr := png.Decode(handle)
	closeErr := handle.Close()
	if decodeErr != nil {
		return nil, fmt.Errorf("decode modern A6 asset %s: %w", path, decodeErr)
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return ebiten.NewImageFromImage(decoded), nil
}

func (a *app) modernScenePicture() *ebiten.Image {
	if a.modernA6 == nil || a.ui.settings.Theme != "modern-a6" || !a.state.SceneCharacterRequested {
		return nil
	}
	key := modernScenePictureKey(a.state.Area.GameArea, a.state.SceneHeadBlock, a.state.SceneBodyBlock)
	if picture := a.modernA6.pictures[key]; picture != nil {
		return picture
	}
	return nil
}

func modernScenePictureKey(area, head, body uint8) string {
	// These aliases are restricted to source composites proven byte-identical
	// by SHA-256. They prevent duplicate redraws from drifting apart.
	if head == 0x03 && body == 0x03 && (area == 2 || area == 3 || area == 5) {
		return "character-area-5-head-03-body-03.png"
	}
	if area == 6 && head == 0x41 && body == 0x41 {
		return "character-area-2-head-41-body-41.png"
	}
	if area == 6 && head == 0x46 && body == 0x46 {
		return "character-area-4-head-46-body-46.png"
	}
	return fmt.Sprintf("character-area-%d-head-%02X-body-%02X.png", area, head, body)
}

func (a *app) drawModernA6Content(screen *ebiten.Image, width, height int) {
	picture := a.modernScenePicture()
	if picture == nil {
		return
	}
	sx, sy := float64(width)/640, float64(height)/480
	destination := image.Rect(int(51*sx), int(39*sy), int(234*sx), int(230*sy))
	drawImageCoverFiltered(screen, picture, destination, ebiten.FilterLinear)
}

func drawImageCoverFiltered(screen, source *ebiten.Image, destination image.Rectangle, filter ebiten.Filter) {
	if source == nil || destination.Empty() {
		return
	}
	scale, translateX, translateY := imageCoverTransform(
		source.Bounds().Dx(), source.Bounds().Dy(), destination)
	target := screen.SubImage(destination).(*ebiten.Image)
	op := &ebiten.DrawImageOptions{}
	op.Filter = filter
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(translateX, translateY)
	target.DrawImage(source, op)
}

func (a *app) drawModernA6AdventureFrame(screen *ebiten.Image, width, height int) bool {
	if a.modernA6 == nil || a.modernA6.adventureFrame == nil {
		return false
	}
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear
	op.GeoM.Scale(float64(width)/640, float64(height)/480)
	screen.DrawImage(a.modernA6.adventureFrame, op)
	return true
}

func (a *app) drawModernA6CombatFrame(screen *ebiten.Image, width, height int) bool {
	if a.modernA6 == nil || a.modernA6.combatFrame == nil {
		return false
	}
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear
	op.GeoM.Scale(float64(width)/640, float64(height)/480)
	screen.DrawImage(a.modernA6.combatFrame, op)
	return true
}

func (a *app) modernCombatSprite(key string) *ebiten.Image {
	if a.modernA6 == nil || a.ui.settings.Theme != "modern-a6" {
		return nil
	}
	return a.modernA6.sprites[key]
}

func (a *app) modernAnimation(key string) ([]combatAnimation, bool) {
	if a.modernA6 == nil || a.ui.settings.Theme != "modern-a6" {
		return nil, false
	}
	frames := a.modernA6.animations[key]
	return frames, len(frames) > 0
}

func (a *app) modernRuntimeImage(family, key string) *ebiten.Image {
	if a.modernA6 == nil || a.ui.settings.Theme != "modern-a6" {
		return nil
	}
	switch family {
	case "combat":
		return a.modernA6.combat[key]
	case "symbols":
		return a.modernA6.symbols[key]
	case "sky":
		return a.modernA6.sky[key]
	default:
		return nil
	}
}

func runtimeImageKey(file string, block uint8, item int) string {
	base := strings.ToLower(strings.TrimSuffix(filepath.Base(file), filepath.Ext(file)))
	return fmt.Sprintf("%s-block-%02X-item-%03d.png", base, block, item)
}

func (a *app) displayTile(index int) *ebiten.Image {
	if a.modernA6 != nil && a.ui.settings.Theme == "modern-a6" {
		if tile := a.modernA6.tiles[modernAreaTileKey(index)]; tile != nil {
			return tile
		}
	}
	if index < 0 || index >= len(a.tileImages) {
		return nil
	}
	return a.tileImages[index]
}

func modernAreaTileKey(index int) string {
	if index >= 0 && index < 26 {
		return fmt.Sprintf("tiles-block-01-item-%03d.png", index)
	}
	if index >= 26 && index < 48 {
		return fmt.Sprintf("tiles-block-02-item-%03d.png", index-26)
	}
	return ""
}
