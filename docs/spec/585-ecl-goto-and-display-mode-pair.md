# 第五百八十五輪：ECL `GOTO` 與顯示模式的成對呼叫

狀態：`READY`。等級：`exact`。日期：2026-08-14

## opcode `01h` ＝ `GOTO`

兩平台都在 `overlay-02:00E8h`，31 bytes。

```text
GOTO:
    READVAR(1)
    ECL_PC := ADDFNC(high[1], low[1])       ← operand 的值直接當新 PC
```

| | PC-98 | DOS |
|---|---|---|
| operand high[1] | `DS:A958h` | `DS:76C6h` |
| operand low[1] | `DS:A998h` | `DS:7706h` |
| ECL PC | `DS:7F21h` | `DS:4FB4h` |

**跳躍目標是 script 內的絕對位移，沒有加任何 base。**

### DOS 的 ECL PC 升級為 `exact`

先前 `DS:4FB4h` 是 `strong inference`（靠 `3Ah` handler 的同形推得，
見 `docs/knowledge/gold-box-ecl-interpreter.md`）。現在兩平台的 `GOTO` handler
在**同一個 overlay-local 位址、同一條指令序列**上分別寫入 `DS:7F21h` 與
`DS:4FB4h`，而前者已是 `exact`——對應關係因此確定，不再是推論。

## 顯示模式的成對呼叫

`dos/overlay-18:006Ah`、`dos/overlay-32:00F4h`、`dos/overlay-32:0134h` 是同一
形狀：

```text
case DS:4FE6h of
    1: <far 0297:2171>(a, b) …
    2: <far 0297:21B0>(a, b) …
```

`DS:4FE6h` 只取 `1` 或 `2`，兩個分支呼叫的是 `0297:2171` 與 `0297:21B0` 這**一
對**常式，參數相同。形狀是「兩種顯示模式各有一套繪圖進入點」。

⚠ 兩個分支的參數順序不一定一致：`overlay-32:00F4h` 的模式 1 傳 `(8, 0)`，
模式 2 傳 `(0, 8)` 之後又傳 `(8, 0)`。照抄，不要以為是對稱的。

## 其他

- `dos/overlay-12:2855h`：連續四次 `sub_1B(0Bh)`／`(35h)`／`(34h)`／`(37h)`，
  之後 `if DS:6FA7h = 0 then DS:6F92h := 64h`。
- `dos/overlay-24:18BEh`：依條件呼叫 `loc_1ECC+1`，一邊傳 `(15h, …)`、
  另一邊傳 `(16h, 26h, 12h, 1)`。

## 明確不宣稱

- `DS:4FE6h` 的兩個值分別代表哪種顯示模式。
- `0297:2171`／`0297:21B0` 的本體。
- `sub_1B` 的本體與那四個 id 的意義。
