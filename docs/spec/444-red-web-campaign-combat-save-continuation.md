# 第四百四十四輪：紅網 campaign 戰鬥存檔 continuation

狀態：`READY`

## 目的

第 443 輪以公開 combat API 證明 save v7 的 Battle／Sleep round-trip，但 encounter
fixture 仍由測試建立。這輪關閉該驗收缺口：從 Standing Stone 正常世界旅行，
沿 GEO6 合法地圖路徑抵達 Myth Drannor 紅網，在真實 ECL6 encounter 第一戰
中存檔；載入全新 State 後完成同一 ECL continuation 的兩場戰鬥。

## 正常玩家路徑

原始 DOS image SHA-256：
`c98698a6271c17177dfdb27f34b0389b7d34f58ef206e92575393f4655f5b26d`。
測試只讀取此 image；version 7 save 寫入 `t.TempDir()`，不覆寫原始資料。

路徑沿用第 383–387 輪已由原始 ECL／GEO 證明的流程：

1. `arriveAtWorldLocation(4)` 抵達 Standing Stone，完成提朗瑟克斯揭露；
2. `JOURNEY ON → MYTH DRANNOR → WILDERNESS → Enter`，進 ECL6 block `40h`；
3. 由 exact spawn `(2,15,E)` 沿 GEO6 可通行格前進；
4. terrain `01h` 精靈幽魂、Picture 72、GREET 與 Journal 25；
5. 沿合法東向路徑到 terrain `82h` 紅網；
6. `SPEAK` 輸入 `Krrkik` 並返回，再選 `ENTER IT`；
7. ECL 建立四隻 `MON6 42h` 蜘蛛，State 進 active combat party-turn boundary。

沒有直接設定 ECL PC、注入 combat request、teleport、grant item 或 forced win。
測試英雄數值很高以縮短長路徑戰鬥，但 encounter、地圖、分支、續跑位址、
怪物資料與 completion flag 均來自原始資料路徑；因此規則強度不是原版難度驗收。

## Save／Load transaction

在第一場蜘蛛戰尚未執行 ALT+M／ALT+Q 前：

1. 取得 `BattleSnapshot` 與 `SessionSnapshot`；
2. `SavePartyFile` 寫入 version 7 JSON 至測試 writable temp；
3. 由同一玩家自備 ECL blocks 建立全新 `State`；
4. 重新掛入真實 `MON6CHA／MON6SPC／ITEM6` decoded catalogs；
5. `LoadPartyFile`；
6. loaded Battle／ECL snapshot 與 save 前逐欄 `DeepEqual`；
7. 後續只使用 loaded state。

loaded state 接著：

- 啟用 Quick Magic／全隊 Quick，擊敗四蜘蛛；
- 同一 ECL session 戰後續跑並顯示 Picture 72 羅剎妖；
- 建立並擊敗 `MON6 43h` 羅剎妖第二戰；
- 顯示繁中自由訊息、寫 `4CBF=1`；
- Continue 回到 dungeon，重踏不重播紅網事件；
- 原長路徑後續仍繼續 Burial Glen／Myth Drannor 事件，整個既有 regression 通過。

## 證據等級

- ECL／GEO 路徑、怪物 IDs、兩戰 continuation、`4CBF`：`exact`，沿用 READY
  spec 383–387 的原始 bytes／runtime evidence。
- save v7 loaded Battle／ECL session snapshot equality：`proven`，本輪真實資料
  deterministic regression。
- loaded state 完成兩戰及旗標：`proven`，本輪 normal player-path regression。
- 高數值測試英雄的戰鬥難度：只屬測試加速，不支持原版 AD&D encounter
  balance 已完成。
- SSI 原版 SAVGAM active-combat 格式：`unknown`；本輪仍是 remake JSON save。

## 尚未完成

- 一般能力值隊伍、手動逐回合操作與失敗／撤退分支的同一路徑 save corpus。
- mid-animation／音訊 sample offset 無縫存讀檔。
- 原版 SAVGAM combat record 與 RNG consumer 反組譯。
- 其他 encounter 類型：allied NPC、persistent cloud、line spell、treasure boundary
  及 final battle 的 active save sampling。
