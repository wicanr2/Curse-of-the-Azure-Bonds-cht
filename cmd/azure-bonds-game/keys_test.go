package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/game"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
)

// scriptedKeys 是測試用的鍵盤來源：一次餵一幀該按的鍵。
//
// ⚠ `JustPressed` 的語意是「**這一幀**剛按下」，所以同一幀問第二次也要回 true，
// 但下一幀就必須是 false。前端同一幀會對同一顆鍵問好幾次（不同分支各問各的），
// 讀完就清掉會讓後面的分支看不到那一顆。
type scriptedKeys struct {
	just map[ebiten.Key]bool
	held map[ebiten.Key]bool
}

func newScriptedKeys() *scriptedKeys {
	return &scriptedKeys{just: map[ebiten.Key]bool{}, held: map[ebiten.Key]bool{}}
}

func (s *scriptedKeys) JustPressed(key ebiten.Key) bool { return s.just[key] }
func (s *scriptedKeys) Pressed(key ebiten.Key) bool     { return s.held[key] || s.just[key] }

// press 安排下一幀按下這些鍵。
func (s *scriptedKeys) press(keys ...ebiten.Key) {
	s.just = map[ebiten.Key]bool{}
	for _, key := range keys {
		s.just[key] = true
	}
}

// release 清掉這一幀的按鍵，讓下一幀是空的。
func (s *scriptedKeys) release() { s.just = map[ebiten.Key]bool{} }

// ★ 這個測試是接縫的守門員。前端只要有人再直接呼叫 `inpututil.IsKeyJustPressed`
// 或 `ebiten.IsKeyPressed`，按鍵驅動的測試就**永遠走不到那一行**——而測試會照樣
// 綠，因為它根本到不了那裡。沉默的漏洞比紅燈貴，所以這裡把它變成紅燈。
func TestFrontendReadsKeysOnlyThroughTheSeam(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	fileSet := token.NewFileSet()
	offenders := []string{}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || name == "keys.go" || strings.HasSuffix(name, "_test.go") {
			continue
		}
		parsed, parseErr := parser.ParseFile(fileSet, filepath.Join(".", name), nil, 0)
		if parseErr != nil {
			t.Fatalf("%s: %v", name, parseErr)
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, ok := selector.X.(*ast.Ident)
			if !ok {
				return true
			}
			switch {
			case pkg.Name == "inpututil" && strings.HasPrefix(selector.Sel.Name, "IsKey"),
				pkg.Name == "ebiten" && selector.Sel.Name == "IsKeyPressed":
				offenders = append(offenders,
					fileSet.Position(call.Pos()).String()+" "+pkg.Name+"."+selector.Sel.Name)
			}
			return true
		})
	}
	if len(offenders) > 0 {
		t.Fatalf("前端要透過 `a.justPressed`／`a.keyDown` 讀鍵盤，直接呼叫的有：\n  %s",
			strings.Join(offenders, "\n  "))
	}
}

// ★ 游標不能停在**已經不存在的選項**上。
//
// `choiceCursor` 是前端自己的狀態，而選項是 ECL 換頁時整批換掉的：商店選單有
// 五十幾項，挪到第 8 項之後換到只有三項的荒野選單，游標還是 8——按下去 `Select`
// 直接回 `choice 8 is invalid in mode 1`，**玩家看到的是一行錯誤訊息**。
//
// ⚠ 重設原本只發生在幾條特定路徑上（選完一項、載入存檔），涵蓋不了所有換頁。
func TestChoiceCursorIsClampedToTheCurrentChoices(t *testing.T) {
	for _, item := range []struct {
		name    string
		cursor  int
		choices int
		want    int
	}{
		{name: "選項變少時夾到最後一項", cursor: 8, choices: 3, want: 2},
		{name: "沒有選項時歸零", cursor: 8, choices: 0, want: 0},
		{name: "在範圍內就不要動", cursor: 1, choices: 3, want: 1},
		{name: "負的歸零", cursor: -1, choices: 3, want: 0},
	} {
		t.Run(item.name, func(t *testing.T) {
			state := game.NewState(locale.Catalog{})
			state.Choices = make([]string, item.choices)
			application := &app{state: &state, choiceCursor: item.cursor}
			application.clampChoiceCursor()
			if application.choiceCursor != item.want {
				t.Fatalf("游標 ＝ %d，want %d", application.choiceCursor, item.want)
			}
		})
	}
}
