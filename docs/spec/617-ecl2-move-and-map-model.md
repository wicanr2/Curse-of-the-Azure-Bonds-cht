# 第六百一十七輪：`MOVEFORWARD` 與 16×16 環繞地圖

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-07:1AECh`（145 bytes）。

```text
MOVEFORWARD:
    case DS:A2ABh of                          ← 方向
        0: if DS:A2AAh > 0    then DS:A2AAh := DS:A2AAh − 1 else DS:A2AAh := 0Fh
        2: if DS:A2A9h < 0Fh  then DS:A2A9h := DS:A2A9h + 1 else DS:A2A9h := 0
        4: if DS:A2AAh < 0Fh  then DS:A2AAh := DS:A2AAh + 1 else DS:A2AAh := 0
        6: if DS:A2A9h > 0    then DS:A2A9h := DS:A2A9h − 1 else DS:A2A9h := 0Fh
    DS:A2ADh := <far 017C:003E>(DS:A2AAh, DS:A2A9h)
    DS:A2ACh := <far 017C:0034>(DS:A2ABh, DS:A2AAh, DS:A2A9h)
    DS:BDFAh := 1
```

## 地圖是 16×16，而且會環繞

座標範圍 `0..0Fh`，**走出邊界不是被擋住而是繞到另一端**：`y = 0` 再往北會變成
`y = 0Fh`。四個方向都有對應的環繞分支，沒有例外。

remake 若在邊界擋住玩家，走法會與原版不同。

## 方向編碼對上了 spec 563

| | ECL 看到 | 實體 `DS:A2ABh` |
|---|---:|---:|
| 北 | 0 | `0` |
| 東 | 1 | `2` |
| 南 | 2 | `4` |
| 西 | 3 | `6` |

[spec 563](563-ecl-memory-model-and-operand-resolution.md) 從讀寫路徑解出
`C04Dh` **讀取時 ÷2、寫入時 ×2**，當時只知道「round-trip 一致，係數只有繪圖端
看得到」。這一輪從移動邏輯確認了另一半：**實體值就是 `0`／`2`／`4`／`6`**，
`MAXRANGE`（[spec 615](615-ecl2-findguy-maxrange.md)）的方向編碼也是同一組。

三個虛擬地圖暫存器的身分因此確定：

| ECL 位址 | 實體 | 內容 |
|---|---|---|
| `C04Bh` | `DS:A2A9h` | **x** |
| `C04Ch` | `DS:A2AAh` | **y** |
| `C04Dh` | `DS:A2ABh` | **方向**（ECL 0..3，實體 0/2/4/6） |

移動後還會更新 `DS:A2ADh`（由 x, y 算）與 `DS:A2ACh`（由方向, x, y 算），
並把 `DS:BDFAh` 設為 1。

## `GETMONSTERS`（`0556h`）

```text
out^ := nil
New(out^, 1A7h)                               ← 423 bytes，與 0Bh 同一種記錄
<載入>(id, out^)
<處理 out^^[14Eh] 的物品鏈>
```

記錄大小與 `0Bh`（[spec 604](604-ecl-spawn-monsters.md)）一致，再次確認
**角色／怪物記錄是 423 bytes**。

## 明確不宣稱

- `017C:003E`（由座標算）與 `017C:0034`（由方向與座標算，`MAXRANGE` 也用它做
  可通行判定）分別回傳什麼。
- `DS:BDFAh` 的語意。
