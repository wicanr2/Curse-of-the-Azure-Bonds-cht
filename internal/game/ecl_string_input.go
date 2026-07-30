package game

import (
	"fmt"
	"unicode/utf8"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

func (s *State) beginECLStringInput(request ecl.StringInputRequest) {
	s.eclStringEditing = true
	s.eclStringValue = ""
	s.eclStringMaxLength = int(request.MaxLength)
	s.Mode = ModeEvent
	s.Prompt = s.catalog.Text("ecl_string_input_prompt", "請輸入回答")
	s.Choices = nil
	s.currentOriginalChoices = nil
}

// ECLStringEditing reports whether the current ECL transaction is paused at
// INPUT STRING. It is distinct from character naming: the submitted text is
// returned to the same VM program counter and stored at the script operand's
// destination address.
func (s *State) ECLStringEditing() bool {
	return s.eclStringEditing
}

func (s *State) ECLStringValue() string {
	return s.eclStringValue
}

func (s *State) ECLStringMaxLength() int {
	return s.eclStringMaxLength
}

func (s *State) AppendECLString(chars []rune) error {
	if !s.eclStringEditing {
		return fmt.Errorf("ECL 字串輸入尚未啟用")
	}
	current := []rune(s.eclStringValue)
	remaining := s.eclStringMaxLength - len(current)
	if remaining <= 0 {
		return nil
	}
	if len(chars) > remaining {
		chars = chars[:remaining]
	}
	s.eclStringValue += string(chars)
	return nil
}

func (s *State) BackspaceECLString() error {
	if !s.eclStringEditing {
		return fmt.Errorf("ECL 字串輸入尚未啟用")
	}
	if s.eclStringValue == "" {
		return nil
	}
	_, size := utf8.DecodeLastRuneInString(s.eclStringValue)
	s.eclStringValue = s.eclStringValue[:len(s.eclStringValue)-size]
	return nil
}

func (s *State) SubmitECLString() error {
	if !s.eclStringEditing {
		return fmt.Errorf("ECL 字串輸入尚未啟用")
	}
	if s.session == nil {
		return fmt.Errorf("ECL 字串輸入需要可續跑的 block session")
	}
	value := s.eclStringValue
	blockBefore := s.session.CurrentBlockID()
	result, err := s.session.ResumeInteractiveInputSeed(
		500, nil, nil, &value, s.eclSeed, s.eclPartyContext(),
	)
	if err != nil {
		return err
	}
	s.eclStringEditing = false
	s.eclStringValue = ""
	s.eclStringMaxLength = 0
	s.requestMusicIfBlockChanged(blockBefore)
	s.eclBlock = s.session.CurrentData()
	handled, err := s.applyDungeonLifecycleResult(result)
	if err != nil {
		return err
	}
	if !handled {
		s.Mode = ModeDungeon
		s.Message = ""
	}
	return nil
}
