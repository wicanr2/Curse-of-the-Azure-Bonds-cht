# 1199：提爾佛頓下水道的圖緣交接（三個區域 ＋ 火刀據點入口）

狀態：`READY`

## 結論

`GEO2/0x03`（提爾佛頓下水道）**不是一張連通的地圖**。以走路可達性算，它被切成
三塊互不相通的區域，而區域之間唯一的通道是**走出圖外**：踏出邊界時
`ECL2/0x03` 的圖緣處理常式讀 `C04B`（X）與 `C04D`（朝向），把隊伍**傳送回同一張
圖的另一個 X**，或者 `NEWECL` 到別的 block。

| 區域 | X 範圍 | 走得到的格子 |
|---|---|---:|
| A（公會進來的那一側）| 0..4 | 63 |
| B | 5..9 | 73 |
| C（通火刀據點）| 10..15 | 87 |

火刀據點的入口在**區域 C 的南緣**。從區域 A 直接走是走不到的——中間要經過兩次
圖緣傳送。

## 圖緣處理常式（`ECL2/0x03` 入口 `0x00F2`）

`7ED5h = 1`（引擎標記「這一步想走出圖外」）時才進這一段；`C04D = 0`（朝北）走
上緣那一支，其餘走下緣那一支。

### 下緣（往南走出圖外）

```text
0x0119  SAVE 00 C04C            { Y := 0 }
0x011F  COMPARE C04B 02 → X := 09 → 停在本圖
0x0131  COMPARE C04B 05 → X := 11 → 停在本圖
0x0143  COMPARE C04B 09 → X := 15 → 停在本圖
0x0155  COMPARE 4C2A 01 → 告示牌（見下）
0x016A  SAVE 00 C04C            { Y := 0 }
0x0170  SUBTRACT 02 C04B C04B   { X := X − 2 }
0x0179  NEWECL 04               { 火刀據點 }
```

### 上緣（往北走出圖外）

```text
0x017C  SAVE 0F C04C            { Y := 15 }
0x0182  COMPARE C04B 09 → X := 02 → 停在本圖
0x0194  COMPARE C04B 0B (>=) → X := X − 6 → 停在本圖
0x01A9  COMPARE 4C2A 01 → 告示牌
0x01BE  NEWECL 02               { 提爾佛頓城 }
```

## 十一個走得出去的邊界格

GEO 的移動遮罩（`cmd/geo-move-mask`，位元 N=1／E=2／S=4／W=8 代表**走得出去**）
與上面的 X 判斷完全對得上。這十一格就是全部：

| 方向 | X | 腳本做什麼 | 落點 |
|---|---:|---|---|
| 北 | 0 | `NEWECL 02` | 提爾佛頓城 |
| 北 | 4 | `NEWECL 02` | 提爾佛頓城 |
| 北 | 9 | 同圖傳送 | `(2,15)` |
| 北 | 11 | 同圖傳送 | `(5,15)` |
| 北 | 15 | 同圖傳送 | `(9,15)` |
| 南 | 2 | 同圖傳送 | `(9,0)` |
| 南 | 5 | 同圖傳送 | `(11,0)` |
| 南 | 9 | 同圖傳送 | `(15,0)` |
| 南 | 10 | `NEWECL 04` | 火刀據點 `(8,0)` |
| 南 | 13 | `NEWECL 04` | 火刀據點 `(11,0)` |
| 南 | 15 | `NEWECL 04` | 火刀據點 `(13,0)` |

## `4C2A` 不是「開啟荒野出口」的旗標

`4C2A` 只在 `ECL2/0x04:042F`（救出喬吉、手札第 53 條）寫過一次。它在圖緣處理
常式裡的作用是**把還沒用過的那條 `NEWECL` 換成告示牌**：

```text
COMPARE 4C2A 01 / IF = / GOTO 0x81C1
0x01C1  PRINTCLEAR "THE WAY IS BLOCKED. A PLACARD PROCLAIMS, 'SEALED"
        PRINT      "BY ORDER OF HIS MAJESTY KING AZOUN IV.' DO YOU WISH"
        PRINT      "EXIT TO THE WILDERNESS?"
        → NEWECL 50（世界地圖）
```

也就是說：**打完火刀據點之前**，這些邊界格通往據點／城裡；**打完之後**，同一格
改成封條 ＋ 出荒野的詢問。旗標關的是舊路，不是開新路。

## remake 這一側

`gamepack/pack/00-core.json` 的 `tilverton.sewers.first-person.external_exits`
現在逐格宣告這十一格（`event_id: ecl-boundary`，`confidence: exact`）。宣告的
意義只有一個：**這一格踏出去要交給 ECL 判**，而不是照 `wrap: true` 繞回對邊。
落點由腳本寫 `C04B`／`C04C` ＋ `CALL 2E10` 決定，remake 走既有的
`projectDungeonCoordinatesFromView`（spec 1172），不需要在 JSON 裡寫目的地。

## 已被推翻的斷言

**「下水道往火刀據點的交接在 `(8,15,S)`」是錯的。**（原 spec 537 表格標成
`exact`，pack 也照這個座標宣告。）

- GEO 的移動遮罩說 `(8,15)` 的南面**走不出去**（遮罩 `3` ＝ 只有 N、E 通）。
  宣告在那一格的 `external_exit` 會**蓋過牆**——`CanMoveDungeon` 對已宣告的出口
  直接回 true，不看 GEO——所以測試綠，而正常玩家永遠走不到。
- 腳本的兩個方向互為反函數：去程 `X := X − 2`，回程（`ECL2/0x04`）`X := X + 2`。
  spec 537 自己也記著回程落在下水道 `(10,15)`，**去程的來源格因此必須是
  `(10,15)`**。`8` 是把減二套錯方向寫出來的。
- 連帶：據點的落點不是 `(6,1)` 而是 `(8,1)`。

教訓寫成規則：**game pack 的 `external_exit` 會蓋過 GEO 的牆，所以每宣告一格，
都要先用 `cmd/geo-move-mask` 確認那一格那個方向本來就走得出去。** 宣告一格走不出
去的邊界，症狀是「測試綠、玩家走不到」，而這兩件事在報表上分不出來。

## 交叉驗證

- `cmd/geo-move-mask -set 2 -block 3`：十一格的遮罩。
- `cmd/ecl-window -member ECL2.DAX -block 03`：上面的反組譯。
- `internal/game` 的 `TestRealNewGameBeginsAtGlobalBlockOne`：從角色建立一路走到
  `(10,15)` 踏出去，落在據點 `(8,1)`，再走到首領 `(3,13)`。
- 按鍵重放（`cmd/azure-bonds-game` 的 `TestKeysDriveARealSessionFromTheTitle`）：
  走到過的 ECL 段從 `0x01 0x02 0x03` 變成 `0x01 0x02 0x03 0x04 0x50`。
