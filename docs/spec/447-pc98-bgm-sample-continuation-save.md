# 第 447 輪：PC-98 BGM 同取樣點存檔續跑

狀態：`READY`

## 問題

第 446 輪只保證戰鬥 visual 已送出的離散 cue 不重播。背景音樂仍由 frontend
持有，remake JSON save 沒有 track、driver machine、YM2203 phase、resampler
remainder 或 audio-device read-ahead，因此載入只能重新從曲首播放。

## 作品中立 engine

engine commit `f06493f` 暴露 vendored BSD `ymfm_saved_state`：

- `ymfm.Snapshot／Restore` 保存完整 YM2203 FM／SSG 內部狀態，restore 前要求
  opaque byte 長度完全相同；consumer 必須另帶格式版本。
- `pcm.LinearResamplerSnapshot` 保存 input／output rate、rational phase、previous
  sample 與 started flag；rate 或 phase 不一致時失敗即關閉。
- exact synth core 在 FM＋SSG 已推進 777 native samples 後 snapshot；全新 chip
  restore 的後續 2,048 samples 與不中斷分支逐 sample 相同。resampler 也以跨
  chunk continuation 逐 sample 相同驗證。

這些 API 不含 CoAB selector、曲名、driver bytes 或劇情 cue。

## CoAB save v8

`TrackPCMStreamSnapshot` 保存：

- 七聲道 `SequenceMachine` PC、call／loop stack、duration、volume、mode、tempo、
  counter 與 register shadow；
- Sound BIOS intent→register renderer 的 parameter state；
- opaque YM2203 state、resampler state、Timer B value；
- 尚未消耗的 800ms `MSCPLAY` silence 與 pending PCM；
- stable game-pack track ID、driver selector、output sample rate及 snapshot version。

Ebiten `Position()` 會扣除 Oto buffered bytes，代表 audible frame，不等於 decoder
read position。`TrackPCMStream` 因此保存四秒 bounded emitted history；snapshot
把 audible frame 到 decoder current 之間尚未聽見的 PCM prepend 回 pending。
播放器固定 250ms buffer，超出 retained history、超大 pending、未知 version／
selector、越界 sample rate或 sequence stack 一律失敗即關閉。

F5 在 JSON save 前取得播放器 snapshot；F9 與 `-party-load` 會用玩家本機 exact
driver restore。stable track ID 必須與 game-pack selector一致，不能讓存檔自行
改選曲目。舊 v1–v7 可載入，但沒有從未保存的 BGM continuation。

## 驗證與證據等級

- `exact`：上游 ymfm full-state API、YM2203／resampler 逐 sample continuation、
  Go 官方 dirhash 對舊 engine commit 重算與既有 `go.sum` 完全相同。
- `strong inference`：合成七聲道 fixture 在 decoder 已預讀 2,048 bytes 時，
  從 audible byte 2,048 snapshot；restore 第一個 byte 起與「buffer backlog＋
  不中斷後續」完全相同。sequence event continuation與損壞 PC／buffer bounds
  另有 regression。
- `unknown／未驗證`：本機目前沒有 SHA-256
  `bddbe63a90078bd9c8c8da5c45417c7ec3afcdf7fd5b724877a83ad9bb7b12f5` 的完整
  `MSCDRV.EXE`；玩家 VFD 的 driver sector 有缺口。因此本輪不能宣稱十二首真實
  曲目已做 runtime save/load 動態 oracle。
- engine `go test ./...`：`ENGINE_ROUND447_FORMAL_EXIT=0`；CoAB Docker／Xvfb／
  `--network none` `go test ./cmd/... ./gamepack ./internal/...`：
  `ROUND447_FORMAL_EXIT=0`。

## 邊界

- active PC-speaker／WAV one-shot 的 sample position仍未保存。
- 原版 SAVGAM 不含 remake v8 extension；本輪只涵蓋 remake JSON save。
- ymfm opaque state 與目前 vendored core 版本綁定；日後升級 core 必須提供明確
  migration 或拒絕舊 snapshot，不能猜測轉換。
- Timer B reload 的 CPU／OPN 共時 phase 仍缺原機 trace；保存目前重建模型不會
  自動升級為 cycle-perfect PC-9801 timing。
