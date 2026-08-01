# 409：ECL session 存檔與持續亂數 continuation

狀態：`READY`

## 範圍

本輪把 remake JSON save 升級為 version 6，保存可變 ECL session：current
block、resume PC、call stack、共享 work memory、字串記憶體、比較旗標、輸入
offset、尚未交付的怪物 setup，以及持續 PRNG 的下一個位置。

這關閉「存檔後隨機遭遇從 seed 起點重播」的 fidelity 缺口，但不代表原版
DOS／PC-98 RNG 演算法已逐位元還原，也不代表任意 UI pause 的畫面文字、選項
與動畫時間軸已完整保存。

## 作品中立 engine contract

獨立 `golden-box-remake-engine/randomstream` package 保持現行 Go
`math/rand` 輸出，snapshot 只含：

```json
{"seed": 1032, "draws": 17}
```

`draws` 是底層 `Source.Int63／Source64.Uint64` 呼叫數，不是高階 `Intn`
次數；rejection sampling 因此仍可 exact continuation。Restore 以同 seed
重播底層 state step，並以 `MaxReplayDraws` 拒絕惡意巨大 save。

證據等級：

- `proven`：相同 seed 的既有 `math/rand` prefix 不變。
- `proven`：混用 `Intn／Int63n／Uint32` 後 snapshot，還原結果逐值相同。
- `proven`：超過 replay 上限的 snapshot 失敗即關閉。
- `unknown`：Go source 與 SSI 原版 RNG 是否相同；不得由本輪擴大宣稱。

## ECL snapshot 與商業資料邊界

`BlockSession.Snapshot／RestoreSnapshot` 保存 runtime-owned mutable state。
`0x8000..0x9DFF` code window 不會整包進 JSON：snapshot 逐位址與目前玩家
自備 ECL block 比對，只保存 runtime 改寫過的差異。載入時先從原始 blocks
重建 code window，再覆蓋差異。因此公開程式碼與一般 save schema 不內嵌
商業 ECL payload。

一旦 session 已有 stream，後續 runner 的 compatibility seed 不得覆蓋它；
明確重播／測試改 seed 只能呼叫 `ResetRandomSeed`。State round-trip 特別使用
非預設 seed 驗證，避免載入後被 frontend 預設值悄悄重設。

Restore 另驗證：

- snapshot version 與 current block 必須存在；
- PC／stack 必須落在該 block payload 邊界；
- input offsets 不得為負；
- regular memory 與 code-memory differences 不得混用位址空間；
- PRNG replay 必須通過有界驗證。

## CoAB save version 6

`SavePartyFile` 在有 ECL session 時取得 snapshot，與 roster、AREA、地圖
位置及 game clock 一起寫入。`LoadPartyFile` 必須已有玩家自備原始 ECL blocks，
才可恢復 session；缺少 blocks 時明確報錯，不會忽略劇情 continuation。

舊版 v1–v5 仍可載入。它們沒有 ECL snapshot，因此只能沿用舊相容行為，
不能冒稱保存了亂數下一值或 ECL resume PC。

## 驗證

- engine `TestStreamMatchesMathRandAndRestoresContinuation`。
- engine `TestRestoreRejectsExcessiveReplay`。
- ECL 合成 PROGRAM boundary：memory、PC 與下一次 RANDOM round-trip。
- save v6 JSON encode／decode 與 State `SavePartyFile／LoadPartyFile` round-trip。
- 真實 `ECL6.DAX` block `40h` terrain `04h`：先消耗原作 RANDOM，snapshot
  後比較不中斷與 restore 分支；raw random values、怪物生成與文字完全相同。

## 尚未完成

- 保存當下 UI message、choices、pending treasure／combat、動畫 elapsed 與
  音訊播放位置，使任何 frame 都能無縫讀檔。
- 原版 SAVGAM 的 RNG／interpreter continuation 欄位 consumer；version 6
  是 remake JSON 格式，不應回寫未知 DOS 欄位。
- 長時間遊戲的 draw-count 上限 corpus 稽核，以及第二款 Gold Box 遊戲採用。
