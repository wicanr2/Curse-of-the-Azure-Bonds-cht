package party

import "fmt"

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
	{0x000, 1, "名字長度", DOSFieldDecoded, "185"},
	{0x001, 15, "名字（Pascal 短字串本體）", DOSFieldDecoded, "185"},
	{0x010, 1, "力量 基準", DOSFieldDecoded, "1079"},
	{0x011, 1, "力量 現值", DOSFieldDecoded, "1079"},
	{0x012, 1, "智力 基準", DOSFieldDecoded, "1079"},
	{0x013, 1, "智力 現值", DOSFieldDecoded, "1079"},
	{0x014, 1, "睿智 基準", DOSFieldDecoded, "1079"},
	{0x015, 1, "睿智 現值", DOSFieldDecoded, "1079"},
	{0x016, 1, "敏捷 基準", DOSFieldDecoded, "1079"},
	{0x017, 1, "敏捷 現值", DOSFieldDecoded, "1079"},
	{0x018, 1, "體質 基準", DOSFieldDecoded, "1079"},
	{0x019, 1, "體質 現值", DOSFieldDecoded, "1079"},
	{0x01A, 1, "魅力 基準", DOSFieldDecoded, "1079"},
	{0x01B, 1, "魅力 現值", DOSFieldDecoded, "1079"},
	{0x01C, 1, "特殊力量 現值（18/xx 的百分位）", DOSFieldDecoded, "1086"},
	{0x01D, 1, "特殊力量 基準", DOSFieldDecoded, "1086"},
	{0x01E, 84, "記憶法術清單 84 格（最高位 ＝ 還在記憶中）", DOSFieldDecoded, "1016／792"},
	{0x072, 1, "最短休息時數 `MINREST`", DOSFieldDocumented, "1164"},
	{0x073, 1, "命中能力（八個職業槽查表取最大；重算時抄進 +199h）", DOSFieldDecoded, "1000／1140"},
	{0x074, 1, "種族編號 1..7", DOSFieldDecoded, "998／1084"},
	{0x075, 1, "職業組合編號 0..10h", DOSFieldDecoded, "1093"},
	{0x076, 2, "年齡（word）", DOSFieldDecoded, "1099"},
	{0x078, 1, "最大 HP", DOSFieldDecoded, "185"},
	{0x079, 100, "每法術旗標 100 格（法術編號 1..100）", DOSFieldDecoded, "815"},
	{0x0DD, 1, "橫掃上限的來源（回合初始化抄進戰鬥狀態 +5）", DOSFieldDocumented, "806／833"},
	{0x0DE, 1, "體型 `SIZE`", DOSFieldDocumented, "1164"},
	{0x0DF, 5, "五個豁免門檻（毒／石化／法杖／噴吐／法術）", DOSFieldDecoded, "1111"},
	{0x0E4, 1, "移動力基準（重算時抄進 +1A5h）", DOSFieldDocumented, "1000／683"},
	{0x0E5, 1, "最高職業等級（0 ＝ 零級生物）", DOSFieldDecoded, "811／815"},
	{0x0E6, 1, "前一個最高等級 `HIGHESTPREVLEVEL`（spec 185 讀成「多職角色的現行等級」，兩種讀法待對）", DOSFieldDecoded, "185／1164"},
	{0x0E7, 1, "被吸掉的等級數（復原術每次還一級）", DOSFieldDecoded, "1125"},
	{0x0E8, 1, "被吸掉的 HP 總數（還一級就還其中 1/N）", DOSFieldDecoded, "1125"},
	{0x0E9, 1, "不死生物種類 1..10（轉化矩陣的列索引）", DOSFieldDocumented, "834"},
	{0x0EA, 8, "盜賊技能八項", DOSFieldDecoded, "185"},
	{0x0F2, 4, "效果鏈頭（遠指標）", DOSFieldDecoded, "185"},
	{0x0F6, 1, "被復活過 `RAISED`", DOSFieldDocumented, "1164"},
	{0x0F7, 1, "控制／士氣（>= 80h 是 NPC）", DOSFieldDecoded, "758"},
	{0x0F8, 1, "派生值已重算 `MODIFIED`", DOSFieldDocumented, "1164"},
	{0x0F9, 1, "換職前的職業 `OLDCLASS`", DOSFieldDocumented, "1164"},
	{0x0FA, 1, "換職前的等級 `OLDLEVEL`", DOSFieldDocumented, "1164"},
	{0x0FB, 14, "七種貨幣（銅銀琥珀金白金寶石首飾，各一個 word）", DOSFieldDecoded, "1000"},
	{0x109, 8, "第一職業的八個欄位（等級在前）", DOSFieldDecoded, "728"},
	{0x111, 8, "第二職業的八個平行欄位（雙職角色才看）", DOSFieldDocumented, "728"},
	{0x119, 1, "性別", DOSFieldDecoded, "1086"},
	{0x11A, 1, "種族大類 `RACETYPE`", DOSFieldDocumented, "1164"},
	{0x11B, 1, "陣營", DOSFieldDecoded, "1102"},
	{0x11C, 1, "攻擊次數基準的第一個武器槽 `BASEATTBLOWS[0]`；spec 1010 由 DOS 側讀成「武器槽選擇」，兩種讀法待對", DOSFieldDocumented, "1010／1164"},
	// ⚠ 三組基準值的**陣列基底**是 `+11Dh`／`+11Fh`／`+121h`，但原作寫的是
	// `基底 ＋ i`（i ＝ 1、2），所以實際的格子從基底的下一個位元組起算。
	{0x11D, 1, "攻擊次數基準的第二個武器槽 `BASEATTBLOWS[1]`", DOSFieldDocumented, "1164"},
	{0x11E, 2, "攻擊骰數 基準（兩個武器槽）", DOSFieldDocumented, "1000／795"},
	{0x120, 2, "攻擊面數 基準", DOSFieldDocumented, "1000／795"},
	{0x122, 2, "傷害加值 基準（有號）", DOSFieldDocumented, "1000／795"},
	{0x124, 1, "護甲起點（重算時抄進 +19Ah；建角寫 32h ＝ AC 10）", DOSFieldDecoded, "1000／1140"},
	// `+125h` 為 0 時力量的命中／傷害調整直接回 0（spec 694／697）。
	// `MON*CHA` 的 81 筆裡 61 筆是 0、20 筆是 1，所以它不是「玩家角色」的同義詞。
	{0x125, 1, "力量調整的開關（0 ＝ 不套用命中／傷害調整）", DOSFieldDecoded, "694／697"},
	{0x126, 1, "角色亂數種子 `RANDOMID`", DOSFieldDocumented, "1164"},
	{0x127, 4, "經驗值（dword）", DOSFieldDecoded, "185"},
	{0x12B, 1, "職業可用性遮罩（與物品類別表 `+0Dh` 做 and）", DOSFieldDocumented, "1120"},
	{0x12C, 1, "基準最大 HP（不含裝備加成）", DOSFieldDecoded, "185"},
	{0x12D, 15, "每環可施放次數：牧師／德魯伊／法師各五環", DOSFieldDecoded, "1016"},
	{0x13C, 2, "基準經驗值 `BASEEXP`（word）", DOSFieldDocumented, "1164"},
	{0x13E, 1, "每點 HP 的經驗 `EXPPERHP`", DOSFieldDocumented, "1164"},
	{0x13F, 1, "頭部造型 `HEAD`", DOSFieldDocumented, "1164"},
	{0x140, 1, "身體造型 `BODY`", DOSFieldDocumented, "1164"},
	{0x141, 1, "頭像 block `ICONHEAD`", DOSFieldDecoded, "185／1164"},
	{0x142, 1, "身體圖示 block `ICONBODY`", DOSFieldDecoded, "185／1164"},
	{0x143, 1, "圖示編號 `ICONINDEX`", DOSFieldDecoded, "185／1164"},
	{0x144, 1, "人像高度 `ICONHEIGHT`（1 小／2 中）", DOSFieldDecoded, "1093／1164"},
	{0x145, 7, "人像配色表 `COLORLIST`（7 格，每格是 EGA 色號對）", DOSFieldDocumented, "1164"},
	{0x14C, 1, "物品件數（重算時數出來）", DOSFieldDocumented, "1000"},
	{0x14D, 4, "物品鏈頭（遠指標）", DOSFieldDecoded, "1000"},
	{0x151, 52, "13 個裝備槽遠指標（槽 9 雙持佔兩格）", DOSFieldDocumented, "1000"},
	{0x185, 1, "已裝備物品佔用的手數（上限 2）", DOSFieldDocumented, "1000／1004"},
	{0x186, 1, "豁免加值（派生欄位，重算時先歸零）", DOSFieldDecoded, "1000"},
	{0x187, 2, "總重（含硬幣枚數）", DOSFieldDocumented, "1000／974"},
	{0x189, 4, "隊伍／戰鬥員鏈的 next（遠指標）", DOSFieldDocumented, "689／815"},
	{0x18D, 4, "戰鬥狀態記錄的遠指標（22 bytes 的那一份）", DOSFieldDocumented, "806"},
	{0x191, 1, "解除詛咒的來源標記 `PDLNREMOVECURSE`", DOSFieldDocumented, "1164"},
	{0x192, 1, "ECL 旗標（投影位址 7CE4h）；Pascal 那邊叫保留欄 `DUM1`", DOSFieldDecoded, "1098／1164"},
	{0x193, 1, "保留欄 `DUM2`", DOSFieldDocumented, "1164"},
	{0x194, 1, "保留欄 `DUM3`", DOSFieldDocumented, "1164"},
	{0x195, 1, "狀態碼（8 ＝ 被摧毀）", DOSFieldDocumented, "833／1010"},
	{0x196, 1, "站著且能行動（1 ＝ 可以）", DOSFieldDocumented, "1010"},
	{0x197, 1, "隊號（0 是一邊，非 0 是另一邊）", DOSFieldDocumented, "777／1112"},
	{0x198, 1, "畫圖示時用的旗標（決定圖號 3 或 1）", DOSFieldDocumented, "837"},
	{0x199, 1, "重算時由 +73h 抄過來", DOSFieldDocumented, "1000"},
	{0x19A, 1, "護甲值（重算時由 +124h 起算；顯示是 60 − 它）", DOSFieldDocumented, "1000"},
	{0x19B, 1, "護甲值第二格（背後攻擊用；＝ +19Ah − 敏捷防禦調整 − 盾牌槽 − 2）", DOSFieldDocumented, "1000／1137"},
	{0x19C, 2, "兩個武器槽本回合的攻擊次數（spec 1010 寫 +19Bh ＋ 槽）", DOSFieldDocumented, "808／1010"},
	{0x19E, 2, "攻擊骰數 現值（兩個武器槽）", DOSFieldDocumented, "1000／795"},
	{0x1A0, 2, "攻擊面數 現值", DOSFieldDocumented, "1000／795"},
	{0x1A2, 2, "傷害加值 現值（有號）", DOSFieldDocumented, "1000／795"},
	{0x1A4, 1, "目前 HP", DOSFieldDecoded, "185"},
	{0x1A5, 1, "目前移動力（重算時由 +0E4h 抄過來）", DOSFieldDocumented, "1000／683"},
}

// ValidateDOSPlayerRecordFields 檢查台帳蓋滿整份記錄、沒有洞也沒有重疊。
//
// ★ 這是整張表的價值所在。 少一段就會讓 `unknown` 的數字偏低，而那個數字正是
// 「還剩多少不知道」的答案；重疊則會讓同一個位元組被算兩次。
func ValidateDOSPlayerRecordFields() error {
	next := 0
	for index, field := range DOSPlayerRecordFields {
		if field.Size < 1 {
			return fmt.Errorf("第 %d 段 +%03Xh 的長度是 %d", index, field.Offset, field.Size)
		}
		if field.Offset != next {
			return fmt.Errorf("第 %d 段從 +%03Xh 開始，前一段結束於 +%03Xh：中間有洞或重疊",
				index, field.Offset, next)
		}
		if field.Name == "" {
			return fmt.Errorf("第 %d 段 +%03Xh 沒有名字", index, field.Offset)
		}
		switch field.Status {
		case DOSFieldDecoded, DOSFieldDocumented:
			if field.Spec == "" {
				return fmt.Errorf("+%03Xh 是 %s 卻沒有出處", field.Offset, field.Status)
			}
		case DOSFieldUnknown:
			if field.Spec != "" {
				return fmt.Errorf("+%03Xh 是 unknown 卻附了出處 %q", field.Offset, field.Spec)
			}
		default:
			return fmt.Errorf("+%03Xh 的狀態 %q 不是三種之一", field.Offset, field.Status)
		}
		next = field.Offset + field.Size
	}
	if next != DOSPlayerRecordSize {
		return fmt.Errorf("台帳只蓋到 +%03Xh，記錄是 %#X bytes", next, DOSPlayerRecordSize)
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
