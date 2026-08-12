package screenshotmanifest

import (
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateRejectsMissingFile(t *testing.T) {
	manifest := validManifest(t, "missing.png")
	err := Validate(manifest, t.TempDir())
	containsError(t, err, "missing file")
}

func TestValidateRejectsWrongDimensions(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "shot.png")
	writePNG(t, file, 320, 200)
	manifest := validManifest(t, "shot.png")
	manifest.Screenshots[0].Width = 320
	manifest.Screenshots[0].Height = 200
	manifest.Screenshots[0].SHA256 = hashFile(t, file)
	manifest.Canvas = Canvas{Width: 640, Height: 480}
	err := Validate(manifest, root)
	containsError(t, err, "declared dimensions")
}

func TestValidateRejectsUnknownEvidenceLevel(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "shot.png")
	writePNG(t, file, 640, 480)
	manifest := validManifest(t, "shot.png")
	manifest.Screenshots[0].SHA256 = hashFile(t, file)
	manifest.Screenshots[0].EvidenceLevel = "pixel-perfect-ish"
	err := Validate(manifest, root)
	containsError(t, err, "unknown evidence level")
}

func TestValidateRejectsHashDrift(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "shot.png")
	writePNG(t, file, 640, 480)
	manifest := validManifest(t, "shot.png")
	manifest.Screenshots[0].SHA256 = strings.Repeat("0", 64)
	err := Validate(manifest, root)
	containsError(t, err, "sha256 drift")
}

func TestValidateAcceptsPlannedGap(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "shot.png")
	writePNG(t, file, 640, 480)
	manifest := validManifest(t, "shot.png")
	manifest.Screenshots[0].SHA256 = hashFile(t, file)
	manifest.Planned = []Planned{{ID: "view", Status: "planned", Description: "正常 VIEW 角色資訊頁", NormalPath: "新遊戲 → CAMP → VIEW", EvidenceLevel: "unknown", Specs: []string{"docs/spec/406-dos-gui-draw-contract.md"}}}
	if err := Validate(manifest, root); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func validManifest(t *testing.T, file string) Manifest {
	t.Helper()
	return Manifest{Version: 1, Canvas: Canvas{Width: 640, Height: 480}, Screenshots: []Screenshot{{ID: "shot", File: file, Width: 640, Height: 480, SHA256: strings.Repeat("0", 64), GenerationMode: "docker-xvfb:test", EvidenceLevel: "layout-only", Spec: "docs/spec/549-dos-character-creation-and-screenshot-polish.md"}}}
}

func writePNG(t *testing.T, path string, width, height int) {
	t.Helper()
	imageData := image.NewRGBA(image.Rect(0, 0, width, height))
	imageData.Set(0, 0, color.White)
	file, err := os.Create(path)
	if err != nil { t.Fatal(err) }
	defer file.Close()
	if err := png.Encode(file, imageData); err != nil { t.Fatal(err) }
}

func hashFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil { t.Fatal(err) }
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func containsError(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %v, want substring %q", err, want)
	}
}
