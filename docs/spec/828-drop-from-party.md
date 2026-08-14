# 828 — 把角色踢出隊伍；最後一個成員被踢掉就直接離開遊戲

- 證據等級：`exact`（兩平台各自逐條讀完）
- 作法見 spec 783

## `overlay-15:199Ch`（entry#20，DOS 136 條／PC-98 160 條）

| | DOS | PC-98 |
|---|---|---|
| 離開遊戲確認 | `'quit TO DOS: '`（13，**結尾一個空白**） | （該段被改寫，無對應） |
| 預告 | `'will be gone'`（12） | `'は完全にいなくなる。'`（20 bytes） |
| 提問 | `'Drop from party? '`（17，**結尾一個空白**） | `'かまわず離脱させますか？'`（24 bytes） |
| 好聚好散 | `'bids you farewell'`（17） | `'は別れを告げた。'`（16 bytes） |
| 丟進水溝 | `'is dumped in a ditch'`（20） | `'はドブに捨てられた。'`（20 bytes） |
| 反悔 | `'Breathes A sigh of relief'`（25，**A 是大寫**） | `'は安堵のため息をもらした。'`（26 bytes） |

`retf`（無參數）。

```pascal
if 只剩這一個 then begin
    顯示('quit TO DOS: ');
    if <far 16AEh>(0Eh, 0Ah, 0Fh) = 'Y' then begin
        <far 0F0Fh>(1, 0);
        <far 06EAh:0000h>();                    { 直接離開遊戲，不返回 }
    end;
    離開;
end;

<far 1554h>(DS:6506h, 名字 ＋ 'will be gone', 0Ah, 0);
顯示('Drop from party? ');
if <far 16AEh>(0Eh, 0Ah, 0Fh) <> 'Y' then begin
    <far 1554h>(DS:6506h, 名字 ＋ 'Breathes A sigh of relief', 0Ah, 1);
    離開;
end;

if DS:6506h^[196h] <> 0 then
    <far 1554h>(DS:6506h, 名字 ＋ 'bids you farewell', 0Ah, 1)
else
    <far 1554h>(DS:6506h, 名字 ＋ 'is dumped in a ditch', 0Ah, 1);

<far 0F0Fh>(1, 0);
<far 1EC9h+4>(0Bh, 26h, 1, 11h);
<far 14F8h+2>(DS:6506h);
```

### 兩句道別依 `+196h` 分流

`+196h` ≠ 0（還能動）→ `'bids you farewell'`；＝ 0（不能動，spec 824 的同一個
條件）→ `'is dumped in a ditch'`。**倒下的隊員是被丟進水溝的。**

### 「最後一個」的判定兩平台完全不同

**DOS**：直接看目前角色是不是唯一的成員。

```pascal
只剩這一個 := (DS:6506h^[189h] = NIL) and (DS:6506h = DS:650Ah);
```

**PC-98**：改成走一遍整個隊伍。

```pascal
只剩這一個 := 1;
p := DS:9598h;
while (p <> NIL) and (只剩這一個 <> 0) do begin
    if ((p^[196h] = 0) or (p^[196h] = 4))
       and (p^[0F7h] = 0)
       and (p <> DS:9594h) then
        只剩這一個 := 0;
    p := p^[18Ah];
end;
```

**這是重寫，不是翻譯。** DOS 只看串列結構；PC-98 額外要求「還有另一個成員的
`+196h` 是 0 或 4 且 `+0F7h` 是 0」才算不是最後一個。兩邊在同一個隊伍狀態下可能
給出不同答案，remake 必須挑一邊並記下來。

### 最後一個成員被踢掉 ＝ 離開遊戲

`<far 06EAh:0000h>` 就是 spec 799 認過的 Turbo Pascal halt 路徑，**不返回**。
所以 `'quit TO DOS: '` 那句問的是真的要不要結束程式。

### 中文化的兩個小地方

- `'Breathes A sigh of relief'` 的 **`A` 是大寫**，是原文的排版失誤；PC-98 譯成
  `'は安堵のため息をもらした。'`，沒有跟著保留。
- 三句道別／預告都接在名字後面，由 `1554h(角色, 字串, 0Ah, 旗標)` 顯示；
  **第一句的旗標是 0，其餘是 1**。

## 明確不宣稱

- 沒有宣稱 `far 0F0Fh(1, 0)` 做什麼（兩處都在「要離開」時呼叫）。
- 沒有宣稱 `far 1EC9h+4` / `14F8h+2` 做什麼。
- 沒有宣稱 PC-98 那個掃描條件（`+196h` ∈ {0, 4} 且 `+0F7h` ＝ 0）在找什麼樣的成員。
- 沒有宣稱 `far 16AEh` 三個參數的意義。
