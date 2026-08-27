# OGG 音訊資產驗證（2026-08-26）

## 決策與來源

- 使用者決定 remake 正式音樂、語音與音效以 Ogg Vorbis（`.ogg`）為主。
- 九個音效由原本 22050 Hz mono PCM WAV 轉換。
- 十二首音樂由 PC-98 Disk 1 的 `MSCDRV.EXE` 渲染；抽出檔 SHA-256：
  `bddbe63a90078bd9c8c8da5c45417c7ec3afcdf7fd5b724877a83ad9bb7b12f5`。
  該 driver 的媒體缺口不穿過十二首音序列，證據邊界見 spec 358／366。
- `cmd/pc98-render-track` 以 44.1 kHz stereo s16 渲染每首 120 秒；FFmpeg
  `libvorbis -q:a 5` 編碼。沒有把 800ms game-transition silence 烘進檔案。

## 機器驗證

- 21／21 檔皆由 FFprobe 辨識為 Vorbis 並可由 Ebiten `audio/vorbis` 解碼。
- 12 首音樂：44,100 Hz、stereo、每首 120 秒；峰值介於 -21.8 至 -13.9 dBFS，
  全部非靜音且沒有 clipping。
- 9 個音效：22,050 Hz、mono、時長 0.054195 至 1.728889 秒；峰值介於
  -24.4 至 -1.9 dBFS，全部非靜音且沒有 clipping。
- `internal/sound`、`cmd/azure-bonds-game`、`internal/game` 測試通過。

## Runtime 契約

- 正式切曲先載入 `<track_id>.ogg`，以整檔循環播放。
- OGG 缺失或解碼失敗時，只有明確提供 `-pc98-music-driver` 才回退即時合成。
- remake save 保存 stable track ID；OGG 載入後由檔頭重新播放。精確 sample-frame
  續播仍只屬研究用 PC-98 stream snapshot，不冒稱 OGG 已保持同一 sample。

## 未關閉的閘門

- 尚未由人耳逐首聆聽三輪，也未證明任意 120 秒邊界是原曲小節邊界；目前整檔循環
  可能在接點產生節拍或波形不連續。正式發行前須逐首找出 musical loop point，
  重切 OGG 並做三輪疲勞聆聽。
- 音樂 OGG 是商業原作衍生資產，轉檔不改變權利狀態；目前由 `.gitignore` 保持本地，
  不進公開 repository。納入公開發行包前必須另行確認散布權。
