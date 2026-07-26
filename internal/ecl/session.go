package ecl

import "fmt"

// BlockSession owns decoded ECL blocks and the current block identity. It is
// intentionally separate from RunSubset: VM execution state can be extended
// later without confusing a DAX block switch with a fallthrough branch.
type BlockSession struct {
	blocks  map[uint8][]byte
	current uint8
}

func NewBlockSession(blocks map[uint8][]byte, current uint8) (*BlockSession, error) {
	if len(blocks) == 0 {
		return nil, fmt.Errorf("ECL session has no blocks")
	}
	if _, ok := blocks[current]; !ok {
		return nil, fmt.Errorf("ECL session has no current block 0x%02X", current)
	}
	owned := make(map[uint8][]byte, len(blocks))
	for id, data := range blocks {
		owned[id] = append([]byte(nil), data...)
	}
	return &BlockSession{blocks: owned, current: current}, nil
}

func (s *BlockSession) CurrentBlockID() uint8 { return s.current }

func (s *BlockSession) CurrentData() []byte {
	return append([]byte(nil), s.blocks[s.current]...)
}

func (s *BlockSession) Switch(id uint8) error {
	if _, ok := s.blocks[id]; !ok {
		return fmt.Errorf("ECL session target block 0x%02X is unavailable", id)
	}
	s.current = id
	return nil
}

// InitialEntry returns the fifth vm_init_ecl entry for the current block as a
// decoded payload offset.
func (s *BlockSession) InitialEntry() (int, error) {
	points, _, err := EntryPoints(s.blocks[s.current], 5)
	if err != nil {
		return 0, err
	}
	return int(points[4]) - CodeAddressBase, nil
}

// ApplyResult applies a bounded runner's NEWECL signal to this session.
func (s *BlockSession) ApplyResult(result RunResult) error {
	if result.NewECLBlockID == nil {
		return nil
	}
	return s.Switch(*result.NewECLBlockID)
}
