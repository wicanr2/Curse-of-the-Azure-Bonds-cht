# 796 — Cure Disease 與 Neutralize Poison：神殿服務表補齊兩列

- 證據等級：`exact`（DOS 側逐條讀完；PC-98 側的運算元差異已逐條列出；效果碼表
  取自常駐資料段 dump）
- 作法見 spec 783

## `overlay-04:0324h`（entry#6，75 條，差異 19 條）— Cure Disease

| | DOS | PC-98 |
|---|---|---|
| 「沒生病」 | `'is not Diseased.'`（16，**大寫 D**） | `'は病いに冒されてはいませんよ。'`（30 bytes） |
| 法術名 | `'Cure Disease'`（12） | `'キュア・ディジーズ'`（18 bytes） |

```pascal
有 := 0;
for i := 1 to 6 do
    if <014Dh:00A7h>(@暫存, byte[0DDh + i], DS:6506h) then 有 := 1;
if 有 = 0 then begin
    顯示(名字 ＋ 'is not Diseased.');
    答 := <本模組 0047h>();
end;
if 答 = 'Y' then begin
    顯示('Cure Disease');
    答 := <本模組 00F3h>(3E8h);              { 1000 }
    if 答 = 'Y' then begin
        DS:6F9Ch := 1;
        for i := 1 to 6 do
            <0141h:002Fh>(DS:6506h, byte[0DDh + i], longint(0));
        DS:6F9Ch := 0;
    end;
end;
```

`retf`（無參數）。

### 疾病效果碼表在 `DS:0DDh`（PC-98 `DS:01D9h`）

索引 **1..6**（索引 0 不用），兩平台完全相同：

| i | 1 | 2 | 3 | 4 | 5 | 6 |
|---|---|---|---|---|---|---|
| 碼 | `1Fh`（31） | `22h`（34） | `2Bh`（43） | `2Ch`（44） | `32h`（50） | `39h`（57） |

**六個碼任一存在就算生病，治好時六個一次全清。** 迴圈不會提早結束——就算第一個
就命中，其餘五個照樣查／照樣清。

## `overlay-04:0892h`（entry#9，77 條，差異 23 條）— Neutralize Poison

| | DOS | PC-98 |
|---|---|---|
| 「沒中毒」 | `'is not poisoned.'`（16，**小寫 p**） | `'は毒を受けてはいませんよ'`（24 bytes） |
| 法術名 | `'Neutralize Poison'`（17） | `'ニュートラライズ・ポイズン'`（26 bytes） |

```pascal
if <014Dh:00A7h>(@暫存, 37h, DS:6506h) then 有 := 1;
if 有 = 0 then begin
    顯示(名字 ＋ 'is not poisoned.');
    答 := <本模組 0047h>();
end;
if 答 = 'Y' then begin
    顯示('Neutralize Poison');
    答 := <本模組 00F3h>(3E8h);              { 1000 }
    if 答 = 'Y' then begin
        DS:6F9Ch := 1;
        <0141h:002Fh>(DS:6506h, 37h, longint(0));
        <0141h:002Fh>(DS:6506h, 16h, longint(0));
        <0141h:002Fh>(DS:6506h, 0Fh, longint(0));
        DS:6F9Ch := 0;
    end;
end;
```

`retf`（無參數）。

**檢查與移除不對稱**：只查 `37h`（55），卻一次移除 `37h`、`16h`（22）、`0Fh`（15）
三個。所以身上只有 `16h` 或 `0Fh` 的角色會被判定成「沒中毒」，但玩家硬要付錢
的話仍會被清掉。

## 神殿服務表（累積）

| 服務 | spec | 檢查 | 移除 | 價 |
|---|---|---|---|---|
| 石化解除 | 763 | — | — | `7D0h`（2000） |
| Cure Blindness | 785 | `21h` | `21h` | `3E8h`（1000） |
| Cure Disease | 本支 | `DS:0DDh[1..6]` 任一 | 同六個 | `3E8h`（1000） |
| Neutralize Poison | 本支 | `37h` | `37h` / `16h` / `0Fh` | `3E8h`（1000） |
| Remove Curse | 794 | 物品 `+36h` 或效果 `24h` | 走 `DS:7435h` ＋ `011Ah:004Dh` | `0DACh`（3500） |

五支都是同一個模子（**先問狀態 → 不符再確認 → 帶價錢問一次 → 動手**），
五支都**沒有扣款動作**。

`DS:6F9Ch`（PC-98 `DS:0A036h`）在動手前設 1、做完設 0——只有這兩支移除多個效果
的服務有這個包夾，`overlay-04:0280h`（Cure Blindness）與 `0B31h`（石化）沒有。

### 移除函式的參數順序

`<0141h:002Fh>` 的推入順序是：角色遠指標 → 效果碼 → `longint 0`（`xor ax,ax` ／
`xor dx,dx` 推兩個 word）。先推的是第一個宣告參數，所以簽名是
`(角色, 碼, longint)`。spec 785 把它記成 `(0, 0, 21h, DS:6506h)`（推入順序），
指的是同一件事。

### 中文化：兩句「沒有…」的大小寫不一致

`'is not Diseased.'` 用大寫 `D`、`'is not poisoned.'` 用小寫 `p`，兩句都有句點；
而 spec 785 的 `'is not blind.'`、spec 794 的 `'is not cursed.'` 也各自獨立。
**四句是四個不同的字串常數，逐句替換，不能用樣板湊。**

## 明確不宣稱

- 沒有宣稱那六個疾病碼各自對應什麼狀態。
- 沒有宣稱 `16h` / `0Fh` 是什麼（只確定解毒會一起清掉）。
- 沒有宣稱 `3E8h` 是金幣價格；本支沒有扣款動作。
- 沒有宣稱 `DS:6F9Ch` 由誰讀。
- 沒有宣稱 `DS:0DDh` 索引 0 與索引 7（DOS `30h` / `0Eh`、PC-98 `30h` / `48h`）
  屬於這張表——本支只用 1..6。
