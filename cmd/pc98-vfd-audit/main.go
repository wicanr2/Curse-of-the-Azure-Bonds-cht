// Command pc98-vfd-audit inventories a user-supplied VFD1.00 PC-98 disk
// without modifying or extracting the original media.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98vfd"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
)

type missingSector struct {
	Cylinder int `json:"cylinder"`
	Head     int `json:"head"`
	Sector   int `json:"sector"`
	Size     int `json:"size"`
}

type report struct {
	Path           string          `json:"path"`
	SHA256         string          `json:"sha256"`
	Size           int             `json:"size"`
	Geometry       string          `json:"geometry"`
	SectorCount    int             `json:"sector_count"`
	MissingSectors []missingSector `json:"missing_sectors"`
}

func main() {
	cylinders := flag.Int("cylinders", 77, tooltext.Text("pc98_vfd_audit.cylinders"))
	heads := flag.Int("heads", 2, tooltext.Text("pc98_vfd_audit.heads"))
	sectors := flag.Int("sectors", 8, tooltext.Text("pc98_vfd_audit.sectors"))
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, tooltext.Text("pc98_vfd_audit.usage"))
		os.Exit(2)
	}

	path := flag.Arg(0)
	data, err := os.ReadFile(path)
	if err != nil {
		fatal(err)
	}
	image, err := pc98vfd.Parse(data, *cylinders, *heads, *sectors)
	if err != nil {
		fatal(err)
	}
	sum := sha256.Sum256(data)
	result := report{
		Path:        path,
		SHA256:      hex.EncodeToString(sum[:]),
		Size:        len(data),
		Geometry:    fmt.Sprintf("%dx%dx%d", *cylinders, *heads, *sectors),
		SectorCount: len(image.Sectors),
	}
	for _, sector := range image.MissingSectors() {
		result.MissingSectors = append(result.MissingSectors, missingSector{
			Cylinder: int(sector.Cylinder),
			Head:     int(sector.Head),
			Sector:   int(sector.Number),
			Size:     sector.Size(),
		})
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
