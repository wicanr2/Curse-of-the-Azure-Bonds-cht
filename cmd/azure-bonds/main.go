package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	partySave "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/save"
)

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "original DOS image ZIP")
	member := flag.String("member", "ECL1.DAX", "DAX member to inspect")
	trace := flag.Bool("trace", false, "trace known ECL cursor commands")
	traceStart := flag.Int("trace-start", -1, "decoded payload offset for -trace (default 0)")
	runStart := flag.Int("run-start", -1, "decoded payload offset for -run-subset (default initial entry)")
	blockID := flag.Int("block-id", -1, "inspect only this decimal DAX block ID")
	stringsOnly := flag.Bool("strings", false, "print ECL packed-text candidates")
	graph := flag.Bool("graph", false, "trace statically reachable ECL branches")
	allEntries := flag.Bool("all-entries", false, "use all five ECL initialization entries with -graph")
	findOpcode := flag.Int("find-opcode", -1, "print reachable instructions with this opcode when used with -graph")
	scanOpcode := flag.Int("scan-opcode", -1, "print linearly scanned instruction candidates with this opcode")
	monsterRecord := flag.Bool("monster-record", false, "decode the selected block as a MON*CHA record")
	monsterItems := flag.Bool("monster-items", false, "decode the selected block as MON*ITM records")
	monsterAffects := flag.Bool("monster-affects", false, "decode the selected block as MON*SPC records")
	baseItems := flag.Bool("base-items", false, "decode the standalone ITEMS base-item catalog")
	encounterStart := flag.Int("encounter-start", -1, "decode LOAD MONSTER sequence at this payload offset")
	monsterMember := flag.String("monster-member", "MON1CHA.DAX", "MON*CHA member for -encounter-start")
	entryPoints := flag.Bool("entrypoints", false, "print five ECL initialization entry points")
	entrySmoke := flag.Bool("entry-smoke", false, "run each ECL initialization entry with bounded semantics")
	runSubset := flag.Bool("run-subset", false, "run the bounded ECL command subset from the initial entry")
	interactive := flag.Bool("interactive", false, "pause -run-subset at the first unselected menu")
	sessionInfo := flag.Bool("session", false, "print decoded ECL block session entries")
	selectionList := flag.String("select", "", "comma-separated HORIZONTAL MENU selections for -run-subset")
	importCharacter := flag.Bool("import-character", false, "import one DOS .SAV/.GUY character with optional .FX/.SWG sidecars")
	characterID := flag.String("character-id", "dos-character", "remake ID for -import-character")
	characterRecord := flag.String("character-record", "", "DOS .SAV or .GUY path for -import-character")
	characterEffects := flag.String("character-effects", "", "optional DOS .FX path for -import-character")
	characterInventory := flag.String("character-inventory", "", "optional DOS .SWG path for -import-character")
	outParty := flag.String("out-party", "", "write imported character as versioned party JSON")
	setAge := flag.String("set-age", "", "set the signed DOS player age when used with -character-record")
	outRecord := flag.String("out-record", "", "write a patched DOS .SAV/.GUY record to this new path")
	flag.Parse()
	if strings.TrimSpace(*setAge) != "" {
		if strings.TrimSpace(*characterRecord) == "" {
			log.Fatal("-character-record is required with -set-age")
		}
		if strings.TrimSpace(*outRecord) == "" {
			log.Fatal("-out-record is required with -set-age; the source record is never overwritten")
		}
		parsedAge, err := strconv.ParseInt(strings.TrimSpace(*setAge), 10, 16)
		if err != nil {
			log.Fatalf("invalid -set-age %q: %v (expected -32768..32767)", *setAge, err)
		}
		record, err := os.ReadFile(*characterRecord)
		if err != nil {
			log.Fatal(err)
		}
		character, err := party.ParseDOSPlayerRecord(record, "cli-age-patch")
		if err != nil {
			log.Fatal(err)
		}
		before := character.Age
		projected, err := character.Character()
		if err != nil {
			log.Fatal(err)
		}
		projected.Age = int16(parsedAge)
		patched, err := party.PatchDOSPlayerRecord(record, projected)
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(*outRecord, patched, 0o600); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("已修改 DOS 角色 %q 年齡：%d → %d，輸出：%s\n", character.Name, before, projected.Age, *outRecord)
		return
	}
	if *importCharacter {
		if strings.TrimSpace(*characterRecord) == "" {
			log.Fatal("-character-record is required with -import-character")
		}
		record, err := os.ReadFile(*characterRecord)
		if err != nil {
			log.Fatal(err)
		}
		effects, err := optionalFile(*characterEffects)
		if err != nil {
			log.Fatal(err)
		}
		inventory, err := optionalFile(*characterInventory)
		if err != nil {
			log.Fatal(err)
		}
		character, err := party.ParseDOSPlayerFiles(*characterID, party.DOSPlayerFiles{Record: record, Effects: effects, Inventory: inventory})
		if err != nil {
			log.Fatal(err)
		}
		data, err := partySave.EncodeParty(party.Roster{character})
		if err != nil {
			log.Fatal(err)
		}
		if *outParty == "" {
			fmt.Print(string(data))
		} else if err := os.WriteFile(*outParty, data, 0o600); err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("已匯入 DOS 角色 %q，輸出：%s\n", character.Name, *outParty)
		}
		return
	}
	if *baseItems {
		data, err := zipMember(*image, "ITEMS")
		if err != nil {
			log.Fatal(err)
		}
		catalog, err := monster.ParseBaseItems(data)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ITEMS: header=0x%04X descriptors=%d\n", catalog.Header, len(catalog.Items))
		for _, item := range catalog.Items {
			fmt.Printf("  type=0x%02X zh-name=%q slot=%d hands=%d damage-small=%dd%d%+d damage-large=%dd%d%+d ac-adjust=%d class-mask=0x%02X ammo=%d\n",
				item.Type, monster.ChineseName(monster.ItemRecord{Type: item.Type}), item.Slot, item.HandsRequired,
				item.SmallDamageDice, item.SmallDamageSides, item.SmallDamageBonus,
				item.LargeDamageDice, item.LargeDamageSides, item.LargeDamageBonus,
				item.ACAdjustment, item.ClassUsabilityMask, item.AmmunitionType)
		}
		return
	}
	selections, err := parseSelections(*selectionList)
	if err != nil {
		log.Fatal(err)
	}
	data, err := zipMember(*image, *member)
	if err != nil {
		log.Fatal(err)
	}
	blocks, err := dax.Parse(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s: %d blocks\n", *member, len(blocks))
	if *sessionInfo {
		byID := make(map[uint8][]byte, len(blocks))
		for _, block := range blocks {
			byID[block.Entry.ID] = block.Data
		}
		session, err := ecl.NewBlockSession(byID, blocks[0].Entry.ID)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  session current=0x%02X initial=+0x%04X\n", session.CurrentBlockID(), mustInitial(session))
		for _, block := range blocks {
			entry, entryErr := ecl.NewBlockSession(byID, block.Entry.ID)
			if entryErr != nil {
				log.Fatal(entryErr)
			}
			fmt.Printf("  block 0x%02X initial=+0x%04X\n", entry.CurrentBlockID(), mustInitial(entry))
		}
	}
	for _, block := range blocks {
		if *blockID >= 0 && int(block.Entry.ID) != *blockID {
			continue
		}
		fmt.Printf("block %d: %d decoded bytes\n", block.Entry.ID, len(block.Data))
		if *entrySmoke {
			reports, smokeErr := ecl.SmokeInitializationEntries(block.Data, 5, 500, selections)
			if smokeErr != nil {
				fmt.Printf("  entry smoke stopped safely: %v\n", smokeErr)
			} else {
				for _, report := range reports {
					fmt.Printf("  entry[%d] address=0x%04X start=+0x%04X steps=%d pc=+0x%04X menu=%t combat=%t spawns=%d program=%v err=%v\n",
						report.Index, report.Address, report.Start, report.Result.Steps, report.Result.PC,
						report.Result.WaitingForMenu, report.Result.CombatRequested, len(report.Result.MonsterSpawns),
						report.Result.ProgramIDs, report.Err)
				}
			}
		}
		if *monsterRecord {
			record, recordErr := monster.Parse(block.Data)
			if recordErr != nil {
				fmt.Printf("  monster record stopped safely: %v\n", recordErr)
			} else {
				fmt.Printf("  monster name=%q mod=0x%02X team=%d hp=%d/%d ac=%d attack-bonus=%d damage=%dd%d%+d initiative=%d\n", record.Name, record.ModID, record.CombatTeam, record.HitPoints, record.MaxHitPoints, record.ArmorClass, record.AttackBonus, record.DamageDiceCount, record.DamageDiceSides, record.DamageBonus, record.InitiativeBonus)
				if len(record.Raw) >= monster.RecordSize {
					fmt.Printf("  player race=0x%02X class=0x%02X multiclass-level=%d levels=%v control-morale=0x%02X\n",
						record.Raw[0x74], record.Raw[0x75], record.Raw[0xE6],
						record.Raw[0x109:0x111], record.Raw[0xF7])
				}
			}
		}
		if *monsterItems {
			items, itemErr := monster.ParseItems(block.Data)
			if itemErr != nil {
				fmt.Printf("  item records stopped safely: %v\n", itemErr)
			} else {
				for index, item := range items {
					fmt.Printf("  item[%d] name=%q zh-name=%q type=0x%02X name-numbers=%#v plus=%d count=%d readied=%t cursed=%t affects=%#v\n", index, item.Name, monster.ChineseName(item), item.Type, item.NameNumbers, item.Plus, item.Count, item.Readied, item.Cursed, item.Affects)
				}
			}
		}
		if *monsterAffects {
			affects, affectErr := monster.ParseAffects(block.Data)
			if affectErr != nil {
				fmt.Printf("  affect records stopped safely: %v\n", affectErr)
			} else {
				for index, affect := range affects {
					fmt.Printf("  affect[%d] zh-name=%q kind=0x%02X value=0x%04X duration=%d active=%t data=%#v\n", index, monster.ChineseAffectName(affect), affect.Kind, affect.Value, affect.Duration, affect.Active, affect.Data)
				}
			}
		}
		if *encounterStart >= 0 {
			instructions, traceErr := ecl.TraceAt(block.Data, *encounterStart, 128)
			spawns := make([]ecl.MonsterSpawn, 0)
			for _, instruction := range instructions {
				if instruction.Command.Opcode == 0x24 {
					break
				}
				if instruction.Command.Opcode != 0x0B {
					continue
				}
				spawn, spawnErr := ecl.DecodeMonsterSpawn(instruction)
				if spawnErr != nil {
					fmt.Printf("  encounter spawn stopped safely: %v\n", spawnErr)
					break
				}
				spawns = append(spawns, spawn)
			}
			monsterData, memberErr := zipMember(*image, *monsterMember)
			if memberErr != nil {
				fmt.Printf("  encounter monster member stopped safely: %v\n", memberErr)
			} else if monsterBlocks, parseErr := dax.Parse(monsterData); parseErr != nil {
				fmt.Printf("  encounter monster DAX stopped safely: %v\n", parseErr)
			} else {
				records := make(map[uint8]monster.Record, len(monsterBlocks))
				for _, monsterBlock := range monsterBlocks {
					record, recordErr := monster.Parse(monsterBlock.Data)
					if recordErr == nil {
						records[monsterBlock.Entry.ID] = record
					}
				}
				enemies, enemyErr := monster.BuildEnemies(spawns, records)
				if enemyErr != nil {
					fmt.Printf("  encounter build stopped safely: %v\n", enemyErr)
				} else {
					fmt.Printf("  encounter spawns=%d enemies=%d\n", len(spawns), len(enemies))
					for _, enemy := range enemies {
						fmt.Printf("    enemy id=%s name=%q hp=%d/%d ac=%d attack=%d damage=%dd%d%+d\n", enemy.ID, enemy.Name, enemy.HitPoints, enemy.MaxHitPoints, enemy.ArmorClass, enemy.AttackBonus, enemy.DamageDiceCount, enemy.DamageDiceSides, enemy.DamageBonus)
					}
				}
			}
			if traceErr != nil {
				fmt.Printf("  encounter trace stopped safely: %v\n", traceErr)
			}
		}
		if *scanOpcode >= 0 {
			instructions, scanErr := ecl.ScanKnownInstructions(block.Data)
			for _, instruction := range instructions {
				if instruction.Command.Opcode == byte(*scanOpcode) {
					fmt.Printf("  scan opcode 0x%02X at +0x%04X operands=%#v next=+0x%04X\n", instruction.Command.Opcode, instruction.Offset, instruction.Operands, instruction.Next)
				}
			}
			if scanErr != nil {
				fmt.Printf("  scan stopped safely: %v\n", scanErr)
			}
		}
		if *trace {
			var instructions []ecl.Instruction
			var err error
			if *traceStart >= 0 {
				instructions, err = ecl.TraceAt(block.Data, *traceStart, 40)
			} else {
				instructions, err = ecl.Trace(block.Data, 40)
			}
			for _, instruction := range instructions {
				fmt.Printf("  +0x%04X %-16s operands=%#v\n", instruction.Offset, instruction.Command.Name, instruction.Operands)
				for _, operand := range instruction.Operands {
					if len(operand.Packed) > 0 {
						fmt.Printf("    packed=%q\n", ecl.DecodePackedText(operand.Packed))
					}
				}
			}
			if err != nil {
				fmt.Printf("  trace stopped safely: %v\n", err)
			}
		}
		if *stringsOnly {
			for _, candidate := range ecl.FindPackedTextCandidatesAt(block.Data) {
				fmt.Printf("  +0x%04X text=%q\n", candidate.Offset, candidate.Text)
			}
		}
		if *graph {
			starts := []int(nil)
			if points, _, entryErr := ecl.EntryPoints(block.Data, 5); entryErr == nil && len(points) == 5 {
				entryIndex := 4
				if *allEntries {
					starts = make([]int, 0, len(points))
					for index, point := range points {
						offset := int(point) - ecl.CodeAddressBase
						starts = append(starts, offset)
						fmt.Printf("  graph entry[%d]=+0x%04X (code 0x%04X)\n", index, offset, point)
					}
				} else {
					starts = []int{int(points[entryIndex]) - ecl.CodeAddressBase}
					fmt.Printf("  graph start=+0x%04X (initial entry 0x%04X)\n", starts[0], points[entryIndex])
				}
			} else if entryErr != nil {
				fmt.Printf("  entry point unavailable; graph fallback start=+0x0000: %v\n", entryErr)
			}
			result, err := ecl.TraceGraph(block.Data, starts, 2000)
			fmt.Printf("  graph instructions=%d edges=%d\n", len(result.Instructions), len(result.Edges))
			for _, edge := range result.Edges {
				fmt.Printf("  edge +0x%04X -> +0x%04X (%s)\n", edge.From, edge.To, edge.Kind)
			}
			for _, instruction := range result.Instructions {
				if *findOpcode >= 0 && instruction.Command.Opcode == byte(*findOpcode) {
					fmt.Printf("  found opcode 0x%02X at +0x%04X operands=%#v next=+0x%04X\n", instruction.Command.Opcode, instruction.Offset, instruction.Operands, instruction.Next)
				}
				if instruction.Command.Opcode == 0x20 && len(instruction.Operands) == 1 {
					fmt.Printf("  NEWECL at +0x%04X operand=%#v\n", instruction.Offset, instruction.Operands[0])
				}
			}
			if err != nil {
				fmt.Printf("  graph stopped safely: %v\n", err)
			}
		}
		if *entryPoints {
			points, next, err := ecl.EntryPoints(block.Data, 5)
			if err != nil {
				fmt.Printf("  entry points stopped safely: %v\n", err)
				continue
			}
			fmt.Printf("  entry points cursor=+0x%04X", next)
			for _, point := range points {
				fmt.Printf(" 0x%04X", point)
			}
			fmt.Println()
		}
		if *runSubset {
			start := *runStart
			if start < 0 {
				start = 0
				if points, _, entryErr := ecl.EntryPoints(block.Data, 5); entryErr == nil && len(points) == 5 {
					start = int(points[4]) - ecl.CodeAddressBase
				}
			}
			var result ecl.RunResult
			var runErr error
			if *interactive {
				result, runErr = ecl.RunSubsetInteractive(block.Data, start, 500, selections)
			} else {
				result, runErr = ecl.RunSubsetWithSelections(block.Data, start, 500, selections)
			}
			fmt.Printf("  subset steps=%d stop=+0x%04X texts=%d\n", result.Steps, result.PC, len(result.Text))
			if result.WaitingForMenu {
				fmt.Println("  subset waiting for menu selection")
			}
			if result.NewECLBlockID != nil {
				fmt.Printf("  subset requested ECL block 0x%02X\n", *result.NewECLBlockID)
			}
			if result.CombatRequested {
				fmt.Println("  subset requested COMBAT")
			}
			if result.LoadPiecesRequested {
				fmt.Printf("  subset LOAD PIECES=%v\n", result.LoadPieces)
			}
			if len(result.ProgramIDs) > 0 {
				fmt.Printf("  subset PROGRAM ids=%v external-boundary=%t\n", result.ProgramIDs, result.ProgramExit)
			}
			if len(result.NPCIDs) > 0 {
				fmt.Printf("  subset ADD NPC ids=%v\n", result.NPCIDs)
			}
			for _, message := range result.Text {
				fmt.Printf("    text=%q\n", message)
			}
			for _, menu := range result.Menus {
				fmt.Printf("    menu location=0x%04X selected=%d options=%q\n", menu.Location, menu.Selected, menu.Options)
			}
			if runErr != nil {
				fmt.Printf("  subset stopped safely: %v\n", runErr)
			}
		}
	}
}

func mustInitial(session *ecl.BlockSession) int {
	entry, err := session.InitialEntry()
	if err != nil {
		log.Fatal(err)
	}
	return entry
}

func parseSelections(value string) ([]uint16, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parts := strings.Split(value, ",")
	selections := make([]uint16, 0, len(parts))
	for _, part := range parts {
		parsed, err := strconv.ParseUint(strings.TrimSpace(part), 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid menu selection %q: %w", part, err)
		}
		selections = append(selections, uint16(parsed))
	}
	return selections, nil
}

func optionalFile(path string) ([]byte, error) {
	if strings.TrimSpace(path) == "" {
		return nil, nil
	}
	return os.ReadFile(path)
}

func zipMember(path, member string) ([]byte, error) {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer archive.Close()
	for _, file := range archive.File {
		if file.Name != member {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		return io.ReadAll(reader)
	}
	return nil, fmt.Errorf("member %q not found", member)
}
