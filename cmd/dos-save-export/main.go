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
//
//	      ⚠ 只換得到**同一張圖裡的任一格**。`-map-block` 換不到別張圖：
//	      原版的第一人稱地圖由存檔裡的 ECL 狀態決定，不是由那個位元組決定
//	      （第 675 輪實測，見下方註解與 spec 1185）。
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
	"archive/zip"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"io"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/area"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	partySave "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/save"
)

// 原版存檔的固定前綴版面（spec 181／1134，DOS 側 ECL 區塊是 1E00h）。
const (
	mapPosOffset = 1 + partySave.SAVGAMArea1Size + partySave.SAVGAMArea2Size +
		partySave.SAVGAMRuntimeStateSize + partySave.SAVGAMECLMemorySize
	area1MapBlockOffset = 1 + 0x18A
	area1LastECLOffset  = 1 + 0x1E4
	// ⚠ `GameArea` 在存檔裡有**兩份**：容器的第一個位元組，以及 Area2 的
	// `0x624`。載入端讀的是 Area2 那一份（`area1.GameArea = area2.GameArea`），
	// 只改容器那一份的話章節看起來變了、實際沒變——而且不會報錯。
	area2GameAreaOffset = 1 + partySave.SAVGAMArea1Size + 0x624
	// 三組牆面參數接在 map state 五個位元組與兩個 state 位元組之後。
	setBlocksOffset = mapPosOffset + partySave.SAVGAMMapStateSize + partySave.SAVGAMStateBytes
	// ECL 程式碼視窗在存檔的第四塊，接在前三塊之後（spec 1163）。
	eclWindowOffset = 1 + partySave.SAVGAMArea1Size + partySave.SAVGAMArea2Size +
		partySave.SAVGAMRuntimeStateSize
	area1LastXOffset    = 1 + 0x1E0
	area1LastYOffset    = 1 + 0x1E2
)

// setFlags 回報使用者實際打了哪些旗標。`-base` 模式只覆寫打過的欄位，
// 沒打的一律沿用底檔——底檔是原版自己寫出來的，猜它的值只會弄壞它。
func setFlags() map[string]bool {
	seen := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { seen[f.Name] = true })
	return seen
}

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
	eclBlock := flag.Int("ecl-block", -1, "把這一段 ECL 的位元組碼換進存檔的程式碼視窗（換圖用；-1 不動）")
	image := flag.String("image", "curseoftheazurebonds.zip", "原版 image ZIP（-ecl-block 要用）")
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
		given := setFlags()
		if given["area"] {
			patched[0] = byte(*gameArea)
			patched[area2GameAreaOffset] = byte(*gameArea)
		}
		if given["map-block"] {
			patched[area1MapBlockOffset] = byte(*mapBlock)
			// ⚠ 這個位元組**改不動原版的第一人稱地圖**。實測（第 675 輪）：
			// 拿提爾佛頓的存檔寫 `-map-block 3`（下水道）再讓原版讀進去，
			// 站上 (1,8) 拍到的畫面與 remake 畫的**提爾佛頓** (1,8) 逐格相同
			// （88×88 格差 0 格），與下水道那張差 48%。存檔裡也**沒有**任何一塊
			// GEO 區塊的位元組（整份 13,149 bytes 掃過六個 GEO 檔集都沒有），
			// 所以地圖幾何不是跟著存檔走的——原版是照存檔裡的 ECL 狀態重跑
			// `LOAD FILES` 決定載哪一張。要換圖得換 ECL 狀態，不是換這個位元組。
			//
			// remake 這一側反而**認**這個位元組（`Current3DMapBlockID`），
			// 所以兩邊會指向不同的圖而且都不會報錯——這正是最危險的形狀。
			fmt.Fprintln(os.Stderr, "⚠ -map-block 只改存檔裡的標記；原版的第一人稱地圖由 ECL 狀態決定。"+
				"要換圖請加 -ecl-block（同一個 GEO 檔集內有效；跨檔集實測無效，見 spec 1185）。")
		}
		// Area1 的 LastX／LastY 與地圖那五格都要改：前者是「上一個座標」，
		// 後者才是載入之後站的位置，只改一邊會在畫面與地圖標記之間打架。
		patched[area1LastXOffset] = byte(*posX)
		patched[area1LastXOffset+1] = 0
		patched[area1LastYOffset] = byte(*posY)
		patched[area1LastYOffset+1] = 0
		patched[mapPosOffset] = byte(*posX)
		patched[mapPosOffset+1] = byte(*posY)
		patched[mapPosOffset+2] = byte(*facing)
		// 牆型與地形沒指定時，**照目標地圖算**，不要沿用底檔。
		//
		// ⚠ 沿用底檔的後果只在天空色上看得出來：地形位元組 ≥ 80h 是室內（黑天），
		// 否則是室外。拿提爾佛頓的 `80h` 去配一張室外的圖，原版畫黑天、remake 照
		// 自己算出來的地形畫亮天——2,224 個像素的差異，而兩邊各自看起來都正常
		// （第 683 輪在 GEO3 段 0x15 上踩到）。
		derivedType, derivedRoof, derivedOK := mapCellState(*image, *gameArea, *mapBlock, *posX, *posY, *facing)
		switch {
		case *wallType >= 0:
			patched[mapPosOffset+3] = byte(*wallType)
		case derivedOK:
			patched[mapPosOffset+3] = derivedType
		}
		switch {
		case *wallRoof >= 0:
			patched[mapPosOffset+4] = byte(*wallRoof)
		case derivedOK:
			patched[mapPosOffset+4] = derivedRoof
		}
		// -ecl-block 換圖。 原版的第一人稱地圖不是存檔裡某個編號決定的，是**當前
		// ECL 段的進入碼**跑 `LOAD FILES` 決定的（第 675 輪實測）。而存檔的第四塊
		// 就是那一段的位元組碼本身：實測 slot B 的視窗正是 `ECL2` 段 `0x01` 的
		// 資料**從位移 2 開始**（前兩個 byte 是段頭，不進記憶體）。所以換圖＝把
		// 另一段的位元組碼換進這個視窗，並把段編號一起改掉。
		if *eclBlock >= 0 {
			code, pieces, piecesOK, err := eclBlockBytes(*image, *gameArea, uint8(*eclBlock))
			if err != nil {
				log.Fatal(err)
			}
			window := patched[eclWindowOffset : eclWindowOffset+partySave.SAVGAMECLMemorySize]
			for i := range window {
				window[i] = 0
			}
			copy(window, code)
			binary.LittleEndian.PutUint16(patched[area1LastECLOffset:], uint16(*eclBlock))
			// ⚠ 牆面參數也要一起換。 存檔記著三組「用哪一塊牆磚」，載入時原版會
			// 拿它們去 `WALLDEF<章>.DAX` 重載。只換地圖不換這個，原版會拿**上一張
			// 圖的**選圖去新章的檔案裡找，然後印
			// `Unable to load 1 from WALLDEF4.` 收場（第 679 輪實測）。
			// 值就取自我們正要裝進去的那一段自己發的 `37h LOAD PIECES`。
			// ⚠ **只有換章的時候才改牆面參數。** 這一段是為了跨章而加的：底檔的
			// 選圖（提爾佛頓的 1,2,3）在 `WALLDEF4` 裡不存在，載入會停在
			// `Unable to load 1 from WALLDEF4.`。但**同一章之內不能改**——實測
			// 同一格 (1,8) E 的原版畫面，改了之後與沒改差 3,735 格，而沒改的那張
			// 與 remake 逐格相同。同章之內段自己的 `LOAD PIECES` 會把選圖設對，
			// 先寫進去反而讓原版走到不一樣的狀態。
			if piecesOK && *gameArea != int(prefix[0]) {
				for index, piece := range pieces {
					value := piece
					if piece == 0xFF {
						value = 0xFFFF
					}
					// ⚠ 第二個欄位是**槽號**（1..3），不是選圖編號。底檔分不出這兩種
					// 解讀——提爾佛頓的選圖剛好是 1,2,3，與槽號同號。寫成選圖編號的
					// 話原版會**一面牆都不畫**，而畫面看起來仍然正常（天空與地板都在）。
					binary.LittleEndian.PutUint16(patched[setBlocksOffset+index*4:], value)
					binary.LittleEndian.PutUint16(patched[setBlocksOffset+index*4+2:], uint16(index+1))
				}
				fmt.Fprintf(os.Stderr, "牆面參數改成 %d,%d,%d（取自該段自己的 LOAD PIECES）\n",
					pieces[0], pieces[1], pieces[2])
			} else if !piecesOK {
				fmt.Fprintln(os.Stderr, "⚠ 這一段的 LOAD PIECES 不是常數，牆面參數維持底檔的值")
			}
			fmt.Fprintf(os.Stderr, "換入 ECL%d 段 0x%02X 的位元組碼 %d bytes（視窗 %d bytes）\n",
				*gameArea, *eclBlock, len(code), partySave.SAVGAMECLMemorySize)
		}
		if err := os.WriteFile(filepath.Join(*out, prefixName), patched, 0o644); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s（%d bytes，以 %s 為底）area=%d block=0x%02X 位置=(%d,%d) 朝向=%d wallType=0x%02X wallRoof=0x%02X\n",
			prefixName, len(patched), *base, patched[0], patched[area1MapBlockOffset],
			*posX, *posY, *facing, patched[mapPosOffset+3], patched[mapPosOffset+4])
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

// eclBlockBytes 取出一段 ECL 要放進存檔程式碼視窗的位元組。
//
// ⚠ 前兩個 byte 是段頭，不進記憶體：實測 slot B 的視窗與 `ECL2` 段 `0x01` 的
// 資料**從位移 2 起**逐 byte 相同。少切這兩個 byte，整段碼會整體偏移兩格，
// 而 ECL 位元組碼沒有對齊檢查——它會照樣執行，只是執行的是別的東西。
func eclBlockBytes(imagePath string, area int, block uint8) ([]byte, [3]uint16, bool, error) {
	archive, err := zip.OpenReader(imagePath)
	if err != nil {
		return nil, [3]uint16{}, false, err
	}
	defer archive.Close()
	name := fmt.Sprintf("ECL%d.DAX", area)
	var payload []byte
	for _, file := range archive.File {
		if !strings.EqualFold(file.Name, name) {
			continue
		}
		handle, openErr := file.Open()
		if openErr != nil {
			return nil, [3]uint16{}, false, openErr
		}
		payload, err = io.ReadAll(handle)
		handle.Close()
		if err != nil {
			return nil, [3]uint16{}, false, err
		}
	}
	if payload == nil {
		return nil, [3]uint16{}, false, fmt.Errorf("image 裡沒有 %s", name)
	}
	blocks, err := dax.Parse(payload)
	if err != nil {
		return nil, [3]uint16{}, false, err
	}
	for _, item := range blocks {
		if item.Entry.ID != block {
			continue
		}
		if len(item.Data) < 2 {
			return nil, [3]uint16{}, false, fmt.Errorf("%s 段 0x%02X 只有 %d bytes", name, block, len(item.Data))
		}
		code := item.Data[2:]
		if len(code) > partySave.SAVGAMECLMemorySize {
			return nil, [3]uint16{}, false, fmt.Errorf("%s 段 0x%02X 的碼 %d bytes 放不進 %d bytes 的視窗",
				name, block, len(code), partySave.SAVGAMECLMemorySize)
		}
		pieces, ok := ecl.BlockWallPieces(item.Data)
		return code, pieces, ok, nil
	}
	return nil, [3]uint16{}, false, fmt.Errorf("%s 裡沒有段 0x%02X", name, block)
}

// mapCellState 從目標地圖算出這一格的牆型與地形位元組。
func mapCellState(imagePath string, area, block, x, y, facing int) (wall, roof uint8, ok bool) {
	archive, err := zip.OpenReader(imagePath)
	if err != nil {
		return 0, 0, false
	}
	defer archive.Close()
	name := fmt.Sprintf("GEO%d.DAX", area)
	catalog := geo.NewCatalog()
	for _, file := range archive.File {
		if !strings.EqualFold(file.Name, name) {
			continue
		}
		handle, openErr := file.Open()
		if openErr != nil {
			return 0, 0, false
		}
		payload, readErr := io.ReadAll(handle)
		handle.Close()
		if readErr != nil {
			return 0, 0, false
		}
		if addErr := catalog.AddDAX(uint8(area), payload); addErr != nil {
			return 0, 0, false
		}
	}
	grid, found := catalog.Lookup(geo.MapRef{Set: uint8(area), BlockID: uint8(block)})
	if !found {
		return 0, 0, false
	}
	wallValue, wallOK := grid.WallWrapped(x, y, facing)
	if !wallOK {
		return 0, 0, false
	}
	return wallValue, grid.CellWrapped(x, y).Terrain, true
}
