# PC-98 睡眠術受傷解除

狀態：`READY`

## 結論

PC-98 `EFFECTS` overlay 23 的 `PUTDAMAGE 1FFDh` 只在實際正傷害路徑呼叫
`REMOVEFX 158Ah`。`REMOVEFX` 以 `DS:[159Dh+index]` 讀取 19 個效果編號，
其中包含 `35h`（Sleep），再逐一呼叫 `SPELLOFF 010Eh` 解鏈並釋放九 byte
`EFFECTREC`。因此睡眠中的戰鬥者受到至少一點傷害後會醒來；零傷害不會走
此解除路徑。

## 位址空間與原始證據

| 證據 | 原始位置 | 結論 | 等級 |
| --- | --- | --- | --- |
| Borland symbol | overlay module `013E`：`PUTDAMAGE 1FFD`、`REMOVEFX 158A`、`SPELLOFF 010E` | 函式身分與 local offset | `exact` |
| IDA 9.4 連續指令 | overlay 23 `22A9h..22B3h` | 正傷害路徑呼叫 `REMOVEFX` | `exact` |
| IDA 9.4 連續指令 | overlay 23 `1590h..15B8h` | 迴圈 1..19，由 `DS:[159Dh+index]` 取效果編號並呼叫 `SPELLOFF` | `exact` |
| resident raw bytes | `GAME.EXE` MZ file `E1CEh..E1E1h`，即 `DS:159Eh..15B1h` | 表列 `07 0B 0D 15 17 1E 1F 20 33 34 35 3A 3B 5F 62 88 89 8B 90`，含 `35h` | `exact` |

`[di+159Dh]` 使用預設 `DS`，不是 overlay 的 `CS:159Dh`。這個區分是本輪
必要校正；讀 overlay 同數字位置會得到錯誤表格。原始 executable、overlay
與既有 IDA database 均未修改；audit 只在 `/tmp` 的副本執行，語意以本規格
附加保存。

輸入雜湊：

- `PC98-GAME.EXE`：`8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0`
- code-only overlay 23：`a3ea0d9528be57a92c33fc345baa3e27eef375c84822afba0cfbb141c2faabc9`

## 實作邊界

`Battle.applyPositiveDamage` 在扣除正傷害與中斷待施法後，移除動態 `35h`
效果。`MON*SPC` 載入的 innate record 並非 `PUTEFFECT` 建立的動態 spell
record，本輪不把它當成可解除睡眠；其他 18 個 `REMOVEFX` 編號也要逐一
證明 gameplay projection 後再接入，不能只憑表格名稱猜語意。

## 尚未完成

- `duration=5×caster level` 的實際時間／回合遞減 consumer；
- 戰鬥結束、存讀檔與玩家角色效果鏈的完整生命週期；
- 醒來文字、twinkle 與聲音的原版時序；
- DOS／PC-98 實機畫面的受傷醒來動態對照。
