# 1112 — overlay 的 far call 到底打到哪：`<far 1590h>` 從來就不是位址

- 證據等級：`exact`（段位移常數由 5,079 個 DOS far call 與 5,262 個 PC-98 far call
  反推，兩平台各自收斂到單一值，`unresolved = 0`；三筆先前手工解出的目標逐一吻合）
- 工具：`cmd/far-call-map`
- 產出：`docs/audit/far-call-map-dos.{md,json}`、`docs/audit/far-call-map-pc98.{md,json}`

## 問題

IDA 把每個 overlay 的 `.bin` 從位址 0 載入。於是

```asm
9A B6 00 4D 01     call far ptr 014D:00B6
```

會被它顯示成 `call far ptr loc_1584+2`——因為 `0x14D0 + 0xB6 = 0x1586`
正好落在**這個 overlay 自己的位元組**裡。那個標籤與真正的目標毫無關係。

規格裡累積下來的 `<far 1590h>`、`<far 145Ah+3>`、`<far 154Ch+3>` 全部是這樣來的：
**看起來像位址，其實是工具的誤解**，而且每一份規格的「明確不宣稱」都要為它留一條。

## 解法

far pointer 指的是常駐段裡的 **TPOV 控制記錄**（`docs/audit/tpov-overlay-control-table.md`）：

```
段    → 哪一個 overlay 的控制記錄
位移  → (位移 − 20h) ÷ 5 ＝ entry 編號     （stub 表在記錄的 +20h，每筆 5 bytes）
```

段與控制記錄在檔案裡的段位址之間差一個常數，就是 EXE header 的長度：

| 平台 | 常數 | far call | 解出 entry | 目標是常駐 | 段對得上但位移不合法 |
|---|---:|---:|---:|---:|---:|
| DOS | `7Bh` 段（＝ `7B0h` bytes） | 5,079 | 2,267 | 2,812 | **0** |
| PC-98 | `9Ah` 段（＝ `9A0h` bytes） | 5,262 | 2,291 | 2,971 | **0** |

⚠ **常數是量出來的，不是查出來的。** `cmd/far-call-map` 掃過 0..400h 每一個候選值，
數「有幾個目標正好落在 stub 邊界、而且 entry 編號在該 overlay 的範圍內」，取命中
最多的。位移不在 stub 邊界時**回錯誤而不是取最近的 entry**——四捨五入會把
「假設錯了」變成一份看起來合理的對照表。

## 三筆獨立驗證

先前有三支規格是**手工**解出目標的，本工具逐一重現：

| 規格 | 手工結論 | 工具輸出 |
|---|---|---|
| 1010（機會攻擊） | `overlay-13 sub_95A` 叫 `overlay-24 entry#32`、`overlay-32 entry#18`、`overlay-31 entry#4` | 一模一樣 |
| 1009（法術目標） | `'Already been targeted'` 走 `overlay-24 entry#19` 的訊息框 | `sub_D1C` 的 `<far 154Ch+3>` ＝ `overlay-24 entry#19` |
| 1009（模式 5 上限） | `2d4` 是 `overlay-23 entry#9` | `sub_D1C` 的 `<far 145Ah+3>(2,1)` ＝ `overlay-23 entry#9`，參數 `(2,1)` 就是 1d2 |

## 立刻解掉的兩個未定項

### 一、`overlay-24 entry#30` ＝ 「對面那一隊的隊號」

```pascal
function 對面隊號(p: 遠指標): byte;   { overlay-24 entry#30，code 25B0h }
begin
    if p^[197h] = 0 then 對面隊號 := 1 else 對面隊號 := 0;
end;
```

`+197h` 是隊號（spec 1010）。這一支回的是**相反的那一隊**。

### 二、spec 777 的陣營判斷是反的

`overlay-09:02B1h`（範圍法術施放前的掃描）原本寫成

```pascal
我方 := <呼叫>(DS:6506h);
if 目標^[197h] <> 我方 then …
```

`<呼叫>` 就是上面那支 `overlay-24 entry#30`，回的是**對面**的隊號。所以
`目標^[197h] <> 對面隊號` ＝ **目標和施法者同一隊**——這一支掃的是**自己人**。

這與呼叫端完全吻合：spec 802／835 的 AI 在施放前叫它，**任何一個目標回真就不放**。
「有隊友會中招就別放」是合理的；原本的讀法會變成「有敵人會中招就別放」。

> **`overlay-09:02B1h` ＝ 範圍法術的友軍誤傷檢查。** remake 的怪物 AI 要照這個
> 方向實作，否則會得到一支專打自己人的 AI。

## 逃跑判定因此完整了（spec 799 的補完）

`overlay-13:0D1Ch` 的每一個外呼現在都有名字：

```pascal
function 逃跑判定(角色: 遠指標): byte;
begin
    逃掉 := 0;
    if <overlay-24 entry#32>(0FFh, 角色) = 0 then     { 沒有任何戰鬥員貼著 }
        逃掉 := 1
    else begin
        半 := 有號(<本模組 0124h>(角色)) div 2;        { 我的移動率，格 }
        v  := <本模組 2EACh>(角色);                    { 對面隊伍中最快的移動率 }
        if v < 半 then 逃掉 := 1                       { 敵人全都比我慢 }
        else if v = 半 then
            if <overlay-23 entry#9>(2, 1) = 1 then 逃掉 := 1;   { 1d2，平手偏向失敗 }
    end;
    if 逃掉 <> 0 then <overlay-23 entry#12>(角色, 3, 名字 ＋ 'Got Away')
    else              <overlay-24 entry#19>('Escape is blocked');
    逃跑判定 := <overlay-24 entry#34>(角色);           { 回傳的不是逃跑成敗 }
end;
```

`2EACh` 走 `DS:650Ah` 的戰鬥員鏈，條件是 `對面隊號(角色) = p^[197h]` 且
`p^[196h] <> 0`（站著、能行動，spec 1010），取 `0124h(p) div 2` 的最大值。
**所以比的是「我的速度」對「敵方最快的速度」**——和 AD&D 的逃跑規則同形。

## 怎麼用

讀某一支 overlay 函式時，把它的位址丟進 JSON 查：

```sh
tools/go.sh run ./cmd/far-call-map -platform dos \
    -output docs/audit/far-call-map-dos.md -json docs/audit/far-call-map-dos.json
```

JSON 每個呼叫點一列（來源模組／函式／位址／原始 far ptr／目標 entry／code offset）；
Markdown 是反向索引（某個 entry 被誰叫），用來回答「改這支會影響誰」。

## 明確不宣稱

- 沒有宣稱「目標是常駐程式碼」那 2,812 筆各自打到哪。常駐段沒有 stub 表，
  要另外用 `START.EXE` 的符號或呼叫圖去解。
- 沒有宣稱兩個平台的 entry 編號可以直接對應。兩邊的 overlay 內容雖然同源，
  但 entry 數不同（DOS `overlay-24` 有 53 個、PC-98 不一定），要逐支比對。
- 沒有宣稱 stub 的 `flags` 那個位元組是什麼。
