package combat

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// packDerivedFighterFields 是**故意不進存檔**的欄位：它們是 game pack 的規則表，
// 讀檔之後由遊戲層重新注入，存進去只會讓存檔跟著資料檔一起過期。
//
// 值是重新注入用的 setter 名字。`TestPackDerivedFieldsAreReinjectedOnLoad`
// 會去遊戲層的讀檔路徑找那個呼叫——注入被拿掉時，豁免就不再成立，
// 而那正是「讀檔之後怪物突然不會施法」的成因。
var packDerivedFighterFields = map[string]string{
	"DamageRules":              "SetDamageRules",
	"ConditionalModifierRules": "SetConditionalModifierRules",
	"MagicResistanceRules":     "SetMagicResistanceRules",
	"PostHitRules":             "SetPostHitRules",
	"MonsterSpellRules":        "SetMonsterSpellRules",
}

// 存檔完整性的閘：`Fighter` 的**每一個**匯出欄位都要能存進快照、再讀回來。
//
// ★ 為什麼要用反射掃欄位，而不是逐欄寫斷言。 逐欄斷言只擋得住「已經想到的欄位」；
// 真正會出事的是**下一個新增的欄位**——加了欄位、忘了存檔，症狀是讀檔之後某個
// 狀態悄悄回到零值，而所有既有測試照樣綠。反射版會在新增欄位的那一刻變紅。
//
// ⚠ 這條測試不驗語意，只驗「存得進、讀得回」。欄位的意義由各自的規格負責。
func TestEveryFighterFieldSurvivesASaveRoundTrip(t *testing.T) {
	fighter := Fighter{ID: "probe", Side: SideParty}
	filled := fillExportedFields(t, reflect.ValueOf(&fighter).Elem(), 1)
	if filled == 0 {
		t.Fatal("一個欄位都沒填，反射填值失效了")
	}
	// 幾個欄位有語意限制，填成反射的通用值會讓 NewBattle 直接拒絕。
	fighter.ID = "probe"
	fighter.Side = SideParty
	fighter.HitPoints = 7
	fighter.MaxHitPoints = 9
	fighter.CombatSize = 1
	fighter.CombatX, fighter.CombatY = 3, 4
	fighter.HasCombatPosition = true
	fighter.Escaped = false
	fighter.ArmorClass = 5

	battle, err := NewBattle([]Fighter{fighter,
		{ID: "foe", Side: SideEnemy, HitPoints: 5, MaxHitPoints: 5,
			HasCombatPosition: true, CombatX: 20, CombatY: 10}}, 3)
	if err != nil {
		t.Fatal(err)
	}
	stored, ok := battle.Fighter("probe")
	if !ok {
		t.Fatal("probe 不在戰鬥裡")
	}

	snapshot, err := battle.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	var decoded BattleSnapshot
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreBattle(decoded)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := restored.Fighter("probe")
	if !ok {
		t.Fatal("讀回來之後 probe 不見了")
	}
	compareExportedFields(t, reflect.ValueOf(stored), reflect.ValueOf(got), "Fighter")
}

// 逃離的隊員也要存得起來。這一條擋的是「新增了戰鬥結果狀態，卻沒改快照的
// 上限檢查」——存檔當下沒事，讀的時候才炸。
func TestFledBattleSurvivesASaveRoundTrip(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "runner", Side: SideParty, HitPoints: 10, MaxHitPoints: 10,
			HasCombatPosition: true, CombatX: 0, CombatY: 0, MovementAllowance: 12},
		{ID: "orc", Side: SideEnemy, HitPoints: 8, MaxHitPoints: 8, HitDice: 1,
			HasCombatPosition: true, CombatX: 20, CombatY: 10, MovementAllowance: 3},
	}, 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := battle.AttemptEscape("runner"); err != nil {
		t.Fatal(err)
	}
	if battle.Status() != StatusPartyFled {
		t.Fatalf("status=%v", battle.Status())
	}
	snapshot, err := battle.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreBattle(snapshot)
	if err != nil {
		t.Fatalf("逃離結局的存檔讀不回來：%v", err)
	}
	if restored.Status() != StatusPartyFled {
		t.Fatalf("讀回來的 status=%v", restored.Status())
	}
	runner, ok := restored.Fighter("runner")
	if !ok || !runner.Escaped {
		t.Fatalf("離場旗標沒有存下來：%+v", runner)
	}
}

// fillExportedFields 把每個匯出欄位填成可辨識的非零值，回傳填了幾個。
// 零值填不出來的（介面、函式、channel）跳過並回報，免得測試靜靜地少驗一塊。
func fillExportedFields(t *testing.T, value reflect.Value, seed int) int {
	t.Helper()
	filled := 0
	structType := value.Type()
	for index := 0; index < value.NumField(); index++ {
		field := value.Field(index)
		name := structType.Field(index).Name
		if !field.CanSet() {
			continue
		}
		switch field.Kind() {
		case reflect.String:
			field.SetString(name + "-value")
		case reflect.Bool:
			field.SetBool(true)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			field.SetInt(int64(seed + index + 1))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			field.SetUint(uint64(seed + index + 1))
		case reflect.Slice:
			element := reflect.New(field.Type().Elem()).Elem()
			if element.Kind() == reflect.Struct {
				fillExportedFields(t, element, seed+index)
			} else {
				fillScalar(element, seed+index+1)
			}
			field.Set(reflect.Append(field, element))
		case reflect.Array:
			for element := 0; element < field.Len(); element++ {
				item := field.Index(element)
				if item.Kind() == reflect.Struct {
					fillExportedFields(t, item, seed+index+element)
				} else {
					fillScalar(item, seed+index+element+1)
				}
			}
		case reflect.Struct:
			fillExportedFields(t, field, seed+index)
		default:
			t.Logf("欄位 %s 是 %s，反射填不出值，這一欄沒有驗到", name, field.Kind())
			continue
		}
		filled++
	}
	return filled
}

func fillScalar(value reflect.Value, seed int) {
	switch value.Kind() {
	case reflect.String:
		value.SetString("v")
	case reflect.Bool:
		value.SetBool(true)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value.SetInt(int64(seed))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value.SetUint(uint64(seed))
	}
}

// compareExportedFields 逐欄比對，錯的那一欄名字直接印出來——
// 整個結構 DeepEqual 只會說「不一樣」，找不出是哪一欄沒存到。
func compareExportedFields(t *testing.T, want, got reflect.Value, path string) {
	t.Helper()
	structType := want.Type()
	for index := 0; index < want.NumField(); index++ {
		name := structType.Field(index).Name
		if structType.Field(index).PkgPath != "" {
			continue // 未匯出欄位不進 JSON，本來就不該比
		}
		wantField, gotField := want.Field(index), got.Field(index)
		if wantField.Kind() == reflect.Struct {
			compareExportedFields(t, wantField, gotField, path+"."+name)
			continue
		}
		if _, derived := packDerivedFighterFields[name]; derived {
			if !gotField.IsZero() {
				t.Errorf("%s.%s 是 pack 衍生欄位，不該出現在存檔裡，卻讀回 %#v",
					path, name, gotField.Interface())
			}
			continue
		}
		if !reflect.DeepEqual(wantField.Interface(), gotField.Interface()) {
			t.Errorf("%s.%s 沒有存進存檔：存前 %#v，讀回 %#v",
				path, name, wantField.Interface(), gotField.Interface())
		}
	}
	if strings.Count(path, ".") > 4 {
		t.Fatalf("巢狀太深，可能是遞迴型別：%s", path)
	}
}

// 豁免的欄位必須真的在讀檔路徑被重新注入，否則豁免只是把 bug 寫成規則。
func TestPackDerivedFieldsAreReinjectedOnLoad(t *testing.T) {
	path := filepath.Join("..", "game", "creation.go")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("讀不到遊戲層的讀檔路徑：%v", err)
	}
	text := string(source)
	for field, setter := range packDerivedFighterFields {
		if !strings.Contains(text, "battle."+setter+"(") {
			t.Fatalf("%s 被豁免在存檔之外，但 %s 沒有在讀檔路徑呼叫 %s——"+
				"要嘛把注入加回去，要嘛把欄位存進存檔", field, path, setter)
		}
	}
	// 反向：Fighter 上以 Rules 結尾的切片欄位都要在豁免表裡，
	// 否則新增一組規則表時會靜靜地少注入一次。
	fighterType := reflect.TypeOf(Fighter{})
	for index := 0; index < fighterType.NumField(); index++ {
		name := fighterType.Field(index).Name
		if !strings.HasSuffix(name, "Rules") {
			continue
		}
		if _, ok := packDerivedFighterFields[name]; !ok {
			t.Fatalf("Fighter.%s 看起來是規則表卻不在豁免表裡："+
				"確認它是存檔狀態還是 pack 衍生", name)
		}
	}
}
