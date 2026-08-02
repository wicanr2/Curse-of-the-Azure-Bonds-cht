# 第四百七十一輪：提爾佛頓公會至火刀據點文字資料化

狀態：`READY`

## 範圍

本輪把 ECL2 盜賊公會與轉入火刀據點的十四個作品文字 boundary 移入 CoAB
game-pack：公會首領問候／簡報、側門爆破、火刀命令、淬毒匕首、混戰開始、
豎琴半身人、犬舍戰前／戰後、猴籠、訪客簿、下水道痕跡、下水道入口與火刀
據點入口。每條規則保存原始 ECL token、英文訊息與繁中訊息；`State` 不再保存
這些作品譯文或比對分支。

## 證據與行為

- 原始 `ECL2.DAX` block 2 驅動公會事件、四名友軍盜賊對十三名敵人的混合戰，
  戰後續跑至手札 4；其位址與 team write 證據仍以 READY spec 293 為準。
- 同一真實 image integration session 續跑豎琴半身人、犬舍與三個房間事件，
  再經 block 3 下水道轉入 block 4；`NEWECL`、GEO 與 LOAD PIECES contract 仍以
  READY spec 297 為準。
- 十四個畫面文字均以 `requireGamePackText` 取得正式 stable ID 期望；測試不再
  複製繁中片段。公會突襲六個連續 pause 逐段驗證，避免只靠盲目選擇抵達戰鬥。

長回歸從真實新遊戲 session 建立狀態並使用真實 DAX／GEO／MON 資料，但在城市
內為縮短測試會直接設定已知 GEO 格；因此它證明事件 continuation 與資料綁定，
不單獨證明玩家逐步行走的完整正常路徑。混合戰與犬舍戰均由 combat runtime
實際完成，沒有直接清除敵方 HP。

## 驗證

- `TestTilvertonGuildAndHideoutTransitionIsGamePackDriven`：十四條規則的 en／zh-TW
  stable ID、非空訊息與無誤掛手札。
- `TestRealNewGameBeginsAtGlobalBlockOne`：真實 image session 的十四個畫面 boundary、
  兩場戰鬥、手札 4、block 2→3→4 continuation 與 GEO／LOAD PIECES handoff。
- Go 漢字字串基線：`594 → 580`；`localization_debt 99 → 85`，frontend 135、
  runtime 360 不變。

本輪沒有改 renderer，故不新增畫面，也不能擴大宣稱 UI 或整段遊戲已完成。
