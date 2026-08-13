# 第五百九十五輪：`DS:9594h` 是目前目標

狀態：`READY`。等級：`exact`。日期：2026-08-14
模組：`INTERPET`（overlay-02）。位址為 PC-98 overlay-local。

## `39h`（`2EDAh`）：選目標

```text
READVAR(1)
<far 0418:14C7>(16h, 26h, 11h, 1)         ← 開視窗（座標與 11h/12h 那組相同）
s := <解 packed text>(DS:A9DAh)
p := DS:9594h
p := <far 014A:00E8>(@p, 0, s)            ← 以 s 為提示讓玩家選一個
if p = nil then <opcode 00h 的 handler>   ← 取消 ⇒ 結束整個 script
else DS:9594h := p                        ← 換成選中的
```

**取消選單會直接結束 script**（走 `00h`，連帶清空 GOSUB 堆疊，
[spec 588](588-ecl-return-and-gosub-stack.md)）。remake 把取消當成 no-op
會讓後續指令繼續跑，行為完全不同。

## `DS:9594h` ＝ 目前目標的 far pointer

這支把選中的結果寫回 `DS:9594h`，而同一個位址在別處被當成「要操作的對象」
傳出去：

| 出處 | 用法 |
|---|---|
| `KILLDUDE` 尾段 | `<far 014A:002A>(DS:9594h)`（[spec 579](579-character-status-fields.md)） |
| `sub_269` | 遍歷 effect 鏈時的 subject（[spec 577](577-attempttohit-and-effect-chain-walk.md)） |
| `3Fh`（下方） | 查詢效果的對象 |
| `3Dh` | 重繪時傳出去（[spec 594](594-ecl-random-and-indexed-store.md)） |

先前 `DS:9594h` 只知道「是個會被傳來傳去的 far pointer」，現在確定它就是
**ECL 當前操作的目標**，而且 `39h` 是唯一（目前已知）會改寫它的指令。

## `3Fh`（`364Bh`）：查詢目標身上有沒有某效果

```text
FillChar(DS:A88Ah, 6, 0)                  ← 清六個比較旗標
READVAR(1)
id := ADDRESSVALUE(1)
if <far 014A:00A7>(@ctx, id, DS:9594h) then DS:A88Ah := 1
else                                        DS:A88Bh := 1
```

`014A:00A7` 就是 `sub_269` 用的那支效果查找常式。

**它借用比較旗標**，所以後面接 `16h`（相等）就是「有這個效果」、接 `17h`
就是「沒有」。ECL 沒有獨立的布林分支指令，條件查詢一律走同一組旗標。
只設 `A88Ah`／`A88Bh` 兩個，所以 `18h`～`1Bh` 接在它後面永遠跳過。

## 明確不宣稱

- `0418:14C7` 與 `014A:00E8` 的本體（視窗與選單）。
- `39h` 的 operand 除了決定提示文字之外還有沒有別的作用。
