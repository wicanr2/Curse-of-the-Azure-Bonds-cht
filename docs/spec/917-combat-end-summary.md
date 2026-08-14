# 917 — 戰鬥結束結算：逃跑就丟掉全部戰利品

- 證據等級：`exact`（兩平台都逐條讀完；相似度 0.963，只有一個差異區塊）
- 作法見 spec 783

## `overlay-05:00A7Eh`（DOS 253 條）↔ `overlay-05:00A5Ch`（PC-98 270 條）

`retf 4` ＝ 一個 **longint（每人份的經驗值）**。

| | DOS | PC-98 |
|---|---|---|
| 逃跑 | `'The party has fled.'` | `'パーティーは逃げた。'` |
| 敗北 | `'You have lost the fight.'` | `'きみたちは戦いに敗れた。'` |
| 決鬥勝 | `'You have won the duel.'` | `'きみは決闘に勝った。'` |
| 隊伍勝 | `'The party has won.'` | `'パーティーは勝った。'` |
| 無戰鬥 | `'The party has found Treasure!'` | `'パーティーは宝物を見つけた！'` |
| 經驗值前綴（決鬥） | `'The duelist receives '` | `'決闘者は'` |
| 經驗值前綴（一般） | `'Each character receives '` | `'各キャラクターは'` |
| 經驗值後綴 | `'experience points.'` | `'点の経験点を得た。'` |
| 等按鍵 | `'press <enter>/<return> to continue'` | `'リターン・キーを押してください          '` |

```pascal
開視窗();                                            { 01A0h:0000h }

if (DS:47ECh = 0) and (DS:8B56h = 0) then
    印字(1, 3, 0Ah, 0, 'The party has found Treasure!')

else if DS:47EDh <> 0 then begin                     { 逃跑 }
    印字(1, 3, 0Ah, 0, 'The party has fled.');
    經驗值 := 0;                                     { ★ 逃跑不給經驗值 }
    p := 遠指標(DS:6F8Ch);                           { ★ 戰利品物品鏈 }
    while p <> NIL do begin
        暫 := p^[2Ah];  FreeMem(p, 3Fh);  p := 暫;
    end;
    FillChar(DS:6F70h, 1Ch, 0);                      { ★ 清空七個公用池 }
end

else begin
    勝 := (DS:8B5Ch <> 0) or
          ((DS:8B56h = 0) and (遠指標(DS:4F9Dh)^[5CCh] <> 0));   { word }
    if not 勝 then begin
        遠指標(DS:4F9Dh)^[58Eh] := 80h;              { word }
        印字(1, 3, 0Ah, 0, 'You have lost the fight.');
        經驗值 := 0;                                 { ★ 敗北不給經驗值 }
    end
    else if DS:8B56h <> 0 then
        印字(1, 3, 0Ah, 0, 'You have won the duel.')
    else
        印字(1, 3, 0Ah, 0, 'The party has won.');
end;

Str(經驗值, 0, @數字, 0Ah);
if DS:8B56h <> 0 then 句 := 'The duelist receives '  ＋ 數字
else                 句 := 'Each character receives ' ＋ 數字;
印字(1, 5, 0Ah, 0, 句);
印字(1, 7, 0Ah, 0, 'experience points.');

等按鍵(@結果, 'press <enter>/<return> to continue', 0Fh, 0Fh, 0Fh, 1, 0, @緩衝);
```

## ⚠ 逃跑會把戰利品整條丟掉

逃跑分支做兩件破壞性的事：

1. 走 `DS:6F8Ch` 的鏈，每個節點 `FreeMem(節點, 3Fh)`——**節點大小與 next 位移
   （`+2Ah`）跟角色身上的物品節點完全相同**（spec 832／849），所以
   **`DS:6F8Ch` 就是這場戰鬥掉落的戰利品鏈**。
2. `FillChar(DS:6F70h, 1Ch, 0)`——**`1Ch` ＝ 28 bytes ＝ 七個 longint**，
   正好是 spec 913 那個 `for j := 0 to 6 do longint[6F70h + j*4]` 的公用錢池。
   **spec 913 的「七格公用池」到這裡尺寸閉合。**

所以**逃跑 ＝ 錢與物品全部歸零**，而且經驗值也清成 0。

⚠ 這條鏈**沒有把頭指標 `DS:6F8Ch` 寫回 NIL**——只有節點被釋放，全域指標還指著
已釋放的記憶體。下一次戰鬥若沒有先重設它就會踩到懸空指標。兩平台都一樣。

## 三個旗標

| 位址（DOS／PC-98） | 語意 |
|---|---|
| `DS:47ECh`／`DS:78A2h` | **有沒有發生戰鬥**（0 且非決鬥 → 直接「找到寶藏」） |
| `DS:47EDh`／`DS:78A3h` | **有沒有逃跑** |
| `DS:8B56h`／`DS:0BDE8h` | **是不是決鬥**（決定「決鬥者」還是「各角色」的措辭） |
| `DS:8B5Ch`／`DS:0BDEEh` | **贏了沒** |

非決鬥時就算 `DS:8B5Ch` 是 0，只要 `遠指標(DS:4F9Dh)^[5CCh]`（word）不是 0
一樣算贏。敗北時往 `遠指標(DS:4F9Dh)^[58Eh]` 寫 `80h`。

## 中文化：經驗值句被拆成兩列

DOS 的句子橫跨兩列：

```
列 5:  Each character receives 1234
列 7:  experience points.
```

PC-98 照同樣的兩列切，但**切點落在日文文法能接的地方**
（`各キャラクターは1234` ／ `点の経験点を得た。`）。

繁中版同理可切成「各角色獲得 1234」／「點經驗值。」，
**但列 5 與列 7 中間空一列是寫死的**，不能合併成一句。

## PC-98 的兩處差異

- 等按鍵前多一段 10 bytes 的熱鍵字串指派到 `DS:0A335h`（同 spec 913 等六處的模式）。
- 等按鍵的三個屬性參數 `0Fh, 0Fh, 0Fh` 改成 `7, 0, 0`。

## 明確不宣稱

- 沒有宣稱 `遠指標(DS:4F9Dh)` 那筆記錄是什麼（`+58Eh`、`+5CCh` 兩個 word）。
- 沒有宣稱 `80h` 寫進 `+58Eh` 的意義。
- 沒有宣稱等按鍵那七個參數的排列。
- 沒有宣稱誰負責重設 `DS:6F8Ch`。
