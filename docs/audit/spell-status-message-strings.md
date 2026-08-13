# 常數字串函式攜帶的訊息

由 `scripts/small_function_reader.py` 的判讀結果整理（第 569 輪）。
這些函式的整個 body 就是「把一段 Pascal 字串常數複製進區域緩衝，再呼叫
訊息 routine」，因此字串內容是函式行為的一部分，不是推測。

⚠ **兩平台尚未證明逐筆對應**：位址不同、數量也不同，下表分平台列出。
要建立 en／ja 對照必須另外證明是同一個函式，不得依表格列序配對。

## DOS（23 筆）

| 模組 | 位址 | 字串 |
|---|---|---|
| overlay-22 | `1D2Ah` | is Cursed |
| overlay-22 | `1DD5h` | is affected |
| overlay-22 | `1E0Fh` | is protected |
| overlay-22 | `1E4Eh` | is cold-resistant |
| overlay-22 | `2232h` | is shielded |
| overlay-22 | `245Bh` | is fire resistant |
| overlay-22 | `2494h` | is silenced |
| overlay-22 | `2682h` | is invisible |
| overlay-22 | `26BBh` | Knock-Knock |
| overlay-22 | `2754h` | is weakened |
| overlay-22 | `2FA4h` | is blind |
| overlay-22 | `3093h` | is diseased |
| overlay-22 | `36A7h` | has been cursed! |
| overlay-22 | `36E0h` | is blinking |
| overlay-22 | `3CEDh` | is Slowed |
| overlay-22 | `4043h` | is paralyzed |
| overlay-22 | `40D5h` | is invisible |
| overlay-22 | `472Ch` | is highlighted |
| overlay-22 | `4759h` | is invisible |
| overlay-22 | `4EF7h` | is protected |
| overlay-22 | `563Bh` |  |
| overlay-22 | `5669h` |  |
| overlay-22 | `5697h` |  |

## PC98（22 筆）

| 模組 | 位址 | 字串 |
|---|---|---|
| overlay-22 | `1F8Bh` | は呪いを受けた。 |
| overlay-22 | `203Dh` | は魔法にかかった。 |
| overlay-22 | `20BFh` | は冷気に対する防護を得た。 |
| overlay-22 | `24C1h` | は魔法の楯に守られた。 |
| overlay-22 | `26FCh` | は火炎に対する防護を得た。 |
| overlay-22 | `2736h` | は沈黙した。 |
| overlay-22 | `2933h` | は透明になった。 |
| overlay-22 | `2969h` | こんこん |
| overlay-22 | `2A04h` | は弱くなった。 |
| overlay-22 | `3273h` | は視力を奪われた。 |
| overlay-22 | `3369h` | は病いに冒された。 |
| overlay-22 | `3983h` | は呪われた！ |
| overlay-22 | `39C1h` | は点滅している。 |
| overlay-22 | `3FA8h` | は減速された。 |
| overlay-22 | `4313h` | は麻痺した。 |
| overlay-22 | `43ACh` | は透明になった。 |
| overlay-22 | `4A2Dh` | に光がまとわりついた。 |
| overlay-22 | `4A5Eh` | は透明になった。 |
| overlay-22 | `529Eh` | は守られた。 |
| overlay-22 | `59D8h` |  |
| overlay-22 | `5A06h` |  |
| overlay-22 | `5A34h` |  |
