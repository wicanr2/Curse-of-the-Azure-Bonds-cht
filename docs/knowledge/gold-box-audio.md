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

`MSCDRV.EXE` 已證明會把 IVT `7Eh` 直接設成自身 `CS:0080`。public ABI
是 `AH=0, AL=0-based track` 播放，`AH=1` 停止；handler 再透過 D2h client
操作 PC-9801 Sound BIOS。NEC 官方 BIOS 手冊證明 `CEE0` 是固定介面表，
`CEE0:[0006]` 是 entry offset，N88-BASIC 預設使用 D2h。這些 anchors
全位於 driver file `0x0280..0x14A1`，
不與缺失的 `0x4000..0x43FF` 重疊。可重用方法是「IDA 定位＋原始 file
offset byte audit＋缺口隔離」，不能因缺檔便放棄所有位於完整區段的 ABI
結論，也不能反過來宣稱整份 driver 已恢復。

本作實際有 `INITIALIZE`、`CLEAR`、`READREG`、`WRITEREG`、
`SETTOUCH`、`NOTE`、`SETLENGTH`、`SETPARABLOCK`、`READPARA`、
`WRITEPARA`、`ALLSTOP`、`CONTPLAY`、`MODUON/OFF`、`SETINTCOND`、
`HOLDSTATE` 與 `SETVOLUME` client，但沒有觀察到官方 `PLAY` 或
`SETTEMPO` wrapper。另有直接讀寫 YM2203 `0x188/0x18A` 的 helper。
可重用的音訊擷取器因此必須同時觀察 BIOS 命令與硬體 I/O，不能假設所有
狀態變更都經過單一 interrupt。

本作 `CURRENTECL` switch 已證明 selector 值 `3/4/5/6/8/9/12`，selector
5／6 又已對回戶外導航／城鎮服務場景。跨作品可沿用的是 symbol-table
parser、IVT wrapper 模式與 track intent；CoAB ECL block 表仍只屬 game
pack，不進共用 engine hardcode。

CoAB driver 的 `DS:0330` 是公開 track pointer table。前十二筆各指向
64-byte descriptor，內含七組 sequence offset／length；84 個 stream 全在
file `0x1B61..0x3C58`，早於缺失的 `0x4000..0x4400`。因此「driver 有缺
sector」不能再泛化成「十二首音符資料都殘缺」。可重用匯入原則是逐 track、
逐 channel 計算 half-open range 與 SHA-256，只開放沒有跨媒體缺口的資料。

曲名與音訊 bytes 也要分層。Hoot 的本作設定提供 0-based code→曲名，
game pack 以跨 locale `title_id` 保存 metadata；商業 sequence 仍只從
使用者媒體 runtime import。引擎 schema 不保存作品曲名，也不提交抽出資料。

正式 remake 保留兩種來源：

- DOS 忠實 theme：既有 selector 與後續 PC Speaker／Tandy 行為。
- PC-98 配樂 theme：使用者原始媒體經雜湊驗證後 runtime import，或另有
  明確授權的重製音源；原始商業音軌不進 repo。

scene role → stable track ID、loop、來源與 confidence 屬 game-pack JSON；
play／stop／fade／resume、獨立 music／effect volume 與 save/load resume
屬共用 engine。

2026-07-30 第 367 輪由 IDA `sub_10410` 證明音序列不是單一共用 grammar：
FM 0–2、PSG 3–5 共用 `A0–A4` jump／call／return／counted loop，但
`90` 只屬 FM、`91/92` 只屬 PSG，`B0` 在 FM 只跳過兩參數、PSG 則直接
寫 OPN register。兩種 stack 均為 16 entries；overflow／underflow 是
no-op，不是 fatal error。

Timing channel 6 更不能套一般 range parser。它只特別消耗 `85/8A`，
其他高 opcode 逐 byte 略過，而且原驅動不檢查 descriptor end。真實曲目
因此會 read-through 到相鄰資料。可重用安全模式是先驗證原檔雜湊，再以完整
driver data、命令 budget 與 event budget 限制 trace；不可為了記憶體安全
假造會改變節奏的 descriptor end。84 組 stream 已各跑 256 個 timed
events；下一層仍須把 note／參數語意交叉成 Sound BIOS／YM2203 register
events，才可選定合成後端。

2026-07-30 第 368 輪完成正常配樂路徑的事件 runtime。可重用邊界不是
「note 直接變 PCM」，而是先保留三種事件：

- YM2203 `register_write`；
- Sound BIOS `set_volume`；
- Sound BIOS `set_parameter_block`。

FM 音高讀 verified driver 的 DS `0210h` 12-word F-number table；PSG
音高讀 DS `0228h` 71-word period table，不能只擷取最前面的 24 words。
71 的邊界由 `00h..60h` note consumer 算式直接決定。Timer scheduler
同時維護 duration、PSG envelope pointer、`91/92` modulation 與 register
shadow。十二首各 4,096 ticks 的 event count／SHA 已進 auditor，商業表與
sequence 仍只在使用者媒體通過 SHA 後 runtime import。

2026-07-30 第 370 輪已取得 Hoot S98 v3 外部 register oracle。跨作品可
沿用的工具放在 engine `audio/s98`：嚴格驗證 header／device／wait／end，
並抽取 YM2203 tone burst 與 key-on snapshot；作品端只保存 selector
對映、input hash 與稽核結論。Hoot 會保存上次游標，批次擷取必須每次
`Home` 後使用絕對 row；S98 流水號不是 XML track code。

CoAB 二十組 NEC WORD parameter 與全部十二首 trace 證明：

- logical operator 1,2,3,4 要按 YM register slot 1,3,2,4 重排；
- rate 與 sustain-level 是反向 parameter，寫晶片前分別以 31／15 相減；
- signed DETUNE 採 8-bit left shift，不能先截成 3-bit；
- total level 會被 volume／carrier 規則改寫，timbre signature 應排除它；
- descriptor 的超範圍 parameter index 仍會形成真實 register 副作用，
  但若第一個 stream `85h` 在 note 前覆蓋，就不能誤列為可聽音色需求。

這項方法適用其他使用 NEC Sound BIOS 的 PC-98 Gold Box：分開追蹤「所有
SETPARABLOCK 呼叫」與「key-on 當下有效音色」，並以 runtime trace 判定，
不能只看靜態 index 聯集便推論缺 bank。

2026-07-30 第 371 輪再用指定 IDA Pro 9.4 正確以 8086／16-bit 載入
`SOUND.ROM`，並與同一批 S98 交叉驗證：

- parameter `OUTPUT_LEVEL` 寫晶片時是 `127-parameter`；
- `SETVOLUME` 只依 algorithm 改 carrier，順序是 operator `4,2,3,1`；
- algorithm 0–3／4／5–6／7 分別有 1／2／3／4 個 carrier；
- 欄位 5 `OPERATOR_MASK` 左移四位後形成 YM2203 `28h` key-on 高 nibble；
- 十二首共 72 組 descriptor／first-stream output-level sequence 全部相符。

可重用的 algorithm/carrier 與 logical→physical operator 拓樸放在 engine
`audio/ym2203`；NEC 50-WORD block 與作品 driver offset 仍留在作品端。
遇到其他 PC-98 Gold Box，不應把 FM key-on 固定寫成 `F0h`，也不能把
total level 當作 volume-independent timbre signature。

目前仍不能宣稱「PC-98 音樂已還原」。第 376 輪已接入 BSD 授權 `ymfm`
與 PCM mixer；第 377 輪則證明正常換曲不是 fade，而是 stop 後等待 800ms，
正常 BGM 的 loop count 0 會轉成 `0xFF` 無限循環。剩餘重點是 SFX 真實
caller、save/resume、reload phase 與類比 gain。

2026-07-30 第 372 輪進一步還原 `SOUND.ROM` 的軟體 LFO。IDA 的一般分析
會漏掉 timer ISR `CF47Ah` 透過 `jmp bx` 前往 `CF4C3h`、`CF501h`、
`CF5F3h` 的三條路徑；做 PC-98 ROM 反組譯時，遇到 register-held near
offset 必須手動建立 code path，不能把後段 `db` 當成資料。

可沿用的結論：

- MSCDRV FM opcode `90h` 是 tempo `+4`，與 LFO 無關；
- `SETPARABLOCK` 會自動開啟 modulation，`MODUON/OFF` 只切換聲道 bit 7；
- 六種 waveform 使用 signed WORD phase／step，不是浮點 sine；
- pitch 依 sample、有效 depth 與 base F-number 做兩次 `/32767`；
- amplitude LFO 保留 8086 的 byte-sized signed multiply/divide 與 wrapping；
- 共用核心放在 engine `audio/pc98soundbios`，作品端只映射 NEC parameter。

S98 稽核也要保存「沒有觀測到」的限制。十二首 first-stream 共 18 個聲道
使用非零 LFO 參數，但現有約五秒 Hoot capture 的獨立 pitch／TL update
皆為零。這不能推論原作 LFO 關閉；在 timer cadence 取得 Hoot 長時間 trace
或 NP2kai／test harness 外部證據前，不應把 scheduler 接成假精確。

2026-07-30 第 373 輪先補齊 Sound BIOS 本身的動態證據。selector 9
的三個 FM 聲道都使用
parameter 3 且實際 key-on；45.01 秒 Hoot S98 仍沒有獨立 LFO write，
當時只能確認該 BGM path 沒有執行 LFO。改以 Unicorn 8086
直接執行 exact `SOUND.ROM` 的 command routines 與 `CF5F3h`：

- YM status bit 0 進 `CF501h` Timer A note path，bit 1 進 `CF5F3h`
  Timer B LFO path；
- sync 8 在第 1 tick 建立 29，再由 state 1 低 byte 倒數；
- 第 30 tick reset oscillator 並首次輸出；
- tick 30..80 共 51 組 `A4/A0` 與 51×4 組 TL，counter 最後是 51。

可重用 engine scheduler 因此應以「一次 Timer B overflow」為 API，
不能把 S98 的 10 ms 記錄刻度當成 LFO 固定週期。作品 adapter 只映射
waveform、sync、speed；Timer B register `26h` 到 PCM sample clock、
Timer A gate／length、fade／SFX 與播放器仍是下一層工作。

2026-07-30 第 374 輪進一步證明 CoAB 正常 BGM 根本不走 Sound BIOS
Timer ISR。MSCDRV 保存舊硬體中斷向量後，把 vector 改成自己的
`CS:0F54h`；該 ISR 讀 `188h` status，只測 bit 1，寫 `27h=20h` 後呼叫
track interpreter，結尾寫 `27h=0Ah`。它不測 Timer A、也不鏈回舊 handler。

因此知識庫必須分清兩層：

- Sound BIOS LFO scheduler 是 ROM 本身的 exact、可重用行為；
- CoAB MSCDRV 的 faithful 配樂路徑接管 Timer B IRQ，故不執行該 LFO。

非零 LFO 參數不等於實際聽得到 modulation。其他 PC-98 Gold Box／軟體
必須各自追到 interrupt owner，不能只因沿用同一 parameter format 就啟用
LFO。CoAB 下一步是 register `26h` 與硬體 clock 到 PCM sample clock 的
無漂移 bridge，不是把第 373 輪 scheduler 接進 `TrackPlayback`。

2026-07-30 第 375 輪把上述 bridge 拆成「已證明的完整 count period」與
「尚未證明的 reload phase」。selector 9 的 45.01 秒 S98 header 指定
YM2203 clock `3,993,600 Hz`，且 `2Dh..2Fh` prescaler writes 為零；
`26h` 實際值包含 `BAh／D3h／E3h`。官方 ymfm OPN 核心則交叉提供：

```text
Timer B chip clocks =
  16 × (256 − register26) × 12 operators × prescale
```

YM2203 reset prescale 是 6，所以 CoAB 的完整 count period 是
`1152 × (256 − register26)` clocks。共用 engine 使用整數 numerator 與
remainder 累加 PCM samples，不能每 tick 各自四捨五入，否則曲長會逐步
漂移。

但 `×16` divider 是 free-running；MSCDRV 每次 IRQ 都寫
`27h=20h→0Ah`，首週期仍會受 reload phase 與 CPU I/O timing 影響。
知識庫不可把「長時間無 rounding drift」寫成「每個 IRQ edge cycle-exact」。

2026-07-30 第 377 輪補上遊戲 wrapper 與 driver 之間的重要分層：

- `GAME.EXE MSCPLAY` 先 `MSCSTOP`，再呼叫 Borland `DELAY(0x320)`，最後
  才送新 selector；`0x320` 是 800 毫秒，不是 fade counter。
- MSCDRV queue 只有在 active track 非零時才建立 `0x28` fade counter；
  正常 GAME wrapper 已先 stop，所以不走這條 fade。
- public play 傳 loop count 0；driver 初始化轉成 `0xFF`，FM／PSG end
  consumer 對它不遞減並重設七聲道 cursor，形成無限循環。
- MSCDRV 內部另有 FM0 SFX interpreter 與 request priority，但 request
  欄位只找到清零、未找到正常 GAME 非零 producer；`GAME SOUNDFX` 則走
  Borland `SOUND`／`DELAY`，兩者不能混為一談。

其他 PC-98 Gold Box 版本即使使用相似 driver，也必須同時追遊戲 wrapper、
public ABI、request producer 與 driver consumer，不能看到 fade／SFX
程式碼就假設正常玩家路徑使用。

2026-07-30 第 378 輪把 CoAB 正常短音效追到 `GAME.EXE SOUNDFX` 與
`GAME.OVR` caller。這條路徑不使用 YM2203，也不使用 IBM PC PIT：
Borland `SOUND` 只保存一個 WORD，另一 routine 依該值忙等，交替向
PC-98 port `37h` 寫 `06h／07h`。

可沿用的分析方法：

- 先用 TPOV control chain 分離 code 與 relocation；
- 以 exact `PUSH DS:[constant] + CALL FAR` 找直接 caller；
- 再用 Borland symbol module 判讀功能群；
- raw overlay 載入 IDA 時必須明確設成 8086／16-bit；
- executable 的 data segment 不能假設全檔共用一個 linear→file 差值；
  應從 IDA 匯出資料視窗，再回搜 raw bytes 與整段雜湊。

CoAB 的 42 個直接 caller 分布於 `INTRO／INTERPET／PROTECT／COMSTUFF／
MOVEMENT／SPELLS／GENERIC／LOS`。第 379 輪進一步解析同一份 Borland
symbol table，證明 selector 2–15 分別是 `CASTFX` 到 `CRASHFX`，並把
caller 收斂到 `REALMOVE／ANYUNDEAD／SHOWARROW／CASTSPELL／TWINKLE／
SCAN` 等具名函式。證據鏈應同時包含常數符號、DS address、consumer 與
caller，不能只憑 module 名稱猜成某一招法術。

聲音重建時也要保留語意邊界。原 WORD 是 busy-loop period／count 參數，
不是已證明的 Hz。DOS 的箭矢 selector 2 與 PC-98 `ARROWFX=12` 已直接證明
數字不能跨平台共用；遊戲規則只發 `arrow／spell_hit／step` 等 intent，
平台 adapter 才映射 native selector。

共用 engine 第 379 輪加入 `audio/cyclepcm`，只把 cycles＋level 區段做
duty-cycle 積分，不知道作品 selector。第 380 輪依 NEC V30 官方表更正為
`LOOP taken=5／final=13`，並以 exact 指令路徑分離第一次／後續 gate-on
及一般／最後 gate-off overhead，再把 port `37h` 6／7 gate edge 轉成
PCM。這是值得跨作品沿用的分層：

1. 作品 executable importer 保存 selector、sequence 與 cycle intent；
2. 平台 profile 保存 CPU clock、instruction／I/O wait 假設；
3. 共用 renderer 只做無 drift 的 sample integration；
4. runtime mixer 使用語意事件選擇平台 backend。

目前 8 MHz profile 已可播放且 deterministic。第 381 輪已讓 NP2kai
i286c/V30 core 直接執行 exact routine，證明 OUT 順序，卻同時證明該 core
沿用 80286 `LOOP taken=8／exit=4`，不能作原機 V30 wall-clock oracle。
原機 edge、prefetch 與 wait 仍未校準，必須標為 timing-reconstructed。
