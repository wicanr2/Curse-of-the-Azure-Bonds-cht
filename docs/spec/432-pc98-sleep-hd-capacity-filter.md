# 第 432 輪：PC-98 睡眠術 HD 容量篩選

狀態：`READY`（限全域法術 `15h` dispatch、`4d4` 容量與 ordered HD filter）

## 結論

PC-98 全域法術 `15h` 的專屬處理器先擲一次 `4d4` 容量，再依既有目標陣列
順序扣除 HD 成本：HD `<=1／2／3／4／5／>=6` 分別花費
`1／2／4／6／10／20`；五 HD 目標的 record `+74h` 非零時也花費 `20`。
已有 raw effect `35h` 或剩餘容量不足的目標會從陣列清掉，但掃描不會因此
中止，後面的較低成本目標仍可能入選。

本輪沒有替 record `+74h` 猜 gameplay 名稱，也沒有從 AD&D 紙本規則補上
尚未閉合的目標幾何、豁免、解除、動畫或聲音。remake 因此只新增作品中立的
容量 primitive，尚未把睡眠術列為可在戰鬥 UI 選用的完整法術。

## 非破壞性輸入

| 輸入 | SHA-256 | 用途 | 等級 |
|---|---|---|---|
| `PC98-GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | 全域 spell record、Borland symbols、resident stubs | `exact` |
| `PC98-GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | typed TPOV resolution | `exact` |
| overlay 22 | `c54729525d576c11d731d64a1b06ee2547b2562b73e3708a1beaafc535cabbe8` | `INITSPELLS`、Sleep handler、共用 effect writer | `exact` |

原始檔與既有 database 全程唯讀。`scripts/ida/pc98_sleep_budget_audit.idc`
只在 `/tmp/coab-ida-432b` 的 overlay 副本建立附加 ledger，保留 local offset、
raw bytes 與原始 disassembly，不 rename、patch 或回寫來源。

## Dispatch 與 spell record

全域 spell table 的 16-byte record 公式沿用 spec 423。`15h @ GAME.EXE file
12F34h` 是：

```text
02 01 03 04 00 05 09 00 00 04 35 01 01 02 01 01
```

其中 `+0Ah=35h` 是共用 effect writer 使用的 effect kind；其餘未閉合欄位
不因排列相似就自行命名。

overlay 22 `INITSPELLS 66EBh` 以四 byte far-pointer slots 建立
`DOSPELL` table。法術第 21 筆的 slot 位於 `A390h + 20*4 = A3E0h`；
`6816h..681Fh` exact 寫入 `0117:00EDh`。typed resolver 將 stub `00EDh`
解析為 overlay 22 entry 41、handler local `2547h`。這條資料流把全域
`15h` 與容量處理器直接閉合，不靠日文字串或法術名稱猜測。

## Handler 指令證據

overlay 22 local `2547h..267Ch`：

1. `2552h..255Dh` 呼叫共同擲骰 helper，參數為 `4,4`，結果保存到
   `7D93h`：exact `4d4`。
2. `2560h..2578h` 依 `NUMSPELLTARGETS A520h`，從 index 1 開始依原順序
   走 `SPELLTARGET A51Dh` 的 far pointers。
3. `2583h..25E9h` 讀 target record `+E5h`，建立成本
   `1／2／4／6／10／20`；五 HD 時另讀 `+74h`，零為 10、非零為 20。
4. `25F4h..2622h` 先排除 target `+196h==1`，再以 effect predicate
   查 raw `35h`；已有同效果者不消耗容量。
5. `2624h..2634h` 只有 `remaining >= cost` 才扣除；失敗則與上述排除分支
   一樣在 `2636h..2647h` 把該 target pointer 清零。迴圈仍繼續掃下一筆。

`CASTSPELL` 會經目前場景安裝的 `DOSPELLTARGETING` function pointer，在專屬
handler 前建立候選陣列；戰鬥中的正確 consumer 是 overlay 13 `225Fh`，不是
預設／非戰鬥的 `GETSPELLTARGETS 112Ch`。這項第 432 輪舊斷言已由 spec 433
supersede。handler 結束後再進共用 effect writer。local `0F62h..1118h` 會讀 spell record
`+0Ah` 並呼叫 effect-list writer，但它同時包含 save、magic resistance、
動畫與 duration 分支。本輪尚未把這些分支對 `15h` 的所有實際參數閉合，故
不能把「effect kind 已知」擴大成完整睡眠術結算。

### 推論等級

- `exact`：record bytes、`15h → A3E0h → 0117:00EDh → entry 41／2547h`、
  `4d4`、HD 成本表、`35h` existing-effect 排除及 ordered pointer clearing。
- `strong inference`：record `+74h` predicate 代表某種五 HD 目標細分；讀取
  與 10／20 分支 exact，但 gameplay 欄位名稱尚無 writer／type 證據。
- `unknown`：原版 target geometry／ordering 的上游完整規則、save 與 magic
  resistance 對 Sleep 的實際開關、effect duration、解除方式及動態演出。

## Engine 邊界與驗證

獨立 engine 新增 `combat/sleep`：

- `RollCapacity` 嚴格消耗四次合法 d4；
- `HitDiceCost` 保存 exact piecewise table，但把 `+74h` 暫稱為中立的
  `DoubleFiveHitDiceCost`；
- `Filter` 保存輸入順序、略過既有 held、容量不足時繼續掃描，並回傳剩餘值。

測試包含全部 HD 邊界、五 HD predicate 正反例、既有效果、前方昂貴目標不
阻止後方低成本目標，以及 RNG 錯誤。CoAB frontend 未接入，因此本規格不宣稱
正常玩家施法路徑、法術格消耗、Action clear、畫面或音訊完成。

## 下一步

1. 沿 spec 433 已定位的 overlay 31 `SCAN`，以 runtime 關閉幾何欄位、
   large footprint 與 tie order。
2. 追共用 writer 對 `15h` 的 magic resistance 與 raw
   `35h` record 參數。
3. 完成後才把全域 `15h` 加入 game-pack metadata、手動／Quick delayed
   cast、效果寫入、held scheduler、繁中訊息、動畫與正常玩家路徑驗收。
