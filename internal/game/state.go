// Package game contains platform-neutral remake state. Rendering and input
// adapters (Ebiten or a test harness) call Apply; no DOS assumptions belong
// here.
package game

import (
	"fmt"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
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

type Location uint8

const (
	LocationWilderness Location = iota
	LocationShadowdale
)

type State struct {
	Mode     Mode
	Title    string
	Prompt   string
	Choices  []string
	Message  string
	Location Location

	// OriginalOpening records the English sentence found in the ECL payload.
	// It is evidence that the opening state was sourced from the original data,
	// not a replacement for the localized display string.
	OriginalOpening  string
	OriginalChoices  []string
	OriginalEvent    string
	OriginalLocation string

	catalog           locale.Catalog
	eclBlock          []byte
	eclStart          int
	selectionSequence []uint16
}

func NewStateFromECL(catalog locale.Catalog, block []byte) State {
	state := NewState(catalog)
	state.eclBlock = append([]byte(nil), block...)
	for _, candidate := range ecl.FindPackedTextCandidates(block) {
		if strings.Contains(candidate, "YOU ARE AT THE EDGE OF") {
			state.OriginalOpening = "YOU ARE AT THE EDGE OF"
			break
		}
	}
	if points, _, err := ecl.EntryPoints(block, 5); err == nil && len(points) == 5 {
		start := int(points[4]) - ecl.CodeAddressBase
		state.eclStart = start
		if result, runErr := ecl.RunSubset(block, start, 100); runErr == nil || len(result.Menus) > 0 {
			if len(result.Menus) > 0 {
				for _, option := range result.Menus[0].Options {
					state.OriginalChoices = append(state.OriginalChoices, option)
					switch option {
					case "ENTER CITY":
						state.Choices = append(state.Choices, catalog.Text("enter_city", "Enter city"))
					case "JOURNEY ON":
						state.Choices = append(state.Choices, catalog.Text("journey_on", "Journey on"))
					case "CAMP":
						state.Choices = append(state.Choices, catalog.Text("camp", "Camp"))
					default:
						state.Choices = append(state.Choices, option)
					}
				}
			}
		}
	}
	return state
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
		if len(s.Choices) == 0 {
			s.Choices = []string{
				s.catalog.Text("enter_city", "Enter city"),
				s.catalog.Text("journey_on", "Journey on"),
			}
		}
		s.Message = ""
		return nil
	case s.Mode == ModeWilderness && action == ActionEnterCity:
		return s.Select(0)
	case s.Mode == ModeWilderness && action == ActionJourneyOn:
		return s.Select(1)
	default:
		return fmt.Errorf("action %d is invalid in mode %d", action, s.Mode)
	}
}

// Select applies a localized opening choice and, when the state came from an
// ECL block, runs that choice through the bounded ECL subset.
func (s *State) Select(index int) error {
	if s.Mode != ModeWilderness || index < 0 || index >= len(s.Choices) {
		return fmt.Errorf("choice %d is invalid in mode %d", index, s.Mode)
	}
	s.Mode = ModeEvent
	switch index {
	case 0:
		s.Message = s.catalog.Text("enter_city", "Enter city")
	case 1:
		s.Message = s.catalog.Text("journey_on", "Journey on")
	case 2:
		s.Message = s.catalog.Text("camp", "Camp")
	default:
		s.Message = s.Choices[index]
	}
	if len(s.eclBlock) > 0 {
		s.selectionSequence = append(s.selectionSequence, uint16(index))
		result, _ := ecl.RunSubsetInteractive(s.eclBlock, s.eclStart, 180, s.selectionSequence)
		if len(s.selectionSequence) >= 4 && s.selectionSequence[0] == 0 && s.selectionSequence[1] == 0 && s.selectionSequence[2] == 1 && s.selectionSequence[3] == 0 {
			s.Location = LocationShadowdale
			s.OriginalLocation = "SHADOWDALE"
		}
		if result.WaitingForMenu && len(result.Menus) > 0 {
			menu := result.Menus[len(result.Menus)-1]
			s.Choices = make([]string, 0, len(menu.Options))
			for _, option := range menu.Options {
				s.Choices = append(s.Choices, localizeOption(s.catalog, option))
			}
			if menu.Prompt != "" {
				s.Prompt = localizePrompt(s.catalog, menu.Prompt)
			}
			s.Mode = ModeWilderness
			return nil
		}
		if len(result.Text) > 0 {
			s.OriginalEvent = result.Text[len(result.Text)-1]
		}
	}
	return nil
}

func localizeOption(catalog locale.Catalog, option string) string {
	switch option {
	case "ENTER CITY":
		return catalog.Text("enter_city", "Enter city")
	case "JOURNEY ON":
		return catalog.Text("journey_on", "Journey on")
	case "CAMP":
		return catalog.Text("camp", "Camp")
	case "INN":
		return catalog.Text("inn", "Inn")
	case "STORE":
		return catalog.Text("store", "Store")
	case "BAR":
		return catalog.Text("bar", "Bar")
	case "LEAVE":
		return catalog.Text("leave", "Leave")
	case "SHADOWDALE":
		return catalog.Text("shadowdale", "Shadowdale")
	case "ASHABENFORD":
		return catalog.Text("ashabenford", "Ashabenford")
	case "DAGGER FALLS":
		return catalog.Text("dagger_falls", "Dagger Falls")
	case "WILDERNESS":
		return catalog.Text("wilderness", "Wilderness")
	case "EXIT":
		return catalog.Text("exit", "Exit")
	default:
		return option
	}
}

func localizePrompt(catalog locale.Catalog, prompt string) string {
	if prompt == "PRESS BUTTON OR RETURN TO CONTINUE." {
		return catalog.Text("press_button", "Press any button or Enter to continue")
	}
	return prompt
}
