# Remake OGG sound and music assets

正式 remake 播放一律使用 OGG。九個音效由 reference project
`engine/seg044.cs`／`Main/Resource.resx` 證實的 22050 Hz mono PCM WAV 轉為
Ogg Vorbis；`internal/sound` 以原版 `Sound` selector 對應它們。

| Reference selector | OGG | 用途（依 resource／呼叫點） |
| --- | --- | --- |
| `sound_2` | `missle.ogg` | missile |
| `sound_3` | `magic_hit.ogg` | magic hit |
| `sound_5` | `death.ogg` | death |
| `sound_6` | `sound_5.ogg` | reference generic sample |
| `sound_7` | `hit.ogg` | hit |
| `sound_9` | `miss.ogg` | miss |
| `sound_A` | `step.ogg` | step |
| `sound_B` | `sound_10.ogg` | reference generic sample |
| `sound_D` | `start_sound.ogg` | start |

`sound_0`／`sound_FF` 是停止目前播放；其餘沒有來源素材的 selector 不虛構資產。

十二首 `pc98-bgm-selector-01.ogg` … `pc98-bgm-selector-0c.ogg` 由使用者持有的
PC-98 Disk 1 抽出 `MSCDRV.EXE`，以 `cmd/pc98-render-track` 在 44.1 kHz 渲染
120 秒，再以 FFmpeg `libvorbis -q:a 5` 轉換。正式遊戲以 OGG 為主；
`MSCDRV.EXE` 即時合成只保留為研究 oracle／缺檔 fallback。

原始 driver、磁碟映像、PCM 中間檔與由商業原作衍生的音樂 OGG 不提交公開 Git；
發行前必須另行確認散布權利。音效 OGG 延續既有 reference asset 的散布邊界。
