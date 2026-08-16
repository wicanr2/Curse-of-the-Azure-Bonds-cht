package ecl

import "testing"

// `31h SPRITE OFF` 的可觀察效果是「畫面上那隻怪物不見了」。原作在
// `ECL2.DAX/0x02 +0CCBh` 用它接畫面提交點再開新頁：逃走之後怪物不該還在。
func TestSpriteOffIsReported(t *testing.T) {
	result, err := RunSubset([]byte{0x00, 0x00, 0x31, 0x00}, 0, 16)
	if err != nil {
		t.Fatal(err)
	}
	if !result.SpriteOffRequested {
		t.Fatal("SPRITE OFF 沒有回報出來，上層無從得知要關掉圖示")
	}
}

func TestSpriteOffIsNotSetWithoutTheOpcode(t *testing.T) {
	result, err := RunSubset([]byte{0x00, 0x00, 0x00}, 0, 16)
	if err != nil {
		t.Fatal(err)
	}
	if result.SpriteOffRequested {
		t.Fatal("沒有 SPRITE OFF 卻回報要關圖示")
	}
}
