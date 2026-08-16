package monster

import "fmt"

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
var ItemRecordFields = []RecordField{
	{0x00, 42, "名稱（右側補 NUL／空白）", FieldDecoded, "185"},
	{0x2A, 4, "物品鏈的 next（遠指標）", FieldDocumented, "1000／832"},
	{0x2E, 1, "物品類型（索引 DS:5CF6h 那張 16 bytes 的類別表）", FieldDecoded, "1000／832"},
	{0x2F, 3, "名稱編號三格（未鑑定時顯示用）", FieldDecoded, "1036"},
	{0x32, 1, "加值（有號）", FieldDecoded, "1036"},
	{0x33, 1, "豁免用的加值", FieldDecoded, "1036"},
	{0x34, 1, "裝備中（非 0 ＝ 已裝備）", FieldDecoded, "1000／832"},
	{0x35, 1, "名稱隱藏旗標", FieldDecoded, "1036"},
	{0x36, 1, "詛咒", FieldDecoded, "1036"},
	{0x37, 2, "單位重量（word）", FieldDecoded, "1000／762"},
	{0x39, 1, "數量（重量與價值都要乘它）", FieldDecoded, "1000／762"},
	{0x3A, 2, "價值（word，有號）", FieldDecoded, "1035"},
	{0x3C, 3, "三個效果槽（`+3Ch`／`+3Dh`／`+3Eh`）", FieldDecoded, "803／807"},
}

// AffectRecordFields 是 `.FX` 效果記錄（9 bytes）的逐段台帳。
var AffectRecordFields = []RecordField{
	{0x00, 1, "效果碼", FieldDecoded, "1005"},
	{0x01, 2, "持續時間（word，分鐘）", FieldDecoded, "712"},
	{0x03, 1, "強度（0FFh ＝ 永久）", FieldDecoded, "441"},
	{0x04, 1, "生效旗標", FieldDecoded, "441"},
	{0x05, 4, "效果鏈的 next（遠指標，原樣保留）", FieldDocumented, "1000"},
}

// ValidateRecordFields 檢查台帳蓋滿整份記錄、沒有洞也沒有重疊。
func ValidateRecordFields(name string, fields []RecordField, size int) error {
	next := 0
	for index, field := range fields {
		if field.Size < 1 {
			return fmt.Errorf("%s 第 %d 段 +%02Xh 的長度是 %d", name, index, field.Offset, field.Size)
		}
		if field.Offset != next {
			return fmt.Errorf("%s 第 %d 段從 +%02Xh 開始，前一段結束於 +%02Xh：中間有洞或重疊",
				name, index, field.Offset, next)
		}
		if field.Name == "" {
			return fmt.Errorf("%s 第 %d 段沒有名字", name, index)
		}
		switch field.Status {
		case FieldDecoded, FieldDocumented:
			if field.Spec == "" {
				return fmt.Errorf("%s +%02Xh 是 %s 卻沒有出處", name, field.Offset, field.Status)
			}
		case FieldUnknown:
			if field.Spec != "" {
				return fmt.Errorf("%s +%02Xh 是 unknown 卻附了出處", name, field.Offset)
			}
		default:
			return fmt.Errorf("%s +%02Xh 的狀態 %q 不是三種之一", name, field.Offset, field.Status)
		}
		next = field.Offset + field.Size
	}
	if next != size {
		return fmt.Errorf("%s 台帳只蓋到 +%02Xh，記錄是 %#X bytes", name, next, size)
	}
	return nil
}
