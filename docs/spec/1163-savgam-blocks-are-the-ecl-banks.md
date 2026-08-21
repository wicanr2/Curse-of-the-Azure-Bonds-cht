# 1163 — SAVGAM 的前四塊就是 ECL 位址空間的前四區

- 證據等級：`exact`（尺寸、順序與九個具名位移三重對齊）
- 前置：spec 1096（ECL 位址空間分五區）、spec 181（SAVGAM 固定前綴的寫入順序）
- 相關：spec 1162（`4C00h`..`4C0Fh` 是每一段自己的暫存）

## 結論

| SAVGAM 欄位 | 大小 | 區 | ECL 位址 | 寬度 | bank 指標 |
|---|---:|---:|---|---|---|
| `Area1` | `0x800` | 0 | `4B00h`..`4EFFh`（1,024 格）| word | `DS:4F99h` |
| `Area2` | `0x800` | 1 | `7C00h`..`7FFFh`（1,024 格）| word | `DS:4F9Dh` |
| `Runtime`（`stru_1B2CA`）| `0x400` | 2 | `7A00h`..`7BFFh`（512 格）| word | `DS:4FA1h` |
| `ECL` | `0x1E00` | 3 | `8000h`..`9DFFh`（7,680 格）| byte | `DS:4FA5h` |

bank 內位移是 `(位址 − 基底) × 2`（spec 1096），所以

```
ECL 位址 = 基底 + SAVGAM 位移 ÷ 2
```

## 三重對齊

**一、尺寸。** 四塊的大小與四區的格數×寬度一一相符，沒有一塊需要補零或截斷。

**二、順序。** spec 181 從 `SaveGame` 讀到的寫入順序是 `Area1`、`Area2`、
`stru_1B2CA`、`ECL memory`；spec 1096 從寫入端讀到的 bank 指標順序是
`DS:4F99h`、`4F9Dh`、`4FA1h`、`4FA5h`。兩份規格互不相干，順序完全一致。

**三、具名位移。** `internal/area` 那幾個「還沒解讀語意」的位移，套上算式之後
每一個都落在已知的 ECL 位址上，而且語意對得起來：

| Area 欄位 | SAVGAM 位移 | ECL 位址 | 那個位址已知是什麼 |
|---|---|---|---|
| `InDungeon` | `0x1CC` | `4BE6h` | 重畫髒旗標 `bank0^[1CCh]`（spec 1096／1045）|
| `LastXPos` | `0x1E0` | `4BF0h` | 移動前的 X 快照（spec 1155）|
| `LastYPos` | `0x1E2` | `4BF1h` | 移動前的 Y 快照（spec 1155）|
| `LastECLBlockID` | `0x1E4` | `4BF2h` | 上一段 ECL 的編號（spec 1096 直接寫出這一格）|
| `OutdoorSkyColor` | `0x1FA` | `4BFDh` | 各段前導自己寫（`ECL6/0x40:0025h SAVE 0B 4BFD`）|
| `IndoorSkyColor` | `0x1FC` | `4BFEh` | 同上（`ECL4/0x22:004Ch SAVE 08 4BFE`）|
| `CurrentCity` | `0x342` | `4CA1h` | 世界路線表查出來的城市（`ECL1/0x50:110Ch GETTABLE 9D13 4C9D 4CA1`）|
| `HeadBlockID` | `0x5C2` | `7EE1h` | 頭像區塊（`0Eh PICTURE` 讀的 `bank1^[5C2h]`，spec 1148）|
| `GameArea` | `0x624` | `7F12h` | 全域的章節／區域旗標 |

⇒ **那幾個欄位從來就不是「Area 記錄裡的未知欄位」，它們是 ECL 變數。**

## remake 這一側

- 存檔：`State.encodeECLBanksInto` 把 session 記憶體寫進前三塊，**只動有鍵的
  位址**——沒被 VM 碰過的格子維持原版帶進來的位元組。接著 Area 編碼器覆蓋它
  自己那幾個具名欄位：那批位址兩邊都寫得到，remake 以 `s.Area` 為準。
- 讀檔：`State.seedECLBanksFrom` 反向把三塊讀回 session。**值是 0 的格子不收**
  ——原作分不出「寫過 0」與「沒寫過」，收進來會讓 `MemoryValue` 的第二個回傳值
  失去意義，而好幾處判斷靠它分辨。
- 區 3 不寫回：那是 ECL 程式碼本身，`loadCurrentCodeMemory` 每次換段都會從
  玩家提供的 block 重建。⚠ 代價是**自我修改過的 bytecode 不會進原版版面的
  存檔**（remake 自己的存檔有 `CodeMemory` 差異表，不受影響）。

## 對 spec 1162 的影響

原版版面裡 `4C00h`..`4C0Fh` 只有一份，就是**目前這一段**的暫存。別段停在旁邊
那幾份原版存不下來——這反過來說明原作的一次性旗標本來就只保證「當前這一段」
跨存檔存活。remake 自己的存檔格式（`SessionSnapshot.BlockScratch`）把停放的那
幾份也帶著，比原版多留一點狀態，不會少。

## 明確不宣稱

- 沒有直接讀出 `SaveGame` 裡「這四塊分別來自哪個指標」的組語；結論是尺寸、
  順序與九個具名位移三邊交叉出來的。
- 沒有宣稱區 2（`7A00h`..`7BFFh`）裡有哪些具名欄位——只確認它是那一塊。
