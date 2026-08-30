# 建立隊伍與主題切換抽樣驗證（2026-08-30）

## 結論

- 原版核心垂直鏈「建立角色 → 獨立保存 → 從 Curse 清單加入」已在 remake
  正常按鍵路徑重現；建立與加入不是同一個動作。
- DOS 空隊伍功能選單是「建立、加入、讀檔、離開」；remake 現行外層是
  「建立、加入、開始冒險」。核心建隊流程對拍，但整張外層選單不是逐項
  `exact`，只能標為 `material-exact/layout-reconstructed`。
- DOS 的來源選單依序是 Curse／Pool／Hillsfar／Exit；remake 順序相同。
  Curse 角色庫可用，Pool／Hillsfar 尚未實作舊格式匯入並採失敗即關閉
  （fail-closed）。
- DOS 清單以 `*` 標記已入隊角色；remake 保留相同互動語意與防重複加入，
  但使用版本化 JSON 而非 `.guy`，版面也不是逐像素復刻。
- F2 在建立隊伍外層與第一張地圖均能切換 `original`／`modern-a6`，且不改變
  建隊步驟、地圖座標或朝向。
- 現代／原版主題會分派到各自的 sprite 與 tileset；完整資產護欄另確認
  48／48 個區域 tile、512／512 個動畫與隊伍 sprite 索引都有現代素材。
- 建角頁目前沒有角色 sprite／大頭像，也不顯示地圖 tile；所以本輪只證實該頁
  的框線與色彩會切換，以及 sprite／tileset 的實際分派契約。若要求建角頁肉眼
  同時看到角色圖像切換，仍需先依原版證據另寫 READY spec。

## 原版 oracle

正常 DOSBox 玩家路徑（非 direct-entry）：

1. 標題按 Enter，進入 `CHOOSE A FUNCTION`。
2. 選 `ADD CHARACTER TO PARTY`，進入 `ADD FROM WHERE?`。
3. 選 `CURSE`，進入角色清單；已入隊的 `BOB` 以前置 `*` 顯示。

本輪擷取：

- `workplace/dos-oracle/out/tester-party-menu-20260830.png`
- `workplace/dos-oracle/out/tester-add-source-20260830.png`
- `workplace/dos-oracle/out/tester-character-list-20260830.png`

上述檔案是工作區證據，不是 Git 發行素材。原始定位、檔案格式與控制流證據見
[`spec 1246`](../spec/1246-original-outer-party-assembly.md)。

## Remake 驗證

Docker 內執行：

```sh
COAB_KEY_FRAMES=400 tools/go.sh test ./cmd/azure-bonds-game \
  -run 'TestKeysDriveARealSessionFromTheTitle|TestThemeSwitchChangesSpriteAndTileAssetDispatch|TestF1ToF6AreGlobalAndResolutionCycles|TestEveryAreaTileResolvesToModernA6|TestEverySourceAnimatedAndPartySpriteResolvesToDoubleResolutionModernA6' \
  -count=1 -v
```

結果：全部通過。正常按鍵 session 從標題經建角、保存、加入與出發，400 幀走過
47 格、5 種模式、41 段玩家文字；英文回退 0、全滅 0。F2 的兩個狀態保持檢查與
sprite／tileset 正反主題分派測試均通過。

## Fidelity 分級

| 項目 | 結果 | 分級 |
|---|---|---|
| 建立後回外層、再從 Curse 加入 | 相符 | `exact` 行為／重建版面 |
| 來源選單順序 | 相符；兩種舊格式尚未匯入 | `exact` 選項、`partial` 功能 |
| 已入隊標記與防重複 | 相符 | `exact` 語意 |
| 外層完整命令集合 | 讀檔／DOS 離開未在同頁；remake 有開始冒險 | `nearby` |
| 建角途中 F2 | 主題切換且狀態不漂移 | remake 契約已證實 |
| sprite／tileset 切換 | 正反分派與全索引護欄通過 | remake 契約已證實 |
| 建角頁可見角色 sprite | 現行頁面沒有該視覺元素 | 待決定／待 spec |

