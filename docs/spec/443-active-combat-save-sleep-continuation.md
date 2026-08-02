# 第四百四十三輪：active combat 存檔與 Sleep continuation

狀態：`READY`

## 問題

remake JSON save v6 只保存冒險、ECL session 與 ECL PRNG。若在戰鬥中存檔，
`ModeCombat` 雖會寫入，`Battle` 卻完全遺失；讀檔得到 combat mode＋nil battle。
Sleep `35h` 的剩餘 duration、Action、dynamic scheduler、持續區域與戰鬥亂數
也都無法 continuation。這種檔案不能稱為可恢復存檔。

## save v7 contract

version 7 新增兩層 snapshot。

### 作品中立 Battle snapshot

- stable fighter／TeamList order 與完整 Fighter projection；包含 HP、位置、
  `MonsterAffects`、Action、法術／怪物欄位與圖示狀態；
- party／enemy battle-scoped attack-roll modifier；
- round、status、persistent areas、next area ID；
- dynamic initiative entries、目前 selection 與 selected flag；
- 尚未交付的 spell interruption queue；
- 戰鬥 `randomstream.Snapshot{seed,draws}`。

`draws` 是底層 `math/rand.Source` 次數。restore 受 engine
`MaxReplayDraws` 限制，不接受任意重播成本。這只證明 remake continuation
逐值一致；SSI PC-98／DOS 的原始 PRNG 演算法仍是 `unknown`。

### CoAB State snapshot

作品 adapter 另保存 turns／turn index、DELAY markers、target／spell／point／
move cursors、combat speed、Quick Magic、view、return mode、訊息與尚未開始播放
的 `VisualEvent`。`LineTerrain`、`TACTICALMAP provider` 與素材不序列化；它們是
frontend 從目前 game pack／map 重建的唯讀 callback。

## 失敗即關閉

restore 會拒絕：

- 未知 snapshot version、負 round、非法 status；
- 重複／越序／不存在的 scheduler fighter；
- selection 沒有對應 entry；
- persistent area ID／caster 不存在；
- interruption 指向不存在 fighter 或 spell ID 0；
- turn、delayed marker、view／visual fighter、class／speed／return mode 越界；
- `ModeCombat` 沒有 combat snapshot，或非 combat mode 帶 combat snapshot；
- 視覺已進入 travel／impact／death 後的存檔。

最後一項是刻意的 truthful gate：wall-clock elapsed 仍由 Ebiten frontend 擁有，
目前 save 不會假造 mid-animation 的 phase。尚未開始的 Sleep TWINKLE transaction
可以保存；已開始播放時會明確拒絕。

## 可重現整合路徑驗證

固定 seed `443`：

1. 正常手動選格施放 Sleep，成功寫 effect `35h` 與 15 ticks，並建立
   `VisualTwinkle`；
2. 在第一幀前 SavePartyFile，載入新 State；
3. 比對 visual、Battle、scheduler、Action 與 PRNG snapshot；
4. 原狀態與 loaded 狀態各完成同一 handoff，snapshot 完全相同；
5. 第一分支繼續 14 次 round boundary，兩者同時自然解除；
6. 第二個 loaded 分支完成 handoff 後由物理命中造成正傷害，Sleep 解除且已
   消耗的 memorized slot 不會重複改變；
7. 視覺開始後再次存檔，確認 fail-closed。

這是公開 `StartCombat→選格→Cast→Save／Load` transaction；沒有直接注入
effect、PC、強制勝利或 teleport，但戰鬥 fixture 仍由測試建立，不等於從
Standing Stone／ECL encounter 抵達的完整 campaign 玩家路徑。

## 證據等級

- Sleep writer／duration／damage removal／CLEARACTION：`exact`，沿用 READY
  spec 434、440、441、442 的 PC-98 bytes／IDA／consumer 證據。
- scheduler entries／selection 語意：`exact`（PC-98），沿用 READY spec 419–420。
- save v7 JSON round-trip 與 remake PRNG continuation：`proven`（本輪 codec、
  deterministic regression）。
- 原版 SSI 存檔是否保存 mid-combat／PRNG，以及其欄位布局：`unknown`。
- mid-animation seamless restore：`not implemented`，目前明確拒絕。

## 尚未完成

- 保存 Ebiten wall-clock visual elapsed、當前音訊 sample offset，做到任意 frame
  無縫續播。
- 原版 SAVGAM active-combat record／RNG consumer 的反組譯與雙向匯入。
- 固定 PC-98 戰場的長局 save corpus、Quick Sleep 與第二款 Gold Box 驗證。

更新：真實 ECL encounter 的 save/load campaign regression 已由 READY spec 444
接續完成。
