# 1144 — `29h ENCOUNTER MENU` 的三句旁白依距離挑

- 證據等級：`exact`（DOS `overlay-02:2086h`..`2784h` 逐條讀出，649 條）
- 作法見 spec 783；同一支函式的選項與結果碼見 spec 1041
- 查詢工具：`cmd/ecl-encounter-text`

## 結論

`29h` 帶三句旁白（運算元 9、10、11），是同一場遭遇的**近／中／遠**三種描述。
原作依**當下的距離**決定從哪一句開始看，再在三句之間**循環**往下找第一句非空的。

| 距離 | 從第幾句開始 | 循環順序 |
|---:|---:|---|
| 0 | 1 | 1 → 2 → 3 |
| 1 | 2 | 2 → 3 → 1 |
| 2 | 3 | 3 → 1 → 2 |

## 一手證據

進門先把三句旁白從解譯器的文字槽搬進區域緩衝區，`i = 1..3`，stride `100h`：

```asm
2131  mov  di, ax          ; i
2135  shl  di, cl          ; i × 100h
2137  add  di, 7648h       ; 來源：DS:7648h + i × 100h
2149  lea  di, [bp+di+var_406]   ; 目的：SS:var_406 + i × 100h
2153  call far ptr 0A54h:64Eh    ; 上限 0FFh
2158  cmp  [bp+var_409], 3
215D  jnz  short loc_2127
```

⇒ **區域槽 1..3 是三句候選，槽 0（`var_406`）是選中的那一句。**

接著讀距離並三選一：

```asm
2229  mov  ax, es:[di+582h]      ; bank1^[582h] ＝ 當下距離
222E  cmp  ax, 0
2231  jnz  short loc_226E
2233  mov  [bp+var_43B], 1       ; 距離 0 → 從第 1 句開始
...
2259  inc  [bp+var_43B]
225D  cmp  [bp+var_406], 0       ; 選中的還是空的？
2262  jnz  short loc_226B
2264  cmp  [bp+var_43B], 3
2269  jbe  short loc_2238        ; 還沒看完就繼續

226E  cmp  ax, 1
2273  mov  [bp+var_43B], 2       ; 距離 1 → 從第 2 句開始
229D  cmp  [bp+var_43B], 3
22A2  jbe  short loc_22A9
22A4  mov  [bp+var_43B], 1       ; ★ 繞回第 1 句
22B0  cmp  [bp+var_43B], 2       ; 繞回起點就停

22B9  cmp  ax, 2
22BE  mov  [bp+var_43B], 3       ; 距離 2 → 從第 3 句開始
22EF  mov  [bp+var_43B], 1       ; ★ 同樣繞回
22FB  cmp  [bp+var_43B], 3
```

選中的那一句最後餵給 `542h:4A4h` 畫出來（`2325h`）。

## 語意上的獨立佐證

猶拉什地下（`ECL3/0x12` `0x0523`）那一處的三句是：

1. `SOME SHAMBLING MOUNDS ATTEMPT TO PUSH YOU BACK INTO THE CORRIDOR.`（貼身）
2. `SHAMBLING MOUNDS ATTEMPT TO GET AWAY.`（中距）
3. `YOU SEE SHAMBLING MOUNDS IN THE DISTANCE.`（遠距）

近／中／遠與「距離 → 第幾句」的對應完全吻合，兩條證據互相印證。

## 距離本身

`bank1^[582h] := <overlay-07 entry#6>(DS:720Fh, DS:7210h, DS:7211h)`（隊伍座標與
朝向），再被運算元 2（`bank1^[580h]`，距離上限）夾住。選 `ADVANCE` 會 `dec` 它，
減到 0 時第四個選項換成 `PARLAY`（spec 1041）。

⚠ 距離大於 2 時 `2229h` 沒有對應分支，會落到 `2302h` 讀沒寫過的緩衝區。
全遊戲 20 處的距離上限實測是 **16 處為 2、4 處為 0**，所以那條路走不到。

## `ADVANCE` 逐次接近時，旁白會跟著換

函式尾巴是一個重來旗標：

```asm
2768  cmp  [bp+var_40C], 0
276D  jz   short loc_2772
276F  jmp  loc_21D5          ; ★ 重來
```

`var_40C` 由「距離減一並重畫」那幾支設成 1（`2493h`／`250Bh`／`262Ah`／`26B6h`／
`272Eh`），而 `loc_21D5` **在挑旁白（`2225h`）之前**。

⇒ **每重畫一次就重新挑一次旁白，而距離已經少了一。** 玩家按 `ADVANCE` 逼近時，
敘述會從遠距那一句換成中距、再換成貼身那一句。三句是同一場遭遇的三個階段，
不是三個備案。

## remake 目前的近似

remake 沒有「從地圖算距離」的模型，所以拿**距離上限**當當下距離
（`ADVANCE`／`PARLAY` 的判斷本來就已經是同一個近似）。挑法本身照抄，寫在
`internal/ecl/encounter_prompt.go` 的 `EncounterPromptSlots`，執行期與盤點工具
共用同一份，避免兩邊各寫一次而漂開。

⚠ **重畫迴圈還沒接**：remake 的 `29h` 選完就回結果，不會「距離減一、重新出選單」。
所以旁白只會挑一次——挑的那一句是對的，但按 `ADVANCE` 看不到它逐步逼近。

## 明確不宣稱

- 沒有宣稱 `overlay-07 entry#6(座標, 朝向)` 怎麼從地圖算出距離。
- 沒有宣稱三句都空的時候畫出來的是什麼。
