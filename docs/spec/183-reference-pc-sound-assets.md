# 第 183 輪：reference PC sound assets 與 Ebiten playback

狀態：`READY`（限已證實 WAV selector 與目前三個觸發點）

## 證據

`engine/seg044.cs` 的 `SoundInit`／`PlaySound` 搭配 `Main/Resource.resx` 證實 9 個 WAV resource 與 selector mapping。素材以原始檔名放在 `assets/audio/`，`internal/sound` 保存 ID contract；`sound_0`／`sound_FF` 停止播放，未有 resource 的 selector 不會被猜測成音效。

`cmd/azure-bonds-game` 的 Ebiten adapter 現在：

- 啟動標題按 Enter 播放 `sound_D/start_sound.wav`。
- 荒野合法移動與 dungeon preview 合法移動播放 `sound_A/step.wav`。
- `-sound-dir` 可指定素材路徑；素材或 audio backend 不可用時只停用聲音，不阻斷文字／地圖／戰鬥流程。

本輪沒有宣稱完成完整戰鬥音效觸發、背景音樂、MIDI／AdLib、音量／設定選單或所有 ECL routine 的 sound call。
