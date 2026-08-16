package ecl

import "testing"

// `3Dh CLEAR BOX` 的可觀察效果是「文字框空了」，而不是任何記憶體變化。
// 原作用它在換畫面之前擦掉上一幕（`ECL1.DAX/0x52 +001Bh`：
// CLEAR BOX → PICTURE → ADD NPC → PRINTCLEAR）。
func TestClearBoxIsReportedSeparatelyFromText(t *testing.T) {
	// CLEAR BOX 然後 EXIT：這一次執行沒有任何文字。
	block := []byte{0x00, 0x00, 0x3D, 0x00}
	result, err := RunSubset(block, 0, 16)
	if err != nil {
		t.Fatal(err)
	}
	if !result.ClearBoxRequested {
		t.Fatal("CLEAR BOX 沒有回報出來，上層無從得知要清空文字框")
	}
	if len(result.Text) != 0 {
		t.Fatalf("CLEAR BOX 不該產生文字，卻得到 %v", result.Text)
	}
}

func TestClearBoxIsNotSetWhenTheBlockNeverClears(t *testing.T) {
	// 只有 EXIT。
	result, err := RunSubset([]byte{0x00, 0x00, 0x00}, 0, 16)
	if err != nil {
		t.Fatal(err)
	}
	if result.ClearBoxRequested {
		t.Fatal("沒有 CLEAR BOX 卻回報要清空")
	}
}
