package monster

import "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"

// RecordFieldStatus 與 `internal/party` 的三態同義：`decoded` 解析器真的讀它、
// `documented` 有出處但解析器沒讀、`unknown` 還沒查到。
type RecordFieldStatus string

const (
	FieldDecoded    RecordFieldStatus = "decoded"
	FieldDocumented RecordFieldStatus = "documented"
	FieldUnknown    RecordFieldStatus = "unknown"
)

// RecordField 是記錄裡的一段位元組。
type RecordField struct {
	Offset int
	Size   int
	Name   string
	Status RecordFieldStatus
	Spec   string
}

// ItemRecordFields 是 `.SWG` 物品記錄（`3Fh` ＝ 63 bytes）的逐段台帳。
//
// ★ `+2Ah` 是鏈結指標，不是名字的一部分。 名字只到 `+29h`；原作走物品鏈用的
// 就是 `q := q^[2Ah]`（spec 1000／832）。把名字讀成 46 bytes 會把指標讀進字串，
// 而那在英文原版看起來只是「名字後面有幾個怪字元」。
//
// ⚠ PC-98 的 `CHARITEMFILREC` 把這一段宣告成 `NAME: STR40`（41 bytes，
// `+00h`..`+28h`）加一格 `TITLE`（`+29h`，boolean）——remake 這一側是把 42 bytes
// 當右側補白的字串一起讀。兩種讀法在現有資料上結果相同（名字沒有長到 40 字），
// 但 `+29h` 到底是不是名字的一部分還沒對過（spec 1165）。
var ItemRecordFields = []RecordField{
	{0x00, 42, tooltext.Text("h.380c6e56be01"), FieldDecoded, "185／1165"},
	{0x2A, 4, tooltext.Text("h.8051f1fd954a"), FieldDocumented, "1000／832"},
	{0x2E, 1, tooltext.Text("h.584bb84381eb"), FieldDecoded, "1000／832"},
	{0x2F, 3, tooltext.Text("h.87f8c62a9afb"), FieldDecoded, "1036"},
	{0x32, 1, tooltext.Text("h.e8a0ee7f269d"), FieldDecoded, "1036"},
	{0x33, 1, tooltext.Text("h.793db5e6efcc"), FieldDecoded, "1036"},
	{0x34, 1, tooltext.Text("h.b10838e60aa8"), FieldDecoded, "1000／832"},
	{0x35, 1, tooltext.Text("h.0e4e929fb8a6"), FieldDecoded, "1036／1165"},
	{0x36, 1, tooltext.Text("h.9cb5c84762e2"), FieldDecoded, "1036"},
	{0x37, 2, tooltext.Text("h.03a77167a463"), FieldDecoded, "1000／762"},
	{0x39, 1, tooltext.Text("h.ed882b5da2a8"), FieldDecoded, "1000／762"},
	{0x3A, 2, tooltext.Text("h.508f30d856d0"), FieldDecoded, "1035"},
	{0x3C, 3, tooltext.Text("h.146516db986e"), FieldDecoded, "803／807"},
}

// AffectRecordFields 是 `.FX` 效果記錄（9 bytes）的逐段台帳。
var AffectRecordFields = []RecordField{
	{0x00, 1, tooltext.Text("h.5355e3997b24"), FieldDecoded, "1005"},
	{0x01, 2, tooltext.Text("h.ccc6543ac468"), FieldDecoded, "712"},
	{0x03, 1, tooltext.Text("h.1bb0cdbbe288"), FieldDecoded, "441／1165"},
	{0x04, 1, tooltext.Text("h.199cb8a07c55"), FieldDecoded, "441／1165"},
	{0x05, 4, tooltext.Text("h.7c871a69b652"), FieldDocumented, "1000"},
}

// ValidateRecordFields 檢查台帳蓋滿整份記錄、沒有洞也沒有重疊。
func ValidateRecordFields(name string, fields []RecordField, size int) error {
	next := 0
	for index, field := range fields {
		if field.Size < 1 {
			return tooltext.Errorf("h.8c523dae0060", name, index, field.Offset, field.Size)
		}
		if field.Offset != next {
			return tooltext.Errorf("h.b99176242db2", name, index, field.Offset, next)
		}
		if field.Name == "" {
			return tooltext.Errorf("h.9078bac3e49a", name, index)
		}
		switch field.Status {
		case FieldDecoded, FieldDocumented:
			if field.Spec == "" {
				return tooltext.Errorf("h.9e926c0d48d4", name, field.Offset, field.Status)
			}
		case FieldUnknown:
			if field.Spec != "" {
				return tooltext.Errorf("h.4982a6c18f8d", name, field.Offset)
			}
		default:
			return tooltext.Errorf("h.9b4f49eed35b", name, field.Offset, field.Status)
		}
		next = field.Offset + field.Size
	}
	if next != size {
		return tooltext.Errorf("h.c91599ffd9ae", name, next, size)
	}
	return nil
}
