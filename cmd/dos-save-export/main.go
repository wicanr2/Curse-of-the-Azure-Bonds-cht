// dos-save-export 產生一份**原版格式**的存檔槽，讓原版 DOS 版可以用
// `LOAD SAVED GAME` 直接進到指定的地圖與座標。
//
// ★ 存在的理由是取 oracle：要拿原版的第一人稱畫面當比對基準，就得讓原版站在
// 指定的格子上。用鍵盤序列從建角一路走進城太脆——任何一步的載入時間漂移都會
// 讓後面每一個鍵錯位，而錯位的鍵可能剛好按到 `EXIT TO DOS`。直接餵一份存檔
// 就沒有這個問題：主選單按 `L` 就到。
//
// ⚠ 這一支寫的是**原版**版面（名字走原版編碼），與 remake 自己的槽不同；
// 兩者版面相同、只有名字編碼不同，光看檔案分不出來源（spec 1121）。
//
// 用法：
//
//	go run ./cmd/dos-save-export -out workplace/dos-oracle/game/SAVE \
//	  -slot A -area 2 -map-block 1 -x 7 -y 13 -facing 0
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/area"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	partySave "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/save"
)

func main() {
	out := flag.String("out", "", "輸出目錄（原版的存檔路徑，預設是 C:\\SAVE）")
	slot := flag.String("slot", "A", "存檔槽 A..J")
	name := flag.String("name", "ORACLE", "角色名（1..15 bytes）")
	gameArea := flag.Int("area", 2, "GameArea／章節（1..6）")
	mapBlock := flag.Int("map-block", 1, "Current3DMapBlockID：第一人稱地圖的 GEO 區塊")
	posX := flag.Int("x", 7, "地圖座標 X")
	posY := flag.Int("y", 13, "地圖座標 Y")
	facing := flag.Int("facing", 0, "朝向 0..7（0 為北）")
	wallType := flag.Int("wall-type", 0, "MapWallType")
	wallRoof := flag.Int("wall-roof", 0, "MapWallRoof")
	gameState := flag.Int("game-state", 0, "GameState")
	lastGameState := flag.Int("last-game-state", 0, "LastGameState")
	inDungeon := flag.Bool("in-dungeon", true, "InDungeon：第一人稱地圖為真")
	city := flag.Int("city", 0, "CurrentCity")
	charRef := flag.String("char-ref", "", "存檔裡記的角色檔名；留空用 CHRDAT<槽>1 並一併寫出該檔")
	pad := flag.Int("pad-ecl", 0, "在 ECL 區塊後補幾個位元組（原版量到的區塊比 spec 181 長 0x41）")
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

	base := fmt.Sprintf("CHRDAT%c1", key)
	ref := base
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
		MapWallType:   uint8(*wallType),
		MapWallRoof:   uint8(*wallRoof),
		LastGameState: uint8(*lastGameState),
		GameState:     uint8(*gameState),
		PartyCount:    1,
		CharacterRefs: [partySave.SAVGAMCharacterRefCount][]byte{characterRef},
	}
	prefix, err := partySave.EncodeSAVGAM(container)
	if err != nil {
		log.Fatal(err)
	}
	if *pad > 0 {
		// spec 1072 從 PC-98 逐條讀到的 ECL 區塊是 0x1E41，比 spec 181 從
		// decompiler 輸出記的 0x1E00 多 0x41；補在 ECL 區塊之後、地圖那五格之前。
		cut := 1 + partySave.SAVGAMArea1Size + partySave.SAVGAMArea2Size +
			partySave.SAVGAMRuntimeStateSize + partySave.SAVGAMECLMemorySize
		padded := make([]byte, 0, len(prefix)+*pad)
		padded = append(padded, prefix[:cut]...)
		padded = append(padded, make([]byte, *pad)...)
		padded = append(padded, prefix[cut:]...)
		prefix = padded
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

	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatal(err)
	}
	prefixName := fmt.Sprintf("savgam%c.dat", key+('a'-'A'))
	if err := os.WriteFile(filepath.Join(*out, prefixName), prefix, 0o644); err != nil {
		log.Fatal(err)
	}
	if *charRef == "" {
		if err := os.WriteFile(filepath.Join(*out, base+".sav"), record, 0o644); err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(*out, base+".GUY"), record, 0o644); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("%s（%d bytes）與 %s.sav（%d bytes）已寫到 %s\n",
		prefixName, len(prefix), base, len(record), *out)
	fmt.Printf("槽 %c：area=%d block=0x%02X 位置=(%d,%d) 朝向=%d in-dungeon=%v\n",
		key, *gameArea, *mapBlock, *posX, *posY, *facing, *inDungeon)
}
