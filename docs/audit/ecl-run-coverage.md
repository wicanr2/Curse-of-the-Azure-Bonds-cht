# 實跑指令覆蓋（實跑路線執行過 ECL 的哪些指令）

由 `cmd/ecl-run-coverage` 產生，不要手改。

★ 存在的理由：「全城市／全房間走訪（地形分派以外的部分）」先前沒有分母。玩家看得到的內容都是被執行的 ECL 指令，所以拿「實跑路線執行過的指令」對上 `cmd/ecl-effect-coverage` 的可達指令集，序章、支線、可選房間、查表分派的分支全部落進同一個分母，不必一種內容一種盤點。

⚠ **覆蓋率是下界。** 實跑路線只走它要走的路；「沒執行到」不等於「走不到」，更不等於「沒接」。它的用途是把「還沒走過的內容」變成一張可排工作的清單。
⚠ 分母是靜態控制流走訪的**可達**指令（與 `ecl-effect-coverage` 同一套）：IF 的兩側、查表分派的每個目標都在分母裡，但原作真的到不了的死分支也在——分母本身就標明過這個限制。
⚠ 「圖外執行」＝實跑執行了、靜態走訪卻不認得的指令。**不是 0 就要查**：要嘛走訪器有洞，要嘛記錄器把合成 fixture 混進來了。

## 覆蓋輸入怎麼產生（要與程式碼同版）

```bash
rm -f workplace/ecl-run-coverage.lines
COAB_ECL_COVERAGE=/src/workplace/ecl-run-coverage.lines \
  ./tools/go.sh test ./internal/game -run 'TestReal|TestTilvertonRoute' -count=1
COAB_ECL_COVERAGE=/src/workplace/ecl-run-coverage.lines \
  COAB_ROUTE_JSON=/src/workplace/campaign-frames/route-clean-716.json \
  ./tools/go.sh test ./cmd/azure-bonds-game -run TestKeysDriveARealSessionFromTheTitle -count=1
./tools/go.sh run ./cmd/ecl-run-coverage -coverage workplace/ecl-run-coverage.lines
```

實跑的定義沿用 `cell-reachability`：主線 real-session 測試＋提爾佛頓走訪，再加按鍵重放那一場（輸入層的實跑）。

## 逐段覆蓋

| 段 | 可達指令 | 實跑執行 | 覆蓋 |
|---|---:|---:|---:|
| `ECL2.DAX/0x01` | 659 | 590 | 89.5% |
| `ECL2.DAX/0x02` | 374 | 307 | 82.1% |
| `ECL2.DAX/0x03` | 703 | 437 | 62.2% |
| `ECL2.DAX/0x04` | 482 | 282 | 58.5% |
| `ECL3.DAX/0x10` | 913 | 624 | 68.3% |
| `ECL3.DAX/0x11` | 788 | 471 | 59.8% |
| `ECL3.DAX/0x12` | 840 | 470 | 56.0% |
| `ECL3.DAX/0x15` | 334 | 145 | 43.4% |
| `ECL4.DAX/0x20` | 722 | 413 | 57.2% |
| `ECL4.DAX/0x21` | 563 | 261 | 46.4% |
| `ECL4.DAX/0x22` | 526 | 343 | 65.2% |
| `ECL4.DAX/0x23` | 303 | 58 | 19.1% |
| `ECL4.DAX/0x25` | 745 | 73 | 9.8% |
| `ECL5.DAX/0x30` | 72 | 56 | 77.8% |
| `ECL5.DAX/0x31` | 388 | 299 | 77.1% |
| `ECL5.DAX/0x32` | 634 | 460 | 72.6% |
| `ECL5.DAX/0x33` | 709 | 377 | 53.2% |
| `ECL5.DAX/0x35` | 598 | 321 | 53.7% |
| `ECL6.DAX/0x40` | 766 | 661 | 86.3% |
| `ECL6.DAX/0x42` | 603 | 503 | 83.4% |
| `ECL6.DAX/0x43` | 525 | 454 | 86.5% |
| `ECL6.DAX/0x45` | 320 | 156 | 48.8% |
| `ECL1.DAX/0x50` | 798 | 432 | 54.1% |
| `ECL1.DAX/0x51` | 726 | 255 | 35.1% |
| `ECL1.DAX/0x52` | 86 | 0 | 0.0% |
| **合計** | **14177** | **8448** | **59.6%** |

### 已判定的 0%

| 段 | 判定 |
|---|---|
| `ECL1.DAX/0x52` | **demo**（spec 278：`gbl.inDemo` 才進得去，正常 new game 是 `0x01`）——0% 是設計事實，不是缺路線 |

## 沒蓋到的最大叢集（前 20 個）

★ 叢集＝可達指令依位址排序後，連續沒被執行的一段。大叢集通常就是「整間房／整條支線沒走過」；要提高覆蓋，從這裡挑路線補走。

| 段 | 起點位址 | 連續未執行指令數 |
|---|---|---:|
| `ECL4.DAX/0x25` | `0x84C6` | 409 |
| `ECL1.DAX/0x51` | `0x8DFD` | 189 |
| `ECL4.DAX/0x25` | `0x920F` | 149 |
| `ECL5.DAX/0x33` | `0x91AF` | 148 |
| `ECL4.DAX/0x23` | `0x8054` | 122 |
| `ECL4.DAX/0x23` | `0x86EE` | 98 |
| `ECL5.DAX/0x35` | `0x8F79` | 93 |
| `ECL1.DAX/0x52` | `0x8014` | 86 |
| `ECL4.DAX/0x21` | `0x8FCE` | 85 |
| `ECL2.DAX/0x04` | `0x872E` | 84 |
| `ECL4.DAX/0x21` | `0x8A5F` | 82 |
| `ECL3.DAX/0x15` | `0x8711` | 78 |
| `ECL3.DAX/0x15` | `0x811B` | 75 |
| `ECL3.DAX/0x12` | `0x8C60` | 74 |
| `ECL1.DAX/0x50` | `0x94AB` | 73 |
| `ECL6.DAX/0x45` | `0x810A` | 73 |
| `ECL1.DAX/0x50` | `0x8853` | 72 |
| `ECL4.DAX/0x20` | `0x979B` | 71 |
| `ECL4.DAX/0x25` | `0x82CF` | 71 |
| `ECL3.DAX/0x10` | `0x916D` | 64 |

## 圖外執行

實跑執行了、靜態走訪不認得的：**0** 條。
