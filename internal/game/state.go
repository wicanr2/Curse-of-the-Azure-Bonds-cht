// Package game contains platform-neutral remake state. Rendering and input
// adapters (Ebiten or a test harness) call Apply; no DOS assumptions belong
// here.
package game

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
)

type Mode uint8

const (
	ModeTitle Mode = iota
	ModeWilderness
	ModeEvent
)

type Action uint8

const (
	ActionStart Action = iota
	ActionEnterCity
	ActionJourneyOn
)

type State struct {
	Mode    Mode
	Title   string
	Prompt  string
	Choices []string
	Message string

	catalog locale.Catalog
}

func NewState(catalog locale.Catalog) State {
	return State{
		Mode:    ModeTitle,
		Title:   catalog.Text("title", "Curse of the Azure Bonds"),
		Prompt:  catalog.Text("press_enter", "Press Enter to continue"),
		catalog: catalog,
	}
}

func (s *State) Apply(action Action) error {
	switch {
	case s.Mode == ModeTitle && action == ActionStart:
		s.Mode = ModeWilderness
		s.Prompt = s.catalog.Text("you_are_at_the_edge_of", "You are at the edge of")
		s.Choices = []string{
			s.catalog.Text("enter_city", "Enter city"),
			s.catalog.Text("journey_on", "Journey on"),
		}
		s.Message = ""
		return nil
	case s.Mode == ModeWilderness && action == ActionEnterCity:
		s.Mode = ModeEvent
		s.Message = s.catalog.Text("enter_city", "Enter city")
		return nil
	case s.Mode == ModeWilderness && action == ActionJourneyOn:
		s.Mode = ModeEvent
		s.Message = s.catalog.Text("journey_on", "Journey on")
		return nil
	default:
		return fmt.Errorf("action %d is invalid in mode %d", action, s.Mode)
	}
}
