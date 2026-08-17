package game

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// 原版位元組的解碼邊界要靠測試守，因為走錯路在**英文資料上完全沒有症狀**
// （spec 1121）。任何讀 `*.DAX` 區塊的地方都必須走 `ParseOriginalItems`。
//
// ★ 為什麼是掃原始碼而不是跑一次。 走錯路的載入點不會出錯、不會 panic、
// 也不會被任何既有測試碰到——它只是把中文名字讀成亂碼。唯一能可靠發現
// 「又多了一個沒分流的載入點」的辦法，是每次都把所有載入點列出來比對。
func TestEveryDAXItemLoadUsesTheOriginalDecoder(t *testing.T) {
	roots := []string{"../../internal", "../../cmd"}
	pattern := regexp.MustCompile(`monster\.ParseItems\(([^)]*)\)`)
	var offenders []string
	for _, root := range roots {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") ||
				strings.HasSuffix(path, "_test.go") {
				return err
			}
			source, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			for _, match := range pattern.FindAllStringSubmatch(string(source), -1) {
				argument := strings.TrimSpace(match[1])
				// `block.Data` 一定來自 DAX，也就是原版位元組。
				if strings.Contains(argument, "block.") || strings.Contains(argument, "Block") {
					offenders = append(offenders, path+": ParseItems("+argument+")")
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(offenders) > 0 {
		t.Fatalf("DAX 區塊走了 remake 的解碼路徑，中文物品名會變亂碼：\n  %s\n"+
			"改用 monster.ParseOriginalItems（spec 1121）。",
			strings.Join(offenders, "\n  "))
	}
	// 正對照：確認掃描真的看得到原始碼，否則這條測試永遠綠。
	source, err := os.ReadFile("treasure.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(source), "monster.ParseOriginalItems(block.Data)") {
		t.Fatal("treasure.go 沒有掃到預期的呼叫——這條測試的掃描範圍壞了")
	}
}
