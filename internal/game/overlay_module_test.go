package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// `overlay-04` 是**神殿**不是營地（spec 1182）。
//
// ★ 這條直接對遊戲檔斷言，因為那個標籤錯了很久而且**錯了不會有任何徵兆**：
// spec 1030 把它讀成營地主選單，spec 1095／1149 沿用，還因此說 remake 的
// `TempleRequested` 名字不對——反了。
//
// 判準用 overlay 自己的字串：神殿有牧師、治療、復活死者、鑑定；營地沒有。
func TestOverlayFourIsTheTempleNotTheCamp(t *testing.T) {
	path := filepath.Join("..", "..", "workplace", "re-sweep", "dos", "out", "overlay-04.json")
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("找不到 DOS 全掃輸出，跳過：%v", err)
	}
	var sweep struct {
		Strings []struct {
			Text string `json:"text"`
		} `json:"strings"`
	}
	if err := json.Unmarshal(payload, &sweep); err != nil {
		t.Fatal(err)
	}
	if len(sweep.Strings) == 0 {
		t.Fatal("overlay-04 一條字串都沒讀到——正對照失敗，這個測試等於沒跑")
	}
	joined := strings.ToLower(strings.Join(func() []string {
		out := make([]string, 0, len(sweep.Strings))
		for _, item := range sweep.Strings {
			out = append(out, item.Text)
		}
		return out
	}(), "\n"))

	for _, marker := range []string{"priest", "e dead", "appraise", "how can we help you"} {
		if !strings.Contains(joined, marker) {
			t.Errorf("overlay-04 的字串裡找不到 %q——神殿的判準不成立了，先確認資料再改結論", marker)
		}
	}
}
