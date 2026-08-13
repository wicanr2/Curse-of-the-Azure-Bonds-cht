# 第六百八十八輪：DOS `overlay-02` 的四支場景 opcode

狀態：`READY`。等級：`exact`。日期：2026-08-14
工具：`scripts/annotated_dump.py`（把 far call 解回模組 entry 再貼上對面平台的判讀）

## `2D15h`

```text
DS:4FB4h := DS:4FB4h + 1
<resident 01A0:0136>()
<overlay-24 entry#2>(DS:6508h, DS:6506h)        ← 推入 seg 再 off，傳的是 DS:6506h 那個 far 指標
<overlay-24 entry#38>()
<overlay-29 entry#1>(3, 3, 1, DS:7234h, DS:7232h)
<overlay-24 entry#38>()
DS:8B6Eh := 0
DS:8B60h := 1
```

`DS:6506h`（低位）與 `DS:6508h`（高位）合起來是一個 far 指標——`3251h` 用
`les ax, ds:6506h` 讀它，證實這一點。

## `2D5Eh`

```text
READVAR(1)
<far 0542h:0B4Ah>(1, 11h, 26h, 16h)             ← 還原視窗，與 0EBDh 收尾同一支同一組參數
複製 DS:7748h → 區域緩衝（上限 0FFh）            ← 字串槽 1（spec 686）
<overlay-24 entry#40>(@緩衝, 0, @DS:6506h)
```

第三個參數推的是 `push ds; push di`（`di = 6506h`），也就是**指標的位址**，
不是指標的值。

## `321Fh`

```text
FillChar(DS:8B48h, 2, 0)                        ← 0A54h:1AE0h
DS:8B62h := 0
READVAR(1)
<overlay-03 entry#1>()
<overlay-24 entry#37>()
```

`0A54h:1AE0h` 三個參數是 `(@目標, 2, 0)`，`179Ah` 收尾也用同一組——形狀與
Turbo Pascal 的 `FillChar(var, count, value)` 一致。

`sub sp, 8` 配置的區域空間整支沒有用到。

## `3251h`

```text
DS:4FB4h := DS:4FB4h + 1
<overlay-17 entry#3>(0, 1)
DS:47E7h := DS:6506h                            ← far 指標整個複製過去
DS:47E9h := DS:6508h
<overlay-24 entry#2>(DS:6508h, DS:6506h)
```

`sub sp, 2` 同樣沒有用到。

## 明確不宣稱

- `overlay-24` entry#2／#37／#38／#40、`overlay-29` entry#1、`overlay-03`
  entry#1、`overlay-17` entry#3 各自做什麼——兩個平台的台帳都還沒有這些。
- `resident 01A0:0136` 的身分。
- `DS:6506h` 指向什麼、`DS:47E7h` 這份複本給誰用。
- `DS:8B60h`／`8B62h`／`8B6Eh`／`7232h`／`7234h` 的語意。
