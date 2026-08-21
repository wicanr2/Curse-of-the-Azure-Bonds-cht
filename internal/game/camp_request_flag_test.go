package game

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// 原作 `24h COMBAT` 的三選一，第二格是 `bank1^[5C4h]`＝**營地**（spec 1095／1096），
// remake 把同一格（`0x7EE2`）標成 Temple。位址對、語意不符。
//
// ★ 目前**不會**產生錯誤行為，因為那條路走不到：正式程式碼裡沒有任何一處寫
// `0x7EE2`（ECL 那一側也從沒寫過，spec 1095 已全掃 1,355 條指令確認）。
// 也就是說 `TempleRequested` 是一條**沒有 producer 的死路**。
//
// 這一條擋的是「哪天有人補上 producer，卻沒先決定那一格到底是營地還是神殿」。
// ⚠ 真要補 producer，先讀 spec 1095 §「目前 remake 的落差」：原作那一支呼叫的是
// `overlay-04 entry#1`（營地主選單），而 remake 的營地走的是另一條等價路徑
// （`EnterDungeonCamp` → lifecycle entry 2，跑完 ECL 再開營地選單）。
func TestCampRequestFlagHasNoProducer(t *testing.T) {
	root := filepath.Join("..", "..")
	// 已知且允許的兩處：常數宣告，以及 `24h` 讀到就清零的那一段。
	allowed := map[string]bool{
		filepath.Join("internal", "ecl", "runtime.go"): true,
	}
	var offenders []string
	err := filepath.Walk(filepath.Join(root, "internal"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		relative, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		if allowed[relative] {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		for index, line := range strings.Split(string(data), "\n") {
			if !strings.Contains(line, "0x7EE2") {
				continue
			}
			offenders = append(offenders,
				relative+":"+strconv.Itoa(index+1)+"  "+strings.TrimSpace(line))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(offenders) > 0 {
		t.Errorf("有人開始碰 `0x7EE2` 了，但那一格是營地還是神殿還沒決定：\n  %s",
			strings.Join(offenders, "\n  "))
	}
}

