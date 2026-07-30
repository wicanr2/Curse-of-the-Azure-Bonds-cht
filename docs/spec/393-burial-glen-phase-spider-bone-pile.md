# 393：Burial Glen 相位蜘蛛骨堆

狀態：`READY`

## 原始證據

| 成員 | SHA-256 |
|---|---|
| `ECL6.DAX` | `faca339db267cc88fe6f8dc6e42d7e52d362f790b6f2d49672f9758aa26175fb` |
| `GEO6.DAX` | `c2729f8b6d13ec6d497bf185841e5fb7d964dd797bd8c7c822f48053514b886c` |
| `MON6CHA.DAX` | `e739ed3dd2ccbfc6fa87d4c6d230723dafcd44ccba6f1f1f393f9a2b9f05c78b` |

ECL6 block `0x40` 的 SearchLocation selector `0x15` 跳至 payload
`+1A03h`：

1. `4CCFh == 1` 時直接 EXIT。
2. `SETUP MONSTER 41h,2,41h` 使用 `MON6CHA 41h` PHASE SPIDER。
3. 顯示：
   `SPIDERS HAVE GATHERED A PILE OF BONES HERE AND DEFEND IT.`
4. pause 後寫 `7F82h=10`、`4C01h=6`，透過共用 combat routine
   建立六隻相位蜘蛛。
5. 勝利 continuation 先寫 `4CCFh=1`，再詢問：
   `WHAT DO YOU DO WITH THE BONES?`
6. 原始水平選單為 `LOOT／REPLACE IN CRYPTS／IGNORE`。

三個 branch target 與副作用：

| 選項 | payload | 原始操作 | 結果 |
|---|---:|---|---|
| `LOOT` | `+1AA9h` | `SUBTRACT 1,4CBAh,4CBAh`；`CLEARMONSTERS`；`TREASURE 0,0,0,0,0,1,0,FFh`；`COMBAT` | 黛米爾好感減一，加入一顆 gem，`FFh` 明確表示沒有 item block |
| `REPLACE IN CRYPTS` | `+1AC6h` | `ADD 1,4CBAh,4CBAh` | 黛米爾好感加一後退出 |
| `IGNORE` | `+1ACFh` | 清畫面、`PICTURE FFh`、common call | 好感與財寶均不變後退出 |

`TREASURE` 的第六個 coin operand 對應 gems；第七個才是 jewelry。
`ItemBlock=FFh` 在現有 evidence-backed resolver 中是「無物品」，不能誤寫成
隨機裝備。

## GEO 正常路徑

terrain `95h` 位於 GEO6 block `0x40` 的 `(14,10)`。從第 392 輪
terrain `94h` `(14,8)` 向南回到 `(14,9)`，再向南一步即可抵達。兩步均由
`CanMoveDungeonWrapped` 驗證；途中若出現隨機遭遇，正常 FLEE continuation
會回到同一 ECL session。

## 資料化與驗證

- 守衛文字、骨堆問題與三個選項均使用 CoAB game-pack stable ID；沒有新增
  本作專用 Go 劇情分支。
- raw ECL session regression 分別驗證：
  - 六隻 `41h` PHASE SPIDER；
  - `7F82=10`、`4C01=6`、`4CCF=1`；
  - 三個選項；
  - `LOOT` 的 gem／`FFh`、好感減一；
  - `REPLACE` 好感加一；
  - `IGNORE` 好感不變；
  - 完成後重訪直接 EXIT。
- 正常 State 玩家路徑從 Standing Stone 一路延伸至 `(14,10)`，完成戰鬥、
  繁中選單、LOOT、gem service、地城 continuation 與重踏不重播。

## 明確邊界

- PHASE SPIDER 的位移相位、毒素、AI 與完整 AD&D 特殊能力仍未完成。
- 尚未取得這一場戰鬥的 DOS 動態影片時間碼，不能宣稱攻擊動畫、毒素演出、
  音效或回合節奏忠實。
- 本輪只證明 Burial Glen 這一組 `93h／94h／95h` 相位蜘蛛防線；不能據此
  宣稱 Myth Drannor 或整款遊戲已完整可通關。
