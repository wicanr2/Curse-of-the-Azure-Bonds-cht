# Gold Box ECL command-set knowledge base

這份文件是後續 SSI Gold Box 作品共用的 ECL 反組譯入口。它把「已知
opcode 名稱／arity」與「已證實 runtime semantics」分開；command table 本身
只能證明 instruction cursor 要吃幾個 operand，不能自動證明 engine side effect。

## 已整理的 command table

| opcode | command | operands | 目前狀態 |
|---|---|---:|---|
| `00`–`02` | EXIT／GOTO／GOSUB | 0/1/1 | bounded control flow |
| `03`–`08` | COMPARE／ADD／SUBTRACT／DIVIDE／MULTIPLY／RANDOM | 2/3/3/3/3/2 | bounded VM |
| `09`–`14` | SAVE／LOAD CHARACTER／LOAD MONSTER／SETUP MONSTER／APPROACH／PICTURE／…／COMPARE AND | 2/1/3/3/0/1/…/4 | partial signal／bounded |
| `15`–`1F` | menus／IF variants／CLEARMONSTERS／party checks | variable／fixed | partial |
| `20`–`2C` | NEWECL／LOAD FILES／surprise／COMBAT／ON branches／treasure／menus／PARLAY | variable | partial signal／bounded |
| `2D` | CALL | 1 | **未實作；需確認 external dispatch 或 code call** |
| `2E` | DAMAGE | 5 | bounded raw signal（flags／dice／bonus／save flags） |
| `2F`–`30` | AND／OR | 3/3 | bounded 16-bit memory destination |
| `31`–`40` | sprite／item／clock／save table／NPC／pieces／PROGRAM／WHO／delay／spell／protection／… | variable | partial signal／bounded |

完整 arity source of truth 是 `internal/ecl/operand.go` 的 `KnownCommands`；
不要在這份文件另抄一份可漂移的完整 table。每個真正加入 VM 的 opcode 都必須
先新增 `docs/spec/<round>-...md` 的 `READY` contract，再加 synthetic 與 real-image
regression。

## Operand contract

目前已由 parser／既有 arithmetic regression 證實：`0x00` 是 byte literal；
`0x01`、`0x03` 可讀 memory word；`0x02` 是 word literal；`0x80` 是 packed
text；`0x81` 是 string-memory word。三 operand arithmetic 與本輪 `AND`／`OR`
使用前兩個 value、第三個 word destination，但這個形狀不應推廣到其他 command。

## Cross-game reuse rule

後續 Pool、Secret、Savage Frontier 等作品應重用 parser、command metadata、
bounded runner 與 evidence report；作品特有的 block namespace、memory layout、
external `PROGRAM` routine、monster table 與 save side effect 必須由各作品 adapter
注入。若同一 opcode 在另一作品出現不同 arity 或 operand code，優先保留 raw trace，
不要覆寫本表的 CoAB assumption。

## Current evidence boundary

ECL1–ECL6 entry smoke 已實際遇到 `0x1D PARTYSTRENGTH`、`0x22 PARTY SURPRISE`、
`0x2D CALL`／`0x2F AND`。`AND`／`OR` 已是 bounded memory operation，`CALL` 已建立
external-call signal；本輪再加入兩個 party-rule destination signal。它們都不把缺少的
作品 party context 猜成 VM memory side effect。

最新 smoke evidence 顯示，補上 variable monster operands 後，ECL3 block 17／18、
ECL4 block 33／37 等真實 entries 已能抵達 COMBAT 並產生 spawn signal；這是 bounded
descriptor bridge，不是完整外部 routine 或玩家流程完成的證明。

`0x2D CALL` 現在也有 reusable external-call boundary：保存 observed word address
並 return 到下一個 ECL instruction。後續作品可重用 signal contract，再依各作品
的 recognized address table 注入 routine handler；不可把它直接改成 `GOSUB`。

`0x2E DAMAGE` 的公開 CoAB reference operand 順序已確認為
`flags, dice_count, dice_size, damage_bonus, save_flags`。VM 現在只保存這五個 raw
numeric values 並繼續 cursor；target selection、saving throw、signed bonus、random
roll 與 party HP mutation 仍由作品 adapter 實作，避免把 ECL flags 誤當成一般攻擊。

`0x34 ECL CLOCK` 是 reference `vm_LoadCmdSets(2)` 的兩個 numeric operands：
`timeStep`、`timeSlot`。CoAB 的 bounded VM 現在輸出 raw `ClockRequest`，State 再呼叫
共用 `AdvanceGameTime(timeSlot, timeStep)`；這個「VM signal → 作品 time adapter」分層
可重用到其他 Gold Box 遊戲，但各作品仍須重新驗證 slot scale 與事件觸發規則。

`0x35 SAVE TABLE` 的 reference operand contract 是 `value, address, offset`：先把
前兩個 numeric operand 解成 value 與 destination base，再把第三個 numeric operand
加到 destination，將 value 寫入 `memory[address+offset]`。這與 `0x2A GETTABLE`
的 indexed read 成對；bounded VM 已實作兩者，並以 synthetic memory regression 保護
「operand 可是 literal 或 memory value、offset 可是 literal 或 memory value」的語意。
這是可跨 Gold Box 重用的 raw ECL memory primitive，作品專屬的 table schema 仍須另行
由反組譯證據建立。

`0x24 COMBAT` 是 resumable engine boundary，但不保證進入 battle loop。CoAB reference
`CMD_Combat` 在無怪物且 combat type 為 normal 時，先依 `Area2.EnterShop`／
`EnterTemple` 派送 `CityShop`／temple shop，兩個 flag 都未設定才做一般戰後寶物；
有怪物或特殊 combat type 才進入戰鬥。CoAB ECL mirror 已確認 EnterShop=`0x7F6C`、
shop price scale=`0x7F6D`、EnterTemple=`0x7EE2`。VM 必須先輸出 typed service／combat
signal並把 PC 推進到 COMBAT 後，State adapter 在 service 關閉或 party victory 後以同一
`RuntimeState` 續跑，才能接回 text、menu、PICTURE 或 `NEWECL`。

CityShop 的 stock 由前置 `TREASURE` 載入 `items_pointer`；購入時加入商品的
`ShallowClone()`，不移除 stock entry。`field_6DA` 的 `01/02/04/08` 是右移
4/3/2/1，`20/40/80` 是左移 1/2/3，其他值（包括 Tilverton Weaponers 使用的
`0x10`）維持原價。付款先用 selected player 的 typed-coin gold worth，再用 pooled
money；進店會清空既有 pool。後續作品可重用 resumable dispatch 形狀，但 Area2 位址、
service priority、ITEM namespace、價格表及貨幣規則都必須重新驗證，不能把 CoAB 常數
升格成通用 VM semantics。

Temple 沿用同一 external-service boundary，但交易表不同：`ovr005.temple_shop`
提供 Heal／View／Pool／Share／Appraise，`temple_heal` 再列出十個固定價格的
condition／HP services。可共用的是 EnterTemple signal、typed-money payment 與
resume contract；治療 affect IDs、價格、Raise Dead stat penalty 必須逐作品驗證。

ECL event text 也採同一 evidence discipline：只有已由 raw image 解出的 segment
才進入作品 locale catalog，未知句子維持原文，避免跨作品誤套 CoAB 翻譯。

事件 segment 與 menu pause 是兩個不同 observable outputs；State 必須先保存 text，
再切換 choices／waiting state。這項 ordering 可供後續 Gold Box 遊戲共用。

`PRINT RETURN` 同樣只保存 renderer-facing cursor signal，不應誤當成 VM `RETURN`；
它不觸碰 ECL call stack。

`FIND ITEM`／`DESTROY ITEMS` 是另一組可跨作品重用的 inventory boundary；VM 只
保存 item IDs；有作品 party context 時，`FIND ITEM` 依全隊 raw item types 設定
`=`／`<>`，缺 context 時維持 unresolved。`DESTROY ITEMS` 會更新同-run working view，
真正 persistent roster mutation仍由作品 State adapter負責。

`PARTYSTRENGTH (0x1D)` 與 `PARTY SURPRISE (0x22)` 現在由 bounded VM 保存已驗證的 word
destination request 並繼續 cursor；State 注入 `PartyContext` 時會依 reference 計算並寫回
shared ECL memory，沒有 roster context 的純 VM path 則維持 unresolved request，不寫入
猜測值。

`CHECKPARTY (0x1E)` 已釋放三個 reference branches：normalized selector `0xA5..0xAC`
查 thief skills、`0x9F` 查 movement、`8001` 查 active affect；四個 destination 的
min／max／average／found ordering 由 `PartyContext` 寫入。未知 selector 仍只保存 raw
request，這個規則可跨 Gold Box 重用，但各作品需重新驗證 selector table。

`WHO (0x39)` 是一個獨立的 character-selection boundary：它使用目前 prompt 呼叫
`selectAPlayer`，不是從 operand 產生選項。VM 現在保存 `WhoRequest.Prompt`，interactive
path 會 pause，State 以 roster UI 消費 selection，再由 shared `RuntimeState` resume；
selected player ID 已保存，但其他 global routine side effects 仍需各自驗證。

`LOAD CHARACTER (0x0A)` 與 WHO 不同：reference 從 operand value 取得 1-based player
selector，低 7 bits 對應 `TeamList[selector]`，bit 7 是 restore／party-summary redraw
flag。共用 VM 會同時保存 raw word address 與 `LoadCharacterRequest`；CoAB State 已將
有效 selector 接回 persistent roster 的 selected player，無效 selector 保留 not-found
狀態。`FreeCurrentPlayer`、external string context 與 redraw side effects 仍須各作品
依 reference 逐欄接線。

已核對的 `vm_CopyStringFromMemory` 特例是 `0x7C00`：它讀取目前 `SelectedPlayer.name`。
因此共用 runtime 只由作品 `PartyContext` 提供 roster name，並在 `LOAD CHARACTER` 後更新
`RuntimeState.Strings[0x7C00]`；後續 `0x81` `COMPARE`／`PRINT` 可使用該姓名。這不代表
`0x4B00`、`0x7A00`、`0x8000` 等其他 DOS memory regions 已完成。

`FIND SPECIAL (0x3F)` 查的是目前 `SelectedPlayer` 的 active affect，不是全隊 affect
query；`0x3D` 則是 CLEAR BOX，兩者不能因 opcode 接近而混用。共用 `RuntimeState`
保存 selected-player index，讓 `LOAD CHARACTER` 與 WHO selection 都能餵給後續
`FIND SPECIAL`，並跨 pause／shared BlockSession 保留；有 context 時設定 `=`／`<>`，
缺 context 或未選角時維持 unresolved request。

`DUMP (0x3E)` 是角色離隊 command，不是 debug dump。reference 的
`FreeCurrentPlayer(selected, free_icon=true, leave_party_size=false)` 會移除 TeamList member、
減少 party size並回傳前一位／新第一位作為 selection；空隊伍回傳 null。共用 VM 以
ordered `DumpRequest` 更新 working party，作品 State 再同步 persistent roster／fighter。
它與玩家操作的 ALTER DROP 不同，不能強制「至少留一人」。

目前 CoAB 的 State adapter 已把 verified `DESTROY ITEMS` IDs 廣播到 persistent
party roster；這是 ECL effect 的明確 mutation，與玩家操作用、會保護 readied item
的 `Character.RemoveItem` 不同。後續 Gold Box 作品可沿用「VM signal → 作品 party
adapter」分層，但仍須各自驗證 item type namespace、compare result 與角色範圍。
CoAB 的 `PROGRAM (0x38)` 已確認 0/3/8/9 都是「停止本輪 VM，再交給 engine」：
0=start menu、3=party killed、8=game won／全隊恢復／詢問存檔、9=encamp。可重用 VM
只應回傳 ID 與 boundary；作品 State adapter 才能決定 process exit、title screen 與存檔
UI。CoAB remake 讓一般選單與 combat continuation 共用同一 adapter，避免戰後吞掉勝利。

`CALL (0x2D)` 的 raw operand 不能直接當 code pointer：CoAB dispatch 先做 unsigned
`operand - 0x7FFF`。已觀察 raw `0x2E10/0xC01E/0xB200` 分別是 redraw、forced
`MovePositionForward` 與 sound A/B。forced move 是 16×16 cardinal wrap 且不檢查碰撞；
玩家按鍵 movement 的門／牆阻擋不可誤套到 script CALL。其他作品必須重新驗證 dispatch
base 與 address table，不可直接沿用 CoAB 位址。

CoAB `ADD NPC (0x36)` 是 ID＋morale 兩 operands。真實 block 的第二 operand 常以
`0x00,value` 編碼；少吃一個 operand會把 `0x00` 誤判成 EXIT，形成看似通過的假 boundary。
corpus gate除「無 unsupported opcode」外，也應鎖定 steps、PC、signal sequence 與後續文字／
COMBAT，才能抓出這類 framing 錯誤。`DELAY (0x3A)` 是無 ECL-memory side effect 的 engine
timing signal，runner應計數並繼續，renderer再決定實時呈現。
