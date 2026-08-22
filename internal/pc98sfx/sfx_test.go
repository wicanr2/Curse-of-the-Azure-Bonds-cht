package pc98sfx

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestImportRejectsUnknownExecutable(t *testing.T) {
	t.Parallel()

	_, err := Import([]byte("not GAME.EXE"))
	if err == nil || !strings.Contains(err.Error(), "SHA-256") {
		t.Fatalf("Import error=%v", err)
	}
}

func TestDecodeImmediateFormulaAndTablePaths(t *testing.T) {
	t.Parallel()

	game := make([]byte, tableOffset+selectorCount*wordsPerEffect*2+2)
	put := func(selector, index int, value uint16) {
		offset := tableOffset + (selector*wordsPerEffect+index)*2
		game[offset] = byte(value)
		game[offset+1] = byte(value >> 8)
	}
	put(3, 1, 1000)
	put(3, 2, 3000)
	put(3, 3, 500)
	put(3, 4, 0)

	noOp := decodeEffect(game, 13)
	if !noOp.NoOp || len(noOp.Steps) != 0 {
		t.Fatalf("selector 13=%+v", noOp)
	}
	formula := decodeEffect(game, 2)
	if formula.Source != "formula" || len(formula.Steps) != 1 ||
		formula.Steps[0].FrequencyOrPeriod != 20 ||
		formula.Steps[0].PulseCount != 125 {
		t.Fatalf("selector 2=%+v", formula)
	}
	table := decodeEffect(game, 3)
	if table.Source != "table" || len(table.Steps) != 3 ||
		table.Steps[0].PulseCount != 2 ||
		table.Steps[1].Kind != StepDelay ||
		table.Steps[2].PulseCount != 4 {
		t.Fatalf("selector 3=%+v", table)
	}
}

func TestKnownDigestConstantIsCanonicalHex(t *testing.T) {
	t.Parallel()

	decoded, err := hex.DecodeString(GameSHA256)
	if err != nil || len(decoded) != sha256.Size {
		t.Fatalf("GameSHA256=%q err=%v", GameSHA256, err)
	}
}

func TestSelectorForEventUsesBorlandSemanticNames(t *testing.T) {
	t.Parallel()

	tests := map[string]int{
		"cast": 2, "miss": 3, "spell_hit": 4, "dead": 5,
		"whistle": 6, "hit": 7, "lightning": 8, "swish": 9,
		"step": 10, "fireball": 11, "arrow": 12,
		"overture": 13, "combat": 14, "crash": 15, "stop": 255,
	}
	for event, want := range tests {
		got, ok := SelectorForEvent(event)
		if !ok || got != want {
			t.Errorf("SelectorForEvent(%q)=(%d,%v), want (%d,true)", event, got, ok, want)
		}
	}
	if _, ok := SelectorForEvent("unknown"); ok {
		t.Fatal("unknown event was accepted")
	}
}

// 描述子位址曾經是一張手打的對照表，而那張表漏了 4840h／4842h 兩格
// （MISSFX／SPELLHITFX）。漏掉沒有任何徵兆：報表只是把那兩處印成
// 「符號表沒有」。這一組把「不可能漏」釘住。
func TestSelectorsCoverEveryDescriptorContiguously(t *testing.T) {
	list := Selectors()
	if len(list) != 17 {
		t.Fatalf("選擇子 ＋ SOUNDHALT 共 %d 個，want 17", len(list))
	}
	if list[0].Symbol != "SOUNDHALT" || list[0].Descriptor != HaltDescriptor {
		t.Fatalf("第一個應該是 SOUNDHALT@%04Xh：%+v", HaltDescriptor, list[0])
	}
	seen := map[int]string{}
	for _, info := range list[1:] {
		if info.Descriptor != DescriptorBase+info.Selector*2 {
			t.Fatalf("選擇子 %d 的描述子 %04Xh 不等於基底 ＋ 選擇子×2", info.Selector, info.Descriptor)
		}
		if previous, clash := seen[info.Descriptor]; clash {
			t.Fatalf("描述子 %04Xh 同時屬於 %s 與 %s", info.Descriptor, previous, info.Symbol)
		}
		seen[info.Descriptor] = info.Symbol
	}
	// 那兩格必須查得到，而且必須是這兩個名字——這正是舊表漏掉的兩格。
	for address, want := range map[int]string{0x4840: "MISSFX", 0x4842: "SPELLHITFX"} {
		info, ok := SelectorForDescriptor(address)
		if !ok || info.Symbol != want {
			t.Fatalf("%04Xh ＝ %+v，want %s", address, info, want)
		}
	}
	// 負對照：表外的位址必須查不到，否則上面的「查得到」不算數。
	if info, ok := SelectorForDescriptor(0x4836); ok {
		t.Fatalf("4836h 不在表裡卻查得到：%+v", info)
	}
	if info, ok := SelectorForDescriptor(0x485A); ok {
		t.Fatalf("485Ah 不在表裡卻查得到：%+v", info)
	}
}

// 每個有玩法語意的選擇子都要有事件名，而且事件名不能撞號——
// `cmd/sound-trigger-compare` 是拿事件名當鍵去對 remake 的。
func TestSelectorEventsAreUniqueWhenPresent(t *testing.T) {
	events := map[string]string{}
	for _, info := range Selectors() {
		if info.Event == "" {
			// SOUNDOFF／SOUNDON 是驅動開關，沒有玩法事件。
			if info.Symbol != "SOUNDOFF" && info.Symbol != "SOUNDON" {
				t.Fatalf("%s 沒有事件名", info.Symbol)
			}
			continue
		}
		if previous, clash := events[info.Event]; clash {
			t.Fatalf("事件 %q 同時屬於 %s 與 %s", info.Event, previous, info.Symbol)
		}
		events[info.Event] = info.Symbol
		if _, ok := SelectorForEvent(info.Event); !ok {
			t.Fatalf("%s 的事件 %q 反查不回選擇子", info.Symbol, info.Event)
		}
	}
	if len(events) != 15 {
		t.Fatalf("有事件名的選擇子 %d 個，want 15", len(events))
	}
}

// ★ 不作聲的選擇子**剛好這五個**，而且原因不只一處。
//
// ⚠ 這條擋的是一次真實的誤讀：`SOUNDFX` 的第一段早退最後一項是 `0FFh`
// （`SOUNDHALT`，根本不是選擇子），很容易讀成 `0Fh` ＝ 15 ＝ `CRASHFX`，
// 於是把 15 從清單拿掉「修好」它。**15 的早退在第二處**（`18995h`），
// 而表格在第 13 格之後就沒有資料了——拿掉之後 `CRASHFX` 只會讀到別人的位元組。
func TestSilentSelectorsAreExactlyTheFiveTheOriginalReturnsEarlyOn(t *testing.T) {
	t.Parallel()

	game := make([]byte, tableOffset+selectorCount*wordsPerEffect*2+2)
	// 每個選擇子都在表格裡放一個音，才分得出「原本就沒資料」與「被判成無聲」。
	for selector := 0; selector < selectorCount; selector++ {
		offset := tableOffset + (selector*wordsPerEffect+1)*2
		game[offset], game[offset+1] = 0xE8, 0x03 // 1000
	}
	silent := map[int]bool{}
	for selector := 0; selector < selectorCount; selector++ {
		if decodeEffect(game, selector).NoOp {
			silent[selector] = true
		}
	}
	want := map[int]bool{0: true, 1: true, 13: true, 14: true, 15: true}
	if len(silent) != len(want) {
		t.Fatalf("不作聲的選擇子 ＝ %v，want %v", silent, want)
	}
	for selector := range want {
		if !silent[selector] {
			t.Errorf("選擇子 %d（%s）應該是無聲的", selector, selectorMetadata[selector].symbol)
		}
	}
	// `SOUNDHALT` 不在 0..15 裡：它是「停手上那一個」，不是一個音效。
	if _, ok := SelectorForEvent("stop"); !ok {
		t.Fatal("`stop` 應該對到 SOUNDHALT")
	}
	if selector, _ := SelectorForEvent("stop"); selector < selectorCount {
		t.Fatalf("SOUNDHALT ＝ %d，不該落在選擇子範圍內", selector)
	}
}
