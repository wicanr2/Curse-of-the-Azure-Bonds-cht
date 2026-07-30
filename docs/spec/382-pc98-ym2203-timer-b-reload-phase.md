# 382：YM2203 Timer B 重載相位契約

狀態：`READY`（限 free-running divide-by-16 的公式、作品中立 API 與
CoAB 現行零相位模型邊界；原機 CPU／OPN 共時相位仍未完成）

## 1. 問題

第 375 輪只保存完整 Timer B count period：

```text
16 × (256 − register26) × 12 operators × prescale
```

但 ymfm 指出 `×16` 分頻器持續運轉；由 register `27h` 把停止的 Timer B
重新 load 時，第一個週期會依當下相位縮短。先前 engine 只有註解，沒有
可供平台 bridge 明確傳入相位的契約。

## 2. 可重現證據

版本化 BSD-3-Clause ymfm `ymfm_fm.ipp` 的 `engine_mode_write()` 呼叫：

```text
update_timer(1, load_timer_b, -(m_total_clocks & 15))
```

`update_timer()` 先將 Timer B period 設為
`16 × (256 − B) + delta`，再乘 `OPERATORS × clock_prescale`。因此首次
重載的晶片時鐘是：

```text
(16 × (256 − B) − phase) × 12 × prescale
phase ∈ 0..15
```

不能把 phase 誤解成最後只減 1–15 個晶片時鐘。

既有 IDA Pro 9.4 輸出與原始 bytes 又證明 CoAB ISR：

1. `10F85h` 呼叫 helper 寫 `27h=20h`；
2. 執行 `10175h` 的七聲道音序列 interpreter；
3. `10F9Ch` 後由 `11012h` 寫 `27h=0Ah`。

兩次 I/O 間隔會隨聲道事件、branch、busy flag 與硬體 wait 改變，不是固定
常數。S98 只有 10 ms 時間解析度，也不足以反推出 0–15 相位。

## 3. Engine 契約

獨立 engine `audio/ym2203` 新增：

- `TimerBReloadClockCycles(value, prescale, dividerPhase)`；
- phase 僅接受 `0..15`；
- phase 0 與完整週期完全相同；
- 非零 phase 只縮短停止後重新 load 的第一週期；
- `TimerBSampleAccumulator.AdvanceReload()` 會把相位週期納入既有整數
  numerator／remainder，不引入浮點漂移。

API 不包含 CoAB 位址、ISR 或 CPU 型號；決定 `dividerPhase` 的 CPU／晶片
共同時間線仍由作品平台 adapter 負責。

## 4. CoAB 現況與邊界

現行 `TrackPCMStream` 在每個 `TrackPlayback.Tick()` 後立即套用 register
writes，沒有模擬 interpreter 消耗的 CPU 時間。每個完整 Timer B period
產生的 ymfm native sample 數又是 `16 × (256−B) × 3`，必為 16 的倍數；
所以這個抽象模型每次重載都落在 phase 0。

因此本輪不硬寫一個虛構的 CoAB 延遲，也不改變目前 WAV。要升級為原機
cycle-exact，仍需：

- 同時記錄 V30 指令／I/O 與 YM2203 clock 的 runtime trace；
- 校準 PC-9801 機型的 prefetch、memory／I/O wait；
- 逐次保留 register `27h` 重載相位，並涵蓋 save/resume。

本規格完成的是「相位公式與 API READY」，不是「CoAB 原機重載相位完成」。
