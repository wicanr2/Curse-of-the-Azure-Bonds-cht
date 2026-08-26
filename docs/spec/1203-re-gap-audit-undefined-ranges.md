# 1203：未定義區段判定（RE-11）——兩平台 overlay 的機械分類與待人工清單

狀態：`READY`（分類器與報表已落地；「待人工」逐段語意判定是後續輪的工作）

## 這一份在回答什麼

spec 559 的全模組掃描完成後，兩平台 overlay 裡仍有一批位元組不屬於任何
函式（IDA 匯出的 `undefined_ranges`）：DOS 929 段／16,065 bytes、
PC-98 388 段／20,321 bytes。RE-11 問的是：**這些位元組是什麼**——
對齊縫？填充？文字？資料表？還是掃描漏掉的程式碼？

一段一段用 IDA 人工看太貴，而其中大半可以機械判定形狀。本輪落地
`cmd/re-gap-audit`：機械分類能收的收掉，收不掉的逐段附 hex 前綴
排成清單，讓「待人工」有一份有序的台帳而不是一個總數。

## 資料源與分母

- 分母＝`workplace/re-sweep/{dos,pc98}/out/overlay-*.json` 的
  `undefined_ranges`（十進位 offset、end 排除式，直接索引同名
  `overlays/overlay-NN.bin`）。
- **常駐執行檔不在分母**：DOS `START.EXE` 與 PC-98 `GAME.EXE` 的匯出
  `undefined_ranges` 都是 0（IDA 全數認成程式碼或資料）。
- 工具對每段先驗證 `start < end ≤ len(bin)`，範圍出界直接 fail——
  分母與 bytes 對不上時要噴錯，不能靜默跳過。

## 分類法（形狀判定，不證明語意）

依序套用，先中先收：

| 類別 | 判定 | 語意解讀 |
|---|---|---|
| 碎屑（crumb） | ≤2 bytes | 函式尾的對齊縫、單一 `int 3`／`nop` 級殘渣 |
| 同值填充（fill） | 全段同一 byte（實測全為 `00h`） | linker 填充 |
| 文字（text） | 文字覆蓋率 ≥85%，或 Pascal 長度前綴鏈覆蓋 ≥80% | 字串常數（IDA 沒認成 string） |
| 待人工（pending） | 以上皆非 | **還沒回答的部分**——資料表或漏掉的程式碼 |

「文字覆蓋率」的定義：可列印 ASCII（`20h..7Eh`）算 1 byte、合法
Shift-JIS 雙位元組（lead `81h..9Fh`／`E0h..EFh`＋trail `40h..FCh` 去
`7Fh`）整對算 2。**PC-98 版訊息是 Shift-JIS，只認 ASCII 的判定器會把
306 段／18,876 bytes 全部誤判成待人工**（初版實測值）；加上 SJIS 後
待人工掉到 95 段／7,326 bytes。
⇒ 規則：**做文字形狀判定時，判定器要涵蓋該平台實際使用的編碼**，
只認 ASCII 的「可列印比率」在雙位元組平台上是儀器的洞
（同 `~/diagnosis-notes/docs/02-query-returned-empty/` 的「過濾器有洞」型）。

## 結果（2026-08-26，`docs/audit/re-gap-audit.{md,json}`）

| 平台 | 段數 | bytes | 碎屑 | 填充 | 文字 | 待人工（段／bytes）|
|---|---:|---:|---:|---:|---:|---|
| DOS | 929 | 16,065 | 476 | 7 | 259 | 187／5,818 |
| PC-98 | 388 | 20,321 | 51 | 5 | 237 | 95／7,326 |

機械分類收掉 79%（DOS bytes）／64%（PC-98 bytes）。正對照樣本：

- 文字：兩平台 `overlay-01` `0051h` 同一段 credits
  （`$based on the tsr novel "azure bonds"…`，Pascal 長度前綴鏈），
  DOS 小寫、PC-98 大寫混排，互相印證。
- 填充：抽 4 段全為 `00h` 且全段同值。

## 待人工段的初步觀察（形狀線索，非結論）

- **兩平台同座標的同一段**：`overlay-14 0044h` DOS 258 B／PC-98 258 B，
  前綴都是 `8A 46 06 …`（`mov al,[bp+6]`）——像是**同一支沒被 seed 到的
  函式**兩平台都沒收進來。這類段是「漏掉的程式碼」的最強候選。
- 最大段：PC-98 `overlay-21 1C56..2305h`（1,711 B，前綴含 `74 18` jz、
  `0E 57` push cs/push di）、DOS `overlay-22 650E..68A5h`（919 B）、
  DOS `overlay-11 0419..06D8h`（703 B）。形狀像程式碼或碼＋表混合。
- 一批 DOS 段（如 `overlay-13 0256h`、`overlay-23 1F19h`）是**夾控制碼的
  訊息模板**（`06 74 61 6B 65 73 20` = "\x06takes " 後接 `12h` 佔位）——
  Pascal 鏈判定要求內容 100% 文字所以沒收；語意上接近文字類，
  但佔位碼的意義要人工對（可能對上 FINDSTR 的參數替換）。

逐段清單（≥16 B、依大小排、附前 16 bytes hex）在
`docs/audit/re-gap-audit.md` 的「待人工段」表；全量含 <16 B 的在
同名 `.json`（schema `coab-re-gap-audit/1`）。

## 邊界

- 分類只證明**位元組形狀**：「文字」沒有逐段解碼核對語意；「待人工」
  也可能事後判成填充變體或表。把任何一段升級成定論前，要回 IDA 看
  上下文（誰引用它、前後函式是誰）。
- 報表由工具重生（`./tools/go.sh run ./cmd/re-gap-audit/`），不要手改。
- 閾值（85%／80%／16 B）是工程選擇，改動要重跑並重讀待人工清單。
