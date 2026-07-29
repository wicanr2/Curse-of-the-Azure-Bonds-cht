# Gold Box audio knowledge

跨作品可重用的 audio layer 應分成三層：reference sound selector、raw WAV asset catalog、平台 playback adapter。遊戲 state 只發出 deterministic sound intent，不能在 rules／ECL VM 內直接建立 audio device。

CoAB reference `seg044.PlaySound` 使用 `Sound` selector：`2` missile、`3` magic hit、`5` death、`6` generic sample、`7` hit、`9` miss、`A` step、`B` generic sample、`D` start。已從 reference `Main/sounds` 保存 9 個 WAV，`internal/sound` 提供 selector mapping 與 Ebiten player；未被 resource table 證實的 selector 維持 no-op。

目前 adapter 已接入 title start、荒野移動、dungeon preview step，以及 State 發出的戰鬥命中／未命中／死亡／已實作法術 intent。後續 ECL routine 可沿用同一個 selector boundary；尚未因名稱或相鄰音效自行猜測完整音樂、MIDI、AdLib 或作品間共用的音效語意。

## PC-98 音樂層

PC-98 版 CoAB 不是把 WAV 放在磁碟裡播放。`MSCDRV.EXE` 是 DOS
terminate-and-stay-resident driver，接管 `INT D2h`，透過 port `0x188/0x18A`
控制 YM2203，並由 timer interrupt 推進序列。這代表跨 Gold Box 作品可沿用的
邊界應是「music intent／track metadata／chip-stream adapter」，不是把
CoAB 場景判斷寫進播放器。

使用者提供的 VFD1.00 Disk 1 恰好缺少 `MSCDRV.EXE` offset
`0x4000..0x43FF` 的一個 1024-byte sector；任何以零填補後得到的播放結果都
只能算 hypothesis。媒體稽核、缺失 CHRN、目前可證明的 `INT D2h` dispatch
與 12 首公開曲目目錄見 `docs/spec/358-pc98-vfd-and-fm-audio-source.md`。

MAME 官方 `pc98.xml` 已提供本作兩片 1,265,664-byte FDI 的身分雜湊，可供
第二份合法 dump 交叉驗證。VFD 的 absent sector 不能先當成一般壞軌補零：
保留缺 sector 的副本可進 MEGDOS，改變 sector topology、FAT root 或
`LOADER.COM` 則在 banner 前停止，顯示它可能同時承載早期完整性／防拷語意。
音訊工具必須分開保存媒體物理狀態與檔案系統 bytes。

Emulator trace 又確認目標 CHRN 在 baseline 被讀四次；「首讀 not-found、
第二讀起補零」仍無法進入 loader EXEC。這代表匯入器未來若只接受 flat raw
sector bytes，會遺失可能影響啟動與音訊驅動載入的多重讀取／sector-status
資訊。

`LOADER.COM` 的原 bytes 已確認依序 EXEC `SETUP.EXE`、`MSCDRV.EXE`、
`GAME.EXE`。磁碟內跳過第二步會破壞開機驗證；後續應在 NP2kai 加入有界、
只記錄研究 trace 的 guest interrupt／overlay instrumentation，而不是改寫
原始或研究磁碟。

PC-98 `GAME.EXE` 另保留 Borland `0x52FB` legacy symbol table。由完整
symbol／name-pool 邊界驗證後，已把 `MSCPLAY` 定位到 `0893:0114`：
它接受 1-based track byte、抑制重播同曲，轉為 0-based buffer 後透過
IVT vector `7Eh` 的 far-call wrapper 送出。這說明 literal `CD D2` 掃描
為零並不代表沒有音樂呼叫；追蹤 DOS／PC-98 老引擎時，應把「遊戲 wrapper
vector」與「TSR driver service interrupt」分成兩層 contract。

本作 area-code switch 已證明 selector 值 `3/4/5/6/8/9/12`，但 area code
尚未全部對回人類可讀場景。跨作品可沿用的是 symbol-table parser、IVT
wrapper 模式與 track intent；CoAB area code 表仍只屬 game pack，不進
共用 engine hardcode。

正式 remake 保留兩種來源：

- DOS 忠實 theme：既有 selector 與後續 PC Speaker／Tandy 行為。
- PC-98 配樂 theme：使用者原始媒體經雜湊驗證後 runtime import，或另有
  明確授權的重製音源；原始商業音軌不進 repo。

scene role → stable track ID、loop、來源與 confidence 屬 game-pack JSON；
play／stop／fade／resume、獨立 music／effect volume 與 save/load resume
屬共用 engine。
