# 第五百五十四輪：眼魔洞穴 LOOK、Dexam 雙戰正常路徑

狀態：`READY`
日期：2026-08-12

## 結論

同一個從新遊戲開始的玩家 session，現在可在取得手札 59 後以普通 GEO 移動走到
`(14,1)`，使用 `LOOK` 揭露東側 `wall=09/detail=0` 通路，進入 `(15,1)`，完成
Dexam、梅杜莎、眼魔與牛頭人第一戰、戰利品／洛山達護符，以及散提爾堡部隊
第二戰，最後回到同一洞穴。沒有設定 PC、座標或注入戰鬥。

本輪尚未閉合 `(15,1)` 到 terrain `93h`／`(6,3)` 的正常出口 route，因此不得
宣稱眼魔洞穴、世界返回或整作主線已完成。

## 證據與等級

| 項目 | 證據 | 等級 |
|---|---|---|
| Dexam handler | ECL4 block `22h:+0B8Ah` 先比較 `4C03h`；值小於 1 時加一，接著 PICTURE 40／49、兩次戰鬥與戰利品 | `exact` |
| shrine writer | ECL4 block `21h:+0832h` 在 hooded woman 分支寫 `4C03h=1` | `exact` |
| 洞穴進入前歸零 | 若沿用 remake 的全域共享 work memory，Dexam handler 必定被略過；原版 map-local variable bank loader 尚未完整閉合 | `strong inference`；以 CoAB JSON 不透明 `set_memory` 重建，不命名欄位 |
| `(14,1,E)` 通路 | GEO4/25 的唯一 Dexam 隔離 edge、手札 59 相對拓撲與正常 remake LOOK 路徑一致 | `strong inference`；尚無原版 SEARCH／LOOK writer→consumer trace |
| Dexam 雙戰 | 原始 ECL／MON4CHA 遭遇組成，加上同一新遊戲正常 session | `exact`（remake 玩家路徑與原始事件）；不含完整怪物能力／演出 |
| 已揭露 edge 雙向可走 | GEO 同一物理邊由兩格方向 byte 表示；正常玩家可原路退出房間 | `exact`（remake contract）／`strong inference`（原版輸入時序） |

## 勘誤

先前 spec 548／550 只記錄 `4C03h=1` 在死精靈分支保持不變，沒有追到
`+0B8Ah` consumer。新 bytes 證明該值若未在進入新 map-local event bank 時歸零，
Dexam 劇情會被直接略過。舊文件的 raw trace 本身仍有效，但「不可清除」只適用
該局部 continuation，不可擴大成跨所有 ECL block 永久共享的規則。

正常 observer 原本能續跑四段 Dexam 文字，卻未把其 stable message ID 放入
`seen` 白名單，造成事件已完成仍報未覆蓋的假陰性。本輪補齊 ID，不以顯示字串
硬編碼測試。

## 驗收與範圍

- `TestRealNewGameContinuesFromHapToBeholderCaveEntrance`：新遊戲→手札 59→LOOK→
  Dexam 雙戰→戰後洞穴。
- `TestRealBeholderCaveDexamAndZhentilBattles`：精確鎖定兩批敵人、戰利品與局部出口
  ECL continuation。
- `TestDiscoveredDungeonSearchEdgeIsPassableFromBothSides`：同一物理 edge 不要求兩份
  作品 JSON。
- game-pack 測試鎖定 `4C03h=0` action 必須屬於眼魔洞穴 handoff，避免 JSON 陣列
  編輯錯位。

未完成：出口唯一 route、原版 LOOK／SEARCH 動作 trace、Dexam 怪物完整特殊能力、
法術／投射物／聲音逐幀 fidelity、戰敗／重訪與 save/reload。
