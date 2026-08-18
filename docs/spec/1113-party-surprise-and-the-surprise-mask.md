# 1113 — `22h PARTY SURPRISE` 只算遊俠，第二個值永遠是 0；突襲遮罩沒有人設它

- 證據等級：`exact`（DOS `overlay-02:1636h` 逐條讀完；`+596h` 的讀寫由 36 個
  overlay 的完整 body dump 全掃）／**未定**（誰把突襲遮罩設成非 0）
- 對應工作項：`RE-06` → `ENG-07` 的 surprise 那一格

## 一、`22h` 的 handler（DOS `overlay-02:1636h`）

```pascal
<overlay-07 entry#2>(2);                       { 取兩個運算元 }
p := DS:650Ah;                                  { 隊伍鏈 }
遊俠 := 0;  第二值 := 0;  另一格 := 0;
while p <> NIL do begin
    if (p^[75h] = 4) or (p^[75h] = 0Ah) then begin
        遊俠   := 1;                            { [bp−5] }
        另一格 := 1;                            { [bp−7] ← 沒有任何讀者 }
    end;
    p := p^[189h];
end;
<overlay-07 entry#15>(<overlay-07 entry#9>(運算元 0), 遊俠);
<overlay-07 entry#15>(<overlay-07 entry#9>(運算元 1), 第二值);   { ← 恆為 0 }
```

`+75h` 是**職業組合編號**（spec 1093）。含遊俠的組合只有兩個：`4`（單職遊俠）
與 `0Ah`（牧師／遊俠）——與程式碼比對的那兩個值一格不差。

### ★ 第二個目的地寫的是一個從沒被指定過的區域變數

`[bp−5]`、`[bp−6]`、`[bp−7]` 三格都在進場時清成 0。有遊俠時寫的是 `[bp−5]`
與 `[bp−7]`，最後兩次寫回取的是 `[bp−5]` 與 **`[bp−6]`**。也就是說：

> **第一個目的地 ＝ 隊伍裡有沒有遊俠；第二個目的地 ＝ 常數 0。**
> `[bp−7]` 被設成 1 之後沒有任何讀者——那是原作的死碼。

形狀上像是「第二個值本來要放某種突襲判定，寫錯了變數」。不論成因，
**可觀察的行為就是恆 0**，remake 照抄（`internal/ecl` 的
`TestPartySurpriseSecondDestinationIsAlwaysZero` 擋住「補上」它）。

這關掉了 spec 264 留的那一條：「`PARTY SURPRISE` 第二個值必須留在 State
adapter 直到驗證」——現在驗過了，它不需要 adapter。

## 二、突襲遮罩：只有清除，沒有設定

先攻擲骰（spec 806 的 `overlay-13:sub_0`）會看一個遮罩：

```pascal
c^[3] := 有號(<far 1524h+3>(角色)) + <overlay-23 entry#9>(6, 1);   { 1d6 + 反應調整 }
if ((角色^[197h] + 1) and word(DS:4F9Dh^[596h])) <> 0 then c^[3] := c^[3] − 6;
```

`DS:4F9Dh` 是區域記錄的遠指標，`+596h` 就是**突襲遮罩**：位元對到隊號 + 1，
中了就先攻 −6（一整個 1d6 的量，實務上等於讓對方先動一輪）。

把 36 個 DOS overlay 的完整 body dump 全掃過，`+596h` 只出現兩次：

| 位置 | 動作 |
|---|---|
| `overlay-13 sub_0`（先攻） | **讀** |
| `overlay-08 sub_F3` | `xor ax, ax` ＋ **寫 0**（清除） |

**沒有任何 overlay 把它設成非 0。**

### 這一格已經收掉：整個 build 都沒有人設它

常駐段與指標間接的補掃、相鄰欄位的正對照，以及 PC-98 的第二證人都做完了，
結論是**沒有任何地方把它寫成非 0**，先攻的 `−6` 在這一作走不到。
證據與掃描口徑見 [spec 1136](1136-surprise-mask-dead-and-guard-reservation.md)。

⚠ 這**不**代表「原作沒有突襲」：`22h PARTY SURPRISE` 這條指令仍然在跑，
只是它的結果沒有接到先攻上。

## 三、remake 現況

`RollActionDelay(roller, dexterity, combatTeam, surpriseMask)` 已經實作
`(隊號 + 1) and 遮罩 ⇒ −6`，形狀與原作相同；呼叫端目前一律傳 **0**。

傳 0 是**唯一有證據的值**：整個 build 對這個欄位只有一次讀與一次清 0
（spec 1136）。`TestInitiativeStaysInsideTheNoSurpriseRange` 釘住可觀察的後果
——沒有 `−6`，先攻的值域就是 `1..11`。

## 明確不宣稱

- 沒有宣稱 `overlay-07 entry#9`／`entry#15` 的內部（取運算元／寫回目的地是由
  呼叫形狀推的，不是讀過本體）。
- 沒有宣稱 `[bp−7]` 那個死碼在原始 Pascal 裡本來要做什麼。
- 沒有宣稱突襲遮罩的位元寬度上限（讀的時候當 word，隊號只用到低位兩三個位元）。
