# 第六百八十六輪：DOS ECL 的運算元描述表與選單 opcode

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：DOS `overlay-02` 的 `092Ch`、`0B3Bh`、`0EBDh`。

## 三個平行的運算元描述陣列

把三支函式對同一組位址的用法排在一起，索引關係就出來了：

| 函式 | 用到的位址 | 運算元序號 |
|---|---|---|
| `011Eh`（spec 685） | `DS:7686h`、`DS:7687h` | 1、2 |
| `0EBDh` | `DS:76C6h`、`DS:7706h` | 1 |
| `092Ch` | `DS:76C7h`、`DS:7707h` | 2 |

三個陣列**首位址相差 40h**（`7685h` → `76C5h` → `7705h`），各 64 bytes，
**1-based**。取值一律成對：

```text
addr  := ADDFNC(DS:7705h[k], DS:76C5h[k])      ← overlay-07 entry#9
STOREVALUE(addr, value)                        ← overlay-07 entry#15
```

`ADDFNC` 在 PC-98 側已讀完：**不是算術加法**，是把兩個 byte 併成一個 word，
回傳 `(第二參數 shl 8) + 第一參數`。所以

> **`DS:76C5h[k]` 是運算元位址的高位元組，`DS:7705h[k]` 是低位元組。**

寫回的參數順序是 **`(addr, value)`**——位址在前、值在後，`0EBDh` 與 `092Ch`
兩支一致，也和 PC-98 側 `STOREVALUE(addr, value)` 的記法一致。

`DS:7685h[k]` 則是型別標記（`>= 80h` 代表字串，見 spec 685）。三個陣列在 PC-98
的對應是 `A917h`／`A957h`／`A997h`，同樣相差 40h、同樣 1-based——**兩平台各自
獨立讀出來的結論一致**。

## 字串槽：`DS:7648h + k×100h`

`0EBDh` 的節點填充迴圈算出來的來源位址是 `(index shl 8) + 7648h`，index 從 1
起跳。所以字串槽是 256 bytes 一格、以 `7648h` 為基底：槽 1 = `7748h`、
槽 2 = `7848h`——正好就是 spec 685 裡 `011Eh` 字串比較用的那兩個緩衝。
**兩支函式各自算出的位址在這裡對上了**，不是猜的。

## `0EBDh`：選單／清單選取

```text
DS:8B66h := 0
READVAR(3)                                   ← 先取 3 個運算元
dst := ADDFNC(DS:7706h, DS:76C6h)                ← 結果要寫回運算元 1 指到的位址
title := 複製 DS:7748h（上限 0FFh）
n := ADDRESSVALUE(3)                             ← 運算元 3 的值 = 項目數
DS:4FB4h := DS:4FB4h − 1
READVAR(n)                                   ← ⚠ 再取 n 個運算元，數量由資料決定
list := <overlay-26 entry#7>(n)                           ← 配置 n 個節點
head := list
DS:65A0h := 1；DS:65A1h := 11h
<far 0542h:04A4h>(1, 11h, 26h, 16h, 0Ah, 0, 1, @title)   ← 開視窗
i := 1
while list <> NIL do
    複製 40 bytes：DS:[7648h + i shl 8] → list^                ← 項目文字
    list^[29h] := 0
    list := list^[2Ah]                                        ← 下一個節點
    i := i + 1
list := head
choice := 0
<overlay-07 entry#21>  ; ECLMENUV(@list, @choice, @flag, 0, list, 16h, 26h, DS:65A1h+1, 1, 0Fh, 0Ah, 0Dh, @x, @x)
STOREVALUE(dst, choice)                             ← 寫回運算元 1
<overlay-26 entry#8>(@list)                               ← 釋放清單
<far 0542h:0B4Ah>(16h, 26h, 11h, 1)              ← 還原視窗底下的畫面
```

**取運算元的次數不固定**：先固定取 3 個，讀出第 3 個的值當數量，再取那麼多個。
所以這條 opcode 的位元組長度要跑過才知道——靜態切 opcode 邊界時會在這裡切錯。

節點版面（由 `28h` 的複製上限與兩個常數位移確定）：

| 位移 | 內容 |
|---|---|
| `+00h` | Pascal 字串，長度位元組 ＋ 最多 40 字元 |
| `+29h` | 旗標 byte，建立時填 0 |
| `+2Ah` | far 指標，指向下一個節點 |

迴圈終止是**同時檢查 far 指標的兩個 word 是否為零**（`mov ax,[..]; or ax,[..]; jz`），
不是只看 segment。

## `092Ch`：讀一個數值寫回運算元 2

```text
READVAR(2)
buf := 0
dst := ADDFNC(DS:7707h, DS:76C7h)
value := <far 0542h:08B3h>(0, 0Ah, @buf)
STOREVALUE(dst, value)
```

`0Ah` 是輸入長度上限。

## `0B3Bh`：條件不成立就跳過下一條 ECL 指令

```text
DS:4FB4h := DS:4FB4h + 1
v := DS:75FFh
if v 在 16h..1Bh 之間 且 DS:[75F8h + (v − 16h)] = 0 then
    <overlay-07 entry#29>()      ← 六個分支呼叫的是同一支，且無參數
```

`DS:75F8h..75FDh` 這六個 byte 就是 spec 685 裡兩支比較程序（`overlay-07`
entry#22／#23）寫出的六個結果旗標：`=`、`<>`、`<`、`>`、`<=`、`>=`。
`overlay-07` entry#29 在 PC-98 側已讀完，是**跳過下一條 ECL 指令**——它從
bank3 讀出下一個 opcode，查內建的 opcode→arity 表，呼叫 `READVAR(n)` 把該指令
的運算元吃掉。

所以整條的語意是：**opcode `16h..1Bh` 各對應一種比較，對應的旗標為 0（條件不
成立）就跳過下一條指令**。這是 ECL 的條件分支。

編譯後展開成六段 `cmp`／`jnz`，六個分支的目標位址完全相同，差別只在讀哪一個
旗標。`v` 落在 `16h..1Bh` 之外就什麼都不做（不跳，也不執行）。

## 明確不宣稱

- `0542h:04A4h`／`0542h:08B3h`／`0542h:0B4Ah` 這幾支 resident 的內部行為。
- `DS:4FB4h` 數的是什麼（各 opcode 有增有減）。
- `DS:65A0h`／`65A1h` 的單位。
- `DS:75F8h..75FDh` 這六個旗標代表什麼。
- 槽 1 同時當標題又當第一個項目，是不是原始碼的本意。
