# 第一千二百三十九輪：QUICK 火球友軍安全與戰敗後窄化休息

狀態：`READY`（只關閉 Fireball 友軍安全判定與正常路徑測試的戰敗恢復；
不宣稱一般強度連續通關）

## 結論

1. PC-98 overlay-09 `02B1h` 的範圍法術安全檢查會掃描施法者同陣營角色；
   友軍只要有一人未通過該次豁免，就不得把該格當作 QUICK Fireball 中心。
   這項控制流證據已由 spec 777／1112 保存，本輪把它接到 remake 的正式
   `quickAreaTargetLegal`，不再讓阿卡巴朝緊鄰隊友的黑暗精靈領主施放火球。
2. Fireball 採原版陣營修正：玩家方 `-2`，另一方 `+8`，豁免類別為 `4`；
   候選範圍內沒有敵人時同樣不合法。這只改 CoAB adapter，不把作品資料寫進
   共用 engine。
3. 正常玩家路徑的戰敗恢復只在兩種玩家可見證據下啟動：目前敘事明示
   「可以在這裡安全休息」，或剛顯示「戰鬥失敗。」。恢復仍按 `E` 走正式
   CAMP／REST，不注入 HP、時鐘或座標。
4. 曾試作「看過一次安全休息後，整個 ECL block 永久可休息」；1800 幀測試
   在第 758 幀後反覆進出 CAMP／MEMORIZE。此策略已撤回，因為 block 不是原版
   的永久休息權限，也會把測試 driver 自己造成的 loop 偽裝成產品恢復。

## 驗證

- `TestQuickFireballRejectsCenterWhenFriendlyFailsOriginalSafetySave`：固定敵人與
  友軍位置及骰值，確認友軍豁免失敗時拒絕該 Fireball 中心。
- `TestCombatAltMQuickFireballUsesAreaCenterAndPendingDelay`：確認既有 QUICK
  Fireball 中心、延遲與法術格交易沒有回歸；兩項 focused test 在 Docker
  內通過。
- 不載路線 oracle 的 1800 幀正常鍵盤 session：171 格／130 句，最後新進展在
  第 1785 幀，證明窄化後不再停在休息循環；這不是主線覆蓋證據。
- 當輪載入 `route.json` 與被窄測試覆寫的 `route-clean-716.json` 都只到
  `ECL 0x03`；後續 spec 1240 已建立唯一現行入口並完整重生，不再沿用這兩份
  歷史檔。

## 下一步

此項已由 spec 1240 完成；下一步改查原版五／六級建角的法術能力。
