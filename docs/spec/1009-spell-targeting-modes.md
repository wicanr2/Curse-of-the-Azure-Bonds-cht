# 1009 — 法術目標模式：一個 nibble 決定自己／N 個／半徑幾格，以及半徑被吃掉的缺陷

- 證據等級：`exact`（DOS 側 483 條逐條讀完；PC-98 對側 483 條，
  差異只有兩條 `xor ah, ah`、兩條有號/無號跳躍、與一個多存的全域）
- 作法見 spec 783

## `dos overlay-13:02220h` ↔ `pc98 overlay-13:0225Fh`（`retf 8`）

兩側原本都是待解讀。這一支決定**一個法術要打誰**。

```pascal
procedure 選目標(法術: byte;      { bp+0Ch }
                 自動: byte;      { bp+0Ah，非 0 ＝ 不出提示、直接扣額度 }
             var 成功: byte);     { bp+06h..09h，遠指標 }
```

## ★★ 法術屬性表的 `+6` 是目標模式

```pascal
模式 := byte[37E0h ＋ 法術 × 16] and 0Fh;
```

★ 表的基底是 **`DS:37DAh`**（spec 1016 由同時讀 `37DAh` 與 `37DBh` 的程式定案），
筆距 16，所以 `37E0h` 是記錄的 **`+6`**。PC-98 對應 `DS:61B4h`／`61BAh`。
同一張表的 `+0` 是施法職業、`+1` 是法術環數（spec 1016）。

法術名稱表在 `DS:27BDh`，**筆距 41、編號從 1 起算**
（`DS:27BDh ＋ 法術 × 41` 是 Pascal 短字串）。把兩張表對起來：

| 模式 | 做法 | 法術數 | 例子 |
|---|---|---|---|
| `0` | **只打自己** | 31 | Detect Magic、Read Magic、Shield、Find Traps、Burning Hands |
| `2`／`3`／`4`／`6`／`7` | **固定 `(模式 and 3) ＋ 1` 個目標** | 50 | Cure Light Wounds(1)、Charm Person(1)、Hold Person(3)、Hold Monsters(4) |
| `5` | **逐一點選，累計等級權重到擲骰上限** | 4 | Faerie Fire、Charm Monsters |
| `8`..`0Eh` | **範圍效果，半徑 ＝ 模式 `and 7`** | 32 | Lightning Bolt(0)、Sleep(1)、Bless(2)、Fireball(3) |
| `0Fh` | **打已鎖定的目標，沒有就取整片** | 2 | Silence, 15' Radius |

★★★ **與 AD&D 1e 完全對得上**：

- **Hold Person 有兩支**——牧師版（法術 23，模式 `6` ⇒ **3 個**）與法師版
  （法術 49，模式 `7` ⇒ **4 個**）。1e 的牧師 Hold Person 影響 1–3 人、
  法師版 1–4 人，**一個不差**。
- 半徑階序 **Fireball(3) > Bless／Ice Storm(2) > Sleep／Cloud Kill(1) >
  Lightning Bolt／Cone of Cold(0)**——線狀與錐狀法術半徑 0，形狀由別處決定。
- Burning Hands 是模式 `0`（自己），符合「從施法者手上噴出去」。

## ★★ 模式 5 是 AD&D 的「總 HD 額度」

```pascal
if 法術 = 4Fh then 上限 := <overlay-24 entry#36>(4Fh)
              else 上限 := <overlay-23 entry#9>(2, 4);      { ＝ 2d4 }
```

`4Fh` ＝ 79 ＝ **Faerie Fire**，它有自己的一套上限；其餘走 **2d4**。
每選一個目標就依它的等級加權，超過上限就停：

| 一般法術：`目標^[0E5h]` | 加 | Faerie Fire：`目標^[0DEh]` | 加 |
|---|---|---|---|
| 0 或 1 | 1 | 1 | 1 |
| 2 | 2 | 2 或 3 | 2 |
| 3 | 4 | 4 | 4 |
| 其他 | 8 | 其他 | 0 |

★ **`2d4` 與逐級加倍的權重就是 1e Sleep 那張「2d4 HD 的生物」表**。
⚠ 判斷是 `(已選數 > 1) and (總和 > 上限)`——**第一個目標一定收得下**，
就算它自己就超過上限。

## ★★ 缺陷：模式 `0Fh` 的半徑恆為 0

```asm
mov al, [di+37E0h]
and al, 0Fh          ; ← 先把高 nibble 砍掉
xor ah, ah
mov cx, 4
shr ax, cl           ; ← 再右移 4 位 ⇒ 結果一定是 0
push ax              ; 當半徑傳給 overlay-31 entry#6
```

原始位元組 `8a 85 e0 37 / 24 0f / 30 e4 / b9 04 00 / d3 e8 / 50`，
**兩平台一模一樣**（PC-98 在 `24DBh`，基底 `61BAh`）。

資料端證明它不是「本來就想傳 0」：

| 法術 | 名稱 | `+6` | 高 nibble |
|---|---|---|---|
| 25 | **Silence, 15' Radius** | `1Fh` | **1** |
| 113 | （未命名） | `2Fh` | **2** |

> **模式 `0Fh` 的兩支法術在資料裡各自帶了半徑 1 與 2，
> 程式卻因為先 `and 0Fh` 再 `shr 4` 而永遠拿到 0。**
> 對照組：模式 `8..0Eh` 那條路徑寫的是 `and 7`，**沒有多餘的右移**，
> 半徑正確地取到 0..7。

⚠ remake 要不要「修好」是設計決定——**修了 Silence, 15' Radius 的作用範圍
就和原版不同**。本規格只記錄事實。

## 流程

```pascal
成功 := 1;
目標數 := 0;  DS:6F9Dh := 0;                  { 範圍效果旗標 }
中心x := <overlay-32 entry#15>(施法者);        { DS:7559h }
中心y := <overlay-32 entry#16>(施法者);        { DS:755Ah }

case 模式 of
  0:  begin 目標表[1] := 施法者;  目標數 := 1 end;

  5:  repeat                                   { 見上一節 }
        if <sub_200F>(@候選, 0, 自動, 法術) = 0 then break;
        if 候選 已經在目標表裡 then
            if 自動 <> 0 then dec(上限) else 顯示 'Already been targeted'
        else begin
            目標表[inc(n)] := 候選;  中心 := 候選座標;  inc(目標數);
            總和 += 權重(候選);
            if (n > 1) and (總和 > 上限) then break;
        end;
        <overlay-32 entry#7>(候選x, 候選y);
      until 上限 = 0;

  0Fh: if <sub_200F>(@候選, 0, 自動, 法術) = 0 then 成功 := 0
       else if 施法者^[18Dh]^[0Ah] <> NIL then begin
           目標表[1] := 施法者^[18Dh]^[0Ah];  目標數 := 1     { ★ 用已鎖定的目標 }
       end else 取範圍(半徑 ＝ 0);                             { ← 上一節的缺陷 }

  8..0Eh: if <sub_200F>(@候選, 1, 自動, 法術) = 0 then 成功 := 0
          else 取範圍(半徑 ＝ 模式 and 7);

  else begin
      剩 := (模式 and 3) ＋ 1;
      while 剩 > 0 do begin
          if <sub_200F>(@候選, 0, 自動, 法術) = 0 then break;
          if 候選 已經在目標表裡 then
              if 自動 <> 0 then dec(剩) else 顯示 'Already been targeted'
          else begin 目標表[inc(n)] := 候選;  dec(剩);  中心 := 候選座標 end;
          <overlay-32 entry#7>(候選x, 候選y);
      end;
      目標數 := n;
      if n = 0 then 成功 := 0;
      中心 := 目標表[n] 的座標;                { ⚠ 見下 }
  end;
end;
```

「取範圍」那一段是：

```pascal
<overlay-31 entry#6>(遠指標(DS:6E92h), 1, 0FFh, 半徑, 中心y, 中心x);
for i := 1 to DS:6E96h do
    目標表[i] := 遠指標(DS:6D35h ＋ byte[6E94h ＋ i × 3] × 4);
目標數 := DS:6E96h;
DS:6F9Dh := 1;                                { ★ 標記「這是範圍效果」}
```

`DS:6E94h` 起是 `entry#6` 回填的結果，**每筆 3 bytes**，第 0 個 byte 是戰鬥員編號；
`DS:6D35h` 是戰鬥員遠指標表（spec 805／851／1008 同一張）。

## ⚠ 目標數 0 時仍然去讀 `目標表[0]`

目標表用 `基底 ＋ i × 4` 索引且 **i 從 1 起算**，
所以 `基底` 的那四個 byte（DOS `7431h..7434h`）**最後一個就是計數器 `DS:7434h` 本身**。

固定數量那一支在 `n = 0`（一個都沒選到）時仍然執行

```pascal
中心x := <entry#15>(目標表[0]);   { ＝ 讀 DS:7431h 那四個 byte 當遠指標 }
中心y := <entry#16>(目標表[0]);
```

**把計數器與旁邊三個 byte 當成遠指標傳出去。**
兩平台都一樣（PC-98 基底 `0A51Dh`、計數器 `0A520h`）。
`成功` 已經被設成 0，所以呼叫端多半不會用這個中心——但這兩次呼叫本身照跑。

## 兩平台差異

| | DOS | PC-98 |
|---|---|---|
| 施法者遠指標 | `DS:6506h` | `DS:9594h` |
| 目標表基底／計數 | `DS:7431h`／`7434h` | `DS:0A51Dh`／`0A520h` |
| 範圍查詢結果 | `DS:6E92h`／`6E94h`／`6E96h` | `DS:9F2Ch`／`9F2Eh`／`9F30h` |
| 戰鬥員遠指標表 | `DS:6D35h` | `DS:9DCFh` |
| 中心座標 | `DS:7559h`／`755Ah` | `DS:0A645h`／`0A646h` |
| 範圍效果旗標 | `DS:6F9Dh` | `DS:0A037h` |
| 法術屬性表（基底） | `DS:37DAh` | `DS:61B4h` |
| 戰鬥狀態欄位 | 角色 `+18Dh` | 角色 **`+18Eh`** |

PC-98 另外多兩條：`ds:0BE2Dh := 法術編號`（DOS 沒有這個全域）。
`8..0Eh` 的範圍判斷 DOS 用有號跳躍（`jge`／`jle`）、PC-98 用無號（`jnb`／`jbe`）
——值已經 `and 0Fh` 過，兩者等價。

## 中文化

唯一的字串是 `'Already been targeted'`（21 字元，`CS:220Ah`），
玩家重複點同一個目標時出現，走 `overlay-24 entry#19` 的訊息框。

**法術名稱表 `DS:27BDh` 筆距 41**（含長度 byte ⇒ 最多 40 字元 ＝ **20 個漢字**），
在常駐資料段，可以整批換掉。

## 明確不宣稱

- 沒有宣稱 `sub_200F`（同模組的「請玩家點一格」）內部行為，
  也沒有宣稱它第二個參數 0／1 的差別。
- 沒有宣稱 `overlay-31 entry#6` 的前三個參數（`1`、`0FFh`）代表什麼。
- 沒有宣稱 `+0E5h`／`+0DEh` 這兩個等級欄位的完整定義。
- 沒有宣稱 `overlay-24 entry#36(4Fh)` 回什麼（Faerie Fire 的上限）。
- 沒有宣稱法術屬性表 `DS:37DAh` 其餘 13 個 byte 是什麼（`+0`／`+1` 見 spec 1016）。
- 沒有宣稱 `DS:0BE2Dh`（PC-98 才有）給誰用。
