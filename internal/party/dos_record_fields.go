package party

import "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"

// DOSRecordFieldStatus 是一段位元組在 remake 這側的狀態。
//
// ★ 三態的分界寫死，免得下次憑印象填：
//
//	decoded     remake 的解析器**真的讀它**，值會影響匯入結果。
//	            `cmd/save-field-coverage` 用位元組突變逐格量測，對不上會紅。
//	documented  有規格說得出它是什麼，但解析器沒讀。匯入時原樣保留。
//	unknown     沒有任何出處。**不是「沒有用」**——只是還沒查到。
type DOSRecordFieldStatus string

const (
	DOSFieldDecoded    DOSRecordFieldStatus = "decoded"
	DOSFieldDocumented DOSRecordFieldStatus = "documented"
	DOSFieldUnknown    DOSRecordFieldStatus = "unknown"
)

// DOSRecordField 是角色記錄裡的一段位元組。
type DOSRecordField struct {
	Offset int
	Size   int
	Name   string
	Status DOSRecordFieldStatus
	// Spec 是出處。`decoded` 與 `documented` 都必須有；`unknown` 一定沒有。
	Spec string
}

// DOSPlayerRecordFields 是 DOS 角色記錄（`CHARREC`，`1A6h` bytes）的逐段台帳。
//
// ★ 為什麼要逐段而不是只列已知欄位。 「已知欄位清單」回答不了 `RE-05` 的問題
// ——「還有多少不知道」。把**每一個位元組**都放進表裡、用 fail-closed 測試擋住
// 洞與重疊之後，`unknown` 的總數才是可以引用的數字。
//
// ⚠ 匯入不會因為 `unknown` 而遺失資料：`State.LoadSAVGAMSlot` 保留整份原始
// 記錄，`PatchDOSPlayerRecord` 只寫已知位移（spec 185）。這張表衡量的是
// **理解程度**，不是保真度。
var DOSPlayerRecordFields = []DOSRecordField{
	{0x000, 1, tooltext.Text("h.a0d9d9952154"), DOSFieldDecoded, "185"},
	{0x001, 15, tooltext.Text("h.e134154c0699"), DOSFieldDecoded, "185"},
	{0x010, 1, tooltext.Text("h.7516d9f135ec"), DOSFieldDecoded, "1079"},
	{0x011, 1, tooltext.Text("h.123b5d1e3f23"), DOSFieldDecoded, "1079"},
	{0x012, 1, tooltext.Text("h.2a465a0924f7"), DOSFieldDecoded, "1079"},
	{0x013, 1, tooltext.Text("h.58adf2d5ce12"), DOSFieldDecoded, "1079"},
	{0x014, 1, tooltext.Text("h.881e685108f8"), DOSFieldDecoded, "1079"},
	{0x015, 1, tooltext.Text("h.961fa5f18fe0"), DOSFieldDecoded, "1079"},
	{0x016, 1, tooltext.Text("h.68b7af260181"), DOSFieldDecoded, "1079"},
	{0x017, 1, tooltext.Text("h.7a81780969b9"), DOSFieldDecoded, "1079"},
	{0x018, 1, tooltext.Text("h.297256db5435"), DOSFieldDecoded, "1079"},
	{0x019, 1, tooltext.Text("h.87eb78bfed12"), DOSFieldDecoded, "1079"},
	{0x01A, 1, tooltext.Text("h.881a17e66aff"), DOSFieldDecoded, "1079"},
	{0x01B, 1, tooltext.Text("h.d295951be8e2"), DOSFieldDecoded, "1079"},
	{0x01C, 1, tooltext.Text("h.0f2cc6f829c8"), DOSFieldDecoded, "1086"},
	{0x01D, 1, tooltext.Text("h.072b5650b7f9"), DOSFieldDecoded, "1086"},
	{0x01E, 84, tooltext.Text("h.8ac9643659a9"), DOSFieldDecoded, "1016／792"},
	{0x072, 1, tooltext.Text("h.44b3331abe25"), DOSFieldDocumented, "1164"},
	{0x073, 1, tooltext.Text("h.508bd10b607c"), DOSFieldDecoded, "1000／1140"},
	{0x074, 1, tooltext.Text("h.aaa493f24fdf"), DOSFieldDecoded, "998／1084"},
	{0x075, 1, tooltext.Text("h.995bcdf667d4"), DOSFieldDecoded, "1093"},
	{0x076, 2, tooltext.Text("h.09d6a87d4c6c"), DOSFieldDecoded, "1099"},
	{0x078, 1, tooltext.Text("h.3abe5d68a83d"), DOSFieldDecoded, "185"},
	{0x079, 100, tooltext.Text("h.0a4636557ab8"), DOSFieldDecoded, "815"},
	{0x0DD, 1, tooltext.Text("h.390d9cc5a973"), DOSFieldDocumented, "806／833"},
	{0x0DE, 1, tooltext.Text("h.e6d900b06c75"), DOSFieldDocumented, "1164"},
	{0x0DF, 5, tooltext.Text("h.e7635cfb8d9f"), DOSFieldDecoded, "1111"},
	{0x0E4, 1, tooltext.Text("h.7f4d48012e13"), DOSFieldDecoded, "1000／683"},
	{0x0E5, 1, tooltext.Text("h.fbc7bbd0bb4e"), DOSFieldDecoded, "811／815"},
	{0x0E6, 1, tooltext.Text("h.17505dfc26ac"), DOSFieldDecoded, "1079／1164"},
	{0x0E7, 1, tooltext.Text("h.6ffa0b3e4b78"), DOSFieldDecoded, "1125"},
	{0x0E8, 1, tooltext.Text("h.059ba5bdae57"), DOSFieldDecoded, "1125"},
	{0x0E9, 1, tooltext.Text("h.ebd46e349ff6"), DOSFieldDocumented, "834"},
	{0x0EA, 8, tooltext.Text("h.f8036258fa4f"), DOSFieldDecoded, "185"},
	{0x0F2, 4, tooltext.Text("h.ab1f8140b600"), DOSFieldDecoded, "185"},
	{0x0F6, 1, tooltext.Text("h.8c744e30507f"), DOSFieldDocumented, "1164"},
	{0x0F7, 1, tooltext.Text("h.655e442d40f0"), DOSFieldDecoded, "758"},
	{0x0F8, 1, tooltext.Text("h.5ebee22aa84a"), DOSFieldDocumented, "1164"},
	{0x0F9, 1, tooltext.Text("h.c2c22b508f0e"), DOSFieldDocumented, "1164"},
	{0x0FA, 1, tooltext.Text("h.0a25ad3e4542"), DOSFieldDocumented, "1164"},
	{0x0FB, 14, tooltext.Text("h.153f959a23b2"), DOSFieldDecoded, "1000"},
	{0x109, 8, tooltext.Text("h.fda2e58708a6"), DOSFieldDecoded, "728"},
	{0x111, 8, tooltext.Text("h.f8b50d054da9"), DOSFieldDecoded, "728／1196"},
	{0x119, 1, tooltext.Text("h.5f276e7a9b0a"), DOSFieldDecoded, "1086"},
	{0x11A, 1, tooltext.Text("h.96832340eb55"), DOSFieldDocumented, "1164"},
	{0x11B, 1, tooltext.Text("h.35a45173039d"), DOSFieldDecoded, "1102"},
	{0x11C, 1, tooltext.Text("h.694ecdbe1591"), DOSFieldDocumented, "1010／1164／1180"},
	// ⚠ 三組基準值的**陣列基底**是 `+11Dh`／`+11Fh`／`+121h`，但原作寫的是
	// `基底 ＋ i`（i ＝ 1、2），所以實際的格子從基底的下一個位元組起算。
	{0x11D, 1, tooltext.Text("h.405b2132c139"), DOSFieldDocumented, "1164"},
	{0x11E, 2, tooltext.Text("h.d5130220c575"), DOSFieldDocumented, "1000／795"},
	{0x120, 2, tooltext.Text("h.8a4458b23f2f"), DOSFieldDocumented, "1000／795"},
	{0x122, 2, tooltext.Text("h.c566305039c3"), DOSFieldDocumented, "1000／795"},
	{0x124, 1, tooltext.Text("h.bb54296e8cb0"), DOSFieldDecoded, "1000／1140"},
	// `+125h` 為 0 時力量的命中／傷害調整直接回 0（spec 694／697）。
	// `MON*CHA` 的 81 筆裡 61 筆是 0、20 筆是 1，所以它不是「玩家角色」的同義詞。
	{0x125, 1, tooltext.Text("h.efbd770d168a"), DOSFieldDecoded, "694／697"},
	{0x126, 1, tooltext.Text("h.fc6af97a554d"), DOSFieldDocumented, "1164"},
	{0x127, 4, tooltext.Text("h.56e5ec26629f"), DOSFieldDecoded, "185"},
	{0x12B, 1, tooltext.Text("h.410bcc642f04"), DOSFieldDocumented, "1120"},
	{0x12C, 1, tooltext.Text("h.d1250b53804a"), DOSFieldDecoded, "185"},
	{0x12D, 15, tooltext.Text("h.ffacb7d658a1"), DOSFieldDecoded, "1016"},
	{0x13C, 2, tooltext.Text("h.aaf688c52cf1"), DOSFieldDocumented, "1164"},
	{0x13E, 1, tooltext.Text("h.2b11240d2f61"), DOSFieldDocumented, "1164"},
	{0x13F, 1, tooltext.Text("h.605305329bf1"), DOSFieldDocumented, "1164"},
	{0x140, 1, tooltext.Text("h.a1e87c23dd0f"), DOSFieldDocumented, "1164"},
	{0x141, 1, tooltext.Text("h.1dc2fac1934c"), DOSFieldDecoded, "185／1164"},
	{0x142, 1, tooltext.Text("h.83b7b2d5ffa8"), DOSFieldDecoded, "185／1164"},
	{0x143, 1, tooltext.Text("h.a40b9187f6a5"), DOSFieldDecoded, "185／1164"},
	{0x144, 1, tooltext.Text("h.f5e0f6f4d0bc"), DOSFieldDecoded, "1093／1164"},
	{0x145, 7, tooltext.Text("h.5b2075dd4db1"), DOSFieldDocumented, "1164"},
	{0x14C, 1, tooltext.Text("h.b463ad2a025d"), DOSFieldDocumented, "1000"},
	{0x14D, 4, tooltext.Text("h.0f12b114d156"), DOSFieldDecoded, "1000"},
	{0x151, 52, tooltext.Text("h.87640e5aaae6"), DOSFieldDocumented, "1000"},
	{0x185, 1, tooltext.Text("h.b1567be44efe"), DOSFieldDocumented, "1000／1004"},
	{0x186, 1, tooltext.Text("h.80ede011ff3f"), DOSFieldDecoded, "1000"},
	{0x187, 2, tooltext.Text("h.df9e6f38aa25"), DOSFieldDocumented, "1000／974"},
	{0x189, 4, tooltext.Text("h.0ad20698c452"), DOSFieldDocumented, "689／815"},
	{0x18D, 4, tooltext.Text("h.ca508efc3c7f"), DOSFieldDocumented, "806"},
	{0x191, 1, tooltext.Text("h.64fd0db7c2c0"), DOSFieldDocumented, "1164"},
	{0x192, 1, tooltext.Text("h.07ead3cd4ead"), DOSFieldDecoded, "1098／1164"},
	{0x193, 1, tooltext.Text("h.7cbdafc0deec"), DOSFieldDocumented, "1164"},
	{0x194, 1, tooltext.Text("h.b95ac63362f3"), DOSFieldDocumented, "1164"},
	{0x195, 1, tooltext.Text("h.60f33c31115a"), DOSFieldDocumented, "833／1010"},
	{0x196, 1, tooltext.Text("h.2e69a502b8e6"), DOSFieldDocumented, "1010"},
	{0x197, 1, tooltext.Text("h.deae298a72f5"), DOSFieldDocumented, "777／1112"},
	{0x198, 1, tooltext.Text("h.5a09e30ae71a"), DOSFieldDocumented, "837"},
	{0x199, 1, tooltext.Text("h.64f76db77fff"), DOSFieldDocumented, "1000"},
	{0x19A, 1, tooltext.Text("h.d7cfb51a2c88"), DOSFieldDocumented, "1000"},
	{0x19B, 1, tooltext.Text("h.2493435a54c5"), DOSFieldDocumented, "1000／1137"},
	{0x19C, 2, tooltext.Text("h.cab30c3ebe65"), DOSFieldDocumented, "808／1010"},
	{0x19E, 2, tooltext.Text("h.c2ded79d8588"), DOSFieldDocumented, "1000／795"},
	{0x1A0, 2, tooltext.Text("h.e1190769655a"), DOSFieldDocumented, "1000／795"},
	{0x1A2, 2, tooltext.Text("h.1bbc36beacd6"), DOSFieldDocumented, "1000／795"},
	{0x1A4, 1, tooltext.Text("h.6ec628b29160"), DOSFieldDecoded, "185"},
	{0x1A5, 1, tooltext.Text("h.c75066a5b3b0"), DOSFieldDecoded, "1000／683"},
}

// ValidateDOSPlayerRecordFields 檢查台帳蓋滿整份記錄、沒有洞也沒有重疊。
//
// ★ 這是整張表的價值所在。 少一段就會讓 `unknown` 的數字偏低，而那個數字正是
// 「還剩多少不知道」的答案；重疊則會讓同一個位元組被算兩次。
func ValidateDOSPlayerRecordFields() error {
	next := 0
	for index, field := range DOSPlayerRecordFields {
		if field.Size < 1 {
			return tooltext.Errorf("h.1cb01188f695", index, field.Offset, field.Size)
		}
		if field.Offset != next {
			return tooltext.Errorf("h.e13070aaee20", index, field.Offset, next)
		}
		if field.Name == "" {
			return tooltext.Errorf("h.16f47dbab574", index, field.Offset)
		}
		switch field.Status {
		case DOSFieldDecoded, DOSFieldDocumented:
			if field.Spec == "" {
				return tooltext.Errorf("h.e71b1f46af1c", field.Offset, field.Status)
			}
		case DOSFieldUnknown:
			if field.Spec != "" {
				return tooltext.Errorf("h.eeef2792bdd5", field.Offset, field.Spec)
			}
		default:
			return tooltext.Errorf("h.b91921c0176b", field.Offset, field.Status)
		}
		next = field.Offset + field.Size
	}
	if next != DOSPlayerRecordSize {
		return tooltext.Errorf("h.318d402ef9fe", next, DOSPlayerRecordSize)
	}
	return nil
}

// DOSPlayerRecordFieldAt 回傳涵蓋某個位移的那一段。
func DOSPlayerRecordFieldAt(offset int) (DOSRecordField, bool) {
	for _, field := range DOSPlayerRecordFields {
		if offset >= field.Offset && offset < field.Offset+field.Size {
			return field, true
		}
	}
	return DOSRecordField{}, false
}
