package game

import (
	"archive/zip"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

type normalDungeonPoint struct {
	x, y int
}

func normalDungeonDelta(direction int) (int, int) {
	switch direction {
	case 0:
		return 0, -1
	case 2:
		return 1, 0
	case 4:
		return 0, 1
	case 6:
		return -1, 0
	default:
		return 0, 0
	}
}

type normalCampaignObserver struct {
	state                 *State
	seen                  map[string]bool
	lavaPoolsPass         int
	towerGrid             geo.Grid
	towerReady            bool
	towerReturnOption     string
	stopAtWorldEdge       bool
	nextWorldDestinations []string
	stopAtMessageID       string
	stopAtDataPackEventID string
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
		"area5.depart-akabar",
		"area5.depart-akabar-reluctant",
		"area5.dark-elf-gear-decays",
		"post-wizard.dracolich",
		"essembra.edge",
		"hillsfar.edge",
		"hillsfar.places",
		"hillsfar.dockside-bar",
		"hillsfar.red-plumes-spill-drinks",
		"yulash.edge",
		"yulash.entry",
		"yulash.riders-burst-out",
		"yulash.checkpoint-halt",
		"yulash.see-commander",
		"yulash.waiting-room",
		"yulash.zhentarim-spies",
		"yulash.led-to-commander",
		"yulash.commander-business",
		"journal-trigger.yulash-commander-22",
		"yulash.commander-side-door",
		"yulash.pit-entrance",
		"zhentil.patrol_pass",
		"zhentil.edge",
		"zhentil.guards_question",
		"zhentil.guards_warning",
		"zhentil.inner_city",
		"zhentil.fritz-accusation",
		"zhentil.fritz-killed",
		"zhentil.fritz-let-go",
		"zhentil.olive_appears",
		"zhentil.olive_follow",
		"zhentil.dark_shrine_entry",
		"zhentil.olive_explains",
		"zhentil.dimswart_door",
		"zhentil.olive_leaves",
		"zhentil.olive_cell_hint",
		"zhentil.olive_repeats_dimswart",
		"zhentil.dimswart_appears",
		"zhentil.dimswart_join",
		"zhentil.hooded_offer",
		"zhentil.hooded_follow",
		"zhentil.fzoul_interrupts",
		"zhentil.fzoul_retreats",
		"dexam.arrival",
		"dexam.journal_30",
		"dexam.amulet_choice",
		"dexam.fzoul_journal_7",
		"dexam.kills_fzoul",
		"dexam.fzoul_bond_fades",
		"dexam.kill_order",
		"dexam.amulet_rises",
		"dexam.altar_melee",
		"dexam.final_reveal",
		"dexam.attack",
		"dexam.amulet_retrieved",
		"dexam.zhentil_attack",
		"dexam.departure.olive",
		"dexam.departure.dimswart",
		"dexam.departure.gharri",
		"dexam.departure.riders",
		"pit.alias-dragonbait-meet",
		"pit.alias-bonded-reaction",
		"pit.alias-dragonbait-introduction",
		"journal-trigger.alias-story-3",
		"pit.alias-dragonbait-join",
		"pit.alias-dragonbait-joined",
		"pit.stairs-down",
		"pit.stairs-up",
		"pit.dead-zhentrim",
		"pit.zhentrim-scroll",
		"pit.mogion-altar",
		"pit.mogion-self-identifies",
		"pit.alias-identifies-mogion",
		"pit.mogion-greeting",
		"pit.opening-dead-cultists",
		"pit.opening-chosen",
		"pit.trapped",
		"pit.cleric-dies",
		"pit.ambience",
		"pit.bond-paralysis",
		"pit.alias-dragonbait-tendrils",
		"pit.mogion-ritual",
		"pit.dimensional-window",
		"pit.moander-returns",
		"pit.bond-fades",
		"pit.bond-broken",
		"pit.alias-attack-mogion",
		"pit.rift-closes",
		"pit.remnants-scream",
		"pit.remnants-attack",
		"pit.gauntlet",
		"pit.priest-flees",
		"pit.altar-treasure",
		"pit.exit-last-stand",
		"standing-stone.grey-man",
		"standing-stone.four-masters",
		"standing-stone.seek-red",
		"myth-drannor.tyranthraxus-reveal",
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
		if value, ok := o.state.dataPack.Text(messageID, o.state.catalog.Language); ok &&
			(o.state.Message == value || strings.Contains(o.state.Message, value)) {
			o.seen[messageID] = true
		}
	}
}

func (o *normalCampaignObserver) stoppedAtDataPackEvent() bool {
	return o != nil && o.state != nil && o.stopAtDataPackEventID != "" &&
		o.state.appliedDataPackEvents[o.stopAtDataPackEventID]
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

func (o *normalCampaignObserver) isWorldMenuBlock() bool {
	if o.state == nil || o.state.session == nil {
		return false
	}
	block := o.state.session.CurrentBlockID()
	return block == 0x50 || block == 0x51
}

func (o *normalCampaignObserver) selectNextWorldDestination(t *testing.T) bool {
	if len(o.nextWorldDestinations) == 0 || !o.hasOption(o.nextWorldDestinations[0]) {
		return false
	}
	optionID := o.nextWorldDestinations[0]
	o.nextWorldDestinations = o.nextWorldDestinations[1:]
	return o.selectOption(t, optionID)
}

// resolveDungeonBoundary consumes only interactions that are reachable on the
// normal Hap route.  It deliberately fails closed when a new event or menu is
// encountered instead of silently selecting an unrelated branch.
func (o *normalCampaignObserver) resolveDungeonBoundary(t *testing.T) {
	t.Helper()
	for attempt := 0; attempt < 160; attempt++ {
		o.observe()
		if o.stoppedAtDataPackEvent() {
			return
		}
		if o.stopAtWorldEdge && o.state.Mode == ModeWilderness &&
			o.state.Message == requireGamePackText(t, o.state, "essembra.edge") {
			return
		}
		if o.stopAtMessageID != "" &&
			o.state.Message == requireGamePackText(t, o.state, o.stopAtMessageID) {
			return
		}
		switch o.state.Mode {
		case ModeDungeon:
			if o.state.session != nil && o.state.session.CurrentBlockID() == 0x33 &&
				!o.towerReady {
				o.state.TurnDungeonWithGrid(o.towerGrid, (2-int(o.state.DungeonDirection)+8)%8)
				if err := o.state.RunDungeonLifecycle(); err != nil {
					t.Fatalf("run wizard-tower entry lifecycle: %v", err)
				}
				o.towerReady = true
				continue
			}
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
			case o.state.treasureMenu:
				index, found := o.state.OriginalChoiceIndex("TREASURE_EXIT")
				if !found {
					t.Fatalf("normal campaign treasure menu has no exit: %v", o.state.currentOriginalChoices)
				}
				if err := o.state.Select(index); err != nil {
					t.Fatalf("leave normal campaign treasure menu: %v", err)
				}
			case o.state.session != nil && o.state.session.CurrentBlockID() == 0x22 &&
				o.selectOption(t, "ecl-option.leave"):
				// The cave route can cross an optional random beholder event.
				// LEAVE is the original safe branch; use the stable option ID so
				// this observer does not depend on the English prompt text.
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
				o.hasOption("ecl-option.caves") && o.hasOption("ecl-option.wilderness"):
				optionID := o.towerReturnOption
				if optionID == "" {
					optionID = "ecl-option.caves"
				}
				if !o.selectOption(t, optionID) {
					t.Fatalf("wizard-tower return option %s unavailable: %v", optionID, o.state.currentOriginalChoices)
				}
				if optionID == "ecl-option.wilderness" {
					// WILDERNESS returns to the tower's exit cell first; the
					// following dungeon lifecycle exposes VILLAGE/DEPART.
					o.towerReady = false
				}
				// The tower roof can return through the original CAVES or
				// WILDERNESS branch; the choice is data-driven for this path.
			case o.state.session != nil && o.state.session.CurrentBlockID() == 0x33 &&
				o.towerReturnOption == "ecl-option.wilderness" &&
				o.selectOption(t, "ecl-option.depart"):
				// DEPART continues through the original Area5 farewell before
				// handing control back to the world route.
			case o.state.Message == requireGamePackText(t, o.state, "zhentil.edge"):
				if !o.selectOption(t, "ecl-option.enter-city") {
					t.Fatalf("Zhentil ENTER CITY option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "zhentil.fritz-accusation"):
				if !o.selectOption(t, "ecl-option.let-him-go") {
					t.Fatalf("Zhentil Fritz LET HIM GO option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.isWorldMenuBlock() &&
				o.selectNextWorldDestination(t):
			case o.isWorldMenuBlock() &&
				o.selectOption(t, "ecl-option.journey-on"):
			case o.isWorldMenuBlock() &&
				o.selectOption(t, "ecl-option.essembra"):
			case o.isWorldMenuBlock() &&
				o.selectOption(t, "ecl-option.trail"):
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
			case o.state.Message == requireGamePackText(t, o.state, "standing-stone.four-masters"):
				if !o.selectOption(t, "ecl-option.thank-him") {
					t.Fatalf("Standing Stone THANK HIM option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "yulash.checkpoint-halt"):
				if !o.selectOption(t, "ecl-option.parlay") {
					t.Fatalf("Yulash checkpoint PARLAY option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "yulash.see-commander"):
				if !o.selectOption(t, "ecl-option.go-with-guards") {
					t.Fatalf("Yulash GO WITH GUARDS option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "yulash.zhentarim-spies"):
				if !o.selectOption(t, "ecl-option.fight-the-men") {
					t.Fatalf("Yulash FIGHT THE MEN option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "yulash.commander-business"):
				if !o.selectOption(t, "ecl-option.parlay-nice") {
					t.Fatalf("Yulash PARLAY_NICE option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "zhentil.olive_follow"),
				o.state.Message == requireGamePackText(t, o.state, "zhentil.hooded_follow"):
				if !o.selectOption(t, "option.yes") {
					t.Fatalf("Zhentil follow YES option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "zhentil.dimswart_join"):
				if !o.selectOption(t, "option.yes") {
					t.Fatalf("Dimswart join YES option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "dexam.amulet_choice"):
				if !o.selectOption(t, "ecl-option.combat") {
					t.Fatalf("Dexam amulet COMBAT option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "pit.alias-bonded-reaction"):
				if !o.selectOption(t, "ecl-option.parlay") {
					t.Fatalf("Pit Alias PARLAY option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "pit.mogion-greeting") &&
				o.hasOption("ecl-option.parlay"):
				if !o.selectOption(t, "ecl-option.parlay") {
					t.Fatalf("Pit Mogion PARLAY option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "pit.alias-attack-mogion"):
				if !o.selectOption(t, "ecl-option.attack") {
					t.Fatalf("Pit Alias attack option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "pit.stairs-down"),
				o.state.Message == requireGamePackText(t, o.state, "pit.stairs-up"):
				if !o.selectOption(t, "option.yes") {
					t.Fatalf("Pit stair YES option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "pit.dead-zhentrim"):
				if !o.selectOption(t, "ecl-option.examine-corpse") {
					t.Fatalf("Pit Zhentil corpse EXAMINE option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "pit.alias-dragonbait-introduction"):
				if !o.selectOption(t, "ecl-option.press-button-or-return-to-continue") {
					t.Fatalf("Pit Alias introduction continuation unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "journal-trigger.alias-story-3"):
				if !o.selectOption(t, "ecl-option.press-button-or-return-to-continue") {
					t.Fatalf("Pit Alias story continuation unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "pit.alias-dragonbait-join"):
				if !o.selectOption(t, "option.yes") {
					t.Fatalf("Pit Alias join YES option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.hasOption("ecl-option.tell-her-your-story"):
				if !o.selectOption(t, "ecl-option.tell-her-your-story") {
					t.Fatalf("Pit Alias story option unavailable: %v", o.state.currentOriginalChoices)
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
				o.state.Message == requireGamePackText(t, o.state, "lava-tube.intense-heat"),
				o.state.Message == requireGamePackText(t, o.state, "area5.depart-akabar"),
				o.state.Message == requireGamePackText(t, o.state, "area5.depart-akabar-reluctant"),
				o.state.Message == requireGamePackText(t, o.state, "area5.dark-elf-gear-decays"),
				o.state.Message == requireGamePackText(t, o.state, "post-wizard.dracolich"),
				o.state.Message == requireGamePackText(t, o.state, "standing-stone.grey-man"),
				o.state.Message == requireGamePackText(t, o.state, "standing-stone.seek-red"),
				o.state.Message == requireGamePackText(t, o.state, "myth-drannor.tyranthraxus-reveal"),
				o.state.Message == requireGamePackText(t, o.state, "yulash.riders-burst-out"),
				o.state.Message == requireGamePackText(t, o.state, "yulash.waiting-room"),
				o.state.Message == requireGamePackText(t, o.state, "yulash.led-to-commander"),
				o.state.Message == requireGamePackText(t, o.state, "journal-trigger.yulash-commander-22"),
				o.state.Message == requireGamePackText(t, o.state, "yulash.commander-side-door"),
				o.state.Message == requireGamePackText(t, o.state, "yulash.pit-entrance"),
				o.state.Message == requireGamePackText(t, o.state, "pit.alias-dragonbait-meet"),
				o.state.Message == requireGamePackText(t, o.state, "pit.alias-dragonbait-joined"),
				o.state.Message == requireGamePackText(t, o.state, "pit.zhentrim-scroll"),
				o.state.Message == requireGamePackText(t, o.state, "pit.mogion-altar"),
				o.state.Message == requireGamePackText(t, o.state, "pit.mogion-self-identifies"),
				o.state.Message == requireGamePackText(t, o.state, "pit.alias-identifies-mogion"),
				o.state.Message == requireGamePackText(t, o.state, "pit.opening-dead-cultists"),
				o.state.Message == requireGamePackText(t, o.state, "pit.opening-chosen"),
				o.state.Message == requireGamePackText(t, o.state, "pit.trapped"),
				o.state.Message == requireGamePackText(t, o.state, "pit.cleric-dies"),
				o.state.Message == requireGamePackText(t, o.state, "pit.ambience"),
				o.state.Message == requireGamePackText(t, o.state, "pit.bond-paralysis"),
				o.state.Message == requireGamePackText(t, o.state, "pit.alias-dragonbait-tendrils"),
				o.state.Message == requireGamePackText(t, o.state, "pit.mogion-ritual"),
				o.state.Message == requireGamePackText(t, o.state, "pit.dimensional-window"),
				o.state.Message == requireGamePackText(t, o.state, "pit.moander-returns"),
				o.state.Message == requireGamePackText(t, o.state, "pit.bond-fades"),
				o.state.Message == requireGamePackText(t, o.state, "pit.bond-broken"),
				o.state.Message == requireGamePackText(t, o.state, "pit.rift-closes"),
				o.state.Message == requireGamePackText(t, o.state, "pit.remnants-scream"),
				o.state.Message == requireGamePackText(t, o.state, "pit.remnants-attack"),
				o.state.Message == requireGamePackText(t, o.state, "pit.gauntlet"),
				o.state.Message == requireGamePackText(t, o.state, "pit.priest-flees"),
				o.state.Message == requireGamePackText(t, o.state, "pit.altar-treasure"),
				o.state.Message == requireGamePackText(t, o.state, "pit.exit-last-stand"),
				o.state.Message == requireGamePackText(t, o.state, "zhentil.patrol_pass"),
				o.state.Message == requireGamePackText(t, o.state, "zhentil.guards_question"),
				o.state.Message == requireGamePackText(t, o.state, "zhentil.guards_warning"),
				o.state.Message == requireGamePackText(t, o.state, "zhentil.inner_city"),
				o.state.Message == requireGamePackText(t, o.state, "zhentil.fritz-killed"),
				o.state.Message == requireGamePackText(t, o.state, "zhentil.fritz-let-go"),
				o.state.Message == requireGamePackText(t, o.state, "zhentil.olive_appears"),
				o.state.Message == requireGamePackText(t, o.state, "zhentil.dark_shrine_entry"),
				o.state.Message == requireGamePackText(t, o.state, "zhentil.olive_explains"),
				o.state.Message == requireGamePackText(t, o.state, "zhentil.dimswart_door"),
				o.state.Message == requireGamePackText(t, o.state, "zhentil.olive_leaves"),
				o.state.Message == requireGamePackText(t, o.state, "zhentil.olive_cell_hint"),
				o.state.Message == requireGamePackText(t, o.state, "zhentil.olive_repeats_dimswart"),
				o.state.Message == requireGamePackText(t, o.state, "zhentil.dimswart_appears"),
				o.state.Message == requireGamePackText(t, o.state, "zhentil.hooded_offer"),
				o.state.Message == requireGamePackText(t, o.state, "zhentil.fzoul_interrupts"),
				o.state.Message == requireGamePackText(t, o.state, "zhentil.fzoul_retreats"),
				o.state.Message == requireGamePackText(t, o.state, "dexam.arrival"),
				o.state.Message == requireGamePackText(t, o.state, "dexam.journal_30"),
				o.state.Message == requireGamePackText(t, o.state, "dexam.fzoul_journal_7"),
				o.state.Message == requireGamePackText(t, o.state, "dexam.kills_fzoul"),
				o.state.Message == requireGamePackText(t, o.state, "dexam.fzoul_bond_fades"),
				o.state.Message == requireGamePackText(t, o.state, "dexam.kill_order"),
				o.state.Message == requireGamePackText(t, o.state, "dexam.amulet_rises"),
				o.state.Message == requireGamePackText(t, o.state, "dexam.altar_melee"),
				o.state.Message == requireGamePackText(t, o.state, "dexam.final_reveal"),
				o.state.Message == requireGamePackText(t, o.state, "dexam.attack"),
				o.state.Message == requireGamePackText(t, o.state, "dexam.amulet_retrieved"),
				o.state.Message == requireGamePackText(t, o.state, "dexam.zhentil_attack"),
				o.state.Message == requireGamePackText(t, o.state, "dexam.departure.olive"),
				o.state.Message == requireGamePackText(t, o.state, "dexam.departure.dimswart"),
				o.state.Message == requireGamePackText(t, o.state, "dexam.departure.gharri"),
				o.state.Message == requireGamePackText(t, o.state, "dexam.departure.riders"):
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
			case o.selectOption(t, "ecl-option.evade"):
			case o.selectOption(t, "option.flee"):
			case o.selectOption(t, "ecl-option.press-button-or-return-to-continue"):
			default:
				route := []uint16{}
				if o.state.session != nil {
					for address := uint16(0x4C02); address <= 0x4C05; address++ {
						value, _ := o.state.session.MemoryValue(address)
						route = append(route, value)
					}
				}
				t.Fatalf("unexpected normal campaign wilderness boundary mode=%v city=%d message=%q choices=%v route=%v position=(%d,%d,%d)",
					o.state.Mode, o.state.Area.CurrentCity, o.state.Message, o.state.currentOriginalChoices, route,
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

func loadGeoCampaignGrid(t *testing.T, image *zip.ReadCloser, set uint8, filename string, blockID uint8) geo.Grid {
	t.Helper()
	catalog := geo.NewCatalog()
	if err := catalog.AddDAX(set, zipData(t, image, filename)); err != nil {
		t.Fatal(err)
	}
	grid, ok := catalog.Lookup(geo.MapRef{Set: set, BlockID: blockID})
	if !ok {
		t.Fatalf("missing GEO%d block %#x", set, blockID)
	}
	return grid
}

func loadGeo5CampaignGrid(t *testing.T, image *zip.ReadCloser, blockID uint8) geo.Grid {
	t.Helper()
	return loadGeoCampaignGrid(t, image, 5, "GEO5.DAX", blockID)
}

func openNormalDungeonDoor(t *testing.T, state *State, grid *geo.Grid) {
	t.Helper()
	_, _, direction := state.DungeonGeometryView()
	flags, ok := grid.WallDoorFlagsWrapped(state.DungeonX, state.DungeonY, int(direction))
	if !ok || (flags != 2 && flags != 3) {
		t.Fatalf("normal dungeon expected locked door at (%d,%d,%d), flags=%#x ok=%v",
			state.DungeonX, state.DungeonY, direction, flags, ok)
	}
	options := state.DungeonDoorMenuOptions(flags)
	if !options.Bash {
		t.Fatalf("normal dungeon locked door has no Bash action: flags=%#x options=%#v", flags, options)
	}
	result := state.BashDungeonDoor(flags)
	if !result.Opened {
		t.Fatalf("normal dungeon Bash failed at (%d,%d,%d): result=%#v options=%#v",
			state.DungeonX, state.DungeonY, direction, result, options)
	}
	if !grid.UnlockDoorWrapped(state.DungeonX, state.DungeonY, int(direction)) {
		t.Fatalf("normal dungeon Bash succeeded but GEO door did not unlock at (%d,%d,%d)",
			state.DungeonX, state.DungeonY, direction)
	}
	state.DungeonWallType, _ = grid.WallWrapped(state.DungeonX, state.DungeonY, int(direction))
}

func walkNormalDungeonTo(t *testing.T, state *State, grid *geo.Grid, targetX, targetY int, observer *normalCampaignObserver) {
	t.Helper()
	target := normalDungeonPoint{targetX, targetY}
	initialBlock := uint8(0xFF)
	if state.session != nil {
		initialBlock = state.session.CurrentBlockID()
	}
	for hop := 0; hop < geo.Width*geo.Height*4 && (state.DungeonX != targetX || state.DungeonY != targetY); hop++ {
		if observer.stoppedAtDataPackEvent() {
			return
		}
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
				if _, external := state.dungeonExternalExit(current.x, current.y, uint8(direction)); external {
					continue
				}
				deltaX, deltaY := normalDungeonDelta(direction)
				nextX := geo.WrapCoordinate(current.x+deltaX, geo.Width)
				nextY := geo.WrapCoordinate(current.y+deltaY, geo.Height)
				next := normalDungeonPoint{nextX, nextY}
				// Keep the route on the deterministic story corridor in maps where
				// the known random cells are not needed for the main path. The
				// Zhentil shrine route is deliberately left connected: its prison
				// corridor crosses one of these cells in the original GEO.
				if !(state.GeoMapSet == 4 && state.GeoMapBlock == 0x21) &&
					(next == (normalDungeonPoint{10, 2}) || next == (normalDungeonPoint{8, 11}) ||
						next == (normalDungeonPoint{8, 15}) || next == (normalDungeonPoint{12, 10})) {
					continue
				}
				if _, found := previous[next]; found ||
					(!grid.CanMoveDungeonWrapped(current.x, current.y, direction) &&
						!state.searchEdgeDiscovered(current.x, current.y, uint8(direction))) {
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
			type lockedEdge struct {
				source    normalDungeonPoint
				direction int
				distance  int
			}
			var nextDoor lockedEdge
			hasNextDoor := false
			for y := 0; y < geo.Height; y++ {
				for x := 0; x < geo.Width; x++ {
					if _, reachable := previous[normalDungeonPoint{x, y}]; !reachable {
						continue
					}
					for _, direction := range []int{0, 2, 4, 6} {
						if _, external := state.dungeonExternalExit(x, y, uint8(direction)); external {
							continue
						}
						deltaX, deltaY := normalDungeonDelta(direction)
						nextX := geo.WrapCoordinate(x+deltaX, geo.Width)
						nextY := geo.WrapCoordinate(y+deltaY, geo.Height)
						next := normalDungeonPoint{nextX, nextY}
						if !(state.GeoMapSet == 4 && state.GeoMapBlock == 0x21) &&
							(next == (normalDungeonPoint{10, 2}) || next == (normalDungeonPoint{8, 11}) ||
								next == (normalDungeonPoint{8, 15}) || next == (normalDungeonPoint{12, 10})) {
							continue
						}
						if _, alreadyReachable := previous[next]; alreadyReachable {
							continue
						}
						if grid.CanMoveDungeonWrapped(x, y, direction) {
							continue
						}
						flags, _ := grid.WallDoorFlagsWrapped(x, y, direction)
						if flags != 2 && flags != 3 {
							continue
						}
						distance := nextX - targetX
						if distance < 0 {
							distance = -distance
						}
						distanceY := nextY - targetY
						if distanceY < 0 {
							distanceY = -distanceY
						}
						distance += distanceY
						if !hasNextDoor || distance < nextDoor.distance {
							nextDoor = lockedEdge{
								source:    normalDungeonPoint{x, y},
								direction: direction,
								distance:  distance,
							}
							hasNextDoor = true
						}
					}
				}
			}
			if !hasNextDoor {
				t.Fatalf("normal dungeon target (%d,%d) is unreachable from (%d,%d) and no locked door leads onward",
					targetX, targetY, start.x, start.y)
			}
			current := nextDoor.source
			doorPath := make([]struct {
				point     normalDungeonPoint
				direction int
			}, 0)
			for current != start {
				edge := previous[current]
				doorPath = append(doorPath, struct {
					point     normalDungeonPoint
					direction int
				}{point: current, direction: edge.direction})
				current = edge.point
			}
			for left, right := 0, len(doorPath)-1; left < right; left, right = left+1, right-1 {
				doorPath[left], doorPath[right] = doorPath[right], doorPath[left]
			}
			for _, step := range doorPath {
				observer.resolveDungeonBoundary(t)
				if state.Mode != ModeDungeon {
					t.Fatalf("normal dungeon route to locked door mode=%v message=%q", state.Mode, state.Message)
				}
				deltaX, deltaY := normalDungeonDelta(step.direction)
				if err := state.MoveDungeon(*grid, deltaX, deltaY, step.direction); err != nil {
					t.Fatalf("normal dungeon route to locked door (%d,%d): %v",
						nextDoor.source.x, nextDoor.source.y, err)
				}
				observer.resolveDungeonBoundary(t)
				if observer.stoppedAtDataPackEvent() {
					return
				}
				if state.session != nil && state.session.CurrentBlockID() != initialBlock {
					return
				}
			}
			if state.DungeonX != nextDoor.source.x || state.DungeonY != nextDoor.source.y {
				t.Fatalf("normal dungeon did not reach locked door source (%d,%d), ended at (%d,%d)",
					nextDoor.source.x, nextDoor.source.y, state.DungeonX, state.DungeonY)
			}
			state.TurnDungeonWithGrid(*grid,
				(nextDoor.direction-int(state.DungeonDirection)+8)%8)
			openNormalDungeonDoor(t, state, grid)
			continue
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
		if observer.stoppedAtDataPackEvent() {
			return
		}
		if (observer.stopAtWorldEdge || observer.stopAtMessageID != "") && state.Mode != ModeDungeon {
			return
		}
		if state.Mode != ModeDungeon {
			t.Fatalf("normal dungeon route before hop %d mode=%v message=%q", hop, state.Mode, state.Message)
		}
		deltaX, deltaY := normalDungeonDelta(step.direction)
		if err := state.MoveDungeon(*grid, deltaX, deltaY, step.direction); err != nil {
			t.Fatalf("normal dungeon hop %d toward (%d,%d) from (%d,%d): %v", hop, targetX, targetY,
				state.DungeonX, state.DungeonY, err)
		}
		observer.resolveDungeonBoundary(t)
		if state.session != nil && state.session.CurrentBlockID() != initialBlock {
			return
		}
	}
	if observer.stoppedAtDataPackEvent() {
		return
	}
	if state.DungeonX != targetX || state.DungeonY != targetY {
		t.Fatalf("normal dungeon route did not reach (%d,%d), ended at (%d,%d)",
			targetX, targetY, state.DungeonX, state.DungeonY)
	}
}

func TestRealNewGameContinuesFromHapToMythDrannor(t *testing.T) {
	state := runNormalNewGameToEssembra(t)
	if state == nil {
		return
	}
	selectOption := func(t *testing.T, optionID string) {
		t.Helper()
		if err := state.Select(requireGamePackOptionIndex(t, state, optionID)); err != nil {
			t.Fatalf("select %s: %v", optionID, err)
		}
	}
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Fatal(err)
	}
	defer image.Close()
	// 跨段共用的狀態先宣告：底下每一段是一個 subtest，宣告在 subtest 裡的
	// 變數出不了那個閉包。
	var grid, towerGrid, pitGrid, burialGrid geo.Grid
	var observer *normalCampaignObserver
	// 每一段結束時存一份快照，最後一段之後整批驗往返（SEG-11／SEG-30）。
	snapshotDir := t.TempDir()
	var segmentEnds []campaignSegmentEnd
	captureSegmentEnd := func(t *testing.T, name string) {
		t.Helper()
		path := filepath.Join(snapshotDir,
			strings.NewReplacer("/", "-", " ", "-", "：", "-").Replace(name)+".json")
		if err := state.SavePartyFile(path); err != nil {
			t.Fatalf("段界快照存不下去：%v", err)
		}
		segmentEnds = append(segmentEnds, campaignSegmentEnd{
			name: name, path: path, block: state.session.CurrentBlockID(),
			mode: state.Mode, area: state.Area.GameArea, inDungeon: state.Area.InDungeon,
			party: len(state.PartyFighters()),
		})
	}
	if !t.Run("ECL5/0x31 哈普村", func(t *testing.T) {
		selectOption(t, "ecl-option.journey-on")
		selectOption(t, "ecl-option.hap")
		selectOption(t, "ecl-option.trail")
		selectOption(t, "ecl-option.press-button-or-return-to-continue")
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
		selectOption(t, "ecl-option.enter-city")
		if state.PictureRequested {
			if err := state.Continue(); err != nil {
				t.Fatalf("continue Hap entry picture: %v", err)
			}
		}
		selectOption(t, "ecl-option.press-button-or-return-to-continue")
		if state.Mode != ModeDungeon || state.session.CurrentBlockID() != 0x31 ||
			state.GeoMapSet != 5 || state.GeoMapBlock != 0x32 {
			t.Fatalf("Hap dungeon entry mode=%v block=%#x geo=%d/%#x", state.Mode,
				state.session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock)
		}

		treasureItems, err := ParseTreasureItemBlocks(map[uint8][]byte{
			1: zipData(t, image, "ITEM1.DAX"),
			2: zipData(t, image, "ITEM2.DAX"),
			3: zipData(t, image, "ITEM3.DAX"),
			4: zipData(t, image, "ITEM4.DAX"),
			5: zipData(t, image, "ITEM5.DAX"),
			6: zipData(t, image, "ITEM6.DAX"),
		})
		if err != nil {
			t.Fatal(err)
		}
		state.SetTreasureItemBlocks(treasureItems)
		monster4Blocks, err := dax.Parse(zipData(t, image, "MON4CHA.DAX"))
		if err != nil {
			t.Fatal(err)
		}
		monster4Records := make(map[uint8]monster.Record, len(monster4Blocks))
		for _, block := range monster4Blocks {
			record, parseErr := monster.Parse(block.Data)
			if parseErr != nil {
				t.Fatal(parseErr)
			}
			monster4Records[block.Entry.ID] = record
		}
		state.SetMonsterRecordsForECL(4, monster4Records)
		grid = loadGeo5CampaignGrid(t, image, 0x32)
		towerGrid = loadGeo5CampaignGrid(t, image, 0x33)
		observer = newNormalCampaignObserver(t, state)
		observer.towerGrid = towerGrid
		for _, target := range []normalDungeonPoint{{4, 10}, {9, 10}, {3, 13}, {15, 5}} {
			walkNormalDungeonTo(t, state, &grid, target.x, target.y, observer)
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
		captureSegmentEnd(t, "ECL5/0x31 哈普村")
	}) {
		t.FailNow()
	}
	if !t.Run("ECL5/0x32 古熔岩洞與 ECL5/0x33 巫師塔", func(t *testing.T) {
		walkNormalDungeonTo(t, state, &grid, 9, 10, observer)
		if !observer.seen["lava-tube.guarded-door"] || state.Mode != ModeDungeon {
			t.Fatalf("normal lava guarded-door route mode=%v coverage=%v position=(%d,%d,%d)",
				state.Mode, observer.seen, state.DungeonX, state.DungeonY, state.DungeonDirection)
		}
		walkNormalDungeonTo(t, state, &grid, 0, 5, observer)
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

		// Continue the same player session through the verified WILDERNESS roof
		// exit, Area5 farewell and world route.  This intentionally stops at the
		// Essembra world edge; the optional post-wizard encounter is flag-gated and
		// remains covered by its separate source-oracle regression.
		captureSegmentEnd(t, "ECL5/0x32 古熔岩洞與 ECL5/0x33 巫師塔")
	}) {
		t.FailNow()
	}
	if !t.Run("ECL1/0x50 世界路線：艾森布拉到希爾斯法", func(t *testing.T) {
		observer.towerReturnOption = "ecl-option.wilderness"
		observer.towerReady = false
		observer.stopAtWorldEdge = true
		walkNormalDungeonTo(t, state, &grid, 6, 15, observer)
		area5DepartureSeen := observer.seen["area5.depart-akabar"] || observer.seen["area5.depart-akabar-reluctant"]
		if state.Mode != ModeWilderness ||
			state.Message != requireGamePackText(t, state, "essembra.edge") ||
			!area5DepartureSeen {
			t.Fatalf("normal tower-to-world route mode=%v location=%v block=%#x message=%q coverage=%v position=(%d,%d,%d)",
				state.Mode, state.Location, state.session.CurrentBlockID(), state.Message, observer.seen,
				state.DungeonX, state.DungeonY, state.DungeonDirection)
		}

		// From Essembra, continue along the directed world graph to Hillsfar and
		// stop at its edge. The route choice and destination remain game-pack IDs;
		// the travel encounter, if any, is resolved by the same observer.
		observer.stopAtWorldEdge = false
		observer.stopAtMessageID = "hillsfar.edge"
		observer.nextWorldDestinations = []string{"ecl-option.the-standing-stone", "ecl-option.hillsfar"}
		if !observer.selectOption(t, "ecl-option.journey-on") {
			t.Fatalf("Essembra JOURNEY ON option unavailable: %v", state.currentOriginalChoices)
		}
		observer.resolveDungeonBoundary(t)
		if state.Mode != ModeWilderness || state.Location != LocationHillsfar ||
			state.Message != requireGamePackText(t, state, "hillsfar.edge") {
			t.Fatalf("normal Essembra-to-Hillsfar route mode=%v location=%v block=%#x message=%q choices=%v coverage=%v",
				state.Mode, state.Location, state.session.CurrentBlockID(), state.Message,
				state.currentOriginalChoices, observer.seen)
		}
		captureSegmentEnd(t, "ECL1/0x50 世界路線：艾森布拉到希爾斯法")
	}) {
		t.FailNow()
	}
	if !t.Run("希爾斯法城內", func(t *testing.T) {
		// Enter Hillsfar, resolve the dockside-bar provocation, then leave through
		// the original place menu. This keeps city services and the world edge in
		// the same ECL session instead of switching to a fixture state.
		if !observer.selectOption(t, "ecl-option.enter-city") {
			t.Fatalf("Hillsfar ENTER CITY option unavailable: %v", state.currentOriginalChoices)
		}
		if state.PictureRequested {
			if err := state.Continue(); err != nil {
				t.Fatalf("continue Hillsfar entry picture: %v", err)
			}
		}
		if !observer.selectOption(t, "ecl-option.bar") {
			t.Fatalf("Hillsfar BAR option unavailable: %v", state.currentOriginalChoices)
		}
		if !observer.selectOption(t, "ecl-option.relax") {
			t.Fatalf("Hillsfar RELAX option unavailable: %v", state.currentOriginalChoices)
		}
		if !observer.selectOption(t, "option.no") {
			t.Fatalf("Hillsfar refuse provocation option unavailable: %v", state.currentOriginalChoices)
		}
		for turn := 0; turn < 32 && state.Mode == ModeCombat; turn++ {
			if err := state.CombatAct(); err != nil {
				t.Fatalf("Hillsfar dockside combat turn %d: %v", turn, err)
			}
		}
		if state.Mode != ModeWilderness ||
			state.Message != requireGamePackText(t, state, "hillsfar.dockside-bar") {
			t.Fatalf("Hillsfar dockside victory mode=%v message=%q choices=%v coverage=%v",
				state.Mode, state.Message, state.currentOriginalChoices, observer.seen)
		}
		if !observer.selectOption(t, "ecl-option.exit") {
			t.Fatalf("Hillsfar bar EXIT option unavailable: %v", state.currentOriginalChoices)
		}
		if state.PictureRequested {
			if err := state.Continue(); err != nil {
				t.Fatalf("continue Hillsfar bar exit picture: %v", err)
			}
		}
		if !observer.selectOption(t, "ecl-option.leave") {
			t.Fatalf("Hillsfar places LEAVE option unavailable: %v", state.currentOriginalChoices)
		}
		if state.PictureRequested {
			if err := state.Continue(); err != nil {
				t.Fatalf("continue Hillsfar world-edge picture: %v", err)
			}
		}
		if state.Mode != ModeWilderness ||
			state.Message != requireGamePackText(t, state, "hillsfar.edge") {
			t.Fatalf("Hillsfar city exit mode=%v message=%q choices=%v",
				state.Mode, state.Message, state.currentOriginalChoices)
		}
		captureSegmentEnd(t, "希爾斯法城內")
	}) {
		t.FailNow()
	}
	if !t.Run("ECL1/0x51 世界路線：希爾斯法到猶拉什", func(t *testing.T) {
		observer.stopAtMessageID = "yulash.edge"
		observer.nextWorldDestinations = []string{"ecl-option.yulash"}
		if !observer.selectOption(t, "ecl-option.journey-on") {
			t.Fatalf("Hillsfar JOURNEY ON option unavailable: %v", state.currentOriginalChoices)
		}
		observer.resolveDungeonBoundary(t)
		if state.Mode != ModeWilderness || state.Location != LocationYulash || state.Area.CurrentCity != 10 ||
			state.Message != requireGamePackText(t, state, "yulash.edge") {
			t.Fatalf("normal Hillsfar-to-Yulash route mode=%v location=%v block=%#x message=%q choices=%v coverage=%v",
				state.Mode, state.Location, state.session.CurrentBlockID(), state.Message,
				state.currentOriginalChoices, observer.seen)
		}
		captureSegmentEnd(t, "ECL1/0x51 世界路線：希爾斯法到猶拉什")
	}) {
		t.FailNow()
	}
	if !t.Run("ECL3/0x10 猶拉什：地面與指揮部", func(t *testing.T) {
		// Enter Yulash through the same world-menu destination and ask permission.
		// The guards, spies, commander negotiation and side door are all resolved by
		// stable game-pack option IDs; no story text or flag is injected here.
		if !observer.selectOption(t, "ecl-option.enter-city") {
			t.Fatalf("Yulash ENTER CITY option unavailable: %v", state.currentOriginalChoices)
		}
		if state.PictureRequested {
			if err := state.Continue(); err != nil {
				t.Fatalf("continue Yulash entry picture: %v", err)
			}
		}
		observer.observe()
		if !observer.seen["yulash.entry"] {
			t.Fatalf("Yulash entry message was not observed: %q", state.Message)
		}
		if !observer.selectOption(t, "ecl-option.ask-permission") {
			t.Fatalf("Yulash ASK PERMISSION option unavailable: %v", state.currentOriginalChoices)
		}
		observer.resolveDungeonBoundary(t)
		if state.Mode != ModeDungeon || state.session.CurrentBlockID() != 0x10 ||
			state.GeoMapSet != 3 || state.GeoMapBlock != 0x10 ||
			state.DungeonX != 0 || state.DungeonY != 3 || state.DungeonDirection != 2 {
			t.Fatalf("normal Yulash entry mode=%v block=%#x geo=%d/%#x pos=(%d,%d,%d) coverage=%v",
				state.Mode, state.session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock,
				state.DungeonX, state.DungeonY, state.DungeonDirection, observer.seen)
		}
		for _, messageID := range []string{
			"yulash.riders-burst-out",
			"yulash.checkpoint-halt",
			"yulash.see-commander",
			"yulash.waiting-room",
		} {
			if !observer.seen[messageID] {
				t.Fatalf("normal Yulash entry did not cover %s: %v", messageID, observer.seen)
			}
		}

		yulashGrid := loadGeoCampaignGrid(t, image, 3, "GEO3.DAX", 0x10)
		walkNormalDungeonTo(t, state, &yulashGrid, 1, 3, observer)
		for _, messageID := range []string{
			"yulash.zhentarim-spies",
			"yulash.led-to-commander",
			"yulash.commander-business",
			"journal-trigger.yulash-commander-22",
			"yulash.commander-side-door",
		} {
			if !observer.seen[messageID] {
				t.Fatalf("normal Yulash commander route did not cover %s: %v", messageID, observer.seen)
			}
		}
		openNormalDungeonDoor(t, state, &yulashGrid)
		walkNormalDungeonTo(t, state, &yulashGrid, 11, 0, observer)
		if state.Mode != ModeDungeon || state.session.CurrentBlockID() != 0x10 {
			t.Fatalf("normal Yulash route before Pit exit mode=%v block=%#x pos=(%d,%d,%d) coverage=%v",
				state.Mode, state.session.CurrentBlockID(), state.DungeonX, state.DungeonY,
				state.DungeonDirection, observer.seen)
		}
		if !observer.seen["yulash.pit-entrance"] {
			t.Fatalf("normal Yulash route did not reach Pit entrance: %v", observer.seen)
		}
		if err := state.RunDungeonExitLifecycle(); err != nil {
			t.Fatalf("normal Yulash-to-Pit exit lifecycle: %v", err)
		}
		observer.resolveDungeonBoundary(t)
		if state.Mode != ModeDungeon || state.session.CurrentBlockID() != 0x11 ||
			state.GeoMapSet != 3 || state.GeoMapBlock != 0x11 ||
			state.DungeonX != 0 || state.DungeonY != 0 || state.DungeonDirection != 2 {
			t.Fatalf("normal Pit entry mode=%v block=%#x geo=%d/%#x pos=(%d,%d,%d) coverage=%v",
				state.Mode, state.session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock,
				state.DungeonX, state.DungeonY, state.DungeonDirection, observer.seen)
		}
		for _, messageID := range []string{
			"pit.opening-dead-cultists",
			"pit.opening-chosen",
			"pit.trapped",
			"pit.cleric-dies",
			"pit.ambience",
		} {
			if !observer.seen[messageID] {
				t.Fatalf("normal Pit entry did not cover %s: %v", messageID, observer.seen)
			}
		}
		captureSegmentEnd(t, "ECL3/0x10 猶拉什：地面與指揮部")
	}) {
		t.FailNow()
	}
	if !t.Run("ECL3/0x11 猶拉什：地下第一層", func(t *testing.T) {
		pitGrid = loadGeoCampaignGrid(t, image, 3, "GEO3.DAX", 0x11)
		walkNormalDungeonTo(t, state, &pitGrid, 1, 4, observer)
		for _, messageID := range []string{
			"pit.alias-dragonbait-meet",
			"pit.alias-bonded-reaction",
			"pit.alias-dragonbait-introduction",
			"journal-trigger.alias-story-3",
			"pit.alias-dragonbait-join",
			"pit.alias-dragonbait-joined",
		} {
			if !observer.seen[messageID] {
				t.Fatalf("normal Pit Alias route did not cover %s: %v", messageID, observer.seen)
			}
		}
		if len(state.partyRoster) != 3 || state.partyRoster[1].ScriptName != "ALIAS" ||
			state.partyRoster[2].ScriptName != "DRAGONBAIT" {
			t.Fatalf("normal Pit Alias route party=%#v", state.partyRoster)
		}

		walkNormalDungeonTo(t, state, &pitGrid, 15, 11, observer)
		if state.Mode != ModeDungeon || state.session.CurrentBlockID() != 0x12 ||
			state.DungeonX != 15 || state.DungeonY != 14 || state.DungeonDirection != 4 {
			t.Fatalf("normal Pit stairs-down mode=%v block=%#x pos=(%d,%d,%d) coverage=%v",
				state.Mode, state.session.CurrentBlockID(), state.DungeonX, state.DungeonY,
				state.DungeonDirection, observer.seen)
		}
		captureSegmentEnd(t, "ECL3/0x11 猶拉什：地下第一層")
	}) {
		t.FailNow()
	}
	if !t.Run("ECL3/0x12 猶拉什：地下第二層", func(t *testing.T) {
		walkNormalDungeonTo(t, state, &pitGrid, 11, 3, observer)
		for _, messageID := range []string{
			"pit.stairs-down",
			"pit.mogion-altar",
			"pit.alias-identifies-mogion",
			"pit.mogion-greeting",
			"pit.bond-paralysis",
			"pit.alias-dragonbait-tendrils",
			"pit.mogion-ritual",
			"pit.dimensional-window",
			"pit.moander-returns",
			"pit.bond-fades",
			"pit.bond-broken",
			"pit.alias-attack-mogion",
			"pit.rift-closes",
			"pit.remnants-scream",
			"pit.remnants-attack",
			"pit.gauntlet",
			"pit.priest-flees",
		} {
			if !observer.seen[messageID] {
				t.Fatalf("normal Pit Mogion route did not cover %s: %v", messageID, observer.seen)
			}
		}

		walkNormalDungeonTo(t, state, &pitGrid, 12, 0, observer)
		if err := state.SearchDungeonLocation(); err != nil {
			t.Fatalf("normal Pit altar search: %v", err)
		}
		observer.resolveDungeonBoundary(t)
		if !observer.seen["pit.altar-treasure"] || state.Mode != ModeDungeon {
			t.Fatalf("normal Pit altar search mode=%v message=%q choices=%v coverage=%v",
				state.Mode, state.Message, state.currentOriginalChoices, observer.seen)
		}

		walkNormalDungeonTo(t, state, &pitGrid, 15, 14, observer)
		if state.Mode != ModeDungeon || state.session.CurrentBlockID() != 0x11 ||
			state.DungeonX != 0 || state.DungeonY != 0 || state.DungeonDirection != 2 {
			t.Fatalf("normal Pit stairs-up mode=%v block=%#x pos=(%d,%d,%d) coverage=%v",
				state.Mode, state.session.CurrentBlockID(), state.DungeonX, state.DungeonY,
				state.DungeonDirection, observer.seen)
		}
		captureSegmentEnd(t, "ECL3/0x12 猶拉什：地下第二層")
	}) {
		t.FailNow()
	}
	if !t.Run("ECL3/0x11 返回與猶拉什邊界", func(t *testing.T) {
		observer.stopAtMessageID = "yulash.edge"
		walkNormalDungeonTo(t, state, &pitGrid, 0, 12, observer)
		// The last-stand encounter occupies (0,12); the actual Yulash boundary
		// handler is the adjacent (0,11,W) cell after that battle returns.
		walkNormalDungeonTo(t, state, &pitGrid, 0, 11, observer)
		state.TurnDungeonWithGrid(pitGrid, (6-int(state.DungeonDirection)+8)%8)
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatalf("normal Pit final exit lifecycle: %v", err)
		}
		observer.resolveDungeonBoundary(t)
		if state.Mode != ModeWilderness || state.Location != LocationYulash ||
			state.Message != requireGamePackText(t, state, "yulash.edge") ||
			!observer.seen["pit.exit-last-stand"] {
			current, currentOK := state.session.MemoryValue(0x4C9B)
			destination, destinationOK := state.session.MemoryValue(0x4C9C)
			t.Fatalf("normal Pit exit mode=%v location=%v city=%d block=%#x message=%q 4C9B=%#x/%v 4C9C=%#x/%v choices=%v coverage=%v",
				state.Mode, state.Location, state.Area.CurrentCity, state.session.CurrentBlockID(), state.Message,
				current, currentOK, destination, destinationOK, state.currentOriginalChoices, observer.seen)
		}
		captureSegmentEnd(t, "ECL3/0x11 返回與猶拉什邊界")
	}) {
		t.FailNow()
	}
	if !t.Run("ECL4/0x20 散提爾堡內城", func(t *testing.T) {
		// Continue the same session through Zhentil Keep.  The destination queue is
		// expressed as game-pack option IDs; the observer still has to consume the
		// ordinary JOURNEY ON and TRAIL menus before the city patrol appears.
		observer.stopAtMessageID = ""
		observer.nextWorldDestinations = []string{"ecl-option.zhentil-keep", "ecl-option.trail"}
		observer.resolveDungeonBoundary(t)
		if state.Mode != ModeDungeon || state.session.CurrentBlockID() != 0x20 ||
			state.GeoMapSet != 4 || state.GeoMapBlock != 0x20 ||
			state.DungeonX != 2 || state.DungeonY != 0 || state.DungeonDirection != 4 {
			route := []uint16{}
			for address := uint16(0x4C02); address <= 0x4C05; address++ {
				value, _ := state.session.MemoryValue(address)
				route = append(route, value)
			}
			t.Fatalf("normal Zhentil entry mode=%v city=%d block=%#x geo=%d/%#x pos=(%d,%d,%d) route=%v message=%q choices=%v coverage=%v",
				state.Mode, state.Area.CurrentCity, state.session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock,
				state.DungeonX, state.DungeonY, state.DungeonDirection, route, state.Message,
				state.currentOriginalChoices, observer.seen)
		}
		zhentilGrid := loadGeoCampaignGrid(t, image, 4, "GEO4.DAX", 0x20)
		walkNormalDungeonTo(t, state, &zhentilGrid, 10, 11, observer)
		// The ECL cell contract is position plus facing: Olive appears at
		// (10,11,N).  Approach it from the southern neighbor so the normal move,
		// rather than a direct coordinate assignment or same-cell redraw, supplies
		// the north-facing trigger.
		walkNormalDungeonTo(t, state, &zhentilGrid, 10, 12, observer)
		state.TurnDungeonWithGrid(zhentilGrid, (0-int(state.DungeonDirection)+8)%8)
		if err := state.MoveDungeon(zhentilGrid, 0, -1, 0); err != nil {
			t.Fatalf("normal Zhentil Olive north-facing approach: %v", err)
		}
		observer.resolveDungeonBoundary(t)
		for _, messageID := range []string{
			"zhentil.olive_appears",
			"zhentil.olive_follow",
			"zhentil.dark_shrine_entry",
			"zhentil.olive_explains",
			"zhentil.dimswart_door",
			"zhentil.olive_leaves",
		} {
			if !observer.seen[messageID] {
				guard, _ := state.session.MemoryValue(0x7F81)
				t.Fatalf("normal Zhentil Olive route did not cover %s: mode=%v block=%#x geo=%d/%#x pos=(%d,%d,%d) roof=%#x wall=%#x guard=%#x message=%q choices=%v coverage=%v", messageID, state.Mode, state.session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock, state.DungeonX, state.DungeonY, state.DungeonDirection, state.DungeonWallRoof, state.DungeonWallType, guard, state.Message, state.currentOriginalChoices, observer.seen)
			}
		}
		if state.Mode != ModeDungeon || state.session.CurrentBlockID() != 0x21 ||
			state.GeoMapSet != 4 || state.GeoMapBlock != 0x21 {
			t.Fatalf("normal Dark Shrine transition mode=%v block=%#x geo=%d/%#x pos=(%d,%d,%d) coverage=%v",
				state.Mode, state.session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock,
				state.DungeonX, state.DungeonY, state.DungeonDirection, observer.seen)
		}
		captureSegmentEnd(t, "ECL4/0x20 散提爾堡內城")
	}) {
		t.FailNow()
	}
	if !t.Run("ECL4/0x21 散提爾堡：神殿與牢房", func(t *testing.T) {
		shrineGrid := loadGeoCampaignGrid(t, image, 4, "GEO4.DAX", 0x21)
		walkNormalDungeonTo(t, state, &shrineGrid, 6, 13, observer)
		state.TurnDungeonWithGrid(shrineGrid, (0-int(state.DungeonDirection)+8)%8)
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatalf("normal Dimswart cell hint lifecycle: %v", err)
		}
		observer.resolveDungeonBoundary(t)
		walkNormalDungeonTo(t, state, &shrineGrid, 2, 14, observer)
		state.TurnDungeonWithGrid(shrineGrid, (0-int(state.DungeonDirection)+8)%8)
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatalf("normal Dimswart cell lifecycle: %v", err)
		}
		observer.resolveDungeonBoundary(t)
		for _, messageID := range []string{
			"zhentil.dimswart_appears",
			"zhentil.dimswart_join",
		} {
			if !observer.seen[messageID] {
				t.Fatalf("normal Dimswart route did not cover %s: %v", messageID, observer.seen)
			}
		}
		walkNormalDungeonTo(t, state, &shrineGrid, 4, 12, observer)
		state.TurnDungeonWithGrid(shrineGrid, (0-int(state.DungeonDirection)+8)%8)
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatalf("normal hooded woman lifecycle: %v", err)
		}
		observer.resolveDungeonBoundary(t)
		for _, messageID := range []string{
			"zhentil.hooded_offer",
			"zhentil.hooded_follow",
			"zhentil.fzoul_interrupts",
			"zhentil.fzoul_retreats",
			"dexam.arrival",
			"dexam.journal_30",
			"dexam.amulet_choice",
			"dexam.fzoul_journal_7",
			"dexam.kills_fzoul",
			"dexam.fzoul_bond_fades",
			"dexam.kill_order",
			"dexam.amulet_rises",
			"dexam.altar_melee",
		} {
			if !observer.seen[messageID] {
				boundary, _ := state.session.MemoryValue(0x7ED5)
				forcedMove, _ := state.session.MemoryValue(0x7EC9)
				previousBlock, _ := state.session.MemoryValue(0x4BF2)
				t.Fatalf("normal Zhentil Shrine route did not cover %s: mode=%v block=%#x geo=%d/%#x pos=(%d,%d,%d) roof=%#x wall=%#x 7ED5=%#x 7EC9=%#x 4BF2=%#x message=%q coverage=%v", messageID, state.Mode, state.session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock, state.DungeonX, state.DungeonY, state.DungeonDirection, state.DungeonWallRoof, state.DungeonWallType, boundary, forcedMove, previousBlock, state.Message, observer.seen)
			}
		}
		if state.Mode != ModeDungeon || state.session.CurrentBlockID() != 0x22 ||
			state.GeoMapSet != 4 || state.GeoMapBlock != 0x25 {
			t.Fatalf("normal Beholder Cave transition mode=%v block=%#x geo=%d/%#x pos=(%d,%d,%d) coverage=%v",
				state.Mode, state.session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock,
				state.DungeonX, state.DungeonY, state.DungeonDirection, observer.seen)
		}
		// E1 is the original Cave of the Beholder entrance. It is a distinct
		// player-visible anchor from the later dead-elf teleporter destination;
		// the normal ECL transaction below connects the two in a separate step.
		if state.DungeonX != 5 || state.DungeonY != 7 || state.DungeonDirection != 6 {
			t.Fatalf("normal Beholder Cave spawn mode=%v block=%#x geo=%d/%#x pos=(%d,%d,%d) coverage=%v",
				state.Mode, state.session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock,
				state.DungeonX, state.DungeonY, state.DungeonDirection, observer.seen)
		}
		captureSegmentEnd(t, "ECL4/0x21 散提爾堡：神殿與牢房")
	}) {
		t.FailNow()
	}
	if !t.Run("ECL4/0x22 眼魔洞穴與 Dexam", func(t *testing.T) {
		caveGrid := loadGeoCampaignGrid(t, image, 4, "GEO4.DAX", 0x25)
		observer.stopAtDataPackEventID = "zhentil-keep.beholder-cave.same-block-launch"
		// The source cell is reached by ordinary GEO movement.  The ECL transaction
		// then writes the destination through C04B/C04C/C04D; the title pack only
		// projects that original virtual-map handoff back into State.
		walkNormalDungeonTo(t, state, &caveGrid, 5, 9, observer)
		if !observer.stoppedAtDataPackEvent() || state.Mode != ModeWilderness ||
			state.DungeonX != 13 || state.DungeonY != 1 || state.DungeonDirection != 6 ||
			state.DungeonWallType != 8 || state.DungeonWallRoof != 0xC0 {
			t.Fatalf("normal Beholder Cave teleporter mode=%v pos=(%d,%d,%d) wall=%#x roof=%#x applied=%v coverage=%v",
				state.Mode, state.DungeonX, state.DungeonY, state.DungeonDirection,
				state.DungeonWallType, state.DungeonWallRoof,
				state.appliedDataPackEvents[observer.stopAtDataPackEventID], observer.seen)
		}
		for address, want := range map[uint16]uint16{
			0xC04B: 13, 0xC04C: 1, 0xC04D: 3, 0xC04E: 8, 0xC04F: 0xC0, 0x4C03: 0,
		} {
			got, ok := state.session.MemoryValue(address)
			if !ok || got != want {
				t.Fatalf("normal Beholder Cave teleporter memory[%#x]=%#x ok=%v, want %#x", address, got, ok, want)
			}
		}
		if guard, ok := state.session.MemoryValue(0x7F81); !ok || guard != 0 {
			t.Fatalf("normal Beholder Cave teleporter guard=%#x ok=%v, want cleared", guard, ok)
		}
		if state.Message != requireGamePackText(t, state, "dexam.dead-elf.remains") ||
			!observer.hasOption("dexam.dead-elf.examine-remains") || !observer.hasOption("ecl-option.leave") {
			t.Fatalf("normal Beholder Cave dead-elf prompt message=%q choices=%v originals=%v",
				state.Message, state.Choices, state.currentOriginalChoices)
		}
		selectOption(t, "dexam.dead-elf.examine-remains")
		if state.Mode != ModeWilderness || state.Message != requireGamePackText(t, state, "dexam.dead-elf.pouch") ||
			!observer.hasOption("ecl-option.press-button-or-return-to-continue") {
			t.Fatalf("normal Beholder Cave pouch discovery mode=%v message=%q choices=%v",
				state.Mode, state.Message, state.currentOriginalChoices)
		}
		selectOption(t, "ecl-option.press-button-or-return-to-continue")
		if state.Mode != ModeWilderness || !observer.hasOption("dexam.dead-elf.pick-up-pouch") ||
			!observer.hasOption("dexam.dead-elf.poke-pouch") || !observer.hasOption("dexam.dead-elf.find-trap") ||
			!observer.hasOption("ecl-option.leave") {
			t.Fatalf("normal Beholder Cave pouch menu mode=%v message=%q choices=%v",
				state.Mode, state.Message, state.currentOriginalChoices)
		}
		selectOption(t, "dexam.dead-elf.pick-up-pouch")
		if state.Mode != ModeWilderness || state.Message != requireGamePackText(t, state, "dexam.dead-elf.gas-trap") ||
			!observer.hasOption("ecl-option.press-button-or-return-to-continue") {
			t.Fatalf("normal Beholder Cave gas-trap warning mode=%v message=%q choices=%v",
				state.Mode, state.Message, state.currentOriginalChoices)
		}
		selectOption(t, "ecl-option.press-button-or-return-to-continue")
		if state.Mode != ModeWilderness || state.Message != requireGamePackText(t, state, "dexam.dead-elf.map") ||
			!observer.hasOption("ecl-option.press-button-or-return-to-continue") {
			t.Fatalf("normal Beholder Cave Journal 59 mode=%v message=%q choices=%v",
				state.Mode, state.Message, state.currentOriginalChoices)
		}
		if !slices.Contains(state.JournalPages, requireGamePackText(t, state, "journal.59")) {
			t.Fatalf("normal Beholder Cave Journal 59 was not unlocked: %v", state.JournalPages)
		}
		for address, want := range map[uint16]uint16{0x4C03: 0, 0x4C07: 0x80} {
			if got, ok := state.session.MemoryValue(address); !ok || got != want {
				t.Fatalf("normal Beholder Cave raw memory[%#x]=%#x ok=%v, want %#x", address, got, ok, want)
			}
		}
		selectOption(t, "ecl-option.press-button-or-return-to-continue")
		treasureExit, hasTreasureExit := state.OriginalChoiceIndex("TREASURE_EXIT")
		if state.Mode != ModeWilderness || !state.treasureMenu || state.CombatActive() ||
			len(state.PendingTreasureItems()) != 2 ||
			!hasTreasureExit {
			t.Fatalf("normal Beholder Cave post-map treasure boundary mode=%v treasure=%v combat=%v event=%q message=%q choices=%v items=%v",
				state.Mode, state.treasureMenu, state.CombatActive(), state.OriginalEvent, state.Message,
				state.currentOriginalChoices, state.PendingTreasureItems())
		}
		if err := state.Select(treasureExit); err != nil {
			t.Fatalf("normal Beholder Cave leave Journal 59 treasure: %v", err)
		}
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatalf("normal Beholder Cave lifecycle after Journal 59 treasure: %v", err)
		}
		if state.Mode != ModeDungeon || state.CombatActive() || state.treasureMenu ||
			len(state.PendingTreasureItems()) != 0 || state.Message != "" || len(state.currentOriginalChoices) != 0 ||
			state.DungeonX != 13 || state.DungeonY != 1 || state.DungeonDirection != 6 {
			t.Fatalf("normal Beholder Cave post-treasure lifecycle mode=%v combat=%v treasure=%v message=%q choices=%v items=%v pos=(%d,%d,%d)",
				state.Mode, state.CombatActive(), state.treasureMenu, state.Message, state.currentOriginalChoices,
				state.PendingTreasureItems(), state.DungeonX, state.DungeonY, state.DungeonDirection)
		}

		// Journal 59 supplies the player-visible topology; GEO4/25 independently
		// proves that (14,1,E) is the sole detail-zero wall isolating Dexam.  The
		// original SEARCH/LOOK writer is not yet exact, so the game-pack keeps this
		// executable reconstruction explicitly graded as strong inference.
		observer.stopAtDataPackEventID = ""
		walkNormalDungeonTo(t, state, &caveGrid, 14, 1, observer)
		if value, _ := state.session.MemoryValue(0x4C03); value != 0 {
			t.Fatalf("normal Beholder Cave route to Dexam wall changed 4C03=%#x", value)
		}
		state.TurnDungeonWithGrid(caveGrid, (2-int(state.DungeonDirection)+8)%8)
		if state.CanMoveDungeon(caveGrid, 1, 0, 2) {
			t.Fatal("normal Beholder Cave Dexam wall was passable before LOOK")
		}
		if err := state.LookDungeonLocation(); err != nil {
			t.Fatalf("normal Beholder Cave LOOK at Dexam wall: %v", err)
		}
		observer.resolveDungeonBoundary(t)
		if value, _ := state.session.MemoryValue(0x4C03); value != 0 {
			t.Fatalf("normal Beholder Cave LOOK changed 4C03=%#x", value)
		}
		if !state.CanMoveDungeon(caveGrid, 1, 0, 2) {
			t.Fatalf("normal Beholder Cave LOOK did not reveal Dexam wall: edges=%v",
				state.dungeonSearchEdgeIDs())
		}
		if err := state.MoveDungeon(caveGrid, 1, 0, 2); err != nil {
			t.Fatalf("normal Beholder Cave enter Dexam chamber: %v", err)
		}
		if value, _ := state.session.MemoryValue(0x4C03); value != 1 {
			t.Fatalf("normal Beholder Cave Dexam handler guard=%#x, want 1", value)
		}
		// The original fixed encounter is keyed to the chamber's north-facing
		// presentation. Entering from the west is normal movement; turning north
		// is a second ordinary player action, not a direct ECL entry.
		if state.Mode == ModeDungeon {
			state.TurnDungeonWithGrid(caveGrid, (0-int(state.DungeonDirection)+8)%8)
			if err := state.RunDungeonLifecycle(); err != nil {
				t.Fatalf("normal Beholder Cave face Dexam encounter: %v", err)
			}
		}
		observer.resolveDungeonBoundary(t)
		for _, messageID := range []string{
			"dexam.final_reveal",
			"dexam.attack",
			"dexam.amulet_retrieved",
			"dexam.zhentil_attack",
		} {
			if !observer.seen[messageID] {
				guard, _ := state.session.MemoryValue(0x7F81)
				terrain, _ := state.session.MemoryValue(0xC04F)
				flag03, _ := state.session.MemoryValue(0x4C03)
				flag07, _ := state.session.MemoryValue(0x4C07)
				t.Fatalf("normal Beholder Cave Dexam route did not cover %s: mode=%v pos=(%d,%d,%d) wall=%#x roof=%#x C04F=%#x 7F81=%#x 4C03=%#x 4C07=%#x coverage=%v",
					messageID, state.Mode, state.DungeonX, state.DungeonY,
					state.DungeonDirection, state.DungeonWallType, state.DungeonWallRoof,
					terrain, guard, flag03, flag07, observer.seen)
			}
		}
		if state.Mode != ModeDungeon || state.session.CurrentBlockID() != 0x22 ||
			state.DungeonX != 15 || state.DungeonY != 1 {
			t.Fatalf("normal Beholder Cave after Dexam mode=%v block=%#x pos=(%d,%d,%d) coverage=%v",
				state.Mode, state.session.CurrentBlockID(), state.DungeonX, state.DungeonY,
				state.DungeonDirection, observer.seen)
		}
		if !state.CanMoveDungeon(caveGrid, -1, 0, 6) {
			t.Fatalf("normal Beholder Cave cannot return through revealed Dexam wall: edges=%v",
				state.dungeonSearchEdgeIDs())
		}

		// The original Chinese Journal 59 map draws a second door on the east side
		// of Dexam's shrine. GEO4/25 stores it as wall 09/detail 0 at the wrapped
		// (15,1,E)->(0,1,W) edge; after revealing it, every remaining edge on the
		// route to terrain 93 is an ordinary door, archway, or open corridor.
		state.TurnDungeonWithGrid(caveGrid, (2-int(state.DungeonDirection)+8)%8)
		if state.CanMoveDungeon(caveGrid, 1, 0, 2) {
			t.Fatal("normal Beholder Cave shrine east wall was passable before LOOK")
		}
		if err := state.LookDungeonLocation(); err != nil {
			t.Fatalf("normal Beholder Cave LOOK at shrine east wall: %v", err)
		}
		observer.resolveDungeonBoundary(t)
		if !state.CanMoveDungeon(caveGrid, 1, 0, 2) {
			t.Fatalf("normal Beholder Cave LOOK did not reveal shrine east wall: edges=%v",
				state.dungeonSearchEdgeIDs())
		}
		if err := state.MoveDungeon(caveGrid, 1, 0, 2); err != nil {
			t.Fatalf("normal Beholder Cave leave shrine through east wall: %v", err)
		}
		observer.resolveDungeonBoundary(t)
		if state.DungeonX != 0 || state.DungeonY != 1 {
			t.Fatalf("normal Beholder Cave shrine east wrap ended at (%d,%d), want (0,1)",
				state.DungeonX, state.DungeonY)
		}
		for stepIndex, direction := range []int{4, 2, 2, 4, 4, 2, 2, 2, 2, 2, 0, 6} {
			observer.resolveDungeonBoundary(t)
			if state.Mode != ModeDungeon {
				t.Fatalf("normal Beholder Cave exit path step %d mode=%v message=%q",
					stepIndex+1, state.Mode, state.Message)
			}
			state.TurnDungeonWithGrid(caveGrid,
				(direction-int(state.DungeonDirection)+8)%8)
			deltaX, deltaY := normalDungeonDelta(direction)
			if !state.CanMoveDungeon(caveGrid, deltaX, deltaY, direction) {
				flags, _ := caveGrid.WallDoorFlagsWrapped(
					state.DungeonX, state.DungeonY, direction)
				if flags != 2 && flags != 3 {
					t.Fatalf("normal Beholder Cave exit path step %d blocked at (%d,%d,%d), flags=%#x",
						stepIndex+1, state.DungeonX, state.DungeonY, direction, flags)
				}
				openNormalDungeonDoor(t, state, &caveGrid)
			}
			if err := state.MoveDungeon(caveGrid, deltaX, deltaY, direction); err != nil {
				t.Fatalf("normal Beholder Cave exit path step %d from (%d,%d,%d): %v",
					stepIndex+1, state.DungeonX, state.DungeonY, direction, err)
			}
			observer.resolveDungeonBoundary(t)
		}
		if state.Mode != ModeDungeon || state.DungeonX != 6 || state.DungeonY != 3 ||
			state.DungeonWallRoof != 0x93 {
			t.Fatalf("normal Beholder Cave exit route mode=%v pos=(%d,%d,%d) terrain=%#x message=%q",
				state.Mode, state.DungeonX, state.DungeonY, state.DungeonDirection,
				state.DungeonWallRoof, state.Message)
		}
		state.TurnDungeonWithGrid(caveGrid, (2-int(state.DungeonDirection)+8)%8)
		if err := state.RunDungeonExitLifecycle(); err != nil {
			t.Fatalf("normal Beholder Cave exit lifecycle: %v", err)
		}
		observer.observe()
		if state.Message != requireGamePackText(t, state, "dexam.departure.olive") {
			t.Fatalf("normal Beholder Cave departure Olive message=%q", state.Message)
		}
		if err := state.Continue(); err != nil {
			t.Fatalf("continue normal Beholder Cave departure picture: %v", err)
		}
		for _, messageID := range []string{
			"dexam.departure.dimswart",
			"dexam.departure.gharri",
			"dexam.departure.riders",
		} {
			if err := state.Select(0); err != nil {
				t.Fatalf("select normal Beholder Cave departure %s: %v", messageID, err)
			}
			observer.observe()
			if state.Message != requireGamePackText(t, state, messageID) {
				t.Fatalf("normal Beholder Cave departure %s message=%q", messageID, state.Message)
			}
		}
		if err := state.Select(0); err != nil {
			t.Fatalf("leave normal Beholder Cave departure scene: %v", err)
		}
		if err := state.Select(0); err != nil {
			t.Fatalf("continue normal Beholder Cave world handoff: %v", err)
		}
		observer.observe()
		if state.Mode != ModeEvent || state.session.CurrentBlockID() != 0x51 ||
			state.Area.InDungeon || state.Area.GameArea != 1 ||
			state.Message != requireGamePackText(t, state, "zhentil.edge") {
			t.Fatalf("normal Beholder Cave world return mode=%v block=%#x area=%+v message=%q",
				state.Mode, state.session.CurrentBlockID(), state.Area, state.Message)
		}
		if err := state.Continue(); err != nil {
			t.Fatalf("continue normal Beholder Cave Shadowdale edge: %v", err)
		}
		if state.Mode != ModeWilderness ||
			!observer.hasOption("ecl-option.enter-city") ||
			!observer.hasOption("ecl-option.journey-on") {
			t.Fatalf("normal Beholder Cave Shadowdale menu mode=%v choices=%v message=%q",
				state.Mode, state.currentOriginalChoices, state.Message)
		}
		captureSegmentEnd(t, "ECL4/0x22 眼魔洞穴與 Dexam")
	}) {
		t.FailNow()
	}
	if !t.Run("ECL1/0x50 立石群：灰袍男子", func(t *testing.T) {
		// 散提爾堡的世界圖鄰居只有帖許瓦／猶拉什／費蘭，要繞經猶拉什與希爾斯法
		// 才接得到立石群。
		observer.stopAtMessageID = ""
		observer.stopAtWorldEdge = false
		observer.nextWorldDestinations = []string{
			"ecl-option.yulash", "ecl-option.hillsfar", "ecl-option.the-standing-stone",
		}
		if !observer.selectOption(t, "ecl-option.journey-on") {
			t.Fatalf("散提爾堡 JOURNEY ON 選項不在：%v", state.currentOriginalChoices)
		}
		observer.resolveDungeonBoundary(t)
		if state.Location != LocationStandingStone || state.Area.CurrentCity != 4 {
			t.Fatalf("走到立石群失敗：location=%v city=%d block=%#02x",
				state.Location, state.Area.CurrentCity, state.session.CurrentBlockID())
		}
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatalf("立石群生命週期：%v", err)
		}
		observer.observe()
		if state.Message != requireGamePackText(t, state, "standing-stone.grey-man") {
			t.Fatalf("立石群開場=%q", state.Message)
		}
		if !observer.selectOption(t, "ecl-option.press-button-or-return-to-continue") {
			t.Fatalf("立石群開場沒有繼續選項：%v", state.currentOriginalChoices)
		}
		observer.observe()
		// 灰袍男子按「還剩幾位主人」換一種說法，數字是執行期插進同一頁的，
		// 所以每一種數字是一條獨立的文字規則。這條路徑走到這裡剩兩位。
		if state.Message != requireGamePackText(t, state, "standing-stone.two-masters") {
			t.Fatalf("立石群灰袍男子的台詞=%q choices=%v", state.Message, state.currentOriginalChoices)
		}
		if !observer.selectOption(t, "ecl-option.thank-him") {
			t.Fatalf("立石群 THANK HIM 選項不在：%v", state.currentOriginalChoices)
		}
		observer.observe()
		if state.Message != requireGamePackText(t, state, "standing-stone.seek-red") {
			t.Fatalf("立石群指路台詞=%q", state.Message)
		}
		if !observer.selectOption(t, "ecl-option.press-button-or-return-to-continue") {
			t.Fatalf("立石群指路沒有繼續選項：%v", state.currentOriginalChoices)
		}
		observer.observe()
		if !observer.hasOption("ecl-option.journey-on") {
			t.Fatalf("立石群結束之後沒有回到世界選單：mode=%v choices=%v",
				state.Mode, state.currentOriginalChoices)
		}
		captureSegmentEnd(t, "ECL1/0x50 立石群：灰袍男子")
	}) {
		t.FailNow()
	}

	if !t.Run("密斯卓諾：世界路線", func(t *testing.T) {
		observer.stopAtMessageID = "myth-drannor.edge"
		observer.nextWorldDestinations = []string{"ecl-option.myth-drannor"}
		if !observer.selectOption(t, "ecl-option.journey-on") {
			t.Fatalf("立石群 JOURNEY ON 選項不在：%v", state.currentOriginalChoices)
		}
		observer.resolveDungeonBoundary(t)
		if state.Location != LocationMythDrannor || state.Area.CurrentCity != 13 ||
			state.Message != requireGamePackText(t, state, "myth-drannor.edge") {
			t.Fatalf("走到密斯卓諾邊緣失敗：location=%v city=%d block=%#02x message=%q",
				state.Location, state.Area.CurrentCity, state.session.CurrentBlockID(), state.Message)
		}
		captureSegmentEnd(t, "密斯卓諾：世界路線")
	}) {
		t.FailNow()
	}

	if !t.Run("ECL6/0x40 密斯卓諾：墓園", func(t *testing.T) {
		observer.stopAtMessageID = ""
		if !observer.selectOption(t, "ecl-option.enter-city") {
			t.Fatalf("密斯卓諾 ENTER CITY 選項不在：%v", state.currentOriginalChoices)
		}
		if state.session.CurrentBlockID() != 0x40 || state.GeoMapSet != 6 || state.GeoMapBlock != 0x40 {
			t.Fatalf("進遺跡之後 block=%#02x geo=%d/%#02x", state.session.CurrentBlockID(),
				state.GeoMapSet, state.GeoMapBlock)
		}
		observer.observe()
		if state.Message != requireGamePackText(t, state, "myth-drannor.helm-north") {
			t.Fatalf("墓園入口台詞=%q", state.Message)
		}
		captureSegmentEnd(t, "ECL6/0x40 密斯卓諾：墓園")
	}) {
		t.FailNow()
	}

	if !t.Run("ECL6/0x40 墓園：進入遺跡", func(t *testing.T) {
		if err := state.Continue(); err != nil {
			t.Fatalf("龍盔台詞之後推不動：%v", err)
		}
		observer.observe()
		// 原作把出生點寫進 C04B／C04C／C04D，remake 從那裡同步回來。
		if state.Mode != ModeDungeon || state.DungeonX != 2 || state.DungeonY != 15 ||
			state.DungeonDirection != 2 {
			t.Fatalf("墓園出生點 mode=%v pos=(%d,%d,%d)", state.Mode,
				state.DungeonX, state.DungeonY, state.DungeonDirection)
		}
		burialGrid = loadGeoCampaignGrid(t, image, 6, "GEO6.DAX", 0x40)
		state.DungeonWallType, _ = burialGrid.WallWrapped(state.DungeonX, state.DungeonY,
			int(state.DungeonDirection))
		state.DungeonWallRoof = burialGrid.CellWrapped(state.DungeonX, state.DungeonY).Terrain
		captureSegmentEnd(t, "ECL6/0x40 墓園：進入遺跡")
	}) {
		t.FailNow()
	}

	if !t.Run("ECL6/0x40 墓園：紅網", func(t *testing.T) {
		// ⚠ x=3 那一整欄（地形 0x01）是墓園的離場格：踩上去隊伍會被送回世界
		// 地圖。往東要繞 y=12 那一列。
		observer.stopAtMessageID = "myth-drannor.red-web"
		for _, target := range []normalDungeonPoint{{2, 13}, {2, 12}, {6, 12}, {6, 14}} {
			walkNormalDungeonTo(t, state, &burialGrid, target.x, target.y, observer)
			if state.Mode != ModeDungeon && state.Mode != ModeWilderness {
				t.Fatalf("走到 (%d,%d) 時模式是 %v", target.x, target.y, state.Mode)
			}
		}
		observer.observe()
		if state.Message != requireGamePackText(t, state, "myth-drannor.red-web") {
			t.Fatalf("紅網台詞=%q pos=(%d,%d,%d) choices=%v", state.Message,
				state.DungeonX, state.DungeonY, state.DungeonDirection,
				state.currentOriginalChoices)
		}
		// 說出通關語只會讓網更亮（`red-web.brighter`），原作的解法是砍。
		if !observer.selectOption(t, "ecl-option.hack-it") {
			t.Fatalf("紅網 HACK IT 選項不在：%v", state.currentOriginalChoices)
		}
		observer.observe()
		if state.Message != requireGamePackText(t, state, "myth-drannor.red-web.hack") {
			t.Fatalf("砍網之後的台詞=%q choices=%v", state.Message, state.currentOriginalChoices)
		}
		for step := 0; step < 8 && state.Mode != ModeDungeon; step++ {
			switch {
			case state.Mode == ModeCombat:
				for turn := 0; turn < 64 && state.Mode == ModeCombat; turn++ {
					if err := state.CombatAct(); err != nil {
						t.Fatalf("紅網蜘蛛戰第 %d 回合：%v", turn, err)
					}
				}
			default:
				if err := state.Continue(); err != nil {
					if selectErr := state.Select(0); selectErr != nil {
						t.Fatalf("砍網之後推不動：continue=%v select=%v", err, selectErr)
					}
				}
			}
			observer.observe()
		}
		if state.Mode != ModeDungeon || state.DungeonX != 6 || state.DungeonY != 14 {
			t.Fatalf("砍網之後 mode=%v pos=(%d,%d,%d)", state.Mode,
				state.DungeonX, state.DungeonY, state.DungeonDirection)
		}
		captureSegmentEnd(t, "ECL6/0x40 墓園：紅網")
	}) {
		t.FailNow()
	}

	if !t.Run("ECL6/0x40 墓園：黛米爾公主的祝福", func(t *testing.T) {
		observer.stopAtMessageID = ""
		for _, target := range []normalDungeonPoint{{6, 12}, {13, 12}, {13, 13}} {
			walkNormalDungeonTo(t, state, &burialGrid, target.x, target.y, observer)
			if state.Mode != ModeDungeon {
				t.Fatalf("走向 (%d,%d) 時模式是 %v", target.x, target.y, state.Mode)
			}
		}
		// ★ 幽魂是**走進那一格**才出現的，站在那一格跑生命週期不會觸發。
		state.TurnDungeonWithGrid(burialGrid, (4-int(state.DungeonDirection)+8)%8)
		if err := state.MoveDungeon(burialGrid, 0, 1, 4); err != nil {
			t.Fatalf("走進公主那一格：%v", err)
		}
		if state.Message != requireGamePackText(t, state, "myth-drannor.daemir.offer") {
			t.Fatalf("幽魂台詞=%q pos=(%d,%d,%d)", state.Message,
				state.DungeonX, state.DungeonY, state.DungeonDirection)
		}
		if err := state.Continue(); err != nil {
			t.Fatalf("幽魂台詞之後推不動：%v", err)
		}
		if !observer.selectOption(t, "option.accept") {
			t.Fatalf("幽魂 ACCEPT 選項不在：%v", state.currentOriginalChoices)
		}
		if state.Message != requireGamePackText(t, state, "myth-drannor.daemir.blessing") {
			t.Fatalf("接受祝福之後的台詞=%q", state.Message)
		}
		if !observer.selectOption(t, "ecl-option.press-button-or-return-to-continue") {
			t.Fatalf("祝福之後沒有繼續選項：%v", state.currentOriginalChoices)
		}
		if state.Mode != ModeDungeon || state.DungeonX != 13 || state.DungeonY != 14 {
			t.Fatalf("祝福之後 mode=%v pos=(%d,%d,%d)", state.Mode,
				state.DungeonX, state.DungeonY, state.DungeonDirection)
		}
		captureSegmentEnd(t, "ECL6/0x40 墓園：黛米爾公主的祝福")
	}) {
		t.FailNow()
	}

	t.Run("段界快照往返", func(t *testing.T) {
		if len(segmentEnds) != 18 {
			t.Fatalf("存到 %d 份段界快照，應該是 18 份", len(segmentEnds))
		}
		blocks := map[uint8][]byte{}
		for member := 1; member <= 6; member++ {
			parsed, err := dax.Parse(zipData(t, image, "ECL"+strconv.Itoa(member)+".DAX"))
			if err != nil {
				t.Fatal(err)
			}
			for _, block := range parsed {
				blocks[block.Entry.ID] = block.Data
			}
		}
		for _, end := range segmentEnds {
			restored := NewStateFromECLBlocks(testCatalog(), blocks, 0x50)
			if err := restored.LoadPartyFile(end.path); err != nil {
				t.Errorf("%s 的段界快照讀不回來：%v", end.name, err)
				continue
			}
			if got := restored.session.CurrentBlockID(); got != end.block {
				t.Errorf("%s 讀回來停在 block %#02x，存的是 %#02x", end.name, got, end.block)
			}
			if restored.Mode != end.mode {
				t.Errorf("%s 讀回來的模式是 %v，存的是 %v", end.name, restored.Mode, end.mode)
			}
			if restored.Area.GameArea != end.area || restored.Area.InDungeon != end.inDungeon {
				t.Errorf("%s 讀回來的章節／地城是 %d/%v，存的是 %d/%v", end.name,
					restored.Area.GameArea, restored.Area.InDungeon, end.area, end.inDungeon)
			}
			if len(restored.PartyFighters()) != end.party {
				t.Errorf("%s 讀回來隊伍 %d 人，存的是 %d 人", end.name,
					len(restored.PartyFighters()), end.party)
			}
		}
	})
}

// campaignSegmentEnd 是一段走完時的邊界狀態，用來驗那份快照讀得回來。
type campaignSegmentEnd struct {
	name      string
	path      string
	block     uint8
	mode      Mode
	area      uint8
	inDungeon bool
	party     int
}
