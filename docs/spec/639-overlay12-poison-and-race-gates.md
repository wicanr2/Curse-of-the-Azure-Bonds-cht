# 第六百三十九輪：`overlay-12` 七支 —— 毒的即死流程與兩支只差一行的種族閘

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-12` 的 `133Fh`、`147Ah`、`16C2h`、`176Dh`、`17B5h`、
`180Ch`、`1973h`。

## `147Ah`：豁免失敗就死

```text
DS:0A03Dh := arg_4^[18Eh]^[0Ah]
if <far 013Eh:0028h>(arg_2, 0, DS:0A03Dh) = 0 then      ← 豁免失敗
    備妥 'は毒を受けた。'，<far 014Ah:0024h>(DS:0A03Dh, 訊息, 0, 0Ah)
    <far 0418h:14AAh>()
    <sub_1437>(DS:0A03Dh, 37h, 0, 0FFh, 0)
    備妥 'は死んだ。'，<far 013Eh:0005h>(DS:0A03Dh, 訊息, 6)
```

兩句訊息**接連出現在同一條路徑上**：中毒之後沒有任何判定就直接死亡。這是 AD&D
的毒（豁免成功則無事、失敗即死），不是逐回合扣血。

### IDA 給的長度短了 69 bytes

IDA 認定這支是 86 bytes（到 `14D0h`），但實際的 `retf 6` 在 `1512h`——真正的長度是
**155 bytes**。`14CFh` 那條 `call far 0418h:14AAh` 本身就跨出了 IDA 的範圍。

只讀 IDA 給的範圍會**在「は毒を受けた。」之後就停住**，完全看不到後面的死亡流程，
而且看起來像是讀完了。這是 `scripts/show.py --whole` 走 prologue 區間、不信 IDA
邊界的理由。

## `16C2h` 與 `06F9h`：長得像，但一行的位置不同

```text
06F9h（第 633 輪）:
    if arg_0 = 0 and <sub_1547>(…) = 0 then
        DS:0A035h := 1
        DS:0A039h := DS:0A039h − 4        ← 兩行都在條件內

16C2h:
    if DS:9594h^[11Ah] = 13h then
        if <sub_1547>(…) = 0 then
            DS:0A035h := 1                 ← 只有這行在內層條件裡
        DS:0A039h := DS:0A039h − 4         ← 這行照做
```

`16C2h` 的 `jnz` 跳到 `16F2h`，也就是**減法那一行**，不是跳過它。所以只要
`RACETYPE = 13h`，`DS:0A039h` 一定減 4，`sub_1547` 的結果只決定要不要設
`DS:0A035h`。

兩支的差別只有一個跳躍目標。**照著「跟另一支一樣」抄會抄錯**。

## 種族閘三支

```text
16C2h:  DS:9594h^[11Ah] = 13h
17B5h:  DS:0A03Dh^[11Ah] = 3
180Ch:  DS:0A03Dh^[11Ah] = 8
```

`+11Ah` 是 `RACETYPE`（[spec 499](499-pc98-alignment-conditional-effects.md)）。
三支各自針對一個種族值，動作不同：

```text
17B5h:  x := <far 0176h:0506h>(arg_6)            ← cbw，有號
        DS:0A02Eh := byte(ROLLDICE(1, 0Ch) × 3 + 4 + x)
        DS:0A039h := DS:0A039h + 2

180Ch:  DS:0A039h := DS:0A039h + 3
        DS:0A02Eh := DS:0A02Eh + 3
```

`17B5h` 的 `1d12 × 3 + 4` 範圍是 `7..40`，再加上一個有號的 `x`，最後**只存低
byte**。

## 其餘兩支

```text
133Fh:  p := <sub_FCC>(DS:9594h)
        if p = nil or p^[5Ah] = 0 then
            DS:0A02Eh := 0
        else if p^[5Ah] < 3 then
            DS:0A02Eh := DS:0A02Eh div 2          ← 有號除法

176Dh:  if DS:0A02Fh and 20h <> 0 then
            <sub_1B>(0)
            備妥 'は影響を受けなかった。'，<far 014Ah:0024h>(arg_6, arg_8, 訊息, 1, 0Ah)

1973h:  if <far 0176h:0473h>(byte(arg_6^[1A5h]), arg_6) = 0 then
            <sub_3C>(1, arg_2^[3], 4Eh, arg_6)
```

`133Fh` 的第二個條件在原始碼裡重複檢查了一次 `p <> nil`（`or ax, dx` ＋ `jz`），
即使第一個條件已經涵蓋——編譯出來是兩次檢查，不是一次。

## 一筆分類錯誤：`1414h` 不是函式起點

`scripts/verify_size_truncation.py`（本輪新增）查出 `overlay-12:1414h` 被標成
「已解讀」，但它落在 `13F1h..1434h` 那一支的中段——`13F1h` 才有
`push bp / mov bp, sp`，`1434h` 才是 `retf 0Ah`。

原本的分類**剛好相反**：`13F1h` 被標成邊界碎片、`1414h` 被標成已解讀。兩筆已對調。
完整的一支是：

```text
if <sub_3C>(arg_A:arg_8, 3Eh, arg_4^[3], 3Ch) <> 0
   and <far 013Eh:008Eh>(arg_A:arg_8, 1, 1) <> 0 then
    <far 014Ah:00B1h>(arg_A:arg_8)
```

`sub_3C` 的呼叫形狀與 `1973h` 一致（`(far pointer, id, byte, 旗標)`），是同一支
routine 的第二個呼叫端。

## 明確不宣稱

- `+5Ah`／`+1A5h` 以外欄位的意義，以及 `DS:0A02Eh`／`DS:0A039h`／`DS:0A035h` 的身分。
- `RACETYPE` 的 `3`／`8`／`13h` 各對應哪個種族。
- 效果 id `37h`／`4Eh` 與各 far routine 的行為。
