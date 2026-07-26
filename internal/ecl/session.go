package ecl

import "fmt"

// BlockSession owns decoded ECL blocks and the current block identity. It is
// intentionally separate from RunSubset: VM execution state can be extended
// later without confusing a DAX block switch with a fallthrough branch.
type BlockSession struct {
	blocks          map[uint8][]byte
	current         uint8
	selectionOffset int
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

// RunInteractive executes the current block from its original entry and
// automatically follows bounded NEWECL signals. The selection offset lets a
// caller keep one global input sequence while each newly loaded block sees
// only the choices not consumed by earlier blocks.
func (s *BlockSession) RunInteractive(maxSteps int, selections []uint16) (RunResult, error) {
	start, err := s.InitialEntry()
	if err != nil {
		return RunResult{}, err
	}
	return s.RunFrom(start, maxSteps, selections)
}

// RunFrom executes an explicit event entry in the current block. After a
// NEWECL signal, the target resumes at its own initial entry.
func (s *BlockSession) RunFrom(start, maxSteps int, selections []uint16) (RunResult, error) {
	var aggregate RunResult
	var err error
	for transitions := 0; transitions < 8; transitions++ {
		remaining := selections
		if s.selectionOffset < len(selections) {
			remaining = selections[s.selectionOffset:]
		} else {
			remaining = nil
		}
		result, runErr := RunSubsetInteractive(s.CurrentData(), start, maxSteps, remaining)
		aggregate.Text = append(aggregate.Text, result.Text...)
		aggregate.Menus = append(aggregate.Menus, result.Menus...)
		aggregate.Steps += result.Steps
		aggregate.PC = result.PC
		aggregate.CombatRequested = aggregate.CombatRequested || result.CombatRequested
		if result.MonsterSetup != nil {
			setup := *result.MonsterSetup
			aggregate.MonsterSetup = &setup
		}
		aggregate.MonsterSpawns = append(aggregate.MonsterSpawns, result.MonsterSpawns...)
		aggregate.ProgramIDs = append(aggregate.ProgramIDs, result.ProgramIDs...)
		aggregate.ProgramExit = aggregate.ProgramExit || result.ProgramExit
		s.selectionOffset += result.SelectionsConsumed
		if runErr != nil {
			return aggregate, runErr
		}
		if result.NewECLBlockID == nil {
			aggregate.WaitingForMenu = result.WaitingForMenu
			return aggregate, nil
		}
		if err := s.ApplyResult(result); err != nil {
			return aggregate, err
		}
		start, err = s.InitialEntry()
		if err != nil {
			return aggregate, err
		}
	}
	return aggregate, fmt.Errorf("ECL session exceeded block transition limit")
}
