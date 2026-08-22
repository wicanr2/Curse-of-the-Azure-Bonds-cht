# 1181 `21h LOAD FILES` 自己會寫的兩格 ECL 記憶體

狀態：`READY`

## 結論

`21h LOAD FILES` 不只是「記下請求、上層去換檔」。handler 本體在載 3D 地圖那一路
會**自己寫兩格 ECL 記憶體**，而那兩格**腳本讀得到**——少寫是控制流分歧，
不是表現層問題。

```pascal
if (o[1] <> 0FFh) and (o[1] <> 7Fh) and (bank0^[1CCh] <> 0) then begin
    bank0^[18Ah] := o[1];          { 剛載進來的地圖 block }
    <overlay-30 entry#9 LOAD3DMAP>(o[1]);
    bank1^[592h] := 0;             { 「地圖要重載」旗標清掉 }
end;
if (o[3] <> 0FFh) and (bank0^[1CCh] = 0) and (DS:728Ah <> 'P') then
    <overlay-29 entry#9 LOADBIGPIC>('y');
```

handler 全文見 [spec 1087](1087-ecl-handlers-six.md)；`o[1]` 是要載的地圖 block
（22 處運算元全列在 [spec 1153](1153-load-pieces-wall-set-params.md)）。

## 三格的位址

用 spec 1098／1155 的換算對回 ECL 格——bank0 是 `4B00h + 位移 / 2`，
bank1 是 `7C00h + 位移 / 2`：

| 原作 | ECL 格 | 內容 | corpus 存取 |
|---|---|---|---:|
| `bank0^[1CCh]` | `4BE6h` | **第一人稱（3D 地圖）模式**：非零才載地圖，為零改走載大圖 | 22 |
| `bank0^[18Ah]` | `4BC5h` | 剛載進來的地圖 block 編號 | 0 |
| `bank1^[592h]` | `7EC9h` | 腳本寫 `FFh` ＝「地圖要重載」，handler 載完清成 0 | 34 |

## 為什麼 `7EC9h` 不能不寫

34 處裡腳本一路寫 `FFh`，而**唯一的讀取端**是

```
ECL2/0x03  0x00F6  COMPARE  7EC9 FF
           0x00FC  IF =
           0x00FD  SAVE     00 7ED5
```

`7EC9h` 仍是 `FFh` 就把 `7ED5h` 歸零，下一句 `COMPARE 7ED5 00 / IF = / GOTO 824b`
因此成立。清掉它的**只有這個 handler**——remake 不寫的話，腳本設過 `FFh` 之後
就永遠停在那個分支上。

## `4BE6h` 是模式閘門

22 處寫入全部落在 ECL1 的三個世界地圖 block（`0x50`／`0x51`／`0x52`），
在 `0`／`1` 之間來回切。配上 handler 的兩條互斥條件（非零載 3D 地圖、
為零載大圖），形狀就是「現在是不是第一人稱畫面」。

⚠ 只宣稱**它決定走哪一路**，不宣稱它在原作的變數名或其他用途。

## remake 這一側

`21h` 的 handler 現在照原作的條件寫那兩格，並在 `RunResult` 上多帶一個
`LoadFilesLoaded3DMap`，讓上層知道原作這一次走的是哪一路——先前上層是自己
從運算元推的。

⚠ 兩格都走 `recordStore`，所以會進 `SaveWrites`：只改 map 不記錄的話，
存檔重載之後狀態會不一致。

## 回歸

- `TestLoadFilesWritesTheMapCellsOnlyInThreeDMode`：四種輸入（閘門開／關、
  片號 `0FFh`／`7Fh`）逐一驗，**沒走那一路時一格都不能動**——`7EC9h` 要維持
  腳本寫進去的 `FFh`。突變驗過（拿掉清 `7EC9h` 那兩行立刻紅）。
- `TestLoadFilesMapCellsReachSaveWrites`：兩格都要進 `SaveWrites`。

## 不宣稱

- `DS:728Ah`（比對 `'P'`）是什麼，以及載大圖那一路的參數 `'y'` 代表哪一張圖。
- `LOAD3DMAP`／`LOADBIGPIC` 的內部；remake 的資產載入仍走既有的
  `ApplyLoadFiles` → GEO 目錄那條路。
- `4BC5h` 在 corpus 裡沒有任何讀取端，所以只知道 handler 會寫它，不知道誰讀。
