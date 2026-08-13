# 第六百五十二輪：換片重試迴圈，與第三份首位元組判斷

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `PC98-GAME.EXE` 的 `173A1h`、`177BDh`、`1790Dh`、`17A37h`。

## `173A1h`：開不到 `game.ovr` 就一直要求換片

```text
repeat
    s := 'game.ovr'                          ← 每一輪都重新指派
    ok := <sub_17656>(s)
    if ok = 0 then
        <sub_1BDDB>(0, CS:byte_17380, DS:byte_2820Eh)
        <sub_1BD13>()
        <sub_18036>()
until ok <> 0
```

`byte_17380` 是一個 32 bytes 的 Pascal 短字串：

> **ゲームディスクＡを入れてください**（請放入遊戲磁片 A）

所以這是**換片提示的重試迴圈**：開不到 `game.ovr` 就顯示提示、等一下，然後**整個
重來**——包括重新指派檔名字串。沒有重試上限，也沒有放棄的出口。

`sub_18036` 的回傳值存進區域變數後沒有再讀（本輪第四次看到這個模式）。

### 對 remake 的意含

remake 沒有換片，這條迴圈整段不需要。但**訊息本身要保留在中文化的字串表裡**——
玩家路徑驗證若要重現原版流程，這是其中一個畫面。

## `177BDh`：第三份首位元組判斷

```text
回傳 1 的條件：81h..9Fh 或 0E0h..0FCh
```

與 `14342h`（[spec 645](645-pc98-text-layer-primitives.md)）同一族——**上界 `FCh`**、
回傳值放 `AL`（0／1），不是進位旗標。

所以 resident 裡首位元組判斷共有兩族三支：

| 族 | 上界 | 回傳 | 位址 |
|---|---|---|---|
| A | `FCh` | `AL` = 0／1 | `14342h`、`177BDh` |
| B | `F7h` | `clc` = 前導 | `169D9h` |

實際畫字走的是 B 族（[spec 648](648-pc98-text-draw-core.md)）。A 族由誰使用本輪
未追。

## `1790Dh`：`0FFh` 當標記的二選一

```text
if arg_C^^[0] = 0FFh then
    <sub_1C15D>(arg_4^, arg_8^^, arg_C^^ + 1)      ← 跳過標記那個 byte
else
    <sub_17DD5>(arg_4^, arg_8^^（dword）, arg_C^^（dword）)
```

兩條路徑傳的**參數形狀不同**：命中 `0FFh` 時第二、三個參數是 far pointer 的內容
（各一個 word ＋ 位址加一），否則是完整的兩個 dword。所以 `0FFh` 是「短格式」的
標記。

三個參數都是**兩層間接**（`les di, [bp+arg]` 之後再 `les di, es:[di]`），呼叫端傳
的是指標的指標。

## `17A37h`

```text
<sub_1BF69>(arg_4)
<sub_1BF69>(arg_0)
```

同一支 routine 用兩個參數各叫一次，順序是 `arg_4` 先、`arg_0` 後。

## 明確不宣稱

- `sub_17656`（開檔）／`sub_1BDDB`／`sub_1BD13`／`sub_18036`／`sub_1BF69`／
  `sub_1C15D`／`sub_17DD5` 的行為。
- `DS:byte_2820Eh` 的用途。
- A 族首位元組判斷由哪些呼叫端使用。
