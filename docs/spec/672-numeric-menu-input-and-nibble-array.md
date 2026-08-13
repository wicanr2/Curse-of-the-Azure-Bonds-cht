# 第六百七十二輪：數字選單輸入與 nibble 陣列讀取

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：DOS `START.EXE` 的 `12110h`、`129EBh`。

## `12110h`：讀到合法為止

```text
重複:
    c := <ReadKey>()
    byte_20E7Eh := c
    if c 不在 CS:byte_120F0h 的集合裡 then 重來      ← 第一道
    n := c − 30h                                     ← ASCII 數字 → 數值
    清空區域集合
    區域集合 := [arg_0 .. arg_2]                     ← Pascal 的 `+` 建子範圍
    if n 不在區域集合裡 then 重來                     ← 第二道
<Write>(Text, c) ; <Write>(Text)                     ← 通過才回顯
return n
```

**兩道檢查都不通過就回到 `ReadKey`**——按錯鍵完全沒有回饋，畫面上什麼都不會出現。
回顯在兩道檢查之後，所以**不合法的按鍵永遠不會顯示出來**。

允許的字元集合寫死在程式碼段（`CS:byte_120F0h`），而數值範圍由參數
`arg_0`／`arg_2` 決定——兩者各管一層。

IDA 由 Borland 簽章還原出 `@READKEY$qv`、`@Set@MemberOf$q4Byte`、
`@Set@Clear$qv`、`@Set@$brplu$q4Bytet1`、`@Write$qm4Text4Char4Word`，所以這裡用到的
都是 Turbo Pascal 的集合運算與 `Crt` 單元。

## `129EBh`：從 8 bytes 裡取第 N 個 nibble

```text
複製 arg_0 指的 8 bytes 進區域                       ← @$basg（block assign）
if arg_4 是奇數 then
    return 區域[arg_4 div 2] and 0Fh                 ← 低 nibble
else
    return (區域[arg_4 div 2] and 0F0h) shr 4        ← 高 nibble
```

8 bytes ＝ **16 個 nibble**，`arg_4` 是 `0..15` 的索引。奇偶判斷用
`shr al, 1` 後看進位——順手把商也算出來，但下面又用 `idiv 2` 重算一次，**同一個
除法做了兩遍**。

**高 nibble 在前**：索引 0 取高半、索引 1 取低半。remake 解這種打包資料時順序不能
弄反。

除法用的是**有號** `idiv`，但索引本來就非負，所以結果相同。

## 明確不宣稱

- `CS:byte_120F0h` 集合的實際內容。
- 這兩支各自服務哪個畫面。
