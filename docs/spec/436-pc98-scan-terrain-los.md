# 第 436 輪：PC-98 `SCAN` 地形視線與候選 producer

狀態：`READY`（限 `TDEFTYPE`／`TACTICALMAP` 投影、`LOSEXISTS`、footprint
最短距離與三 byte record producer；`INARC` sector、COMPOBJ 建立順序、原版
固定戰場動態 trace 與正常玩家 Sleep UI 仍待續）

## 結論

第 435 輪已關閉三 byte record 與排序，但仍把 terrain property 名稱列為未知。
本輪以 PC-98 Borland 0x52FB symbol／type／member table、全新 IDA 9.4 raw data
report、overlay 31 連續指令及 producer→consumer 資料流交叉驗證後，得到：

- `TDEF 0C29:48B0` 是 65 筆、每筆四 byte 的 array；element type
  `TDEFTYPE` 大小四，Borland 正式 member 依序為 `HT／LOS／SYM`。第四 byte
  沒有 member 名稱，雖然 raw table 有非零規律值，仍只能保存成 `Raw3`，不可
  自行命名。
- `COMBATMAP 0C29:9F2C` 的型別是 `TACMAPPTR → TACTICALMAP`；後者大小
  `04E9h = 7 + 50×25`。member 依序是 `H／V／VX／VY／CURSORON／
  CURSORSIZE／XRAY／TD`，所以 runtime map `+6=XRAY`、`+7=TD[0,0]`。
- `TD` tile byte 是一基底 `TDEF` index。overlay 31 的
  `[48ADh + 4*tile]` 正好是 `TDEF[tile-1].LOS`；`[48AEh + 4*tile]` 是
  `TDEF[tile-1].SYM`。
- `LOSEXISTS 03EAh` 以 source→target 的 Bresenham-like vector 逐格走訪；
  cardinal step 的 metric 加二，diagonal step 加三。每一步在 `XRAY=0` 時
  要求目前 cell 的 `SYM <= source cell 的 LOS`，且 metric 必須
  `<= 2*range+1`。任一條件失敗會回傳 false、停止 cell 與當前 metric；抵達
  target 則回傳 true 與 metric。
- `XRAY` 只略過 `SYM／LOS` 比較，不略過 range budget。不能把它實作成無限
  range，也不能把 `SYM` 在沒有其他 consumer／runtime 證據時改名為「牆高」。

以上只關閉靜態地形資料流；尚未取得 PC-98 固定戰場的動態 wall／corner trace，
所以畫面層仍不得宣稱整個戰鬥 LOS 已完成。

## 非破壞性輸入與報告

| 輸入／報告 | SHA-256 | 用途 |
|---|---|---|
| `PC98-GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | Borland symbols／types／members、resident data |
| overlay 31 | `6cd5e38dddeb1ea5ddc44dd7f3af68c49bd9b67198e7c42247f1d08743050081` | `STARTVECTOR／STEPVECTOR／LOSEXISTS／SCAN` |
| `/tmp/coab-ida-436/pc98-scan-terrain-globals.txt` | `8138250f3420aad96f0d28243ec2845f5c51e6fd778704265addf92fa48a9553` | 65 筆 TDEF raw bytes、三個 resident globals |

原始 EXE、overlay 與基準 `.i64` 全部唯讀保存。IDC 只在
`/tmp/coab-ida-436` 的副本運行，不 rename database。第一次命令雖 exit 0
卻沒有 report；第二次 report 只有 `BADADDR`，兩者均按失敗即關閉規則拒收。
最後一次報告另驗證非空、恰有 65 行 `TDEF_RECORD`，並保留：

```text
resolution=borland_segment_0C29 ida_linear_base=0001C290 segment=dseg
symbol=TDEF original=0C29:48B0 ea=00020B40 ida_name=
symbol=COMBATMAP original=0C29:9F2C ea=000261BC ida_name=
symbol=LASTSIGHT original=0C29:9F30 ea=000261C0 ida_name=
```

空白 `ida_name` 是刻意結果：語意來自外部 Borland symbol table，沒有覆蓋 IDA
原始 database 名稱。

## `LOSEXISTS` exact 資料流

### Stack contract

`SCAN 0AA2h..0AC4h` 依 Pascal 左至右壓棧呼叫 `LOSEXISTS`：

- `[bp+18h／16h]`：source X／Y value；
- `[bp+12h／0Eh]`：target X／Y in-out far pointers；
- `[bp+0Ah]`：metric in-out far pointer；
- `[bp+06h]`：`TACTICALMAP` far pointer。

函式 `retf 14h` 與二十 bytes 參數完全吻合。不能用 C 右至左參數直覺重排。

### Terrain 與距離

- `0432h..0454h`：source `(x,y)` 讀 `TD[y*50+x]`，再讀
  `TDEF[tile-1].LOS`。
- `0457h..046Dh`：取 `max(abs(dx),abs(dy))`。
- `046Dh..0495h`：再次讀同一 source cell 的 `LOS`，建立從
  `(0,LOS)` 到 `(maxDelta,LOS)` 的第二 vector；因此 sight threshold 在這個
  consumer 是 source `LOS` 的常數，不是 target interpolation。
- `04A5h..04ADh`：`TACTICALMAP.XRAY != 0` 時跳過 property gate。
- `04AFh..04D8h`：目前 primary cell 的 `SYM` 大於 secondary current Y
  （即 source `LOS`）就失敗。
- `04DAh..04E2h`：當前 metric 大於 `2*range+1` 就失敗。
- `0507h..052Dh`：兩個 vector 各 step 一次；secondary 已抵達 end 時成功，
  否則回到 property／range gate。primary 與 secondary 都走
  `max(abs(dx),abs(dy))` 次，故同一 iteration 結束。

## Remake 對應

獨立 engine 新增 `combat/scan`：

- `TacticalMap` 保存一基底 TD、`TDEFTYPE` 與 `XRAY`；
- `LineOfSight` 保存原始 2／3 metric、inclusive budget、停止 cell 與錯誤邊界；
- `Build` 展開 adapter 提供的 source／candidate footprint cell pairs，保留第一個
  strict minimum，再建立三 byte record 並呼叫既有 `scanorder.Sort`；
- object ID、footprint 與 direction resolver 都是強制輸入。缺少 direction、
  tile 0、越界 cell、重複／零 object ID 一律失敗，不用現代預設值猜測。

CoAB `BuildScanTargetIDs` 把 explicit legacy object ID 映射到 stable fighter ID，
並已有 `terrain producer → ordered stable IDs → CastSleepOrdered` 的 bounded
transaction 回歸。測試不複製中文顯示文字，也不把 slice index 當 object ID。

## 驗證

- engine：Docker／`--network none` 的 `go test -count=1 ./...` 全數通過。
- CoAB focused：`internal/combat` 與 `internal/game` 在 local workspace 指向同一
  engine HEAD `9c94ddd` 時通過。
- CoAB formal：先由 `go list -buildvcs=false` 確認 31 個 import paths，再於
  Docker／Xvfb／`--network none`／`GOPROXY=off` 逐套件執行
  `go test -buildvcs=false -count=1`。去重後 31 行 package pass ledger 為
  `/tmp/coab-ida-436/package-pass-436-unique.txt`，SHA-256
  `04e7594f05582c1e8a5f16d4c8c8b2b1e532a25e9fe2831c976593ebe5d7bf7b`。
- 正常玩家 Sleep、PC-98 動態 wall／corner 與 cursor 尚未完成；上述 gate 只
  支持本 spec 的 bounded producer／transaction，不支持「完整戰鬥」聲明。

## 未完成與下一步

1. 逐指令關閉 `INARC 054Ah..08D8h` 八個 sector 的邊界與 direction payload；
   目前 engine 要求 adapter callback，不以 `atan2` 猜測。
2. 追 COMPOBJ builder，證明 party reserved slots、monster／temporary ally 與一基底
   object ID 的建立順序；在此之前正常 Battle 不能用 `fighterOrder+1` 代替。
3. 若 PC-98 VFD 可重現，擷取固定戰場 wall／corner／large footprint 的 SCAN
   records；若媒體缺 sector，必須如實保留靜態 exact／動態未驗，不虛構 trace。
4. 上述兩項閉合後才把手動／Quick Sleep 接進正常玩家格點選擇、法術格消耗、
   effect lifecycle、繁中訊息、動畫與音效。
