package game

import (
	"archive/zip"
	"fmt"
	"path/filepath"
	"sort"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unicode"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
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
	// refusedEdges 記下「走過去卻被 ECL 推回來」的邊。原作有一類格子會問
	// 「你要離開嗎」，回答否就把隊伍送回上一格（spec 1157 的 15 處「退回上
	// 一格」）。地圖的牆資料允許那一步，所以尋路演算法看不出來；只有實際
	// 走一次才知道。記下來之後尋路就會繞開，而不是原地來回撞到步數用完。
	refusedEdges map[normalDungeonEdge]bool
	// hapLeaveYes 決定哈普村邊界那句「你們正準備返回荒野。要繼續嗎？」怎麼答。
	// 站在地形碼 2／3／1 的邊界格朝東／西／北就會問；答否會把隊伍推回上一格
	// （ECL5/0x31:0E9Fh），所以只有真的要離村時才答是。
	hapLeaveYes bool
	// messages 記下這條 session 期間玩家看得到的每一句話，供語系不變量檢查。
	messages map[string]bool
}

// normalDungeonEdge 是「某張地圖的某一格往某個方向」這條邊。
type normalDungeonEdge struct {
	set, block uint8
	point      normalDungeonPoint
	direction  int
}

func newNormalCampaignObserver(t *testing.T, state *State) *normalCampaignObserver {
	t.Helper()
	return &normalCampaignObserver{
		state: state, seen: make(map[string]bool), messages: make(map[string]bool),
		refusedEdges: make(map[normalDungeonEdge]bool),
	}
}

func (o *normalCampaignObserver) observe() {
	if o.messages != nil && o.state != nil {
		// 訊息與提示都要記：遭遇選單的敘述走的是 Prompt 那一行。
		if strings.TrimSpace(o.state.Message) != "" {
			o.messages[o.state.Message] = true
		}
		if strings.TrimSpace(o.state.Prompt) != "" {
			o.messages[o.state.Prompt] = true
		}
	}
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
		case ModeMap:
			// 走荒野那條路線會先進世界地圖：目的地已經排好，收掉這一段行程
			// 就會回到目標城鎮的邊緣選單。
			if !o.state.pendingWorldTravel {
				t.Fatalf("normal campaign world map without a pending destination: message=%q choices=%v",
					o.state.Message, o.state.currentOriginalChoices)
			}
			if err := o.state.EnterPlaces(); err != nil {
				t.Fatalf("normal campaign world travel: %v", err)
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
			case o.state.Message == requireGamePackText(t, o.state, "zhentil.enter-prompt"):
				// 黑暗神殿的正門（`ECL4/0x20:0213h`，地形 5 且朝東）。這條路線
				// 走的是劇情那條：先在 `(10,11)` 遇上奧莉薇，由她帶進去。
				// 從正門闖進去會跳過她那一段。
				if !o.selectOption(t, "option.no") {
					t.Fatalf("Zhentil shrine door NO option unavailable: %v", o.state.currentOriginalChoices)
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
			case o.state.Message == requireGamePackText(t, o.state, "wizard-tower.wilderness-exit"):
				// 出塔之後問「要順道去哈普村還是離開這一區」。這條路線已經
				// 走過哈普村了，選 DEPART 繼續世界路線。
				if !o.selectOption(t, "ecl-option.depart") {
					t.Fatalf("wizard-tower exit DEPART option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "wizard-tower.stairs.down"):
				// 巫師塔每層之間靠樓梯事件接（`ECL5/0x33:090Ch` 先用地形碼查
				// 一張「這道樓梯要朝哪個方向站」的表）。這條路線要下樓回熔岩洞，
				// 所以答是——通用 fallback 會先撞到 `option.no`，那等於一直待在
				// 塔頂（spec 1161）。
				if !o.selectOption(t, "option.yes") {
					t.Fatalf("wizard-tower stairs YES option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "hap.leave"):
				optionID := "option.no"
				if o.hapLeaveYes {
					optionID = "option.yes"
				}
				if !o.selectOption(t, optionID) {
					t.Fatalf("Hap leave option %s unavailable: %v", optionID, o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "hap.map-route"):
				if !o.selectOption(t, "ecl-option.caves") {
					t.Fatalf("Hap map route option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "pit.vegepygmies-retreat"),
				o.state.Message == requireGamePackText(t, o.state, "pit.shambling-mounds-push"),
				o.state.Message == requireGamePackText(t, o.state, "pit.shambling-mounds-push-corridor"):
				// 這三場的 `FLEE` 走 `ECL3/0x11:0589h`——會把隊伍推回上一格；
				// `WAIT` 走 `1075h`，跳過那兩句 `SAVE` 直接重畫，隊伍留在原地。
				// 這條路線要往前推進，所以取 `WAIT`。
				if !o.selectOption(t, "ecl-option.wait") {
					t.Fatalf("pit blocking encounter WAIT option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.state.Message == requireGamePackText(t, o.state, "myth-drannor.grave.skeleton"):
				// `GO` 走 `ECL6/0x40:097Fh`，會把隊伍推回上一格；`REBURY SKELETON`
				// 走 `0D28h`，不打也不搶。這條路線要繼續往東，所以重新掩埋。
				if !o.selectOption(t, "myth-drannor.grave.rebury") {
					t.Fatalf("Myth Drannor grave REBURY option unavailable: %v", o.state.currentOriginalChoices)
				}
			case o.selectOption(t, "ecl-option.thank-him"):
				// 絲綢在古熔岩洞出口那一場（`(6,15)`）也給 THANK HIM／ATTACK／
				// LEAVE，敘述只有「WHAT DO YOU DO?」。這條路線一律道謝走人。
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

// refusedEdge 回報「從這一格往這個方向走會被推回來」。
func (o *normalCampaignObserver) refusedEdge(point normalDungeonPoint, direction int) bool {
	if o == nil || o.refusedEdges == nil || o.state == nil {
		return false
	}
	return o.refusedEdges[normalDungeonEdge{
		set: o.state.GeoMapSet, block: o.state.GeoMapBlock,
		point: point, direction: direction,
	}]
}

// recordRefusedEdge 在實際走過去卻沒有離開原地時記下那條邊。
func (o *normalCampaignObserver) recordRefusedEdge(point normalDungeonPoint, direction int) {
	if o == nil || o.refusedEdges == nil || o.state == nil {
		return
	}
	o.refusedEdges[normalDungeonEdge{
		set: o.state.GeoMapSet, block: o.state.GeoMapBlock,
		point: point, direction: direction,
	}] = true
}

// wizardTowerRooms 是 `GEO5/0x33` 上互不相連的五個房間各自有哪些樓梯格，
// 以及每道樓梯「要朝哪個方向站」。
//
// ★ 塔的每一層在 GEO 上是獨立的小房間，層與層之間只靠樓梯事件接：
// `ECL5/0x33:090Ch` 用地形碼查 `1C81h` 那張表拿到方向，朝向不對就直接 `EXIT`
// （畫面上什麼都不會發生）。表的值是 `C04D`（0..3），這裡換成 remake 的
// 0/2/4/6 刻度。房間的分組由 `cmd/geo-probe` 的連通性算出來（spec 1161）。
var wizardTowerRooms = [][]struct {
	point  normalDungeonPoint
	facing uint8
}{
	{{normalDungeonPoint{3, 1}, 0}},
	{{normalDungeonPoint{8, 0}, 4}, {normalDungeonPoint{5, 4}, 6}},
	{{normalDungeonPoint{9, 5}, 2}, {normalDungeonPoint{14, 1}, 0}},
	{{normalDungeonPoint{14, 6}, 4}, {normalDungeonPoint{9, 12}, 6}},
	{{normalDungeonPoint{0, 12}, 2}},
}

// takeWizardTowerStairs 在塔裡走一步：走到這間房裡還沒用過的那道樓梯**前一格**，
// 再朝樓梯要的方向踏上去。
//
// ⚠ 不要直接用 `walkNormalDungeonTo` 走到樓梯格上：踏上去的當下事件就會把隊伍
// 傳到別的房間，走訪器會以為「還沒到目標」而繼續找路，然後回報「目標走不到」。
// 同一間房裡的兩道樓梯一上一下，所以要先挑不是腳下這一格的那道。
func takeWizardTowerStairs(
	t *testing.T, state *State, observer *normalCampaignObserver, used map[normalDungeonPoint]bool,
) bool {
	t.Helper()
	here := normalDungeonPoint{state.DungeonX, state.DungeonY}
	for _, room := range wizardTowerRooms {
		inRoom := false
		for _, stairs := range room {
			if stairs.point == here {
				inRoom = true
			}
		}
		if !inRoom {
			continue
		}
		ordered := make([]struct {
			point  normalDungeonPoint
			facing uint8
		}, 0, len(room))
		for _, stairs := range room {
			if stairs.point != here {
				ordered = append(ordered, stairs)
			}
		}
		for _, stairs := range room {
			if stairs.point == here {
				ordered = append(ordered, stairs)
			}
		}
		for _, stairs := range ordered {
			if used[stairs.point] {
				continue
			}
			used[stairs.point] = true
			deltaX, deltaY := normalDungeonDelta(int(stairs.facing))
			if stairs.point == here {
				// 已經站在樓梯上（劇情把隊伍送到這裡）：原地轉到它要的方向。
				// `7F81` 是「這一步已經演過一場」的旗標，移動時才會被清掉，
				// 原地轉向要照移動的做法清一次，否則事件會被自己的旗標擋住。
				state.TurnDungeonWithGrid(observer.towerGrid,
					(int(stairs.facing)-int(state.DungeonDirection)+8)%8)
				state.session.SetMemoryValue(0x7F81, 0)
				if err := state.RunDungeonLifecycle(); err != nil {
					t.Fatalf("run wizard-tower stairs lifecycle: %v", err)
				}
				observer.resolveDungeonBoundary(t)
				return true
			}
			approach := normalDungeonPoint{
				geo.WrapCoordinate(stairs.point.x-deltaX, geo.Width),
				geo.WrapCoordinate(stairs.point.y-deltaY, geo.Height),
			}
			if approach != here {
				walkNormalDungeonTo(t, state, &observer.towerGrid, approach.x, approach.y, observer)
				if state.session.CurrentBlockID() != 0x33 {
					return true
				}
			}
			if err := state.MoveDungeon(observer.towerGrid,
				deltaX, deltaY, int(stairs.facing)); err != nil {
				t.Fatalf("step onto wizard-tower stairs (%d,%d): %v",
					stairs.point.x, stairs.point.y, err)
			}
			observer.resolveDungeonBoundary(t)
			return true
		}
	}
	return false
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
				// ⚠ 巫師塔（`GEO5/0x33`）也要放行：`(8,15)` 與 `(12,10)` 落在
				// 那張圖的走廊上，排除掉會把第 15 列切斷，看起來就像「塔頂下
				// 不來」（spec 1161）。這份清單是**逐圖**的排除，不是通則。
				if !(state.GeoMapSet == 4 && state.GeoMapBlock == 0x21) &&
					!(state.GeoMapSet == 5 && state.GeoMapBlock == 0x33) &&
					(next == (normalDungeonPoint{10, 2}) || next == (normalDungeonPoint{8, 11}) ||
						next == (normalDungeonPoint{8, 15}) || next == (normalDungeonPoint{12, 10})) {
					continue
				}
				if observer.refusedEdge(current, direction) {
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
							!(state.GeoMapSet == 5 && state.GeoMapBlock == 0x33) &&
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
				refused := []string{}
				for edge := range observer.refusedEdges {
					refused = append(refused, fmt.Sprintf("%+v", edge))
				}
				sort.Strings(refused)
				t.Fatalf("normal dungeon target (%d,%d) is unreachable from (%d,%d) and no locked door leads onward; refused=%v",
					targetX, targetY, start.x, start.y, refused)
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
			doorEdgeRefused := false
			for _, step := range doorPath {
				observer.resolveDungeonBoundary(t)
				if state.Mode != ModeDungeon {
					t.Fatalf("normal dungeon route to locked door mode=%v message=%q", state.Mode, state.Message)
				}
				deltaX, deltaY := normalDungeonDelta(step.direction)
				doorFrom := normalDungeonPoint{state.DungeonX, state.DungeonY}
				if err := state.MoveDungeon(*grid, deltaX, deltaY, step.direction); err != nil {
					t.Fatalf("normal dungeon route to locked door (%d,%d): %v",
						nextDoor.source.x, nextDoor.source.y, err)
				}
				observer.resolveDungeonBoundary(t)
				// 走去開鎖門的路上一樣會撞到「走過去卻被推回來」的邊；記下來
				// 讓外層重新規劃，否則會卡在同一步（同主迴圈那段）。
				if state.DungeonX == doorFrom.x && state.DungeonY == doorFrom.y {
					observer.recordRefusedEdge(doorFrom, step.direction)
					doorEdgeRefused = true
					break
				}
				if observer.stoppedAtDataPackEvent() {
					return
				}
				if state.session != nil && state.session.CurrentBlockID() != initialBlock {
					return
				}
			}
			if doorEdgeRefused {
				continue
			}
			// ⚠ 走到門口的路上劇情可能把隊伍搬走。搬走了就把控制權交回呼叫端
			// 重新規劃——原本的目標很可能在新位置那一區根本走不到。
			if state.DungeonX != nextDoor.source.x || state.DungeonY != nextDoor.source.y {
				return
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
		// 走完一步還在原地，代表這條邊被 ECL 擋下來（例如「要離開嗎」回答否
		// 會把隊伍推回上一格）。記下來讓尋路繞開，否則會一直重試同一步。
		if state.DungeonX == start.x && state.DungeonY == start.y {
			observer.recordRefusedEdge(start, step.direction)
			continue
		}
		// ⚠ 落點不是規劃的那一格，代表劇情把隊伍搬走了（傳送、走位動畫）。
		// 把控制權交回呼叫端重新規劃，不要繼續照舊計畫走——舊計畫的目標很
		// 可能在新位置那一區根本走不到，然後會以「目標走不到」收場，
		// **症狀出現在終點、原因在中間某一步**（spec 1161）。
		if state.DungeonX != step.point.x || state.DungeonY != step.point.y {
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

func TestRealNewGameRunsToTheEnding(t *testing.T) {
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
		roster := make([]string, 0, len(state.partyRoster))
		experience := make(map[string]uint32, len(state.partyRoster))
		equipment := make(map[string]string, len(state.partyRoster))
		spells := make(map[string]string, len(state.partyRoster))
		effects := make(map[string]string, len(state.partyRoster))
		for _, character := range state.partyRoster {
			label := character.ScriptName
			if label == "" {
				label = character.Name
			}
			roster = append(roster, label)
			experience[label] = character.Experience
			equipment[label] = campaignEquipmentSignature(character)
			spells[label] = campaignSpellSignature(character)
			effects[label] = campaignEffectSignature(character)
		}
		segmentEnds = append(segmentEnds, campaignSegmentEnd{
			name: name, path: path, block: state.session.CurrentBlockID(),
			mode: state.Mode, area: state.Area.GameArea, inDungeon: state.Area.InDungeon,
			party: len(state.PartyFighters()), roster: roster, experience: experience,
			music: state.activeMusicTrackID,
			equipment: equipment, spells: spells, effects: effects,
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
		// 第六章的怪物表：密斯卓諾的蜘蛛、盜墓者與最終戰都在這一份。
		monster6Blocks, err := dax.Parse(zipData(t, image, "MON6CHA.DAX"))
		if err != nil {
			t.Fatal(err)
		}
		monster6Records := make(map[uint8]monster.Record, len(monster6Blocks))
		for _, block := range monster6Blocks {
			record, parseErr := monster.Parse(block.Data)
			if parseErr != nil {
				t.Fatal(parseErr)
			}
			monster6Records[block.Entry.ID] = record
		}
		state.SetMonsterRecordsForECL(6, monster6Records)
		grid = loadGeo5CampaignGrid(t, image, 0x32)
		towerGrid = loadGeo5CampaignGrid(t, image, 0x33)
		observer = newNormalCampaignObserver(t, state)
		observer.towerGrid = towerGrid
		// 哈普村只佔 GEO5/0x32 的西半（x≤5）：ECL5/0x31 的每格分派前面就擋著
		// `COMPARE C04F 80h; IF <`，地形碼 ≥ 0x80 的東半是黑暗精靈那張圖
		// （ECL5/0x32）共用同一張 GEO。村子的邊界格是地形碼 1（朝北）、
		// 2（朝東）、3（朝西），走上去就會問「要不要回荒野」。
		//
		// 路線因此全部落在西半：`(4,10)` 是村民（地形碼 4）、`(4,4)` 是阿卡巴
		// （地形碼 A）、`(3,13)` 是伊夫利特的穀倉（地形碼 8）、`(4,12)` 是東側
		// 出村格 `(5,12)` 的前一格。
		walkNormalDungeonTo(t, state, &grid, 4, 10, observer)
		// `4C02` 是「這一步已經演過一場」的閂：每格分派器先寫 0
		// （ECL5/0x31:03D7h），每支場景處理常式看到 1 就 `EXIT`、演完寫 1
		// （`0534h`／`066Eh`／`0B22h`／`0CFAh`）。剛演完村民那場就會是 1。
		if value, ok := state.session.MemoryValue(0x4C02); !ok || value != 1 {
			t.Fatalf("normal Hap peasants scene left memory[0x4c02]=%#x,%v want 1", value, ok)
		}
		for _, target := range []normalDungeonPoint{{4, 4}, {3, 13}, {4, 12}} {
			walkNormalDungeonTo(t, state, &grid, target.x, target.y, observer)
		}
		if !observer.seen["hap.peasants-flee"] || !observer.seen["hap.akabar-join"] ||
			!observer.seen["hap.efreet-map"] {
			t.Fatalf("normal Hap story coverage=%v", observer.seen)
		}
		// `4C02` 在這裡是 0：隊伍停在 `(4,12)`（地形碼 0），分派器已經把閂清掉，
		// 而地形碼 0 那一支不演任何場景，所以不會再寫回 1。
		for address, want := range map[uint16]uint16{0x4C01: 5, 0x4C02: 0, 0x4C5E: 1, 0x4C5F: 1} {
			if got, ok := state.session.MemoryValue(address); !ok || got != want {
				t.Fatalf("normal Hap memory[%#x]=%#x,%v want %#x", address, got, ok, want)
			}
		}
		// 出村：站上 `(5,12)`（地形碼 2）朝東就會問，這一次答「是」。
		observer.hapLeaveYes = true
		if err := state.MoveDungeon(grid, 1, 0, 2); err != nil {
			t.Fatalf("normal Hap east village exit: %v", err)
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
		// ⚠ 走去熔岩池的途中會被劇情搬走：巫師塔那一段的 `ECL5/0x33:022Bh`
		// 把隊伍直接放到塔頂 `(3,1)`。塔裡要一層一層走樓梯下來，最後在
		// 地面層的 `(7,15)` 朝東才會冒出「地道還是荒野」那個出口選單
		// （`0811h`）；選 CAVES 走 `088Bh` ＝ `SAVE 06 C04B; SAVE 0F C04C;
		// NEWECL 32h`（spec 1160／1161）。
		observer.towerReturnOption = "ecl-option.caves"
		towerStairsUsed := map[normalDungeonPoint]bool{}
		for attempt := 0; attempt < 16 && (state.session.CurrentBlockID() != 0x32 ||
			state.DungeonX != 0 || state.DungeonY != 5); attempt++ {
			if state.session.CurrentBlockID() == 0x33 {
				observer.towerReady = true
				if !takeWizardTowerStairs(t, state, observer, towerStairsUsed) {
					// 樓梯走完就到了塔的地面層。出口 `(7,15)`（地形碼 1）不走
					// 樓梯那條分派——`07D7h` 先比地形碼 1，再在 `07E2h` 比
					// `C04D == 1`（朝東）。⚠ 樓梯方向表對索引 1 給的是 5，
					// 不是合法朝向，正說明這一格不是樓梯。
					walkNormalDungeonTo(t, state, &observer.towerGrid, 7, 15, observer)
					if state.session.CurrentBlockID() != 0x33 {
						continue
					}
					state.TurnDungeonWithGrid(observer.towerGrid,
						(2-int(state.DungeonDirection)+8)%8)
					state.session.SetMemoryValue(0x7F81, 0)
					if err := state.RunDungeonLifecycle(); err != nil {
						t.Fatalf("run wizard-tower exit lifecycle: %v", err)
					}
					observer.resolveDungeonBoundary(t)
				}
				continue
			}
			walkNormalDungeonTo(t, state, &grid, 0, 5, observer)
		}
		if state.session.CurrentBlockID() != 0x32 || state.DungeonX != 0 || state.DungeonY != 5 {
			t.Fatalf("normal lava pool approach ended at (%d,%d,%d) block=%#x mode=%v",
				state.DungeonX, state.DungeonY, state.DungeonDirection,
				state.session.CurrentBlockID(), state.Mode)
		}
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
		// ⚠ 古熔岩洞（`ECL5/0x32`）與哈普村（`ECL5/0x31`）共用 `GEO5/0x32`
		// （spec 1158）。從熔岩池走到地圖出口 `(6,15)` 的路上會經過村界，
		// 走進去就換成 0x31；再走一次就會從村子的出口回到 0x32。
		for attempt := 0; attempt < 6 && state.Mode == ModeDungeon &&
			(state.DungeonX != 6 || state.DungeonY != 15); attempt++ {
			walkNormalDungeonTo(t, state, &grid, 6, 15, observer)
		}
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
			// 進猶拉什是腳本演的走位：`ECL3/0x10:006Ch` 先落在 `(0,8)` 朝西，
			// `0127h` 移到 `(1,0)` 朝南，接著一串 `CALL C01Eh` 往南再往西走到
			// `(0,3)`——**收尾朝西**，不是地圖宣告的 spawn 那個朝東。
			state.DungeonX != 0 || state.DungeonY != 3 || state.DungeonDirection != 6 {
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
		// ECL 那一格的條件是位置加朝向：奧莉薇在 `(10,11,N)` 才出現。所以從南邊
		// 那一格走上去，讓一般的移動供給朝北的觸發，而不是直接寫座標或原地重畫。
		// ⚠ 上面那趟路本身就可能經過她那一格：她會直接把隊伍帶進黑暗神殿
		// （`ECL4/0x21`）。已經被帶走就不要再照城區的地圖走一次——那張圖上的
		// 座標在神殿裡沒有意義。
		if state.session.CurrentBlockID() == 0x20 {
			walkNormalDungeonTo(t, state, &zhentilGrid, 10, 12, observer)
		}
		if state.session.CurrentBlockID() == 0x20 {
			state.TurnDungeonWithGrid(zhentilGrid, (0-int(state.DungeonDirection)+8)%8)
			if err := state.MoveDungeon(zhentilGrid, 0, -1, 0); err != nil {
				t.Fatalf("normal Zhentil Olive north-facing approach: %v", err)
			}
			observer.resolveDungeonBoundary(t)
		}
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
		// 只認格子，不認朝向：走到 E1 的最後一步朝哪邊由前面那段的收尾決定，
		// 下一段一開始就會自己轉向。
		if state.DungeonX != 5 || state.DungeonY != 7 {
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
		// ⚠ 傳送現在是腳本自己做的：`ECL4/0x22:061Bh` 在「YOU ARE SUDDENLY
		// SLAMMED AGAINST A WALL.」之後寫 `(13,1)` 朝西、重畫，接著同一次執行
		// 就印出死精靈那一段。停下來的訊號因此改成那句敘述本身——不能再靠
		// game pack 那條 `set_map_position` 替身事件，而且**要在觀察器的
		// 「洞穴裡一律 LEAVE」那條通用規則之前停住**，否則死精靈的選單會被
		// 順手選掉（spec 1161）。
		observer.stopAtMessageID = "dexam.dead-elf.remains"
		// The source cell is reached by ordinary GEO movement.  The ECL transaction
		// then writes the destination through C04B/C04C/C04D; the title pack only
		// projects that original virtual-map handoff back into State.
		for attempt := 0; attempt < 6 && state.Mode == ModeDungeon &&
			(state.DungeonX != 13 || state.DungeonY != 1); attempt++ {
			walkNormalDungeonTo(t, state, &caveGrid, 5, 9, observer)
		}
		// 敘述先停在按鍵那一格；推完之後同一場的選單才交出來。
		// 牆面／地形現在由重畫自己去地圖重讀（spec 1161），到站時就已經是
		// `(13,1)` 那一格的值。
		if (state.Mode != ModeWilderness && state.Mode != ModeEvent) ||
			state.DungeonX != 13 || state.DungeonY != 1 || state.DungeonDirection != 6 ||
			state.DungeonWallType != 8 || state.DungeonWallRoof != 0xC0 {
			t.Fatalf("normal Beholder Cave teleporter mode=%v pos=(%d,%d,%d) wall=%#x roof=%#x coverage=%v",
				state.Mode, state.DungeonX, state.DungeonY, state.DungeonDirection,
				state.DungeonWallType, state.DungeonWallRoof, observer.seen)
		}
		// ⚠ 這裡不再驗 `4C03`：那一格先前是 game pack 那條 `set_map_position`
		// 替身順手寫的（`set_memory 4C03 = 0`）。傳送改由腳本自己做之後，
		// 這一刻 `4C03` 就是腳本留下的值，不該再拿替身的副作用當期望值。
		for address, want := range map[uint16]uint16{
			0xC04B: 13, 0xC04C: 1, 0xC04D: 3, 0xC04E: 8, 0xC04F: 0xC0,
		} {
			got, ok := state.session.MemoryValue(address)
			if !ok || got != want {
				t.Fatalf("normal Beholder Cave teleporter memory[%#x]=%#x ok=%v, want %#x", address, got, ok, want)
			}
		}
		if guard, ok := state.session.MemoryValue(0x7F81); !ok || guard != 0 {
			t.Fatalf("normal Beholder Cave teleporter guard=%#x ok=%v, want cleared", guard, ok)
		}
		// 傳送同一次執行就印出死精靈的敘述（`ECL4/0x22:0631h`），這條路線停在
		// 這裡再自己把後面那段互動走完。
		if state.Message != requireGamePackText(t, state, "dexam.dead-elf.remains") {
			t.Fatalf("normal Beholder Cave dead-elf narration message=%q choices=%v",
				state.Message, state.currentOriginalChoices)
		}
		// `0651h` 用 `4C06` 把那段互動擋成一次性。這一格是眼魔洞穴自己的暫存
		// （spec 1162），進洞時是 0，所以走的是「第一次」那條分支，`066Bh`
		// 已經把它設成 1。
		if flag, ok := state.session.MemoryValue(0x4C06); !ok || flag != 1 {
			t.Fatalf("normal Beholder Cave dead-elf gate 4C06=%#x,%v, want 1", flag, ok)
		}
		// ⚠ 停用停止訊號再往下走：`stopAtMessageID` 還設著的話，觀察器每次都會
		// 在這句敘述上立刻返回，看起來就像「推不動」。
		observer.stopAtMessageID = ""
		// 敘述先停在按鍵那一格；推完之後同一場的選單才交出來。
		for attempt := 0; attempt < 4 &&
			!observer.hasOption("dexam.dead-elf.examine-remains"); attempt++ {
			if state.Mode == ModeEvent {
				if err := state.Continue(); err != nil {
					t.Fatalf("continue the dead-elf narration: %v", err)
				}
				continue
			}
			if state.Mode != ModeWilderness ||
				!observer.hasOption("ecl-option.press-button-or-return-to-continue") {
				break
			}
			selectOption(t, "ecl-option.press-button-or-return-to-continue")
		}
		if !observer.hasOption("dexam.dead-elf.examine-remains") ||
			!observer.hasOption("ecl-option.leave") {
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
		if got, ok := state.session.MemoryValue(0x4C07); !ok || got != 0x80 {
			t.Fatalf("normal Beholder Cave raw memory[0x4c07]=%#x ok=%v, want 0x80", got, ok)
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
		// 灰袍男子唸的是「還剩幾位主人」：`ECL1/0x50:0232h` 把 `4C59`（巫師塔的
		// 崔坎卓斯）、`4C5B`（猶拉什）、`4C5A`（眼魔洞穴的 Dexam）各算一分，
		// 三分就走 `028Bh IF =` 那條——直接揭露烈焰之主並指向密斯卓諾。
		// 這條主線在走到立石群之前三處都已經打完，所以看到的是這一頁，
		// 不是「還剩兩位」。
		if state.Message != requireGamePackText(t, state, "myth-drannor.tyranthraxus-reveal") {
			t.Fatalf("立石群灰袍男子的台詞=%q choices=%v", state.Message, state.currentOriginalChoices)
		}
		for press := 0; press < 4 && !observer.hasOption("ecl-option.journey-on"); press++ {
			if state.Mode == ModeEvent {
				if err := state.Continue(); err != nil {
					t.Fatalf("立石群揭露頁：%v", err)
				}
			} else if !observer.selectOption(t, "ecl-option.press-button-or-return-to-continue") {
				break
			}
			observer.observe()
		}
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
		// 密斯卓諾只有荒野這一條路（`world-route.myth-drannor` 只給
		// `WILDERNESS`／`EXIT`），所以路線選項跟目的地一起排進佇列。
		observer.nextWorldDestinations = []string{"ecl-option.myth-drannor", "ecl-option.wilderness"}
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

	if !t.Run("ECL6/0x42 密斯卓諾：外城遺跡", func(t *testing.T) {
		// ★ block 0x40 的邊界處理是 `ON GOTO C04D`（面向／2）：只有朝東跨出邊界
		// 才會走到「更多遺跡」那個選單，其餘方向都回世界地圖。
		observer.stopAtMessageID = "myth-drannor.more-ruins"
		state.TurnDungeonWithGrid(burialGrid, (2-int(state.DungeonDirection)+8)%8)
		if err := state.RunDungeonExitLifecycle(); err != nil {
			t.Fatalf("朝東的邊界生命週期：%v", err)
		}
		observer.observe()
		if state.Message != requireGamePackText(t, state, "myth-drannor.more-ruins") {
			t.Fatalf("邊界選單台詞=%q choices=%v", state.Message, state.currentOriginalChoices)
		}
		if !observer.selectOption(t, "myth-drannor.path") {
			t.Fatalf("邊界選單沒有 PATH：%v", state.currentOriginalChoices)
		}
		observer.stopAtMessageID = ""
		for step := 0; step < 6 && state.Mode != ModeDungeon; step++ {
			if err := state.Continue(); err != nil {
				if selectErr := state.Select(0); selectErr != nil {
					t.Fatalf("走上小徑之後推不動：continue=%v select=%v", err, selectErr)
				}
			}
			observer.observe()
		}
		if state.session.CurrentBlockID() != 0x42 || state.GeoMapSet != 6 || state.GeoMapBlock != 0x42 {
			t.Fatalf("走上小徑之後 block=%#02x geo=%d/%#02x pos=(%d,%d,%d) mode=%v message=%q",
				state.session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock,
				state.DungeonX, state.DungeonY, state.DungeonDirection, state.Mode, state.Message)
		}
		captureSegmentEnd(t, "ECL6/0x42 密斯卓諾：外城遺跡")
	}) {
		t.FailNow()
	}

	if !t.Run("ECL6/0x43 密斯卓諾：內城遺跡", func(t *testing.T) {
		// 同樣是 `ON GOTO C04D`，這一塊的索引 0（＝朝北）走到大神殿廢墟。
		outerGrid := loadGeoCampaignGrid(t, image, 6, "GEO6.DAX", 0x42)
		observer.stopAtMessageID = "myth-drannor.outer.ruined-temple"
		state.TurnDungeonWithGrid(outerGrid, (0-int(state.DungeonDirection)+8)%8)
		if err := state.RunDungeonExitLifecycle(); err != nil {
			t.Fatalf("朝北的邊界生命週期：%v", err)
		}
		observer.observe()
		if state.Message != requireGamePackText(t, state, "myth-drannor.outer.ruined-temple") {
			t.Fatalf("神殿邊界台詞=%q", state.Message)
		}
		observer.stopAtMessageID = ""
		// 神殿前問兩次「要繼續嗎」：第二次還會提醒前方危機四伏。
		for step := 0; step < 8 && state.session.CurrentBlockID() != 0x43; step++ {
			if observer.hasOption("option.yes") {
				if !observer.selectOption(t, "option.yes") {
					t.Fatalf("神殿問句選不到 YES：%v", state.currentOriginalChoices)
				}
			} else if err := state.Continue(); err != nil {
				if selectErr := state.Select(0); selectErr != nil {
					t.Fatalf("神殿問句推不動：continue=%v select=%v", err, selectErr)
				}
			}
			observer.observe()
		}
		if state.session.CurrentBlockID() != 0x43 || state.GeoMapSet != 6 ||
			state.GeoMapBlock != 0x43 || state.DungeonX != 6 || state.DungeonY != 15 {
			t.Fatalf("進內城遺跡之後 block=%#02x geo=%d/%#02x pos=(%d,%d,%d) mode=%v",
				state.session.CurrentBlockID(), state.GeoMapSet, state.GeoMapBlock,
				state.DungeonX, state.DungeonY, state.DungeonDirection, state.Mode)
		}
		captureSegmentEnd(t, "ECL6/0x43 密斯卓諾：內城遺跡")
	}) {
		t.FailNow()
	}

	if !t.Run("ECL6/0x43 內城遺跡：儀式與爪牙戰", func(t *testing.T) {
		observer.stopAtMessageID = ""
		seen := map[string]bool{}
		for step := 0; step < 60 && state.Mode != ModeDungeon; step++ {
			seen[state.Message] = true
			if state.Mode == ModeCombat {
				for turn := 0; turn < 200 && state.Mode == ModeCombat; turn++ {
					if err := state.CombatAct(); err != nil {
						t.Fatalf("爪牙戰第 %d 回合：%v", turn, err)
					}
				}
				observer.observe()
				continue
			}
			if err := state.Continue(); err != nil {
				if selectErr := state.Select(0); selectErr != nil {
					t.Fatalf("儀式推不動：continue=%v select=%v", err, selectErr)
				}
			}
			observer.observe()
		}
		// 儀式的骨幹：提朗瑟克斯的演說（手札 48）、三件神器被丟進光芒之池、
		// 祭司掀開兜帽是無名者、密語解除枷印的控制、爪牙撲上來。
		for _, messageID := range []string{
			"myth-drannor.inner.ritual.arrival",
			"myth-drannor.inner.ritual.journal",
			"myth-drannor.inner.ritual.pool",
			"myth-drannor.inner.ritual.nameless-reveal",
			"myth-drannor.inner.ritual.bonds-fade",
			"myth-drannor.inner.minions-attack",
		} {
			text := requireGamePackText(t, state, messageID)
			if !seen[text] {
				t.Fatalf("儀式沒有走到 %s", messageID)
			}
		}
		if state.Mode != ModeDungeon {
			t.Fatalf("爪牙戰之後模式是 %v", state.Mode)
		}
		captureSegmentEnd(t, "ECL6/0x43 內城遺跡：儀式與爪牙戰")
	}) {
		t.FailNow()
	}

	if !t.Run("ECL6/0x43 內城遺跡：一樓房間", func(t *testing.T) {
		innerGrid := loadGeoCampaignGrid(t, image, 6, "GEO6.DAX", 0x43)
		observer.stopAtMessageID = ""
		// ★ 每格事件由地形碼分派（`ON GOTO (C04F & 0x7F)`），所以房間的位置是
		// 從 GEO 的地形碼反推的，不是走到哪算哪。
		rooms := []struct {
			x, y       int
			messageID  string
			suppressed bool
		}{
			{8, 13, "myth-drannor.inner.chapel", false},
			{9, 12, "myth-drannor.inner.bedroom", false},
			{11, 12, "myth-drannor.inner.office", false},
			{13, 14, "myth-drannor.inner.kitchen", false},
			// 「下水道塌了」只有從下水道進來（`4C07` 非 0）才唸得到
			// （`ECL6/0x43:0E47h`）。這條路線走的是正門，所以這一格安靜。
			{15, 15, "myth-drannor.inner.sewer-collapsed", true},
			// 這兩間的守衛都帶著 `4C00`（`0E7Fh`／`0EF8h`）——儀式那場打完之後
			// 警報已經拉起來，房裡的人不會再照本宣科招呼你們。
			{14, 11, "myth-drannor.inner.tiered-beds", true},
			{14, 8, "myth-drannor.inner.worshipping-priests", true},
			{4, 9, "myth-drannor.inner.statuary", false},
			{1, 9, "myth-drannor.inner.kennel", false},
		}
		for _, room := range rooms {
			walkNormalDungeonTo(t, state, &innerGrid, room.x, room.y, observer)
			resolveInnerRoom(t, state, observer)
			fired := observer.messages[requireGamePackText(t, state, room.messageID)]
			switch {
			case fired && room.suppressed:
				t.Errorf("(%d,%d) 的 %s 現在出得來了，請把 suppressed 拿掉",
					room.x, room.y, room.messageID)
			case !fired && !room.suppressed:
				t.Errorf("走到 (%d,%d) 沒有出現 %s（停在 (%d,%d,%d) roof=%#02x）",
					room.x, room.y, room.messageID, state.DungeonX, state.DungeonY,
					state.DungeonDirection, state.DungeonWallRoof)
			}
		}
		// `4C00`..`4C0F` 是每一段自己的暫存（spec 1162），所以走進內城時這一區
		// 是乾淨的：`4C07` 是 0（沒從下水道進來），廚房才唸得出來。
		if value, ok := state.session.MemoryValue(0x4C07); ok && value != 0 {
			t.Errorf("進內城時 4C07=%#x，這條路線不是從下水道進來的", value)
		}
		captureSegmentEnd(t, "ECL6/0x43 內城遺跡：一樓房間")
	}) {
		t.FailNow()
	}

	if !t.Run("ECL6/0x43 內城遺跡：二樓與最終戰", func(t *testing.T) {
		innerGrid := loadGeoCampaignGrid(t, image, 6, "GEO6.DAX", 0x43)
		observer.stopAtMessageID = ""
		// ★ 內城的每格事件是 `ON GOTO (C04F & 0x7F)`——用**地形碼**分派（0 起算）。
		// 上二樓的樓梯是地形 `0x97`（索引 23），只有 (10,7) 那一格是。
		walkNormalDungeonTo(t, state, &innerGrid, 10, 7, observer)
		state.DungeonWallRoof = innerGrid.CellWrapped(state.DungeonX, state.DungeonY).Terrain
		if err := state.RunDungeonLifecycle(); err != nil {
			t.Fatalf("樓梯格生命週期：%v", err)
		}
		observer.observe()
		if state.Message != requireGamePackText(t, state, "myth-drannor.inner.stairs-up") {
			t.Fatalf("樓梯台詞=%q choices=%v", state.Message, state.currentOriginalChoices)
		}
		if !observer.selectOption(t, "option.yes") {
			t.Fatalf("樓梯選不到 YES：%v", state.currentOriginalChoices)
		}
		if state.Mode != ModeDungeon || state.DungeonX != 2 || state.DungeonY != 5 {
			t.Fatalf("上樓之後 mode=%v pos=(%d,%d,%d)", state.Mode,
				state.DungeonX, state.DungeonY, state.DungeonDirection)
		}
		// 二樓的房間也照地形碼逐間走一次，再往東北角。
		upstairs := []struct {
			x, y       int
			messageID  string
			suppressed bool
		}{
			{10, 5, "myth-drannor.inner.preservation-room", false},
			// 地形 `0x93` 的分派是一句 `EXIT`（`10EAh`）；有內容的是地形
			// `0x94`／`0x95` 那兩支。這一格本來就不唸。
			{10, 1, "myth-drannor.inner.biers", true},
			{14, 1, "myth-drannor.inner.library", false},
			{14, 4, "myth-drannor.inner.food-storeroom", false},
			{14, 5, "myth-drannor.inner.magic-circle", false},
		}
		for _, room := range upstairs {
			state.SetDungeonGeometryView(room.x, room.y, 0)
			state.DungeonWallRoof = innerGrid.CellWrapped(room.x, room.y).Terrain
			if err := state.RunDungeonLifecycle(); err != nil {
				t.Fatalf("二樓房間 (%d,%d)：%v", room.x, room.y, err)
			}
			resolveInnerRoom(t, state, observer)
			fired := observer.messages[requireGamePackText(t, state, room.messageID)]
			switch {
			case fired && room.suppressed:
				t.Errorf("(%d,%d) 的 %s 現在出得來了，請把 suppressed 拿掉",
					room.x, room.y, room.messageID)
			case !fired && !room.suppressed:
				t.Errorf("二樓 (%d,%d) 沒有出現 %s（roof=%#02x）",
					room.x, room.y, room.messageID, state.DungeonWallRoof)
			}
		}

		// 停屍架那一格安靜的原因是抵達標記 `4C0E`，不是別張地圖留下的殘值
		// （`4C00`..`4C0F` 已經是每一段自己的，spec 1162）。
		// `4C00`..`4C0F` 已經是每一段自己的暫存（spec 1162），走進內城時整區
		// 是乾淨的。停屍架那一格安靜跟旗標無關：它的地形是 `0x93`，
		// `ON GOTO (C04F & 0x7F)` 的第 19 支就是一句 `EXIT`（`10EAh`）。
		if terrain := innerGrid.CellWrapped(10, 1).Terrain; terrain != 0x93 {
			t.Errorf("停屍架那一格的地形變成 %#x 了，分派索引要重算", terrain)
		}

		// 二樓到東北角的路線（spec 408 記下的最短合法路徑）。
		route := []struct {
			x, y      int
			direction uint8
		}{
			{2, 4, 0}, {2, 3, 0}, {2, 2, 0}, {2, 1, 0}, {2, 0, 0},
			{3, 0, 2}, {4, 0, 2}, {5, 0, 2}, {6, 0, 2}, {6, 1, 4},
		}
		seen := map[string]bool{}
		for index, point := range route {
			state.SetDungeonGeometryView(point.x, point.y, point.direction)
			state.DungeonWallRoof = innerGrid.CellWrapped(point.x, point.y).Terrain
			if err := state.RunDungeonLifecycle(); err != nil {
				t.Fatalf("二樓第 %d 步：%v", index, err)
			}
			for boundary := 0; boundary < 24 && state.Mode != ModeDungeon; boundary++ {
				seen[state.Message] = true
				if state.CombatActive() {
					if seen[requireGamePackText(t, state, "myth-drannor.inner.final-amulet")] {
						break
					}
					for turn := 0; turn < 400 && state.CombatActive(); turn++ {
						if err := state.CombatAct(); err != nil {
							t.Fatalf("二樓第 %d 步的戰鬥第 %d 回合：%v", index, turn, err)
						}
					}
					observer.observe()
					continue
				}
				switch state.Mode {
				case ModeEvent:
					if err := state.Continue(); err != nil {
						t.Fatalf("二樓第 %d 步的事件：%v", index, err)
					}
				default:
					if len(state.Choices) == 0 {
						t.Fatalf("二樓第 %d 步停在 mode=%v 且沒有選項", index, state.Mode)
					}
					if err := state.Select(0); err != nil {
						t.Fatalf("二樓第 %d 步的選單：%v", index, err)
					}
				}
				observer.observe()
			}
			if state.CombatActive() {
				break
			}
		}
		// 最終對峙的三段台詞，以及提朗瑟克斯真的開打。
		for _, messageID := range []string{
			"myth-drannor.inner.final-compulsion",
			"myth-drannor.inner.final-defiance",
			"myth-drannor.inner.final-amulet",
		} {
			text := requireGamePackText(t, state, messageID)
			if !seen[text] {
				t.Fatalf("最終對峙沒有走到 %s：看到的是 %v", messageID, len(seen))
			}
		}
		if !state.CombatActive() {
			t.Fatalf("最終戰沒有開打：mode=%v message=%q", state.Mode, state.Message)
		}
		if fighters := len(state.CombatFighters()); fighters < 20 {
			t.Fatalf("最終戰只有 %d 名戰鬥員", fighters)
		}
		captureSegmentEnd(t, "ECL6/0x43 內城遺跡：二樓與最終戰")
	}) {
		t.FailNow()
	}

	if !t.Run("結局：擊敗提朗瑟克斯", func(t *testing.T) {
		for turn := 0; turn < 600 && state.CombatActive(); turn++ {
			if err := state.CombatAct(); err != nil {
				t.Fatalf("最終戰第 %d 回合：%v", turn, err)
			}
		}
		if state.CombatStatus() != combat.StatusPartyWon {
			t.Fatalf("最終戰結果=%v", state.CombatStatus())
		}
		observer.observe()
		// ★ 原作打完最終戰先跑結局過場才回主選單（spec 1082／1154）。
		// 這裡逐頁走過去，順便證明那五頁在真實玩家路徑上到得了、而且是中文。
		for page := 0; state.endingScene; page++ {
			if page >= len(endingSceneKeys) {
				t.Fatalf("結局過場翻不完，停在第 %d 頁", state.endingPageIndex)
			}
			if !campaignMessageHasHan(state.Message) {
				t.Fatalf("結局第 %d 頁落回原文：%q", page+1, state.Message)
			}
			observer.observe()
			if err := state.Select(0); err != nil {
				t.Fatalf("結局第 %d 頁：%v", page+1, err)
			}
		}
		if _, ok := state.OriginalChoiceIndex("PROGRAM_END"); !ok {
			t.Fatalf("勝利之後沒有結束選項：message=%q choices=%v",
				state.Message, state.currentOriginalChoices)
		}
		saveIndex, ok := state.OriginalChoiceIndex("PROGRAM_WIN_SAVE")
		if !ok {
			t.Fatalf("勝利之後沒有存檔選項：%v", state.currentOriginalChoices)
		}
		if err := state.Select(saveIndex); err != nil {
			t.Fatalf("勝利存檔：%v", err)
		}
		if state.Mode != ModeTitle {
			t.Fatalf("勝利存檔之後模式是 %v", state.Mode)
		}
	}) {
		t.FailNow()
	}

	// SEG-32 的語系不變量：整條主線上玩家看得到的每一句話都要是中文。
	// 判準是「有沒有漢字」——原作文字整段是大寫英文，沒有漢字而有英文字母
	// 就是落回原文。
	t.Run("語系：整條主線沒有落回原文", func(t *testing.T) {
		if len(observer.messages) < 100 {
			t.Fatalf("只記到 %d 句話，這條 session 不可能這麼短", len(observer.messages))
		}
		var fallbacks []string
		for message := range observer.messages {
			if campaignMessageHasHan(message) {
				continue
			}
			if !campaignMessageHasLatinWord(message) {
				continue
			}
			fallbacks = append(fallbacks, message)
		}
		sort.Strings(fallbacks)
		for _, message := range fallbacks {
			t.Errorf("落回原文：%q", message)
		}
		t.Logf("整條主線記到 %d 句話，落回原文 %d 句", len(observer.messages), len(fallbacks))
	})

	// SEG-31 的隊伍連續性：**同一個角色**的經驗值不會倒退，隊伍成員只在劇情
	// 安排的地方變動。NPC 同伴會依劇情加入與離開，所以比的是逐個角色，不是總和。
	t.Run("隊伍連續性", func(t *testing.T) {
		changes := []string{}
		previous := campaignSegmentEnd{}
		for index, end := range segmentEnds {
			if len(end.roster) == 0 {
				t.Errorf("%s 的隊伍是空的", end.name)
			}
			if index > 0 {
				for name, value := range end.experience {
					if before, ok := previous.experience[name]; ok && value < before {
						t.Errorf("%s 在 %s 的經驗值從 %d 掉到 %d",
							name, end.name, before, value)
					}
				}
				if strings.Join(end.roster, "／") != strings.Join(previous.roster, "／") {
					changes = append(changes,
						end.name+"：["+strings.Join(end.roster, "／")+"]")
				}
			}
			previous = end
		}
		// 隊伍變動的位置是宣告好的：多一處或少一處都代表同伴的加入／離開被改了。
		want := []string{
			"ECL1/0x50 世界路線：艾森布拉到希爾斯法：[戰士]",
			"ECL3/0x11 猶拉什：地下第一層：[戰士／ALIAS／DRAGONBAIT]",
			"ECL3/0x11 返回與猶拉什邊界：[戰士]",
		}
		if strings.Join(changes, "｜") != strings.Join(want, "｜") {
			t.Errorf("隊伍變動的位置是 %v，宣告的是 %v", changes, want)
		}
	})

	// SEG-33 的音樂綁定：每一段結束時都該有一首正在播的曲子、曲名在 game pack
	// 的曲目表裡，**而且是 PC-98 原作在那個 block 會選的那一首**（spec 355）。
	//
	// ⚠ 只驗「有一首在曲目表裡的曲子」擋不住「每一段都播同一首」。
	t.Run("音樂綁定", func(t *testing.T) {
		pack, err := gamepack.Default()
		if err != nil {
			t.Fatal(err)
		}
		distinct := map[string]bool{}
		for _, end := range segmentEnds {
			if end.music == "" {
				t.Errorf("%s 結束時沒有正在播的曲子", end.name)
				continue
			}
			if _, found := pack.FindMusicTrack(end.music); !found {
				t.Errorf("%s 的曲目 %q 不在 game pack 的曲目表裡", end.name, end.music)
			}
			distinct[end.music] = true
			choices, declared := expectedMusicForBlock(end.block)
			if !declared {
				// `0x30` 不換曲、`0x52` 沒有分支：那兩段沿用前一段的曲子。
				continue
			}
			if !slices.Contains(choices, end.music) {
				t.Errorf("%s（block %#02X）結束時播的是 %s，spec 355 說應該是 %v",
					end.name, end.block, end.music, choices)
			}
		}
		// ⚠ 非空還不夠：整條主線只播出一首也會讓上面每一條都過。
		if len(distinct) < 4 {
			t.Errorf("整條主線只播出 %d 首不同的曲子，選曲的比對等於沒驗", len(distinct))
		}
	})

	// SEG-31 後半：裝備、記憶法術與效果的跨段連續性。
	//
	// ⚠ 這一條**先擋非空**再擋變動。只擋「沒有變動」的話，哪天攜帶物整批讀不
	// 到、每一段都是空的，測試照樣全綠——空的等於空的。
	t.Run("攜帶物的跨段連續性", func(t *testing.T) {
		richestEquipment, castersSeen, effectsSeen := 0, 0, 0
		for _, end := range segmentEnds {
			for label := range end.equipment {
				count := 0
				if end.equipment[label] != "" {
					count = strings.Count(end.equipment[label], ",") + 1
				}
				if count > richestEquipment {
					richestEquipment = count
				}
				if !strings.Contains(end.spells[label], "記憶=[]") {
					castersSeen++
				}
				if end.effects[label] != "" {
					effectsSeen++
				}
			}
		}
		if richestEquipment < 4 {
			t.Errorf("整條路線上最多裝備的角色只有 %d 件，攜帶物的比對等於沒驗", richestEquipment)
		}
		if castersSeen == 0 {
			t.Error("整條路線上沒有一名角色帶著記憶法術，法術的比對等於沒驗")
		}
		if effectsSeen == 0 {
			t.Error("整條路線上沒有一名角色帶著效果，效果的比對等於沒驗")
		}

		// 同一個角色的攜帶物在相鄰兩段之間的變動位置是宣告好的。這條路線上沒有
		// 購買、消耗或劇情剝奪，所以清單是空的——多出一處就代表有東西被靜靜
		// 改掉或掉了。
		changes := []string{}
		previous := campaignSegmentEnd{}
		for index, end := range segmentEnds {
			if index > 0 {
				for label, value := range end.equipment {
					if before, ok := previous.equipment[label]; ok && before != value {
						changes = append(changes, end.name+"／"+label+"：裝備")
					}
				}
				for label, value := range end.spells {
					if before, ok := previous.spells[label]; ok && before != value {
						changes = append(changes, end.name+"／"+label+"：法術")
					}
				}
				for label, value := range end.effects {
					if before, ok := previous.effects[label]; ok && before != value {
						changes = append(changes, end.name+"／"+label+"：效果")
					}
				}
			}
			previous = end
		}
		sort.Strings(changes)
		if len(changes) != 0 {
			t.Errorf("攜帶物在段界之間變動的位置是 %v，宣告的是沒有變動", changes)
		}
	})

	// P0 的 save／reload gate：讀回來之後**還走得動**。
	//
	// ⚠ 「欄位對得上」與「還能繼續玩」是兩件事。上一個 subtest 驗的是前者；
	// 存檔如果掉了 ECL 的續跑位置或地城的生命週期上下文，欄位照樣全對，
	// 但玩家按下一步就會卡住或噴錯——而且不會有任何欄位比對抓得到。
	t.Run("段界快照讀回來還走得動", func(t *testing.T) {
		blocks := map[uint8][]byte{}
		records := map[uint8]map[uint8]monster.Record{}
		for member := 1; member <= 6; member++ {
			parsed, err := dax.Parse(zipData(t, image, "ECL"+strconv.Itoa(member)+".DAX"))
			if err != nil {
				t.Fatal(err)
			}
			for _, block := range parsed {
				blocks[block.Entry.ID] = block.Data
			}
			monsterBlocks, err := dax.Parse(zipData(t, image,
				"MON"+strconv.Itoa(member)+"CHA.DAX"))
			if err != nil {
				t.Fatal(err)
			}
			chapter := map[uint8]monster.Record{}
			for _, block := range monsterBlocks {
				record, parseErr := monster.Parse(block.Data)
				if parseErr != nil {
					t.Fatalf("MON%dCHA block %#02x: %v", member, block.Entry.ID, parseErr)
				}
				chapter[block.Entry.ID] = record
			}
			records[uint8(member)] = chapter
		}
		stepped := 0
		for _, end := range segmentEnds {
			restored := NewStateFromECLBlocks(testCatalog(), blocks, 0x50)
			for chapter, chapterRecords := range records {
				restored.SetMonsterRecordsForECL(chapter, chapterRecords)
			}
			if err := restored.LoadPartyFile(end.path); err != nil {
				t.Errorf("%s 讀不回來：%v", end.name, err)
				continue
			}
			if err := stepRestoredCampaignState(&restored); err != nil {
				t.Errorf("%s 讀回來之後走不動：%v", end.name, err)
				continue
			}
			stepped++
			// ⚠ 只驗「走得動 ＋ 隊伍還在」。**不驗 block 有沒有變**：世界地圖那
			// 幾段的下一步就是出發旅行，換 block 是對的。
			if len(restored.PartyFighters()) == 0 {
				t.Errorf("%s 讀回來走一步之後隊伍空了", end.name)
			}
		}
		// ⚠ 非空閘門：走得動的份數如果塌下來，上面每一條都不會紅。
		if stepped < len(segmentEnds) {
			t.Errorf("%d／%d 份段界快照讀回來走得動", stepped, len(segmentEnds))
		}
	})

	// 段內支線：主線走到某一段的當下，那張地圖上除了主線路線以外的格子演不演
	// 得出來。★ `cmd/cell-sweep` 每一格都重新進段，答的是「這一格演不演得出來」；
	// 這裡答的是**「帶著主線進度走到那一格，演不演得出來」**——處理常式開頭那些
	// `COMPARE 4C2D 01 / IF = / EXIT` 的守衛只有帶著進度才滿足得了。
	t.Run("段內支線在主線進度下演得出來", func(t *testing.T) {
		blocks := map[uint8][]byte{}
		records := map[uint8]map[uint8]monster.Record{}
		catalogs := map[uint8]geo.Catalog{}
		for member := 1; member <= 6; member++ {
			parsed, err := dax.Parse(zipData(t, image, "ECL"+strconv.Itoa(member)+".DAX"))
			if err != nil {
				t.Fatal(err)
			}
			for _, block := range parsed {
				blocks[block.Entry.ID] = block.Data
			}
			monsterBlocks, err := dax.Parse(zipData(t, image, "MON"+strconv.Itoa(member)+"CHA.DAX"))
			if err != nil {
				t.Fatal(err)
			}
			chapter := map[uint8]monster.Record{}
			for _, block := range monsterBlocks {
				record, parseErr := monster.Parse(block.Data)
				if parseErr != nil {
					t.Fatalf("MON%dCHA block %#02x: %v", member, block.Entry.ID, parseErr)
				}
				chapter[block.Entry.ID] = record
			}
			records[uint8(member)] = chapter
			// 第一章沒有 `GEO1.DAX`（世界地圖不走地城地圖）。
			if payload := optionalZipData(image, "GEO"+strconv.Itoa(member)+".DAX"); payload != nil {
				catalog := geo.NewCatalog()
				if err := catalog.AddDAX(uint8(member), payload); err != nil {
					t.Fatalf("GEO%d: %v", member, err)
				}
				catalogs[uint8(member)] = catalog
			}
		}

		swept, playedCells, fallbacks := 0, 0, []string{}
		// ⚠ 記的是「**掃成功**的 block」，不是「看過的 block」。同一個 block 常有
		// 好幾份段界快照，前面那份可能停在世界地圖上推不回地城；用「看過就跳過」
		// 會把整段判成掃不到，而那不是事實。
		sweptBlock := map[uint8]bool{}
		noDispatcher := []string{}
		for _, end := range segmentEnds {
			if sweptBlock[end.block] {
				continue
			}
			// ⚠ 語系要用**正式語系檔**，不是 `testCatalog()`。兄弟 subtest 只比
			// 欄位，用最小語系檔無妨；這裡量的是玩家看到的字，最小語系檔會讓
			// 每一句都退回 stable id（`party_ready`），整批看起來像沒中文化。
			catalog := trainingTestCatalog(t)
			restore := func() (State, error) {
				restored := NewStateFromECLBlocks(catalog, blocks, 0x50)
				for chapter, chapterRecords := range records {
					restored.SetMonsterRecordsForECL(chapter, chapterRecords)
				}
				if err := restored.LoadPartyFile(end.path); err != nil {
					return restored, err
				}
				return restored, nil
			}
			sweep := sweepSideBranches(end.name, restore, blocks, catalogs)
			t.Log(sweep.summarize())
			if sweep.skipped != "" {
				noDispatcher = append(noDispatcher,
					fmt.Sprintf("%s：%s", end.name, sweep.skipped))
				continue
			}
			sweptBlock[end.block] = true
			sweptBlock[sweep.block] = true
			swept++
			for _, cell := range sweep.cells {
				if !cell.played() {
					continue
				}
				playedCells++
				// 語系跟主線同一個判準：沒有漢字而有連續英文字母就是落回原文。
				if campaignMessageHasHan(cell.text) || !campaignMessageHasLatinWord(cell.text) {
					continue
				}
				fallbacks = append(fallbacks, fmt.Sprintf("%s 索引 %d (%d,%d)：%q",
					sweep.name, cell.index, cell.x, cell.y, cell.text))
			}
		}
		sort.Strings(fallbacks)
		for _, message := range fallbacks {
			t.Errorf("段內支線落回原文：%s", message)
		}
		// ⚠ 非空閘門：走訪壞掉會變成「語系全綠但什麼都沒驗到」。掃得到的段數與
		// 演得出來的格數任一塌下來，上面那條語系檢查都不會紅。
		if swept < 8 {
			t.Errorf("只掃到 %d 段的段內支線，太少", swept)
		}
		if playedCells < 110 {
			t.Errorf("段內支線只演出 %d 格，太少", playedCells)
		}
		// 沒有地形分派器的段一律列出來，不要靜靜地不算進分母。
		// 掃不到的段一律列出來並附理由，不要靜靜地不算進分母。
		for _, note := range noDispatcher {
			t.Logf("掃不到：%s", note)
		}
		t.Logf("段內支線：%d 段掃得到，%d 格演得出來，落回原文 %d 格",
			swept, playedCells, len(fallbacks))
	})

	t.Run("段界快照往返", func(t *testing.T) {
		if len(segmentEnds) != 23 {
			t.Fatalf("存到 %d 份段界快照，應該是 23 份", len(segmentEnds))
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
			// SEG-31 後半：交接走的是快照，所以「裝備／記憶法術／效果存下去
			// 讀得回來」就是跨段不變量本身。⚠ 人數對不代表身上的東西還在。
			for _, character := range restored.partyRoster {
				label := character.ScriptName
				if label == "" {
					label = character.Name
				}
				if want, ok := end.equipment[label]; ok {
					if got := campaignEquipmentSignature(character); got != want {
						t.Errorf("%s 的 %s 讀回來裝備是 %q，存的是 %q",
							end.name, label, got, want)
					}
				}
				if want, ok := end.spells[label]; ok {
					if got := campaignSpellSignature(character); got != want {
						t.Errorf("%s 的 %s 讀回來法術是 %q，存的是 %q",
							end.name, label, got, want)
					}
				}
				if want, ok := end.effects[label]; ok {
					if got := campaignEffectSignature(character); got != want {
						t.Errorf("%s 的 %s 讀回來效果是 %q，存的是 %q",
							end.name, label, got, want)
					}
				}
			}
		}
	})
}

// campaignSegmentEnd 是一段走完時的邊界狀態，用來驗那份快照讀得回來。
type campaignSegmentEnd struct {
	name       string
	path       string
	block      uint8
	mode       Mode
	area       uint8
	inDungeon  bool
	party      int
	roster     []string
	experience map[string]uint32
	music      string
	// 下面三個是 `SEG-31` 後半：跨段的**攜帶物**。段與段之間的交接一律走快照，
	// 所以「存下去讀回來還在不在」就是跨段不變量本身。
	equipment  map[string]string
	spells     map[string]string
	effects    map[string]string
}

// campaignEquipmentSignature 把一名角色的裝備列成穩定字串。⚠ 要帶上「有沒有
// 裝備中」與數量：只比件數的話，武器被卸下或用掉一支都看不出來。
func campaignEquipmentSignature(character party.Character) string {
	parts := make([]string, 0, len(character.Equipment))
	for _, item := range character.Equipment {
		readied := ""
		if item.Readied {
			readied = "*"
		}
		parts = append(parts, fmt.Sprintf("%d%s×%d+%d", item.Type, readied, item.Count, item.Plus))
	}
	return strings.Join(parts, ",")
}

// campaignSpellSignature 把記憶法術與法術書列成穩定字串。
func campaignSpellSignature(character party.Character) string {
	return fmt.Sprintf("記憶=%v 法術書=%v 容量=%v",
		character.SpellSlots, character.KnownSpells, character.SpellCastCount)
}

// campaignEffectSignature 把身上的效果列成穩定字串。
func campaignEffectSignature(character party.Character) string {
	parts := make([]string, 0, len(character.Effects))
	for _, effect := range character.Effects {
		active := ""
		if effect.Active {
			active = "*"
		}
		parts = append(parts, fmt.Sprintf("%02X%s/%d", effect.Kind, active, effect.Duration))
	}
	return strings.Join(parts, ",")
}

// stepRestoredCampaignState 讓讀回來的狀態**真的走一步**，走哪一種依它停在哪。
//
// ⚠ 這不是「跑到下一個劇情點」，只是「按得下去」。段與段之間的完整續接由
// `SEG-12` 的 47 條邊負責。
func stepRestoredCampaignState(state *State) error {
	switch {
	case state.CombatActive():
		return state.CombatAct()
	case state.Mode == ModeDungeon:
		return state.RunDungeonLifecycle()
	case len(state.Choices) > 0 && state.Mode != ModeEvent:
		return state.Select(0)
	default:
		return state.Continue()
	}
}

// campaignMessageHasHan 判斷一句話裡有沒有漢字。
func campaignMessageHasHan(message string) bool {
	for _, glyph := range message {
		if unicode.Is(unicode.Han, glyph) {
			return true
		}
	}
	return false
}

// campaignMessageHasLatinWord 判斷一句話裡有沒有連續兩個以上的英文字母。
// 單獨的字母多半是數字旁的單位或原作的代號，不算落回原文。
func campaignMessageHasLatinWord(message string) bool {
	run := 0
	for _, glyph := range message {
		if (glyph >= 'A' && glyph <= 'Z') || (glyph >= 'a' && glyph <= 'z') {
			run++
			if run >= 2 {
				return true
			}
			continue
		}
		run = 0
	}
	return false
}

// resolveInnerRoom 把房間事件推到底：事件按繼續、選單選第一項、遇上就打。
func resolveInnerRoom(t *testing.T, state *State, observer *normalCampaignObserver) {
	t.Helper()
	// 先記一次：有些房間的敘述不會停下來等玩家，迴圈的條件會直接把它跳過。
	observer.observe()
	for step := 0; step < 16 && state.Mode != ModeDungeon; step++ {
		if state.CombatActive() {
			for turn := 0; turn < 400 && state.CombatActive(); turn++ {
				if err := state.CombatAct(); err != nil {
					t.Fatalf("房間戰鬥第 %d 回合：%v", turn, err)
				}
			}
			observer.observe()
			continue
		}
		if err := state.Continue(); err != nil {
			if selectErr := state.Select(0); selectErr != nil {
				t.Fatalf("房間事件推不動：continue=%v select=%v", err, selectErr)
			}
		}
		observer.observe()
	}
}
