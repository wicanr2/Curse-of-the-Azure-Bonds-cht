# 第五百七十三輪：`EFFPROCS` 效果 handler（第一批 12 個）

狀態：`READY`（限本文件列出的 12 個 handler 的指令內容）。日期：2026-08-14

## 結論先行

`EFFPROCS`（overlay-12）是效果 handler 的集中地，136 個 entry。本輪逐條讀完
最小的 12 個，並得到兩個跨規格的印證：

**一、它們就是修改命中骰的那批效果。** 第 571 輪讀出 `TRYTOHIT` 把 d20 存進
全域 `DS:A039h`，並在 `CHECKFX(10h, …)` 之後才比較。本輪看到多個 handler 直接
`inc`／`sub` 這個 `DS:A039h`——**證實效果是靠改寫該全域來影響命中**，不是回傳
修正值。

**二、`DS:A02Fh` 是傷害屬性旗標，且可讀出位元。** 第 412 輪已知 `70h` 與 `87h`
分別在「含 Fire」「含 Electricity」時呼叫保護處理。本輪讀出確切位元：

| handler | 條件 | 位元 |
|---|---|---|
| `249Dh` | `DS:A02Fh and 1` | bit 0 |
| `2461h` | `DS:A02Fh and 2` | bit 1 |
| `2B87h` | `DS:A02Fh and 4` | bit 2 |

依 spec 412 的效果對應，bit 0 是 Fire、bit 2 是 Electricity；**bit 1 尚未有
對應效果證據，不命名**。三者命中時都以參數 `0` 呼叫同一個 routine
（overlay-12 local `1Bh`）。

## 逐一內容

| 位址 | 內容 |
|---|---|
| `008Dh` | `DS:A02Ch` 加一、`DS:A039h` 加一（命中骰 +1） |
| `009Eh` | `DS:A03Ch` 加 5、`DS:A039h` 加一 |
| `0BC8h` | `DS:A039h` 減 4、`DS:A02Ch` 減 4 |
| `1713h` | `DS:A035h := 1`、`DS:A039h` 減 4 |
| `106Dh` | `DS:A030h := DS:A030h ÷ 2`（有號除法） |
| `1A1Dh` | `DS:A02Eh := DS:A02Eh ÷ 2`（有號除法） |
| `2461h` | `DS:A02Fh and 2` 非零時呼叫 `1Bh(0)` |
| `249Dh` | `DS:A02Fh and 1` 非零時呼叫 `1Bh(0)` |
| `2B87h` | `DS:A02Fh and 4` 非零時呼叫 `1Bh(0)` |
| `243Ah` | 依序以常數 `0Bh` 與 `35h` 呼叫 `1Bh` |
| `2BA0h` | 經兩層 far pointer（`arg_6` → `+18Eh`）把目標 record 的 `+6` 清 0 |
| `0075h` | 以 `arg_6`／`arg_8` 呼叫外部 routine 並檢查回傳值 |

所有 handler 的呼叫慣例一致：`retf 0Ah`（5 個 word 參數）。

## 這份規格明確不宣稱

- **`DS:A02Ch`／`A02Eh`／`A030h`／`A035h`／`A03Ch` 的欄位語意**。本輪只描述
  handler 對它們做了什麼。要命名成 AC、傷害、速度等必須先找到 writer 與
  consumer，這正是 `AGENTS.md` 反覆強調的那條線。
- **`1Bh` 這個 routine 的實作**。第 412 輪稱它為 `Protected`，本輪只確認
  三個傷害屬性分支都以參數 `0` 呼叫它。
- **`DS:A02Fh` bit 1 的屬性**。有位元、沒有對應效果證據就不命名。
- **哪個 `MON*SPC` 效果編號對應哪個 handler**。entry index 與效果編號的對應在
  spec 412 有部分結果，本輪沒有擴充。
