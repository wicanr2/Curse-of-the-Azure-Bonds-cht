# 第二百六十三輪：DOS player age patch CLI

狀態：`READY`

## Contract

`cmd/azure-bonds` 現在提供安全的年齡修改入口：

```sh
go run ./cmd/azure-bonds \
  -set-age 42 \
  -character-record CHRDATA1.sav \
  -out-record /tmp/CHRDATA1-age42.sav
```

流程會先以既有 DOS player parser 驗證 record，再只修改 `Player.age` 的 signed
little-endian 欄位 `0x76..0x77`，透過 `PatchDOSPlayerRecord` 輸出新檔。輸入檔永遠不會
被覆寫；未知 bytes、raw class levels、spell、save 與其他已保存欄位由 raw-preserving
patch 保留。

## Boundary

這是單一 `.SAV/.GUY` player record 的修改器，不等於完整 `SAVGAM` slot transaction。
若要讓遊戲讀取修改後角色，仍需把輸出檔放回對應的 `CHRDAT{slot}{n}.sav` bundle，並
依既有 slot loader 規則處理 `.GUY/.FX/.SWG` sidecars。年齡範圍限制為 signed int16，
避免 CLI 產生 parser 無法表示的值。

## Verification

既有 `PatchDOSPlayerRecord` age round-trip regression 加上 `cmd/azure-bonds` build，
覆蓋欄位位置與 CLI 編譯 contract；未知 bytes preservation 仍由 party parser tests
負責。
