# 841 — 抄寫卷軸：`+3Ch`..`+3Eh` 的最高位就是「抄寫中」

- 證據等級：`exact`（DOS 側逐條讀完；PC-98 側**助憶碼完全相同**，52 條差異全是位址與字串）
- 作法見 spec 783

## `overlay-15:00B9Bh` ↔ `overlay-15:00B86h`（entry#14，兩平台各 275 條）

`retf`（無參數）。

| | DOS | PC-98 |
|---|---|---|
| 提示（開頭） | `'Scribe These Spells? '` | `'これらの呪文を書き写しますか？'` |
| 沒有卷軸 | `'has no copyable scrolls'` | `'は書き写せる巻物を持っていない。'` |
| 已學會 | `'You already know that spell'` | `'その魔法はすでに知っている。'` |
| 已在抄寫 | `'You are already scibing that spell'` | `'その魔法はすでに書き写そうとしている。'` |
| 不能抄 | `'You can not scribe that spell.'` | `'その魔法は書き写せない。'` |
| 提示（結尾） | `'Scribe these spells? '` | **與開頭同一個字串** |

```pascal
if not <本模組 0391h>(3) then 離開;
離開 := 0;  選單狀態 := 0FFFFh;

法術 := <far 102Bh+1>(6, 0, @選單狀態, @有選);
重繪 := 1;
if 有選 <> 0 then begin
    if <本模組 16AEh>('Scribe These Spells? ', 0Fh, 0Ah, 0Eh) = 'N' then
        <本模組 01CBh>(遠指標(DS:6506h))
    else 離開 := 1;
end else 重繪 := 0;

選單狀態 := 0FFFFh;
while 離開 = 0 do begin
    法術 := <far 102Bh+1>(3, 3, @選單狀態, @有選);
    if 法術 = 0 then begin
        離開 := 1;
        if 有選 = 0 then <far 1554h>(角色, 名字 ＋ 'has no copyable scrolls', 0Ah, 1)
        else 重繪 := 1;
        continue;
    end;
    重繪 := 1;
    角色 := 遠指標(DS:6506h);

    if 角色^[78h + 法術] <> 0 then begin
        <far 154Eh+1>('You already know that spell');  continue;
    end;

    { 掃物品鏈，看有沒有已經在抄的同一個法術 }
    p := 角色^[14Dh];  撞到 := 0;
    while (p <> NIL) and (撞到 = 0) do begin
        if <本模組 11DEh>(p) then
            for i := 1 to 3 do
                if (p^[3Bh + i] > 7Fh)                       { 無號：最高位設著 }
                   and ((p^[3Bh + i] and 7Fh) = 法術) then begin
                    <far 154Eh+1>('You are already scibing that spell');
                    撞到 := 1;
                end;
        p := p^[2Ah];
    end;
    if 撞到 <> 0 then continue;

    { 該職業該等級還有可記憶欄位嗎 }
    等級 := 有號(byte[37DBh + 法術 * 10h]);
    職業 := 有號(byte[37DAh + 法術 * 10h]);
    if 角色^[12Ch + 職業 * 5 + 等級] > 0 then begin           { 無號 }
        p := 角色^[14Dh];  完成 := 0;
        while (p <> NIL) and (完成 = 0) do begin
            i := 1;
            repeat
                if p^[3Bh + i] = 法術 then begin
                    p^[3Bh + i] := p^[3Bh + i] or 80h;       { 標記抄寫中 }
                    完成 := 1;
                end;
                inc(i);
            until (i > 3) or (完成 <> 0);
            p := p^[2Ah];
        end;
    end else
        <far 154Eh+1>('You can not scribe that spell.');
end;

if 選單狀態 <> 0FFFFh then begin
    選單狀態 := 0FFFFh;
    法術 := <far 102Bh+1>(6, 0, @選單狀態, @有選);
    if 有選 <> 0 then
        if <本模組 16AEh>('Scribe these spells? ', 0Fh, 0Ah, 0Eh) = 'N' then
            <本模組 01CBh>(遠指標(DS:6506h));
end;
if 重繪 <> 0 then <本模組 15A9h>();
```

## 物品節點 `+3Ch`／`+3Dh`／`+3Eh` 的最高位 ＝ 抄寫中

spec 803 只知道那三格是「法術編號 ＋ 最高位旗標」，沒定名。這一支兩段程式碼把它定死：

- **讀**：`(槽 > 7Fh) and ((槽 and 7Fh) = 法術)` → 「這個法術已經在抄了」。
- **寫**：`槽 := 槽 or 80h` → 「開始抄這個法術」。

所以**卷軸上的三個法術槽，最高位就是「已排入抄寫」**。這與角色 `+1Eh` 起那 84 格
的最高位語意（spec 788：還在記憶中）是同一套設計手法。

## 兩張既有表在這裡交會

| 讀取 | 意義 | 出處 |
|---|---|---|
| `角色^[78h + 法術]` | 已學會這個法術（`+79h`..`+0DCh`，100 格） | 803／815 |
| `byte[37DAh + 法術 × 10h]` | 法術屬性 `+0` ＝ 職業類別 | 788 |
| `byte[37DBh + 法術 × 10h]` | 法術屬性 `+1` ＝ 等級 | 788 |
| `角色^[12Ch + 職業 × 5 + 等級]` | 該職業該等級的可記憶數 | 788 |

**抄寫的門檻是「該等級還有可記憶欄位」**，不是金錢也不是智力——這一點與 AD&D 的
原始規則不同，是引擎自己的簡化。

## 中文化：一個錯字、一組重複字串

- **`'You are already scibing that spell'` 少了一個 `r`**（應為 `scribing`）。
  這是原版的錯字，PC-98 翻成 `'その魔法はすでに書き写そうとしている。'` 時修掉了。
  中文化照正確寫法翻即可，但**搜尋原文時要記得用錯字**才找得到。
- DOS 有兩句只差大小寫的提示：`'Scribe These Spells? '`（開頭）與
  `'Scribe these spells? '`（結尾）。**PC-98 把兩處指到同一個字串**
  （`offset loc_AE9` 出現兩次）——可見原版的大小寫不一致是意外。
  中文化跟 PC-98 一樣合併成一句就好。

## 明確不宣稱

- 沒有宣稱 `far 102Bh+1(a, b, @選單狀態, @有選)` 前兩個常數（`6, 0` 與 `3, 3`）的意義。
- 沒有宣稱 `本模組 0391h(3)` / `01CBh` / `11DEh` / `15A9h` / `16AEh` 各做什麼。
- 沒有宣稱 `選單狀態` 那個 `0FFFFh` 哨兵值代表什麼。
- 沒有宣稱抄寫是在哪裡真正完成的（本支只設旗標，實際寫入靠休息流程，見 spec 813／814）。
