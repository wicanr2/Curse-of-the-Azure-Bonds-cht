# 1146 — 遭遇距離：`0Ch` 擺、`0Dh` 走近、`29h` 重設

- 證據等級：`exact`（DOS `overlay-02` 的 `03CAh`（52 條）、`0801h`（22 條）與
  `2086h` 前段逐條讀完）
- 旁白怎麼依距離挑見 spec 1144；選項與結果碼見 spec 1041
- 對照表：`docs/audit/ecl-opcode-handlers-dos.md`

## 兩格工作記憶體

| 原作 | ECL 格 | 內容 |
|---|---|---|
| `bank1^[580h]` | `7EC0h` | **距離上限**，由 `0Ch`／`29h` 的運算元 2 寫入 |
| `bank1^[582h]` | `7EC1h` | **當下距離**，由地圖座標算出再被上限夾住 |

位址由 `bank1^[X] ＝ ECL 格 0x7C00 ＋ X ÷ 2` 換算；三個已知點
（`5C4h`→`7EE2h`、`5C2h`→`7EE1h`、`550h`→`7EA8h`）都對得上。第四個佐證是
`ECL4/0x20` `+00B5h` 的 `SAVE 00 7EC0`——腳本自己會把距離上限歸零。

## `0Ch SETUP MONSTER`（`03CAh`）

```pascal
DS:7601h      := ADDRESSVALUE(1);          { sprite }
bank1^[580h]  := ADDRESSVALUE(2);          { 距離上限 }
DS:7602h      := ADDRESSVALUE(3);          { picture }
bank1^[582h]  := entry#6(720Fh, 7210h, 7211h);        { 從隊伍座標與朝向算距離 }
if bank1^[580h] < bank1^[582h] then bank1^[582h] := bank1^[580h];
entry#8(8B48h, byte(bank1^[582h]), DS:7602h, DS:7601h);   { 畫出來 }
```

## `0Dh APPROACH`（`0801h`，整支 22 條）

```pascal
if bank1^[582h] > 0 then begin
    dec bank1^[582h];
    entry#8(8B48h, byte(bank1^[582h]), DS:7602h, DS:7601h);
end;
```

距離已經是 0 就**什麼都不做**（`0808h` 的 `cmp ... 0 / jbe`）。

## ★ `29h` 會把 `APPROACH` 走過的距離蓋回去

`29h` 進門就重跑一次跟 `0Ch` 一模一樣的那段（`20D1h` 寫上限，`2177h`..`21B5h`
從地圖算再夾住）。

⇒ **`APPROACH` 的減一只影響「畫面上怪物離多遠」，不影響遭遇選單。** 
`SETUP MONSTER / APPROACH / DELAY / APPROACH`（spec 309 的哈普伊弗利特）是
一段**演出**：怪物一步一步走近。等選單開起來，距離重新從地圖算。

這一點反直覺，而且兩種讀法都自洽——只有把 `29h` 的前段也讀完才分得出來。

## remake

`internal/ecl/encounter_prompt.go` 的 `InitEncounterDistance` 與
`ApproachEncounter` 是唯一的兩個入口，`0Ch`、`0Dh`、`29h` 都走它們。

⚠ **remake 沒有「從地圖算距離」的模型**（原作走 `overlay-07 entry#6(座標, 朝向)`），
所以當下距離直接取上限；在這個近似下「被上限夾住」那一步永遠成立，不是被省略。

⚠ **`29h` 的重畫迴圈仍未接**：原作按 `ADVANCE` 會距離減一、重新出選單並換旁白
（spec 1144），remake 選完就回結果。

## 明確不宣稱

- 沒有宣稱 `overlay-07 entry#6(座標, 朝向)` 怎麼從地圖算出距離。
- 沒有宣稱 `entry#8(8B48h, 距離, picture, sprite)` 畫出來的是什麼版面。
- 沒有宣稱 `DS:7603h` 的語意（`1Ch` 把它設成 8，見 spec 1145）。
