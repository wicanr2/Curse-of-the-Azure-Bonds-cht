# 396：Burial Glen 精靈王庭與王后獎勵

狀態：`READY`

## 原始證據

| 成員 | SHA-256 |
|---|---|
| `ECL6.DAX` | `faca339db267cc88fe6f8dc6e42d7e52d362f790b6f2d49672f9758aa26175fb` |
| `GEO6.DAX` | `c2729f8b6d13ec6d497bf185841e5fb7d964dd797bd8c7c822f48053514b886c` |
| `MON6CHA.DAX` | `e739ed3dd2ccbfc6fa87d4c6d230723dafcd44ccba6f1f1f393f9a2b9f05c78b` |

本輪以 DOS 原始 ECL6 block `40h` 的 real-image session、GEO wrapped
movement 與正常 State 玩家路徑交叉驗證。terrain `96h` 只提供王庭內部
走廊幾何，沒有獨立 ECL 事件。

## Terrain `08h`：門口幽魂

玩家由蜘蛛巢 `(10,1)` 向西前進，必經 `(5,2)`：

1. `PICTURE 72` 顯示被壓碎的螳螂人與門口幽魂。
2. pause 後詢問 `DO YOU WANT TO ENTER THE BUILDING?`。
3. `NO` 直接 EXIT，不消耗事件。
4. `YES` 顯示幽魂歡迎玩家覲見王后，並寫座標
   `C04B=4／C04C=2／C04D=3`，即傳送到 `(4,2,S)`。

## Terrain `89h`：魔法盔甲階梯

terrain `89h` 位於 `(3,1)`。初次進入寫 `4CC4=1`：

| 選項 | ECL 結果 |
|---|---|
| `GO UPSTAIRS` | 盔甲躬身；`4CBA +1`；傳送至 `(2,1,W)` |
| `TAKE ARMOR` | 盔甲碎成鏽片；`4CBA -2` |
| `ATTACK` | 與 TAKE ARMOR 共用鏽片分支；`4CBA -2` |
| `RETREAT` | 直接 EXIT；不寫 `4CC4`，可重訪 |

原作認可值以 `80h` 為中立線。從 `80h` 搶奪或攻擊後會成為 `7Eh`，
因此會改變後續王庭與王后分支。

## Terrain `8Ah`：王庭判決

terrain `8Ah` 位於 `(1,2)`，進入時先寫 `4CC5=1`：

- `4CBA >= 80h`：王庭向玩家致意，pause 後 EXIT。
- `4CBA < 80h`：建立十四名敵人：
  - `MON6CHA 42h` 六名，icon `41h`；
  - `MON6CHA 41h` 四名，icon `41h`；
  - `MON6CHA 40h` 四名，icon `40h`。

## Terrain `8Bh`：王后幽魂

terrain `8Bh` 位於 `(1,3)`，兩條路徑都先寫 `4CC6=1` 並使用
`PICTURE 72`。

### 認可值 `>= 80h`

王后給予：

```text
TREASURE 0,0,0,0,0,12,8,41h
```

即 12 gems、8 jewelry 與 `ITEM6.DAX block 41h` 的六筆物品。財寶服務
結束後，王后道別並消失。正常玩家測試必須載入真正 ITEM6 blocks；若測試
沒載入便接受通用零怪物 COMBAT fallback，屬無效驗收。

### 認可值 `< 80h`

王后出現時先執行 `4CBA -5`，再提出離開葬林的交易：

- `YES`：取得 `4 gems、2 jewelry、ITEM6 block 40h`。
- `NO`：沒有財寶，王后宣告讓玩家與亡者同眠。
- 兩者最後都顯示高塔崩塌，並傳送到 `(5,2,S)`。

`4CC6` 在提問前已設定，所以拒絕後也不能重領或重問。

## 正常玩家路徑與資料化

從第 395 輪 `(10,1)`：

```text
(9,1) → (9,2) → (9,3) → (8,3) → (7,3) → (7,2)
→ (6,2) → (5,2 / terrain 08h)
```

接受幽魂邀請後：

```text
(4,2) → (4,1) → (3,1 / terrain 89h)
→ GO UPSTAIRS 至 (2,1)
→ (2,2) → (1,2 / terrain 8Ah)
→ (1,3 / terrain 8Bh)
```

- 所有步驟均由 `CanMoveDungeonWrapped` 驗證。
- 所有繁中由 game-pack stable ID 提供；產品層測試從同一份 JSON 取得
  期望文字。
- raw ECL 測試涵蓋門口 YES／NO、盔甲四分支、`80h` 門檻、十四名敵人、
  兩種王后財寶、扣五點、倒塔傳送與三個完成旗標。
- 正常玩家路徑從 Standing Stone 連續抵達友善王后，實際解出 ITEM6
  `41h` 的六筆物品並驗證 12 gems／8 jewelry。

## 明確邊界

- 本輪完成的是 Burial Glen 西側王庭 vertical slice，不代表整個 Myth
  Drannor 章節或全遊戲可通關。
- terrain `8Ah` 十四名敵人的完整 AD&D 能力、AI、動畫與音效尚未完成。
- 王庭高塔崩塌是否改變 GEO 視覺素材、離開 Burial Glen 的完整出口與下一個
  主線區域仍需繼續稽核。
