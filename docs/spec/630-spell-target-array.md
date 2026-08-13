# 第六百三十輪：選定目標陣列 `DS:0A51Dh` —— 所有法術用的 `DS:0A521h` 是它的第 1 筆

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-22:282Dh`（175 bytes）、`2147h`（227 bytes）。

## 目標陣列

`282Dh` 掃過角色鏈挑目標，存進一個陣列：

```asm
inc  byte ptr ds:0A520h        ; 先加一，所以索引由 1 開始
les  cx, [bp+var_4]            ; 這一筆的 far pointer
mov  bx, es
mov  al, ds:0A520h
xor  ah, ah
mov  di, ax
shl  di, 1
shl  di, 1                     ; di := 索引 × 4
mov  [di-5AE3h], cx            ; -5AE3h 當無號是 0A51Dh
mov  [di-5AE1h], bx
```

| 位址 | 內容 |
|---|---|
| `DS:0A51Dh` | 陣列基底，每筆 4 bytes（far pointer） |
| `DS:0A520h` | **已選數量**（同時是索引上界） |
| `DS:0A521h` | 第 1 筆目標 ＝ `0A51Dh + 4 × 1` |

**這解釋了先前每一支法術函式都在讀的 `DS:0A521h`／`DS:0A523h`**（spec 627／628／
629）——它不是某個特別的全域變數，是**選定目標陣列的第 1 筆**。

索引由 1 開始，所以第 0 筆（`0A51Dh`..`0A520h`）永遠不會被寫；計數 byte 就放在那
四個 bytes 的最後一個。兩者位置重疊但不會互相覆蓋。

## `282Dh` 的挑選規則：預算耗盡為止

```text
budget := DS:9594h^[1A5h]        ← 施法者的 +1A5h，存進 DS:7D93h
DS:0A520h := 0
p := DS:9598h                     ← 角色鏈
while p <> nil do
    if p^[11Ah] = 0Eh then                    ← RACETYPE
        if budget >= p^[1A5h] then
            budget := budget − p^[1A5h]       ← 扣掉這一個的份
            inc DS:0A520h
            陣列[DS:0A520h] := p
    p := p^[18Ah]
推入 DS:0A031h 與兩個 0，備妥 'は魅了された。'，<sub_F62>
```

**逐個扣預算、扣不動就跳過**，而不是固定選 N 個。`+11Ah` 是 `RACETYPE`
（[spec 499](499-pc98-alignment-conditional-effects.md) 由 Borland type table 定
出），這裡只挑 `0Eh` 那一種。

比較與扣減都用 `+1A5h`（目前 HP，[spec 623](623-killthedude-damage-message.md)）。
判斷是 `budget >= p^[1A5h]`（無號），**扣不動的不會被部分影響**。

注意迴圈**不會提早結束**——預算歸零後仍走完整條鏈，只是後面都通不過檢查。

## `2147h`：免疫檢查與效果套用

```text
t := DS:0A521h
if t^[11Ah] > 1 or t^[0DEh] > 1 then
    備妥 'は影響を受けなかった。'，<far 014Ah:0084h>(1, 0Ah)
else
    x := <far 014Ah:00D4h>(DS:0A031h, DS:0A031h)
    n := (signext(DS:9594h^[198h]) shl 7) + x
    推入 n、1、0、0，備妥 'は魅了された。'，<sub_F62>
    if <far 014Ah:00A7h>(t, 0Bh, @var_4) <> 0 then
        <far 013Eh:002Ah>(0, var_4, t, 0Bh)
```

免疫條件是**兩個欄位任一超過 1**：`+11Ah`（`RACETYPE`）與 `+0DEh`。不是等於某個
值，是大於 1——所以「種族編號 0 或 1」才會被影響。

`014Ah:00A7h(效果 id, 目標, @out)` 之後接 `013Eh:002Ah` 的寫法，與
[spec 629](629-spell-pack-idiom-and-uninit.md) 的 `2776h` 相同（那邊 id 是 `37h`，
這邊是 `0Bh`）。

## 明確不宣稱

- `DS:7D93h`／`DS:7D94h`／`DS:0A031h`／`+0DEh`／`+198h` 的身分。
- `RACETYPE` 的 `0Eh` 與「大於 1」各自對應哪些種族。
- `014Ah:00A7h`／`013Eh:002Ah`／`014Ah:0084h`／`014Ah:00D4h`／`sub_F62` 的內部行為。
- 這兩支各自對應哪一個法術名稱。
