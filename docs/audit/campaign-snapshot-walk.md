# 帶著劇情旗標走：從主線快照出發的段內走訪

由 `cmd/campaign-snapshot-walk` 產生，不要手改。

★ 走訪未達成的那幾十個索引，成因幾乎全是「站得上去，但從段入口走不到」——門要**劇情旗標**才開得了（spec 1193）。冷走每一段都開一支新隊伍、沒有旗標，所以那些門對它永遠是牆；主線有旗標，但**主線只走它要走的路**。這一份拿主線各段的快照，**帶著那一刻的旗標**把那一段走遍。

⚠ 這**不是**「玩家走得到」的證明，是**帶著劇情旗標的幾何可達性**：快照那一刻隊伍真的在那一段、旗標真的是那樣，所以那些門真的開得了；但走訪本身仍是機器的走法，不是劇情路線。

⚠ 選單策略跑四種取聯集（第 1／2／3／最後項）：挑第一項會被收費關卡擋在門外，挑最後一項會在「要離開嗎」直接走人。**單一策略的結果看起來都很合理。**

| 快照 | block | 走到的格子 | 走到的地形碼 | 備註 |
|---|---:|---:|---:|---|
| `ECL1-0x50-世界路線-艾森布拉到希爾斯法` | 0 | 0 | 0 | 讀回來推不進地城：停在1 |
| `ECL1-0x50-立石群-灰袍男子` | 0 | 0 | 0 | 讀回來推不進地城：continue=continue is invalid in mode 0 select=choice 0 is invalid in mode 0 |
| `ECL1-0x51-世界路線-希爾斯法到猶拉什` | 16 | 116 | 19 | 讀回來推不進地城：continue=continue is invalid in mode 3 select=choice 0 is invalid in map mode |
| `ECL3-0x10-猶拉什-地面與指揮部` | 17 | 48 | 11 | — |
| `ECL3-0x11-猶拉什-地下第一層` | 18 | 1 | 1 | — |
| `ECL3-0x11-返回與猶拉什邊界` | 0 | 0 | 0 | 讀回來推不進地城：停在1 |
| `ECL3-0x12-猶拉什-地下第二層` | 17 | 4 | 1 | — |
| `ECL4-0x20-散提爾堡內城` | 33 | 2 | 2 | — |
| `ECL4-0x21-散提爾堡-神殿與牢房` | 34 | 5 | 2 | — |
| `ECL4-0x22-眼魔洞穴與-Dexam` | 0 | 0 | 0 | 讀回來推不進地城：停在1 |
| `ECL5-0x31-哈普村` | 50 | 21 | 6 | — |
| `ECL5-0x32-古熔岩洞與-ECL5-0x33-巫師塔` | 50 | 87 | 8 | — |
| `ECL6-0x40-墓園-紅網` | 64 | 199 | 15 | — |
| `ECL6-0x40-墓園-進入遺跡` | 64 | 184 | 13 | — |
| `ECL6-0x40-墓園-黛米爾公主的祝福` | 64 | 135 | 12 | — |
| `ECL6-0x40-密斯卓諾-墓園` | 64 | 184 | 13 | — |
| `ECL6-0x42-密斯卓諾-外城遺跡` | 66 | 92 | 9 | — |
| `ECL6-0x43-內城遺跡-一樓房間` | 67 | 25 | 6 | — |
| `ECL6-0x43-內城遺跡-二樓與最終戰` | 0 | 0 | 0 | 讀回來推不進地城：continue=continue is invalid in mode 0 select=choice 0 is invalid in mode 0 |
| `ECL6-0x43-內城遺跡-儀式與爪牙戰` | 67 | 24 | 1 | — |
| `ECL6-0x43-密斯卓諾-內城遺跡` | 67 | 18 | 2 | — |
| `inside-block-01-1` | 1 | 76 | 17 | — |
| `inside-block-01-2` | 1 | 76 | 17 | — |
| `inside-block-01-3` | 1 | 125 | 24 | — |
| `inside-block-01-4` | 1 | 108 | 22 | — |
| `inside-block-01-5` | 1 | 125 | 24 | — |
| `inside-block-01` | 1 | 76 | 17 | — |
| `inside-block-02-1` | 2 | 21 | 7 | — |
| `inside-block-02-2` | 2 | 21 | 5 | — |
| `inside-block-02-3` | 2 | 18 | 4 | — |
| `inside-block-02-4` | 2 | 9 | 1 | — |
| `inside-block-02-5` | 2 | 4 | 1 | — |
| `inside-block-02` | 2 | 14 | 2 | — |
| `inside-block-03-1` | 3 | 44 | 5 | — |
| `inside-block-03-2` | 3 | 44 | 5 | — |
| `inside-block-03-3` | 3 | 44 | 5 | — |
| `inside-block-03-4` | 3 | 44 | 5 | — |
| `inside-block-03-5` | 3 | 44 | 5 | — |
| `inside-block-03` | 3 | 7 | 2 | — |
| `inside-block-10-1` | 16 | 116 | 19 | — |
| `inside-block-10-2` | 16 | 116 | 19 | — |
| `inside-block-10-3` | 16 | 4 | 2 | — |
| `inside-block-10-4` | 16 | 122 | 22 | — |
| `inside-block-10-5` | 16 | 122 | 22 | — |
| `inside-block-10` | 16 | 120 | 20 | — |
| `inside-block-11-1` | 17 | 48 | 11 | — |
| `inside-block-11-2` | 17 | 48 | 11 | — |
| `inside-block-11-3` | 17 | 48 | 11 | — |
| `inside-block-11-4` | 17 | 48 | 11 | — |
| `inside-block-11-5` | 17 | 48 | 11 | — |
| `inside-block-11` | 17 | 48 | 11 | — |
| `inside-block-12-1` | 18 | 39 | 7 | — |
| `inside-block-12-2` | 18 | 39 | 7 | — |
| `inside-block-12-3` | 18 | 3 | 1 | — |
| `inside-block-12-4` | 18 | 2 | 1 | — |
| `inside-block-12-5` | 18 | 5 | 2 | — |
| `inside-block-12` | 18 | 1 | 1 | — |
| `inside-block-20-1` | 32 | 93 | 35 | — |
| `inside-block-20-2` | 32 | 93 | 35 | — |
| `inside-block-20-3` | 32 | 93 | 35 | — |
| `inside-block-20-4` | 32 | 93 | 35 | — |
| `inside-block-20-5` | 32 | 93 | 35 | — |
| `inside-block-20` | 32 | 93 | 35 | — |
| `inside-block-21-1` | 33 | 4 | 3 | — |
| `inside-block-21-2` | 33 | 4 | 2 | — |
| `inside-block-21-3` | 33 | 9 | 2 | — |
| `inside-block-21-4` | 33 | 9 | 3 | — |
| `inside-block-21-5` | 33 | 108 | 37 | — |
| `inside-block-21` | 33 | 2 | 2 | — |
| `inside-block-22-1` | 34 | 4 | 3 | — |
| `inside-block-22-2` | 34 | 6 | 2 | — |
| `inside-block-22-3` | 34 | 2 | 1 | — |
| `inside-block-22-4` | 34 | 4 | 3 | — |
| `inside-block-22-5` | 34 | 23 | 5 | — |
| `inside-block-22` | 34 | 7 | 2 | — |
| `inside-block-31-1` | 49 | 231 | 20 | — |
| `inside-block-31-2` | 49 | 231 | 20 | — |
| `inside-block-31-3` | 49 | 231 | 20 | — |
| `inside-block-31-4` | 49 | 231 | 20 | — |
| `inside-block-31-5` | 49 | 231 | 20 | — |
| `inside-block-31` | 49 | 231 | 20 | — |
| `inside-block-32-1` | 50 | 21 | 5 | — |
| `inside-block-32-2` | 50 | 19 | 4 | — |
| `inside-block-32-3` | 50 | 11 | 1 | — |
| `inside-block-32-4` | 50 | 21 | 2 | — |
| `inside-block-32-5` | 50 | 24 | 2 | — |
| `inside-block-32` | 50 | 21 | 6 | — |
| `inside-block-33-1` | 51 | 8 | 3 | — |
| `inside-block-33-2` | 51 | 16 | 3 | — |
| `inside-block-33-3` | 51 | 28 | 2 | — |
| `inside-block-33-4` | 51 | 24 | 1 | — |
| `inside-block-33-5` | 51 | 6 | 1 | — |
| `inside-block-33` | 51 | 19 | 4 | — |
| `inside-block-40-1` | 64 | 184 | 13 | — |
| `inside-block-40-2` | 64 | 169 | 11 | — |
| `inside-block-40-3` | 64 | 141 | 8 | — |
| `inside-block-40-4` | 64 | 199 | 15 | — |
| `inside-block-40-5` | 64 | 145 | 12 | — |
| `inside-block-40` | 64 | 184 | 13 | — |
| `inside-block-42-1` | 66 | 61 | 3 | — |
| `inside-block-42-2` | 66 | 66 | 3 | — |
| `inside-block-42-3` | 66 | 66 | 3 | — |
| `inside-block-42-4` | 66 | 84 | 5 | — |
| `inside-block-42-5` | 66 | 56 | 3 | — |
| `inside-block-42` | 66 | 92 | 9 | — |
| `inside-block-43-1` | 67 | 29 | 1 | — |
| `inside-block-43-2` | 67 | 30 | 6 | — |
| `inside-block-43-3` | 67 | 35 | 6 | — |
| `inside-block-43-4` | 67 | 34 | 5 | — |
| `inside-block-43-5` | 67 | 34 | 5 | — |
| `inside-block-43` | 67 | 24 | 1 | — |
| `inside-block-50` | 0 | 0 | 0 | block 0x50 沒有地形分派 |
| `密斯卓諾-世界路線` | 64 | 178 | 12 | 讀回來推不進地城：停在1 |
| `希爾斯法城內` | 0 | 0 | 0 | 讀回來推不進地城：停在1 |

合計 114 份快照、257 個 (block, 地形碼) 組合。
