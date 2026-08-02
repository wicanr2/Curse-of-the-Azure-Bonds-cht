# 第 438 輪：PC-98 `IDLIST` 到 stable fighter ID 投影

狀態：`READY`（身份與 footprint transaction exact／material-reconstructed；
正常 Sleep UI 仍受 `TACTICALMAP TD/TDEF` frontend projection 阻擋）

## Exact 證據

`GAME.EXE` Borland debug table 提供下列正式符號與型別：

| symbol | 地址 | type | 結論 |
|---|---:|---:|---|
| `CHARACTERLIST` | `0C29:9598` | 991 `CHARRECPTR`，4 bytes | linked-list head |
| `IDLIST` | `0C29:9DD3` | 384，`0x120` bytes，element type 991 | 72 筆 `CHARRECPTR` |
| `LASTOBJECT` | `0C29:9740` | 8，1 byte | 最後一基底 object ID |
| `OBJECTLIST` | `0C29:9741` | 1159，`0x120` bytes，4-byte element | 72 筆 map objects |

type 992 正式名稱是 `CHARREC`、大小 `0x1A7`；member table 102..180 是完整
角色欄位，member 163 正式名稱 `NEXT`、型別 `CHARRECPTR`。overlay 10 builder
從 `CHARACTERLIST` 開始，每接受一個 record 便：

- 把同一 far pointer 寫進 `IDLIST[objectID-1]`（指令以
  `9DCFh + 4*objectID` 表示）。
- 把 objectID 本身寫進 `OBJECTLIST[objectID-1].+2`。
- 從角色紀錄與 combat map 寫入 X／Y／footprint-active。
- 沿 `CHARREC.NEXT`（raw offset `18Ah`）前進並遞增 `LASTOBJECT`。

因此 object ID 到角色身分的橋不是名稱、顯示字串或陣列位置推測，而是原作
`IDLIST` 的 pointer identity；等級標為 exact。remake 沒有 DOS far pointer，
以建立 `CHARACTERLIST` 的同一有序 stable fighter ID transaction 重建一基底
`LegacyObjectID`，屬 material-reconstructed。

## Remake transaction

- `StartCombat` 依 party（含臨時盟友）後 enemies 的明確 linked-list 等價順序
  重建 `LegacyObjectID=1..N`；呼叫者傳入的舊值會被重建覆蓋。
- 原作 `OBJECTLIST／IDLIST` 上限是 72；超過即拒絕戰鬥。
- `Battle.LegacyScanObjects` 只用 `LegacyObjectID`，依 `CombatSize & 7`
  展開 footprint；死亡或未配置位置相當於 inactive object，不進候選。
- `BuildLegacyScanTargetIDs` 串接 `INARC → LOSEXISTS → scanorder`，最後透過
  同一 identity table 回傳 stable fighter IDs。
- 缺 ID、重複 ID、ID 超過 72、未知 source 或 source inactive 全部
  fail-closed；測試不依中文名稱或 slice identity。

## 未完成邊界

目前 Ebiten frontend 的 `combatLineTerrain` 只提供 line spell 所需的 valid／
reflect，沒有 PC-98 `TACTICALMAP.TD` 的一基底 tile 與 `TDEF.LOS／SYM`。把它
假裝成 SCAN map 會使牆、門與大型 footprint LOS 失真。因此本輪完成身份
transaction，但不開放正常手動／Quick Sleep；下一步必須解碼並投影目前戰場
的 TD/TDEF，再以 PC-98 固定戰場 runtime trace 驗證 wall／corner。

原始 executable／overlay 保持唯讀；語意均為附加文件與 remake 欄位，沒有
rename 或覆寫原始位置。第 437 輪的 `/tmp/coab-ida-437` 與
`scripts/ida/pc98_objectlist_audit.idc` 是可重現指令入口。
