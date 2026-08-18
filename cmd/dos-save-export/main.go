// dos-save-export 產生一份**原版格式**的存檔槽，讓原版 DOS 版可以用
// `LOAD SAVED GAME` 直接進到指定的地圖與座標。
//
// ★ 存在的理由是取 oracle：要拿原版的第一人稱畫面當比對基準，就得讓原版站在
// 指定的格子上。開場那一格是密室（劇情要求先找出口），走不出去；用鍵盤序列
// 從建角一路走進城又太脆——任何一步的載入時間漂移都會讓後面每一個鍵錯位。
// 直接餵一份存檔就沒有這個問題：主選單按 `L` 就到。
//
// 兩種模式：
//
//	-base 指向一份**原版自己寫出來的** savgam?.dat 時，只覆寫座標／朝向，
//	      其餘位元組原樣保留。取 oracle 一律用這一種——原版在那份檔案裡放的
//	      runtime 欄位（天空色、ECL 狀態）不必自己重建。
//	沒有 -base 時從零合成，用來驗自己的編碼器。
//
// ⚠ 存檔裡那八筆角色檔名接的是 `.sav`／`.FX`，不是 `.GUY`：原版存檔時會把
// `<名字>.GUY`／`.FX` 改名成 `CHRDAT<槽><序>.SAV`／`.FX`（實測，spec 1134）。
// 少寫 `.FX` 或寫成 `.GUY`，原版會載入成功但隊伍是空的——而空隊伍會退回
// 沒有隊伍的主選單，看起來就像「拒絕讀檔」。
//
// 用法：
//
//	go run ./cmd/dos-save-export -base workplace/orig-savgamb.dat \
//	  -out workplace/dos-oracle/game/SAVE -slot A -x 3 -y 7 -facing 2
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/area"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	partySave "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/save"
)

// 原版存檔的固定前綴版面（spec 181／1134，DOS 側 ECL 區塊是 1E00h）。
const (
	mapPosOffset = 1 + partySave.SAVGAMArea1Size + partySave.SAVGAMArea2Size +
		partySave.SAVGAMRuntimeStateSize + partySave.SAVGAMECLMemorySize
	area1LastXOffset = 1 + 0x1E0
	area1LastYOffset = 1 + 0x1E2
)

func main() {
	out := flag.String("out", "", "輸出目錄（原版的存檔路徑，預設是 C:\\SAVE）")
	slot := flag.String("slot", "A", "存檔槽 A..J")
	name := flag.String("name", "ORACLE", "角色名（1..15 bytes）")
	gameArea := flag.Int("area", 2, "GameArea／章節（1..6）")
	mapBlock := flag.Int("map-block", 1, "Current3DMapBlockID：第一人稱地圖的 GEO 區塊")
	posX := flag.Int("x", 7, "地圖座標 X")
	posY := flag.Int("y", 13, "地圖座標 Y")
	facing := flag.Int("facing", 0, "朝向 0..7（0 為北、2 東、4 南、6 西）")
	wallType := flag.Int("wall-type", -1, "MapWallType；負值表示沿用 -base 的值")
	wallRoof := flag.Int("wall-roof", -1, "MapWallRoof（>= 0x80 為室內）；負值表示沿用 -base 的值")
	gameState := flag.Int("game-state", 0, "GameState（只在沒有 -base 時使用）")
	lastGameState := flag.Int("last-game-state", 0, "LastGameState（只在沒有 -base 時使用）")
	inDungeon := flag.Bool("in-dungeon", true, "InDungeon：第一人稱地圖為真（只在沒有 -base 時使用）")
	city := flag.Int("city", 0, "CurrentCity（只在沒有 -base 時使用）")
	charRef := flag.String("char-ref", "", "存檔裡記的角色檔名；留空用 CHRDAT<槽>1 並一併寫出該檔")
	base := flag.String("base", "", "以既有的原版 savgam?.dat 為底，只覆寫座標／朝向")
	flag.Parse()

	if *out == "" {
		log.Fatal("-out is required")
	}
	key := (*slot)[0]
	if key >= 'a' && key <= 'j' {
		key -= 'a' - 'A'
	}
	if key < 'A' || key > 'J' {
		log.Fatalf("-slot %q 不在 A..J", *slot)
	}
	if *facing < 0 || *facing > 7 {
		log.Fatalf("-facing %d 不在 0..7", *facing)
	}

	prefixName := fmt.Sprintf("savgam%c.dat", key+('a'-'A'))
	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatal(err)
	}

	if *base != "" {
		prefix, err := os.ReadFile(*base)
		if err != nil {
			log.Fatal(err)
		}
		if len(prefix) != partySave.SAVGAMFixedPrefixSize {
			log.Fatalf("-base %s 長度 %d，不是原版的 %d", *base, len(prefix), partySave.SAVGAMFixedPrefixSize)
		}
		patched := append([]byte(nil), prefix...)
		// Area1 的 LastX／LastY 與地圖那五格都要改：前者是「上一個座標」，
		// 後者才是載入之後站的位置，只改一邊會在畫面與地圖標記之間打架。
		patched[area1LastXOffset] = byte(*posX)
		patched[area1LastXOffset+1] = 0
		patched[area1LastYOffset] = byte(*posY)
		patched[area1LastYOffset+1] = 0
		patched[mapPosOffset] = byte(*posX)
		patched[mapPosOffset+1] = byte(*posY)
		patched[mapPosOffset+2] = byte(*facing)
		if *wallType >= 0 {
			patched[mapPosOffset+3] = byte(*wallType)
		}
		if *wallRoof >= 0 {
			patched[mapPosOffset+4] = byte(*wallRoof)
		}
		if err := os.WriteFile(filepath.Join(*out, prefixName), patched, 0o644); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s（%d bytes，以 %s 為底）位置=(%d,%d) 朝向=%d wallType=0x%02X wallRoof=0x%02X\n",
			prefixName, len(patched), *base, *posX, *posY, *facing,
			patched[mapPosOffset+3], patched[mapPosOffset+4])
		return
	}

	if len(*name) < 1 || len(*name) > 15 {
		log.Fatalf("-name 要 1..15 bytes，給的是 %d", len(*name))
	}
	state := area.State{
		GameArea:            uint8(*gameArea),
		HeadBlockID:         0xFF,
		InDungeon:           *inDungeon,
		Current3DMapBlockID: uint8(*mapBlock),
		CurrentCity:         uint8(*city),
		LastXPos:            int16(*posX),
		LastYPos:            int16(*posY),
	}
	area1, err := area.EncodeArea1(state, nil)
	if err != nil {
		log.Fatal(err)
	}
	area2, err := area.EncodeArea2(state, nil)
	if err != nil {
		log.Fatal(err)
	}

	baseName := fmt.Sprintf("CHRDAT%c1", key)
	ref := baseName
	if *charRef != "" {
		ref = *charRef
	}
	characterRef, refErr := partySave.SAVGAMCharacterRef(ref)
	if refErr != nil {
		log.Fatal(refErr)
	}
	container := partySave.SAVGAMContainer{
		GameArea:      uint8(*gameArea),
		Area1:         area1,
		Area2:         area2,
		Runtime:       make([]byte, partySave.SAVGAMRuntimeStateSize),
		ECL:           make([]byte, partySave.SAVGAMECLMemorySize),
		MapPosX:       int8(*posX),
		MapPosY:       int8(*posY),
		MapDirection:  uint8(*facing),
		MapWallType:   uint8(max(*wallType, 0)),
		MapWallRoof:   uint8(max(*wallRoof, 0)),
		LastGameState: uint8(*lastGameState),
		GameState:     uint8(*gameState),
		PartyCount:    1,
		CharacterRefs: [partySave.SAVGAMCharacterRefCount][]byte{characterRef},
	}
	prefix, err := partySave.EncodeSAVGAM(container)
	if err != nil {
		log.Fatal(err)
	}

	// 一名人類戰士。數值取得夠高，讓 oracle 不會在路上被打死。
	character := party.Character{
		ID:           "oracle",
		Name:         *name,
		Race:         party.RaceHuman,
		Class:        party.ClassFighter,
		RawClassID:   2,
		Level:        5,
		ClassLevels:  [8]uint8{0, 0, 5},
		Abilities:    party.Abilities{Strength: 18, Intelligence: 12, Wisdom: 12, Dexterity: 16, Constitution: 16, Charisma: 12},
		HitPoints:    45,
		MaxHitPoints: 45,
		IconSize:     2,
	}
	record, err := party.PatchDOSPlayerRecord(make([]byte, party.DOSPlayerRecordSize), character)
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(*out, prefixName), prefix, 0o644); err != nil {
		log.Fatal(err)
	}
	if *charRef == "" {
		if err := os.WriteFile(filepath.Join(*out, baseName+".sav"), record, 0o644); err != nil {
			log.Fatal(err)
		}
		// 效果檔一定要寫：原版存檔時 `<名字>.FX` 會跟著改名，載入端也照樣開。
		if err := os.WriteFile(filepath.Join(*out, baseName+".FX"),
			make([]byte, monster.AffectRecordSize*3), 0o644); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("%s（%d bytes）與 %s.sav／.FX 已寫到 %s\n",
		prefixName, len(prefix), baseName, *out)
	fmt.Printf("槽 %c：area=%d block=0x%02X 位置=(%d,%d) 朝向=%d in-dungeon=%v\n",
		key, *gameArea, *mapBlock, *posX, *posY, *facing, *inDungeon)
}
