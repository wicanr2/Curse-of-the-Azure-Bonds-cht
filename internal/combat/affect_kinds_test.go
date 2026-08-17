package combat

import (
	"os"
	"regexp"
	"sort"
	"testing"
)

// 這一條把 `InterpretedAffectKinds` 綁回原始碼：`battle.go` 裡每一個拿效果碼
// 去比對的常數都要登記，反過來登記了卻找不到實作也要紅。
//
// ★ 為什麼不是人工維護一張表。 人工表的失效方式是安靜的——新增一個
// `case 0x2B:` 而忘了登記，覆蓋報告從此少算一個已判讀的碼，而沒有任何測試
// 會發現。這條測試把「忘了更新」變成編譯後就紅的失敗。
func TestInterpretedAffectKindsMatchTheSource(t *testing.T) {
	source, err := os.ReadFile("battle.go")
	if err != nil {
		t.Fatal(err)
	}
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?:affect|a)\.Kind\s*==\s*0x([0-9A-Fa-f]{2})`),
		regexp.MustCompile(`(?m)^\s*case 0x([0-9A-Fa-f]{2}):`),
	}
	found := map[string]bool{}
	for _, pattern := range patterns {
		for _, match := range pattern.FindAllStringSubmatch(string(source), -1) {
			found[normaliseHex(match[1])] = true
		}
	}
	if len(found) == 0 {
		t.Fatal("原始碼裡一個效果碼常數都沒掃到——正規式壞了，這條測試等於沒跑")
	}
	registered := map[string]bool{}
	for _, kind := range InterpretedAffectKinds {
		registered[normaliseHex(hexByte(kind))] = true
	}
	var missing, extra []string
	for key := range found {
		if !registered[key] {
			missing = append(missing, key)
		}
	}
	for key := range registered {
		if !found[key] {
			extra = append(extra, key)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	if len(missing) > 0 {
		t.Errorf("battle.go 判讀了這些效果碼卻沒登記：%v", missing)
	}
	if len(extra) > 0 {
		t.Errorf("登記了但 battle.go 沒有對應的判讀：%v", extra)
	}
}

func normaliseHex(value string) string {
	const digits = "0123456789ABCDEF"
	out := []byte(value)
	for index, char := range out {
		if char >= 'a' && char <= 'f' {
			out[index] = digits[10+int(char-'a')]
		}
	}
	return string(out)
}

func hexByte(value uint8) string {
	const digits = "0123456789ABCDEF"
	return string([]byte{digits[value>>4], digits[value&0x0F]})
}
