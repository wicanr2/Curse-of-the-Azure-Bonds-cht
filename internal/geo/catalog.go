package geo

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

// MapRef identifies the original GEO DAX set and block ID. The set is the
// chapter/file number (2..6); the block ID is the area/map selector stored in
// the DAX index, not an array position invented by the remake.
type MapRef struct {
	Set     uint8
	BlockID uint8
}

// Catalog contains every decoded GEO block loaded from the original image.
// It is intentionally independent from ECL or area save state so the latter
// can select a map once its current_3DMap_block_id contract is decoded.
type Catalog struct {
	sets map[uint8]map[uint8]Grid
	refs []MapRef
}

func NewCatalog() Catalog { return Catalog{sets: make(map[uint8]map[uint8]Grid)} }

// AddDAX decodes one GEO*.DAX member and preserves its original block IDs.
func (c *Catalog) AddDAX(set uint8, data []byte) error {
	if c.sets == nil {
		c.sets = make(map[uint8]map[uint8]Grid)
	}
	blocks, err := dax.Parse(data)
	if err != nil {
		return fmt.Errorf("parse GEO%d.DAX: %w", set, err)
	}
	if len(blocks) == 0 {
		return fmt.Errorf("GEO%d.DAX contains no blocks", set)
	}
	if c.sets[set] == nil {
		c.sets[set] = make(map[uint8]Grid)
	}
	for _, block := range blocks {
		grid, err := Parse(block.Entry.ID, block.Data)
		if err != nil {
			return fmt.Errorf("GEO%d.DAX block 0x%02X: %w", set, block.Entry.ID, err)
		}
		if _, exists := c.sets[set][block.Entry.ID]; !exists {
			c.refs = append(c.refs, MapRef{Set: set, BlockID: block.Entry.ID})
		}
		c.sets[set][block.Entry.ID] = grid
	}
	return nil
}

func (c Catalog) Lookup(ref MapRef) (Grid, bool) {
	blocks, ok := c.sets[ref.Set]
	if !ok {
		return Grid{}, false
	}
	grid, ok := blocks[ref.BlockID]
	return grid, ok
}

func (c Catalog) Refs() []MapRef { return append([]MapRef(nil), c.refs...) }

func (c Catalog) Len() int { return len(c.refs) }
