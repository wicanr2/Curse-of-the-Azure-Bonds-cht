# 第 437 輪：PC-98 `INARC` 八方向與 `OBJECTLIST` 邊界

狀態：`READY`（`INARC` exact；`OBJECTLIST` 名稱／容量／主要欄位 exact；
combatant linked-list 到 stable ID 的完整投影仍待續）

## 結論與推論等級

### `INARC`：exact

PC-98 overlay 31 local `054Ah..08D5h` 的原始指令證明：

- 參數依序是 source X/Y、target X/Y、arc；座標界限固定 50×25。
- `FFh` 先改成 8；sector 8 接受所有有效座標。
- `DXDIR／DYDIR` 是 N、NE、E、SE、S、SW、W、NW、self。
- source 等於 target，或 target 是該 sector 的相鄰格時立即成功。
- sector `0..7` 使用八組整數斜線半平面與 inclusive 比較。扇區邊界不得以
  浮點角度、`atan2` 或現代圓形 cone 近似。
- `SCAN 0B2Eh..0B8Ch` 對 caller arc `<8` 直接保存；`8／FFh` 則以最佳
  footprint pair 依序測 `0..7`，保存第一個成功 sector。

Engine `combat/scan` 已保存上述 contract。測試將 production predicate 與
逐跳轉形狀的 reference 對 50×25 全 source、全 target、九個 sector 共
14,062,500 組輸入比對，另驗證 `FFh`、越界與無效 sector 失敗即關閉。

### `OBJECTLIST`：部分 exact，作品投影未完成

`GAME.EXE` Borland debug table 的正式符號是：

| symbol | 地址 | 型別證據 | 等級 |
|---|---:|---|---|
| `LASTOBJECT` | `0C29:9740` | 1 byte | exact |
| `OBJECTLIST` | `0C29:9741` | `0x120` byte array，element 4 bytes | exact |

因此 table 是 72 筆四 byte records；一基底 object ID 直接作 index。overlay
31 `SCAN` 與 overlay 10 builder 共同證明 record `+0=X`、`+1=Y`、`+2` 由
builder 寫入自身 object index、`+3` 是傳給 `CALCBIGOFFSET` 的 footprint／
active byte。builder `19F6h` 從 `LASTOBJECT=1` 開始，沿 combatant far-pointer
linked list 建表；每接受一筆就遞增 object ID／`LASTOBJECT`，不適用者把
`+3` 清零。

「`+2` 可稱 combatant ID」目前只是 strong inference；其 consumer 尚未完整
追完。remake 的 stable fighter ID 也不能直接假定等於 slice index。現有
adapter 因此仍要求明確 `object ID → stable ID` 映射；在 linked-list traversal
與玩家／怪物／臨時盟友順序關閉前，正常 Sleep UI 維持 fail-closed。

## 非破壞性 IDA 與原始證據

原始 overlays 全部唯讀，沒有 rename、patch 或覆寫。語意標籤只存在於新增的
`scripts/ida/pc98_objectlist_audit.idc` 報告；IDA databases 與輸出建於
`/tmp/coab-ida-437`。主要 SHA-256：

- overlay 10：`cc0724159c0cd7dc550e9d3937ec6a2a5e8d290716b3178b9e7a31202f14afe4`
- overlay 31：`6cd5e38dddeb1ea5ddc44dd7f3af68c49bd9b67198e7c42247f1d08743050081`
- overlay 32：`c7e2a2b1166676454a54a69330ea318dae8ef9a22e5c5a7ca6e6950ef3609d93`
- overlay 10 OBJECTLIST report：
  `5b9178e3c1a5e260e02f7930253c886c4593ddfc1e03a5c355265cadb3efbfb1`
- overlay 32 report：
  `3983399c40c6f49638d3541f0a8b410e006aac65fcc3423674d28ef85b956f33`

下一步先追 overlay 10 builder 的 combatant linked-list 欄位與 pointer table，
再將真實 object records 接到 State；不得因 `INARC` 已完成就宣稱正常玩家
Sleep、完整 LOS 或完整戰鬥 targeting 已完成。
