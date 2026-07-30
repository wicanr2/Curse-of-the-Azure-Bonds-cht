# 第三百七十四輪 PC-98 MSCDRV Timer B 中斷所有權

狀態：`READY`（限 MSCDRV 硬體中斷接管、Timer B dispatch 與 CoAB
正常 BGM 不執行 Sound BIOS LFO）

本規格修正第 372／373 輪留下的整合假設。第 373 輪已正確還原
`SOUND.ROM` 的 Timer B LFO scheduler，但當時把 45 秒 Hoot S98 沒有
LFO writes 解釋成 Hoot 的可觀測性限制，並把「接進 `TrackPlayback`」
列為下一步。

指定 IDA Pro 9.4、原始 bytes 與同一份長時間 S98 現在證明：CoAB 的
`MSCDRV.EXE` 會保存舊硬體中斷向量後安裝自己的 ISR；新 ISR 只在 YM2203
status bit 1（Timer B）呼叫音序列 interpreter，不鏈回舊 Sound BIOS ISR。
因此 CoAB 正常 BGM 本來就不執行 `SOUND.ROM CF5F3h` 軟體 LFO。把第 373
輪 scheduler 無條件接進 faithful `TrackPlayback` 反而會產生原版沒有的
pitch／total-level writes。

## 1. 輸入、工具與證據

| 輸入 | SHA-256 |
|---|---|
| `MSCDRV.EXE` | `bddbe63a90078bd9c8c8da5c45417c7ec3afcdf7fd5b724877a83ad9bb7b12f5` |
| `SOUND.ROM` | `f05b508d49f31f2a1a61724f013572592abc0833c09c45a72180e84247dc0d0d` |

依 `AGENTS.md`，主要反組譯在 Docker 內使用
`/home/anr2/ida_94_official/dist` 的 IDA Pro 9.4 ver2。可重現入口：

- [`pc98_mscdrv_timer_bridge_audit.idc`](../../scripts/ida/pc98_mscdrv_timer_bridge_audit.idc)
- [`pc98_sound_rom_timer_bridge_audit.idc`](../../scripts/ida/pc98_sound_rom_timer_bridge_audit.idc)

兩支原生 IDC 會輸出 IDA address、file offset、instruction bytes 與
8086 listing；IDA database、log、ROM、driver 與 S98 都只留本機。
`od` 對 exact input 的 file `0x10E0..0x1229` 與 ROM
`0x347A..0x3500` 另做 raw-byte 交叉驗證。

`cmd/pc98-music-audit` 另以 exact driver SHA 逐 byte 驗證四組版本化
anchors：`INSTALL_MUSIC_TIMER_INTERRUPT`、`INITIALIZE_TIMER_B_26_27`、
`TIMER_B_ONLY_PLAYBACK_DISPATCH` 與
`TIMER_B_RESTART_AND_ACKNOWLEDGE`。它們全在缺失 sector
`0x4000..0x43FF` 之前。

Yamaha《YM2203 FM Operator Type-N (OPN)》資料表說明：

- `24h..26h` 是 Timer A／B 設定值；
- `27h` 控制 Timer A／B load、interrupt enable 與 status reset；
- IRQ 是兩個 timer 共用、可由程式 mask 的中斷輸出。

來源：<https://www.bitsavers.org/components/yamaha/YM2203_198911.pdf>

## 2. MSCDRV 安裝自己的硬體 ISR

`MSCDRV 10EE0h` 的 exact 流程：

1. 讀 PC-98 sound hardware configuration；
2. 由本機 table 選出 DOS interrupt vector；
3. `INT 21h/AH=35h` 取得並保存舊 vector 到 `115EA:115EC`；
4. `INT 21h/AH=25h` 把該 vector 改成 `CS:0F54h`；
5. 寫 `26h=BAh` 作初始 Timer B data；
6. 寫 `27h=02h` load Timer B，再由 `11012h` 寫 `27h=0Ah` 啟用
   Timer B interrupt。

關鍵指令：

```text
10EF6  B4 35       mov ah,35h
10EF8  CD 21       int 21h
10F02  BA 54 0F    mov dx,0F54h
10F0B  B4 25       mov ah,25h
10F0D  CD 21       int 21h
10F21  B3 BA       mov bl,0BAh
10F23  E8 C7 02    call 111EDh       ; register 26h
10F26  B0 27       mov al,27h
10F28  B3 02       mov bl,02h
10F2A  E8 FD 00    call 1102Ah
10F2D  E8 E2 00    call 11012h       ; register 27h = 0Ah
```

`10F37h` 的卸載路徑才用 `INT 21h/AH=25h` 恢復保存的舊 vector。播放 ISR
本身沒有 far call／jump 到該 vector。

## 3. 新 ISR 只消費 Timer B

安裝入口 `10F54h` 讀 `188h` status 並遮罩低兩位：

```text
10F78  BA 88 01    mov dx,188h
10F7B  EC          in  al,dx
10F7C  24 03       and al,03h
10F81  A8 02       test al,02h
10F83  74 17       jz  10F9Ch
10F85  E8 9A 00    call 11022h       ; register 27h = 20h
10F94  E8 DE F1    call 10175h       ; TrackPlayback dispatch
```

結論：

- status bit 0（Timer A）沒有 dispatch branch；
- status bit 1（Timer B）先寫 `27h=20h` 清 flag，再呼叫 `10175h`；
- `10175h` 最後呼叫七聲道 `10410h` stream interpreter；
- 共用尾端 `11012h` 寫 `27h=0Ah`，重新 load／enable Timer B；
- ISR 沒有鏈回保存的舊 Sound BIOS handler。

因此現有 `TrackPlayback.Tick()` 的正確語意就是「一次本作 MSCDRV
Timer B overflow」，不是任意 10 ms tick，也不是 Sound BIOS Timer A。
descriptor header word 1 與 opcode `90h` 寫入 register `26h`，會改變下一次
Timer B overflow 的 wall-clock 間隔。

## 4. 為何 Sound BIOS LFO 不會執行

`SOUND.ROM CF47Ah` 原 ISR 的 exact dispatch 是：

```text
CF494  mov  dx,188h
CF497  in   al,dx
CF498  and  al,03h
CF4A3  test al,01h
CF4A5  jnz  CF501h     ; Timer A note／length
CF4A7  test al,02h
CF4A9  jnz  CF5F3h     ; Timer B software LFO
```

但 MSCDRV 安裝 `10F54h` 後，硬體 IRQ 不再進入這個 handler。MSCDRV
雖透過 `INT D2h` 呼叫 `SETPARABLOCK`／`SETVOLUME`／`NOTE`，仍沒有任何
路徑在每個 Timer B overflow 呼叫 `CF5F3h`。

第 373 輪 selector 9 的 45.01 秒 Hoot S98 已有三個 nonzero-LFO
parameter 3 聲道與實際 key-on，卻只有正常 note burst，獨立 pitch／TL
更新均為零。新 IDA／raw-byte control flow 能解釋這個結果：這不是單純
「Hoot 看不到 Sound BIOS LFO」，而是 Hoot 執行的 CoAB driver 正常路徑
本來就繞過 Sound BIOS timer ISR。

## 5. Remake contract

原版忠實的 CoAB PC-98 BGM path：

- `TrackPlayback.Tick()` 由 YM2203 Timer B overflow 推進；
- tempo 值來自 register `26h`；
- 不把 `pc98soundbios.Modulator` 接進正常 BGM；
- 不產生額外 LFO pitch／TL register writes。

共用 engine 的 `audio/pc98soundbios.Modulator` 仍是有 exact ROM harness
證據的正確可重用元件，可供：

- 真正使用 Sound BIOS timer ISR 的其他 PC-98 軟體；
- Sound BIOS 相容模式／研究工具；
- 未來有獨立證據證明使用該路徑的 Gold Box 版本。

它不能因 CoAB 不使用便刪除，也不能因參數非零便擅自啟用。

## 6. 完成與未完成

本輪完成：

1. MSCDRV 保存／安裝／恢復硬體 interrupt vector；
2. Timer B-only status dispatch；
3. `27h=20h` acknowledge 與 `27h=0Ah` restart；
4. `TrackPlayback.Tick()` 的 Timer B overflow 語意；
5. CoAB 正常 BGM 不執行 Sound BIOS LFO 的三方驗證；
6. exact binary auditor 的四組 Timer B ownership anchors；
7. 清除「下一步把 LFO scheduler 接進 TrackPlayback」的錯誤工作項。

仍未完成：

1. PC-9801 實際 YM2203 clock／prescaler 與 register `26h` 的 wall-clock
   公式交叉驗證；
2. Timer B event 到 PCM sample clock 的無漂移排程；
3. YM2203 FM／PSG 合成器與 mixer；
4. fade、SFX/BGM 共存、完整曲長／loop、pause／save-resume；
5. 遊戲內播放與三平台音訊驗收。

因此本規格只能支持「CoAB MSCDRV Timer B ownership READY」，不能支持
「PC-98 音樂完成」或「遊戲內已可播放」。
