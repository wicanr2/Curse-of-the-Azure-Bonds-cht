# 第五百七十一輪：`TRYTOHIT` 命中判定

狀態：`READY`（限 PC-98 `overlay-23:11C4h` 的控制流與比較式）。
日期：2026-08-14

## 結論先行

`EFFECTS` 單元的 `TRYTOHIT` 是原版的命中判定本體，整段可逐條讀出：

```text
TRYTOHIT(modifier: byte, target: far ptr) : boolean
    result := 0
    DS:A039h := ROLLDICE(1, 20)            ← 一顆 d20，存進全域
    if DS:A039h <= 1  then return 0        ← 自然 1 必失手
    if DS:A039h == 20 then DS:A039h := 100 ← 自然 20 改記為 100
    CHECKFX(10h, target)                   ← 效果可改寫上面那個全域
    if DS:A039h < 0 then return 0
    if sign_extend(DS:A039h) + modifier > target^[19Bh] then result := 1
    return result
```

要點：

- **d20 的結果放在全域 `DS:A039h`，不是區域變數。** `CHECKFX(10h, …)` 在骰完
  之後被呼叫，且後面立刻檢查 `< 0`——所以效果（祝福、詛咒之類）是**直接改寫
  這個全域**來影響命中，而不是回傳修正值。remake 若把骰值放區域變數，
  就沒有地方讓效果掛進去。
- **自然 1 直接失手**，不進入比較。
- **自然 20 被改記為 100**，等於必中（除非效果把它改成負值）。
- 比較是**嚴格大於**：`骰值 + modifier > target^[19Bh]`，相等算失手。
- `target^[19Bh]` 是目標 record 內偏移 `19Bh` 的 byte，本輪不宣稱它是 AC
  或其他具名欄位——那需要另外證明 writer 與 consumer。

`ROLLDICE(1, 20)` 的參數順序依 [spec 568](568-rolldice-and-original-rng-entry.md)：
先推 count 再推 sides。

## 這份規格明確不宣稱

- **`DS:A039h` 的完整生命週期**：誰還會讀它、戰鬥流程何時清除。
- **`CHECKFX(10h, …)` 做了什麼**：`10h` 是效果類別或遮罩，本輪只確認它在
  骰後、比較前被呼叫，且能改變骰值。`CHECKFX` 本體（`overlay-23:03FEh`）未讀。
- **`target^[19Bh]` 的欄位語意**。
- **DOS 側**：`overlay-23` 的對應函式尚未逐指令核對，本輪只讀 PC-98。
- **`ATTEMPTTOHIT`（`122Ch`）**：緊接其後、參數更多（`retf` 更大），未讀。
