# PC-98 VFD runtime trace

平台：NP2kai `0.86.0.22`／commit `e2dc904`，Docker／Xvfb，PC-9821 core。
輸入只使用由原始 VFD 產生的可拋棄 D88；使用者提供的 FDD 未修改。

## Baseline

Disk A 保留 VFD 的兩個 absent-sector descriptors。D88 backend 對
`C=3 H=0 R=8 N=3` 的 read request 記錄：

```text
COAB_MISSING_READ count=1 C=3 H=0 R=8 N=3
COAB_MISSING_READ count=2 C=3 H=0 R=8 N=3
COAB_MISSING_READ count=3 C=3 H=0 R=8 N=3
COAB_MISSING_READ count=4 C=3 H=0 R=8 N=3
```

畫面進入 MEGDOS 0.25 並停在 `loader.com`。CPU soft-interrupt probe 在這段
時間內沒有看到 DOS `INT 21h/AH=4Bh`，所以不能把四次 read 說成
`LOADER.COM` 內三次 EXEC 的結果；它們發生在 shell 進入 loader code 以前。

## Second-read synthesized-zero experiment

第一次 read 維持 sector-not-found，第二次起由 emulator 暫時回傳 1024
bytes 零資料：

```text
COAB_MISSING_READ count=1 C=3 H=0 R=8 N=3
COAB_MISSING_READ count=2 C=3 H=0 R=8 N=3
COAB_MISSING_READ synthesized-zero count=2
```

結果停在 MEGDOS banner，沒有出現 `loader.com`，也沒有 DOS EXEC trace。
因此「第一次缺失、第二次補零」仍不是原磁片的正確讀取語意。

## 結論邊界

- exact：同一 CHRN 在 baseline 被請求四次。
- exact：第二次回傳零資料會改變啟動狀態，且尚未進入 loader EXEC。
- hypothesis：該 sector 涉及 copy protection、弱磁區、多重讀取結果或
  其他 VFD 未保存的低階語意。
- 不成立：把 `0xFFFFFFFF` 永久視為零填 sector。
- 待驗證：MAME canonical FDI 對同一 CHRN 保存的 sector status／payload。

