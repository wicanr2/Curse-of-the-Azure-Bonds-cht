# 891 — 戰鬥收場：釋放戰利品鏈、發經驗值、重設畫面模式

- 證據等級：`exact`（DOS 側逐條讀完；PC-98 側差異逐條列出）
- 作法見 spec 783

## `overlay-05:01736h`（DOS 161 條）↔ `overlay-05:01775h`（PC-98 192 條）

`retf` ＝ 無參數。兩句訊息：

| | DOS |
|---|---|
| 全滅 | `'The monsters rejoice for the party has been destroyed'` |
| 等按鍵 | `'Press any key to continue'` |

```pascal
遠指標(DS:4F9Dh)^[58Eh] := 0;
DS:47ECh := 0;
if DS:4FBCh = 0 then <overlay-05:0053Ch>();      { spec 861 戰鬥結束處理 }
DS:4FBAh := 6;                                    { 畫面模式 6，spec 876 }
<本模組 1370h>();
if DS:4FBCh <> 0 then 離開;

p := 遠指標(DS:650Ah);
while p <> NIL do begin <far 1512h+1>(p);  p := p^[189h] end;

if (DS:4FC7h <> 0) and (DS:8B56h = 0) then goto 敗北;

if DS:47EDh <> 0 then 釋放戰利品鏈;               { 有人在逃 → 東西不要了 }

if DS:4FBCh = 0 then begin
    <本模組 14ACh>();
    <本模組 0A7Eh>(longint@DS:8B58h);             { 每人份經驗值 }
    <本模組 106Ch>();
end;

釋放戰利品鏈;
goto 收尾;

敗北:
    遠指標(DS:4F9Dh)^[58Eh] := 80h;
    <far 01A0h:0000h>();
    DS:65A0h := 2;  DS:65A1h := 6;
    訊息('The monsters rejoice for the party has been destroyed');
    <far 0542h:04A4h>(1, 0, 0Ah, 16h, 25h, 5, 2);
    訊息('Press any key to continue');
    <far 0542h:0946h>(0, 0Dh);

收尾:
    DS:4FC9h := 1;
    遠指標(DS:4F9Dh)^[6E0h] := 0;
    遠指標(DS:4F9Dh)^[6E2h] := 0;
    遠指標(DS:4F9Dh)^[6E4h] := 0;
    遠指標(DS:4F9Dh)^[5C6h] := 0;
    遠指標(DS:4F9Dh)^[5CCh] := 0;
```

其中「釋放戰利品鏈」兩處寫法完全相同：

```pascal
q := 遠指標(DS:6F8Ch);
while q <> NIL do begin
    暫 := q;  q := q^[2Ah];
    FreeMem(暫, 3Fh);                             { PC-98 67h }
end;
DS:6F8Ch := NIL;
```

## 關掉 spec 858 的最後一個問號

spec 858 記著「`DS:6F8Ch` 那條戰利品鏈由誰消費、誰釋放，仍未讀到」。
**這一支就是釋放端**，而且釋放兩次：

- **有人逃跑時**（`DS:47EDh ≠ 0`，spec 861 設的）先釋放一次，
  然後**跳過發經驗值那一段**——逃走就沒有戰利品、也沒有經驗值。
- 正常結束時發完經驗值再釋放。

節點大小 `3Fh`／PC-98 `67h` 與 spec 832／849／858 的物品節點一致。

`本模組 0A7Eh(longint@DS:8B58h)` 收的正是 spec 861 寫進 `DS:8B58h` 的每人份，
所以整條戰後鏈到這裡收尾：

```
overlay-05:0053Ch （861 戰鬥結束處理）→ 0039h（858 算每人份）→ 0351h（857 發下去）
overlay-05:01736h （本支 收場）      → 0A7Eh（再發一次？）→ 釋放戰利品鏈
```

⚠ `0A7Eh` 與 `0351h` 都吃 `DS:8B58h`，**本規格不宣稱它們是不是同一件事**。

## 收場時清掉的五格

`遠指標(DS:4F9Dh)` 上的 `+6E0h`／`+6E2h`／`+6E4h`／`+5C6h`／`+5CCh` 全部歸零。
其中兩格先前已有出處：

- `+5C6h`：spec 858 讀到「＝ 1 就跳過物品收集」。
- `+5CCh`：spec 861 讀到「≠ 0 時瀕死者直接復原」。

所以那些都是**單場戰鬥有效的設定，收場時一律清掉**。

## ⚠ PC-98 自己重算「戰鬥是不是真的結束」

DOS 有三處 `DS:4FBCh` 的分支，PC-98 全部沒有；PC-98 改成走一遍隊伍鏈：

```pascal
if DS:7F34h <> 0 then begin                       { ＝ DOS 的 DS:4FC7h }
    DS:7F34h := 1;
    p := 遠指標(DS:9598h);
    while (p <> NIL) and (DS:7F34h <> 0) do begin
        if not (p^[196h] in [1, 2, 6, 7, 8]) then DS:7F34h := 0;
        p := p^[18Ah];
    end;
end;
```

`+196h`（PC-98）＝ DOS 的 `+195h` 狀態碼（spec 825）：
`1` Animated、`2` tempgone、`6` Dead、`7` Stoned、`8` Gone。
**只要還有一個人不在這五種狀態裡，就不算全滅。**

DOS 的 `DS:4FC7h` 是在 spec 861 裡算好的；**PC-98 在這裡又算了一次**，
而且用的是不同的判準。remake 要選一邊。

PC-98 另外多一段 `far 0893h:0114h(DS:8BF3h)` ＋ `DS:7F36h := 1`。
本規格不宣稱那是什麼。

## 敗北路徑就是全滅

字串 `'The monsters rejoice for the party has been destroyed'` 把那條路徑定死了：
`(DS:4FC7h ≠ 0) and (DS:8B56h = 0)` 成立就是**全隊被消滅**。

`DS:4FC7h` 是 spec 861 算出來的「戰鬥真的結束」旗標，
所以「真的結束」＋「`DS:8B56h` 沒設」＝ 玩家輸了。
反過來說 `DS:8B56h` ≠ 0 那條路（spec 858 的固定經驗值、spec 861 的特殊分支）
**即使全隊倒下也不會走全滅畫面**——形狀上是劇情安排的必敗戰鬥。

## 明確不宣稱
- 沒有宣稱 `DS:4FBCh`／`DS:8B56h`／`DS:47ECh`／`DS:4FC9h`／`DS:65A0h`／
  `DS:65A1h` 的意義。
- 沒有宣稱 `本模組 1370h`／`14ACh`／`0A7Eh`／`106Ch`／`far 1512h+1` 各做什麼。
- 沒有宣稱「敗北」那條路徑是不是真的全滅（形狀上是，但沒有字串佐證）。
