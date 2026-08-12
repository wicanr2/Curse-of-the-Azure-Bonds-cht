// Package screenshotmanifest validates versioned screenshot evidence without
// making claims about renderer fidelity.
package screenshotmanifest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
)

var knownEvidenceLevels = map[string]bool{
	"exact":                              true,
	"nearby":                             true,
	"material-exact/layout-reconstructed": true,
	"layout-only":                        true,
	"hypothesis":                         true,
	"unknown":                            true,
}

type Manifest struct {
	Version     int          `json:"version"`
	Canvas      Canvas       `json:"canvas"`
	Screenshots []Screenshot `json:"screenshots"`
	Planned     []Planned    `json:"planned"`
}

type Canvas struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Screenshot struct {
	ID             string `json:"id"`
	File           string `json:"file"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	SHA256         string `json:"sha256"`
	GenerationMode string `json:"generation_mode"`
	EvidenceLevel  string `json:"evidence_level"`
	Spec           string `json:"spec"`
}

type Planned struct {
	ID            string   `json:"id"`
	Status        string   `json:"status"`
	Description   string   `json:"description"`
	NormalPath    string   `json:"normal_path"`
	EvidenceLevel string   `json:"evidence_level"`
	Specs         []string `json:"specs"`
}

func Load(path string) (Manifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("read manifest %q: %w", path, err)
	}
	var manifest Manifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode manifest %q: %w", path, err)
	}
	return manifest, nil
}

func Validate(manifest Manifest, root string) error {
	if manifest.Version < 1 {
		return fmt.Errorf("version must be positive")
	}
	if manifest.Canvas.Width <= 0 || manifest.Canvas.Height <= 0 {
		return fmt.Errorf("canvas dimensions must be positive")
	}
	seen := make(map[string]bool)
	for index, shot := range manifest.Screenshots {
		if err := validateScreenshot(index, shot, root, manifest.Canvas, seen); err != nil {
			return err
		}
	}
	for index, planned := range manifest.Planned {
		if planned.ID == "" {
			return fmt.Errorf("planned[%d]: id is required", index)
		}
		if seen[planned.ID] {
			return fmt.Errorf("planned[%d]: duplicate id %q", index, planned.ID)
		}
		seen[planned.ID] = true
		if planned.Status != "planned" {
			return fmt.Errorf("planned[%d] %q: status must be planned", index, planned.ID)
		}
		if planned.Description == "" || planned.NormalPath == "" || len(planned.Specs) == 0 {
			return fmt.Errorf("planned[%d] %q: description, normal_path and specs are required", index, planned.ID)
		}
		if !knownEvidenceLevels[planned.EvidenceLevel] {
			return fmt.Errorf("planned[%d] %q: unknown evidence level %q", index, planned.ID, planned.EvidenceLevel)
		}
	}
	return nil
}

func validateScreenshot(index int, shot Screenshot, root string, canvas Canvas, seen map[string]bool) error {
	if shot.ID == "" || shot.File == "" || shot.GenerationMode == "" || shot.Spec == "" {
		return fmt.Errorf("screenshots[%d]: id, file, generation_mode and spec are required", index)
	}
	if seen[shot.ID] {
		return fmt.Errorf("screenshots[%d]: duplicate id %q", index, shot.ID)
	}
	seen[shot.ID] = true
	if !knownEvidenceLevels[shot.EvidenceLevel] {
		return fmt.Errorf("screenshots[%d] %q: unknown evidence level %q", index, shot.ID, shot.EvidenceLevel)
	}
	if shot.Width != canvas.Width || shot.Height != canvas.Height {
		return fmt.Errorf("screenshots[%d] %q: declared dimensions %dx%d do not match canvas %dx%d", index, shot.ID, shot.Width, shot.Height, canvas.Width, canvas.Height)
	}
	if len(shot.SHA256) != sha256.Size*2 {
		return fmt.Errorf("screenshots[%d] %q: sha256 must be 64 hex characters", index, shot.ID)
	}
	if _, err := hex.DecodeString(shot.SHA256); err != nil {
		return fmt.Errorf("screenshots[%d] %q: invalid sha256: %w", index, shot.ID, err)
	}
	path := filepath.Join(root, filepath.Clean(shot.File))
	rootClean, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("screenshots[%d] %q: resolve root: %w", index, shot.ID, err)
	}
	pathClean, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("screenshots[%d] %q: resolve file: %w", index, shot.ID, err)
	}
	if pathClean != rootClean && (len(pathClean) <= len(rootClean) || pathClean[:len(rootClean)] != rootClean || rune(pathClean[len(rootClean)]) != filepath.Separator) {
		return fmt.Errorf("screenshots[%d] %q: file escapes repository root", index, shot.ID)
	}
	file, err := os.Open(pathClean)
	if err != nil {
		return fmt.Errorf("screenshots[%d] %q: missing file %q: %w", index, shot.ID, shot.File, err)
	}
	defer file.Close()
	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return fmt.Errorf("screenshots[%d] %q: decode image: %w", index, shot.ID, err)
	}
	if config.Width != shot.Width || config.Height != shot.Height {
		return fmt.Errorf("screenshots[%d] %q: actual dimensions %dx%d differ from manifest %dx%d", index, shot.ID, config.Width, config.Height, shot.Width, shot.Height)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("screenshots[%d] %q: rewind file: %w", index, shot.ID, err)
	}
	hash := sha256.New()
	if _, err := readInto(hash, file); err != nil {
		return fmt.Errorf("screenshots[%d] %q: hash file: %w", index, shot.ID, err)
	}
	actual := hex.EncodeToString(hash.Sum(nil))
	if actual != shot.SHA256 {
		return fmt.Errorf("screenshots[%d] %q: sha256 drift: manifest=%s actual=%s", index, shot.ID, shot.SHA256, actual)
	}
	return nil
}

type reader interface { Read([]byte) (int, error) }

func readInto(hash interface{ Write([]byte) (int, error) }, input reader) (int64, error) {
	buffer := make([]byte, 32*1024)
	var total int64
	for {
		n, err := input.Read(buffer)
		if n > 0 {
			if _, writeErr := hash.Write(buffer[:n]); writeErr != nil {
				return total, writeErr
			}
			total += int64(n)
		}
		if err != nil {
			if err == io.EOF {
				return total, nil
			}
			return total, err
		}
	}
}
