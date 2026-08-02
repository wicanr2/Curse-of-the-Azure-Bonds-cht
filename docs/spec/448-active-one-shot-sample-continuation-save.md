# 第 448 輪：活動短音效同取樣點存檔續跑

狀態：`READY`

## 問題

第 447 輪已保存 PC-98 BGM 的合成器與 audible/read-ahead 狀態，但 DOS WAV 與
PC-98 software-speaker 短音效仍由 frontend player 持有。戰鬥 visual cue marker
只能防止讀檔後重新發出同一事件，無法讓已經播放一半的聲音從同一 sample續跑。

## remake save v9 contract

`internal/audiostate.Snapshot` 是 renderer-neutral 的 bounded record：

- `version=1`；
- 全域 `enabled`；
- 最多 64 個仍在播放的 one-shot；
- backend 只能是 `dos_wav` 或 `pc98_speaker`；
- key 分別是 DOS 原始 selector 的十進位字串，或 PC-98 gameplay `Event` stable ID；
- position 以 44,100 Hz 的 audible sample frame 保存，不保存 wall-clock 字串。

同一 backend／key 不得重複，key 長度限定 1–64。`enabled=false` 代表現行
`SetSound(false)` 已終止所有短音效，因此不能夾帶 active record。JSON v1–v8
仍可載入，但沒有從未保存的 one-shot position；frontend 會停止載入前的舊音效，
不讓它洩漏到還原後的世界。

## player transaction

DOS WAV 與 PC-98 software-speaker 都使用 seekable Ebiten player，且每個 selector／
semantic event 最多一個實例；同一 key retrigger 會 rewind，異 key 可以重疊。
snapshot 只接受 `IsPlaying()` 在讀取 Position 前後都成立的 player，因此自然結束
於取樣競爭窗口的音效不會復活。輸出依 backend／key 排序，使 JSON deterministic。

Ebiten Position 是由 PCM byte position 向下換算的 `time.Duration`。保存時以
ceiling 還原 frame；載入 seek 時則以 floor 產生 duration，兩者組合可回到同一
frame。restore 先停止所有 pre-load one-shot，再解析 backend、asset 與 frame；
同一 player 可同時持有兩組 backend，但每筆 identity 都必須找到對應資產。
backend 未載入、asset 缺失、位置 overflow 或 seek 失敗都保持全停，不能部分播放。

F5 依序取得 BGM 與 one-shot snapshot，再寫 v9 JSON；F9 與 `-party-load` 先還原
one-shot 的 enabled／position，再安裝 BGM，讓相同 enabled 狀態同時控制兩條路徑。
若本次執行沒有 sound player，F5 會清除先前載入的 one-shot snapshot；不能把未實際
播放、也無法重新量測的舊 cursor 再寫成 active continuation。

## 驗證與證據等級

- `exact`：播放器內 sample frame round-trip；同時 DOS／PC-98 active records；
  自然結束不保存；停用不復活；混合 backend 自身 round-trip；backend asset
  缺失／seek failure 後全停；save v9
  bounds、舊版拒絕 v9 payload及 State defensive-copy round-trip。
- `strong inference`：Ebiten `Position／SetPosition` 的整數換算由官方實作語意與
  可替換 player regression閉合；尚未以實體 Oto device loopback 錄音逐 sample
  比較，所以不能宣稱 DAC 輸出波形無縫。
- `layout-only／不適用`：本輪沒有新增畫面或視覺完成聲明。
- `unknown／未驗證`：原版 DOS／PC-98 在原生存檔時如何處理正在播放的短音效；
  原版 SAVGAM 不含 remake v9 extension。
- Docker／Xvfb／`--network none` 正式 gate
  `go test ./cmd/... ./gamepack ./internal/...`：`ROUND448_FORMAL_EXIT=0`；完整日誌
  位於 `/tmp/coab-round448-formal.log`。

## 邊界

- one-shot PCM bytes 仍由玩家本機 DOS WAV 或 exact PC-98 `GAME.EXE` 產生，存檔
  不嵌入原作音訊素材。
- 本輪保存目前 adapter 的 audible cursor，不證明 DOS PC Speaker／Tandy／AdLib
  backend 已完成。
- 同一 semantic event 同時多聲部重疊不在現行 player contract；若日後 mixer
  改成真正 voice pool，snapshot identity 必須增加 stable instance serial。
