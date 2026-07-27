# Reference PC sound assets

這些 WAV 由 reference project `engine/seg044.cs`／`Main/Resource.resx` 證實並抽出，保留原始檔名與 22050 Hz mono PCM 格式。`internal/sound` 以原版 `Sound` selector 對應它們，Ebiten adapter 會轉成目前 audio context 的 sample rate。

| Reference selector | WAV | 用途（依 resource／呼叫點） |
| --- | --- | --- |
| `sound_2` | `missle.wav` | missile |
| `sound_3` | `magic_hit.wav` | magic hit |
| `sound_5` | `death.wav` | death |
| `sound_6` | `sound_5.wav` | reference generic sample |
| `sound_7` | `hit.wav` | hit |
| `sound_9` | `miss.wav` | miss |
| `sound_A` | `step.wav` | step |
| `sound_B` | `sound_10.wav` | reference generic sample |
| `sound_D` | `start_sound.wav` | start |

`sound_0`／`sound_FF` 是停止目前播放，`sound_1`、`sound_4`、`sound_8`、`sound_C`、`sound_E`、`sound_F` 在 reference resource table 沒有可用 WAV；因此 catalog 不為它們虛構素材。
