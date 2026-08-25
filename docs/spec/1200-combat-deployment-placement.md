# 1200 — 戰鬥佈陣：兩隊的起點、模板與找空位

- 狀態：`READY`（演算法與資料表已解讀；第 712 輪接進 remake，見「七、remake 接線」）
- 證據等級：`exact`（PC-98 COMPREP＝overlay-10 逐指令讀完：`17EEh..1C18h`
  佈署主流程、`1364h` 找空位、`1220h` 放格子、`1C28h` INITCOMBAT；
  DSEG 資料表逐 byte 取自 `pc98-dseg-dseg.bin`）
- 上游：spec 086（八方向 delta）、spec 089（team start）、round-86/88（reference
  公式）、spec 1119（TACMAP entry#19 障礙判定）、spec 1146（遭遇距離）
- ⚠ DOS 側依台帳為同源同構（overlay-10 多數函式與 PC-98 助憶碼序列完全相同，
  差異只有資料段位址）；本規格的位址一律是 PC-98 的。

## 一、整體流程（`17EEh`，由 `INITCOMBAT 1C28h` 呼叫）

```pascal
FIGGOODBAD();                       { 014A:2807：數兩隊人數 → DS:[-5FD6h+隊] }
for id := 1 to 255 do 佈署[id].size := 0;   { 步幅 4 的佈署記錄清空 }
CALCWHOISWHERE();                   { 0189:06D7 }

{ 兩隊的起點（隊 0 ＝ 我方、隊 1 ＝ 敵方）}
teamX[0] := 0;  teamY[0] := 0;
dirGroup[0] := A2ABh div 2;         { A2ABh ＝ 地圖朝向 0..7 }
teamX[1] := 遭遇距離 × dx[A2ABh];   { 距離 ＝ player^[582h]，SETUP MONSTER 寫的 }
teamY[1] := 遭遇距離 × dy[A2ABh];   { dx/dy ＝ DS:489Eh／48A7h（spec 086 的表）}
dirGroup[1] := ((A2ABh + 4) mod 8) div 2;
attempts[隊] := (隊伍人數 + 1) div 2;      { 7ABEh[隊] }

{ occupancy 展開：DS:78AAh[隊][組0..3][列0..5][欄0..10] }
for 隊 := 0 to 1 do
  for 組 := 0 to 3 do begin
    形狀 := if 組 = 1 then 4 else dirGroup[隊];
    for 列 := 0 to 5 do for 欄 := 0 to 10 do
      occ[隊][組][列][欄] := (欄 >= lo[形狀][列]) and (欄 <= hi[形狀][列]);
  end;

{ 逐戰鬥員放格 }
n := 0;
for f in 戰鬥員鏈（9598h 起） do begin
  隊(7AC8h) := f^.side(+198h);
  佈署[id].id := id;  佈署[id].size := f^.field_DE and 7;
  if 找空位(id) then begin          { 1364h }
    if f 不是隱形類(+197h=0 等三個閘) then begin
      n := n + 1;
      戰場圖[y×50 + x + 7] 的符號存回 [-610Eh+n×7] 後改寫成 1Fh;
      icon 表[n]（步幅 7）:= (f, x, y);
    end;
  end else begin
    佈署[id].size := 0;
    if f^.+18Eh^.+13h = 1 then 把 f 從戰鬥移除（00E9:002F(0,1)）;   { ★ 放不下就不上場 }
  end;
  CALCWHOISWHERE();
end;
```

`INITCOMBAT` 收尾拿**第一個戰鬥員**的位置定相機：
`map+2 := FINDX(第一人) − 3`、`map+3 := FINDY(第一人) − 3`（0189:12EBh／1313h）。

## 二、佈署模板：DSEG `702h` 的「每列欄範圍」，不是任意格

5 種形狀 × 6 列 × (lo, hi)（`lo > hi` ＝ 該列沒有格）：

| 形狀 | 用於 | 列 0..5 的 [lo..hi] |
|---|---|---|
| 0 | dirGroup 0（北）| —, —, —, [2..9], [3..10], [4..10] |
| 1 | dirGroup 1（東）| [0..2], [0..3], [1..4], [2..5], [3..6], [4..7] |
| 2 | dirGroup 2（南）| [0..6], [0..7], [1..8], —, —, — |
| 3 | dirGroup 3（西）| [3..6], [4..7], [5..8], [6..9], [7..10], [8..10] |
| 4 | **組 1 專用**（兩隊皆同）| [0..6], [0..7], [1..8], [2..9], [3..10], [4..10] |

佈署區塊是 **6 列 × 11 欄**（`1220h` 的界內判斷寫死 `0<=欄<=0Ah`、`0<=列<=5`）。
每隊有 4 個組（`arg_0`＝0..3），組 1 一律用形狀 4。occupancy 放好一格就寫 0，
**同一格不會放兩個人**。

## 三、放格子（`1220h`，retf 0Ch）

參數 `(組, a2, a4, 列, 欄, id)`，其中 `a4 = teamX[隊]`、`a2 = teamY[隊]`
（`1364h` 由 `7ABAh[7AC8h]`／`7ABCh[7AC8h]` 讀出，可能已被牆面調整位移過）：

```text
通過 occupancy（occ[7AC8h][組][列][欄] <> 0）後：
  x := a2×5 + a4×6 + 欄 + 22        （寫進佈署[id].x，DS:[-68C3h+id×4]）
  y := a2×5 + 列 + 10               （佈署[id].y，DS:[-68C2h+id×4]）
再過地面檢查：TACMAP entry#19（0189:007Fh，spec 1119 的障礙判定）回報
  該格沒被擋、且站著的物件（var_3）在 `48ACh` 表（步幅 4）首 byte <> FFh
成功 ⇒ occ 該格寫 0（占用）。
```

與 round-88 的 reference 公式逐項相符（`+22`／`+10`、×6／×5）；reference 叫
`groupRow` 的那個乘數就是 `teamY`。

## 四、找空位（`1364h`，retf 2，參數＝id）

以「理想格」為中心的**狀態機掃描**（狀態 `var_7` ∈ 1/2/3 交替換方向、
`var_10` 逐圈放大、`var_13` 計步）：

- 掃描方向種子：`var_9 := [6DEh + dirGroup×4 + 組] div 2`；三個狀態各用
  `(var_9+2) mod 4`／`(var_9+1) mod 4`／`(var_9+3) mod 4` 查 `6EEh` 表得軸向，
  再查 `489Eh`／`48A7h` 得步進；初始格由 `6F2h`／`6FAh` 表（依「組>0」選半）
  加上 `attempts×步進` 組出。
- 每個候選過 `11EFh`（「兩個都不在 0..10／0..5 內才回 true」的界內判斷）再叫
  `1220h`。
- 掃太遠的止損：組內步數超過 `attempts[隊]`（首組）或 `0Bh`（後續組）就換組
  （`var_14` 0→3，組 1 會落到形狀 4 的模板）；四組都失敗回 0。
- **地城牆面調整**（`161Fh`，只在隊 0、朝向為偶數、組 0、第一次失敗時做一次）：
  以「隊伍的地城座標 `A2A9h`／`A2AAh` ＋ 該隊 `7ABAh`／`7ABCh` 位移」為查詢格，
  用 `6CEh` 表（4 組 × 4 個方向碼，8＝停）逐方向叫 `378h`（牆面從兩側各查一次，
  回 1＝被擋）；**任一方向走得過**（`1690h` 的 `cmp al,1; jz` 不成立才立 `var_6`）
  就把失敗計數再加一——種子格因此多推一步——然後重新開掃。
- **組升級的方向選擇**（`16AFh`，完全出界且步數用盡時）：往下一個組之前，
  查該組在 `6CEh` 表的方向；**走得過**（`174Ch` 的 `al = 1` 就跳過該組）才把
  `teamX`／`teamY` 往 `489Eh`／`48A7h` 的該方向位移並取用該組。
- `7F27h = 3`（非地城模式）兩處都跳過 `378h` 查詢，直接走成功分支：
  首次失敗一律多推、組升級一律取用。

方向掃描表（DSEG，逐 byte）：

```text
6CEh: 08 04 06 02 | 08 06 04 00 | 08 00 06 02 | 08 02 00 04   （牆檢方向，8＝終止）
6DEh: 00 00 02 06 | 02 02 00 04 | 04 04 02 06 | 06 06 04 00   （÷2 → 掃描方向組）
6EEh: 07 02 03 06                                             （軸向重映射）
6F2h: 05 04 05 06 03 08 07 02                                 （初始格，組=0 半）
6FAh: 03 02 02 03 00 02 05 03                                 （初始格，組>0 半）
```

## 五、全域欄位對照

| 位址 | 意義 |
|---|---|
| `A2A9h`／`A2AAh`／`A2ABh` | 隊伍的地城 X／Y／朝向（0..7）|
| `7F09h^[582h]` | 遭遇距離（`0Ch` 擺、`0Dh APPROACH` 減、spec 1146）|
| `7ABAh..7ABBh`、`7ABCh..7ABDh` | teamX[2]、teamY[2] |
| `7ABEh..7ABFh` | attempts[2] ＝ (人數+1)÷2 |
| `7AC0h..7AC1h` | dirGroup[2] |
| `7AC8h` | 佈署中的「目前隊」索引（0/1）|
| `78AAh` | occupancy：2 隊 × 4 組 × 6×11，非 0＝空位 |
| `f^.+198h`／`+0DEh and 7`／`+197h` | 戰鬥員的隊別／size／隱形類旗標 |
| 戰場圖 | 一列 50 bytes（`×32h`），格 +7 是符號；放人寫 `1Fh` |

## 六、明確不宣稱

- 沒有宣稱 `48ACh` 步幅 4 記錄的完整語意（只用到「首 byte = FFh ⇒ 拒絕」）。
- 沒有宣稱 `+197h`／`0BDE8h`／`+18Eh^.+13h` 三個閘各自的完整語意——只知道
  它們決定「放好之後要不要畫 icon」與「放不下時要不要移除」。
- 沒有宣稱 `6EEh`／`6F2h`／`6FAh` 表的**設計意圖**（為什麼是這幾個值）；
  remake 照表轉錄即可。
- DOS 側未逐指令重讀（台帳記載兩平台此模組同構、差異僅資料位址）；
  若行為對不上，先回來比對 DOS overlay-10 同位置的 bytes。

## 七、remake 接線（第 712 輪）

`internal/combat/deployment.go`（`NewDeployment`／`Place`）接進
`StartCombat`（`internal/game/combat_state.go`），輸入對照：

| 原作 | remake |
|---|---|
| `A2ABh` 地圖朝向 | `State.DungeonDirection` |
| `player^[582h]` 遭遇距離 | session 記憶體 `EncounterDistanceCell`（`0x7EC1`，spec 1146）|
| TACMAP entry#19 地面檢查 | `combatMovementTerrain`（移動地形投影）＋ 戰鬥地圖邊界 50×25 |
| `378h` 牆面查詢 | `geo.Grid.CanMoveDungeonWrapped(x, y, dir)`，true＝走得過（非地城交 nil ＝ 原作跳過查詢：首次失敗一律多推、組升級一律取用）|
| 放不下 → 移出戰鬥 | 該戰鬥員不加入 `Battle` |

- 產出的是**原生戰鬥地圖座標**（隊伍區塊 x≈22..32、y≈10..15，與
  `mapdata.DungeonFloor` 的窗口公式同族），整場戰鬥切到 reference 座標模式。
- **只有「全部戰鬥員都沒有預置座標」的開戰**（ECL 遭遇的正常路徑）走佈署；
  帶預置座標的呼叫（存檔還原、測試自組）維持呼叫端的座標命名空間與
  `combatFormationTile` fallback——混用會把兩個座標系疊在同一場戰鬥。
- 沒有 session 的呼叫端（單元測試自組戰鬥）距離取 1；有 session 照實讀，
  含腳本寫 0（兩隊同一個區塊、貼臉遭遇）——下水道巨魔房實測距離 0，
  佈出來的形狀與原版擷取定性一致（巨魔貼臉一排、鱷魚第二排）。
