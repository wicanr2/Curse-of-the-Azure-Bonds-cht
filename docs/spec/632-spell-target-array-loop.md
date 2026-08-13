# 第六百三十二輪：對整個目標陣列施放 —— `for i := 1 to DS:0A520h` 的固定寫法

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-22:4A9Ah`（146 bytes）、`496Bh`（194 bytes）。

## 迴圈寫法

```asm
mov  al, ds:0A520h          ; 數量
mov  [bp+var_B], al
mov  al, 1                  ; 由 1 開始
cmp  al, [bp+var_B]
ja   短：整段跳過            ; 數量為 0 就不進迴圈
...
mov  di, ax
shl  di, 1
shl  di, 1                  ; 索引 × 4
push word ptr [di-5AE1h]    ; 陣列[i] 的 segment
push word ptr [di-5AE3h]    ; 陣列[i] 的 offset
```

`-5AE3h` 當無號是 `0A51Dh`——[spec 630](630-spell-target-array.md) 的目標陣列。
兩支各自獨立走這個迴圈，是該 spec 的第二、三個佐證。

**數量為 0 時整段跳過**（`cmp 1, n` ＋ `ja`），不會誤跑一輪。迴圈尾是
`cmp i, n` ＋ `jnz`，所以 `i` 走完 `n` 才結束。

## `4A9Ah`：先印訊息再逐個套用

```text
備妥 'は魅了された。'，<sub_1D0B>(0, 訊息)      ← 迴圈之外，只印一次
for i := 1 to DS:0A520h do
    if <far 014Ah:00A7h>(0Bh, 陣列[i], @var) <> 0 then
        <far 013Eh:002Ah>(0Bh, 陣列[i], var, 0)
```

訊息**在迴圈外面**，效果在裡面——不管命中幾個目標，訊息只出現一次。

效果 id `0Bh` 與 `014Ah:00A7h` → `013Eh:002Ah` 的接法，跟
[spec 630](630-spell-target-array.md) 的 `2147h` 完全一樣，差別只在那支處理單一
目標、這支跑整個陣列。

## `496Bh`：先看 bank 0 的旗標決定要不要做

```text
if bank0^[1CCh] <> 0 then return                ← 直接結束
for i := 1 to DS:0A520h do
    if 陣列[i] <> nil then                       ← 逐筆檢查 nil
        t := 陣列[i]
        v := <far 013Eh:0048h>(0, 4, t)
        x := <sub_E75>(t, 88h, 88h)
        備妥 'は絡みつかれた。'
        <far 013Eh:0089h>(v, 1, 0, 0, x, 訊息)
```

這支**每一筆都檢查 nil**，`4A9Ah` 則不檢查。陣列裡可能有 nil——
[spec 629](629-spell-pack-idiom-and-uninit.md) 的 `2776h` 就會把 `DS:0A521h`
（第 1 筆）清成 nil。所以不檢查的那支在該情況下會把 nil 傳下去。

`bank0^[1CCh]` 這個旗標在 [spec 626](626-goduel-and-charrec-size.md) 的 `068Bh`
也被讀（那裡是決定要不要載入 `SPRIT`）。同一個旗標控制兩件不同的事。

## 明確不宣稱

- `bank0^[1CCh]` 代表什麼。
- 效果 id `0Bh`、常數 `88h`、`013Eh:0048h`／`013Eh:0089h`／`sub_E75` 的語意。
- 這兩支各自對應哪一個法術名稱。
