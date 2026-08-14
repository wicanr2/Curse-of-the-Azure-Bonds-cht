# 818 — Take 的主迴圈（拿一件就從鏈上摘掉並釋放）、依方向偏移找記錄

- 證據等級：`exact`（DOS 側逐條讀完；PC-98 側的差異已逐條列出）
- 作法見 spec 783

## `overlay-05:0DB1h`（entry#7，DOS 101 條／PC-98 102 條）— Take 主迴圈

`retf`（無參數）。

```pascal
repeat
    選中 := NIL;  清單 := DS:6F8Ch;  旗標 := 0;
    <本模組 0CE6h>(@鍵, @選中, @清單, @旗標);          { spec 817 的 Take 清單 }
    if (鍵 <> 54h) and (鍵 <> 0Dh) then                { 'T' 或 Enter 才繼續 }
        結束 := 1
    else begin
        結束 := 0;
        <far 069Ah>(@失敗, 選中);
        if 失敗 = 0 then begin
            q := DS:6F8Ch;
            if q = 選中 then
                DS:6F8Ch := 選中^[2Ah]
            else begin
                while q^[2Ah] <> 選中 do q := q^[2Ah];
                q^[2Ah] := 選中^[2Ah];
            end;
            選中^[2Ah] := NIL;
            if 選中 <> NIL then FreeMem(選中, 3Fh);    { 63 bytes }
            if DS:6F8Ch = NIL then 結束 := 1;
        end;
    end;
until 結束 <> 0;
<far 15A9h>();
```

### 與 spec 817 的參數對得起來

`overlay-05:0CE6h` 的宣告順序是 `(參數C, 參數8, 參數4, 參數0)`，效果是
`參數C^ := 按鍵`、`參數8^ := 參數4^`。這裡傳的是 `(@鍵, @選中, @清單, @旗標)`，
而 `清單` 的初值是 `DS:6F8Ch` ——**選單把清單裡選到的那一格寫回 `選中`**。
兩支互相印證。

### 物品節點大小再次確認

`FreeMem(選中, 3Fh)` ＝ **63 bytes**（PC-98 `67h` ＝ 103），與 spec 789 由
`GetMem` 那一側量到的完全相同。

### 三個要照抄的細節

- **`'T'`（54h）與 Enter（0Dh）都算「確定」**，其他按鍵一律結束迴圈。
- 摘鏈的 `while q^[2Ah] <> 選中` **沒有 NIL 守衛**——選中的節點必須真的在鏈上。
- `if 選中 <> NIL then FreeMem(...)` 這個判斷**到不了 NIL**：前面已經把
  `選中^[2Ah]` 寫過了。是編譯出來的多餘檢查。
- 鏈清空（`DS:6F8Ch` ＝ NIL）就自動結束，不必再按鍵。

PC-98 在選單呼叫之後多一次 `<far 019Eh:0000h>()`；DOS 沒有（DOS 是在
`overlay-05:0CE6h` 裡呼叫 `01A0h:0000h`，spec 817）。

## `overlay-23:0D07h`（entry#7，97 條，實質差異 6 條）— 依方向偏移找記錄

`retf 0Eh` ＝ 7 個 word。依宣告順序：`(參數C, 參數A, 參數6, 參數4, 參數0)`，
其中 `參數6` 與 `參數0` 是遠指標。回傳 `dx:ax`。

```pascal
p := 參數6;  找到 := 0;
while (p <> NIL) and (找到 = 0) do begin
    for k := 1 to 參數4 do begin                       { 無號 }
        if p^[10h + k] = 0 then continue;
        d := 參數0^[k];
        if (p^[1Ah] + byte[2694h + d]) <> 參數C then continue;
        if (p^[1Bh] + byte[269Dh + d]) <> 參數A then continue;
        找到 := 1;
    end;
    if 找到 = 0 then p := p^[4];                       { next 在 +4 }
end;
if 找到 <> 0 then 結果 := 遠指標(p^[0]) else 結果 := DS:650Ah;
```

### 記錄的版面

| 位移 | 內容 |
|---|---|
| `+00h` | 一個遠指標（找到時就回傳它） |
| `+04h` | next |
| `+10h + k` | 每個 `k` 一個 byte 的旗標（**1 起算**，上界由 `參數4` 給） |
| `+1Ah` / `+1Bh` | 座標 x / y |

方向表 `DS:2694h` / `DS:269Dh` 是 spec 758 那一組。`參數0` 是一個 byte 陣列，
第 `k` 格是方向碼。

### 找不到時回的是隊伍串列頭

`DS:650Ah` ——**不是 NIL**。呼叫端拿到的一定是個有效的遠指標，所以分不出「找到
了」與「沒找到」，除非再比對一次。remake 照抄時要保留這個行為。

內層 `for` 找到之後**不會提早跳出**，剩下的 `k` 照跑（只是不會再改變結果）。

## 明確不宣稱

- 沒有宣稱 `far 069Ah` 判什麼（形狀上是「拿不拿得動」，回傳一個失敗旗標）。
- 沒有宣稱 `far 15A9h` 做什麼。
- 沒有宣稱 `overlay-23` 那條記錄鏈裝的是什麼。
- 沒有宣稱 `+10h + k` 那些旗標代表什麼。
