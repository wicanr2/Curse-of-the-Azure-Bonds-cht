# 397：Burial Glen 紅羽戰士誘餌與羅剎妖陷阱

狀態：`READY`

## 原始證據

| 成員 | SHA-256 |
|---|---|
| `ECL6.DAX` | `faca339db267cc88fe6f8dc6e42d7e52d362f790b6f2d49672f9758aa26175fb` |
| `GEO6.DAX` | `c2729f8b6d13ec6d497bf185841e5fb7d964dd797bd8c7c822f48053514b886c` |
| `MON6CHA.DAX` | `e739ed3dd2ccbfc6fa87d4c6d230723dafcd44ccba6f1f1f393f9a2b9f05c78b` |

手札 33 另由使用者提供的
`Curse-of-the-Azure-Bonds_Misc_DOS_EN_Adventurers-Journal.pdf`
第 10 個 PDF 頁面、印刷頁 17–18 直接核對。該頁證明 Caemir 自稱祖先
長眠於迷斯卓諾，並以祖父打造的魔法弓請玩家清除墓穴蜘蛛。

## 地圖位置與正常路徑

GEO6 block `40h` 的 terrain `05h` 位於 `(13,6)`。友善王后事件結束在
`(1,3)`；由原始牆面資料求得的最短合法路徑為 19 步：

```text
(1,3) → (1,2) → (1,1) → (2,1) → (3,1) → (4,1)
→ (4,2) → (5,2) → (6,2) → (7,2) → (7,3)
→ (8,3) → (9,3) → (10,3) → (11,3) → (12,3)
→ (12,4) → (12,5) → (12,6) → (13,6)
```

這證明精靈王庭是支線房間，不是章節出口。正常玩家 regression 逐格呼叫
`CanMoveDungeonWrapped` 與地城 lifecycle；途中若遇共用隨機遭遇，會依
正常選單撤退後繼續原座標。

## ECL 分支

初見提示是 `HE MAKES A GESTURE OF FRIENDSHIP`，共用遭遇選項為
`COMBAT／WAIT／FLEE／ADVANCE`。選 `WAIT`：

1. 寫入完成旗標 `4CC2h=1`。
2. 顯示紅羽戰士的故事並解鎖手札 33。
3. 提供 `AGREE／REFUSE PAYMENT／DISAGREE`。

`AGREE` 與 `REFUSE PAYMENT` 都會帶玩家前往墓穴；後者多一句感謝玩家
不收報酬。當 `4CBAh` 高於中立線時，精靈幽魂會警告這是陷阱。玩家可在
`DO YOU CONTINUE?` 選 NO 離開，事件不進戰鬥。

`DISAGREE` 會讓對方獨自穿門而入；慘叫後玩家可拒絕調查。若繼續，幽魂仍會
警告，最後進入同一組墓穴敵人，但不承受羅剎妖預先射擊。

## 伏擊傷害與戰鬥

答應同行並繼續時，`PICTURE 43h` 顯示紅羽戰士變形成羅剎妖。原始 ECL
只發出一條：

```text
DAMAGE flags=2, dice=1d6+6, saveFlags=35h
```

既有 DOS 傷害 consumer 證明 `flags=2` 是兩次隨機目標攻擊，不是全隊傷害，
`saveFlags=35h` 是命中加值。正常玩家測試以作品中立的
`ResolvePendingECLDamageWithDefaultHitResolver` 取得兩筆 outcome，再續入：

- `MON6CHA 41h` PHASE SPIDER ×6，icon `41h`；
- `MON6CHA 49h` RAKSHASA ×1，icon `43h`。

戰鬥勝利後接回既有的精靈骸骨 `LOOT GRAVE／REBURY SKELETON／GO` 選單。

## 資料化與繁中

- 手札 33、事件敘事及三個新選項皆由 CoAB game-pack stable ID 提供。
- `State` 的 ECL menu prompt 先查 game-pack `text_rules`，未命中才退回
  共用舊提示翻譯；新作品提示不再寫死於 Go。
- 產品測試從同一份 JSON 取得繁中期望文字，避免 JSON 改譯後測試因複製
  舊譯文而失效。

## 明確邊界

- 原始 `COMBAT／FLEE` encounter action 的完整外部 routine 副作用尚未還原；
  本輪正常路徑使用已驗證的 `WAIT` 分支。
- 羅剎妖的完整 AD&D 抗性、法術、AI、長弓與魔法箭戰利品、箭矢飛行動畫及
  音效仍需由 MON6ITM、戰鬥影片與 DOS runtime 逐項驗證。
- terrain `07h` 的獨立蜘蛛墓穴、terrain `0Ch` 的手札 56 人物，以及
  Burial Glen 後續出口、最終神殿與結局仍未完成。
- 本輪是可連續遊玩的事件切片，不代表 Burial Glen 或整款遊戲已完整可通關。
