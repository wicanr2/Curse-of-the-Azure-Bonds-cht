# 第五百五十五輪：眼魔洞穴東門、正常出口與世界返回

狀態：`READY`
日期：2026-08-12

## 結論

同一個從新遊戲開始的玩家 session，現在已由手札 59、Dexam 雙戰與戰後洞穴，
繼續以 `LOOK` 揭露神殿東門，沿十二步普通 GEO 移動／開門抵達 terrain `93h`，
完成 Olive、Dimswart、Gharri 四段離場敘事，最後回到散提爾堡世界邊界及正常
世界選單。全程沒有直接設定座標、PC、旗標或注入戰鬥。

這關閉的是「眼魔洞穴主線正常路徑」；它不代表洞內所有隨機事件、分支、怪物
特殊能力、重訪／讀檔、後續章節或整作結局已完成。

## 證據與推論等級

| 項目 | 證據 | 等級 |
|---|---|---|
| 神殿東、西兩扇門 | 《軟體世界》中文掃描 `084.jpg`／印刷頁 52 的手札 59，在 Dexam's Shrine 西、東牆各畫一扇門 | `layout-only` |
| 神殿東門原始位置 | GEO4 block `25h` 的 `(15,1,E)` 與 wrapped `(0,1,W)` 均為 `wall=09/detail=0`；它正對手札的東門 | `strong inference`；尚缺原版 SEARCH writer→movement consumer trace |
| 東門後至出口路網 | 原始 `curseoftheazurebonds.zip` SHA-256 `c98698a6271c17177dfdb27f34b0389b7d34f58ef206e92575393f4655f5b26d`，GEO4 block 37；允許原始 detail 2/3 普通門後，`(0,1)` 至 `(6,3)` 有十二步路徑 | `exact`（GEO bytes／remake movement contract） |
| 出口事件 | terrain `93h`／`(6,3)` 觸發 ECL4 block `22h:+1305h` 的離場 continuation | `exact` |
| 世界目的地 | 真正正常 session 返回 `CurrentCity=12` 的散提爾堡邊界；局部 Dexam fixture 因缺少完整前置 world state 仍落在暗影谷，不能拿來證明目的地 | `exact`（正常 remake session） |

十二步路徑由 `(0,1)` 開始，方向依序為：

`S, E, E, S, S, E, E, E, E, E, N, W`

途中三道 `wall=0A/detail=2` 與一道異材質雙側 detail 2 都走既有普通門操作；
`(7,4,N)` 是 `wall=03/detail=1` 拱門。沒有新增第二道推測性 SEARCH edge。

## 實作

- CoAB game-pack 新增 `zhentil-keep.beholder-cave.dexam-shrine-east`，座標、牆值、
  發現方式與 confidence 都留在 JSON；State／frontend 沒有 Dexam 座標特判。
- `scripts/research/geo_route_components.py` 泛化為 `--set`，並新增
  `--open-doors`，可區分普通 detail 2/3 門與 detail 0 候選牆。
- 正常長路徑測試以 `TurnDungeonWithGrid`、`LookDungeonLocation`、
  `MoveDungeon` 與既有開門操作重播路線；離場按鍵沿局部 ECL 回歸已證實的順序，
  不讓 generic observer 自動消費下一座城市選項。
- 正式 gate 發現第 551–552 輪新增 audit 程式後，
  `docs/audit/go-han-literals-baseline.json` 仍是空 baseline；本輪使用正式
  `coab-audit -write-baseline` 列入九筆 `runtime_ui_debt` 精確 AST 雜湊。
  這是既有 audit 技術債補帳，不表示允許新增產品層硬編碼中文。

## 勘誤

- spec 554 的「`(15,1)` 到 `(6,3)` 尚未閉合」由本規格 supersede。
- 手札 59 不是只能證明模糊相對拓撲：它清楚畫出 Dexam 神殿的兩扇門；但它仍
  不能單獨證明 GEO 座標或原版 SEARCH 實作，因此東門維持 `strong inference`。
- `TestRealBeholderCaveDexamAndZhentilBattles` 是從 ECL block 直接建立的局部 fixture，
  其暗影谷回返是缺少完整世界前置狀態的測試預設，不是正常劇情結論。正常新遊戲
  session 證明離洞後回到散提爾堡。

## 驗收

- `TestRealNewGameContinuesFromHapToBeholderCaveEntrance`：新遊戲至散提爾堡世界
  選單的同一正常 session。
- `TestRealBeholderCaveDexamAndZhentilBattles`：局部兩戰、戰利品與 ECL 離場回歸；
  不作世界目的地 oracle。
- `scripts/research/geo_route_components.py ... --set 4 --block 37 --start 0 1
  --target 6 3 --open-doors`：原始 GEO 十二步 route。

未完成：眼魔洞穴全支線／隨機事件、Dexam 與梅杜莎完整能力、戰敗、存檔重訪、
離洞後後續主線至最終戰與結局。
