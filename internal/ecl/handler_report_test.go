package ecl

import (
	"os"
	"regexp"
	"strconv"
	"testing"
)

// handlerReportRow 抓 `docs/audit/ecl-opcode-handlers-dos.md` 逐 opcode 那張表的
// 第一欄（opcode）與最後一欄（副作用狀態）。
var handlerReportRow = regexp.MustCompile(
	"(?m)^\\| `([0-9A-F]{2})h` \\|.*\\| `(done|partial|consumed|unknown)` \\|$")

// ★ 這條擋的是**排順序用的那張表會不會過期**。收尾一支 opcode 改的是
// `OpcodeEffects`，而 `ecl-opcode-handlers-dos.md` 的狀態欄是手寫的——沒有東西
// 會在收尾那一刻更新它。第 751 輪抽查時六支已經是 `done` 的還記著 `partial`
// （`1Ch`／`24h`／`27h`／`2Eh`／`37h`），摘要因此把「還剩多少」多報了四倍。
//
// ⚠ 只比對兩邊都有的 opcode：那張表列的是**分派器列出的**，主迴圈在分派前就
// 處理掉的比較指令不在裡面，而 `OpcodeEffects` 兩種都收。
func TestHandlerReportStatusMatchesOpcodeEffects(t *testing.T) {
	raw, err := os.ReadFile("../../docs/audit/ecl-opcode-handlers-dos.md")
	if err != nil {
		t.Skipf("讀不到 handler 對照表：%v", err)
	}
	matches := handlerReportRow.FindAllStringSubmatch(string(raw), -1)
	if len(matches) < 40 {
		t.Fatalf("只抓到 %d 列，表的格式可能改了", len(matches))
	}
	checked := 0
	for _, match := range matches {
		value, parseErr := strconv.ParseUint(match[1], 16, 8)
		if parseErr != nil {
			t.Fatalf("opcode %q 解不出來：%v", match[1], parseErr)
		}
		effect, ok := OpcodeEffects[uint8(value)]
		if !ok {
			continue
		}
		checked++
		if string(effect.Status) != match[2] {
			t.Errorf("opcode %sh：報表寫 %q，`OpcodeEffects` 是 %q",
				match[1], match[2], effect.Status)
		}
	}
	if checked < 40 {
		t.Fatalf("只對到 %d 個 opcode，兩邊的鍵可能對不上", checked)
	}
}
