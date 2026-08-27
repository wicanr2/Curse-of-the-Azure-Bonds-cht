package game

import (
	"archive/zip"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

func TestRealShadowdaleToVoonlarArmyBranches(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	blocks, records := loadAllECLAndMonsters(t, image)

	newRoute := func(condemned bool) *State {
		state := NewStateFromECLBlocks(trainingTestCatalog(t), blocks, 0x50)
		for area, areaRecords := range records {
			state.SetMonsterRecordsForECL(area, areaRecords)
		}
		if err := state.SetParty([]combat.Fighter{{
			ID: "world-route-probe", Name: "World Route Probe", Side: combat.SideParty,
			HitPoints: 999, MaxHitPoints: 999, ArmorClass: -10,
			AttackBonus: 100, DamageDiceCount: 1, DamageDiceSides: 1, DamageBonus: 100,
			AttacksPerTurn: 8, InitiativeBonus: 100,
		}}); err != nil {
			t.Fatal(err)
		}
		if condemned {
			state.session.SetMemoryValue(0x4C5A, 1)
		}
		return &state
	}
	reachVanguard := func(condemned bool) *State {
		state := newRoute(condemned)
		// 分段驗收從暗影谷的原版抵達 entry 起跑；之後只走玩家看得到的
		// JOURNEY ON → VOONLAR → TRAIL，不寫 4C9D 或事件分派索引。
		if err := state.arriveAtWorldLocation(1); err != nil {
			t.Fatal(err)
		}
		if err := state.Select(1); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(state.currentOriginalChoices,
			[]string{"TILVERTON", "ASHABENFORD", "DAGGER FALLS", "VOONLAR"}) {
			t.Fatalf("Shadowdale destinations=%v", state.currentOriginalChoices)
		}
		if err := state.Select(3); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(state.currentOriginalChoices, []string{"TRAIL", "WILDERNESS", "EXIT"}) {
			t.Fatalf("Voonlar routes=%v message=%q", state.currentOriginalChoices, state.Message)
		}
		if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
		if state.Message != requireGamePackText(t, state, "world.wilderness.voonlar-vanguard") ||
			state.Mode != ModeEvent {
			t.Fatalf("Voonlar vanguard choices=%v message=%q", state.currentOriginalChoices, state.Message)
		}
		if err := state.Continue(); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(state.currentOriginalChoices, []string{"MEET THEM", "ATTACK", "HIDE"}) {
			t.Fatalf("Voonlar vanguard actions=%v", state.currentOriginalChoices)
		}
		return state
	}
	finishCombat := func(state *State) {
		for turn := 0; turn < 200 && state.CombatActive(); turn++ {
			if err := state.CombatAct(); err != nil {
				t.Fatal(err)
			}
		}
		if state.CombatActive() {
			t.Fatal("Voonlar army combat did not finish")
		}
	}

	t.Run("hide", func(t *testing.T) {
		state := reachVanguard(false)
		if err := state.Select(2); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP"}) {
			t.Fatalf("hide arrival choices=%v message=%q", state.currentOriginalChoices, state.Message)
		}
	})
	t.Run("meet and pass", func(t *testing.T) {
		state := reachVanguard(false)
		if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
		if state.Message != requireGamePackText(t, state, "world.wilderness.marching-on-shadowdale") {
			t.Fatalf("meet result=%q", state.Message)
		}
		if !reflect.DeepEqual(state.currentOriginalChoices, []string{"ATTACK", "PASS"}) {
			t.Fatalf("officer choices=%v", state.currentOriginalChoices)
		}
		if err := state.Select(1); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP"}) {
			t.Fatalf("pass arrival choices=%v message=%q", state.currentOriginalChoices, state.Message)
		}
	})
	t.Run("meet and attack", func(t *testing.T) {
		state := reachVanguard(false)
		if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
		if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
		if !state.CombatActive() {
			t.Fatalf("officer attack mode=%v message=%q", state.Mode, state.Message)
		}
		finishCombat(state)
		if state.Message != requireGamePackText(t, state, "world.wilderness.army-routs-north") {
			t.Fatalf("officer attack result=%q", state.Message)
		}
	})
	t.Run("condemned", func(t *testing.T) {
		state := reachVanguard(true)
		if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
		if state.Message != requireGamePackText(t, state, "world.wilderness.penalty-for-spying") {
			t.Fatalf("condemned result=%q", state.Message)
		}
		if state.Mode == ModeEvent {
			if err := state.Continue(); err != nil {
				t.Fatal(err)
			}
		}
		if !state.CombatActive() {
			t.Fatalf("condemned mode=%v message=%q", state.Mode, state.Message)
		}
		finishCombat(state)
		if state.Message != requireGamePackText(t, state, "world.wilderness.army-routs-north") {
			t.Fatalf("condemned combat result=%q", state.Message)
		}
	})
}

func TestRealDaggerFallsToTeshwaveMercenaryBranch(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	blocks, records := loadAllECLAndMonsters(t, image)
	newRoute := func() *State {
		state := NewStateFromECLBlocks(trainingTestCatalog(t), blocks, 0x50)
		for area, areaRecords := range records {
			state.SetMonsterRecordsForECL(area, areaRecords)
		}
		if err := state.SetParty([]combat.Fighter{{
			ID: "teshwave-route-probe", Name: "Teshwave Route Probe", Side: combat.SideParty,
			HitPoints: 999, MaxHitPoints: 999, ArmorClass: -10,
			AttackBonus: 100, DamageDiceCount: 1, DamageDiceSides: 1, DamageBonus: 100,
			AttacksPerTurn: 8, InitiativeBonus: 100,
		}}); err != nil {
			t.Fatal(err)
		}
		return &state
	}
	pressPages := func(state *State) {
		for step := 0; step < 8 && reflect.DeepEqual(state.currentOriginalChoices,
			[]string{"PRESS BUTTON OR RETURN TO CONTINUE."}); step++ {
			if err := state.Select(0); err != nil {
				t.Fatal(err)
			}
		}
	}
	finishCombat := func(state *State) {
		for turn := 0; turn < 200 && state.CombatActive(); turn++ {
			if err := state.CombatAct(); err != nil {
				t.Fatal(err)
			}
		}
		if state.CombatActive() {
			t.Fatal("Teshwave content-census combat did not finish")
		}
	}
	reachInvestigation := func() *State {
		state := newRoute()
		// 以原版抵達 entry 直入匕首瀑布，之後只用玩家選單走
		// JOURNEY ON → TESHWAVE → WILDERNESS。ECL 自己算出 4C9D=13
		// 與非道路旅行旗標，沒有注入分派索引。
		if err := state.arriveAtWorldLocation(3); err != nil {
			t.Fatal(err)
		}
		if err := state.Select(1); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(state.currentOriginalChoices, []string{"SHADOWDALE", "TESHWAVE"}) {
			t.Fatalf("Dagger Falls destinations=%v", state.currentOriginalChoices)
		}
		if err := state.Select(1); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(state.currentOriginalChoices, []string{"BY BOAT", "WILDERNESS", "EXIT"}) {
			t.Fatalf("Teshwave routes=%v message=%q", state.currentOriginalChoices, state.Message)
		}
		if err := state.Select(1); err != nil {
			t.Fatal(err)
		}
		if state.Message != requireGamePackText(t, state, "world.wilderness.used-path-north") {
			t.Fatalf("Teshwave path message=%q", state.Message)
		}
		if state.Mode == ModeEvent {
			if err := state.Continue(); err != nil {
				t.Fatal(err)
			}
		}
		if !reflect.DeepEqual(state.currentOriginalChoices, []string{"YES", "NO"}) {
			t.Fatalf("Teshwave path choices=%v", state.currentOriginalChoices)
		}
		return state
	}
	reachBriefing := func() *State {
		state := reachInvestigation()
		if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
		if state.Message != requireGamePackText(t, state, "world.wilderness.craggy-peak-patrol") {
			t.Fatalf("Teshwave investigation=%q", state.Message)
		}
		pressPages(state)
		if !reflect.DeepEqual(state.currentOriginalChoices, []string{"ATTACK", "PRETEND TO BE MERCENAIES"}) {
			t.Fatalf("Teshwave briefing choices=%v message=%q", state.currentOriginalChoices, state.Message)
		}
		return state
	}
	reachCommand := func() *State {
		state := reachBriefing()
		if err := state.Select(1); err != nil {
			t.Fatal(err)
		}
		pressPages(state)
		want := []string{"ATTACK THE LEADERS", "SLIP AWAY TO WARN DAGGER FALLS", "MARCH ON DAGGER FALLS", "MARCH ON TESHWAVE"}
		if !reflect.DeepEqual(state.currentOriginalChoices, want) {
			t.Fatalf("Teshwave command choices=%v message=%q", state.currentOriginalChoices, state.Message)
		}
		return state
	}

	t.Run("ignore path", func(t *testing.T) {
		state := reachInvestigation()
		if err := state.Select(1); err != nil {
			t.Fatal(err)
		}
		if state.CombatActive() {
			t.Fatal("ignoring the northern path unexpectedly started combat")
		}
	})
	t.Run("attack patrol", func(t *testing.T) {
		state := reachBriefing()
		if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
		finishCombat(state)
		if state.Message != requireGamePackText(t, state, "teshwave.army-melts-away") {
			t.Fatalf("attack patrol result=%q", state.Message)
		}
	})
	for index, name := range []string{"attack leaders", "warn Dagger Falls", "march on Dagger Falls", "march on Teshwave"} {
		index, name := index, name
		t.Run(name, func(t *testing.T) {
			state := reachCommand()
			if err := state.Select(index); err != nil {
				t.Fatal(err)
			}
			switch index {
			case 0:
				if !state.CombatActive() {
					t.Fatalf("attack leaders mode=%v message=%q", state.Mode, state.Message)
				}
			case 1:
				if state.Message != requireGamePackText(t, state, "teshwave.reach-dagger-falls") {
					t.Fatalf("warn Dagger Falls result=%q", state.Message)
				}
			case 2:
				if state.Message != requireGamePackText(t, state, "teshwave.monsters-infight") {
					t.Fatalf("march on Dagger Falls opening=%q", state.Message)
				}
			case 3:
				if state.Message != requireGamePackText(t, state, "teshwave.sweep-down-melee") {
					t.Fatalf("march on Teshwave opening=%q", state.Message)
				}
			}
			pressPages(state)
			finishCombat(state)
			if index == 0 || index == 2 {
				if state.Message != requireGamePackText(t, state, "teshwave.army-melts-away") {
					t.Fatalf("command %d combat result=%q", index, state.Message)
				}
			}
			if index == 3 && state.Message != requireGamePackText(t, state, "teshwave.both-routed") {
				t.Fatalf("march on Teshwave combat result=%q", state.Message)
			}
			pressPages(state)
			if !reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP"}) {
				t.Fatalf("command %d arrival choices=%v message=%q", index, state.currentOriginalChoices, state.Message)
			}
		})
	}
}

func TestRealDaggerFallsTeshwaveBoatEncounters(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	blocks, records := loadAllECLAndMonsters(t, image)
	state := NewStateFromECLBlocks(trainingTestCatalog(t), blocks, 0x50)
	for area, areaRecords := range records {
		state.SetMonsterRecordsForECL(area, areaRecords)
	}
	if err := state.SetParty([]combat.Fighter{{
		ID: "boat-route-probe", Name: "Boat Route Probe", Side: combat.SideParty,
		HitPoints: 999, MaxHitPoints: 999, ArmorClass: -10,
		AttackBonus: 100, DamageDiceCount: 1, DamageDiceSides: 1, DamageBonus: 100,
		AttacksPerTurn: 8, InitiativeBonus: 100,
	}}); err != nil {
		t.Fatal(err)
	}
	pressPages := func() {
		for step := 0; step < 8 && reflect.DeepEqual(state.currentOriginalChoices,
			[]string{"PRESS BUTTON OR RETURN TO CONTINUE."}); step++ {
			if err := state.Select(0); err != nil {
				t.Fatal(err)
			}
		}
	}
	finishCombat := func() {
		for turn := 0; turn < 200 && state.CombatActive(); turn++ {
			if err := state.CombatAct(); err != nil {
				t.Fatal(err)
			}
		}
		if state.CombatActive() {
			t.Fatal("boat content-census combat did not finish")
		}
	}
	chooseBoat := func(destination int) {
		if err := state.Select(1); err != nil { // JOURNEY ON
			t.Fatal(err)
		}
		if err := state.Select(destination); err != nil {
			t.Fatal(err)
		}
		if len(state.currentOriginalChoices) == 0 || state.currentOriginalChoices[0] != "BY BOAT" {
			t.Fatalf("boat routes=%v message=%q", state.currentOriginalChoices, state.Message)
		}
		if err := state.Select(0); err != nil {
			t.Fatal(err)
		}
	}

	if err := state.arriveAtWorldLocation(3); err != nil {
		t.Fatal(err)
	}
	chooseBoat(1) // Dagger Falls -> Teshwave
	if state.Message != requireGamePackText(t, &state, "hillsfar.boat-forced-ashore") {
		t.Fatalf("first boat encounter=%q", state.Message)
	}
	pressPages()
	if !state.CombatActive() {
		if err := state.Continue(); err != nil {
			t.Fatal(err)
		}
	}
	pressPages()
	if !state.CombatActive() {
		t.Fatalf("first boat encounter mode=%v choices=%v message=%q", state.Mode, state.currentOriginalChoices, state.Message)
	}
	finishCombat()
	if state.Message != requireGamePackText(t, &state, "hillsfar.captain-impressed") {
		t.Fatalf("first boat combat result=%q", state.Message)
	}
	pressPages()
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP"}) {
		t.Fatalf("Teshwave arrival choices=%v message=%q", state.currentOriginalChoices, state.Message)
	}

	// 反向航線使用另一個原版計數器；先正常回到匕首瀑布，
	// 再次搭同一條順向航線，才會命中 4C88 == 2 的海盜逃跑分支。
	chooseBoat(0) // Teshwave -> Dagger Falls
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP"}) {
		if err := state.Continue(); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP"}) {
		t.Fatalf("reverse boat arrival choices=%v message=%q", state.currentOriginalChoices, state.Message)
	}
	chooseBoat(1) // Dagger Falls -> Teshwave, second time
	if state.Message != requireGamePackText(t, &state, "hillsfar.pirates-flee") {
		t.Fatalf("second boat encounter=%q", state.Message)
	}
	if err := state.Continue(); err != nil {
		t.Fatal(err)
	}
	pressPages()
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP"}) {
		t.Fatalf("second Teshwave arrival choices=%v message=%q", state.currentOriginalChoices, state.Message)
	}
}

func TestRealDaggerFallsCityLeaveThenJourneyOnDoesNotReenterPlaces(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	blocks, _ := loadAllECLAndMonsters(t, image)
	state := NewStateFromECLBlocks(trainingTestCatalog(t), blocks, 0x50)

	if err := state.arriveAtWorldLocation(3); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP"}) {
		t.Fatalf("Dagger Falls arrival choices=%v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(0); err != nil { // ENTER CITY
		t.Fatal(err)
	}
	if state.Mode == ModeEvent {
		if err := state.Continue(); err != nil {
			t.Fatal(err)
		}
	}
	places := []string{"INN", "STORE", "BAR", "LEAVE"}
	if !reflect.DeepEqual(state.currentOriginalChoices, places) {
		t.Fatalf("Dagger Falls places=%v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(3); err != nil { // LEAVE
		t.Fatal(err)
	}
	if state.Mode == ModeEvent {
		if err := state.Continue(); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(state.currentOriginalChoices, []string{"ENTER CITY", "JOURNEY ON", "CAMP"}) {
		t.Fatalf("Dagger Falls edge choices=%v message=%q", state.currentOriginalChoices, state.Message)
	}
	if err := state.Select(1); err != nil { // JOURNEY ON
		t.Fatal(err)
	}
	if reflect.DeepEqual(state.currentOriginalChoices, places) {
		t.Fatalf("Dagger Falls JOURNEY ON reentered city places: message=%q", state.Message)
	}
	if len(state.currentOriginalChoices) < 2 {
		t.Fatalf("Dagger Falls JOURNEY ON did not open destinations: choices=%v message=%q",
			state.currentOriginalChoices, state.Message)
	}
}

func TestRealShadowdaleToAshabenfordUsesTheSelectedTravelDispatcher(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	blocks, _ := loadAllECLAndMonsters(t, image)
	state := NewStateFromECLBlocks(trainingTestCatalog(t), blocks, 0x50)

	if err := state.arriveAtWorldLocation(1); err != nil {
		t.Fatal(err)
	}
	if err := state.Select(1); err != nil { // JOURNEY ON
		t.Fatal(err)
	}
	ashabenford := -1
	for index, option := range state.currentOriginalChoices {
		if option == "ASHABENFORD" {
			ashabenford = index
			break
		}
	}
	if ashabenford < 0 {
		t.Fatalf("Shadowdale destinations=%v", state.currentOriginalChoices)
	}
	// Reproduce the continuous-session precondition found by the key-driven
	// run: ECL's current cell still names the previously visited Dagger Falls.
	state.session.SetMemoryValue(0x4C9B, 3)
	if err := state.Select(ashabenford); err != nil {
		t.Fatal(err)
	}
	if state.Area.CurrentCity != 2 || state.OriginalLocation != "ASHABENFORD" {
		t.Fatalf("selected destination was not projected: area=%+v original=%q",
			state.Area, state.OriginalLocation)
	}
	want := requireGamePackText(t, &state, "world-route.ashabenford")
	if state.Message != want {
		t.Fatalf("Ashabenford selected but ECL dispatched another destination: message=%q want=%q choices=%v",
			state.Message, want, state.currentOriginalChoices)
	}
}

// TestRealOverlandArrivalAndRouteGraphCoverage keeps the world-map contract
// honest at the boundary that can be checked without inventing city stories:
// every native location declared by the CoAB pack must be reachable through
// the directed adjacency graph, and ECL1's real arrival entry must project
// each native value into the corresponding world state. Individual city
// events remain owned by their ECL continuation tests.
func TestRealOverlandArrivalAndRouteGraphCoverage(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()

	allBlocks := make(map[uint8][]byte)
	for chapter := 1; chapter <= 6; chapter++ {
		blocks, parseErr := dax.Parse(zipData(t, image, "ECL"+strconv.Itoa(chapter)+".DAX"))
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		for _, block := range blocks {
			allBlocks[block.Entry.ID] = block.Data
		}
	}

	probe := NewStateFromECLBlocks(trainingTestCatalog(t), allBlocks, 0x50)
	if probe.dataPack == nil {
		t.Fatal("CoAB game pack is unavailable")
	}
	overland, found := probe.dataPack.FindMapByKind("overland")
	if !found {
		t.Fatal("CoAB overland map declaration is unavailable")
	}
	if len(overland.Locations) != 14 {
		t.Fatalf("overland locations=%d, want 14", len(overland.Locations))
	}

	declared := make(map[uint8]bool, len(overland.Locations))
	for _, point := range overland.Locations {
		declared[point.Value] = true
	}
	for _, point := range overland.Locations {
		for _, destination := range point.Destinations {
			if !declared[destination] {
				t.Fatalf("location %q points to undeclared destination %d", point.ID, destination)
			}
		}
	}

	// The route table is directed in the game pack. Check that all declared
	// world locations are reachable from the normal opening location without
	// assuming that every road has a reverse edge.
	reachable := map[uint8]bool{0: true}
	for changed := true; changed; {
		changed = false
		for _, point := range overland.Locations {
			if !reachable[point.Value] {
				continue
			}
			for _, destination := range point.Destinations {
				if !reachable[destination] {
					reachable[destination] = true
					changed = true
				}
			}
		}
	}
	for _, point := range overland.Locations {
		if !reachable[point.Value] {
			t.Fatalf("world location %q (native value %d) is unreachable from Tilverton", point.ID, point.Value)
		}
	}

	// Each arrival is executed through the real ECL1 entry 1. The memory
	// values are the normal world-session preconditions used by the existing
	// new-game path; no location or menu text is injected into State.
	for _, point := range overland.Locations {
		point := point
		t.Run(point.ID, func(t *testing.T) {
			state := NewStateFromECLBlocks(trainingTestCatalog(t), allBlocks, 0x50)
			state.session.SetMemoryValue(0x4C59, 1)
			state.session.SetMemoryValue(0x4C5A, 1)
			state.session.SetMemoryValue(0x4C5B, 0xFF)
			if err := state.arriveAtWorldLocation(point.Value); err != nil {
				t.Fatalf("arrival native value %d: %v", point.Value, err)
			}
			if state.Area.CurrentCity != point.Value || state.Area.GameArea != 1 || state.Area.InDungeon {
				t.Fatalf("arrival %q projected area=%+v", point.ID, state.Area)
			}
			if got, ok := state.session.MemoryValue(0x4C9B); !ok || got != uint16(point.Value) {
				t.Fatalf("arrival %q current-location memory=%#x,%v, want %#x", point.ID, got, ok, point.Value)
			}
			if state.Location == LocationWilderness || state.LocationName == "" || state.OriginalLocation == "" {
				t.Fatalf("arrival %q did not project a world location: location=%v name=%q original=%q", point.ID, state.Location, state.LocationName, state.OriginalLocation)
			}
		})
	}
}
