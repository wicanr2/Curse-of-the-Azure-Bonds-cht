# 第 435 輪：PC-98 `SCAN` 三欄記錄與原版排序

狀態：`READY`（三欄與排序仍有效；terrain property 與地形感知 producer 已由
spec 436 接續，實機 wall boundary、COMPOBJ 建立順序、手動／Quick Sleep UI
仍待續）

## 訂正與結論

spec 433 已證明 Sleep 的戰鬥 targeting handler 會呼叫 overlay 31
`SCAN 08D8h`，但當時把「第三欄與奇偶 tie」列為 `strong inference`。本輪以
擴大後的連續 IDA 9.4 audit、raw bytes 與 producer→sort→consumer 資料流訂正：

| record byte | exact producer | exact 語意 |
|---|---|---|
| `+0` | `0B04h..0B17h` | 一基底 combat object ID |
| `+1` | `0B1Bh..0B2Ah` | 所有有效 footprint cell pair 中，成功 `LOSEXISTS` 回傳最小加權距離的低 byte |
| `+2` | `0B2Eh..0B8Ch` | caller arc `<8` 時直接保存 arc；否則由最佳 cell pair 找到第一個符合的 `0..7` 方向 sector |

排序程序 local `0035h` 只讀 `+0／+1`，完全不讀 `+2`。主要鍵是 `+1`
遞增；距離相同且後方 object ID 較小時，才比較兩個 ID 的 `ID % 2`。唯一不
交換的是「後方較小 ID 為奇數、目前較大 ID 為偶數」。交換會以三 byte copy
搬移完整 record，因此方向只跟著記錄移動，不是 tie-breaker。

這個 pairwise 規則不是一般全序；不得改用 `sort.Slice` 搭配自創 comparator。
engine `combat/scanorder.Sort` 逐迴圈保存原指令形狀，CoAB
`OrderScanTargetIDs` 再把一基底 object ID 映射到 stable fighter ID；未知、零、
重複 object／fighter ID 一律失敗即關閉。

## 非破壞性輸入與重現

| 輸入 | SHA-256 | 用途 |
|---|---|---|
| overlay 31 | `6cd5e38dddeb1ea5ddc44dd7f3af68c49bd9b67198e7c42247f1d08743050081` | `STARTVECTOR／STEPVECTOR／LOSEXISTS／INARC／SCAN／sort` |

原始 overlay 唯讀保存。`scripts/ida/pc98_spell_targeting_audit.idc` 的 overlay
31 audit 範圍擴大為：sort `0035h..019Dh`、`STARTVECTOR 019Dh..0246h`、
`STEPVECTOR 0246h..03EAh`、`LOSEXISTS 03EAh..054Ah`、`INARC
054Ah..08D8h`、`SCAN 08D8h..0BA5h`。IDA database 與輸出只建立在
`/tmp/coab-ida-435-final` 全新副本；本輪報告 1,187 行、SHA-256
`744f8553f9b31a4fd3a1240b867f16e861d0712999ce1226d0f660601e206403`。
exit code 0 之外另驗證報告非空；原始 overlay、symbol table 與既有 database
均未修改或 rename。

## Producer 資料流

`SCAN 08D8h`：

- `0945h` 清 `LASTSIGHT 9F30h`；`094Ah..095Fh` 以 object ID `1..9740h`
  掃描。
- object table 每筆四 bytes；`973Dh+4*ID` 是 X、`973Eh+4*ID` 是 Y，
  `9740h+4*ID` 傳入 typed `CALCBIGOFFSET`。source size 也用相同 helper，
  cell index `0..3`；無效 cell 寫 `FFh`。
- 所有有效 cell pair 先經 `INARC 054Ah`，再進 `LOSEXISTS 03EAh`。成功
  路徑若比目前 word `00FFh` 小，保存距離與兩端最佳 cell index。
- `STEPVECTOR` 對 cardinal step 將 metric 加二；secondary axis 同步前進時
  再加一，故 diagonal step 為三。`LOSEXISTS` budget 是
  `2*inputRange+1`，並從 50×25 combat map 讀 tile／terrain properties。
- 至少一對 cell 成功才建立三 byte record。Sleep caller 傳 arc `FFh`，
  初次 `INARC` 因 normalized direction 8 接受全部方向；建立 record 前再由
  最佳 pair 依序試 `0..7`，保存第一個符合 sector。

以上 exact 關閉記錄來源與數值流。terrain property table 的正式 `HT／LOS／
SYM` 名稱、`TACTICALMAP.XRAY／TD` 與靜態阻擋比較已由 spec 436 接續；牆角
runtime 邊界仍未閉合，不能把 bounded producer 寫成「完整戰鬥 LOS」。

## 排序指令

local `0035h` 的關鍵位址：

- `007Ch..00BEh`：以 `3*index` 取兩筆 `+0／+1`。
- `00C2h..00C4h`：後者 `+1` 較小即交換。
- `00D2h..00EAh`：距離不等且後者不小就不交換。
- `00EDh..00F5h`：只有後者 object ID 較小才進奇偶判斷。
- `00F8h..0114h`：兩個 byte 先 zero-extend，再各自 `idiv 2`；後者餘數
  大於目前餘數時不交換。
- `0116h..017Eh`：以三次 3-byte copy 交換完整記錄。

因此方向欄沒有 comparator consumer。spec 433 的舊敘述已在同一里程碑
supersede，避免 compact 後再次採用。

## Remake 邊界與驗證

- engine package：`combat/scanorder`，不認識 CoAB spell、map 或 fighter。
- CoAB adapter：`OrderScanTargetIDs`，只驗證 legacy object→stable ID 投影並
  回傳排序後 ID；不自行發現候選。
- engine 回歸覆蓋距離主要鍵、奇偶例外、方向不參與，以及 256 組 byte corpus
  對 literal nested-loop reference。
- CoAB 回歸覆蓋 stable ID 映射、輸入不被修改，以及零／重複／空 ID
  fail-closed。

spec 436 已建立 `terrain producer → ordered stable IDs → CastSleepOrdered` 的
bounded transaction，但仍不是正常玩家 Sleep 路徑。下一步必須關閉 `INARC`
與 COMPOBJ builder，並以 PC-98／DOSBox 固定戰場交叉驗證 wall、large
footprint 與 cursor，才能接入手動／Quick UI。
