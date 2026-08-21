package game

import (
	"math/rand"
	"testing"
)

// `27h TREASURE` 的 `ItemBlock > 80h` 那一路是隨機產生 `n − 80h` 件戰利品。
// 原作的表在 DOS `overlay-02:1D02h`..`1E6Bh`（spec 1151），每一段都是**兩端
// 都夾**的區間：
//
//	第一擲 1d100
//	  1..60  → 第二擲 1d100
//	           1..47、50..59  → 就是那個號碼（45 例外，改 59）
//	           60..90         → 再擲 1d10 → 36／35／34／37／38
//	           91..94         → 73
//	           95..97         → 93
//	           98..100        → 77
//	           其餘（48、49）  → 59
//	  61..85 → 61
//	  86..92 → 62
//	  93..98 → 再擲 1d15 → 71／84／79
//	  99,100 → 59
//
// ⚠ 兩個方向都要驗：只驗「48 會回 59」的話，把 `[60,90]` 那一段整個刪掉同樣
// 會過，而那會讓 31 個點全部掉到預設值、劍再也開不出來。
func TestRandomTreasureItemTypeFollowsTheOriginalRanges(t *testing.T) {
	// 把整張表逐點展開：用一個把 Intn 換成固定序列的 rng 不方便，
	// 改用「窮舉 seed 直到擲出想要的組合」太慢——直接呼叫內部分支函式。
	cases := []struct {
		name  string
		rolls []int
		want  uint8
	}{
		{"第二擲 1 → 就是那個號碼", []int{10, 1}, 1},
		{"第二擲 47 → 就是那個號碼", []int{10, 47}, 47},
		{"第二擲 45 → 改成 59", []int{10, 45}, 59},
		{"第二擲 48 → 落到預設值 59", []int{10, 48}, 59},
		{"第二擲 49 → 落到預設值 59", []int{10, 49}, 59},
		{"第二擲 50 → 就是那個號碼", []int{10, 50}, 50},
		{"第二擲 59 → 就是那個號碼", []int{10, 59}, 59},
		{"第二擲 60 → 進劍那一段", []int{10, 60, 1}, 36},
		{"第二擲 90 → 進劍那一段", []int{10, 90, 10}, 38},
		{"劍 1d10 = 4", []int{10, 70, 4}, 36},
		{"劍 1d10 = 5", []int{10, 70, 5}, 35},
		{"劍 1d10 = 8", []int{10, 70, 8}, 34},
		{"劍 1d10 = 9", []int{10, 70, 9}, 37},
		{"第二擲 91 → 73", []int{10, 91}, 73},
		{"第二擲 94 → 73", []int{10, 94}, 73},
		{"第二擲 95 → 93", []int{10, 95}, 93},
		{"第二擲 97 → 93", []int{10, 97}, 93},
		{"第二擲 98 → 77", []int{10, 98}, 77},
		{"第二擲 100 → 77", []int{10, 100}, 77},
		{"第一擲 61 → 61", []int{61}, 61},
		{"第一擲 85 → 61", []int{85}, 61},
		{"第一擲 86 → 62", []int{86}, 62},
		{"第一擲 92 → 62", []int{92}, 62},
		{"第一擲 93 → 藥水那一段", []int{93, 1}, 71},
		{"藥水 1d15 = 9", []int{95, 9}, 71},
		{"藥水 1d15 = 10", []int{95, 10}, 84},
		{"藥水 1d15 = 11", []int{95, 11}, 79},
		{"第一擲 98 → 藥水那一段", []int{98, 15}, 79},
		{"第一擲 99 → 59", []int{99}, 59},
		{"第一擲 100 → 59", []int{100}, 59},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			rng := rand.New(scriptedSource(test.rolls))
			if got := randomTreasureItemType(rng); got != test.want {
				t.Fatalf("擲 %v 得到物品 %d，應該是 %d", test.rolls, got, test.want)
			}
		})
	}
}

// scriptedSource 讓 `rng.Intn(n)` 依序回傳指定的 1..n 值。
//
// `Intn(n)`（n < 2^31）走的是 `Int31n`，而 `Int31()` ＝ `Int63() >> 32`，
// 所以要讓 `Intn` 回傳 k 就得把 k 左移 32 位再交出去；直接回傳小整數會被
// 移位吃掉、每次都得到 0。`Int31n` 之後做的是 `v % n`，v 小於 n 時就是 v 本身。
type scripted struct {
	values []int
	index  int
}

func scriptedSource(values []int) rand.Source {
	return &scripted{values: values}
}

func (s *scripted) Int63() int64 {
	if s.index >= len(s.values) {
		panic("scripted rng 用完了：測試給的擲骰數不夠")
	}
	value := s.values[s.index]
	s.index++
	return int64(value-1) << 32
}

func (s *scripted) Seed(int64) {}
