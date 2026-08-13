# 架構邊界與早期確認事項

獨立 engine／JSON game pack 的分工，以及當時已確認與尚未確定的清單。

> 本檔由 2026-08-13 的 `CONTEXT.md` 分冊而來，內容逐行保留原文。
> 歷史敘述以當時證據為準；與目前 worktree、READY spec 或
> `docs/knowledge/coab-re-coverage-matrix.md` 衝突時，以後者為準。

## 架構轉向：獨立 Golden Box engine + JSON game pack

使用者要求停止把 CoAB 劇情資料直接 hardcode 在 `internal/game/state.go`。
已建立並推送獨立 repo
[`wicanr2/golden-box-remake-engine`](https://github.com/wicanr2/golden-box-remake-engine)：
通用 Go／Ebiten
engine、ECL/DAX/GEO codec、戰鬥規則、JSON schema 與 declarative event runtime
放在該 repo；本 repo 僅保留 CoAB 原版證據、繁中翻譯、素材、攻略、截圖、
JSON game pack 與端到端整合測試。

禁止新增 CoAB 專屬旗標、block mapping、座標、敵人編成、NPC 離隊、英文文字
比對或繁中敘事到共用 Go engine。若 JSON 尚無法表達，先擴充獨立 engine
schema/runtime，再以 CoAB JSON 描述。

第一個遷移驗證是摩安德之坑離場：`block 51 + 4C5B=FF + 7F12=1`、依 script
name 移除 ALIAS／DRAGONBAIT、Alias 生死告別分支及返回 pending world menu。
上述資料已移至 `gamepack/events/pit-of-moander.json`；舊的
`applyPitOfMoanderDeparture` 已刪除，由通用 `applyDataPackEvent` adapter
投影 engine runtime。散提爾堡巡邏、城外、盤問、伏筆、內城繁中與手札 32
也由 JSON `text_rules` 驅動，不再新增 Go 字串分支。

目前另有尚未提交的散提爾堡 continuation：真實 ECL session 已由離坑後
`JOURNEY ON → ZHENTIL KEEP → TRAIL` 經巡邏放行，抵達 world value `12`，
`ENTER CITY` 切入 ECL4/GEO4 block `0x20`，盤問並解鎖手札 32，最後在
`(2,0,S)` 進入 Dungeon mode。保留此成果，但新增劇情文字必須改由 JSON 驅動。

詳細協作限制見根目錄 [`AGENTS.md`](AGENTS.md)。兩個 repo 必須維持獨立歷史，
仍採重大成果才集中 commit + push。

使用者另確認「地圖畫面也必須 remake」：範圍同時包含 GEO/WALLDEF/8X8D
組成的第一人稱城市／地城探索視窗，以及荒野／世界旅程地圖。現有 Far/Mid/Near
wall traversal、碰撞、門與 sky 只是初步證據；完整牆面拼圖、遮擋、door overlay、
camera 與 640×480 nearest-neighbour 呈現應遷入獨立 engine renderer，CoAB JSON
只保存 map set/block、素材 manifest 與作品規則。

最新成果：摩安德之坑已可完整撤離。祭壇藏寶後回到上層 `(0,12)` 會觸發
10 名教徒、5 名蕈人與 5 名蔓生怪的最後阻擊；勝利返回地城後在 `(0,11,W)`
執行原出口 handler，寫入 `4C5B=FF／7F12=1` 並 `NEWECL 0x51`。愛麗雅絲與
龍餌依原劇情繁中告別並從 persistent roster／combat projection 離隊，Continue
才回世界選單。combat continuation 現保存 dungeon return mode；packed-text
工具也會輸出 block-relative offset，成果收進 READY spec 340、繁中攻略與
Gold Box ECL／state 知識庫。

先前成果：摩安德祭壇藏寶與手札 20 已接通。取得護手後回到上層 `(12,0)`，
面東或西主動 SEARCH，ECL3 selector `0x10` 會一次性給予 20 顆寶石、6 件珠寶，
並從 ITEM3 block `0x10` 解出 `+2` 精製牧師卷軸與 `+2` 護手／手套。零
monster spawn 的 COMBAT 現正確視為 treasure-service boundary；關閉財寶選單
會 resume 同一 ECL session，顯示繁中神殿地圖、解鎖手札 20 並返回地城。
ITEM1～6 已在整合流程以 chapter namespace 載入。成果收進 READY spec 339、
繁中攻略與 Gold Box state 知識庫。

先前成果：摩貢首戰、三塊摩安德殘軀與護手已接通。首戰勝利會沿同一 ECL3
block `0x12` runtime 關閉異次元裂隙；三塊已穿越的殘軀長出數百張嘴尖叫後，
以 MON3 `0x1A BIT O' MOANDER` 進入第二戰。每塊 140 HP、CombatSize `4`
（2×2），名稱繁中為「摩安德殘軀」。勝利後找到摩安德護手，下一個
continuation 寫入 `4C5B=1`，證實神器是 plot flag 而非普通 inventory item；
祭司隨後高喊「他們殺了神」。post-combat runtime budget 由 180 提升為 bounded
500，修復長連戰 script 在 payload offset 7078 中止的問題。成果收進 READY
spec 338、繁中攻略與 Gold Box state 知識。

先前成果：主線已從尤拉什跨入摩安德之坑。取得指揮官通行後，隊伍可抵達
`(11,0,N)` 的 terrain `0x26`，看到原版巨坑警告與 picture pause；向北跨界
由 ECL3/GEO3 block `0x10` 切到 `0x11`，同步 `(0,0,E)`。入口的三名死去
邪教徒、受傷牧師狂熱呼喊、擊石封死入口、咳血死亡與遠方戰鬥／烤麵包氣味
已完整繁中，最後返回可探索 Dungeon mode。成果收進 READY spec 334、中文攻略、
README 與 Gold Box state 知識庫。

先前成果：尤拉什等候室後的散塔林間諜與紅羽衛指揮官主線已接通。
從 `(0,3,E)` 踏入 `(1,3)` 的 terrain `0x9A` 會依 ECL3 SearchLocation 建立
MON3 `0x13 ×1／0x14 ×8／0x15 ×2`，繁中顯示散塔林牧師、戰士與法師。
勝利後晉見指揮官，選 NICE 可取得自由通行並解鎖繁中手札 22／52，最後由側門
回到 `(1,3,E)`；`4C41` 保證事件只觸發一次。共用 lifecycle 也修正空 packed
text 被誤判為可見事件、吞掉 SearchLocation 的問題。成果收進 READY spec 333、
中文攻略與 Gold Box state 知識庫。

先前成果：主線已由希爾斯法延伸到尤拉什戰區。`YULASH→TRAIL` 因隊伍仍帶
散塔林枷印，固定遭十二名紅羽衛指控為間諜並開戰；勝利寫入 `4C9B=10`。
ENTER CITY 由 world ECL `0x51` 切至 ECL3 block `0x10`，走通廢墟入口、
首次騎士／紫衣女子事件、紅羽衛檢查哨、PARLAY 與護送。等候室 Continue
現正確同步 Area 3、GEO3 block `0x10` 與 `(0,3,E)`，不再帶著法師塔舊 GEO
座標進尤拉什。成果收進 READY spec 332 與 Gold Box state 知識庫。

先前成果：世界主線已由艾森布拉延伸到希爾斯法。隊伍返回立石群後選
HILLSFAR／TRAIL，ECL world dispatcher 由 `0x50` 切入 `0x51`，先遭第二次
event 12 的六名偽裝火刀伏擊；勝利以 `4C9B=11` 抵達希爾斯法。State 現完整
投影 Area 1 A–N 的 zero-based `0..13` world values，不再只認南方城市。
PICTURE 80 城市入口、六場所與碼頭酒館已接通；仍帶 Fzoul 枷印時，RELAX
會遭紅羽衛打翻酒水挑釁，拒絕後與六名原版 fighter 戰鬥並返回酒館。成果收進
READY spec 331、跨區 real-session regression 與 Gold Box state 知識庫。

先前成果：艾森布拉 ECL1 block `0x50` 的露天酒館已走通完整
`BAR → RELAX／HAVE A DRINK → Continue → EXIT`。`RELAX` 依原 script
觸發 Tavern Tale 60，手冊正文為龐大身影飛越森林向南；飲酒選單解出
`BEER/ALE/PORT/MEAD/WHISKEY/EXIT`，選啤酒觸發 Tale 44，提示以寒冷攻擊
對付紅袍法師偏好的火焰生物。兩則已由 Adventure Journal 正文補成繁中，
EXIT 可返回艾森布拉六場所選單。另以攻略交叉確認灰袍人／手札 18 屬暗影谷
酒館，清除「DAX 相鄰字串即同城市分支」的錯誤推論。

先前成果：哈普 ECL5 block `0x31` terrain `0x88` 的伊弗利特頭目主線已接通。
新支援的 `APPROACH (0x0D)` 以 renderer-neutral count 保存兩次靠近動畫；
State 不再吞掉 ECL runner error。正式戰鬥包含 MON5 `0x34` 伊弗利特一名、
`0x32` 法師六名與 `0x33` 牧師六名。勝利後寫入 `4C01=5`、`4C5E=1`，
取得村莊／洞穴地圖，播放 PICTURE 50 解放群眾，並由長老指出附近法師塔。
成果收進 spec 309 及 Gold Box ECL／state 知識庫。

先前成果：哈普 ECL5 block `0x31` 的 terrain `0x80` 黑暗精靈巡邏已接通。
依公開 reference 修正 opcode `0x29 ENCOUNTER MENU`：五個 script values 是
behavior modes，COMBAT 對 modes 0/1/3/4 應解析為 destination `1`，不能直接
複製 mode。seed 3 的正式流程建立三名 MON5 `0x31` 戰士與一名 `0x32` 法師，
套用原版 icon 與繁中名稱；勝利後 `4C47=1` 並返回哈普 dungeon，而不是村外
wilderness menu。規格與共用引擎結論收進 spec 308 與 ECL 指令集知識庫。

先前成果：正式流程已從哈普村外進入 Area 5／ECL5 block `0x31`。入口載入
map pieces `12,FF,FF`，顯示 640×480／24px 繁中荒村敘事後進入探索模式；
terrain `0x84` 已接通 HEAD／PICTURE 50 的躲藏村民事件，可選離開或繼續
交談，最後返回同一 resumable runtime。visited `4C02=1`、terrain dispatch
與未知 `4BC9 > 14` gate 已收進 spec 307 及 Gold Box state 知識庫。

先前成果：正式世界流程已由立石群延伸到艾森布拉與哈普。原始複合選項
`PATROL FOREST` 已按單一 selection index 正規化；Standing Stone → Essembra
會寫入 `4C9B=8`，Essembra → Hap 途中建立三隻 MON1 `0x35` 黑龍，勝利後恢復
同一份 ECL session，以 `4C9B=9` 抵達哈普村外。規格與跨作品注意事項收進
spec 306 與 Gold Box state 知識庫。畫面契約維持 640×480、原圖整數
nearest-neighbour、24px 繁中（緊湊欄位約 16×15）。

先前成果：ECL2 block 4 terrain `0x87` 火刀首領主線已反組並接入 State。原始
PICTURE 12／手札 11 後建立 20 名 record 1 與 1 名 record 3；COMBAT 前的
treasure packet 現依 monster spawns 判成遭遇獎勵，只有勝利後才解析為
17,000 GP、8 gems、4 jewelry 與兩件隨機物品。戰後手札 54／53、PICTURE
14／13、四段主人夢境與 `4CFF → 4C2A → 7F12` 時序已收進 spec 303 與共用
Gold Box state knowledge。畫面契約維持 640×480、原圖整數 nearest-neighbour、
24px 繁中（緊湊欄位 16×15）。

後續已延伸至阿沙本福德：修正 encounter loot 的 ECL continuation ownership、
略過財寶後清除剩餘物，避免在下一場戰鬥重現；戰後 State 可走完 BIGPIC 121、
Tilverton 禁止入城、JOURNEY ON→ASHABENFORD→TRAIL、提爾隘口八隻鷹馬戰，
最後以 `4C83=1`、`4C9B=2`、`4CA1=2` 落在 LocationAshabenford。圖片關閉時
仍會保留原始 Continue menu，不會截斷多段 ECL 劇情。

正式流程再延伸至立石群：Ashabenford ENTER CITY 顯示 PICTURE 80 與
INN/STORE/HALL/TEMPLE/BAR/LEAVE；HALL 的 PROGRAM 0 依 choice context 進入
訓練服務。河畔酒館已繁中化 Tavern Tale 28。離城選 Standing Stone TRAIL
會遭六名 MON1 `0x59`／icon `0x20` 偽裝火刀伏擊；勝利後灰袍人指出尚有四位
主人，THANK HIM 顯示「往南方尋找紅色之人」，並以 `4C9B=4` 同步
LocationStandingStone。

## 目前輪次

第一百五十九輪：State LOAD PIECES request。

第 1 輪 commit：`d87b8c3`，已推送至 GitHub `main`。
第 2 輪 commit：`f46bb3d`，已推送至 GitHub `main`。
第 3 輪 commit：`b2631a9`，已推送至 GitHub `main`。
第 4 輪 commit：`beb9de1`，已推送至 GitHub `main`。
第 5 輪 commit：`9162462`，已推送至 GitHub `main`。
第 6 輪 commit：`26b889c`，已推送至 GitHub `main`。
第 7 輪 commit：`002acab`，已推送至 GitHub `main`。
第 8 輪 commit：`c079c50`，已推送至 GitHub `main`。
第 9 輪 commit：`ec85c89`，已推送至 GitHub `main`。
第 10 輪 commits：`d1aea12`、`a4df55b`，已推送至 GitHub `main`。
第 11–12 輪 commits：`a798531`、`07c9309`、`697ff34`，已推送至 GitHub `main`。
第 13 輪 commits：`a6529c8`、`2489507`，已推送至 GitHub `main`。
第 14 輪 commits：`77375cd`、`b3aa461`，已推送至 GitHub `main`。
第 15 輪 commit：`2d50510`，已推送至 GitHub `main`。
第 16 輪 commit：`499e3d3`，已推送至 GitHub `main`。
第 17 輪 commit：`1e79695`，已推送至 GitHub `main`。
第 18 輪 commit：`e6f9fd2`，已推送至 GitHub `main`。
第 19 輪 commit：`624c9e1`，已推送至 GitHub `main`。
第 20 輪 commit：`e36d1ef`，已推送至 GitHub `main`。
第 21 輪 commit：`9d7087f`，已推送至 GitHub `main`。
第 22 輪 commit：`2b5ff95`，已推送至 GitHub `main`。
第 23 輪 commit：`1f6849e`，已推送至 GitHub `main`。
第 24 輪 commit：`1730d11`，已推送至 GitHub `main`。
第 25 輪 commit：`9cc0db0`，已推送至 GitHub `main`。
第 26 輪 commit：`357b7da`，已推送至 GitHub `main`。
第 27 輪 commit：`ad4e98d`，已推送至 GitHub `main`。
第 28 輪 commit：`80f9c46`，已推送至 GitHub `main`。
第 29 輪 commit：`2c03327`，已推送至 GitHub `main`。
第 30 輪 commit：`1edfc84`，已推送至 GitHub `main`。
第 31 輪 commit：`dd3dee6`，已推送至 GitHub `main`。
第 32 輪 commit：`01372f1`，已推送至 GitHub `main`。
第 33 輪 commit：`84df0f9`，已推送至 GitHub `main`。
第 34 輪 commit：`047e63f`，已推送至 GitHub `main`。
第 35 輪 commit：`32aa9c8`，已推送至 GitHub `main`。
第 36 輪 commit：`4b2edd3`，已推送至 GitHub `main`。
第 37 輪 commit：`3d25b13`，已推送至 GitHub `main`。
第 38 輪 commit：`4adc4d5`，已推送至 GitHub `main`。
第 39 輪 commit：`251e087`，已推送至 GitHub `main`。
第 40 輪 commit：`f18a058`，已推送至 GitHub `main`。
第 41 輪 commit：`00684e8`，已推送至 GitHub `main`。
第 42 輪 commit：`230d3f7`，已推送至 GitHub `main`。
第 43 輪 commit：`b756398`，已推送至 GitHub `main`。
第 44 輪 commit：`f4b1d32`，已推送至 GitHub `main`。
第 45 輪 commit：`d7eb4f6`，已推送至 GitHub `main`。
第 46 輪 commit：`d780fda`，已推送至 GitHub `main`。
第 47 輪 commit：`d2a1f3d`，已推送至 GitHub `main`。
第 48 輪功能 commit：`cfa0fce`，已推送至 GitHub `main`；本行為後續文件同步提交。
第 49 輪功能 commit：`6464ea6`，已推送至 GitHub `main`；本行為後續文件同步提交。
第 50 輪功能 commit：`e0b7030`，已推送至 GitHub `main`；本行為後續文件同步提交。
第 51 輪功能 commit：`3e679a9`，已推送至 GitHub `main`；本行為後續文件同步提交。
第 52 輪功能 commit：`4574181`，已推送至 GitHub `main`；本行為後續文件同步提交。
第 53 輪功能 commit：`4d8b91e`，已推送至 GitHub `main`；本行為後續文件同步提交。
第 54 輪功能 commit：`46e7fda`，已推送至 GitHub `main`；本行為後續文件同步提交。
第 55 輪功能 commit：`bb66b25`，已推送至 GitHub `main`；本行為後續文件同步提交。
第 56 輪功能 commit：`a2ba9ad`，已推送至 GitHub `main`；本行為後續文件同步提交。
第 57 輪功能 commit：`a983eca`，已推送至 GitHub `main`；本行為後續文件同步提交。
第 58 輪功能 commit：`0a56423`，已推送至 GitHub `main`；本行為後續文件同步提交。
第 59 輪功能 commit：`595c6ce`，已推送至 GitHub `main`；本行為後續文件同步提交。
第 60 輪功能 commit：`5d13970`，已推送至 GitHub `main`；本行為後續文件同步提交。
第 61 輪功能 commit：`7a085aa`，已推送至 GitHub `main`；本行為後續文件同步提交。

第 62 輪功能／文件 commit：`f742f7a`，已推送至 GitHub `main`；本行為後續文件同步提交。

第 63 輪功能／文件 commit：`af7bf19`，已推送至 GitHub `main`；本行為後續文件同步提交。

第 64 輪功能／文件 commit：`8e4d433`，已推送至 GitHub `main`；本行為後續文件同步提交。

第 65 輪功能／文件 commit：`f9d65c1`，已推送至 GitHub `main`；本行為後續文件同步提交。

第 66 輪功能／文件 commit：`ae4e32a`，已推送至 GitHub `main`；本行為後續文件同步提交。

第 67 輪功能／文件 commit：`8144392`，已推送至 GitHub `main`；本行為後續文件同步提交。

第 68 輪功能／文件 commit：`18d7aca`，已推送至 GitHub `main`；本行為後續文件同步提交。

第 69 輪功能／文件 commit：`a116bee`，已推送至 GitHub `main`；本行為後續文件同步提交。

第 70 輪功能／文件 commit：`bb428df`，已推送至 GitHub `main`；本行為後續文件同步提交。

第 71 輪功能／文件 commit：`85196f2`，已推送至 GitHub `main`；本行為後續文件同步提交。

第 72 輪功能／文件 commit：`06d5447`，已推送至 GitHub `main`；本行為後續文件同步提交。

第 73 輪功能／文件 commit：`1049135`，已推送至 GitHub `main`；本行為後續文件同步提交。

第 74 輪功能／文件 commit：`5f52fff`，已推送至 GitHub `main`；本行為後續文件同步提交。

第 75 輪功能／文件 commit：`e720de7`，已推送至 GitHub `main`；本行為後續文件同步提交。

第 76 輪功能／文件 commit：`3e11803`，已推送至 GitHub `main`。已加入 `gfx.MergePictures` 的透明／bitwise-OR 合成規則；從 `CHEAD.DAX`／`CBODY.DAX` 產生六張 `party-block-XX.png`，Ebiten party fighter 優先顯示合成小人；新增 `docs/knowledge/gold-box-graphics.md` 與第 76 輪規格。測試與素材重建通過。

第 77 輪功能／文件 commit：`419cd9d`，已推送至 GitHub `main`。`combat.Fighter` 現在保存 party icon head/body、normal／attack 與 direction state；`gfx.Picture.FlipHorizontal` 與 Ebiten renderer 已接入方向 flip；generator 產生 normal／attack CHEAD＋CBODY 合成圖，並補上第 77 輪規格與知識庫更新。`go test ./...` 通過。

第 78 輪功能／文件 commit：`0a5f96c`，已推送至 GitHub `main`。依 reference 反組譯確認新建角色 `head_icon=0`、`weapon_icon=0`，並將 dwarf／gnome／halfling 的 small `icon_size=1`、其餘種族 normal `icon_size=2` 接入 `party.Character`、fighter projection 與舊 JSON 相容路徑；移除 party-slot 假造外觀的規則。`go test ./...` 與素材重建通過。

第 79 輪功能／文件 commit：`381c13b`，已推送至 GitHub `main`。SPRIT frame 的 `x/y` 已寫入 animation manifest 並由 Ebiten `combatAnimation` 套用至實際 draw origin；更新 SPRIT 規格、共用 Gold Box 圖像知識庫與 README，清除 frame position 尚未實作的過時斷言。`go test ./...` 與素材重建通過。

第 80 輪功能／文件 commit：`aadd3fe`，已推送至 GitHub `main`。依 reference `load_pic_final` 實作 PIC／FINAL 後續 frame 對第一幀 packed bytes 的 XOR decoder；SPRIT 維持 full-frame mode。PIC1–PIC6 已抽出 152 張 PNG，新增 parser regression、規格與知識庫內容。`go test ./...` 與素材重建通過。

第 81 輪功能／文件 commit：`f2640a6`，已推送至 GitHub `main`。ECL `PICTURE` opcode 現在產生 renderer-neutral block request；game state 進入可恢復 `ModeEvent`，Ebiten 依 `picN-block-XX` 播放 PIC frames，Enter 返回原流程；新增 ECL／game regression 與第 81 輪規格。`go test ./...` 通過。

第 82 輪功能／文件 commit：`2d004e5`，已推送至 GitHub `main`。依 reference `CMD_Picture` 將 block `>= 0x78` 分流為 BIGPIC；加入 `BIGPIC1/2/6.DAX` unmasked picture extraction、4 張 bigpic PNG、renderer-neutral `BigPictureRequested` 與 Ebiten 置中大圖事件畫面。`go test ./...` 與素材重建通過。

第 83 輪功能／文件 commit：`325df5a`，已推送至 GitHub `main`。依 reference `HEAD<area>`／`BODY<area>` 與 body `y+5` 規則，抽出 HEAD2–6／BODY2–6 共 40／31 張 PNG，合成 30 張 scene character composites；新增 `gfx.MergePicturesAt` offset API、regression、共用 Ebiten loader 與第 83 輪規格。`go test ./...` 與素材重建通過。

第 84 輪功能／文件 commit：`7092643`，已推送至 GitHub `main`。依 reference `Area2.HeadBlockId == 0xFF` 分流 PICTURE：非 sentinel 時 body 使用 PICTURE block、head 使用 scene state，Ebiten 顯示 HEAD/BODY composite；新增 `SetSceneCharacter`、可恢復事件 regression、第 84 輪規格與知識庫更新。`go test ./...` 通過。

第 85 輪功能／文件 commit：`fcb63d7`，已推送至 GitHub `main`。確認並接入 reference `Area2.HeadBlockId @ 0x5C2`：`area.State`／Area2 codec 讀寫該欄位，`game.SetAreaState` 同步到 PICTURE HEAD/BODY branch；新增 raw codec 與 Area2→game regression、第 85 輪規格。`go test ./...` 通過。

第 86 輪功能／文件 commit：`4a60738`，已推送至 GitHub `main`。依 reference `MapDirectionDelta` 建立八方向 `combat.DirectionDelta`，新增 deterministic `FormationTile`，Ebiten 戰鬥 party／enemy sprite、姓名與 HP 改由 tile-derived 座標繪製；補 placement regression、第 86 輪規格與知識庫更新。`go test ./...` 通過。

第 87 輪功能／文件 commit：`162b0e6`，已推送至 GitHub `main`。依 reference `CombatantMap {pos,size,screenPos}` 將 `HasCombatPosition`／`CombatX`／`CombatY`／`CombatSize` 接入 Fighter、StartCombat 與 Ebiten；外部真實 position 優先，缺少時才用 formation fallback。新增 position regression、第 87 輪規格與知識庫更新。`go test ./...` 通過。

第 88 輪功能／文件 commit：`d73436b`，已推送至 GitHub `main`。反組譯 `ovr011.try_place_combatant`，封裝 `pos.x = candidateColumn + teamX*6 + groupRow*5 + 22`、`pos.y = candidateRow + teamY*5 + 10` 為 `combat.ReferencePlacement`，加入 regression、規格與知識庫；team／occupancy inputs 尚未強行假設。`go test ./...` 通過。

## 已確認

- `curseoftheazurebonds.zip` 是 94 個檔案、約 1.2 MiB 的 DOS 遊戲映像，包含 `START.EXE`、`GAME.OVR`、六組 `ECL*.DAX`、圖像、精靈、地城與怪物資料。
- `START.EXE` 為 16-bit DOS MZ executable；`GAME.OVR` 是約 272 KiB 的 overlay／資料檔候選。
- `ECL1.DAX` 至 `ECL6.DAX` 的大小約 16–27 KiB，應是按章節分組的腳本資料；尚未宣稱其 opcode 或 header 已確定。
- DAX 容器的 2-byte header offset、9-byte block index 與 signed-byte RLE 已由第二輪工具在 ECL/GEO 樣本上驗證。
- `scripts/dax_dump.py` 已能輸出 block metadata 與 ECL 6-bit packed text 取樣；`tests/test_dax_dump.py` 覆蓋正常 RLE 與截斷錯誤。
- `START.EXE` 的 MZ header 與 `GAME.OVR` 的 `TPOV` 前綴已完成位元組級盤點；loader／overlay 的真正載入邊界仍未宣稱確定。
- `internal/dax` 已實作已確認的 DAX container parser；`internal/locale` 與 `assets/locale/zh-TW.json` 建立第一版繁中資源層。
- Go 1.22 容器驗證通過：`go test ./...`，CLI 能解析 `ECL1.DAX` 的 3 blocks。
- `internal/ecl.ParseOperands` 已實作並測試 ECL 的局部 cursor／word framing；尚未執行 opcode。
- `internal/ecl.Trace` 已加入 command metadata 與安全停止行為；可追蹤已知命令，但尚未實作分支或副作用。
- `internal/ecl.DecodePackedText`／`FindPackedTextCandidates` 已能抽取 ECL 英文候選；真實 `YOU ARE AT THE EDGE OF` payload regression test 通過。
- `internal/game.State` 已將 `zh-TW` catalog 接入 Title → Wilderness → Event 的 opening flow，並有狀態轉移測試。
- `cmd/azure-bonds-game` 已接入 Ebiten、鍵盤輸入與外部 TTF/OTF 字型；仍是 prototype，未宣稱完整遊戲。
- Ebiten opening 已改為讀取 `ECL1.DAX` 第一個 block；`game.State.OriginalOpening` 會記錄從原始 payload 辨識出的 opening marker。
- 第十輪驗證通過：internal tests 與 `cmd/azure-bonds-game` Ebiten compile；GUI 仍需有 display 才能實跑。
- `internal/ecl.TraceGraph` 已能追蹤 code-segment `GOTO/GOSUB` targets；不執行條件或副作用。
- 參考公開 CoAB 重寫程式後，確認 ECL 初始化會先連續讀五組 word-valued command-set；`internal/ecl.EntryPoints` 已加入此安全解析器與 regression test。
- 實際 `ECL1.DAX` 三個 block 的 initial entry 都解析為 `0x8014`；CLI `-graph` 已優先使用該入口對應的 payload offset `+0x0014`，不再盲目從 `+0x0000` 開始。
- `ParseOperands` 已正確消耗 `0x80 length payload` compressed-string operand；`TraceAt` 可從 initial entry 開始，block 81 已回歸解出 `AS YOU DEPART...`、`AS YOU LEAVE...` 等原始事件文字。
- `internal/ecl.RunSubset` 已支援 bounded `SAVE/COMPARE/IF/GOTO/GOSUB/RETURN/PRINT/ON GOTO/ON GOSUB`；實際 ECL1 仍會在其他尚未接入的副作用 command 安全停止。
- `RunSubset` 已加入 `0x14 COMPARE AND`、`0x2A GETTABLE`、`0x2B HORIZONTAL MENU`；實際 ECL1 已讀出 TILVERTON／SHADOWDALE 開場與 `ENTER CITY/JOURNEY ON/CAMP` menu。
- `game.NewStateFromECL` 與 Ebiten opening 已接上原始 menu 的繁中 locale 映射；未提供 sequence 時 runner 仍以 deterministic index 0。
- `RunSubsetWithSelections`、`game.State.Select` 與 Ebiten cursor 已接上 menu index；實際 selection 1 會走到不同 ECL branch。
- `0x15 VERTICAL MENU` 已加入 prompt／options／selection parser；實際 ECL1 可讀到 `INN/STORE/BAR/LEAVE`。
- `RunSubsetInteractive` 與 `State.selectionSequence` 會在 sequence 用完時停在下一個 menu；繁中已加入客棧／商店／酒館／離開／繼續提示。
- `-interactive` CLI 已加入；實際 ECL1 sequence `0,0,1` 會停在 `SHADOWDALE/ASHABENFORD/DAGGER FALLS`，三個地點已接入繁中 locale。
- 實際 sequence `0,0,1,0` 會停在 Shadowdale 的 `WILDERNESS/EXIT` menu；`game.State.Location` 已轉為 `LocationShadowdale` 並保留原文地點。
- `LocationName` 已接入 locale／Ebiten UI；Shadowdale 地點狀態可見，WILDERNESS 回野外、EXIT 後續事件仍待實作。
- `0x20 NEWECL` 已加入 `RunResult.NewECLBlockID` signal；第 55 輪已將 ECL1–ECL6 合併為 global block namespace。
- `BlockSession` 已封裝 decoded block ID／payload、initial entry、validated switch；`ECL1.DAX -session` 已驗證 `0x50/0x51/0x52`。
- `cmd/azure-bonds-game` 已載入 ECL1–ECL6 全部 blocks；`NewStateFromECLBlocks`／`State.Select` 使用 global BlockSession 與 selection offset，能為後續 NEWECL 接續保留 bounded session。
- `BlockSession.RunInteractive` 已實際接入 State，會依 `SelectionsConsumed` 傳遞 global sequence 並套用 NEWECL target；完整玩家流程仍未抵達所有 real entry。
- `-all-entries -graph` 已掃描 ECL1–ECL6；ECL4 block `0x25` payload `+0x022B` 的 real `NEWECL → 0x50` 已由 `-block-id 37 -run-start 555` 驗證。
- ECL5 block `0x30` payload `+0x0098` 的第二條 real `NEWECL → 0x50` 已由 `-block-id 48 -run-start 152` 驗證；ECL1／ECL4／ECL5 合併 session 也已真實跳到 ECL1 block `0x50` 的 Tilverton opening menu。
- Shadowdale `WILDERNESS/EXIT` 已接成第一個 `ModeMap` 垂直切片：State 保存 `(MapX, MapY)`，Ebiten 方向鍵可移動、Esc 可返回；目前沒有宣稱原始 tile 或場所資料已解碼。
- 原始 ECL1 block `0x51` 已觀察到 `INN/STORE/BAR/LEAVE`；本輪接入 `ModePlace`、繁中選項與事件回復，但尚未宣稱場所內部 command path 已完成。
- `0x27 TREASURE` 已依公開 command table 解碼 7 種 pooled money（Copper／Silver／Electrum／Gold／Platinum／Gems／Jewelry）與第 8 個 `ITEM{area}.DAX` block operand；`RunResult`／`BlockSession`／State 已接 raw signal 與 exactly-once queue，尚未宣稱已完成 ITEM DAX 抽物、random branch 或 inventory mutation。
- `ITEM1.DAX`～`ITEM6.DAX` 已由啟動器載入並交給 `game.ParseTreasureItemBlocks`；State 可解析 deterministic block、累加 reference coin conversion、保留 Gems／Jewelry treasure pool，並透過明確 `TakeTreasureItem` 寫入 party equipment。random item branch、完整 loot UI 與劇情入口仍保留 boundary。
- `TREASURE` `0x80+n` random item branch 已依 reference d100 table 使用 seeded resolver 產生 n 件 item；State／Ebiten 已提供繁中 loot menu，玩家可選物品與收下角色。完整 name-number／identify、capacity 與所有真實 loot 劇情入口仍保留 boundary。
- 同一 ECL result 的 TREASURE／COMBAT 順序已接回：先保留 loot 並進入戰鬥，party victory 後先跑 resumable ECL continuation，再顯示 loot menu；headless 未載入 ITEM DAX 時仍不阻斷 COMBAT regression。
- reference `ClassId 8..16` multi-class player records 已可由 DOS parser 接受；`ClassLevel[8] @ 0x109`、`multiclassLevel @ 0xE6` 保存到 Character／JSON 並可透過 player writeback 回寫。primary-class combat projection、multi-class rules／creation UI 仍保留 boundary。
- `ItemRecord.NameNumbers` 現依 reference `GenerateName` 與 `HiddenNameFlags` bits 組合已確認的繁中 magic components；unknown name numbers 仍 raw-preserving，完整 itemNames／Identify side effects 尚待逐項驗證。
- 角色建立已依 reference `Gbl.RaceClasses` 擴充至 40 個單／多職業選項；18 個 multi-class 會保存 RawClassID／ClassLevels，並以 primary class 接目前 party／combat。完整 multi-class rules、alignment、training 與專用副作用仍保留 boundary。
- 真實 ECL1 block `0x51` 的 `+0x0643 COMBAT` 已轉為 `RunResult.CombatRequested`，並保留 `SETUP MONSTER`／`LOAD MONSTER` descriptors；`State.StartEncounter` 可用 `MON*CHA` records 建立 Battle，完整玩家流程仍未完成。
- `internal/combat.Battle.Attack` 與 `internal/game.State.StartCombat/CombatAct` 已接成可由 seed 重現的玩家／敵人回合切片；Ebiten 已有繁中 HP、目標與攻擊畫面，但完整 AD&D 戰鬥仍未完成。
- ECL1 block `0x51 +0x1293` 的 `SETUP MONSTER`／三個 `LOAD MONSTER` 已由 descriptor decoder 驗證，接續 `+0x12B0 COMBAT`；尚未解碼 `MON*CHA` stats。
- `MON1CHA` fixed `0x1A6` record parser 已接入；實際 block `0x56` 解出 BUGBEAR `24/24 HP`、raw AC `55`、attack bonus `44`、`2d4`、initiative `9`，可轉成 `combat.Fighter`。
- `MON*ITM` `0x3F`-byte parser 與 `MON*SPC` 9-byte affect parser 已接入；MON1 block `0x59` 有 5 筆 item、block `0x35` 有 2 筆 affect，名稱組合與 effects merge 尚未完成。
- 已為實際觀察到的 item／effect IDs 接入繁中名稱：弩矢、輕弩、闊劍、盾牌、鏈甲、偵測隱形、酸液吐息；未知 IDs 仍明確 fallback。
- `BuildEnemies` 已將 ECL1 block `0x51 +0x1293` 的 3 個 spawn 與 `MON1CHA` 合併，真實輸出 24 個 enemies（FIGHTER×4、BUGBEAR×10、WORG×10），在第一個 COMBAT 邊界停止；`RunResult`／`State.StartEncounter` 已可傳入 Battle。
- `cmd/azure-bonds-game -encounter` 已直接執行 ECL1 block `0x51 +0x1293`，讀取 `MON1CHA.DAX` 並進入戰鬥；此入口使用明確標示的 debug party，正常 opening 尚未自動抵達 encounter。
- `PROGRAM 0/3/8/9` 已依參考重寫程式接成外部 routine boundary 與共用 State adapter；真實 CAMP selection 在 `PROGRAM 9` 停止，不再錯誤重複跑場所 menu，0/3/8 則處理返回標題、全滅與勝利恢復／存檔選擇。
- 已加入 `docs/manual/` 繁中遊玩手冊、`docs/history.md` 中文金盒子歷史筆記，以及遊戲內 `J`／`Esc` 冒險手札；`State.Camp` 接收 `PROGRAM 9` 並開啟 CAMP Menu，REST 的自然恢復已接入窄 boundary，完整時鐘／中斷規則仍未完成。
- 冒險手札已擴充為八頁 locale-backed 摘要，State 保存頁碼並支援方向鍵翻頁；完整 59 個 Journal Entry／Tavern Tale 逐條觸發仍未完成。
- `State.SetParty`／`PartyFighters` 已保存 party roster；戰鬥結束同步 HP，CAMP 對已保存 party 恢復 MaxHP，完整角色欄位與原始 CAMP side effects 仍未完成。
- `internal/party` 已依 RuleBook／reference 建立六種玩家種族、六種基本職業、能力值 3–18、最低值與 1–6 人 roster validation。
- Ebiten 已接 `C` 角色建立 starter UI：方向鍵選模板、Enter 加入、D 完成；三個模板會經 validation、投影成 combat fighter 並保存到 State。
- 建立畫面按 `N` 可輸入 Unicode／繁中姓名，按 `A` 可編輯六項能力值，按 `R` 可重擲六項 3d6；均在完成時經 validation。
- `internal/save` 已建立 version 1 的 remake party JSON；F5／F9 與 `-party-load` 可保存／載入角色描述，原版 DOS save/import、裝備、XP／等級與完整 party state 仍未完成。
- ECL `COMBAT` 現在在 State 已有 party 與 `MON1CHA` records 時，會由真實 spawn descriptors 建立可操作 Battle；缺少資料時仍安全停在事件 boundary。
- ECL `ENCOUNTER MENU` opcode `0x29` 已解析 14 operands，支援 `COMBAT／WAIT／FLEE／ADVANCE／PARLAY` action mapping、interactive pause 與繁中顯示；實際接近距離、外部遭遇 routine 與戰鬥規則仍未完成。
- ECL interactive menu pause 現在保存 PC、numeric／string memory、比較旗標與 call stack；`BlockSession` 保存共享 runtime context 與 cumulative selection offset，可從同一 menu resume。
- `NEWECL` 現在將 bounded runtime context 帶到 target block initial entry；synthetic regression 已確認 source memory 可被 target 讀取，完整原始 block continuation 與所有 DOS memory semantics 仍未完成。
- 真實 ECL1 block `0x51` 的 `JOURNEY ON → STORE` 已建立 regression；第 284 輪
  進一步證明該 `COMBAT` opcode 依 EnterShop 分派 CityShop，並非缺 monster
  descriptor 的 encounter。
- ECL `RANDOM` opcode `0x08` 已加入 seeded bounded runner；`RunResult` 保存 random values，State 可用 `SetECLSeed` 重播事件，完整遭遇表與其外部 routine 仍未完成。
- 建立畫面按 `R` 可重擲六項 3d6 能力值；核心接受 seed regression，完成加入時仍經職業最低值 validation。
- 原始映像與 PDF／RAR 手冊是本地研究素材，第一輪不直接納入 Git 追蹤。

## 尚未確定

- DAX 的 header、壓縮方式、索引與圖像／腳本子格式。
- `GAME.OVR` 與 `START.EXE` 的載入關係。
- ECL opcode、字串編碼、分支／呼叫慣例。
- unknown opcode `0x85` 的完整語意與 IF／menu 的 runtime state 仍未完成。
- TREASURE 的 party inventory／獎勵規則仍未完成；目前僅有安全 operand prefix。
- COMBAT signal、party／enemy model、基本回合／骰點／傷害與戰鬥 UI 垂直切片已完成；法術、物品、逃跑／PARLAY、戰場與完整原版流程仍未完成。
- party／enemy、initiative、攻擊與傷害 core 及基本 UI、ECL encounter direct-entry 已完成；完整玩家流程、戰場、法術、物品、逃跑／PARLAY 仍未完成。
- ECL monster spawn descriptor、`MON*CHA` HP／AC／攻擊資料與 ECL-to-combat direct-entry adapter 已完成；完整玩家流程接線仍未完成。
- `MON*CHA` raw HP／AC／攻擊 parser 與 Fighter adapter 已完成；`MON*ITM`／`MON*SPC`、完整 ECL-to-Battle setup 仍未完成。
- `MON*CHA`、`MON*ITM`、`MON*SPC` raw parser 已完成；item name catalog、effects merge 與完整 ECL-to-Battle setup 仍未完成。
- 已觀察 IDs 的 monster item/effect 繁中顯示完成；完整 item type／name-number catalog、locale resource integration、effects merge 與 ECL-to-Battle setup 仍未完成。
- 已觀察 IDs 的 monster item/effect 繁中顯示、enemy adapter、party／Battle direct-entry 與基本 UI 完成；完整 item catalog、effects merge、原始 party roster、戰場與玩家流程仍未完成。
- CAMP／其他城市場所功能、完整 menu rendering／input semantics 與後續事件仍未完成；Shadowdale 首層場所 menu 已有 state contract。
- 地點選定後的 map／place state、CAMP 與後續事件仍未完成。
- Shadowdale 已有資料中立的座標／輸入 contract；原始 tile、碰撞與場所事件仍未完成。
- 全 ECL global block loader 與 bounded memory／call stack transfer 已完成；GEO 16×16 geometry parser 已完成，原始 tile art／碰撞與完整場所功能仍未完成。
- `internal/gfx` 已解析原始 `TILES.DAX` 24×24 indexed pictures、`8X8D2–6.DAX` 8×8 symbol pictures 與可串接的 `WALLDEF2–6.DAX` records；palette mapping、畫面組合與碰撞仍未完成。
- `internal/gfx.Picture.RGBA` 已接 reference EGA16 palette 與 mask index 16 的透明語意；尚未接入 Ebiten map viewport。
- `cmd/azure-bonds-game` 已載入 `TILES.DAX` 兩個 block 並以 `T` 顯示 48 個原始 24×24 tile 的 Ebiten gallery；GEO tile index mapping、完整 map viewport 與碰撞仍未完成。
- `cmd/azure-bonds-game` 已以 `G` 顯示 `GEO2` 第一個 block 的 16×16 raw wall geometry viewport；background floor construction、tile mapping、碰撞與 camera 仍未完成。
- `geo.Grid.CanMove` 已依 raw wall fields 建立 shared 四方向 traversal contract；viewport 黃點遵守目前／相鄰 cell wall 與邊界，完整 movement cost、encounter 與 floor collision 仍未完成。
- `internal/mapdata` 已保存 reference `BackGroundTiles` 實際 74 筆 metadata，保留 0xFF impassable sentinel 與 reserved tail；尚未宣稱完成 floor／tile mapping。
- 根目錄 `README.md` 已加入目前成果與限制，`docs/screenshots/tiles-gallery.png`、`geo-geometry.png` 由 `scripts/render_previews.go` 從原始 DAX 以現有 parser 可重現產生。
- README 截圖證明 TILES indexed graphics 與 GEO raw wall geometry 管線已初步完成；不代表完整地圖、完整劇情或完整遊戲已完成。
- `mapdata.GenerateWilderness` 已依 reference `SetupWildernessFloor01–03` 建立 50×25 floor，保留 city flags、骰點與 background entry → pixel tile 的資料邊界。
- `game.State` 進入 Shadowdale map slice 時生成 seeded wilderness floor；`Move` 已檢查 map boundary 與 `BackgroundTile.MoveCost` impassable sentinel。
- `scripts/render_previews.go` 新增 `docs/screenshots/wilderness-floor.png`，由原始 TILES 與同一 floor generator 組合產生。
- `mapdata.GenerateDungeon` 已依 reference `SetupDungeonFloor` 的四段 builder，將 GEO wall／door detail 組成 13×5 dungeon background-entry window。
- `cmd/azure-bonds-game` 新增 `D` dungeon floor preview；`docs/screenshots/dungeon-floor.png` 使用同一組 GEO、mapdata 與 TILES pipeline 產生。
- dungeon floor 的桌椅 decoration 已完成 reference `sub_370D3` 的 seeded pass；area selection、camera、combat placement、encounter 與音效仍未完成。
- `mapdata.GenerateDungeonSeeded` 已讀取 GEO `terrain & 0x40`（reference `MapInfo.x2`），依 `sub_370D3` flags／tile index／1d10 dice 寫入 table entry `0x1A` 與 chair entry `0x1B`；seeded regression 通過。
- dungeon preview／PNG 仍使用同一 decoration-enabled generator；完整 area selection、camera、combat placement、encounter 與音效仍未完成。
- `geo.Catalog` 已保留 GEO2–GEO6 全部 16 個原始 DAX block IDs，並以 `MapRef{Set,BlockID}` 查找；原始 block shape regression 通過。
- `cmd/azure-bonds-game` 新增 `-geo-set`／`-geo-block`，選中的 GEO block 會同時驅動 G geometry 與 D dungeon floor preview；area pointer 自動選圖尚未完成。
- ECL opcode `0x21 LOAD FILES` 已解碼三個 operand；有效第三值會經 `game.State` pending request，由 Ebiten 從 `geo.Catalog` 自動切換 GEO geometry／dungeon floor。
- 完整 `Area1.inDungeon`／`game_area`／save loader、WALLDEF reload 與非 dungeon picture side effect 尚未完成。
- `internal/area.State` 已建立 reference `current_3DMap_block_id`／`inDungeon`／`game_area` 邊界；`LOAD FILES` 現在只有在 dungeon state 才產生 GEO request，非 dungeon 分支保留 big-picture effect signal。
- `game.State` 已接入 `Area.State`，可用 `SetInDungeon(true)` 驗證 ECL map load；完整 Area1／Area2 save/import 與自動 game-area 載入仍未完成。
- party／map state 的完整 VM semantics 尚未跨 block 保存，仍不是完整 VM。
- real transition 已有 entry-level regression，但尚未由完整玩家流程抵達。
- 中文化的字型格式與字串長度限制。
- 完整 Adventurer's Journal 條目、CAMP 恢復／中斷規則與原始 party state 仍未完成。
