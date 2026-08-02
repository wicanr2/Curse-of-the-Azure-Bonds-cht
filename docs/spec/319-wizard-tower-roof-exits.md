# 第三百一十九輪：法師塔屋頂出口與祕道

狀態：`READY`

## 反組譯證據

- GEO5 block `0x33` 唯一 terrain `0x01` 位於 `(7,15)`。per-turn entry
  `+0x07D7` 比較 terrain `1`，並要求 `C04D==1`（renderer 面東）後導向
  `+0x07EA`。文字說明此處可下到洞穴，也發現能直達荒野的祕道。
- `+0x085D` ordinary HORIZONTAL MENU 不是兩項，而是三項：
  `CAVES / WILDERNESS / STAY HERE`。
- CAVES：
  - 寫入 `C04B/C04C/C04D = 6/15/0`；
  - `NEWECL 0x32`，回到熔岩洞。
- WILDERNESS 不應由 State 看到 label 後立即 `enterMap()`；原 ECL 先顯示
  `VILLAGE / DEPART` 第二層 menu：
  - VILLAGE → `NEWECL 0x31`，返回哈普圖斯村；
  - DEPART → `NEWECL 0x30`，交給 Area 5 外部旅程 dispatcher。
- STAY HERE 直接 EXIT，保留 block `0x33` 與塔頂地城。

## 引擎與中文契約

- 普通 menu label `WILDERNESS` 仍不等於 engine action。Shadowdale 等城市的
 作品 adapter 可把特定 WILDERNESS 選擇投影成 map，但 block `0x33` 的同名
 label 必須先 resume script，否則會吞掉 VILLAGE／DEPART 與 NEWECL writes。
- NEWECL source registers 必須在 target initial entry 前保存；CAVES 的
 source write 是 `(6,15,N)`；完整 session 接著執行 block `0x32` initial entry
 後成為 `(6,15,E)`。frontend 不可猜成法師塔入口 `(7,15,W)`。
- STAY HERE 是真正的取消選擇，不可清除地城 picture／piece／position state。
- 三項第一層與兩項第二層均使用 24px 繁中敘事及可容納中文字的選單版面；
  原始 dungeon 圖塊維持 640×480 nearest-neighbor 整數放大。

## 驗收

- real-image regression 鎖定四條 effective outcomes：
  CAVES→`0x32`、VILLAGE→`0x31`、DEPART→`0x30`、STAY HERE→EXIT。
- 驗證 CAVES source registers `6/15/0`。
- State 從真實 GEO `(7,15,E)` 觸發三項繁中 menu；STAY HERE 留在塔頂。
- 重訪後依序驗證 CAVES 返回熔岩洞，以及 WILDERNESS 先出現
  VILLAGE／DEPART，不被泛用 map adapter 提前攔截。
- `-wizard-tower-exit` 經正式 Ebiten／Xvfb 產生
  [`wizard-tower-roof-exit.png`](../screenshots/wizard-tower-roof-exit.png)：
  固定 640×480 logical canvas，24px Noto CJK 正文與選項，未使用 DOS 8px
  英文字格或平滑縮放。

更新：第 451 輪已把屋頂出口與荒野第二層提示移入 CoAB JSON；兩層 menu
tokens 與 NEWECL writes 仍由原 ECL 驅動。見
[`spec 451`](451-wizard-tower-branches-json.md)。
