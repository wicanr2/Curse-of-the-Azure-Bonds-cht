package game

import (
	"archive/zip"
	"path/filepath"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
)

type normalDungeonPoint struct {
	x, y int
}

type normalCampaignObserver struct {
	state         *State
	seen          map[string]bool
	lavaPoolsPass int
	towerGrid     geo.Grid
	towerReady    bool
}

func newNormalCampaignObserver(t *testing.T, state *State) *normalCampaignObserver {
	t.Helper()
	return &normalCampaignObserver{state: state, seen: make(map[string]bool)}
}

func (o *normalCampaignObserver) observe() {
	for _, messageID := range []string{
		"hap.abandoned-village",
		"hap.hiding-peasants",
		"hap.peasants-flee",
		"hap.akabar-join",
		"hap.inn-before-liberation",
		"hap.efreet-barn",
		"hap.efreet-threat",
		"hap.efreet-map",
		"hap.liberated-crowd",
		"hap.elder-thanks",
		"hap.elder-wizard-tower",
		"hap.akabar-secret-routes",
		"hap.leave",
		"hap.map-route",
		"lava-tube.entry",
		"lava-tube.ambush",
		"lava-tube.guarded-door",
		"lava-tube.dream-warning",
		"lava-tube.salamander-pools",
		"lava-tube.intense-heat",
		"lava-tube.fireproof-casks",
		"lava-tube.cask-heat-retreat",
		"wizard-tower.courtyard",
		"wizard-tower.dracandros.arrival",
		"wizard-tower.dracandros.freezes-party",
		"wizard-tower.dragon-roof",
		"wizard-tower.dragon-steps-out",
		"wizard-tower.dracandros.attack-order",
		"wizard-tower.dragon-illusion",
		"wizard-tower.dracandros.journal-15",
		"wizard-tower.dracandros.bond-fades",
		"wizard-tower.dragons-depart",
		"wizard-tower.dracandros.calls-troops",
		"wizard-tower.safe-roof",
		"wizard-tower.dragons-convinced",
		"wizard-tower.dragons-condemn",
		"wizard-tower.take-dragon-heart",
		"wizard-tower.dragon-bodies",
		"wizard-tower.dragon-heart-acid",
		"wizard-tower.roof-exit",
	} {
		if value, ok := o.state.dataPack.Text(messageID, o.state.catalog.Language); ok && o.state.Message == value {
			o.seen[messageID] = true
		}
	}
}

func (o *normalCampaignObserver) selectOption(t *testing.T, optionID string) bool {
	t.Helper()
	index, found := findGamePackOptionIndex(o.state, optionID)
	if !found {
		return false
	}
	if err := o.state.Select(index); err != nil {
		t.Fatalf("select %s: %v", optionID, err)
	}
	return true
}

func (o *normalCampaignObserver) hasOption(optionID string) bool {
	_, found := findGamePackOptionIndex(o.state, optionID)
	return found
}

// resolveDungeonBoundary consumes only interactions that are reachable on the
// normal Hap route.  It deliberately fails closed when a new event or menu is
// encountered instead of silently selecting an unrelated branch.
func (o *normalCampaignObserver) resolveDungeonBoundary(t *testing.T) {
	t.Helper()
	for attempt := 0; attempt < 96; attempt++ {
		o.observe()
		switch o.state.Mode {
		case ModeDungeon:
			return
		case ModeCombat:
			if err := o.state.CombatAct(); err != nil {
				t.Fatalf("normal campaign combat turn: %v", err)
			}
		case ModeEvent:
			if err := o.state.Continue(); err != nil {
				t.Fatalf("continue normal campaign event message=%q: %v", o.state.Message, err)
			}
		case ModeWilderness:
			switch {
			case o.state.session != nil && o.state.session.CurrentBlockID() == 0x33 &&
				o.selectOption(t, "wizard-tower.option.attack-wizard"):
				// Use the smallest verified tower combat branch; the branch key
				// remains a game-pack option rather than a display-string switch.
			case o.state.session != nil && o.state.session.CurrentBlockID() == 0x33 &&
				o.hasOption("ecl-option.combat"):
				if !o.selectOption(t, "ecl-option.wait") {
					t.Fatalf("wizard-tower arrival WAIT option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.session != nil && o.state.session.CurrentBlockID() == 0x33 &&
				o.selectOption(t, "ecl-option.caves"):
				// The tower roof returns through the original CAVES branch.
			case o.state.session != nil && o.state.session.CurrentBlockID() == 0x32 &&
				o.hasOption("ecl-option.combat") && o.hasOption("ecl-option.wait"):
				// The picture/press boundary is followed by the original encounter
				// menu.  Take WAIT once to record the verified friendly-parlay
				// branch, then revisit the same cell and take COMBAT.
				if o.lavaPoolsPass == 0 {
					if !o.selectOption(t, "ecl-option.wait") {
						t.Fatalf("lava pools WAIT option unavailable: %v", o.state.currentOriginalChoices)
					}
					o.lavaPoolsPass++
				} else if !o.selectOption(t, "ecl-option.combat") {
					t.Fatalf("lava pools COMBAT option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "hap.hiding-peasants"):
				if !o.selectOption(t, "ecl-option.try-to-talk-further") {
					t.Fatalf("Hap peasants option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "hap.akabar-join"):
				if !o.selectOption(t, "option.yes") {
					t.Fatalf("Akabar join option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "hap.inn-before-liberation"):
				if !o.selectOption(t, "option.no") {
					t.Fatalf("Hap inn option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "hap.efreet-barn"),
				o.state.Message == requireGamePackText(t, o.state, "hap.efreet-threat"),
				o.state.Message == requireGamePackText(t, o.state, "hap.liberated-crowd"),
				o.state.Message == requireGamePackText(t, o.state, "hap.elder-thanks"),
				o.state.Message == requireGamePackText(t, o.state, "hap.elder-wizard-tower"),
				o.state.Message == requireGamePackText(t, o.state, "hap.akabar-secret-routes"),
				o.state.Message == requireGamePackText(t, o.state, "lava-tube.entry"),
				o.state.Message == requireGamePackText(t, o.state, "lava-tube.ambush"),
				o.state.Message == requireGamePackText(t, o.state, "lava-tube.guarded-door"),
				o.state.Message == requireGamePackText(t, o.state, "lava-tube.dream-warning"),
				o.state.Message == requireGamePackText(t, o.state, "lava-tube.intense-heat"):
				if !o.selectOption(t, "ecl-option.press-button-or-return-to-continue") {
					t.Fatalf("press continuation unavailable for message=%q choices=%v",
						o.state.Message, o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "lava-tube.salamander-pools"):
				if o.hasOption("ecl-option.press-button-or-return-to-continue") {
					if !o.selectOption(t, "ecl-option.press-button-or-return-to-continue") {
						t.Fatalf("lava pools opening continuation unavailable: %v", o.state.currentOriginalChoices)
					}
				} else if o.lavaPoolsPass == 0 && o.selectOption(t, "ecl-option.wait") {
					o.lavaPoolsPass++
				} else if o.selectOption(t, "ecl-option.parlay-nice") {
					// The first WAIT branch exposes the parlay tactics menu.
				} else if !o.selectOption(t, "ecl-option.combat") {
					t.Fatalf("lava pools encounter option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "lava-tube.nice-parlay"):
				if !o.selectOption(t, "ecl-option.press-button-or-return-to-continue") {
					t.Fatalf("lava parlay continuation unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "lava-tube.fireproof-casks"):
				if !o.selectOption(t, "option.yes") {
					t.Fatalf("lava cask YES option unavailable: %v", o.state.currentOriginalChoices)
				}
				if err := o.state.Select(0); err != nil {
					t.Fatalf("select normal cask volunteer: %v", err)
				}
			case o.state.Message == requireGamePackText(t, o.state, "lava-tube.cask-heat-retreat"):
				if !o.selectOption(t, "option.no") {
					t.Fatalf("lava cask retreat option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.session != nil && o.state.session.CurrentBlockID() == 0x33 &&
				o.selectOption(t, "ecl-option.press-button-or-return-to-continue"):
				if o.state.Mode == ModeDungeon && !o.towerReady {
					o.state.TurnDungeonWithGrid(o.towerGrid, (2-int(o.state.DungeonDirection)+8)%8)
					if err := o.state.RunDungeonLifecycle(); err != nil {
						t.Fatalf("open wizard-tower roof exit: %v", err)
					}
					o.towerReady = true
				}
			case o.state.Message == requireGamePackText(t, o.state, "hap.leave"):
				optionID := "option.no"
				if o.state.DungeonX == 15 && o.state.DungeonY == 5 {
					optionID = "option.yes"
				}
				if !o.selectOption(t, optionID) {
					t.Fatalf("Hap leave option %s unavailable: %v", optionID, o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "hap.map-route"):
				if !o.selectOption(t, "ecl-option.caves") {
					t.Fatalf("Hap map route option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.selectOption(t, "ecl-option.parlay-nice"):
			case o.selectOption(t, "option.no"):
			case o.selectOption(t, "option.flee"):
			case o.selectOption(t, "ecl-option.press-button-or-return-to-continue"):
			default:
				t.Fatalf("unexpected normal campaign wilderness boundary mode=%v message=%q choices=%v position=(%d,%d,%d)",
					o.state.Mode, o.state.Message, o.state.currentOriginalChoices,
					o.state.DungeonX, o.state.DungeonY, o.state.DungeonDirection)
			}
		default:
			t.Fatalf("unexpected normal campaign mode=%v message=%q choices=%v",
				o.state.Mode, o.state.Message, o.state.currentOriginalChoices)
		}
	}
	t.Fatalf("normal campaign boundary did not settle: mode=%v message=%q choices=%v block=%#x position=(%d,%d,%d)",
		o.state.Mode, o.state.Message, o.state.currentOriginalChoices,
		o.state.session.CurrentBlockID(), o.state.DungeonX, o.state.DungeonY,
		o.state.DungeonDirection)
}

func loadGeo5CampaignGrid(t *testing.T, image *zip.ReadCloser, blockID uint8) geo.Grid {
	t.Helper()
	catalog := geo.NewCatalog()
	if err := catalog.AddDAX(5, zipData(t, image, "GEO5.DAX")); err != nil {
		t.Fatal(err)
	}
	grid, ok := catalog.Lookup(geo.MapRef{Set: 5, BlockID: blockID})
	if !ok {
		t.Fatalf("missing GEO5 block %#x", blockID)
	}
	return grid
}

func walkNormalDungeonTo(t *testing.T, state *State, grid geo.Grid, targetX, targetY int, observer *normalCampaignObserver) {
	t.Helper()
	target := normalDungeonPoint{targetX, targetY}
	for hop := 0; hop < geo.Width*geo.Height*4 && (state.DungeonX != targetX || state.DungeonY != targetY); hop++ {
		start := normalDungeonPoint{state.DungeonX, state.DungeonY}
		queue := []normalDungeonPoint{start}
		previous := map[normalDungeonPoint]struct {
			point     normalDungeonPoint
			direction int
		}{start: {point: start, direction: -1}}
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			for _, direction := range []int{0, 2, 4, 6} {
				nextX, nextY := current.x, current.y
				switch direction {
				case 0:
					nextY--
				case 2:
					nextX++
				case 4:
					nextY++
				case 6:
					nextX--
				}
				if nextX < 0 || nextX >= geo.Width || nextY < 0 || nextY >= geo.Height {
					continue
				}
				next := normalDungeonPoint{nextX, nextY}
				// Keep the route on the deterministic story corridor; these cells
				// are optional random-encounter cells, not required rooms.
				if next == (normalDungeonPoint{10, 2}) || next == (normalDungeonPoint{8, 11}) {
					continue
				}
				if _, found := previous[next]; found || !grid.CanMoveDungeonWrapped(current.x, current.y, direction) {
					continue
				}
				previous[next] = struct {
					point     normalDungeonPoint
					direction int
				}{point: current, direction: direction}
				queue = append(queue, next)
			}
		}
		if _, found := previous[target]; !found {
			t.Fatalf("normal dungeon target (%d,%d) is unreachable from (%d,%d)", targetX, targetY, start.x, start.y)
		}
		current := target
		path := make([]struct {
			point     normalDungeonPoint
			direction int
		}, 0)
		for current != start {
			edge := previous[current]
			path = append(path, struct {
				point     normalDungeonPoint
				direction int
			}{point: current, direction: edge.direction})
			current = edge.point
		}
		for left, right := 0, len(path)-1; left < right; left, right = left+1, right-1 {
			path[left], path[right] = path[right], path[left]
		}
		step := path[0]
		observer.resolveDungeonBoundary(t)
		if state.Mode != ModeDungeon {
			t.Fatalf("normal dungeon route before hop %d mode=%v message=%q", hop, state.Mode, state.Message)
		}
		if err := state.MoveDungeon(grid,
			step.point.x-state.DungeonX, step.point.y-state.DungeonY, step.direction); err != nil {
			t.Fatalf("normal dungeon hop %d toward (%d,%d) from (%d,%d): %v", hop, targetX, targetY,
				state.DungeonX, state.DungeonY, err)
		}
		observer.resolveDungeonBoundary(t)
	}
	if state.DungeonX != targetX || state.DungeonY != targetY {
		t.Fatalf("normal dungeon route did not reach (%d,%d), ended at (%d,%d)",
			targetX, targetY, state.DungeonX, state.DungeonY)
	}
}

func TestRealNewGameContinuesFromHapToDracolichCave(t *testing.T) {
	state := runNormalNewGameToEssembra(t)
	if state == nil {
		return
	}
	selectOption := func(optionID string) {
		t.Helper()
		if err := state.Select(requireGamePackOptionIndex(t, state, optionID)); err != nil {
			t.Fatalf("select %s: %v", optionID, err)
		}
	}
	selectOption("ecl-option.journey-on")
	selectOption("ecl-option.hap")
	selectOption("ecl-option.trail")
	selectOption("ecl-option.press-button-or-return-to-continue")
	for turn := 0; turn < 128 && state.Mode == ModeCombat; turn++ {
		if err := state.CombatAct(); err != nil {
			t.Fatalf("Hap travel combat turn %d: %v", turn, err)
		}
	}
	if state.Mode == ModeEvent {
		if err := state.Continue(); err != nil {
			t.Fatalf("continue Hap travel combat: %v", err)
		}
	}
	selectOption("ecl-option.enter-city")
	if state.PictureRequested {
		if err := state.Continue(); err != nil {
			t.Fatalf("continue Hap entry picture: %v", err)
		}
	}
	selectOption("ecl-option.press-button-or-return-to-continue")
	if state.Mode != ModeDungeon || state.session.CurrentBlockID() != 0x31 ||
		state.GeoMapSet != 5 || state.GeoMapBlock != 0x32 {
		t.Fatalf("Hap dungeon entry mode=%v block=%#x geo=%d/%#x", state.Mode,
			state.session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock)
	}

	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Fatal(err)
	}
	defer image.Close()
	grid := loadGeo5CampaignGrid(t, image, 0x32)
	towerGrid := loadGeo5CampaignGrid(t, image, 0x33)
	observer := newNormalCampaignObserver(t, state)
	observer.towerGrid = towerGrid
	for _, target := range []normalDungeonPoint{{4, 10}, {9, 10}, {3, 13}, {15, 5}} {
		walkNormalDungeonTo(t, state, grid, target.x, target.y, observer)
	}
	if !observer.seen["hap.peasants-flee"] || !observer.seen["hap.akabar-join"] ||
		!observer.seen["hap.efreet-map"] {
		t.Fatalf("normal Hap story coverage=%v", observer.seen)
	}
	for address, want := range map[uint16]uint16{0x4C01: 5, 0x4C02: 1, 0x4C5E: 1, 0x4C5F: 1} {
		if got, ok := state.session.MemoryValue(address); !ok || got != want {
			t.Fatalf("normal Hap memory[%#x]=%#x,%v want %#x", address, got, ok, want)
		}
	}
	if err := state.MoveDungeon(grid, 1, 0, 2); err != nil {
		t.Fatalf("normal Hap east external exit: %v", err)
	}
	observer.resolveDungeonBoundary(t)
	if state.session.CurrentBlockID() != 0x32 || state.GeoMapSet != 5 || state.GeoMapBlock != 0x32 ||
		state.DungeonX != 15 || state.DungeonY != 5 || state.DungeonDirection != 6 ||
		!observer.seen["hap.leave"] || !observer.seen["hap.map-route"] ||
		!observer.seen["lava-tube.entry"] || !observer.seen["lava-tube.ambush"] {
		t.Fatalf("normal Hap cave handoff block=%#x geo=%d/%#x pos=(%d,%d,%d) mode=%v coverage=%v",
			state.session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock,
			state.DungeonX, state.DungeonY, state.DungeonDirection, state.Mode, observer.seen)
	}

	walkNormalDungeonTo(t, state, grid, 9, 10, observer)
	if !observer.seen["lava-tube.guarded-door"] || state.Mode != ModeDungeon {
		t.Fatalf("normal lava guarded-door route mode=%v coverage=%v position=(%d,%d,%d)",
			state.Mode, observer.seen, state.DungeonX, state.DungeonY, state.DungeonDirection)
	}
	walkNormalDungeonTo(t, state, grid, 0, 5, observer)
	// The original ECL5 pool branch is time-gated. Advance the shared clock
	// through the engine time service, equivalent to a normal CAMP/REST period;
	// do not write 4BC9 directly in this normal-session test.
	if err := state.AdvanceGameTimeHours(8); err != nil {
		t.Fatalf("advance normal dungeon rest time: %v", err)
	}
	state.TurnDungeonWithGrid(grid, (8-int(state.DungeonDirection))%8)
	state.syncDungeonECLRegisters()
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatalf("run normal lava pool cell lifecycle: %v", err)
	}
	observer.resolveDungeonBoundary(t)
	if !observer.seen["lava-tube.salamander-pools"] || state.Mode != ModeDungeon {
		t.Fatalf("normal lava pool opening coverage=%v mode=%v position=(%d,%d,%d)",
			observer.seen, state.Mode, state.DungeonX, state.DungeonY, state.DungeonDirection)
	}
	// The original WAIT → PARLAY branch returns to the same dungeon cell.
	// Re-entering that cell's per-turn handler is the normal route to the
	// salamander battle; it is not a direct flag or PC injection.
	if err := state.RunDungeonLifecycle(); err != nil {
		t.Fatalf("rerun normal lava pool cell lifecycle: %v", err)
	}
	observer.resolveDungeonBoundary(t)
	if !observer.seen["lava-tube.salamander-pools"] || !observer.seen["lava-tube.intense-heat"] {
		flag, _ := state.session.MemoryValue(0x4C48)
		boss, _ := state.session.MemoryValue(0x4C60)
		leader, _ := state.session.MemoryValue(0x4C01)
		gate, _ := state.session.MemoryValue(0x4C5E)
		guard, _ := state.session.MemoryValue(0x7F81)
		hour, _ := state.session.MemoryValue(0x4BC9)
		boundary, _ := state.session.MemoryValue(0x7ED5)
		forced, _ := state.session.MemoryValue(0x7EC9)
		wall, _ := state.session.MemoryValue(0xC04E)
		roof, _ := state.session.MemoryValue(0xC04F)
		t.Fatalf("normal lava pool route coverage=%v mode=%v position=(%d,%d,%d) 4C48=%#x 4C60=%#x 4C01=%#x 4C5E=%#x 7F81=%#x 4BC9=%#x 7ED5=%#x 7EC9=%#x C04E=%#x C04F=%#x",
			observer.seen, state.Mode, state.DungeonX, state.DungeonY, state.DungeonDirection,
			flag, boss, leader, gate, guard, hour, boundary, forced, wall, roof)
	}
	if !observer.seen["wizard-tower.courtyard"] ||
		!observer.seen["wizard-tower.dracandros.arrival"] ||
		!observer.seen["wizard-tower.safe-roof"] ||
		state.session.CurrentBlockID() != 0x32 || state.GeoMapSet != 5 || state.GeoMapBlock != 0x32 {
		t.Fatalf("normal Hap-to-tower round trip block=%#x geo=%d/%#x mode=%v position=(%d,%d,%d) coverage=%v",
			state.session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock, state.Mode,
			state.DungeonX, state.DungeonY, state.DungeonDirection, observer.seen)
	}
}
