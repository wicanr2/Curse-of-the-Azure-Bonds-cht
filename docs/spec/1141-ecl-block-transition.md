# 1141 — ECL block 怎麼換：主迴圈讀 `LastECL`，`NEWECL` 改它

- 證據等級：`exact`（`overlay-02:3772` 的分派段逐條讀完；轉移圖由
  `cmd/ecl-block-graph` 對整份 image 重算）／**明確不宣稱**（`006B:0034` 的內部）
- 上游 spec 1095（ECL 主迴圈與戰鬥續跑）、spec 1096（bank 映射）、spec 1104

## 主迴圈怎麼決定載哪一個 block

`overlay-02:3772`（679 bytes，overlay 進入點）開頭那一段：

```pascal
if bank0^[1E4h] = 0 then begin                 { LastECL，新遊戲是 0 }
    DS:8B6Eh := 0;
    if DS:4FBCh <> 0 then DS:8B5Eh := 52h      { 開場 }
    else begin DS:8B5Eh := 1; <far 014D:002A>(DS:6506h) end;
end else
    DS:8B5Eh := bank0^[1E4h];                  { ★ 就是 LastECL }

if bank0^[1CCh] = 0 then DS:4FBAh := 3;        { InDungeon 為 0 時的模式 }
…
<far 006B:0034>(DS:8B5Eh);                     { 帶著 block 編號進去 }
…
bank0^[1E4h] := DS:8B5Eh;                      { ★ 載完寫回去 }
```

⇒ **`bank0 +1E4h`（ECL 位址 `4BF2h`，spec 1096 的 bank 映射）就是「下一個要載的
block」**。主迴圈讀它、載入、再寫回去。

## 兩件容易搞反的事

★ **`SAVE <自己的編號> → 4BF2h` 不是轉移**，是「記錄我是誰」。六個 block
（`0x50`、`0x51`、`0x23`、`0x25`、`0x33`、`0x43`）寫的都是自己的編號。

★ **`+342h`（City，ECL `4CA1h`）不是拿來載區域的**。它唯一的引擎側讀取點是
`overlay-10:0FC7`，那一支把它抄進 `DS:4A14h` 之後接的是三支**野外戰場地形生成**
（spec 749 的障礙、植被、地貌密度）。世界地圖 hub 用 `GETTABLE` 填它，
是為了讓那一格的野外戰鬥長得像那個地方，不是為了換地圖。

## 轉移是 `20h NEWECL` 做的

改寫目的地的是 `NEWECL`。完整的圖見
[`docs/audit/ecl-block-graph.md`](../audit/ecl-block-graph.md)（由
`cmd/ecl-block-graph` 產生）：**25 個 block、47 條出邊、只有兩個 block 沒有出邊**
——`ECL1/0x52`（開場）與 `ECL6/0x43`（結局）。

兩個世界地圖 hub 是樞紐：

| hub | 出邊 |
|---|---|
| `ECL1/0x50` | `0x03`、`0x25`、`0x31`、`0x35`、`0x40`、`0x51` |
| `ECL1/0x51` | `0x10`、`0x11`、`0x15`、`0x20`、`0x45`、`0x50` |

## ⚠ 事件目錄的圖是假零，不要拿來判斷「沒有出口」

`docs/audit/ecl-event-catalog.md` 的可達性**不跟 `ON GOTO`／`ON GOSUB` 的目的地**
（它自己的「限制」那一節寫了）。用那一份算出來的是 **21 條邊**，而且兩個 hub
會看起來**一條出邊都沒有**——那是整條主線的樞紐。

差距也看得出來：同一個 block，事件目錄看到 85 條指令，跟 `ON GOTO` 之後是 726 條。
兩份的分母本來就不同（4,222 對 14,177），不是同一件事的兩個版本。

判斷「某個 block 有沒有出口」一律用 `cmd/ecl-block-graph`。

## 明確不宣稱

- 沒有宣稱 `<far 006B:0034>` 內部做什麼；只宣稱它拿到的是 block 編號，
  而且呼叫前後 `+1E4h` 的讀寫構成「讀 → 載 → 寫回」。
- 沒有宣稱 `DS:4FBCh`（開場旗標）、`DS:4FBAh`（模式）、`DS:8B6Eh` 由誰設。
- 沒有宣稱 47 條邊在遊戲裡**都走得到**：靜態可達不等於劇情可達，
  條件分支要靠實機路徑驗。
