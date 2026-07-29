# 第三百五十三輪：眼魔洞窟離場

狀態：`READY`

## 原版證據

- ECL4 block `0x22` entry 0 先檢查 `(C04F & 0x3F)==0x13` 與
  `7ED5==1`，因此離場是 boundary lifecycle，不是一般 terrain event。
- GEO4 block `0x25` 的 runtime 出口格為 `(6,3)`、terrain `0x93`。
  攻略標示的 `(11,15)` 使用另一套地圖座標，不可直接寫入 runtime。
- `+0x1305` 起依序顯示 Olive／Dimswart 告別、遠方騎士呼喊與紫衣女子。
- 結尾寫入 `4CE2=1`、`7F12=1`，執行 `DESTROY ITEMS 0x61/0x60`，
  再 `NEWECL 0x51` 回到 Shadowdale 世界選單。

## 實作邊界

- 玩家路徑必須由 `RunDungeonExitLifecycle` 觸發；普通
  `RunDungeonLifecycle` 在同格只會走 SearchLocation 分支。
- 劇情文字由 game-pack JSON 翻譯，通用引擎不得認識 Olive、Dimswart、
  Shadowdale 或神器。
- regression 從 Dexam 與 Zhentil Keep 兩場勝利後開始，鎖定 terrain、
  PICTURE 42、四段繁中、兩個離場旗標、block `0x51` 與世界選單。
- `DESTROY ITEMS 0x60/0x61` 的實際裝備移除仍須另以帶 DOS inventory 的
  player-path fixture 驗證；本輪只鎖定原 ECL transaction 可走到世界畫面。

## 驗收

- `TestRealBeholderCaveDexamAndZhentilBattles`
- game-pack JSON 可解析。
- `go test ./internal/game` 通過。

