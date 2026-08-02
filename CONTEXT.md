# 專案現況

更新日期：2026-07-28

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

## 下一步

1. 建立可重現的映像 manifest／樣本檢查工具。
2. 對 `ECL*.DAX`、`GAME.OVR` 做十六進位與反組譯取樣，找出共通 header／索引。
3. 依證據更新 `docs/spec/`，未驗證內容維持 `DRAFT`。
4. 建立 ECL block 的欄位／事件邊界測試，才評估是否能把 ECL 規格標為 `READY`。
5. 將 ECL decoded payload 分成事件 header、控制資料與文字，建立最小場景 trace。
6. 以 operand framing 為基礎加入未知 opcode 安全停止的 trace walker。
7. 對齊 ECL branch targets 與原版場景入口，建立最小可執行事件狀態。
8. 對全部 ECL 建立去重原文 catalog，開始逐項接入繁中翻譯。
9. 建立 Ebiten input／render adapter，先視覺化 opening state。
10. 將 ECL event trace 與 Ebiten event screen 接起來，加入第一個可驗證劇情場景。
11. 建立 ECL branch target graph，將 opening marker 後的選項與事件序列接入 state。
12. 對 ECL graph 的 entry points 做原版事件文字對齊，建立第一個完整 event screen。
13. 用 `EntryPoints` 對實際 ECL1–ECL6 做入口 regression，再逐步加入可執行 VM command subset。
14. 將 `ON GOTO/GOSUB`、選單與 bounded memory model 以 regression 驗證後接入遊戲 state；完整 DOS memory model 仍保留 boundary。
15. 將 menu selection 變成 Ebiten input／runner action，完成 Enter City／Journey On 第一個事件分支。
16. 實作 `VERTICAL MENU` 的可觀測選項與 input，再擴展第一個城市事件。
17. 將 successive menu sequence 保存為 game event state，完成城市場所選擇與離開分支。
18. 建立城市場所的 ECL event regression，接入第一個可玩的場景功能。
19. 將城市選擇寫入 map state，完成第一個地點（SHADOWDALE）的場所入口。
20. 接通 Shadowdale `WILDERNESS/EXIT` 後的移動／場所 menu。
21. 建立第一個場所的 map／ECL event state。
22. 建立 ECL1–ECL6 DAX block session，驗證 NEWECL target。
23. 將 BlockSession 串入 game runtime，保存跨 block event state。
24. 建立 real NEWECL transition regression，補齊跨 block memory／call stack。
25. 掃描各 ECL event entry，找出可達 NEWECL transition 並建立實際 regression。
26. 將 ECL4 real transition 接入可導航的 event entry，補齊跨 block memory／call stack。
27. 解碼原始地圖 tile／座標規則，並接入 ECL1 block 0x51 的完整場所 command path。
28. 實作場所內部的角色、交易、休息、情報與 AD&D 規則。
29. 解碼 TREASURE table 與 party inventory 效果，補真實 event regression。
30. 建立 party／enemy combat model，接入 AD&D 回合與戰鬥 UI。
31. 將 ECL `LOAD MONSTER`／`SETUP MONSTER` 接到 combat fighter 與 battle map。
32. 解碼 `MON1CHA` 等 monster records，建立 ECL-to-combat adapter。
33. 接入 `MON*ITM`／`MON*SPC`，將 ECL spawn sequence 建立成 Battle。
34. 建立 item type／name-number catalog，合併 monster equipment/effects 並建立 Battle setup。
35. 擴充完整 item catalog，整合 locale 與 ECL-to-Battle equipment/effects。
36. 建立 party fighter 與 Battle，接入 encounter equipment/effects 與戰場 UI。
37. 將 ECL `COMBAT` result 的 spawn descriptors／MON*CHA records 自動接入 `State.StartCombat`。
38. 由完整 opening／城市／地圖流程抵達 ECL1 encounter，移除 debug party 依賴。
39. 整理完整 Adventure Journal 條目、接入 CAMP 恢復／中斷規則與 party state。
40. 將 Journal Entry／Tavern Tale 做成資料驅動的已讀條目與劇情觸發。
41. 反組 party creation/import 與完整 CAMP 恢復／中斷規則。
42. 將 `internal/party` 接入繁中角色建立 UI 與 State.SetParty。
43. 補自訂姓名／能力值、裝備、XP／等級與 party save/import。
44. 接能力值逐項分配與規則化的角色完成流程。
45. 補擲骰／重擲點數池、性別／年齡與 alignment，接角色存檔。
46. 補裝備選擇、XP／等級與 party save/import。
第 89 輪功能／文件 commit：`d389002`，已推送至 GitHub `main`。依 `ovr011.PlaceCombatants` 將遭遇距離與八方向 `mapDirection` 的 team origin／facing group 封裝為 `combat.EncounterTeamStart`，刻意未假設尚未解出的 occupancy／candidate ordering。

第 90 輪功能／文件 commit：`0151599`，已推送至 GitHub `main`。原始 `ITEMS` 解析與 128-type 繁中 catalog 已由實際 ZIP 驗證。

第 91 輪功能／文件 commit：待本輪提交。新增 `monster.ItemRecord.Effect`、packed AC／base damage adapter 與 `party.Character.FighterWithEquipment`；只套用 readied 基本武器／護甲，charges、Affects／法術、雙持、彈藥與完整 DOS inventory 仍明確保留未完成。完成 Docker 測試後提交。

第 91 輪功能／文件 commit：`ebc8715`，已推送至 GitHub `main`。ITEMS readied 武器／護甲效果已接入 party fighter，舊 JSON 行為保持相容。

第 92 輪功能／文件 commit：`fe5edbd`，已推送至 GitHub `main`。party class／slot transaction 已完成，stack、cursed 與完整 inventory mutation 待後續處理。

第 93 輪功能／文件 commit：`5cdb800`，已推送至 GitHub `main`。inventory stack mutation 與 cursed readied lock 已完成。

第 94 輪功能／文件 commit：`5834646`，已推送至 GitHub `main`。scroll／potion／wand consumable data signal 與 inventory mutation 已完成。

第 95 輪功能／文件 commit：`a604bcc`，已推送至 GitHub `main`。ECL SPELL／PROTECTION bounded signals 與 regression 已完成。

第 96 輪功能／文件 commit：`0c1a122`，已推送至 GitHub `main`。新增 `Character.SpellSlots`、`Roster.FindSpell`、`State.ResolveSpellSearch`，並由 Ebiten bootstrap 載入原始 `ITEMS`，使 creation／party load 的 readied equipment 進入 fighter projection；原始 DOS spell offsets 與 ECL memory writeback 仍未完成。

第 97 輪功能／文件 commit：`2280184`，已推送至 GitHub `main`。新增 bounded DOS player spell record parser、truncated-record guard、known／memorized tests 與可重用知識庫規格。

第 98 輪功能／文件 commit：`43b26d0`，已推送至 GitHub `main`。新增 `ParseDOSPlayerRecord`，解析公開 `.SAV/.GUY` 單職業核心欄位，並保留原始 HP、icon、金幣與 spell 到 party projection；`.SWG` inventory／`.FX` effects／多職業與完整 DOS save container 仍未完成。

第 99 輪功能／文件 commit：`938d2a7`，已推送至 GitHub `main`。新增 DOS player item/effects pointer preservation、`.SWG` `0x3F` item stream adapter 與 party equipment projection；`.FX` effects、pointer address-space 與完整 save container 仍未完成。

第一百輪功能／文件 commit：`5bcea79`，已推送至 GitHub `main`。新增 DOS `.FX` 9-byte effects stream adapter、Character preservation 與常見效果繁中名稱；effect gameplay tick／解除與完整 save container 仍未完成。

第一百零一輪功能／文件 commit：`2179dd6`，已推送至 GitHub `main`。修正 `.FX` 16-bit duration／strength 欄位語意，新增 finite/permanent duration tick 與 party adapter；effect-specific gameplay 與 CAMP／戰鬥時間接線仍未完成。

第一百零二輪功能／文件 commit：`0332a09`，已推送至 GitHub `main`。新增 `ParseDOSPlayerFiles` sidecar bundle importer，並將 gold/gems/jewelry 保存到 `Character`；`SAVGAM?.DAT` party／area container 仍未解析。

第一百零三輪功能／文件 commit：`6497c18`，已推送至 GitHub `main`。新增 `cmd/azure-bonds -import-character`，把已證實的 `.SAV/.GUY` + optional `.FX/.SWG` bundle 輸出成 versioned remake party JSON。

第一百零四輪功能／文件 commit：`d3bc6e9`，已推送至 GitHub `main`。新增 `cmd/azure-bonds-game -dos-character-record` startup bridge，直接載入單一原版角色 sidecar bundle；完整六人 SAVGAM party／area container 仍未解析。

第一百零五輪功能／文件 commit：`f57ba3e`，已推送至 GitHub `main`。將 active Bless／Curse effect 投影為 fighter attack +1／-1，其他 target/phase effects 保留未套用。

第一百零六輪功能／文件 commit：`253fdfb`，已推送至 GitHub `main`。新增 active Blind／Bestow Curse／friendly Prayer fighter projection；Haste、Protection、Mirror Image 與完整 target/action rules 仍未完成。

第一百零七輪功能／文件 commit：`c3e93f3`，已推送至 GitHub `main`。保存 DOS player `icon_id @ 0x143` 到 party／combat metadata；runtime icon slot allocation、CombatMap position／camera 與完整戰鬥流程仍未完成。

第一百零八輪功能／文件 commit：`75ae586`，已推送至 GitHub `main`。城市 `INN` 會恢復 party roster／fighter HP 並以繁中訊息返回場所選單；完整 CAMP、商店與酒館情報仍未完成。

第一百零九輪功能／文件 commit：`755d83c`，已推送至 GitHub `main`。建立 price-injected Buy／Sell／200 GP ID party transaction；完整 money pool、shop stock 與 Shop Menu UI 仍未完成。

第一百一十輪功能／文件 commit：`e8ddd14`，已推送至 GitHub `main`。城市 STORE 已接入繁中 BUY／VIEW／TAKE／POOL／SHARE／APPRAISE／EXIT menu，未知 stock／money-pool action 保留明確 boundary。

第一百一十一輪功能／文件 commit：`8ec1f86`，已推送至 GitHub `main`。接入 injected shop offers、party money pool、POOL／TAKE／SHARE 與 pool-funded BUY；實際 item selection UI、VIEW／APPRAISE 仍未完成。

第一百一十二輪功能／文件 commit：`981d01d`，已推送至 GitHub `main`。BUY 會顯示繁中商品／價格、扣除 party pool、加入 active character inventory 並返回 Shop Menu；VIEW／TAKE 數量／APPRAISE 仍未完成。

第一百一十三輪功能／文件 commit：`502b53e`，已推送至 GitHub `main`。VIEW 會列出角色 HP／金幣／繁中裝備摘要並返回 Shop Menu；完整 ALTER／ID／APPRAISE 與角色選擇 side effects 仍未完成。

第一百一十四輪功能／文件 commit：`5ef7203`，已推送至 GitHub `main`。TAKE 會選角色與 1／10／100／全部金額，更新 pool／角色 gold 並返回 Shop Menu；任意數字輸入與 APPRAISE 仍未完成。

第一百一十五輪功能／文件 commit：`6d69011`，已推送至 GitHub `main`。APPRAISE 會選角色與 gems／jewelry，接受 injected offer 後清空財寶、將 GP 加入 pool 並返回 Shop Menu；拒絕報價分支仍未完成。

第一百一十六輪功能／文件 commit：`8cdde8b`，已推送至 GitHub `main`。APPRAISE 新增接受／拒絕／返回確認；只有接受才清除財寶並入帳 pool。

第一百一十七輪功能／文件 commit：`993ced5`，已推送至 GitHub `main`。建立繁中 CAMP Menu，接入 REST return 與 EXIT 荒野返回；SAVE／VIEW／MAGIC／ALTER／FIX 保留明確 placeholder boundary。

第一百一十八輪功能／文件 commit：`b7c3f44`，已推送至 GitHub `main`。CAMP VIEW 新增角色選單、只讀繁中摘要與返回 CAMP Menu；未識別物品與 ALTER side effects 仍不猜測。

第一百一十九輪功能／文件 commit：`e4734d0`，已推送至 GitHub `main`。CAMP MAGIC 新增角色選單與已記憶 spell-slot ID 查看；當時完整 spell catalog、prepare／cast／recovery rules 尚未接入；目前已另加 bounded 一級法術名稱 catalog。

第一百二十輪功能／文件 commit：`e99977e`，已推送至 GitHub `main`。CAMP SAVE 透過一次性 state request 接到 Ebiten versioned party save；原版 SAVGAM slot／area container 尚未解析。

第一百二十一輪功能／文件 commit：`29284fb`，已推送至 GitHub `main`。CAMP ALTER／ORDER 新增兩階段角色重排，並同步 party roster 與 combat fighter；DROP／SPEED／ICON／PICS／FIX 仍未完成。

第一百二十二輪功能／文件 commit：`7c73ab3`，已推送至 GitHub `main`。CAMP ALTER／DROP 新增二次確認的永久角色移除，並同步 party roster／combat fighter；SPEED／ICON／PICS／FIX 仍未完成。

第一百二十三輪功能／文件 commit：`f235e96`，已推送至 GitHub `main`。CAMP ALTER／PICS 新增怪物圖片／動畫 runtime toggle，並接到事件／戰鬥 renderer；SPEED／ICON／FIX 仍未完成。

第一百二十四輪功能／文件 commit：`cf49211`，已推送至 GitHub `main`。CAMP ALTER／SPEED 新增 1–5 級訊息速度與 Ebiten Unicode message reveal；ICON／FIX 仍未完成。

第一百二十五輪功能／文件 commit：`8ae1d64`，已推送至 GitHub `main`。CAMP ALTER／ICON 新增已驗證 CHEAD／CBODY block 選擇，並同步 party／combat fighter icon；FIX 仍未完成。

第一百二十六輪功能／文件 commit：`2de8b01`，已推送至 GitHub `main`。CAMP FIX 依已記憶的 Cure Light Wounds（目前由一級牧師表順序映射為 ID `3`）以 deterministic `1d8` 治療受傷 roster，並同步 combat fighter HP；spell catalog、時間推進與中斷規則仍待反組譯。

第一百二十七輪功能／文件 commit：`c585d1b`，已推送至 GitHub `main`。城市 BAR 已接 ordered Tavern Tale menu、前六則繁中整理與城市場所返回；買酒價格、城市條件、完整 62 則內容與 ECL trigger 仍待反組譯。

第一百二十八輪功能／文件 commit：`c82919e`，已推送至 GitHub `main`。CAMP／PROGRAM 9 現在只開啟 CAMP Menu；REST 接入 `ADD／SUBTRACT／EXIT` 與每 24 小時自然恢復 1 HP，並同步 roster／fighter。法術記憶、遊戲時鐘與遭遇中斷仍待反組譯。

第一百二十九輪功能／文件 commit：`ce44a6d`，已推送至 GitHub `main`。CAMP MAGIC 現在將已核對的一級牧師／魔法師前八個 spell IDs 顯示為繁中名稱，未知 ID 保留 hex；完整 spell catalog、CAST／MEMORIZE／SCRIBE 與 recovery rules 仍待反組譯。

第一百三十輪功能／文件 commit：`a885b9f`，已推送至 GitHub `main`。DOS known-spell flags 現在保存到 `Character.KnownSpells`／party save，並在 CAMP MAGIC 顯示已記憶／可用數量；CAST／MEMORIZE／SCRIBE 與消耗規則仍未完成。

第一百三十一輪功能／文件 commit：`8ba67b5`，已推送至 GitHub `main`。依 RuleBook 將 CAMP MAGIC 接成 `CAST／MEMORIZE／SCRIBE／DISPLAY／REST／EXIT` command menu；DISPLAY／REST 已接入，CAST／MEMORIZE／SCRIBE 仍保留 rules boundary。

第一百三十二輪功能／文件 commit：`312f9b4`，已推送至 GitHub `main`。MEMORIZE 可從 KnownSpells 選法術，先保存 pending state，REST_START 才寫回 SpellSlots；完整 capacity、準備時間、遭遇中斷與 CAST／SCRIBE 仍未完成。

第一百三十三輪功能／文件 commit：`3b61c33`，已推送至 GitHub `main`。對已核對的一級法術套用 RuleBook 的最低準備時間檢查；不足休息時保留 pending selection，不猜測高等級與遭遇中斷結果。

第一百三十四輪功能／文件 commit：`2d7af62`，已推送至 GitHub `main`。戰鬥接入 RuleBook 證實的 Magic Missile（spell ID `7`）、slot consumption、2–5 damage per missile、level scaling 與 Ebiten S 鍵；其他 spell effects 仍未完成。

第一百三十五輪功能／文件 commit：`2a7b7c7`，已推送至 GitHub `main`。戰鬥接入 RuleBook 證實的 Cure Light Wounds（spell ID `3`）、牧師 slot、1d8 封頂治療與 Ebiten H 鍵；完整 target cursor、施法中斷與其他 spell effects 仍未完成。

第一百三十六輪功能／文件 commit：`25c0019`，已推送至 GitHub `main`。S／H 先進入施法目標選擇，左右切換、Enter 確認、Esc 取消；Magic Missile／Cure Light Wounds 分別使用敵方／我方 target list。

第一百三十七輪功能／文件 commits：`89b7620`、`a7aaca3`，已推送至 GitHub `main`。已接入 RuleBook Bless（spell ID `1`）的 B／Enter／Esc 無目標施法 transaction；成功消耗牧師 slot，隊伍攻擊加值效果與後續 adjacency／duration 收斂於下一輪。

第一百三十八輪功能／文件 commits：`64214fe`、`3090264`，已推送至 GitHub `main`。Bless 已依 RuleBook `6r` 接入 6 回合 duration，並依 CombatMap 八方向相鄰排除鄰近存活怪物的隊友；缺少位置資料時採 bounded 不排除 fallback。

第一百三十九輪功能／文件 commits：`82e2c64`、`9279fc2`，已推送至 GitHub `main`。已接入 RuleBook Curse（spell ID `2`）的 C／敵方目標選擇／Enter confirmation；未與我方八方向相鄰的敵人攻擊加值降低 1，持續 6 回合後恢復，並保留無 position direct API fallback。

第一百四十輪功能／文件 commits：`a2d968b`、`dabafa2`，已推送至 GitHub `main`。已接入 RuleBook Cause Light Wounds（spell ID `4`）的 W／敵方 touch target selection／Enter confirmation；相鄰敵人承受 deterministic 1d8 damage、無 saving throw，並保留無 position direct API fallback。

第一百四十一輪功能／文件 commits：`ee6d6d1`、`883321b`，已推送至 GitHub `main`。已接入 RuleBook Protection from Evil（spell ID `6`）的 P／party touch target／Enter confirmation；明確標記 `Evil=true` 的攻擊者對受防護目標只獲得 AC +2 門檻，duration 為 `3×caster level` 回合；saving throw、alignment import 與 dispel 保留 boundary。

第一百四十二輪功能／文件 commits：`cd8942a`、`9e6e660`，已推送至 GitHub `main`。已修正 class-local spell ID collision，牧師 ID `7` 由 G 施放 Protection from Good，魔法師 ID `7` 由 S 施放 Magic Missile；明確 `Good=true` 攻擊者才觸發 AC +2，duration 為 `3×caster level` 回合。

第一百四十三輪功能／文件 commits：`c7fac1f`、`2f078a1`，已推送至 GitHub `main`。ECL encounter 的 FLEE 進入繁中可恢復撤退事件；PARLAY 提供 HAUGHTY／SLY／MEEK／NICE／ABUSIVE 五種 tactic，選擇後返回荒野事件。怪物速度、追擊、speaker／reaction 與完整 conversation script 保留 boundary。

第一百四十四輪功能／文件 commits：`7849d19`、`4a26a36`，已推送至 GitHub `main`。戰鬥按 M 進入 MOVE，方向鍵單格移動當前 party fighter，Battle 驗證 occupancy 後更新 CombatMap 座標並消耗回合；地形、負重、進入敵格、facing 與完整離場規則仍保留 boundary。

第一百四十五輪功能／文件 commits：`1fab06e`、`b8c5616`，已推送至 GitHub `main`。MOVE 成功後若角色離開敵人鄰接範圍，Battle 對該移動者觸發存活 enemy free attack，State 顯示繁中反擊訊息並沿用勝負 transition；背面 AC／facing／reach／地形仍保留 boundary。

第一百四十六輪功能／文件 commit：`77e4245`，已推送至 GitHub `main`。MOVE 移入存活敵人格時回傳既有 attack transaction，保留 party fighter 原座標；隊友格仍拒絕，離開敵人鄰接範圍的 free attack 仍維持既有 branch。新增 RuleBook 規格、繁中 README／共用 state knowledge 與 game/combat tests。

第一百四十七輪功能／文件 commit：`960cffb`，已推送至 GitHub `main`。依 RuleBook active character centered camera，新增可重用 `CombatCamera`、State active fighter read API、Ebiten 座標轉換、測試與 graphics knowledge 更新；viewport 尺寸、scroll animation、地圖遮擋與真實 Area camera 仍保留 boundary。

第一百四十八輪功能／文件 commit：`223fc04`，已推送至 GitHub `main`。Combat Menu `VIEW` 以 `V` 開啟繁中 read-only fighter summary，Enter／Esc 關閉且不消耗回合；新增 State／renderer tests、READY 規格與共用 state knowledge。完整 View Menu、物品／交易與 Combat FLEE 的速度／追擊規則仍保留 boundary。

第一百四十九輪功能／文件 commit：`418fe09`，已推送至 GitHub `main`。依 RuleBook 與 ITEMS RateOfFire，將已裝備武器 raw `4/6` 投影為每回合 2/3 次攻擊，接入 Battle sequence、目標倒下後換下一存活敵人、繁中摘要與測試；彈藥消耗已於第 150 輪接入注入式 adapter，職業等級額外攻擊、Aim／range 與 back stab 仍保留 boundary。

第一百五十輪功能／文件 commit：`02a161a`，已推送至 GitHub `main`。保存武器 raw AmmunitionType，建立 raw code→inventory type mapping 的注入式 atomic consumption，CombatAct 在 attack 前 preflight 本回合 shots；mapping 缺失／彈藥不足不修改 inventory。新增 party／game tests、READY 規格與共用 state knowledge。

第一百五十一輪功能／文件 commit：`d4aef0b`，已推送至 GitHub `main`。Combat Menu `DONE` 以 `D` 結束目前 party turn，不攻擊、不消耗彈藥，重用 enemy／next-party advancement；新增繁中提示、測試、READY 規格與共用 state knowledge。hold／delay 與其他 Combat Menu command 仍保留 boundary。

第一百五十二輪功能／文件 commit：`900db82`，已推送至 GitHub `main`。依 RuleBook Armor List 將 armor type `50–58` 的 12／9／6 movement allowance 投影到 fighter，MOVE 每個方向鍵逐格扣除，剩餘格數不推進回合；新增 UI 提示、party／game tests、READY 規格與共用 state knowledge。負重、地形、障礙、邊界與 FLEE speed 仍保留 boundary。

第一百五十三輪功能／文件 commit：`e6971c8`，已推送至 GitHub `main`。依 RuleBook missile adjacent prohibition 與 dart thrown exception，將 ITEMS weapon group profile 投影到 fighter，Battle 在有 CombatMap 座標時拒絕 missile 近身攻擊；新增 tests、READY 規格與共用 state knowledge。完整 Range、line-of-sight、Aim cursor 與其他 thrown weapon 仍保留 boundary。

第一百五十四輪功能／文件 commit：`e0d4e31`，已推送至 GitHub `main`。修正第 153 輪 guard 與第 150 輪 ammunition atomic contract 的順序，Battle 提供不擲骰 `ValidateAttack`，CombatAct 在扣除彈藥前先拒絕無效相鄰 missile attack；新增 regression、READY 規格與共用 state knowledge。完整 ranged multi-target transaction、Range、line-of-sight 與其他 thrown weapon 仍保留 boundary。

第一百五十五輪功能／文件 commit：`07634ab`，已推送至 GitHub `main`。將已投影的 `Fighter.AttacksPerTurn` 套用到 enemy turn，讓敵方也使用 deterministic `AttackSequence` 與繁中多次攻擊摘要；新增 regression、READY 規格與共用 state knowledge。enemy AI、彈藥、Aim／line-of-sight 與額外職業攻擊仍保留 boundary。

第一百五十六輪功能／文件 commit：`531f892`，已推送至 GitHub `main`。將玩家 combat input error 接成 `ReportCombatError`／繁中訊息，尤其讓相鄰 missile／彈藥／目標錯誤留在戰鬥畫面而不結束 Ebiten game loop；新增 error sentinel、輸入攔截、regression、READY 規格與共用 state knowledge。完整 error catalog、ranged rules 與資料／啟動錯誤仍保留 boundary。

第一百五十七輪曾將 `ADD NPC (0x36)` 誤判為單 operand，使 block 0x52 的 morale
operand code `0x00` 被當成 EXIT；第 277 輪已推翻並修正。正確流程是連續加入
`0x55/0x58/0x5A`、播放完整 demo 展示序列並抵達 COMBAT；第 278 輪確認正常 new game
改走 global block `0x01`。

第一百五十八輪功能／文件 commit：`f9164e4`，已推送至 GitHub `main`。依跨 ECL 實際掃描將 `LOAD PIECES (0x37)` 接成三 selector signal，讓 ECL2 block 0x01 等實際 entry 不再因 opcode 停止；新增 synthetic／ECL1／ECL2 regression、CLI／BlockSession propagation、READY 規格與共用 state knowledge。地城 floor、wall、tile、碰撞與 camera side effect 仍保留 boundary。

第一百五十九輪功能／文件 commit：`eec71dd`，已推送至 GitHub `main`。將 `LoadPiecesRequested` 從 ECL runner 接到 game State 一次性 `ConsumeLoadPiecesRequest()`，與 GEO `LOAD FILES` request 對齊；新增 state regression、READY 規格與共用 state knowledge。當時尚未接入 map-piece file adapter；完整地城副作用仍保留 boundary。

第一百六十輪功能／文件 commit：`7ba3006`，已推送至 GitHub `main`。依公開 CoAB `LoadWalldef` reference 將 State 的 `LOAD PIECES` request 接到 `WALLDEF{area}`／`8X8D{area}` raw `PieceSet` catalog；補上單／多 WALLDEF record selector regression、原始 ZIP area 2 regression、dungeon preview 載入狀態與共用 graphics knowledge。WALLDEF row／column 的牆面拼圖、0x7F 特殊分支、碰撞與完整 3D renderer 仍保留 boundary。

第一百六十一輪功能／文件 commit：`1da9e05`，已推送至 GitHub `main`。依 reference `WallDefBlock.Offset` 將 WALLDEF graphic IDs 套用 dungeon global bases `0x2E／0x74／0xBA`，新增 `PieceSet.WallSymbol` bounded cell-to-8×8D lookup、offset regression、READY 規格與共用 graphics knowledge；九種 3D viewport layout、GEO 深度遍歷與完整 renderer 仍保留 boundary。

第一百六十二輪功能／文件 commit：`4b88cf6`，已推送至 GitHub `main`。依 reference `draw_3D_8x8_titles` 建立十組 `idxOffset／rowCount／colCount` wall layout，輸出 `WallStamp` 並接到 dungeon preview 的原始 8×8D sample；補上 layout regression、READY 規格與共用 graphics knowledge。`Draw3dWorldFar/Mid/Near` 方向遍歷、遮擋、sky／roof、door 與 camera 仍保留 boundary。

第一百六十三輪功能／文件 commit：`519ec4f`，已推送至 GitHub `main`。依 reference `Draw3dWorldFar/Mid/Near` 將 party direction 與 GEO wall fields 接成 ordered `WallLayoutCall` traversal，dungeon preview 改用 Far／Mid／Near 全段 wall stamps；補上 depth／座標 regression、READY 規格與共用 graphics knowledge。GEO wrap、sky／roof、door、遮擋與 camera 仍保留 boundary。

第一百六十四輪功能／文件 commit：`b890ffc`，已推送至 GitHub `main`。dungeon preview 保存 `(dungeonX,dungeonY)`，方向鍵透過 GEO 雙側 wall contract 移動，成功後重建 floor 與 Far／Mid／Near wall stamps；新增 position／camera slice 規格與共用 graphics knowledge。Area／save 真實座標、direction、movement cost、encounter、wrap 與 scroll animation 仍保留 boundary。

第一百六十五輪功能 commit：`6f530a9`，已推送至 GitHub `main`。將 dungeon preview 的 `(x,y,direction)` 接入 remake game save version 3；v1/v2 舊檔與越界值回到安全預設，F9／啟動載入後重建 floor／wall view。新增 save codec／game adapter round-trip regression、READY 規格與共用 state knowledge。原版 DOS `SAVGAM?.DAT` container 與 Area 真實座標寫回仍保留 boundary。

第一百六十六輪功能 commit：`e45cd03`，已推送至 GitHub `main`。將 dungeon preview Q/E 接成 reference 八方向 facing rotation（`±2`），轉向後重建 Far／Mid／Near wall view；新增 state wrap regression、READY 規格與共用 state knowledge。Area 真實 `mapDirection`、轉向時間、sky／roof／door overlay 與完整 3D viewport 仍保留 boundary。

第一百六十七輪功能 commit：`9324a03`，已推送至 GitHub `main`。依 reference `getMap_XXX`／`MovePositionForward` 將 dungeon 16×16 coordinate wrap 接成明確 `geo` wrapped API、wrapped Far／Mid／Near traversal 與 preview 跨邊界 movement；strict API 保留給不允許 wrap 的 context。原版 `mapWallType/mapWallRoof` 五-byte save segment、ECL 例外判斷與完整 3D overlay 仍保留 boundary。

第一百六十八輪功能 commit：`5cc1b42`，已推送至 GitHub `main`。依 reference `ovr017` 5-byte map segment 將 `mapWallType`／`mapWallRoof` 接入 remake save version 4；wrapped GEO refresh 會在位置／方向變更後重算，v3 舊檔 cache 回到 0。新增 save codec／game adapter round-trip 與 v3 compatibility regression、READY 規格及共用 state knowledge。完整 `SAVGAM?.DAT` container、slot、Area／ECL memory 與 player records 仍保留 boundary。

第一百六十九輪功能 commit：`f714494`，已推送至 GitHub `main`。解碼 Area1 `0x1FA/0x1FC` outdoor／indoor sky colour，依 `mapWallRoof > 0x7F` 接入 dungeon preview EGA sky background；新增 Area codec regression、READY 規格與共用 state knowledge。完整 `Draw3dWorldBackground`、roof geometry、door overlay 與原版 save container 仍保留 boundary。

第一百七十輪功能 commit：`a8c050e`，已推送至 GitHub `main`。依 reference `WallDoorFlagsGet` 接入 GEO wrapped no-wall default `1`／walled `x3` detail API，並在 dungeon preview 顯示目前 facing 的 door/detail evidence；新增 exact GEO regression、READY 規格與共用 state knowledge。開門、解鎖、撬門、撞門與 door symbol overlay 仍保留 boundary。

第一百七十一輪功能 commit：`1d292c0`，已推送至 GitHub `main`。依 reference `TryStepForward`／`MapSetDoorUnlocked` 接入 detail `1` unlocked doorway movement 與雙側 `UnlockDoorWrapped` raw mutation；detail `2/3` 保持阻擋。新增 GEO movement／mutation regression、READY 規格與共用 state knowledge。完整 bash／pick／knock menu、技能／骰點／法術消耗與 door graphics 仍保留 boundary。

第一百七十二輪功能 commit：`a6a1622`，已推送至 GitHub `main`。依 reference `Player.thief_skills` 將 DOS record `0xEA–0xF1` 保存到 `DOSPlayerRecord`／`Character`／JSON，提供 `OpenLocksSkill()` index 1 adapter；新增 synthetic parser regression、READY 規格與共用 party/state knowledge。thief skill 重算、pick-lock dice、door menu、bash／knock 與完整 DOS save container 仍保留 boundary。

第一百七十三輪功能 commit：`939b258`，已推送至 GitHub `main`。依 reference `pick_lock()`／`Spells.knock` 建立 `internal/dungeon` 的 injected d100 pick-lock resolver 與 Knock `0x1F` first-slot consume；新增逐位 roll、健康狀態、inclusive roll、失敗消耗嘗試與 spell removal regression、READY 規格與共用 dungeon knowledge。door mutation、完整 door menu、bash 與 thief skill 重算仍保留 boundary。

第一百七十四輪功能 commit：`ec2d0bb`，已推送至 GitHub `main`。將 pick-lock／Knock 接入 Ebiten dungeon preview 的 P/K action adapter；P 只允許 detail 2，K 允許 detail 2/3，成功後呼叫 GEO 雙側 unlock，新增 seeded State action regression 與 READY 規格。完整 locked-door menu、bash、door graphics 與劇情 integration 仍保留 boundary。

第一百七十五輪功能 commit：`db51598`，已推送至 GitHub `main`。依 reference `bash_door()` 保存 DOS `Str.full`／`Str00.cur`，建立 detail 2/3 的 strength／exceptional die resolver，並接入 dungeon preview `B` 撞門與 GEO 雙側 unlock；新增 bash table、extra-roll 與 DOS import regression、READY 規格。完整 locked-door menu、side effects、door graphics 與劇情 integration 仍保留 boundary。

第一百七十六輪功能 commit：`a53744b`，已推送至 GitHub `main`。依 reference `locked_door` 建立 detail 2/3 的 Bash／Pick／Knock／Exit capability resolver，並將方向鍵撞上上鎖門時接到 preview menu；新增 menu capability regression 與 READY 規格。完整 DOS 視窗樣式、door graphics、時間／傷害 side effects 與劇情 integration 仍保留 boundary。

第一百七十七輪功能 commit：`b86c485`，已推送至 GitHub `main`。依 reference `seg001.Init` 將 State、save／preview fallback 與 startup dungeon default 從 `(8,8,0)` 修正為 `(7,13,0)`，並清理已過時的 README／knowledge assertions；新增 READY 規格。`InitAgain` direction 2、完整 SAVGAM context 與劇情 integration 仍保留 boundary。

第一百七十八輪功能 commit：`cd2c0d9`，已推送至 GitHub `main`。依 reference `LoadPlayerCombatIcon`／`chead_cbody_comspr_icon` 將 DOS `icon_size=1` 的 CHEAD/CBODY raw slot 映射到 `+0x40`，載入 extracted raw layers 並在 renderer on-demand 合成；新增 small／normal icon regression、READY 規格與共用 party knowledge。direction-specific placement、recolor、animation 與完整 CombatIcon runtime 仍保留 boundary。

第一百七十九輪功能 commit：`2c3426c`，已推送至 GitHub `main`。依 reference `CombatIcon.LoadIcons` 將 attack layer 映射到 normal block `+0x80`，並接入 CHEAD／CBODY on-demand attack composition；新增 attack block regression、READY 規格與共用 icon knowledge。direction-specific placement、recolor 與完整 CombatIcon runtime cache 仍保留 boundary。

第一百八十輪功能 commit：`9f7c476`，已推送至 GitHub `main`。依 reference `SetupCombatActions`／`HalfDirToIso` 將 map direction 映射到 party／enemy `IconDirection`，接入 StartCombat 與水平 flip adapter；新增 placement／StartEncounter regression、READY 規格與共用 combat knowledge。完整 Area/ECL direction source、CombatMap placement、recolor 與 runtime cache 仍保留 boundary。

第一百八十一輪功能 commit：`bed7e56`，依 reference `ovr017.SaveGame/loadSaveGame` 建立 `SAVGAM?.DAT` 固定前綴 raw codec：保存 game area、Area1／Area2、runtime／ECL raw bytes、5-byte map state、game states、三組 block/set pair、party count 與 8 筆固定 CHRDAT name records；新增 strict-size validation、trailing player-file suffix boundary regression、READY 規格與共用 Gold Box save knowledge。完整 slot、Area 欄位解碼、個別 player files 與 file side effects 仍保留 boundary。

第一百八十二輪功能 commit：`216f7b2`，將 `SAVGAM` fixed prefix 接到 `game.State.LoadSAVGAMPrefix`／`SaveSAVGAMPrefix`；依 Area codec 更新已知 Area／GEO/map 欄位，保留未知 runtime／ECL／raw records，並以 signed map position、facing、wall cache 建立 State regression。此為 prefix load/export adapter，不取代 F5 remake JSON，也不宣稱已完成個別 CHRDAT player files、slot 選擇與 multi-file side effects。

第一百八十三輪功能 commit：`aa04200`，依 reference `seg044.SoundInit/PlaySound` 與 `Main/Resource.resx` 保存 9 個 PC WAV，建立 `internal/sound` selector catalog、WAV decode regression 與 Ebiten playback adapter；title start、荒野移動、dungeon preview 移動已播放對應原版音效，並加入 `-sound-dir`。完整戰鬥 sound calls、背景音樂、MIDI／AdLib 與音量設定仍保留 boundary。

第一百八十四輪功能 commit：`e0d871f`，建立 renderer-neutral `game.SoundEvent` queue，將武器命中／未命中、擊倒、移動時攻擊／免費反擊與已實作法術的 sound intent 接到 Ebiten `internal/sound` player；title start 與 wilderness movement 也改由 State queue 發送。新增 one-shot event order regression、READY 規格並清理上一輪「完整戰鬥音效尚未接入」的過時斷言；背景音樂、MIDI／AdLib、音量設定與所有 ECL sound calls 仍保留 boundary。

第一百八十五輪功能 commit：`9186a08`，依 `ovr017.SaveGame/loadSaveGame` 的實際檔名與 side-effect 順序，新增 `State.LoadSAVGAMSlot`：載入 `savgamA..J.dat`、`CHRDAT{slot}{1..6}.sav` 與 optional `.swg/.fx`，重用既有 DOS player parsers 建立 party／fighter 並進入繁中 wilderness；新增 `-savgam-dir/-savgam-slot` 啟動入口、synthetic slot regression、READY 規格與 state knowledge。Player.StructSize writeback、原始檔刪除與 CAMP multi-file save transaction 仍保留 boundary。

第一百八十六輪功能 commit：`9e64347`，依同一組 `ovr017` 證據新增 `PatchDOSPlayerRecord`、`.swg/.fx` encoders 與 `State.SaveSAVGAMSlot`。已載入角色的已證實欄位可回寫，未知 `.sav` bytes 保留；輸出先進 staging directory 再逐檔替換，仍保留原版刪檔、多職業、未知 sidecar 與 CAMP multi-file atomic transaction boundary。`go test ./...` 與兩個 CLI build 已於 Docker 通過。

第一百八十七輪功能 commit：`b40587c`，將 loaded SAVGAM slot 接到 Ebiten F5 與 CAMP SAVE；`-savgam-dir/-savgam-slot` 模式寫回同一 slot，一般模式維持 remake JSON。新增 workflow 規格與 README／PLAN／state knowledge 更新；原版刪檔、多職業、未知 sidecar 與跨檔案 atomic transaction 仍保留 boundary。

第一百八十八輪功能 commit：`ce606ad`，依已知 `SaveGame` side effect，將 SAVGAM slot 的 prefix 與 `CHRDAT{key}{1..6}` 檔案先移入 backup，再替換 staged bundle；隊伍縮編的 stale player sidecars 會被清理，替換失敗可 rollback。新增 stale-file regression 與 READY 規格；多職業、未知欄位與完整 player serialization 仍保留 boundary。

第一百八十九輪功能 commit：`84def37`，修正戰鬥結束時只同步 renderer-facing party、未同步持久 `partyRoster` 的狀態問題；現在 HP／MaxHP 會依 fighter ID 回寫 roster，供 CAMP 與兩種 save path 使用，並新增 regression／READY 規格與 state knowledge。

第一百九十輪功能 commit：`7b33384`，修正戰鬥結果按 Enter 後保留 stale ECL choices 的主流程問題；新增 `restoreWildernessMenu`，統一返回繁中荒野主選單，並加入 continuation regression／READY 規格與 state knowledge。

第一百九十一輪功能 commit：`a9806e9`，將城市商店 SELL 接入繁中 Shop Menu，依已解碼 item `Value` 將非 readied／非 cursed 物品售出並將 GP 放入 party pool；新增 menu／transaction regression、繁中 locale、READY 規格與共用 shop knowledge。城市 stock／ID／鑑定 routine 仍保留 boundary。

第一百九十二輪功能 commit：`befdc61`，將既有 `PayIdentifyFee` 接入繁中 Shop Menu，完成角色／物品選擇與 200 GP ID transaction，保留 `HiddenNameFlags` 與未解碼 magic result；新增 regression、locale、READY 規格與 shop knowledge。

第一百九十三輪功能 commit：`00682ff`，修正 `BlockSession` 跨 `NEWECL` 遺失 `LOAD FILES`／`PICTURE`／`SPELL`／`PROTECTION` 結果的問題；新增跨三個 synthetic ECL block 的 signal regression。`go test ./internal/ecl` 已於 Docker 通過；完整 `go test ./...` 仍受容器缺少 ALSA／X11 headers 及既存 game integration failure 影響。

第一百九十四輪功能 commit：`630d1b3`，修正真實 ECL1 JOURNEY ON integration regression：PICTURE 已是明確的繁中事件畫面，測試現在先驗證 request，再以 `Continue()` 模擬 Enter，最後確認流程抵達 COMBAT boundary。Docker non-Ebiten internal packages 全部通過。

第一百九十五輪功能 commits：`ad676f2`、`12a0fd7`，將 ECL `SPELL`／`PROTECTION` 結果接到 State pending queue，新增一次性 consume API，並驗證真實 `State.Select` wiring；State 保留原始 signal 順序／位址，不猜測未知 party memory side effect。`go test ./internal/game ./internal/ecl` 已於 Docker 通過。

第一百九十六輪功能 commit：`35fffaa`，新增 `SmokeInitializationEntries` 與 `cmd/azure-bonds -entry-smoke`，逐一 bounded 執行 ECL1–ECL6 全部 block 的五個 initialization entries；實際 image smoke run 已記錄 menu／COMBAT／monster spawn／unsupported opcode per entry，並新增 READY 規格與 ECL knowledge。

第一百九十七輪功能 commit：`d1327af`，以真實 ECL2 block 3 entry 3 與 `MON2CHA.DAX` 建立 playable Battle regression；修正 `MON*CHA` 50..60 packed ArmorClass 的 `60-raw` adapter，並新增 `-encounter-monster-member` 支援跨章節 direct encounter。ECL2 direct entry 已於 Docker 通過，正常玩家流程仍待完整 ECL continuation。

第一百九十八輪功能 commit：`860f7c4`，遊戲啟動載入 `MON1CHA`–`MON6CHA`，State 依 ECL global block namespace 選擇 chapter-local monster table；新增 ECL2 chapter selection regression。`go test ./internal/game ./internal/monster ./internal/ecl` 已於 Docker 通過。

第一百九十九輪功能 commit：`cb70681`，以原始 ECL1 block `0x50` payload `+0x5B5` 驗證 `NEWECL 0x03` 會切換到 ECL2 block `3`，新增 global session regression；target 後續 unsupported routine 仍保留 bounded stop boundary。

第二百輪功能 commit：`f822c89`，依既有 ECL command table／operand contract 將 `0x2F AND` 與 `0x30 OR` 接入 bounded 16-bit memory destination semantics，新增 regression 與 READY 規格；另建立可供後續 Gold Box 作品沿用的 [`gold-box-ecl-command-set.md`](docs/knowledge/gold-box-ecl-command-set.md) 指令集知識庫。ECL1–ECL6 smoke 已遇到的 `0x2D CALL` 仍維持 unsupported，待確認 external dispatch／return context 後再實作。

第二百零一輪功能 commit：`a04f6d6`，以原始 ECL3／ECL4／ECL6 smoke 的 `code 0x01` monster operands 為證據，將 `LOAD MONSTER`／`SETUP MONSTER` 接到 bounded runtime memory resolution，加入 byte-range validation 與 variable descriptor regression；ECL3 block 17／18、ECL4 block 33／37 real entries 已抵達 COMBAT／spawn boundary。完整 `CALL`／external routine、party memory 與玩家流程仍保留 boundary。

第二百零二輪功能 commit：`c45888e`，以 ECL1–ECL6 raw scan 的 `0x2E10`／`0xC01E`／`0xB200` 非 code-segment CALL operands 與 ECL3 opening 的 CALL→PRINT／menu sequence 為證據，新增 `RunResult.CallAddresses` external dispatch signal；bounded VM 從 CALL 後續 instruction 繼續，ECL3 block 16／17／18／21 smoke 已越過原本 `0x2D` stop。真正 engine routine side effect 仍保留 boundary。

第二百零三輪功能 commit：`76a1fa4`，將已由 ECL3 block 16 entry 4 raw image 驗證的 Yulash smoke text 接入 zh-TW locale 與 State event message；未知 ECL text 原樣保留，raw runner 結果不變，新增 localization regression。完整 ECL 對話翻譯與其他作品文字仍需逐段反組譯／翻譯。

第二百零四輪功能 commit：`eb4ab29`，依 ECL3／ECL4 raw event runs 新增邪教徒／受傷牧師、戰火城市與小型魔法商店的 zh-TW catalog mapping；State 將 ECL text message 提前保存到 menu pause 前，新增 unknown fallback regression 與 READY 規格。完整事件分支、CALL routine side effect 與全部 ECL 文字仍保留 boundary。

第二百零五輪功能 commit：`11ea665`，依 ECL3 block 16 entry 4 的 raw `PRINT RETURN`→後續 menu sequence，新增 `RunResult.PrintReturnCount` 與 session aggregation；真實 entry 已由原本 `0x33` stop 推進至 `menu=true`，新增 bounded regression／READY 規格。DOS text-window layout 與完整後續事件仍保留 boundary。

第二百零六至二百零八輪功能 commit：`3dd0645`，依 ECL5 block 48 raw trace／scan 將 `LOAD CHARACTER`、`FIND ITEM`、`DESTROY ITEMS` 接成 bounded signals，保留 character address、inventory query／destroy IDs 並繼續 control flow；real entry 已由 `0x0A`／`0x32`／`0x40` stops 推進到 `NEWECL` boundary。party ownership、compare result、實際 item deletion 與完整事件分支仍保留 boundary。

第二百零九輪功能 commit：`74adb2f`，依 ECL5 sunlight event 的 raw evidence 建立
`DESTROY ITEMS` → persistent party roster adapter；新增可刪除 readied／stacked item
units 的 `Character.DestroyItemType`，並以 party／State regression 驗證。`FIND ITEM`
仍維持 query-only signal，compare result、完整 item namespace 與事件分支仍保留
boundary。Docker 已通過 `internal/party`、`internal/game`、`internal/ecl`、
`internal/locale` 測試。

第二百一十輪功能 commit：`ed4f162`，依公開 CoAB reference `CMD_Damage` 與六個
ECL raw scan，將 `DAMAGE` 五欄 `flags／dice_count／dice_size／damage_bonus／save_flags`
保存為 `RunResult.DamageRequests`，並接入 `BlockSession` aggregation；新增 synthetic
continuation 與真實 ECL2 block 3 `+0x1599` operand regression、READY spec 與 command-set
knowledge。Docker 已通過 `internal/ecl`、`internal/game`、`internal/party`、
`internal/locale`；target／saving throw／random roll／HP mutation 仍保留 party adapter
boundary。

第二百四十七輪重大里程碑：依 reference `Area1.field_6A00_Get/Set` 確認七個 game-time
words 位於 `0x18C..0x198`，接入 `area.State.GameTime`、Area1 binary codec、SAVGAM prefix
load、`State.SetAreaState`、`AdvanceGameTime` 與 remake save synchronization；新增 Area1
raw round-trip 與 State mirror regression、READY spec、README／PLAN／Gold Box state knowledge。
完整 calendar UI 與其他 unknown Area1 fields 仍保留 boundary。

第二百四十五輪重大里程碑：依 reference `race_ages`／`ovr018` 補上 single-class
`StartingAgeSpecFor` 與 deterministic `RollStartingAge`，明確映射六 race、六 supported
classes 的 base age／dice count／size，unsupported 組合回傳錯誤。新增 human fighter
range regression、READY spec 與 README／PLAN／Gold Box state knowledge；目前 starter
templates 與完整角色建立 UI 尚未自動套用，multi-class／half-orc 仍保留 boundary。

第二百四十六輪重大里程碑：將 `RollStartingAge`／`WithAgeEffects` 接入
`State.AddCreationCharacter`：保留可編輯 template 不變，加入隊伍時對 copied character 生成
deterministic age、套用六項 ability effects，再寫入 roster；新增 real creation regression。
完整 race/class／alignment 建立選單、多職業與 half-orc 仍保留 boundary。

第二百一十一輪功能 commit：`3068c2b`，將 `RunResult.DamageRequests` 接入 State
pending queue／`ConsumeDamageRequests()` exactly-once API，避免事件／選單 pause 遺失
script damage effect；新增 READY spec 與共用 State knowledge。由於 remake 尚未保存
原版五類 `saveVerse` 與 selected-character memory mapping，本輪不猜固定 HP mutation。
Docker 已通過 `internal/game`、`internal/ecl`、`internal/party`、`internal/locale`。

第二百一十二輪功能 commit：`387b9fb`，依 reference player `saveVerse` `0xDF–0xE3`
與 `CMD_Damage` flags，保存五類 saving throws 到 DOS parser／Character／JSON／record
writeback；新增 selected／whole-party DAMAGE resolver、natural 1／20、注入骰點、
transactional roster HP 與 stable-ID fighter sync。random-target／`CanHitTarget`、
affect save bonus 與死亡 continuation 仍保留 boundary。Docker 已通過
`internal/party`、`internal/game`、`internal/ecl`、`internal/locale`。

第二百一十三輪功能 commit：`7250778`，依 reference `CanHitTarget` 接入 ECL DAMAGE
random-target branch：低八位 target count、party-size random selection、raw hit bonus、
natural 1／20 與注入式 hit resolver；State 提供 resolver variant 並 transactional sync
roster／fighter HP。新增 party／game regressions 與 READY spec；AC／invisibility affect、
save-effect bonus 與死亡 continuation 仍保留 boundary。Docker 已通過
`internal/party`、`internal/game`、`internal/ecl`、`internal/locale`。

第二百一十四輪功能 commit：`442b51a`，依 reference `CanHitTarget`／`RollSavingThrow`
補上 DOS player `field_186 @ 0x186` signed saving bonus 的 parser／Character／JSON／
writeback，並接入 ECL DAMAGE save threshold；State 新增 default hit resolver，使用
fighter／equipment projected AC 與已證實的 invisibility `0x19`／`0x47` -4 attack roll。
blink／displace／其他 `CheckAffectsEffect` 與死亡 continuation 仍保留 boundary。
Docker 已通過 `internal/party`、`internal/game`、`internal/ecl`、`internal/locale`。

第二百一十四輪文件 commit：`efa1056`，更新 README、PLAN、ECL／State 共用知識庫與
212／213 READY specs，明確區分已完成的 save bonus／AC／invisibility adapter 與後續
作品專屬效果邊界。

第二百一十五輪功能 commit：`b1a3298`，依 reference `CheckAffectsEffect(Type_16)`／
`AffectBlink` 修正 `CanHitTarget` 的 natural-20 順序：先放大為 100，再套用 effects；
active blink `0x25` 在 `actions.delay == 0` 時可將 attack roll 設為 -1。新增
`ECLHitContext` 與 State context adapter，讓戰鬥回合能傳入 action delay；displace 的
persistent affect-data bit、其他 effects 與 death continuation 仍保留 boundary。
Docker 已通過 `internal/party`、`internal/game`、`internal/ecl`、`internal/locale`。

第二百一十五輪文件 commit：`0da9742`，更新 ECL／State 共用知識庫、READY spec、
README 與 PLAN，清除 blink 已完成後的過時描述。

第二百一十六輪功能 commit：`4cfa81e`，依 reference `AffectDisplace` 將 ECL `DAMAGE`
命中投影延伸到 displace `0x59`：FX effect-data 第一 byte 的 `0x10` consumed bit
使首次攻擊 miss、後續攻擊可命中；combat round 0 且 attack roll 為 0 時清除此 bit。
第二個功能 commit：`d4f4a51`，State working roster deep-copy effects，確保多筆 DAMAGE
transaction 在後續 request error 時 rollback displace bit，不污染 live roster。Docker
已通過 `internal/party`、`internal/game`、`internal/ecl`、`internal/locale`。

第二百一十六輪文件 commit：`6a5f2ed`，更新 ECL AC／effects READY spec、Gold Box ECL／
State 知識庫、README 與 PLAN，記錄 displace data mapping 與 transactional rollback。

第二百一十七輪功能 commit：`8dc0c1e`，依 reference `damage_player` 將 ECL DAMAGE
傷害結果投影到可向後相容的 `Character.HealthStatus`／`DamageOutcome.Health`：exact zero
為 unconscious、1..9 overkill 為 dying、10+ overkill 為 dead，animated exact zero
亦為 dead；非 OK／animated 狀態 HP 寫回 0。DOS 固定 player record 未被臆測新增欄位。
Docker 已通過 `internal/party`、`internal/game`、`internal/ecl`、`internal/locale`。

第二百一十七輪文件 commit：`9ad9f79`，更新 ECL DAMAGE READY spec、Gold Box ECL／
State 知識庫、README 與 PLAN，區分已完成的 health-state projection 與尚未接入的
`CheckAffectsEffect(Death)`、bleeding、combatant removal、party win/loss continuation。

第二百一十八輪功能 commit：`d8825a4`，新增 `combat.Battle.SetHitPoints` external bridge；
State active-combat ECL DAMAGE resolution 會把 roster HP 同步到 Battle fighter、重新
計算 party／enemy status，並在 status 結束時走既有 `finishCombat` continuation。新增
active party defeat regression；完整 `CheckAffectsEffect(Death)`、bleeding、effect removal
仍保留 boundary。Docker 已通過 `internal/combat`、`internal/game`、`internal/party`、
`internal/ecl`、`internal/locale`。

第二百一十八輪文件 commit：`e6f8e2b`，更新 ECL death continuation READY spec、Gold Box
ECL／State 知識庫、README 與 PLAN，清除 active combat win/loss 已完成後的過時描述。

第二百一十九輪功能 commit：`37d2678`，依 reference `RemoveCombatAffects` 建立
`Character.RemoveCombatAffects` 的 19-kind cleanup table，並在 active-combat ECL damage
角色倒下時接入 State；blink `0x25`／invisibility `0x19`／`0x47` 因不在 reference
清單中保留。新增 party cleanup 與 State death regression；`CheckAffectsEffect(Death)`、
bleeding、完整 combatant removal 仍保留 boundary。Docker 已通過 `internal/combat`、
`internal/game`、`internal/party`、`internal/ecl`、`internal/locale`。

第二百一十九輪文件 commit：`b31e551`，更新 death/effect READY spec、Gold Box ECL／
State 知識庫、README 與 PLAN，記錄 cleanup table 與尚未解出的 Death side effects。

第二百二十輪功能 commit：`eb24633`，依 reference `CheckAffectsEffect(Death)`／
`sub_3BEE8`／`AffectTrollFireOrAcid` 建立 `DeathEffectContext` 與
`Character.ApplyDeathEffects`：Bleeding 保存 overkill；affect_63 對 dying／unconscious
在明確 combat-heal 條件下恢復並建立永久 affect_5F；troll effect `0x64` 只在已知非
火／酸 damage flags 時以 3d6 建立 TrollRegen `0x66`。`State.ResolveDeathEffects` 以
deep-copy transaction 接入 roster／Battle sync；未猜測 ECL DAMAGE 缺少的 damage type。
Docker 已通過 `internal/combat`、`internal/game`、`internal/party`、`internal/ecl`、
`internal/locale`。

第二百二十輪文件 commit：`800c064`，更新 Death side-effect READY spec、Gold Box ECL／
State 知識庫、README 與 PLAN，明確保留 dragon-slayer target side effect 與其他未知
Death routine boundary。

第二百二十一輪功能 commit：`a0738a9`，依 reference `AffectDragonSlayer` 建立
`Character.ResolveDragonSlayer` 與 `State.ResolveDragonSlayer`：只有 explicit
`MonsterTypeDragon` target 才以 injected d12 計算 `1d12*3 + 4 + strength damage bonus`
並回傳 attack roll `+2`；非龍目標不觸發。target kind／Strength bonus 不從 ECL DAMAGE
五 operands 猜測。Docker 已通過 `internal/combat`、`internal/game`、`internal/party`、
`internal/ecl`、`internal/locale`。

第二百二十一輪文件 commit：`bd0b6da`，更新 Death effect READY spec、Gold Box ECL／
State 知識庫、README 與 PLAN，記錄 dragon-slayer explicit target contract。

第二百二十二輪功能 commit：`5abf8be`，依 reference `RemoveFromCombat`／`CombatantKilled`
將 Battle fighter HP=0 的 combat position 清為 `HasCombatPosition=false`，對應原版
`CombatMap[player_index].size=0`；既有 Battle win/loss 與 finishCombat continuation
保持不變。新增 State active-combat regression；skull overlay、actions clear、完整 map
redraw 與其他 Death routine 仍保留 boundary。Docker 已通過 `internal/combat`、
`internal/game`、`internal/party`、`internal/ecl`、`internal/locale`。

第二百二十二輪文件 commit：`4353c21`，更新 Death／combatant removal READY spec、
Gold Box ECL／State 知識庫、README 與 PLAN，記錄 position removal 與剩餘 renderer side
effects。

第二百二十三輪功能 commit：`eea573d`，將 reference `CombatantKilled` 的死亡視覺需求
建立為 renderer-neutral `Fighter.DeathOverlay` signal；Battle 在 HP=0 時同時保留死亡時
CombatX/Y anchor、清除 `HasCombatPosition`，治療時清除 overlay signal。Ebiten combat
renderer 已在該 anchor 畫出可見的繁中「倒下」overlay；exact `combat_icons[24]/[25]`
skull asset 尚未因 CPIC/COMSPR byte-family 證據不足而硬編。新增 Battle regression。
Docker Go 1.23 已通過 `internal/combat`、`internal/game`、`internal/party`、`internal/ecl`、
`internal/locale`；`cmd/azure-bonds-game` 編譯仍受容器缺 ALSA/X11 headers／Ebiten backend
限制，並非本輪核心測試失敗。

第二百二十三輪文件 commit：`bd31db2`，更新 Death／combatant removal READY spec、Gold Box
ECL／State 可重用知識庫、README 與 PLAN，記錄 DeathOverlay contract、目前 renderer fallback
與 exact skull 素材證據邊界。

第二百二十四輪重大里程碑 commit：`258fde2`，追證 `seg001.Init` 的 COMSPR icon
initialization：`combat_icons[24].GetIcon(Attack, 0)` 對應 `COMSPR` block `0x8B`，
`combat_icons[25].GetIcon(Normal, 0)` 對應 `COMSPR` block `0x19`。Ebiten 現在載入
COMSPR derived PNG，依 DeathOverlay signal 在死亡座標以 100ms phase 交替顯示原版
skull／blank overlay；更新 graphics／ECL／README／PLAN 知識庫。Docker Go 1.23 已通過
`internal/combat`、`internal/game`、`internal/party`、`internal/ecl`、`internal/locale`。

第二百二十五輪重大里程碑 commit：`f2372a7`，依 reference `Action.Clear` 將死亡後的
per-fighter `CombatAction`（delay、move、spell ID、guarding）建立成共用資料 contract，
Battle 在 HP=0 時清零；State 若倒下者正是 current turn，也清除施法、移動、檢視與 target
selection。新增 combat／State regressions，並更新 ECL／State READY spec、README、PLAN、
Gold Box knowledge base。Docker Go 1.23 通過 `internal/combat`、`internal/game`、
`internal/party`、`internal/ecl`、`internal/locale`。

第二百二十六輪重大里程碑 commit：`51ae23b`，補上 save／encounter 初始 HP=0 fighter
的 CombatantKilled boundary：`NewBattle` 立即清除 `HasCombatPosition` 與
`CombatAction`、發出 `DeathOverlay`，且 `StartRound` 不把倒下者放入 turns。新增
initially-downed regression，確認倒下角色不會佔用碰撞格；更新 README、PLAN 與 Gold Box
ECL knowledge。Docker Go 1.23 通過 `internal/combat`、`internal/game`、`internal/party`、
`internal/ecl`、`internal/locale`。

第二百二十七輪重大里程碑 commit：`e6aa0a2`，依 reference `Tile_DownPlayer=0x1F`／
`downedPlayers` 將 `Fighter.DownedCorpse` 與死亡 flash 分離：team party HP=0 角色保留
死亡座標但不佔用 CombatMap position；Ebiten 在 skull flash 後以繁中「倒下」corpse marker
顯示。Cure Light Wounds target routing 現可選到非 dead 的 unconscious／dying corpse，
普通治療同步 roster HP、清除 DeathOverlay，但不讓角色恢復戰鬥格；新增 combat／State
regressions。Docker Go 1.23 通過 `internal/combat`、`internal/game`、`internal/party`、
`internal/ecl`、`internal/locale`。

第二百二十八輪重大里程碑 commit：`a907601`，依 reference `combat_heal` 建立
`Battle.RestoreCombatant(fighterID, position)` explicit stand-up contract：只有
`DeathEffectContext.CombatHealAllowed` 的 affect_63 recovery，在 HP 恢復為 OK 後才以保存的
CombatX/Y 清除 `DownedCorpse`、恢復 `HasCombatPosition`；普通 Cure Light Wounds 仍只加 HP
並讓 corpse 留在原地。新增 Battle／State placement regressions，更新 ECL／State spec、
README、PLAN 與 Gold Box knowledge。Docker Go 1.23 通過 `internal/combat`、`internal/game`、
`internal/party`、`internal/ecl`、`internal/locale`。

第二百二十九輪重大里程碑 commit：`1041099`，建立 renderer-neutral
`combat.DeathOverlayFrame`：以 100ms cadence 交替 `COMSPR 0x8B`／`0x19` 九次後結束
flash。Ebiten 保存每個 fighter 的 flash start time；party 隨後顯示 `DownedCorpse`，enemy
停止繪製死亡小人，治療時清理 lifecycle state。新增 9-cycle core regression，更新 Death／
graphics／README／PLAN 知識庫。Docker Go 1.23 通過 `internal/combat`、`internal/game`、
`internal/party`、`internal/ecl`、`internal/locale`。

第二百三十輪重大里程碑 commit：`d40f877`，依 reference `RemoveFromCombat` 完成
Ebiten render lifecycle：enemy 的 DeathOverlay 九次 phase 結束後完全停止繪製名稱／HP，
並從戰場畫面退出；team party 的 DownedCorpse 則保留原座標與繁中「倒下」marker。更新
ECL／graphics／README／PLAN 知識庫。Docker Go 1.23 通過 `internal/combat`、`internal/game`、
`internal/party`、`internal/ecl`、`internal/locale`，`git diff --check` 通過。

第二百三十一輪重大里程碑 commit：`d5c6f8c`，依 reference `find_target`／
`BuildNearTargets` 建立 `Battle.SelectCombatTarget`：enemy turn 從 sorted、存活的 party
candidate 以 seeded RNG 選擇目標，同一回合的 multi-attack 維持同一 target，不再固定攻擊
party[0]。新增 combat／State deterministic regressions、READY spec 與 Gold Box state
knowledge；visibility／pathfinding／persistent Action.target／AI spell priority／guarding
仍保留明確 boundary。Docker Go 1.23 通過 `internal/combat`、`internal/game`、
`internal/party`、`internal/ecl`、`internal/locale`，`git diff --check` 通過。

第二百三十二輪重大里程碑 commit：`c5c0cc5`，以 headless Xvfb 啟動目前 Ebiten
`-encounter` direct-entry，擷取 640×400 實際繁中戰鬥畫面 `docs/screenshots/combat-game.png`，
並更新 README gallery、sprites manifest 與 live screenshot READY spec。畫面證明目前
renderer 可顯示繁中戰鬥訊息、party／enemy 小人、HP 與操作提示；明確標示這仍是
direct-entry vertical slice，不宣稱完整玩家流程。重新產生 parser screenshot，
`git diff --check` 通過。

第二百三十三輪重大里程碑 commit：`58668fa`，依 reference `PoolRadPlayer.field_33`／
`field_B5..B7` 保存 MON*CHA spell-list slots 與 magic-user level-use counts 到
`combat.Fighter`。敵方回合現在會先嘗試已核對的 Magic Missile `0x0F`：一級單枚、
2–5 damage、成功後 atomic 消耗 level-1 use，失敗或無可用次數才回到 physical attack。
新增 monster parser／combat／State regressions、READY spec、README／PLAN／Gold Box state
knowledge；其他 monster spells、MON*SPC effects、AI priority／range／saving throw 仍保留
明確 boundary。Docker Go 1.23 核心測試與 `git diff --check` 通過。

第二百三十四輪重大里程碑 commit：`7ece3a3`，依 reference `load_mob` 將
`MON1SPC`–`MON6SPC` 以 chapter-local monster ID 載入，並把九-byte raw affect records
以 copy 掛到 enemy fighter 的 `MonsterAffects`。新增 `BuildEnemiesWithAffects`、State
chapter table adapter、CLI loader、copy-isolation regression、READY spec 與 README／PLAN／
Gold Box state knowledge。隱形／加速／睡眠等效果的戰鬥語意仍未猜測。Docker Go 1.23
完整 `go test ./...` 與 `git diff --check` 通過。

第二百三十五輪重大里程碑 commit：`f703d80`，依 reference
`CanHitTarget`／`CheckAffectsEffect(Type_16)` 將 active monster affect `0x19`／`0x47`
投影為 combat target AC +4；inactive effect 不影響命中，raw record 不被消耗。新增
combat exact-boundary regression、READY spec 與 README／PLAN／Gold Box state knowledge。
其他 `MON*SPC` effect kinds 仍保留逐項證據 boundary。Docker Go 1.23 核心測試與
`git diff --check` 通過。

第二百三十六輪重大里程碑：依 reference `load_mob` 的 `field_A1` 解析
`MON*CHA[0xA1]` 為 `Record.AttacksPerTurn`，並依 `AffectHaste` 將 active
`MON*SPC` affect `0x27` 加倍接到 enemy fighter 的每回合攻擊次數。新增 raw offset／
active-inactive Haste regressions、READY spec 與 README／PLAN／Gold Box state knowledge。
完整 Docker Go 1.23 `go test ./...` 與 `git diff --check` 通過。

第二百三十六輪追加規則：依 reference `AffectSlow` 將 active `MON*SPC` affect `0x2A`
套用為每回合攻擊次數減半，並與 Haste 組合測試；目前 adapter 保留至少一攻下限。
movement half-actions、遠程彈藥與完整 weapon profile 仍保留 boundary。

第二百三十七輪重大里程碑：依 reference `Player.IsHeld`／`AttackTarget01` 接入 active
`MON*SPC` helpless／snake charm／paralyze／sleep（`0x1F`／`0x33`／`0x34`／`0x35`）。
Held enemy 會在 State enemy turn 跳過 physical／spell action；held target 由 combat core
套用 guaranteed-hit 例外，raw effect 不消耗。新增 combat、monster、State regressions、
READY spec 與 README／PLAN／Gold Box state knowledge。解除／豁免／持續時間／治療仍保留
boundary；Docker Go 1.23 核心測試與 `git diff --check` 通過。

第二百三十八輪修正：enemy 單次 physical attack 的中文訊息不再固定使用 `party[0]`，
而是依 `AttackResult.TargetID` 重新查找實際命中的隊員；當第一位 party fighter 已倒下、
第二位成為 target 時新增 regression，避免戰鬥規則與畫面文字分離。Docker Go 1.23
`internal/game`／`internal/combat`／`internal/monster` 測試與 `git diff --check` 通過。

第二百三十九輪重大里程碑：依 reference `ovr021.step_game_time`／`CheckAffectsTimingOut`
建立 `State.AdvanceGameTime` 七-slot raw clock adapter，採用
`{10,10,6,24,30,12,0x100}` 級聯進位與 slot→elapsed-minute conversion；party `.FX` 與
active battle raw effects 共用 timeout transaction，`Strength=0xFF` 永久 effect 保留，
slot-6 overflow 以 age cycles 保存。新增 clock normalization、slot-2 十分鐘換算、finite/
permanent party／battle effect regressions、READY spec 與 README／PLAN／Gold Box state
knowledge。完整 Docker Go 1.23 `go test ./...` 與 `git diff --check` 通過；REST interruption、
calendar UI、DOS age writeback 與完整 time-triggered ECL 仍保留 boundary。

第二百四十輪重大里程碑：依 reference REST loop 的 `step_game_time(1,5)`，將
`REST_START` 接到 `AdvanceGameTimeHours`，每 requested hour 推進 60 個 slot-1 minutes，
先處理 finite effect timeout，再執行既有每 24 小時 +1 HP natural healing。新增 REST
clock／effect-order regression、READY spec 與 README／PLAN／Gold Box state knowledge。
完整 Docker Go 1.23 `go test ./...` 與 `git diff --check` 通過；rest interruption、safe
location、spell-learning side effects 與完整 rest encounter table 仍保留 boundary。

第二百四十一輪重大里程碑：依 reference `CMD_EclClock`／`vm_LoadCmdSets(2)` 確認
`ECL CLOCK (0x34)` 必須吃 `timeStep`、`timeSlot` 兩個 operand；修正 command metadata，
新增 `ClockRequest` 與跨 `BlockSession` aggregation，並由 State 呼叫既有七-slot
`AdvanceGameTime`，讓 ECL 與 REST 共用 finite effect timeout。新增 ECL／State regression、
READY spec 與共用 Gold Box command-set knowledge。核心 `internal/ecl`／`internal/game`
測試通過；完整 `go test ./...` 在目前容器因缺 ALSA/X11 headers 的 `cmd/azure-bonds-game`
與 `internal/sound` build dependency 失敗，非本輪 Go logic failure。memory-backed operand
值、time-triggered event table 與完整玩家流程仍保留 boundary。

第二百四十二輪重大里程碑：將七-slot game clock 與 age-cycle overflow 接入 remake JSON
save version 5；`State.SavePartyFile`／`LoadPartyFile` 現在會保存並恢復時間進度，versions
1–4 仍可載入並使用零時鐘。新增 save round-trip regression、READY spec、README／PLAN／
Gold Box state knowledge。DOS SAVGAM Area1 clock raw offset、calendar UI 與完整 time-triggered
event table 仍保留 boundary。

第二百四十三輪重大里程碑：依 reference `Player.age @ 0x76`、Pool/Rad `age @ 0x30` 與
`NormalizeClock`，將 normal DOS player `.SAV/.GUY` age 接入 parser、`Character`、slot-6
overflow writeback 與 `PatchDOSPlayerRecord`；每次 slot-6 overflow 讓 party roster 每人加一歲，
並以 int16 saturation 防止 wrap。新增 parser／writeback／game clock regressions、READY spec、
README／PLAN／Gold Box state knowledge。Pool/Rad importer、age-based ability modifiers、
完整 DOS age UI 與多職業序列化仍保留 boundary。

第二百四十四輪重大里程碑：依 reference `StatValue.AgeEffects`／`Limits.RaceAgeBrackets`
建立 `Abilities.WithAgeEffects`，保存 dwarf／elf／gnome／half-elf／halfling／human 五段
age brackets 與六項 ability deltas，新增 human bracket regression。確認這是新角色生成
規則；既有 DOS player record 已含結果，故沒有把 helper 隱式接到 import／Fighter，避免
double-count。creation age UI、完整 race/class limits 與 Pool/Rad `0x30` importer 仍保留
boundary。

第二百四十八輪重大里程碑：依 reference `display_map_position_time` 與 Area1 raw field
mapping，建立 `State.GameTimeDisplay`／`GameTimeText` renderer-neutral 繁中 clock HUD；
一般畫面與荒野地圖現在顯示 `HH:MM` 及依七-slot scale 推出的日／月／年欄位。新增 READY
spec、回歸測試並更新 README／PLAN／Gold Box state knowledge；完整原版日曆規則與
time-triggered ECL 仍保留 boundary。

第二百四十九輪重大里程碑：將已驗證的 race/class restrictions、class minimums 與
`race_ages` single-class table 接入角色建立選單，從三個固定模板擴充為 22 個可驗證組合；
加入隊伍仍使用 copied-character age／ability transaction，Ebiten 以五列捲動顯示。新增
creation option regression、READY spec 與 README／PLAN／Gold Box state knowledge；
多職業、half-orc 與原版完整 create/modify/drop menu 仍保留 boundary。

第二百五十輪重大里程碑：依 reference `Gbl.RaceClasses`、`race_ages`、`Limits` 與 DOS
raw race `6`，新增 `RaceHalfOrc`、半獸人 age effects、DOS parser 與繁中建立選單的
cleric／fighter／thief 選項；依 `ClassId` index 修正 fighter `13+1d4` age。角色建立
目前共有 22 個已驗證 single-class 選項。新增 READY spec、parser／rules regression 與
README／PLAN／knowledge 更新；half-orc multi-class 與完整原版建立流程仍保留 boundary。

第二百五十一輪存檔／知識庫里程碑：整理 `SAVGAM?.DAT`、`CHRDAT{slot}{1..6}.SAV/.GUY`
與 optional `.FX/.SWG` sidecar 的可驗證邊界，確認角色 age 位於 `0x76..0x77` signed
little-endian，並加入 race `0x74`、class `0x75` 與 shared ECL flag mapping 文件；同時
完成 ECL `SAVE TABLE (0x35)` indexed write 與 regression。這些 raw-preserving contract
可供後續 Gold Box 作品沿用；完整 multi-class serializer 與未知 sidecar schema 仍保留 boundary。

第二百五十二輪主流程里程碑：修正 ECL `COMBAT (0x24)` 在 bounded VM 只回傳 next PC、卻
未保存 resumable runtime state 的問題；現在真實／synthetic battle victory 會接回同一個
ECL session，繼續 menu／picture／NEWECL／PROGRAM 9 結果。新增跨戰鬥 continuation regression、
READY spec 與共用 Gold Box ECL knowledge；完整各 block engine routine 與劇情 side effects
仍保留 boundary。

第二百五十三輪法術流程里程碑：將 `CAMP → MAGIC → CAST` 從 placeholder 改為可操作的
施法者／memorized slot／受傷目標三段選單；已核對的 Cure Light Wounds 會消耗 slot、擲
`1d8`、同步 roster／fighter HP，並以繁中訊息回到 MAGIC。SCRIBE、未知 spell rules、
高等級／多職業 slot 與完整施法時間仍保留 boundary。

第二百五十四輪玩家流程里程碑：新增 `CAMP → ALTER → RENAME`，以最多 15 bytes 的中文
input boundary 更新角色名稱、roster 與 fighter display name；SAVGAM writer 會保留角色
ID／sidecar basename 與未知 bytes，只 patch DOS name field。原版 code-page transcoding、
多職業 serializer 與完整 delete semantics 仍保留 boundary。

第二百六十二輪 multi-class rules 里程碑：新增 `Character.HasClass`，以保存的
`ClassLevels[8]` 判斷角色實際擁有的職業，舊 JSON／缺少 metadata 時退回 primary class。
`CanEquip`、CAMP MAGIC 與 combat cleric／magic-user gates 已使用此判定，Protection from
Good 也不再被 primary-class projection 誤導。新增 READY spec 與 Gold Box 可重用邊界文件；
THAC0、生命骰、高等級 spell capacity、部分 UI label 與完整 multi-class serializer 仍待
逐欄反組譯驗證。

第二百六十三輪工具里程碑：將已驗證的 DOS player age `0x76..0x77` 接成
`cmd/azure-bonds -set-age` 安全修改流程，要求明確 `-out-record`，不覆寫來源檔並保留
未知 bytes。新增 READY spec 與 CLI 文件；完整 SAVGAM slot replacement、sidecar atomic
transaction 與原版 player delete semantics 仍維持既有 boundary。

第二百六十四輪 ECL engine 里程碑：依 `ovr003.CMD_PartyStrength`／`CMD_PartySurprise`
接入 `PARTYSTRENGTH (0x1D)` 與 `PARTY SURPRISE (0x22)` 的 word destination request，
並讓 bounded VM／BlockSession 正常繼續與聚合。新增 READY spec、synthetic regression 與
共用 ECL command knowledge；實際 party stat calculation、AC scale、multi-class level
來源與 surprise result 仍待 State adapter 逐欄接線。

第二百六十五輪 ECL／State 里程碑：新增可注入的 `ecl.PartyContext`，由 State 使用目前
roster／fighter projection 解析 `PARTYSTRENGTH` 與 `PARTY SURPRISE`，將結果寫回 shared
ECL memory，並在 `NEWECL`／menu continuation 間保留。新增 context-resolved regression 與
READY spec；reference AC internal scale、完整 multi-class level／THAC0 table 仍保留邊界。

第二百六十六輪 ECL／State 里程碑：依 `ovr003.CMD_CheckParty` 接入 normalized selector
`0xA5..0xAC` thief skills、`0x9F` movement 與 `8001` active-affect branches；四個結果
destination 由 PartyContext 寫回 shared ECL memory，新增 READY spec 與 min／max／average
regression。未知 selector、NPC／temporary party 語意與各作品 scaling 仍保留 boundary。

第二百六十七輪 ECL engine 里程碑：依 `ovr003.CMD_Who` 接入 `WHO (0x39)` prompt request，
跨 `NEWECL` 聚合並繼續 cursor；新增 READY spec、command knowledge 與 no-prompt regression。
這輪刻意保留 State roster UI／selected-player transaction，沒有自動替玩家選第一位角色。

第二百六十八輪 State／ECL 里程碑：將 `WHO` 做成真正可恢復的角色選擇 transaction；
interactive VM 會停在 WHO，State 顯示繁中 roster，選擇後透過同一個 BlockSession resume，
並保存 selected player ID。新增 READY spec 與 State regression；selected player 對所有
DOS global routine 的完整 side effects、NPC／temporary party 語意仍保留 boundary。

第二百六十九輪 ECL／State 里程碑：依 `ovr003.CMD_LoadCharacter` 將 `LOAD CHARACTER`
從 raw address signal 擴充為 decoded `LoadCharacterRequest`，保留低 7 bits 的 1-based
player selector 與 bit 7 restore/redraw metadata；State 已映射到 persistent `partyRoster`
並與 WHO 共用 selected player ID，無效 selector 有明確 not-found regression。完整
`FreeCurrentPlayer`、party summary redraw、external string context 與 NPC 語意仍保留
boundary。新增 READY spec、command knowledge 與可跨 Golden Box 重用的 VM→roster adapter。

第二百七十輪 ECL string-memory 里程碑：依 `ovr008.vm_CopyStringFromMemory` 的明確
`0x7C00 == SelectedPlayer.name` 特例，將 `PartyMemberContext.Name` 接入 resumable
`RuntimeState.Strings`。現在 `LOAD CHARACTER` 後的 `0x81` string operand 可經
`COMPARE → IF/GOTO` 走姓名分支，新增互斥 success/failure regression、READY spec、README
與共用 command knowledge。原先「完整 external string context 未完成」已收斂為
「`0x7C00` 姓名已完成，其餘 DOS memory regions 仍待逐區驗證」。

第二百七十一輪 ECL inventory-query 里程碑：依 `ovr003.CMD_FindItem` 將
`PartyMemberContext.ItemTypes` 接入全隊 item-record 查詢；resolved `FIND ITEM` 現會
清空 compare flags並設定 `=`／`<>`，可直接驅動原始 `IF/GOTO`。同一 VM run 的
`DESTROY ITEMS` 也會更新 working inventory view，後續查詢不再看到已毀 type；persistent
roster mutation仍由 State 負責。新增 found／not-found／find-destroy-find regressions、
READY spec、README 與共用 Golden Box ECL knowledge，並移除舊「尚未設定 compare」斷言。

第二百七十二輪 ECL selected-affect 里程碑：確認 opcode `0x3F` 是 `FIND SPECIAL`、
`0x3D` 是 CLEAR BOX；新增 `RuntimeState.SelectedPlayerIndex/Set`，讓 LOAD CHARACTER 的
1-based selector 與 WHO 的 0-based UI selection 更新同一份可恢復 selected identity。
`FIND SPECIAL` 現在查 selected member 的 active effects、回傳 resolved request 並設定
`=`／`<>`。新增 LOAD CHARACTER branch 與 WHO pause/resume 第二角色 regressions、READY
spec、README、State／command knowledge；缺 context 或尚未選角仍維持 unresolved。

第二百七十三輪 real-image verification 里程碑：重跑原始 ECL1–ECL6 共 25 decoded blocks／125 個 initialization entries，全部以正常 EXIT、menu、COMBAT、PROGRAM boundary 或 NEWECL 返回，unsupported-opcode error 為零。新增 corpus regression，若 member／block count、entry framing 或 bounded semantics 退化會指出精確 member／block／entry／PC。另以 ECL5 block `0x30:+0x0014` 與含 item `0x5E` 的 PartyContext 驗證真實 FIND ITEM found branch 抵達 `SUNLIGHT` 裝備腐朽文字。更新 READY spec、README、ECL knowledge，並移除第 196 輪仍記載 `0x2D/0x2F` unsupported stop 的過時斷言。

第二百七十四輪 ECL party-departure 里程碑：依 `CMD_Dump`／`FreeCurrentPlayer` 確認
`DUMP (0x3E)` 會移除 selected TeamList member、釋放 icon、減 party size，並選前一位／
新第一位；空隊伍則清除 selection。VM 新增 ordered `DumpRequest` 與 mutable working
PartyContext，讓後續 inventory／affect／party-rule query看見移除後狀態；State同步移除
persistent roster與同 ID fighter，且不套 ALTER DROP 的 last-member guard。新增中間／
最後角色 regressions，並鎖定 real ECL5 block `0x30:+0x020E` 的 Akabar DUMP opcode。
補充 cross-NEWECL regression：BlockSession 使用 caller PartyContext 的 deep copy作為同一
session mutable working party，target block可見已離隊結果，而呼叫端 context保持不變。

第二百七十五輪 ECL／State 里程碑：依 `ovr003.CMD_Program` 將外部 routine 0/3/8/9
集中到 `State.applyECLProgram`。一般事件與戰鬥後 ECL continuation 現在共用 start-menu、
party-killed、game-won／全隊 HP 與健康恢復／存檔詢問，以及 CAMP transaction。DOS 勝利
後 process exit 在桌面重製版明確映射為返回標題；新增四 routine regression、READY spec、
README 與可跨 Gold Box 沿用的 VM／作品 adapter 分層知識。

第二百七十六輪 ECL／map 里程碑：反組 `CMD_Call` 的 `operand - 0x7FFF` dispatch，
將原始 image observed `0x2E10/0xC01E/0xB200` 對應到 redraw、forced
`MovePositionForward` 與 sound A/B。State 現會讓 `0xC01E` 依 0/2/4/6 方向在 16×16
座標 wrap，frontend exactly-once 重建 dungeon floor／wall／roof；`0xB200` 先重現
reference default sound A。新增 ECL3 block 16 real CALL、四方向 wrap、ordered request
regressions與 READY spec；`word_1EE76 == 10` sound-B transient 仍保留 evidence boundary。

第二百七十七輪 demo／NPC 里程碑：依 `CMD_AddNPC.vm_LoadCmdSets(2)` 修正 NPC ID＋
morale framing，新增 `NPCRequest` 與 `DELAY` timing signal；真實 ECL1 block 0x52
現執行 53 steps，加入 RUSTLE／CYNTHIA／GRENDEL、輸出 11 段青色枷展示文字、聚合
`CALL 0x6803`／DELAY，最後抵達 COMBAT。State 依 chapter-local MON*CHA／SPC／ITM
建立 persistent NPC Character／fighter、control morale 與最低空 icon slot；NPC 專用
parser以唯一 ClassLevel修正 stale class_id，普通 save import仍嚴格。PICTURE 後 deferred
combat transaction也已接通，11 段文字逐行翻成繁中。第 278 輪依 `sub_29758` 確認
0x52 僅供 `inDemo`，正常玩家流程不可加入這三名 NPC。

第二百七十八輪正式 new-game 里程碑：`FinishCharacterCreation` 在 production ECL
session 現會 fresh reset 到 global block `0x01`（ECL2），而非人造荒野 menu。真實 entry
載入 FILES `1,2,FF`／PIECES `1,2,3`，依序顯示「小房間醒來、裝備與記憶消失」及
PIC 0x0A「持劍手臂出現奇異圖紋、全隊相同印記」兩段繁中。State 將 picture deferred
boundary擴充到 menu，Ebiten圖片下方顯示三行漸顯文字；real regression從角色建立完成
鎖定 block identity、兩段文字、picture、menu與 pieces。配合繁中文字較大，Ebiten
邏輯畫布也由 640×400 擴為 640×480；88px PIC／人物圖以 nearest-neighbor 3×、BIGPIC
以 2× 整數像素放大，文字則以 24px 高解析字型重繪，下方保留三行訊息與獨立 Enter 提示列。

第二百七十九輪 commit `6384eef` 曾正確找到 `+0x1CBB EXIT` 與
`C04B/C04C/C04D = 7/13/1`，但因 remake Area 零值誤稱為 wilderness `ModeMap`。
第 280 輪依 `seg001.Init/InitAgain` 的 `inDungeon=1` 推翻該 adapter：正式起點是提爾佛頓
GEO2 block 1 的 DungeonMap，script half-direction 1 對應 renderer 東向 2。

第二百八十輪未提交成果同時修正 `CMD_LoadFiles` operand 次序：operand 1 是 dungeon
GEO selector，operand 3 才是 outdoor BIGPIC。BlockSession 新增五-entry lifecycle API，
EXIT 保存 shared memory writes；正式 dungeon 成功前進會同步 `C04B..C04F`，依序執行
per-turn／SearchLocation，並把文字、PICTURE、menu、combat signal 接回 State。
Ebiten 正式流程會自動顯示既有 GEO／WALLDEF／8X8D 3D renderer，不再要求按 D。
另外確認 `ovr011.SetupWildernessFloor` 的 50×25 buffer 是野外遭遇 combat floor，
不是世界地圖；README、PLAN、spec 63／67 與共用知識庫已清除舊斷言。正式地城 UI
已依 640×480／24px 中文重排左右圖像區與分行狀態列，font loader 補 TTC collection；
`-opening` 走過真實序幕後以 Xvfb 擷取 `docs/screenshots/tilverton-opening.png` 並更新 README。

第二百八十一輪成果：依 `TryEncamp` 接通 ModeDungeon 的 `E → PreCampCheck(entry2)
→ CAMP → optional CampInterrupted(entry3)`。真實 block 1 在 `(7,13)` 寫 rest encounter
`0/0`，`x<5 || y<13` 寫 `1/100`；unsafe 24h rest 只推進 1 小時即中斷，不套完整 healing
／memorization，執行皇家巡邏繁中事件，Continue 後回原 dungeon。CAMP EXIT 以
`campReturnMode` 返回 3D view；一般 640×480 event text 改成 24px 五行換行。新增 READY
spec 281 與 real-image regression。

第二百八十二輪成果：修正 VM 長期缺少的 ECL byte code-memory mapping；reference
`0x8000..0x9DFF` 現在隨 block load／NEWECL switch 重載，GETTABLE 可讀腳本內 dispatch
table，code window 外的 shared memory 仍保留。AND／OR 也依 `CMD_AndOr` 補上
`compare_variables(result,0)` side effect。正式序幕後由 GEO2 `(7,13)` 往西抵達
`(6,13)`，SearchLocation selector `0x86` 進入 Windlord's Inn PICTURE 3 與兩段繁中事件；
事件引用 Journal Entry 31 時才將使用者提供 PDF 的中文全文加入遊戲內手札，最後返回原
地城格。PICTURE opcode 同步保存稍後會被 script 清除的 HeadBlockId，讓 HEAD3／BODY3
原始人物素材正確顯示；手札改為 24px、22 字寬七行排版。新增 READY spec 282、
synthetic VM gates、real-image regression 與 640×480 `tilverton-inn.png` README 截圖。

第二百八十三輪成果：反組 `CMD_Rob (0x28)` 的三 operands 與 reference
`RobMoney/RobItems`，VM 現發出 selected/all-party、loss percent、item chance request；
State 對 Copper／Silver／Electrum／Gold／Platinum 逐欄向下取整，並依 inventory 順序、
重量 `>24/-50`、`>255/-90` 與 deterministic `1d100` 處理物品。DOS player
`0xFB..0x108` 七個 money/treasure words 全部進入 typed parse/project/writeback。
正式 Tilverton GEO `(6,5)` selector `0x8A` 現顯示 HEAD5／BODY5 賢者菲拉妮；
「是 → 如實相告」執行真實 `ROB 1,50,0`，解鎖使用者提供 Journal Entry 38 的三頁繁中
全文，經兩個 Continue 返回原格。新增 READY spec 283、ECL/save knowledge、
`-filani` 重現入口與 640×480 `tilverton-filani.png` README 截圖。

第二百八十四輪成果：反組 `ovr003.CMD_Combat` 與 `ovr007.CityShop`，確認
`COMBAT (0x24)` 會在無怪物的 normal context 依 Area2 `EnterShop／EnterTemple`
分派 engine service。VM 現輸出一次性 shop／temple signal並保存 next-PC；State
以同一結果的 TREASURE／ITEM block 建立商品，套用原版 shift 計價、角色五種硬幣
優先與 pooled-money fallback，購買 clone 不耗盡庫存。正式 Tilverton `(2,12)`
selector `0x84` 已完成 Weaponers PICTURE 4、YES／NO、ITEM2 block 5、購買與離店後
`MAY YOU ALWAYS STRIKE TRUE.` continuation，最後返回原格；General Store 的舊
「COMBAT」測試斷言也已修正為 CityShop。新增 READY spec 284、共用 command-set
知識、`-weapon-shop` 重現入口與 640×480 `tilverton-weaponers.png` README 截圖。

第二百八十五輪成果：依 ECL2 real scan 確認 GEO2 `(0,7)` terrain `0x92` 是剛德祭壇，
會聚合 PICTURE 6（HEAD2 `9`／BODY2 `6`）與 EnterTemple service。State 現先保留
PICTURE boundary，Enter 後進入神殿；`ovr005.temple_shop/temple_heal` 的十種治療、
固定價格、`1d8／2d8+1／3d8+3／Heal`、blind／disease／poison／curse／stone／raise
effect transaction 與 typed-coin payment 已接入，離開後 resume ECL 並返回原格。
Raise Dead 的 Constitution／多職業 max-HP penalty 保留明確 boundary。另因這個事件
首次證明 HEAD／BODY selectors 不同，新增可擴張 masked scene compositor，BODY `y+5`
再覆蓋 HEAD，修正舊預產同號圖造成的缺圖／裁頭。新增 READY spec 285、
`-temple` 入口與 640×480 `tilverton-gond-temple.png` README 截圖。

第二百八十六輪成果：ECL2 GEO2 `(5,2)` terrain `0x8C` 的 Hall of Training 已由
PICTURE 4／中文詢問接到場所限定 `PROGRAM 0`。依 `ovr018.train_player` 保存 DOS
`Player.exp @ 0x127` dword，加入 cleric／fighter／paladin／ranger／magic-user／thief
經驗門檻、1000 GP 角色付款、擲兩次 hit die 取高、Constitution 與 class-level／HP
成長。一般 `PROGRAM 0` 返回標題的既有語意不變。新增 READY spec 286、`-training`
重現入口與 640×480 README 截圖；高等級固定 HP、種族上限、完整多職業 CON 與升級
選法術當時維持 boundary；前三項已於第 287 輪補齊，目前只保留升級選法術與
dual-class HP gate。

第二百八十七輪成果：回讀公開 CoAB reference `sub_509E0`／`get_con_hp_adj`／
`Limits.RaceClassLimit`，補上六職業 hit-dice 上限後的固定 `+1/+2/+3 HP`、只對
未達 hit-dice 上限職業計算的 Constitution、多職業除數與 dwarf／elf／gnome／
half-elf／halfling 職業等級上限。正式角色建立整合測試現從真實 ECL2／GEO2
`(5,2)` 跑過 PICTURE 4、中文詢問、YES、場所限定 PROGRAM 0、角色確認、扣 1000 GP、
升級／HP 成長，再離開返回同一格。dual-class HP gate 已於第 288 輪完成，目前只保留
升級選法術 boundary。

第二百八十八輪成果：補回 DOS Player `HitDice @ 0xE5` 與既有
`multiclassLevel @ 0xE6` 的 Character／JSON／raw patch round-trip。訓練升級後會像
`ReclacClassBonuses` 以 active class level 更新 HitDice；若尚未超過 dual-class
舊職業等級，仍扣款並升級但不增加 HP，超過後恢復一般 HP 成長。新增抑制／恢復
regression 與 READY spec 288。升級選法術經 reference 確認還需要
`spellCastCount[class, spellLevel]` 篩選；此模型與選單已於第 289 輪接續完成。

第二百八十九輪成果：定位 DOS Player `spellCastCount[3,5] @ 0x12D..0x13B`，加入
Character／JSON／raw patch round-trip。訓練依 `MU_spell_lvl_learn` 與 ranger
`unk_1A758` 重算容量，再用 CoAB spell class／level metadata 排除超過 5 級、容量為零、
monster／cleric 與已知法術。magic-user 升級及 ranger 新等級大於 8 會顯示不可取消的
繁中法術選單，選一個加入 KnownSpells；9 級遊俠 regression 同時鎖定 druid 與
magic-user 候選。

第二百九十輪成果：GEO2 `(6,10)` terrain `0x88` 的真實提爾佛頓酒館已接通
PICTURE 4／HEAD4／BODY4、酒館動作與四種飲料選單。`LEMONADE → YES` 會走過紫色
腰帶女子及側邊騷動，調查後找到華麗火焰形匕首；Adventure Journal Entry 17 原本只有
插圖，因此遊戲內手札使用忠實的中文圖像描述，不杜撰額外劇情。real-image regression
鎖定完整分支、手札解鎖、EXIT 返回同格及 stale choice 清理。新增 `-tavern` 重現入口與
`tilverton-tavern.png`；事件畫面改成獨立 640×480 layout，原始圖 3× nearest-neighbor、
中文 24px 直接重繪，修正探索 HUD 與人物圖重疊。共用知識庫同步記錄 16×15／24×24
中文字級與圖片、文字分離 pipeline。

第二百九十一輪成果：定位 GEO2 `(1,10)` terrain `0x8F` 為剛德神殿高階祭司。
真實 ECL2 流程顯示 PICTURE 6／HEAD6／BODY6，YES 分支施展 Remove Curse 並記錄
Journal Entry 19。使用者提供的 Adventure Journal 掃描 PDF 證實完整內容為青色枷
發光、射出藍焰並令眾人劇痛，祭司因神力遭抵抗而停止；此中文全文只在事件發生後
解鎖。real-image regression 鎖定問答、兩段 press-button pause 與同格返回；新增
`-high-priest` 及 640×480 README 實機圖。事件 caption 由 34 個英文字元假設改成
每行 22 個 Unicode 字元，24px 中文不再超出右邊界。Gold Box command-set 知識庫也
修正 CALL 的過期「未實作」斷言，新增五層 opcode 支援矩陣、signal exactly-once
時序與 ECL／engine memory ownership 表。

第二百九十二輪成果：正式新遊戲 session 已跑通 Tilverton `(1,0)` 皇家馬車主線。
Weaponers、Filani 與第一次城門警告是原 ECL memory 條件；第二次進入才顯示 PICTURE 11，
國王聲音使青色枷發光並強迫隊伍攻擊。整合測試載入 `MON2CHA.DAX`，要求建立單一
test hero 加五名 Royal Guard 的 active battle，不再接受「戰鬥規則尚未完成」占位；
以真實 combat actions 勝利後，續跑紅袍人劫走假國王、YES／NO 投降、牢房、
PICTURE 2／HEAD2／BODY2 盜賊救援與 Thieves' Guild 描述，最後確認 `NEWECL 0x02`
及 `(1,12,0)` map registers。新增 `-carriage` 正式條件 bootstrap、READY spec 292、
完整繁中敘事與 640×480 `tilverton-carriage.png` README 實機圖。共用 ECL 知識庫新增
「location state → combat boundary → pauses → chapter switch」不得 fresh-reset 的契約。

第二百九十三輪成果：反組 ECL2 block 2 `+0x046B..+0x04BC`，確認原版以
`LOAD CHARACTER 10..13` 逐名讀 selected-player `in_combat @ +0x100`，再寫
`combat_team/quick_fight @ +0x10C = 0x80`，讓四名 THIEF 成為我方 AI 友軍。
VM 現能投影 selected TeamList player-window、跨 pause 保存 team writes，並把
單一 15 人 spawn 拆成 4 名我方與 11 名敵方；混合陣營 AI 會攻擊相反 side，
不再停成四個玩家回合。正式新遊戲 regression 已由 Weaponers、Filani、皇家馬車、
投降與牢房一路抵達公會，驗證 hero + 4 allied THIEF 對 2 FIRE KNIFE + 11 THIEF，
勝利後顯示繁中遺言並解鎖只有地圖圖像的 Journal Entry 4。ECL block 2 的 local
`(1,12,N)` 與 GEO2 combined `(9,3,S)` 已有雙向 renderer adapter。戰鬥 HUD 改為
640×480 專用畫面：24×24 原始小人 nearest-neighbor 2×，隊伍色條／選取框及下方
24px 中文名稱與 HP，十八組文字不再互相重疊。新增 `-guildmaster`、READY spec 293、
知識庫與 `tilverton-guildmaster-battle.png` README 實機圖。

第二百九十四輪成果：公會戰 ECL `PartyMask` 產生的四名 QuickFight THIEF 現標記為
temporary allies，戰鬥結束後連同屍體從 active party projection 清除，避免污染下一場
犬舍戰與 `PARTYSTRENGTH`。正式新遊戲 regression 已繼續走過抱豎琴半身人、
1 FIRE KNIFE＋依隊伍強度縮放的 FIGHTING DOG、猴籠、奧莉芙・拉斯凱托訪客簿及
綠色黏液門，全部加入繁中。依 reference 邊界移動與 ECL2 block 2 entry 0，
renderer 現只回報 passable boundary attempt，由 State 寫 `0x7ED5=1` 後照常執行
`CALL 0xC01E → NEWECL 3`，真實流程已進入提爾佛頓下水道而非硬指定 block。
README／knowledge base／READY spec 294 同步確立 640×480 logical canvas：
原始像素素材 nearest-neighbor 整數放大，繁中以 24px（緊湊欄位可用 16×15）
獨立高解析重繪。

第二百九十五輪成果：ECL2 block 2 的下水道出口現會先把 combined GEO
`(10,15,S)` 寫入 source registers，避免 local guild X 被錯誤減十成負數；`NEWECL 3`
後再讀回 target `C04B/C04C/C04D`，正式落在 GEO2 block 3 `(0,1,S)`。block 3
initial entry 的惡臭、黏液、低天花板與濕滑戰鬥環境已繁中化。real-image regression
接著抵達 `(1,8)` terrain `0x81` 的火刀檢查哨，拒絕投降、驗證五名
MON2 FIRE KNIFE、實際打贏戰鬥，再由同一 resumable ECL 顯示藏起屍體的繁中
continuation。新增 `-sewers` 全故事重現入口、READY spec 295；subagent 的 block 3
唯讀盤點也整理出五入口 lifecycle ABI、`C04F&0x3F → ON GOTO` terrain dispatch、
camp entry 分工與主要 encounters，已收斂進共用 Gold Box ECL 知識庫。

第二百九十六輪成果：正式新遊戲 regression 在五名火刀檢查哨戰後繼續前往
GEO2 block 3 `(13,10)` terrain `0x83`，跑過遭屠滅的檢查哨與迷斯卓諾騎士出場、
青色刺青質問，以及 FIRE KNIVES／PRINCESS NACACIA／NO ONE 三項效忠 menu。
繁中 display labels 保持原始 menu index；選「娜卡西亞公主」後，騎士提示別殺
拿戰鎚的牧師並放行。最後 Continue 會落實 ECL first-visit／friend state，返回
ModeDungeon；重訪同 terrain 已驗證不重播。新增 READY spec 296，知識庫補上
multi-pause dialogue 必須保存 pending PC、plot mutation 可能延後到最後 Continue、
localized label 不可取代 script key 的共用契約。

第二百九十七輪成果：追蹤 ECL2 block 3 entry 0 後確認，下水道邊界除了
`0x7ED5` 還會先檢查 engine movement sentinel `0x7EC9`；公會轉場留下的 `0xFF`
若未在新步伐清除，會取消 exit attempt，接著把 E2 格誤派成 Otyugh 房間。
`RunDungeonExitLifecycle` 現在先同步 combined GEO、清除 stale sentinel、再交回
原始 ECL。正式新遊戲 regression 已從騎士分支走到 `(8,15,S)`，由 script 執行
`CALL 0xC01E → Y=0 → X-2 → NEWECL 4`，進入 GEO2 block 4 `(6,1,S)`。
target initial entry 的 `LOAD FILES 4,2,0xFF`、`LOAD PIECES 1,2,4` 與
`YOU ARE ENTERING THE HIDEOUT` 均在同一 session 聚合，入口文字已繁中化。
新增 READY spec 297，Gold Box 知識庫補上 boundary work flag 與 movement sentinel
是兩個不同 lifecycle 狀態的契約。

第二百九十八輪成果：盤點 ECL2 block 4 的 `C04F&0x3F` dispatch，確認 GEO2
terrain `0x99` 對應 selector `0x19` 的旋轉刀刃屏障。真實 ECL regression 鎖定
`ENTER THE BLADES / WAIT / RETREAT` 原始順序，並驗證 WAIT 不產生 DAMAGE、
刀刃減速消散的 continuation。State 新增「闖入刀刃／等待／撤退」、機關描述與
消散結果繁中；READY spec 298、README 與 Gold Box 共用知識庫同步記錄
640×480 畫布、24px 閱讀字級、16×15 緊湊字級及原始像素整數放大契約。

第二百九十九輪成果：補完刀刃屏障的危險分支。原始 ENTER index 0 先顯示
`THE BLADES TEAR INTO YOU`，下一個 press-button continuation 才送出
`DAMAGE flags=0xE0, dice=8d8, bonus=0, saveFlags=0`，最後與 WAIT 匯流到刀刃
消散。State 現只自動提交這種 whole-party auto-damage packet，以 seeded dice
對所有隊員套用同一傷害並同步 persistent roster／renderer fighter HP；選角、豁免
與 random-hit 形式維持既有 pending boundary。真實 ECL、State 兩人隊伍 100→62 HP
與 exactly-once consume 均有 regression，新增 READY spec 299。

第三百輪成果：接通 ECL2 block 4 selector `0x1A`／GEO terrain `0x9A` 的定身房。
原始 `RETREAT / INTERROGATE / KILL` 順序已繁中為「撤退／審問／殺死」而不改
script index；審問會先繳械逐漸恢復行動的火刀、取得情報並解鎖手札 26，殺死分支
則忠實顯示趁定身尚未解除時屠殺。手札中文依使用者 Adventurer's Journal 核對：
入侵牧師為營救南方首領房囚犯而施展定身，最後在此房被制伏。`4CFE & 0x40`
在選單前設定，因此三分支均消耗事件；真實 ECL 與 playable State regression 已
鎖定返回地城、手札不提前洩漏及重訪不重播。新增 READY spec 300。

第三百零一輪成果：補上原版地城 `SEARCH` 操作；640×480 地城按 `S` 時只在
SearchLocation invocation 期間設定 `7ECA=1`。火刀辦公室 GEO2 block 4
`(14,11)`／terrain `0x9B` 首訪描述房間並令 `4C10:0→1`，普通重訪無事；搜索才令
`4C10:1→2`、設定 `4CFE&0x80`，找到花梨木書桌文件並解鎖手札 9。手札中文依
使用者 PDF 第 12 頁圖像忠實記錄「燃燒靈氣、能附身其他軀體、與光芒之池有關」。
原始 `TREASURE(0,0,0,500,500,3,2,0x82)` 已接成 3000 GP 等值 pool、3 gems、
2 jewelry 與兩件 seeded random items；後續 COMBAT 正確視為 treasure service，
寶物 UI 返回 ModeDungeon。real-image／playable State regression 均鎖定防重複。

第三百零二輪成果：以 table-driven real-image／State regression 一次接通火刀據點
selector `0x1C–0x20`／terrain `0x9C–0xA0`。五個事件分別使用 `4C11..4C15`
visited byte：走廊奇怪煙味、整齊得異常且由看不見僕人復原的臥室、仍冒煙的焚毀
圖書館、遭同一烈焰完全摧毀的實驗室，以及標示「待復活／待埋葬」的兩排覆屍。
圖書館保留原始兩次 Continue；第二段取走焦屍手中未燒毀紙張後才解鎖手札 29。
中文依使用者 Adventurer's Journal 核對，保留盟友控制火焰、在軀體間移動、
異次元力量及「烈焰之主就是泰蘭索斯」線索。所有房間均驗證返回 ModeDungeon
且重訪不重播；新增 READY spec 302。
第三百一十輪成果：反組 ECL5 block `0x31` terrain `0x8A` 與共用子程序
`+0x0E0A`，完成 PICTURE 59 阿卡巴會面、YES／NO、`ADD NPC 0x3B,0x64`、
MON5CHA／SPC／ITM 入隊資料及解放前旅店。阿卡巴實際為 38 歲五級人類魔法師，
有兩件裝備、11 個 known spells 與 `4/2/1` 容量。子程序從 TeamList slot 0
逐人比對 `AKABAR BEL AKAS`，據此修正共用 `LOAD CHARACTER` 為 zero-based，
並新增 `Character.ScriptName`，使中文顯示名不再破壞 ECL script identity。
哈普解放後現會正確顯示阿卡巴的祕密商路提示。視覺契約仍為 640×480：
原圖／戰鬥小人 nearest-neighbour 整數放大，繁中正文 24×24 級、緊湊欄位
16×15 級，兩條 raster pipeline 分離。
第三百一十一輪成果：哈普地城出口現依 `4C5E` 提供地圖 CAVES 路線，
由真實 `NEWECL 0x32` 載入 Area 5 GEO block 50、pieces `8,FF,FF`，落在
`(15,5,W)` 的古老熔岩洞。入口伏擊使用 MON5 `0x39×4 + 0x31×3`，即四隻
火蜥蜴與三名黑暗精靈戰士；戰勝後保留同一 block 探索。修正 menu transaction
內 block switch 未清除來源 `7ED5/7EC9`，避免勝利後誤返荒野。新增
`-lava-tube` 真實 initial-entry 預覽與 640×480 `hap-lava-tube.png`；一般 ECL
menu 現能同時顯示 24px 中文 narrative，不再只剩選項。
第三百一十二輪成果：盤點 ECL5 block `0x32 +0x05B5` terrain dispatch，
確認 `ON GOTO` selector 1 才是第一個 target；terrain `0x8A` 第十項因此落在
`+0x10C6`，不是零起算會誤判的 `0x89`。真實 GEO5 `(9,10)` 現可觸發
火蜥蜴守門巡邏，使用 MON5 `0x39×3 + 0x31×3 + 0x33×1`。勝利後
`4C48 |= 0x08`，同一 resumable ECL 直接顯示繁中夢境警告並返回熔岩洞探索。
規格 312 與 Gold Box 指令集知識庫已保存 selector base 與戰後 presentation
不一定先停泛用勝利頁的契約。
第三百一十三輪成果：ECL5 block `0x32` per-turn 事件已在真實 GEO5
`(0,5,N)` 接通。terrain `0x89` 必須面北才顯示 PICTURE 57 的間歇泉與熔岩池，
接著保留 `COMBAT/WAIT/FLEE/PARLAY` 原始順序；已驗證 COMBAT 路徑建立 15 隻
MON5 `0x39` 火蜥蜴。勝利後 `4C48 |= 1`，發現六只防火桶；YES 進入 WHO，
一般英雄因熱度過強退回，再選 NO 返回洞內。繁中與 regression 已涵蓋每個
PICTURE／PRINT RETURN／menu／COMBAT／WHO boundary；知識庫新增方向敏感
per-turn terrain 與戰後環境志願者 selection 契約。

第三百一十四輪成果：重新對照 `CMD_EncounterMenu` reference 與 ECL5 block
`0x32 +0x01B7/+0x0281`，推翻上一輪「WAIT 直接進戰鬥」斷言。distance 0 的
WAIT／PARLAY 都把 behavior mode 4 解析為 result 3，進入五態度 PARLAY；
FLEE 的 mode 1 解析為 result 2，只有 COMBAT 才進 15 隻火蜥蜴。VM 新增
可恢復 PARLAY opcode，State 不再提前攔截 ECL FLEE／PARLAY；真實長流程驗證
WAIT→友善警告→無旗標離開，重訪後 COMBAT→戰鬥→防火桶。同步完成 640×480
renderer pass：24px 正文／16px compact 雙 CJK face、系統字型自動尋找、
dungeon 24×24 tile nearest-neighbor 2×、Combat／Dungeon HUD 分行與 Unicode
rune 換行，並移除 ModePlace 選單重複繪製。

第三百一十五輪成果：熔岩洞 GEO5 `(6,15,W)` 現依 block `0x32` per-turn
方向 gate 寫入 `C04B/C04C/C04D=7/15/3` 並 `NEWECL 0x33`，正式載入
GEO5 block `0x33`、pieces `14/15/FF` 與 PICTURE 51 五層法師塔庭院。
同一 resumable session 接續德拉坎德羅斯兩次 APPROACH、普通
`COMBAT/WAIT/FLEE/PARLAY` 選單、塔頂黑龍群、屠龍命令與煙霧幻象。
使用者提供的 Adventurer's Journal 條目 15 已整理成兩頁繁中，只在原 ECL
真正輸出 journal marker 後解鎖；事件再次寫入 `4CFF=1` 並令德拉坎德羅斯的
枷印消退，但不把已被火刀事件設定的同位址誤稱為第二枚計數器。最後停在
`ATTACK DRAGONS/ATTACK WIZARD/FLEE/PARLAY WITH THE DRAGONS` 真實 vertical
menu。新增 `-wizard-tower` 640×480 重現入口、READY spec 315 與共用知識庫。

第三百一十六輪成果：法師塔四項真實選單的 `ATTACK WIZARD`（index 1）已
沿同一 ECL5 block `0x33` session 接通。黑龍宣告不介入人類爭端後飛離，
德拉坎德羅斯召來守軍並逃下樓；戰場由 MON5 原始 records 建立 1 名伊弗利特、
2 名黑暗精靈戰士與 1 名法師。勝利後從原 COMBAT PC 續接「屋頂可安全休息」，
再由 EXIT 回 block `0x33` 地城。新增 `-wizard-tower-battle` 可重現入口、
READY spec 316，並在 Gold Box 知識庫保存一般文字選單不可按 label 提前攔截、
CLEAR MONSTERS 只清 encounter build list 的契約。

第三百一十七輪成果：法師塔 `PARLAY WITH THE DRAGONS` 已接入原始
`+0x05EA PARLAY 1,0,0,0,1,[7F79]`。五態度依序為
HAUGHTY/SLY/MEEK/NICE/ABUSIVE；傲慢與威嚇進入 14 隻 MON5 `0x35` 黑龍戰，
其餘三種會播放「沒有對付龍族的陰謀」繁中對話，再匯入德拉坎德羅斯四名
守軍戰，勝利後仍回安全屋頂。新增獨立 `gold-box-parlay.md` 知識文件、
READY spec 317、`-wizard-tower-parlay` 重現入口與 640×480 黑龍繁中實機圖。

第三百一十八輪成果：法師塔 outer menu 的 `ATTACK DRAGONS` 與 `FLEE`
均由原 ECL 匯入 MON5 `0x35×14` 黑龍戰，證實此處撤退不成功。勝利後保存
`7EC7 > 0x80` raw 重戰 gate；`4C61==1` 時可選擇取龍心，YES 會播放酸液繁中、
自動解析全隊 `DAMAGE 0xC0,3d4+3,save type 1` 並寫 `4C64=1`，NO／不符合
條件則跳過。State whole-party resolver 現可從混合 pending queue 取出 `0xC0`
packet 而保留需 selected target 的舊 packet。新增 READY spec 318、
`-encounter-area` graphics namespace 與 640×480 Area 5 原版 14 黑龍實機圖。

第三百一十九輪成果：法師塔塔頂 GEO5 block `0x33 (7,15,E)` terrain `0x01`
出口已由真實 ECL 接通。第一層保留 `CAVES/WILDERNESS/STAY HERE`；WILDERNESS
不再被泛用 label adapter 提前攔截，會繼續顯示 `VILLAGE/DEPART`，四條結果分別
鎖定 block `0x32/0x31/0x30` 或留在 `0x33`。NEWECL 後同步 destination GEO 與
target initial registers，新增 READY spec 319、`-wizard-tower-exit` 及
640×480／24px 繁中實機圖。共用 Gold Box 知識庫同步確立：原圖 nearest-neighbour
整數放大、CJK 直接在高解析 logical canvas rasterize，ordinary menu label 不可
脫離 block context 當成引擎 action。

第三百二十輪成果：法師塔祕道 DEPART 已沿真實 session 完成 ECL5 block `0x30`
離場程序。`LOAD CHARACTER` 現將 party `ControlMorale` 投影到 selected-player
`0x7CB8`，阿卡巴不再因只有姓名投影而被錯判不存在；完成塔與哈普時，他會以
繁中告別並由原 DUMP 離隊。下一個獨立 Continue 顯示日光使黑暗精靈裝備腐朽，
並銷毀 item `0x5E/0x60/0x61`。最後 `NEWECL 0x50` 顯示 BIGPIC 121，回到
ENTER CITY／JOURNEY ON／CAMP 世界流程。新增 READY spec 320、synthetic
selected-window regression、real block 0x30 regression 與完整塔頂長流程驗證；
Gold Box 共用知識庫同步記錄 control/morale projection 與 DEPART cleanup 契約。

第三百二十一輪成果：Area 5 離場回到哈普後，`JOURNEY ON → ESSEMBRA → TRAIL`
現沿 ECL1 block `0x50 +0x149A` 顯示龍巫妖復仇繁中事件並進入戰鬥。此 script
雖位於 ECL1，卻 `LOAD MONSTER 0x3C,1,0x3C`，實際 record 是 MON5 DRACOLICH；
State 因此改為逐 spawn 依 monster ID range 選擇 MON*CHA／MON*SPC，ECL chapter
只作 fallback。MON5 record 解出 66 HP、raw AC 66→AC -6、3d8，並使用 CPIC5
原版小人。勝利後正式抵達艾森布拉城外。新增 READY spec 321、real-image 與
法師塔至艾森布拉長流程 regression，並更新 640×480 README 實機圖。

第三百二十二輪成果：依 DOS 原版城市／戰鬥截圖重建 640×480 畫面拓撲。
冒險畫面恢復左圖、右隊伍 AC／HP、下方敘事與最底命令列；戰鬥恢復左戰術格、
右 active／target 狀態、下方訊息與命令列。reference `draw_head_and_body`
的 `row+5` 已由錯誤 5px 修正為五個 8px 列（40px），並以 HEAD→BODY layering
修復臉貼在胸口。戰場改為 clipping target，不讓大型怪物越入狀態欄；在
combat terrain selector 尚未解出前使用中性戰術格，不再誤鋪 TILES icon atlas。
新增 READY spec 322、Docker DOSBox reference captures、兩張 Xvfb 實機圖，
並把可沿用規則寫入 Gold Box graphics knowledge。

第三百二十三輪成果：設計審查以 DOS `combat-aim`／`fight-black-dragon`
逐像素量測，推翻第 322 輪仍過高的戰鬥面板。上方改回原生 320×184 的精確
2× geometry：戰場 `(16,16,336,336)`、16px 中央石框、256px active status；
640×480 多出的 80px 僅作中文 log，最後 32px 保留兩列 footer。移除可見
checkerboard、紅藍 team bars 與右欄 target card，改用 EGA 灰底、青／綠／黃
資訊層級及 48×48 active cursor。大型敵人 occupancy 未解出前不顯示錯誤的一格
target box。新增 READY spec 323；terrain、原始 stone-frame tiles 與大型怪物
anchor 明確保留為下一輪 visual RE boundary。

第三百二十四輪成果：確認原版戰鬥 terrain 不在一般 TILES，而是
`DUNGCOM/WILDCOM/RANDCOM`。三個單 block payload 均為 17-byte SSI header 加
24×24 4bpp items，CoAB 分別為 25／34／6 張；新增 bounded codec、原始檔
regression 與三套 gallery。Ebiten 地城戰場現由既有 `GenerateDungeon`
50×25 background buffer 取 7×7 slice，再依 `BackgroundTile.TileIndex`
繪製真正 DUNGCOM 石牆／斜牆／轉角，取代單色 placeholder。完整 encounter
terrain-mode selector、WILDCOM procedural placement、RANDCOM decoration、
原始 stone-frame tiles 與大型怪物 occupancy 仍保留為後續 RE boundary。

第三百二十五輪成果：正式接通 WILDCOM 野外戰場。renderer 以 `MapX/MapY`
為中心查 `SetupWildernessFloor` 已還原的 50×25 buffer，將 7×7 entries 的
`TileIndex 0..33` 對應 WILDCOM 34 張原始 tile；實機圖已顯示樹、倒木、岩石、
草地與水岸。terrain family selector 改為只依 `Area.InDungeon` 選
DUNGCOM／WILDCOM，不再使用 `GameArea>1` heuristic；`-combat-terrain`
只保留作 deterministic visual verification。RANDCOM 六張特殊物件明確保持
decoration overlay boundary。新增 selector／camera tests、READY spec 325 與
640×480 野外戰鬥截圖。

第三百二十六輪成果：由 reference `CombatMap.size = player field_DE & 7` 與
MON1–MON6 原始 records 收斂怪物形狀碼：1／2／3／4 分別是
1×1／1×2／2×1／2×2。`CombatSize` 現由 monster parser 投影到 fighter，
移動、復活、鄰接與 camera 均使用完整矩形 footprint。Ebiten marker 依同一
shape 顯示；2×2 龍巫妖為 96×96，CPIC 鏡像仍保留 `6-x` 左上錨點，避免因
額外扣除寬度而被戰場 clipping。更新 READY spec 326、Gold Box graphics
知識庫與 640×480 DUNGCOM 龍巫妖實機圖。

第三百二十七輪成果：接通已存在但先前未繪出的 RANDCOM 原版裝飾 pass。
reference `sub_370D3` 在 GEO terrain bit `0x40` 的開放區，以 dice stream 寫入
table／chair BackgroundTiles entries `0x1A/0x1B`；其 graphic ID
`0x22/0x23` 屬於全域 namespace，應映射 RANDCOM `0/1`，不是拿去查只有
25 張的 DUNGCOM。renderer 現先畫 DUNGCOM floor `0x16`，再透明疊加
RANDCOM `id-0x22`；WILDCOM `0..33` 保持獨立。原始 catalog 掃描與
`GEO2 block 01, center (13,0), seed 1` 的 640×480 Xvfb 圖均確認桌椅可見。
新增 READY spec 327、atlas bounds tests、`-dungeon-x/-dungeon-y` deterministic
visual flags，並同步 README 與 Gold Box graphics 知識庫。

第三百二十八輪成果：將 BackgroundTiles 的 `MoveCost`／`0xFF` 從畫面資料接入
可玩的 combat MOVE transaction。新增 renderer-neutral `MovementTerrain`
callback；Battle 在 occupancy／attack／座標 mutation 前檢查目的地完整
footprint，任一格不可通行即原子拒絕，多格 cost 取最大值，並在 remaining
points 不足時保持位置與 move mode。State 改依 `MoveResult.MovementCost` 扣點。
作品 adapter 分流 reference `x≈22,y≈10` 絕對 CombatMap 座標與目前 0..6
formation fallback；地城查 `(18+x,7+y)`，野外查 MapX/MapY centered floor，
coordinate namespace 在 StartCombat 固定，不會移動途中重新猜測。新增 READY
spec 328、2×2／2×1 terrain regressions、State budget regression，並更新
README 與 Gold Box state 知識庫。

第三百二十九輪成果：移除 640×480 renderer 的 3px 仿石紋戰鬥框。兩張公開
DOS oracle 的 frame／divider／裂紋位置完全相同，原始 ZIP 94 members 又沒有
獨立 UI frame DAX，因此把 boundary 修正為固定 320×184 panel raster，而非尚待
尋找的 encounter stone tiles。新增 `gfx.CombatFrame()`：透明 battlefield／
status interiors、五個 8px frame regions、原生 1px EGA bevel、alternating
dotted inner edge 與固定 crack pixels；Ebiten 啟動時轉一次並 nearest-neighbour
2×。新增 native geometry／transparency tests、READY spec 329，更新 Gold Box
graphics 知識庫與 README 的 640×480 龍巫妖實機圖。

第三百三十輪成果：地圖引擎拆分。原本位於 CoAB `internal/geo` 的完整
16×16 GEO decoder／牆／門／移動規則，以及 `internal/gfx` 的原版
Draw3dWorld 遠中近視角走訪，已移至獨立 `golden-box-remake-engine`
的 `geometry`／`viewport` package；CoAB 僅保留相容 wrapper。game-pack
schema 新增 `maps`，散提爾堡內城以 JSON 宣告 area 4、GEO block `0x20`、
spawn `(2,0,S)`、wrapped 與 2× nearest-neighbour。下一步是 WALLDEF/8X8D
composition/resource loader 搬移，以及 640×480 第一人稱地圖實機截圖。

第三百四十二輪成果：獨立 engine 新增 `graphics`，接管 SSI indexed picture、
EGA RGBA、WALLDEF、LOAD PIECES global offsets 與 8X8D stamps；中立 block map
消除對 CoAB `internal/dax` 的反向依賴。CoAB map JSON 現指定 GEO/WALLDEF/8X8D
檔名並驗證 base filename。該輪曾把 recovered `-5..15` wall traversal columns
誤當成 176px panel 寬；第 347 輪 DOS oracle 已證實實際 chrome 是左 128px、
右 192px。舊 debug floor 不再混入 production。Docker/Xvfb 實機圖為
`docs/screenshots/tilverton-first-person-remake.png`。door／roof overlays、
斜向逐像素 DOS oracle 與 wilderness world map 仍是後續 map fidelity 邊界。

第三百四十三輪成果：世界地圖改讀原始 `BIGPIC1 block 0x79`，不再把
WILDCOM 50×25 combat floor 誤稱 overland。使用者提供的 Clue Book PDF 第 35
頁與攻略確認 CoAB 只能在興趣點間旅行。獨立 engine schema／`worldmap` 支援
作品中立 image、localized points 與 cardinal selection；CoAB JSON 保存 A–N
14 個 values／座標／翻譯。正式 `ModeWilderness` 顯示 608×240 nearest-neighbour
地圖、目前位置、旅行選單與繁中 HUD；`-world-map` 的 Docker/Xvfb 實機圖為
`docs/screenshots/coab-overland-map-remake.png`。route graph 自 ECL 匯出、
Shadowdale AREA overhead map 與 optional travel encounters 尚待後續。

第三百四十四輪成果：補上與世界地圖、戰鬥地板分離的 AREA 俯視地圖。獨立
engine `areamap.Project` 由 16×16 GEO grid 產生 terrain cells、去重實體牆段、
wall type 與 door detail；engine commit 為 `2ef18ca`。CoAB game-pack JSON
新增 `tilverton.area-map`，指定 Area 2、`GEO2.DAX` block 1；
與 2× scale。正式地城按 A 開啟、A／Esc 返回，舊 `ModeMap` 畫面也不再誤用
WILDCOM combat floor。中文字改讀本機倚天 `STDFONT.15` Big5 分區字模，
以 Monkey Island 2 已驗證的逐列水平 1px embolden 顯示 16×15 粗體；
optional `SPCFONT.15` 處理全形符號，字型檔因著作權不提交。640×480 實機圖為
`docs/screenshots/coab-area-map-remake.png`。本輪的向量 renderer 與
`8X8D2/01` 推測已由第 345 輪原版 symbol renderer 取代。

第三百四十五輪成果：依公開 reference commit `9dc46f1` 的
`ovr031.DrawAreaMap`／`seg001` 還原原版 AREA。全域 symbol set 4 在
`game_area=1` 時載入 `8X8D1.DAX/CA`；11×11 camera offset 為
`clamp(party-5,0,5)`，每格 N/E/S/W wall presence 組成 `1/2/4/8` mask，
選 local item `4+mask`，隊伍方向選 item `direction>>1`。原版不顯示門；
reference door pass 是 cheat。獨立 engine `443281a` 新增
`areamap.BuildOriginal` 與 schema `symbol_block`。CoAB renderer 現直接以
2× nearest-neighbour 畫 8×8 EGA 灰牆與白色方向箭頭，取代向量 16×16 全圖；
JSON 指定 `8X8D1.DAX/CA`。更新後的 640×480 倚天粗體實機圖仍為
`docs/screenshots/coab-area-map-remake.png`。

第三百四十六輪成果：修復正式第一人稱 viewport 的背景與 screen transform。
公開 reference `ovr031.Draw3dWorldBackground`／`seg040.DrawColorBlock` 證實
native sky/horizon/ground 為 `(24,24,88,44)`、`(24,68,88,2)`、
`(24,70,88,42)`；SKY FA／FB 依戶外 palette、hour、方向顯示，FC 固定地面
overlay。獨立 engine `8ea72d9` 新增 reusable projection 與 schema
`sky_file/sky_blocks`；CoAB JSON 指定 `SKY.DAX [250,251,252]`，原始 image
regression 驗證 88×16／24×24／88×48。另修正 wall stamp：只保留 logical
row/column 0..10，native position 是 `(column+3,row+3)×8`；舊 renderer
錯誤右移 48px、上移 32px，才形成 README 舊圖的三片巨牆。`-eten-font`
現在同時接管 regular/compact face，全畫面使用倚天 16×15 embolden。最新
Docker/Xvfb 圖已覆寫 `docs/screenshots/tilverton-first-person-remake.png`。

後續 DOSBox oracle 進度：已完整走到原版角色建立的 stats／命名／戰鬥小人
配色流程，確認選單需以游標與彩色功能鍵混合操作。實機新建 male dwarf
fighter 顯示 `AGE 55`（另一輪為 52），原生 320×200 證據保存為
`docs/reference/original-dos/character-age-create.png`，並補入 spec 251 與
Gold Box save-format 知識庫。角色 pool 的下一個阻礙是原版要求 A: floppy；
`/tmp` DOSBox harness 已嘗試以同一暫存目錄掛載 A:，尚未取得可載入 party，
不應把黑畫面誤列為 adventure oracle。倚天字型則已確認與 Monkey Island 2
`build_eten_font.py` 完全同構：原生 `STDFONT.15` 16×15，每列將 source pixel
向右 OR 1px；不做容易黏筆劃的 24→16 縮圖。README 已將此啟動方式列為建議值。

第三百四十七輪成果：發現原版標題的 `D` 可直接進入內建 demo，無須先完成
A: character-pool 流程；因此取得真正 native 320×200 冒險 chrome oracle
`docs/reference/original-dos/tilverton-first-person-demo.png`。畫面訊息明示
`NOWHERE IN THE REAL...`，只可用於 GUI／SKY／status layout，不可當 GEO2/01
牆配置證據。實測 top row 在 native x=128 分割、y=136 進入 message：
640×480 remake 現為 first-person 256×272、roster 384×272、message
640×176、footer 640×32；多出的 80px 僅擴充繁中訊息。CoAB JSON 新增
`tilverton.first-person`，明列 GEO2/01、WALLDEF2、8X8D2、SKY FA–FC、
spawn `(7,13,N)` 與 outdoor sky selector 3。獨立 engine `908cfb7` 已推送，
新增 `FindMapByKindLocation` 與 indoor/outdoor sky schema，使同一 geometry
block 的 AREA／first-person projections 不再互相誤選。正式 `-opening`
Docker/Xvfb 圖已更新 `docs/screenshots/tilverton-first-person-remake.png`。

2026-07-28 使用者要求先暫停新增功能並盤點成果。新增
`docs/project-status.md`，集中列出 engine／CoAB 已完成且可驗證能力、仍未完成
範圍、測試方式與目前 checkpoints（CoAB `762c012`、engine `908cfb7`）。
README 已連結此盤點。後續若恢復反組譯，可使用
`/home/anr2/ida_94_official/dist` 的 IDA Pro；仍須遵守
反組譯證據 → `docs/spec/` READY → 實作的 SDD 流程。目前不要自行續跑新的
combat demo／IDA 工作。

2026-07-29 已恢復作業。第 348 輪以本機 DOS 320×200 runtime capture 抽出
透明 cracked stone chrome，取代 generic 灰框；事件／人物圖以 cover＋clip
填滿左上內格。戰鬥框改用同一 oracle 的真實石框 strip，依已量測 x=176
divider 重組，證據標籤為 material-exact/layout-reconstructed，不冒充整框
pixel-exact。MobyGames PC-98 640×400 gallery（adventure 464680、combat
464695）納入 typography density oracle；一般正文／roster／HUD 以倚天
16×15 水平加粗為主，24px 僅留給標題。新增
`docs/knowledge/golden-box-remake-for-chinese-readers.md`，規定原版／美化
theme 分離、長文分頁與攻略的冒險紀行／逐區／無雷三層。使用者仍要求重大
進度才集中 commit/push。

2026-07-29 使用者要求先整理 compact 後記憶。`AGENTS.md` 已重寫為單一工作
入口，整合最終目標、雙 repo 邊界、SDD／證據標籤、原始資料與工具、DOS／
PC-98 視覺職責、倚天 16×15、長文／攻略三層、Git／驗證門檻、目前 milestone
與 compact 恢復清單。舊的「暫停」及「每小輪 commit」已移除。`CLAUDE.md`
縮成原始需求／資料索引並明確指向 `AGENTS.md`，避免兩份規則漂移。專案專屬
資訊留在 repo，沒有寫入全域 `~/.codex`。

2026-07-29 第 349 輪接續主線：真實 `ECL4/GEO4` probe 證實散塔林堡內城
`(10,11,N)` terrain `0x04` 的奧莉芙事件。玩家可閱讀繁中手札 50、選擇跟隨，
由 block `0x20` 穿牆切到幽暗神殿 `0x21`，再閱讀繁中手札 51，最後於
`GEO4/0x21 (10,6,N)` 恢復地城操作。人物、地圖、事件翻譯與四頁手札均留在
CoAB JSON game pack；State 只新增通用規則：若 title pack 明確宣告 NEWECL
目的地的 first-person map，就以該定義修正可能殘留來源 block 的 LOAD FILES
結果。下一個最早主線缺口是神殿內救出迪姆斯沃特。

2026-07-29 第 350 輪接續神殿主線：真實 `ECL4/GEO4 0x21` terrain sweep
定位 `(6,13)` 奧莉芙南方牢房提示與 `(2,14)` terrain `0x85`／PICTURE 35
迪姆斯沃特事件。完整手札 12 依附帶 Adventure Journal 譯為六頁，保存五道
枷印、五方勢力與三件終局神器說明；接受同行後回到同一神殿、重訪不重播。
ECL 沒有 `ADD NPC`，fighter roster 不變，故知識庫新增 story escort 與
persistent party NPC 的區別。下一個主線缺口是帶著迪姆斯沃特探索神殿，
觸發兜帽女子與離場路徑。

2026-07-29 第 351 輪延伸至眼魔洞穴：迪姆斯沃特 escort flag 成立後，
`GEO4/0x21 (4,12)` terrain `0x86` 觸發 PICTURE 39 兜帽女子。選 YES 後
真實 session 依序通過弗佐爾闖入、ECL `0x21→0x22`、德克薩姆與牛頭人、
手札 30、迪姆斯沃特指出洛山達護符、五項 encounter menu、手札 7、德克薩姆
殺死弗佐爾、第四枷印消退、護符離場與兩派混戰，最後在
`GEO4/0x25 (4,5,N)` 恢復地城。CoAB JSON 以 `script_block=0x22`、
`geometry_block=0x25` 新增洞穴 exact map definition，
全段繁中與四頁手札；下一個缺口是洞穴探索、德克薩姆／梅杜莎／牛頭人決戰
及取得洛山達護符。

2026-07-29 第 352 輪完成眼魔洞穴雙戰：真實 GEO4 `0x25 (15,1,N)` terrain
`0x90` 觸發 PICTURE 49，兜帽女子揭露為梅杜莎；第一戰精確建立 MON4
`0x27×1／0x28×1／0x29×10`。勝利後先處理四件普通戰利品，再從德克薩姆
遺骸取回洛山達護符；raw continuation 明確輸出散提爾堡部隊攻擊，第二戰為
`0x20×11／0x21×4／0x22×3／0x48×1`，勝利後回到 ECL `0x22`／GEO `0x25`。
先前單英雄 probe 的「戰鬥失敗」已證實是兩批敵方合法先攻造成，不是 DAMAGE
重複或 HP sync bug。研究 probe 已換成 focused real-image regression，戰後
localization 修正為走 JSON pack。另依使用者要求，Gold Box graphics 知識庫
新增 gameplay video temporal-oracle 規則與 DOS／C64／Amiga／PC-98 來源；
弓箭、投射物、法術與死亡效果須逐幀驗證，不能以數值正確代替畫面完成。

2026-07-29 第 353 輪完成眼魔洞窟離場。ECL4 block `0x22` entry 0 證實
離場 gate 是 `(C04F&0x3F)==0x13 && 7ED5==1`；runtime 出口為
GEO4/`0x25 (6,3)` terrain `0x93`，攻略 `(11,15)` 屬不同座標系。正式
regression 從雙戰勝利延伸，依序驗證 PICTURE 42、奧莉芙／迪姆斯沃特告別、
「Gharri」騎士與紫衣女子、`4CE2=1`、`7F12=1`、DESTROY ITEMS
`0x61/0x60` 的原 ECL 路徑，以及 `NEWECL 0x51` 暗影谷世界選單。繁中放在
CoAB JSON，不寫入共用引擎。

同輪完成戰鬥視覺唯讀 audit：近戰、弓箭與 Magic Missile 的規則已有部分
實作，但 renderer 缺 attack pose write、projectile、caster/impact、一般受擊；
`SoundMissile` 尚未送出，敵方回合仍在一次 update 同步結算。spec 354 將下一個
戰鬥重大 milestone 定義為 renderer-neutral Combat Action Timeline，先完成
melee、bow、Magic Missile、death 四條可錄影的端到端視聽路徑。

2026-07-29 第 354 輪第一階段建立 `combat.VisualEvent` 時間軸：
windup／travel／impact／commit／death／handoff 有 deterministic duration；
Ebiten opt-in 後鎖住輸入，State 每次只執行一個敵方 action，播放完成才推進
turn。近戰接入 attack icon 與 impact，弓箭開始送 `SoundMissile` 並畫 travel，
Magic Missile 保存飛彈數與路徑，致死 target 在畫面 commit 前維持原格，
死亡 phase 使用 COMSPR `0x8B/0x19`。新增 core/state/frontend geometry tests，
以及 `-combat-visual-demo melee|bow|magic|kill` 四張 640×480 Xvfb frame。
箭與法術 projectile 仍明確標為 fallback；spec 354 保持 IN PROGRESS，待
DOSBox／影片時間碼與原始 projectile block 定位後才能宣稱 fidelity 完成。
deterministic screenshot 已改為凍結指定 timeline elapsed；不可再直接用
`time.Since(start)` 截 phase，否則 Ebiten／Xvfb 啟動延遲會吞掉箭矢 travel。

同輪第二階段由 COMSPR raw blocks、reference consumer 與 DOS 公開影片
交叉定位 projectile。弓箭八方向使用 `0x00/01/02`、`0x80/81/82` 與 flip；
共同 spell missile 使用 `0x05/0x85` 四格，Magic damage impact 使用
`0x0A/0x8A` 四格。renderer 的線段／方塊 fallback 已移除，新增
`magic-impact` deterministic screenshot。影片 `wwYsij1wDC4`
`00:42:25.20–25.40` 顯示 Stinking Cloud 的同形青色 projectile，
`25.50` 清除後約 `25.60` 才出現文字與雲格。這證明 travel→effect ordering，
但弓箭與各距離 wall-clock cadence 尚未完成，spec 354 仍 IN PROGRESS。

同輪第三階段把 COMSPR 對照移出 frontend。獨立 engine commit `511ef40`
新增 `combat_visuals` schema、八方向／逐格驗證、trigger/phase 查詢；
CoAB JSON 宣告 arrow、Magic Missile travel 與 magic impact 的 source
block、flip、scale、原始 delay。三張 deterministic 戰鬥圖已由 Docker／
Xvfb 重拍且完整遊戲／核心測試通過。下一步仍以 DOS／網路實機影片建立
弓箭逐距離 timing，以及 Fireball、Lightning Bolt、Stinking Cloud／
Cloudkill 等不同法術的 travel、area effect、impact oracle。

同輪第四階段依 DOS 影片 `00:36:15.40–17.00` 完成 Fireball vertical
slice。`VisualEvent` 新增向後相容的 ordered multi-impact contract，一次
travel 後可逐名 impact／commit／optional death；前端會保留尚未輪到死亡的
目標，聲音也按 impact index 發送。反組譯 `sub_5F782`、`SpellEntry(0x2F)`、
`DoSpellCastingWork`、`RollSavingThrow` 證實 radius 2、敵我皆中、一次
caster-level d6 與每名 Spell save 半傷。正常玩家路徑現可由 memorized
`0x2F` slot 按 F、方向鍵指定 32×16 任意中心、Enter 消耗 slot 並施放。
三張 640×480 倚天／原版 COMSPR screenshot 顯示 travel、impact[0]、
impact[1]；deterministic screenshot 延後到第三個 Draw 讀回，避免 offscreen
battlefield 半幀。尚未完成 terrain-aware wall reachability、同距離原
combatant-array／direction tie order 與原版 wall-clock timing，spec 354
保持 IN PROGRESS。

2026-07-29 第 355 輪完成 Lightning Bolt 玩家 vertical slice。公開 DOS/EGA
影片 `wwYsij1wDC4` 的 `07:40:22.50–25.60` 清楚顯示 COMSPR `06/86`
青白電弧抵達 dark elf mage、`0A/8A` 紅白 damage impact／電擊傷害／
saving throw，再由命中格繼續前進；五張原版關鍵幀已保存。reference
`SpellLightningBolt`／`sub_5FA44`／`SteppingPath` 交叉證實全域 ID `0x33`、
一次 caster-level d6、逐目標 Spell save 半傷、正交／對角 2／3 weighted
step、玩家 budget 14、地城 `move_cost=0xFF` 首次牆面反向、近施法者首次
反彈 8-step penalty、large footprint 連續格不重傷及離開後重入可再命中。

共用 combat 時間軸新增作品中立 `VisualPathSegment`／
`VisualSegmentTravel`，可在 primary travel、ordered impact 與 continuation／
reflection segment 間交錯；line spell 規則由 `LineTerrain` 注入 valid／
reflect cell，不知道 CoAB DUNGCOM 或 spell ID。獨立 engine 的
`combat_visuals` schema 新增 `line` phase；CoAB JSON 宣告
lightning_bolt travel `05/85`、line `06/86`、impact `0A/8A`。正常玩家必須
真的 memorized `0x33`，戰鬥按 L、用 32×16 tile cursor 指定方向、Enter
才消耗 slot；失敗會回復。`sound_8` intent 被保留，但 recovered PC resource
無 WAV，不虛構音檔；damage impact 使用原 `sound_3`。三張 640×480 倚天／
原版素材 remake checkpoint 顯示 target impact、continued line、wall
reflection。牆角／多次反彈 DOS runtime 畫面、敵方 throws-lightning、
原版逐距離 wall-clock timing 尚未完成；下一個戰鬥效果是
Stinking Cloud／Cloudkill persistent area。

2026-07-29 第 356 輪完成 Stinking Cloud 持續區域 vertical slice。DOS
影片 `wwYsij1wDC4` `00:42:25.20–27.00` 補得 projectile 清除、建立文字、
四格綠白雲與下一段 action 仍持續的三張關鍵幀。reference
`SpellStinkingCloud`／`in_poison_cloud`／`StinkingCloudAffect` 交叉證實
spell ID `0x22`、target-anchored `{8,2,3,4}` 2×2、passable-cell filter、
Poison save、成功 cough 一回合、失敗 `d4+1` helpless、caster-level
duration 與重疊雲先還原再重畫的 instance semantics。Clue Book 另支持
「持續數回合、導引敵人、使怪物 helpless」。

共用 combat core 新增作品中立 `PersistentArea`、area terrain callback、
stable footprint dedup、action-counted coughing／helpless、overlap query
與 round expiry；`VisualEvent.PersistentAreaID` 讓 renderer 在 travel
期間只隱藏新雲，不誤隱藏舊的重疊區。獨立 engine `combat_visuals` 新增
`area` phase；CoAB JSON 宣告 generic `05/85` travel 與全域
BackgroundTiles `0x1E→graphic 0x26→RANDCOM item 4` 綠白雲 raster。正常
玩家必須 memorized `0x22`，按 N、以 32×16 tile cursor 選 2×2 西北角、
Enter 才消耗；四格全阻擋會 rollback。兩張 640×480 倚天／原版素材畫面已
保存。仍缺移動入雲逐步觸發、coughing AC 內部值換算、Cloudkill 3×3／
HD 4–6 即死與 exact DOS timing。

2026-07-29 第 357 輪完成 Cloudkill 3×3 持續毒雲 vertical slice。
`SpellCloudKill`、`in_poison_cloud`、`CloudKillAffect`、`CanMove` 與
`BattleRoundChecks` 交叉證實 spell ID `0x5B`、target 加八方向、passable
cell filter、caster-level instance，以及 HD 0–4 自動死亡、HD 5 的 `-4`
Poison save、HD 6 無修正、HD 7+ 無效果。低於 7 HD 的未受保護角色不能
主動踏入 Cloudkill；每回合重複判定則留待完整 affect 免疫接線。

party offset `0xE5` 與 MON*CHA offset `0xE5` 的 HitDice 已投影到共用
combat Fighter。CoAB JSON 宣告 generic `05/85` travel 與
`Tile_CloudKill 0x1C→graphic 0x24→RANDCOM item 2` 藍白雲；持續區 renderer
改由 JSON `source_file/block/scale` 選 atlas，連 Stinking Cloud 的 item 4
硬編碼也一併移除。正常玩家必須 memorized `0x5B`，按 K、方向鍵指定中心、
Enter 才消耗。Docker/Xvfb 640×480 checkpoint
`combat-timeline-cloudkill-persistent.png` 已保存。PC-98 screenshot 目前只
支持 layout；DOS 動態時間碼、protect magic／未命名 affect 免疫及每回合
重複毒雲死亡時序仍未完成。

2026-07-29 第 358 輪開始 PC-98 音樂來源研究。使用者提供兩張
`VFD1.00` FDD；新增唯讀 `internal/pc98vfd` 與 `cmd/pc98-vfd-audit`，
可重現 77×2×8×1024 geometry、媒體 SHA-256 與缺失 CHRN。Disk 1 的
`3/0/8` 對應 FAT12 cluster 46、`MSCDRV.EXE 0x4000..0x43FF`；`42/0/2`
對應 `CED3.DAX 0x10000..0x103FF`。Disk 2 尾端另缺 24 sectors。原始 FDD
已加入 `.gitignore`，不可提交。

Docker 內建置 NP2kai `0.86.0.22 e2dc904`，以可拋棄的可寫 2HD 研究副本及
Reset 成功進入 MEGDOS 0.25／`shell=loader.com`；唯讀副本會黑畫面，原始
媒體仍未修改。由於 music driver 正好遺失 1 KiB，loader／driver 後停止，
尚未進入標題。IDA 與 bytes 已證實 `MSCDRV.EXE` 是 `INT D2h` TSR，使用
YM2203 ports `0x188/0x18A` 與 timer playback；公開 register-log pack
另列出 Title、Character Creation、Town、Thieves Guild、Combat、
Dungeon 1、Wilderness、Village、Dungeon 2、City、Dungeon 3、Ending
12 首，但 scene→track number 尚未證實。完整限制與 READY gate 已寫入
spec 358；下一步優先找回同版 driver 缺失 sector，再做 caller／runtime
YM trace，不以補零結果冒充原版。

2026-07-29 第 359 輪確認 PC-98 開機保護與 loader 音訊鏈。MAME 官方
`pc98.xml` 的 `azurebnd` 提供 Disk A／B 1,265,664-byte FDI CRC32／SHA-1
身分，可作第二份合法 dump oracle。原 `LOADER.COM` SHA-256
`290e1031aea90a76b644f84175556ab0eba85897bc3204a181fcac2f339b18f3`；
三次 `INT 21h/AH=4Bh` exact 順序為 `setup.exe`、`mscdrv.exe`、
`game.exe`，第二段位於 file offset `0x3E..0x6D`。

同條件 NP2kai 對照顯示：保留 absent sectors 的未修改 D88 可進 MEGDOS；
補 sector、縮短 FAT entry、改 loader 或在開機過程過早覆寫副本，均在
MEGDOS banner 前停止。這支持 early integrity／copy-protection
`hypothesis`，尚未定位 routine。`GAME.EXE`／`GAME.OVR` raw bytes 無裸露
`CD D2`；因本作使用 VROOMM `INT 3Fh` overlays，不能推論無音訊呼叫。
下一步改做 emulator memory／interrupt instrumentation 或 VROOMM 解包。
一次性 no-op TSR／磁碟改寫 probe 已刪除，未提交、未修改使用者 VFD。

NP2kai D88 backend trace 隨後抓到 `C3/H0/R8/N3` baseline 共四次 read。
CPU soft-interrupt probe 在這之前尚未看到 `INT 21h/AH=4Bh`，所以不能把它們
誤標成 loader 的 EXEC。只讓首讀 not-found、第二讀起合成零 sector 時，
read 降為兩次，但畫面停在 MEGDOS banner、沒有 `loader.com` 或 EXEC trace。
這排除「永久零填」及「第一次缺失、重試回零」兩個簡化模型；可提交 trace
在 `docs/reference/original-pc98/vfd-runtime-trace.md`。

同輪依使用者提醒，唯讀回查 `~/.claude` 的 PC-98／retro 抽取記憶。採納
「先盤點既有解碼器與原生無損截圖，再決定是否硬解格式」的工作順序，以及
商業媒體／抽取素材不上公開 repo 的界線；但既有記憶沒有本作
`VFD1.00` absent-sector 的可直接套用解法，故四次 sector read trace 與
第二份合法 dump 仍是本作判讀的必要證據。

2026-07-29 第 360 輪建立 PC-98 Borland symbol／TPOV 可重現工具鏈。
`GAME.OVR` 以 `TPOV` 開頭；`cmd/pc98-ovr-audit` 由 resident
`GAME.EXE` control records 重建 36 段連續 code＋relocation chain，且只掃
code 後確認 literal `INT D2h` 為零。`GAME.EXE` MZ load image 截止
`0x144B0`，其後是 `0x52FB`／version `0x0208` legacy debug data。

`internal/borlanddebug`／`cmd/pc98-symbol-audit` 已以 9-byte record、
ASCIIZ name pool、count 與 EOF 邊界解析 1,725 symbols／2,305 names／
53 modules。錯用 10-byte stride 只會每九筆偶然重新對齊，已由 regression
排除。精確 symbols 為 `SOUNDFX 0893:0000`、`INITSOUND 0893:010D`、
`MSCPLAY 0893:0114`、`MSCSTOP 0893:015E`、`BGMPLAY 0893:0177`。

`MSCPLAY` 接受 1-based byte、減一、抑制同曲重播，再經通用 IVT `7Eh`
far-call wrapper 送出；`BGMPLAY` 已還原 area-code switch，會選
`3/4/5/6/8/9/12`。目前只可稱 internal area→selector exact；仍須將 area
寫入點對回 ECL／map scene，並證明 vector `7Eh` 與殘缺 `MSCDRV.EXE` 的
INT D2h 關係，才能建立正式 scene-role JSON。

同輪依使用者要求建立
`/home/anr2/my_skill/knowledge-base/retro-cht/reverse-engineer-borland-dos-pc98`，
保存 Borland Open Architecture ZIP／文字版、Turbo Pascal 7 原廠手冊、
來源 SHA-256／權利備註及可重用解析流程；skill 驗證通過並已 push
`wicanr2/my_skill` commit `b2a1497`。商業遊戲 binary／磁碟與工具沒有放入
skill。

2026-07-29 第 361 輪完成 PC-98 `CURRENTECL → BGMPLAY` 資料流與第一階段
音樂 contract。IDA／Borland symbol 校準證明 `CURRENTECL=0C29:BDF0`、
`WLDTWN=0C29:7F11`；TPOV overlay 2 `INTERPET` 的 loader writer 保存、
寫入及恢復 `CURRENTECL`，BGMPLAY far call bytes
`9A 77 01 93 08` 則只在 overlay 26 local `0x160` 出現一次。legacy debug
parser 另新增 53 筆 16-byte compiler module table，可列出 `GAME`、
`INTERPET`、`MENUS`、`COMBAT` 等 unit；overlay/module 對應目前只把有實例
支持的部分標為 nearby。

exact selector 表已寫入 `docs/spec/355-pc98-ecl-bgm-selector.md` 與 CoAB
game-pack JSON。獨立 engine commit `272e53c` 新增嚴格驗證的
`music_tracks`／`music_bindings` 及 exact-context fallback resolver，文件
維持繁體中文；CoAB State 會在初始 ECL 與 block transition 發出一次性
`MusicEvent`。`0x30` 不換曲、`0x52` 無分支；第 361 輪當時尚未解開
`0x50/0x51` 的 `WLDTWN` context。這個歷史缺口已由下方第 362 輪取代，
但曲名與實際播放仍未完成。

2026-07-29 第 362 輪完成 PC-98 Disk B／DAX codec 與 `WLDTWN` 語意證據。
Disk B 雖無標準 BPB，原 bytes 仍證明兩份相同 FAT12、32-byte root entries、
data start `0x2C00`；`ECL.DAX` 是 cluster 212–341、file offset `0x37400`、
size `0x20636`。先前 `0x1C00` data-start 假設已被推翻。

官方 IDA Pro 9.4 反組譯 `GETDATABLOCK 0723:0824`／IDA `0x17A54` 與
decoder `0x17DD5`，證明 PC-98 沿用 9-byte DAX index，但每個 block 有
raw／專用三分支 RLE flag。`internal/dax.ParsePC98` 對真實 24 個 ECL blocks
全部精確符合 decoded size，稽核 script 保存於
`scripts/ida/pc98_dax_codec_audit.py`。

overlay `INITECL`／`STOREVALUE`、decoded ECL 全 corpus 與 BGMPLAY consumer
三方證明 `WLDTWN` 特殊 ECL address 是 `0x5208`。block 0x50 在
`+0x063B/+0x0740` 寫 1/0，block 0x51 在 `+0x038B/+0x046D` 寫 1/0。
兩個 value=1 流程與 DOS 的 `YOU ARE IN ... WHAT PLACE WILL YOU VISIT?`
及 INN／STORE／BAR 完整對齊；value=0 回到主要區域／世界導航。因此 selector
5 是區域／戶外導航、selector 6 是城鎮設施選單。JSON 已移除 zero opaque
context，非零改名 `pc98-town-services-menu`；正常 DOS 玩家路徑的同 block
5↔6 cue 尚未完成，音樂與曲名仍不可宣稱完成。

2026-07-29 第 363 輪完成 PC-98 同一 ECL block 內的 selector 5↔6 音樂
情境切換。獨立 engine commit `a81b963` 新增嚴格驗證的 `music_cues`：
只將 `ECL block + signal + raw value` 映射成作品自訂 context，拒絕重複
cue、未知 signal 與沒有 binding 的 context，不讀城市名稱、選單文字或
作品旗標。CoAB JSON 將 blocks `0x50/0x51` 的 `PICTURE 0x50` 映射到
`pc98-town-services-menu`／selector 6，`PICTURE 0x79` 映射到
`pc98-world-navigation`，再由 context-free fallback 選 selector 5。

State 已在三個正式 `RunResult.PictureRequested` 入口解析 cue，並依 IDA
證實的 `MSCPLAY` 行為抑制相同 track 重播。真實
`TestFireKnifeLeaderStateVictoryReturnsToTilverton` 玩家路徑同時鎖定
阿沙本福德 block `0x50` 與希爾斯法 block `0x51`：入城 `5→6`、服務選單
內返回不重播、離城 `6→5`。這只完成選曲 intent；缺失 driver sector、
IVT `7Eh`／INT D2h 轉送、YM runtime trace、曲名與實際播放器仍未完成。

2026-07-29 第 364 輪完成 PC-98 `GAME.EXE → IVT 7Eh → MSCDRV → INT D2h`
bridge。依 AGENTS 規範優先使用
`/home/anr2/ida_94_official/dist` 的 IDA Pro 9.4 Docker 映像；
`scripts/ida/pc98_music_wrapper_audit.py` 還原 GAME 的 18-byte register
image trampoline 與 IRET 後 writeback，
`scripts/ida/pc98_mscdrv_bridge_audit.py` 則證明 MSCDRV file `0x02E3`
把 `0000:01F8`（vector `7Eh × 4`）直接設為自身 `CS:0080`。

GAME play buffer 是 `AX=0x00TT`，stop 是 `AX=0x01TT`；driver public
handler 因而以 `AH=0/AL=0-based track` 播放、`AH=1` 停止。driver
`sub_110CA` 保存舊 D2h vector，再把 D2h 設為固定低階服務
`CEE0:[0006]`；啟用前會檢查 `CEE0:0004 == 0x00D2`。目前不能把
`CEE0` 誤認為 MSCDRV 自身 segment，其 producer 仍待 runtime memory trace。

新增 `internal/pc98music` 與 `cmd/pc98-music-audit`，以 GAME／殘缺 MSCDRV
精確 SHA-256 驗證六個 raw-byte anchors。driver anchors 位於
`0x0280..0x1376`，均早於缺失 sector 對應的 file
`0x4000..0x43FF`，所以 bridge ABI 不受該缺口影響；這不代表音序列或整份
driver 已恢復。READY spec 364 與 Gold Box audio 知識庫已同步；剩餘缺口
是 `CEE0` provider、D2h `10h..1Fh` 命名、完整 driver sector、YM runtime
trace、曲名與播放器。

2026-07-29 第 365 輪完成 PC-9801 Sound BIOS／INT D2h ABI。NEC 官方
《PC-9800 Technical Databook BIOS 1992》「サウンド BIOS」章證明實體
`CEE00h` 是固定介面表、entry offset 位於 `CEE0:[0006]`，N88-BASIC
預設使用 D2h；因此第 364 輪的「CEE0 producer 未知」已被新證據取代。

依 AGENTS 規定在 Docker 內優先使用 IDA Pro 9.4，
`scripts/ida/pc98_sound_bios_audit.py` 列出本作 17 組 D2h client：
`INITIALIZE`、`CLEAR`、`READREG/WRITEREG`、`SETTOUCH`、`NOTE`、
`SETLENGTH`、`SETPARABLOCK`、`READPARA/WRITEPARA`、`ALLSTOP`、
`CONTPLAY`、`MODUON/OFF`、`SETINTCOND`、`HOLDSTATE`、`SETVOLUME`。
官方 `PLAY` 與 `SETTEMPO` 沒有出現在 wrapper 區，故不自行補寫。IDA 另
定位 direct YM2203 read/write helper；`internal/pc98music` 與
`cmd/pc98-music-audit` 現會逐 byte 驗證命令與兩個 `0x188/0x18A` helper。
所有新 anchors 都早於缺失 `0x4000..0x43FF`。READY spec 365 保存來源
URL、官方 PDF SHA-256、命令 register contract 與剩餘 runtime trace 邊界。

2026-07-29 第 366 輪完成 PC-98 十二首 track table 與 runtime import
邊界。IDA `sub_1021E` 證明 `DS:0330` 是 track pointer table，前十二筆
依序指向 `038A..064A`；`sub_10253` 證明每筆 64-byte descriptor 是
8-byte header 加七組 8-byte channel record，其中前兩個 word 是
sequence offset／length。

`internal/pc98music` 現會將 84 個 sequence 映回 driver file range並計算
SHA-256。真實 binary audit 證明十二首完整聯集是
`0x1B61..0x3C58`，完全早於缺失 `0x4000..0x4400`；所以缺 sector 沒有
切斷公開十二首 sequence，但整份 driver／後段工作區仍不可宣稱完整。
`ExtractTrackSequences` 只接受 exact driver SHA 與 selector 1–12，回傳
七聲道資料副本，不提交商業 bytes。

本機 Hoot `ponyca.xml`（SHA-256
`aae112a387d3e163273c191d8b0d826e0cd85b0a02fd4ee615d4ebab81e89b8d`）
的 CoAB entry 以 Shift-JIS 保存 0-based code→十二首曲名，與公開
register-log 清單交叉一致。獨立 engine `5363177` 新增跨 locale 驗證的
可選 `music_tracks.title_id`；CoAB JSON 現保存 selector 1–12、driver
index 0–11 及中英文曲名。仍缺 `sub_10410` stream interpreter、後段參數
consumer、YM register runtime trace 與實際播放器。

2026-07-30 第 367 輪完成 PC-98 `sub_10410` sequence bytecode 第一層。
IDA 與 raw bytes 證明 FM 0–2、PSG 3–5 的 note/rest width、`85/8A`、
family-only `90/91/92`、`A0–A4` 控制流與 `B0`；call／loop stack 各為
16 entries，overflow／underflow 依原版 no-op。

第七個 timing channel 不是一般聲部：只消耗 `85/8A` 參數，其他高 opcode
逐 byte 略過，而且會 read-through descriptor end。`internal/pc98music`
現分開執行 FM／PSG strict range audit 與 timing bounded read-through；
真實十二首 84 組 sequence 各通過 256 個 timed events，共 21,504 events。
READY spec 367 保存語法、邊界與尚未完成的 register-event／YM 播放層。

2026-07-30 第 368 輪完成 PC-98 正常配樂 OPN event runtime。指定 IDA Pro
9.4 batch script `pc98_music_event_audit.py` 重證 `sub_10410`、
direct write、tempo、SETVOLUME／SETPARABLOCK helpers，並定位 DS `0210h`
12-word FM F-number 與 DS `0228h` 71-word PSG period tables。

`internal/pc98music.TrackPlayback` 現從 exact-hash driver runtime import
七聲道，逐 timer tick 產生 register write／Sound BIOS intent，包含 FM
key on/off、PSG envelope／`91/92` modulation、`B0`、tempo 與 duration。
十二首各跑 4,096 ticks，共 68,291 events，auditor 保存各 selector 的 count
與 SHA-256。仍缺 fade／SFX 共存、NP2kai／Hoot 外部 register trace、
parameter block 展開及實際 YM2203 合成器；spec 368 已 READY。

2026-07-30 第 369 輪完成 PC-98 FM 音色 bank 的位址修正與完整性稽核。
指定 IDA Pro 9.4 證明 `sub_112B0` 以
`seg003:0542 + parameter_index×100` 呼叫 NEC Sound BIOS
`AH=16h/DL=0`；正確 file base 是 `0x45A2`，不是先前未驗證的
`dseg:0542`／`0x1A12`。IDA 可重現範圍恰有二十組 100-byte／50-WORD
參數，整體 SHA-256
`7bd538f4b80856aa67195f2ddcfe66226632d2f35e6e0d46d04bccfd2031d113`。

`internal/pc98music.FMParameterBlock` 現依 NEC 官方欄位解析並保留 raw
words；auditor 會在十二首各 4,096 ticks execution 中收集初始化及 opcode
`85h` 的全部音色索引。聯集為 `0..21,23..27,58`，內嵌 bank 缺
`20,21,23,24,25,26,27,58`；只有 selector 3、5、11、12 完整。
`BridgeReport` 現輸出 bank provenance/hash、逐曲索引及
`embedded_parameters_complete`。Hoot metadata 雖要求 `MSCD_98.COM`，
本機尚無合法可執行來源可證明它就是額外 bank producer，故維持 unknown，
不補零、不取模、不複製鄰近音色。DETUNE 真實值同時有 sign-extended
negative 與 `4..7`，在 Sound BIOS consumer／register trace 前不強行
正規化。READY spec 369 記錄證據與下一步。

2026-07-30 第 370 輪完成 PC-98 Hoot S98／YM2203 執行期音色驗證。先建立
Wine 8、Xvfb、Mesa software GL、PulseAudio null sink 的有界 Docker
工具鏈；Hoot 2023 需官方 x86 VC runtime，且原 Shift-JIS XML 在 Wine
MSXML 下必須轉成 UTF-8。指定 IDA Pro 9.4 native batch assembly 又證明
Hoot 的「device initialization failed」位於 display backend，Wine D3DX
trace 最後定位缺少 `skin/monotone/back.bmp`，補齊 skin 後可正常載入
PC-98 driver。所有工具、archive、runtime、S98、Wine prefix、IDA database
與 log 只留 `/tmp`，未提交 repository。

最初以相對 Down 擷取的十二檔因 Hoot 保存上次游標而有重複曲目，已作廢。
修正版每次重新搜尋遊戲、進曲目表、先送 Home，再走到絕對 row；十二首各用
獨立 `docker run --rm` 擷取約五秒 S98 v3。`ponyca.xml` row code 順序是
`00,01,02,08,06,0B,04,05,0A,07,03,09`，不能把 S98 流水號誤作 selector。

獨立 engine commit `1a6a252` 新增作品中立 `audio/s98`：嚴格解碼 S98 v3
header、device、write／wait／end，並抽取 YM2203 Sound BIOS tone burst
與 key-on signature。CoAB 新增 `cmd/pc98-s98-audit`，以 exact MSCDRV、
deterministic stream intent、二十組內嵌 signature 與十二首外部 trace
四方驗證。全部二十組證明 NEC rate／level 欄位要反相，logical operator
到 YM register slot 是 1,3,2,4，Sound BIOS burst 寫入順序是 4,2,3,1；
signed DETUNE 採 8-bit left shift，不應先截成 3-bit。

第 369 輪的「額外八組可聽音色／外部 bank producer」推論已被推翻。
`20,21,23,24,25,26,27,58` 全部只在 FM descriptor 初始化短暫載入，
第一個 stream `85h` 隨即改用內嵌 `0..19`，其後才首次 key-on。它們仍是
真實 register 副作用，不能刪除或取模；但十二首可聽音色均由二十組 bank
覆蓋，`PlaybackAudit` 現分開輸出所有呼叫與 audible indices。READY spec
370 保存 archive／trace hash、row→selector、轉換公式與完成邊界。

音樂仍未完成：total-level／operator-mask／algorithm carrier 公式、LFO、
fade／SFX 共存、完整 loop、YM2203 合成器、遊戲內播放及 save/resume
仍是下一階段。

2026-07-30 第 371 輪完成 PC-98 Sound BIOS total-level、algorithm carrier
與 operator-mask key-on。指定 `/home/anr2/ida_94_official/dist` 的 IDA Pro
9.4 在 Docker 內把 16 KiB `SOUND.ROM` 正確載為 CC000h 的 8086／16-bit
raw image；先前 IDA 預設 64-bit 所產生的 qword 分析已明確作廢。
公開 entry `CEE08h` 經 `CEF3Dh` dispatch 到 `SETPARABLOCK CF309h` 與
`SETVOLUME CF41Fh`。`CFB94h` 的欄位反相表 index-0 哨兵位於 ROM file
`3BEFh`／raw-load linear `CFBEFh`，欄位 1 從 `3BF0h` 開始，證明
AR／DR／SR 以 31、SL／RR 以 15、TL 以 127 相減；`CFD72h` 依
`00,08,04,0C` operator offset 重繪 tone。

`SETVOLUME` 依 algorithm 0–3／4／5–6／7 選 1／2／3／4 個 carrier，
以 logical operator `4→2→3→1` 順序逐一把 OutputLevel parameter 換成
track volume，每次均重繪完整 tone。十二首修正版 S98 的
12×3×2=72 組 descriptor／first-stream 序列全部符合
`TL=127-OutputLevel` 與上述更新順序。`CFC69h` 另證明 parameter 欄位 5
的低四位左移後形成 YM2203 register `28h` key-on；十二首 first-key
operator mask 稽核全數通過。

獨立 engine 新增 `audio/ym2203` algorithm/carrier 與 logical→physical
operator 拓樸，`audio/s98.YM2203KeyOn` 保存 operator mask。CoAB
`FMParameterBlock.YM2203LevelSequence` 與 `pc98-s98-audit` 現驗證 base TL、
carrier 次序及 first-key mask；`TrackPlayback` 不再固定寫 `F0h`，而由
active parameter mask 驅動。READY spec 371 與 native IDC 腳本已保存。
音樂仍缺 LFO、fade／SFX 共存、完整 loop、合成器、PCM mixer、遊戲內播放
及 save/resume，不能宣稱完成。

2026-07-30 第 372 輪完成 PC-98 Sound BIOS 軟體 LFO 靜態核心與動態
可觀測性稽核。指定 IDA Pro 9.4 證明 timer ISR `CF47Ah` 透過 register-held
near offsets `06C3h/0701h/07F3h` 跳到 `CF4C3h/CF501h/CF5F3h`；一般
auto-analysis 因 `jmp bx` 把後段誤留為 `db`。native IDC 現會手動建立三條
路徑，以及 `CF817h` pitch、`CF847h` TL、`CFAB9h` phase accumulator 與
六個 waveform 函式。

`SETPARABLOCK` 與 `MODUON` 都設定聲道 flag bit 7 並重算
`LFO_PITCH_COARSE × LFO_PITCH_DEPTH`；`MODUOFF` 只清 bit 7。MSCDRV 的
正常播放路徑沒有 MODUON/OFF caller，FM stream `90h` 仍只是 tempo +4，
不能與 LFO 混為一談。waveform 0..5 分別位於
`CFA49/CFA67/CFA10/CFA8A/CFA55/CFA24`，使用 signed 16-bit phase／step、
shared timer／noise 與原 8086 boundary quirks。pitch 依 sample、depth、
base F-number 做兩次 `/32767`；amplitude LFO 保留 byte-sized
`IMUL/IDIV` 與 wrapping。

獨立 engine commit `77683a3` 新增 `audio/pc98soundbios` 的 oscillator、
Pitch、TotalLevel，
`audio/s98` 新增一般 note／tone-load 排除後的 software-LFO update
extractor。CoAB `FMParameterBlock.YM2203Modulation` 映射 NEC parameter。
十二首 first-stream 共 18 個聲道使用非零 LFO 參數，但現有約五秒 S98
每首 `observed_lfo_pitch_updates=0`、`observed_lfo_level_updates=0`、
`dynamic_lfo_observed=false`。這只表示目前 Hoot capture 無動態觀測，
不能推論原作關閉 LFO。spec 372 已 READY；仍需長時間 Hoot 或
NP2kai／test harness 證明 cadence、sync delay，之後才能接 TrackPlayback
scheduler、合成器與遊戲內播放器。

2026-07-30 第 373 輪完成 PC-98 Sound BIOS LFO Timer B cadence、
sync delay 狀態機與 ROM 動態 harness。先將 Hoot selector 9 capture
延長至 45.01 秒；三個 FM 聲道都載入 first-stream parameter 3 且 key-on，
22,743 次 register writes 仍只有正常 note burst，獨立 pitch／TL LFO
更新皆為零。第 373 輪當時把它列為 Hoot `pc98dos` 可觀測性限制；第 374
輪後續 IDA／raw-byte 證據已推翻此原因，證明是 MSCDRV 繞過 Sound BIOS
timer ISR。

指定 IDA Pro 9.4 證明 `CF47Ah` 讀 YM status：bit 0 進 `CF501h` Timer A
note path，bit 1 進 `CF5F3h` Timer B LFO path。sync state 3 將
`uint8(sync-1)*4` 寫入 counter，再走共用 increment；state 1 每 tick
只遞減 counter 低 byte，成零時在同 tick reset oscillator 並輸出。

版本化 `scripts/research/pc98_sound_rom_lfo_harness.py` 使用 Unicorn 2.1.4
的 8086 mode，驗證 exact ROM／driver SHA 後直接執行 `CF309/CF41F/CF239/
CF3E7` 及 80 次 `CF5F3`，攔截 `0x188/0x18A`。parameter 3 的
`note_state=FFh`、flags `83h`；第一組 pitch／TL 在 tick 30，tick 30..80
共 51 組 `A4/A0` 與 204 筆 TL，final counter 51。

獨立 engine 新增 `audio/pc98soundbios.Modulator`，exact 保存 enable、
phase state、WORD counter、sync 0／1／2+、低 byte waiting decrement 與
shared phase；CoAB `FMParameterBlock.SoundBIOSModulationConfig` 只映射
waveform／sync／speed。engine 全測試、CoAB focused tests 與 ROM harness
均通過。第 373 輪當時把 TrackPlayback／Timer A note-state 接線列為缺口；
第 374 輪已取消這個錯誤工作項。Timer B→PCM sample clock、fade／SFX、
完整 loop、合成器與遊戲內播放器仍未完成。

2026-07-30 第 374 輪推翻「下一步把 Sound BIOS LFO 接進 CoAB
TrackPlayback」的整合假設。依 AGENTS 規定，在 Docker 內以指定
`/home/anr2/ida_94_official/dist` 的 IDA Pro 9.4 原生 IDC 分析 exact
`MSCDRV.EXE` 與 `SOUND.ROM`，再以 raw bytes 交叉驗證。

MSCDRV `10EE0h` 會用 DOS `INT 21h/AH=35h` 保存舊 sound IRQ vector，
再以 `AH=25h` 安裝自己的 `CS:0F54h`。初始化寫 `26h=BAh`、
`27h=02h`，接著寫 `27h=0Ah`。新 ISR 讀 `188h` status，只測 bit 1；
Timer B 到達時寫 `27h=20h`、呼叫 `10175h` TrackPlayback dispatch，
結尾再寫 `27h=0Ah`。它不測 Timer A，也不鏈回保存的 Sound BIOS ISR；
舊 vector 只在卸載 `10F37h` 恢復。

因此 `SOUND.ROM CF5F3h` scheduler 雖已由第 373 輪 ROM harness 動態
證實，CoAB 正常 BGM 並不執行它。selector 9 的 45.01 秒 Hoot S98 中
三個 parameter 3 聲道已 key-on、卻沒有獨立 LFO writes，現在可由 exact
driver control flow 解釋，不再歸因為 Hoot 可觀測性限制。faithful
`TrackPlayback` 不得額外產生 LFO pitch／TL；其 `Tick()` exact 語意是
一次 MSCDRV Timer B overflow。

READY spec 374 與兩支 native IDC 已保存。下一步是交叉驗證 PC-9801
YM2203 clock／prescaler 與 register `26h` 的 wall-clock 公式，建立
Timer B→PCM sample clock 的無漂移 bridge，再做 YM2203 合成器、fade、
SFX 共存、完整 loop 與遊戲內播放。

2026-07-30 第 375 輪完成 YM2203 Timer B 完整 count period 與 PCM
有理數排程。selector 9 的 45.01 秒 S98 header exact 指定
`3,993,600 Hz`；版本化 auditor 證明 `2Dh..2Fh` prescaler writes 為零，
`26h` 分布是 `00×2／BA×2／D3×1／E3×2`，`27h` 則有
`20×3558／0A×3561`，與第 374 輪 ISR acknowledge／restart 相符。

BSD 授權官方 ymfm commit
`81aec25ccbb98f4873a255f7551ac4dadac59b4a` 的 OPN 核心交叉證明
`16 × (256-data) × 12 operators × prescale`；YM2203 default prescale
為 6。因此 CoAB 完整 period 是 `1152 × (256-data)` chip clocks。

獨立 engine `audio/ym2203` 新增 `TimerBClockCycles` 與
`TimerBSampleAccumulator`；後者使用整數 numerator／remainder，1,000 個
period regression 不累積逐 tick rounding drift。CoAB adapter 只保存
3,993,600 Hz 與 prescale 6。仍未完成 `27h` reload 時 free-running
divide-by-16 phase、CPU I/O timing、YM2203 合成器、fade／SFX、loop 與
遊戲內播放；完整 period READY 不代表 cycle-perfect IRQ edge。

2026-07-30 第 376 輪完成 PC-98 YM2203 合成與第一條遊戲播放路徑。獨立
engine 固定 BSD-3-Clause `ymfm` commit
`81aec25ccbb98f4873a255f7551ac4dadac59b4a`，新增作品中立 register／native
PCM 封裝及整數有理數 phase 線性重取樣；cgo 關閉時明確回報 backend
不可用。

CoAB `YM2203EventRenderer` 依 spec 369–371 的 S98 exact order 展開
SETPARABLOCK／SETVOLUME，`TrackPCMStream` 依 Timer B period 推進
TrackPlayback、合成、重取樣並輸出 44.1 kHz signed 16-bit dual-mono。
Ebiten `internal/sound.Player` 可原子切換曲目，game-pack stable track ID
透過 `reference_selector` 選曲；啟動參數是
`-pc98-music-driver MSCDRV.EXE`。新增 `cmd/pc98-render-track` 供無 GUI
稽核。

exact driver selector 5 在 Docker 內兩次輸出十秒 WAV，SHA-256 均為
`fded75fe89d5e5af860e92e1541f83f14738c228fe7d792506c282c6bd5847c0`；
FFmpeg 證明 441,000 samples/channel、非靜音、peak -21.142021 dB、無
int16 clipping。WAV、driver 與 build cache 都只留本機。READY spec 376
保存完整邊界；fade／SFX arbitration、完整 loop、save/resume、analog
mixer gain 與 `27h` reload phase仍未完成。

2026-07-30 第 377 輪完成 PC-98 正常換曲延遲與 loop 邊界。依
`AGENTS.md` 在 Docker 內使用指定 IDA Pro 9.4 原生 IDC，交叉分析 exact
`GAME.EXE`／`MSCDRV.EXE`，再由 raw bytes 與 executable auditor 驗證。

`GAME MSCPLAY 18A44h` 對不同 selector 固定先呼叫 `MSCSTOP`，再把
`0x0320` 傳給 Borland `DELAY 19259h`，800ms 後才經 IVT `7Eh` play。
正常 public play 將 loop count 0 傳給 driver；初始化轉成 `0xFF`，FM
`105CCh`／PSG `10952h` end consumer 對它不遞減並重設七聲道 cursor，
所以正常十二首採無限循環。

MSCDRV 自己雖有 `0x28`／40 Timer-B-tick fade 與 FM0 SFX interpreter，
但 GAME wrapper 已先 stop，故正常換曲不進 fade。driver 的 SFX request
只找到清零與 consumer，沒有正常 GAME 非零 producer；GAME `SOUNDFX`
另走 Borland `SOUND`／`DELAY`。兩條路徑已明確分開，未擅自接入未知 FM
SFX。

CoAB 新增 `NewGameTrackPCMStream`，在每次新曲前輸出 35,280 frames／
141,120 bytes 靜音且不推進 TrackPlayback；`internal/sound.Player`
改用這條 wrapper。`pc98-render-track -game-transition` 以 exact driver
selector 5 產生三秒 WAV；FFmpeg 證明前 0.8s 共 35,280 samples/channel
全零，0.8–3.0s min/max `-418/2873` 且已發聲。READY spec 377、兩支
native IDC、兩個 GAME raw anchors 與 regression 已保存。仍缺 FM SFX
真實 caller、PC-speaker 完整 mapping、reload phase、save/resume 與
analog mixer gain。

正式 `cmd/azure-bonds-game -opening -pc98-music-driver` 也已在
Docker／Xvfb／ALSA null device 正常消耗 `MusicEvent`、建立新版 player、
輸出 deterministic screenshot 後退出；本機 PNG SHA-256 為
`8ab3e88ed74668788dfb3d37e5d6fdafbccf672de365fe827a933a5213c30fdd`。

2026-07-30 第 378 輪完成 PC-98 正常短音效程式與 caller 分布。指定
IDA Pro 9.4 以 8086／16-bit 分析 `GAME.EXE` 及八個 raw overlay，並由
`cmd/pc98-ovr-audit -soundfx` 對 TPOV code bytes 交叉驗證。42 個
`PUSH DS:[constant] → CALL FAR 0893:0000` caller 完全吻合；selector
分布涵蓋 1–13、15、255，其中 `MOVEMENT` 三處全為 10。

`SOUNDFX 18930h` 對 `0/1/13/14/15/255` 立即返回；`2/4/6/9` 走公式；
其餘使用 DS `1A3Ch` 的 20-WORD table。Borland `SOUND` 只保存 WORD，
`19D1Eh` 依其值忙等並向 PC-98 port `37h` 交替寫 `06h/07h`。IDA 匯出
640 bytes 後回搜 raw executable，唯一有效 base 是 file `E66Ch`，
整表 SHA-256 `65e9cf2cd93ae31edb497666415b54f936b98b8d89f87d475ebf3c4815c59ac4`。

新增 `internal/pc98sfx` exact-SHA importer、`cmd/pc98-sfx-audit`、overlay
direct-call typed auditor、focused tests、兩支 native IDC 與 READY
spec 378。商業表格、IDA DB、log 均未提交。下一步先追 selector 3–12
逐 caller 事件，再用 V30／8086 cycle 或 NP2kai audio trace 校準
port 37h pulse→PCM，最後接 Ebiten one-shot mixer；原 WORD 不可誤稱 Hz。

2026-07-30 第 379 輪完成 PC-98 SOUNDFX 具名語意與第一條可播放短音效
路徑。`GAME.EXE` Borland symbols exact 證明 `SOUNDHALT=255`、
`SOUNDOFF=0`、`SOUNDON=1`，以及 `CASTFX=2` 到 `CRASHFX=15`。
指定 IDA Pro 9.4 以 MZ load base `10000h` 在 `20AC8h..20AE8h` 逐 WORD
驗證；raw TPOV auditor 的 42 個 caller 總數不變，並命名
`REALMOVE／ANYUNDEAD／SHOWARROW／CASTSPELL／TWINKLE／SCAN`。

State 的 sound queue 已由數字 selector 改為平台中立語意。DOS adapter
保留既有 WAV selector；PC-98 adapter 使用 exact symbols，因此不再把
DOS arrow 2 誤套到 PC-98 `ARROWFX=12`，也不再把 DOS magic-hit 3 誤套到
PC-98 `SPELLHITFX=4`。

獨立 engine `fcf9b46` 新增作品中立 `audio/cyclepcm`，以整數 duty-cycle
積分 cycles＋level segments，窄 pulse 不會因落在 sample edge 間而消失。
CoAB `pc98sfx.RenderPCM` 當時使用的 NEC V30 `LOOP taken=13/final=5`
解讀已於第 380 輪作廢；selector、caller、作品中立事件與 engine
`audio/cyclepcm` 分層仍有效。更正後數值與 WAV 稽核如下節。

Ebiten `sound.Player` 可由 `-pc98-sfx-game GAME.EXE
-pc98-sfx-clock 8000000` 建立 one-shot players，並與 YM2203 music 共用
audio context。正式 `-opening` 已在 Docker／Xvfb／ALSA null device
載入 backend，640×480 screenshot hash 仍為
`8ab3e88ed74668788dfb3d37e5d6fdafbccf672de365fe827a933a5213c30fdd`。
READY spec 379 保存證據與邊界。下一步是 NP2kai／原機 port 37h edge
trace、不同 machine wait profile、analog mixer gain、save/resume 及
Timer B reload phase；不可把目前 8 MHz profile 寫成原機 cycle-perfect。

2026-07-30 第 380 輪以指定 IDA Pro 9.4、raw bytes、Unicorn 8086 動態
harness 與 NEC V30 官方表，校正 PC-98 software-speaker routine。exact
輸出順序是 `6,6,7,6,7,7`，位址依序為
`19D2A／19D3C／19D4B／19D3C／19D4B／19D57`；period 1000、pulse 2
共執行 4,045 條指令，兩個 busy loop 各 2,000 次。

官方 `BCWZ/LOOP` 執行時間是 branch taken 5、exit 13 clocks，不是第
379 輪讀反的 13／5。8 MHz、prefetched、no-wait profile 現分離第一次
gate-on `+98`、後續 gate-on `+30`、一般 gate-off `+56` 與最後 gate-off
`+28` cycles。ARROWFX 更正為 1,865 frames／0.042290s，兩次 hash 均為
`b9fc898253a380679e84c2026c84c9725a30302dbb15f7814a411664f2f50a5a`；
FIREBALLFX 為 736 frames／0.016689s，hash
`b8922db10390746d5bb5b06f28385c0ca779a17f35bc01fef29a3bbeea7c5be8`。
這仍是 timing-reconstructed；prefetch、pre-decode、I/O／memory wait、
caller gap 與原機類比路徑仍需後續 edge trace 校準。第 381 輪已證明
NP2kai 的 clock model 不適合承擔這項原機校準。

同輪新增繁體中文音訊知識庫
`docs/knowledge/pc98-gold-box-music-reconstruction.md`；共用 engine 另以
`docs/knowledge/golden-box-audio-architecture.md` 保存作品中立分層。

2026-07-30 第 381 輪完成 NP2kai i286c/V30 core 的 exact speaker edge
probe，並關閉「可否直接用 NP2kai clock 校準原機」的歧義。版本化
`pc98_speaker_probe_disk.py` 驗證本機 GAME SHA、抽取 file
`A6BEh..A6FDh` routine，保留 PC-98 IPL header；兩份 NP2kai patch 分別
記錄 `sysp_o37` 的 `CPU_CLOCK` 與提供 opt-in direct RAM injection，
repository 不提交商業 routine、probe image、binary 或 runtime log。

`period=1／pulse=1` 的 exact edges 是 clocks
`142,201,242,282`，values `6,6,7,7`；`period=1000／pulse=2` 是
`142,201,8234,16282,24315,32347`，values `6,6,7,6,7,7`。版本化
auditor 驗證 CS、IP、count、sequence 與 delta。

NP2kai source `i286c_mn.c:_loop` 明確使用 exit 4／taken 8，動態結果也
滿足 `8×(N-1)+4`：period 1000 busy 7,996、`6→7` 為 8,033；NEC V30
官方 execution busy 則是 5,008。故 NP2kai 本輪只證明控制流與 I/O
sequence，不能作原機 V30 wall-clock oracle，現行 NEC `5/13` profile
不變且仍標 timing-reconstructed。READY spec 381 保存證據與邊界；
下一步需原機 edge／錄音或經 microbenchmark 校準的 V30 emulator。

2026-07-30 第 382 輪釐清 YM2203 Timer B register `27h` 重載相位。
版本化 ymfm 的 `engine_mode_write()` 使用
`-(m_total_clocks & 15)`，而 `update_timer()` 會再乘
`12 operators × prescale`；故首次週期是
`(16×(256−B)−phase)×12×prescale`，不能只減 1–15 個 chip clocks。

獨立 engine 新增 `TimerBReloadClockCycles` 與
`TimerBSampleAccumulator.AdvanceReload`，驗證 phase `0..15` 並保留有理數
sample remainder。IDA 既有 ISR listing 則證明 CoAB 在
`27h=20h→0Ah` 間執行資料相依的七聲道 interpreter，因此不能硬寫固定 CPU
延遲。現行 instant-ISR PCM renderer 每個完整 period 都回到 phase 0；
這是模型邊界，仍需 CPU／OPN 共時 trace 才能完成原機 reload phase。

2026-07-30 第 383 輪修正 ECL session 第一次明確 `RunFrom` 的記憶體
生命週期。舊流程只設定 PC，未設定 runtime `Started`，因此
`SetMemoryValue` 預載的存檔、AREA 與劇情旗標不會匯入 runner，執行後還會
被空白 memory 覆蓋。現在首次 `RunFrom` 會先啟動 runtime；合成測試以
`0x9000=7` 驗證首輪讀取與保存。

真實 `ECL1.DAX` block `0x50` 測試以 `4C59=1／4C5A=1／4C5B=FF`
重現 Standing Stone：工作位址 `7F79` 計數為 3，灰袍人揭露自己是
Tyranthraxus，並要求隊伍到 Myth Drannor。SAVE writer 掃描另定位
ECL5 `0x33:+0FB6`、ECL4 `0x22:+0C8E`、ECL3
`0x11:+04E7`，並確認 `4C5B` 採非零判定。READY spec 383 保存證據與
邊界。下一步接通 JOURNEY ON 後的 ECL6/GEO6 block `0x40` 正常玩家路徑；
本輪未宣稱 Burial Glen、最終神殿或結局已完成。

2026-07-30 第 384 輪接通 Standing Stone→JOURNEY ON→MYTH DRANNOR→
WILDERNESS→AREA arrival→ENTER CITY→ECL6/GEO6 block `0x40` 的正常玩家
路徑。ECL1 尾端 14×4 adjacency table 已移入 JSON；共用 engine
`7ba2f8e` 提供中立 destination graph schema／validation。

測試同時釐清 destination work cell 必須複製原始 location byte，不能一律
減一；條件式縮短目的地表則保留 ECL 自有 work cells。新增測試不硬編碼
繁中劇情文字，而以 `message_id` 從受測 game pack 取得期望值。既有
Ashabenford→Standing Stone→Essembra→Hap 長回歸路徑與新 Burial Glen
入口路徑同時通過。完整 AREA travel、Burial Glen 事件與結局仍未完成。

2026-07-30 第 385 輪由 ECL6 block `0x40:+0014` 證明 Burial Glen exact
出生點是 `(2,15,E)`。修正帶文字的 `NEWECL` 入口：Continue 回到地城時也
會完成 pending map-register handoff，不再沿用 `(7,13,N)`。

GEO6 證明 `(2,15)→(3,15)→(3,14)` 兩步可通行，終點 terrain `01h` 會
觸發 PICTURE 72 精靈幽魂。原始 ECL 三分支 oracle 已驗證
`GREET／FLEE／ATTACK`；正常玩家路徑選 GREET 後，State 由 game-pack
stable ID 顯示繁中事件並解鎖 Journal 25。Journal 英文由公開文件轉錄與
Burial Glen 攻略交叉核對；本地 PDF 是純影像，OCR 本輪未可靠抽出該條。
紅網 terrain `82h`、`Krrkik`、蜘蛛／rakshasa 與後續章節仍未完成。

2026-07-30 第 386 輪解出 Burial Glen terrain `82h` 紅網。ECL6 block
`40h:+08A8h` 是 `INPUT STRING 8,[7F79h]`；輸入後沒有字串比較，而是
無條件顯示紅網更亮並回到四選單，所以 Journal 25 的 `Krrkik` 不可被
remake 假造成唯一答案選單。另一組 `+0425h` 輸入 12 字元並立即比較
`TYRANTHRAXUS`，交叉證明 string-memory destination。

VM 新增獨立、可續跑的字串 input request／offset；State 與 Ebiten 支援
Unicode、Backspace、Enter、ECL 最大長度及 uppercase literal vocabulary。
正常玩家 regression 從精靈幽魂沿 GEO
`(4,14)→(5,14)→(6,14)` 抵達紅網並完成輸入回返。real-image session
另鎖定 ENTER 的四蜘蛛→PICTURE 72→rakshasa 兩戰 continuation、
HACK 蜘蛛、RETREAT 與 `4CBFh=1`。新增 640×480 原版石框／16×15 倚天
截圖及 READY spec 386；兩場戰鬥完整 State 勝敗路徑、逐鍵 DOS fidelity、
strength side effect 與 Burial Glen 後續仍未完成。

2026-07-30 第 387 輪把紅網 ENTER 接到完整 State 勝利玩家路徑。回歸從
Standing Stone→Myth Drannor→Burial Glen 正常 GEO 行走進入 terrain
`82h`，以真實 `MON6CHA.DAX` block `42h` 的四隻 GIANT SPIDER 開戰；
一般 `CombatAct` 勝利後自動續跑同一 ECL session，顯示 PICTURE 72、
揭露 block `43h` RAKSHASA 並開始第二戰。第二次勝利顯示脫困文字、
寫入 `4CBFh=1`、返回地城，重踏事件不再觸發。

ECL6 block `40h:+07E3h..+0894h` trace 也關閉 Journal 25「獲得強大力量」
的歧義：ENTER 分支只寫 combat work `7F72h`、幽魂狀態 `4CBEh` 與完成
旗標 `4CBFh`，沒有角色力量／能力／effect destination。該敘述是羅剎妖
陷阱，remake 不得自行新增 Strength buff；`7F72h` 在外部 ABI 完成前仍保持
未知 combat work 命名。Docker／Xvfb／ALSA null device 另由正式
`-burial-red-web-battle` 保存 640×480 原版裂石框、MON6 CPIC 四蜘蛛戰鬥
checkpoint。READY spec 387 保存證據與邊界；戰敗、蜘蛛毒素、羅剎妖完整
能力、逐幀音效及 Burial Glen 後續仍未完成。

2026-07-30 第 388 輪沿紅網北方正常 GEO 路徑接通 terrain `04h` 的隨機
墳墓掠奪者。ECL6 exact 組成是 2 GIANT SPIDER、3 PHASE SPIDER、1
THRI-KREEN；一般 `CombatAct` 勝利後顯示三選單。動態 HORIZONTAL MENU
真正結束於 `+0CFCh`，LOOT／REBURY 分別跳到 `+0D0Bh`／`+0D28h`，關閉了
固定 arity tracer 對內嵌字串的假解碼。

`4CBAh` 由 ECL 初始化為偏移中立 `80h`；重新安葬 raw `+1`，搜刮 raw
`-1` 並發出一件珠寶、`ItemBlock=FFh` 的 TREASURE。State 現在會為只有
coin／gem／jewelry 的 request 開啟 treasure-service boundary，不再落入
零怪物 COMBAT；選項譯文由共用 engine `option_rules` 與 CoAB JSON
stable ID 驅動。正常玩家 regression 已連續驗證 REBURY 後再 LOOT，
珠寶 pool 為一且 ECL 回地城。正式 Docker／Xvfb 截圖
`burial-glen-grave-looters.png` 已保存；thri-kreen 小人目前受地形／佈陣
遮擋，placement／occlusion 仍是明確缺口。READY spec 388 保存全部 bytes、
分支、畫面與未完成邊界。

同輪曾短暫嘗試以 `seed+invocation serial` 避免隨機 terrain 永遠重複，
正式全套測試證明它會改變 Hap 與散提爾堡既有遭遇數量，故未提交該設計。
最終實作由 session shared `RuntimeState` 保存持續 PRNG stream；固定基底
seed 可重現，跨 ECL invocation 會自然取下一值。PRNG save/resume 尚未完成。

2026-07-30 第 389 輪由新增的作品中立 GEO 最短路徑 auditor，證明
Burial Glen 墳墓 `(6,12)` 可沿九步原始牆／門路徑抵達 `(13,14)`
terrain `03h`。正常玩家 regression 由 Standing Stone 一路完成前置事件後
逐格抵達黛米爾公主幽魂，而不是 direct-entry。

real-image ECL 測試鎖定 ACCEPT／REJECT／KILL／FLEE：認可度 `80h`
以上接受會寫 `4CBAh=85h／4CBBh=02h`；`7Fh` 接受只恢復中立；拒絕與
KILL 共用 `4CBAh=76h／4CBBh=FEh`；逃離不改值。`4CC0h=1` 使事件只出現
一次。全部提問、結果與選項已由 game-pack stable ID 資料化，正式
640×480 原版石框／PICTURE 72／16×15 倚天截圖已保存。

指定 IDA Pro 9.4 fresh load START.EXE 得到 21 個 `4CBAh／4CC0h`
instruction hits，但 START／raw GAME.OVR 都沒有 `4CBBh` literal consumer；
這不排除 interpreter 基底指標的間接存取。故本輪只保留 `4CBBh` raw
work cell，不把攻略所稱命中 `+2／-2` 寫進 runtime。READY spec 389
保存 writer、陰性 IDA 結果與下一步 DOS 相鄰值實驗邊界。

2026-07-30 第 390 輪關閉黛米爾 `4CBBh` 的間接 consumer。ECL6 block
`40h:+0249h` 及 blocks `42h／43h` 多個戰鬥入口均有 exact
`SAVE [4CBBh]→[7F71h]`。PC-98 Borland type table parser 新增舊式
8-byte type slots 與 5-byte members；`VARLISTPTR` 指向
`VARLISTTYPE size=0800h`，element type size=2，搭配
`VARLISTBASE=7C00h／END=7FFFh` 證明 `VARLIST+06E0／06E2` 分別是
work `7F70／7F71`，不是 PARTY offset。

指定 IDA Pro 9.4 對 EFFECTS overlay 23 的 `ATTEMPTTOHIT` 證明 side work
byte 經 `CBW` 後加入 attack roll；POSTCOM overlay 05 的
`DOPOSTCOMBAT` 則在戰後清零兩個 word。因此黛米爾 `02h／FEh` 是玩家側
每場戰鬥 `+2／-2`，不能當 unsigned 254，也不能永久改寫角色 AttackBonus。

獨立 engine 新增資料化 `combat_modifiers` schema、signed low-byte decoder
與繁中跨作品知識庫；CoAB JSON 宣告 `7F70` enemy／`7F71` party
attack-roll bindings。Battle 新增 side-scoped modifier；real-image ECL
正負投影、正常 Standing Stone→Burial Glen→黛米爾 ACCEPT 玩家路徑及
命中邊界均通過，並證明 Fighter 基礎值不變。READY spec 390 保存 hashes、
bytes、IDA 公式與未完成邊界；下一步另依使用者提供的原版實機截圖量測
人物 HEAD／BODY 組合與左上視窗版面。
第三百九十一輪延續使用者提供的 1014×759 DOS 實機擷取，先還原 320×200
邏輯座標，再確認畫中人物是 HEAD2 02＋BODY2 02、88×88 素材位於 `(28,24)`。
共用 engine 新增 game-pack `presentation.scene_character` native
anchor／clip；CoAB renderer 將 HEAD／BODY 與場景 cover 分流，加入 EGA
量化的黃色裂紋人物內框，並以 `-inn`、`-temple` 正常玩家路徑更新兩張
640×480 畫面。詳細證據見 READY spec 391；提交 hash 於本 milestone push
後補記。

第三百九十二輪從黛米爾 `(13,14)` 沿原始 GEO6 wrapped movement 正常步行
到 `(12,10)` terrain `93h` 與 `(14,8)` terrain `94h`。ECL6 SearchLocation
先做 `C04F & 7Fh`，selector `13h／14h` 分別跳到 payload
`+195Eh／+19B1h`；兩段使用真實 MON6CHA `41h` PHASE SPIDER，數量為
十隻／八隻，勝利後寫 `4CCD／4CCE=1`。

繁中敘事已移入 game-pack stable IDs；raw ECL session 與從 Standing Stone
開始的完整 State 玩家路徑均驗證 pause、combat、victory continuation 與
重訪不重播。沒有新增 Go 劇情分支。terrain `95h` 的六蜘蛛、骨堆
LOOT／REPLACE 選單、戰利品與好感副作用仍是下一個邊界；PHASE SPIDER
毒素／相位能力及原版動態演出也尚未完成。READY spec 392 保存 hashes、
bytes、路徑與完成範圍。

第三百九十三輪接續 Burial Glen terrain `95h`。GEO6 block `40h`
證明它位於 `(14,10)`，從上一輪 `94h` `(14,8)` 向南兩步可達。ECL6
selector `15h` 跳到 payload `+1A03h`：六隻 MON6CHA `41h`
PHASE SPIDER，combat work `7F82=10／4C01=6`，勝利後
`4CCF=1`，再顯示骨堆三選一選單。

三個 exact branch 是 `LOOT +1AA9h`、`REPLACE IN CRYPTS +1AC6h`、
`IGNORE +1ACFh`。LOOT 使黛米爾好感 `4CBA-1`，並執行
`TREASURE 0,0,0,0,0,1,0,FFh`；第六格是一顆 gem，`FFh` 是無 item
block，不是隨機裝備。REPLACE 使 `4CBA+1`，IGNORE 不改好感。
守衛文字、問題與三選項均已移入繁中 game-pack stable IDs；raw ECL 三分支
及從 Standing Stone 起的正常玩家路徑均驗證戰鬥、treasure service、
地城 continuation 與重訪 EXIT。詳細證據見 READY spec 393。

第三百九十四輪由 terrain `95h` `(14,10)` 沿最短合法 GEO 路徑推進到
`8Eh` `(10,7)`、`8Fh` `(9,9)` 與 `90h` `(8,9)`。ECL6 payload
`+1687／+16CE／+1713` 分別以 `4CC8／4CC9／4CCA` 控制三道螳螂人防線。
`8Eh` 是十二人、`8Fh` 是六人；`90h` 先打十二人，再檢查前兩旗標，
只在外圍守軍尚未清除時追加兩波各六人。

raw ECL regression 同時鎖定乾淨狀態的 `12→6→6` 與預設
`4CC8=4CC9=1` 時只剩首波。正常玩家路徑會穿過第二個共用 `4CC8` 的
terrain `8Eh` 而不重播，再在營地取得 exact
`TREASURE 0,0,0,2000,1500,4,6,81h`：9500 gold、4 gems、6 jewelry
與一件 deterministic random item。六段文字已移入繁中 game pack。
State／combat continuation 的通用 treasure service 也改為保留當次
`result.Text`，避免原 ECL 的「收起值錢物品」被通用提示覆蓋。
READY spec 394 保存 bytes、路徑、條件波次與未完成邊界。

第三百九十五輪把同一正常玩家路徑延伸到 GEO6 `(9,2)` terrain `91h`
與 `(10,1)` terrain `92h`。前者建立八隻 MON6CHA `42h` GIANT SPIDER，
使用 `7F82=7／4C00=8`，勝利後 `4CCB=1`。後者先顯示漏斗蛛網，再以
`4CBA < 80h` 判斷幽魂是否願意警告。

高好感會顯示蜘蛛守巢警告與 YES／NO；NO 直接 EXIT、不寫 `4CCC`，所以
可重訪。YES 或低好感直接分支會在戰前寫 `4CCC=1`，顯示蛛卵，建立四隻
GIANT SPIDER，並寫敵方 attack-roll work `7F70=2`、`7F82=0／4C00=4`。
raw session 已覆蓋高好感 NO／YES 與低好感無警告；Standing Stone 起始的
正常路徑則先選 NO、重踏再選 YES，驗證繁中選項、敵方命中修正與完成後
不重播。READY spec 395 保存 exact bytes、GEO 路徑與未完成邊界。

第三百九十六輪從蜘蛛巢 `(10,1)` 沿合法 GEO 路徑進入西側精靈王庭。
必經 terrain `08h` 的門口幽魂，YES 傳送至 `(4,2,S)`；terrain `89h`
四選項分別造成好感 `+1／-2／-2／不變`，RETREAT 不消耗 `4CC4`。
terrain `8Ah` 以 `4CBA >= 80h` 判斷友善，低於門檻會建立
MON6CHA `42h×6／41h×4／40h×4`；terrain `8Bh` 的友善王后給
12 gems、8 jewelry 與 ITEM6 block `41h` 六筆物品，敵對分支則先扣五點，
YES 給 4 gems、2 jewelry 與 block `40h`，NO 不給財寶，兩者都倒塔並
傳送 `(5,2,S)`。raw ECL 與 Standing Stone 起始正常玩家路徑均已通過，
所有繁中來自 stable ID／JSON。READY spec 396 是權威細節。

同輪保存使用者指定的 1014×759 DOS 角色資訊實機圖於
`docs/reference/user-provided/dos-character-info-layout-20260730.png`，
並更新 spec 391：左上是黃色裂紋 HEAD／BODY 專用舞台，右側 roster 與
下方長文各有獨立石框；不可把角色圖當一般 PIC cover，也不可把頭像塞入
身體或縮成 roster icon。

2026-07-30 第三百九十七輪證明精靈王庭不是 Burial Glen 出口。GEO6
block `40h` 從王后 `(1,3)` 到 terrain `05h` `(13,6)` 有 19 步合法路徑；
Standing Stone 起始正常玩家 regression 已逐格走完，途中共用隨機遭遇
會從正常 `FLEE` 選單返回原座標。

terrain `05h` 的紅羽戰士在 `WAIT` 後寫 `4CC2=1`、解鎖手札 33，並提供
`AGREE／REFUSE PAYMENT／DISAGREE`。使用者提供的 Adventurer’s Journal
PDF 第 10 頁（印刷頁 17–18）直接證明 Caemir 的祖父墓穴與魔法弓說詞；
英文、繁中、選項及事件提示都已移入 game-pack stable IDs。

答應同行並無視幽魂警告時，`PICTURE 43h` 顯示變形成羅剎妖。ECL 只發出
一條 `DAMAGE flags=2,dice=1d6+6,saveFlags=35h`；既有 random-target
consumer 證明它代表兩次帶命中判定的箭擊。之後載入 MON6CHA
`41h` PHASE SPIDER ×6 與 `49h` RAKSHASA ×1，勝利後接回精靈骸骨選單。
State 的 menu prompt 現會先查作品 JSON text rules，未命中才使用舊共用
fallback，避免把新作品提示硬寫入 Go。READY spec 397 保存完整證據與
限制；terrain `07h／0Ch`、區域出口、encounter `COMBAT／FLEE` 外部語意、
羅剎妖完整能力／戰利品／弓箭演出仍待後續。

2026-07-30 第三百九十八輪完成 Burial Glen 剩餘兩種 terrain 與東界
handoff。terrain `07h` `(14,3)` 每次都建立六隻 PHASE SPIDER 與一隻
RAKSHASA，沒有持久完成旗標，不可擅自改成一次性。terrain `0Ch`
`(4,8)` 初見寫 `4CC7=1`；`WAIT／PARLAY` 解鎖由使用者 Adventurer's
Journal PDF 印刷頁 23–24 核對的完整繁中手札 56，重訪直接離開。

正常 Standing Stone 起始玩家路徑已繞行 terrain `0Ch`，再完成紅羽戰士
戰鬥並由 `(15,6,E)` 呼叫真正出口 lifecycle。`PATH／WOODS／TURN BACK`
分別進入 ECL block `42h` `(0,12,E)`、`42h` `(0,6,E)` 或留在 `40h`。
State 現會對所有地城內 NEWECL 轉換投影原始 `C04B／C04C／C04D`，
沒有加入本作座標 hardcode。READY spec 398 保存證據；下一步接續
block `42h` 遺跡，最終神殿與結局仍未完成。

2026-07-30 第三百九十九輪把 Standing Stone 起始正常玩家路徑延伸至
ECL／GEO6 block `42h` terrain `01h`。初見寫 `4CD0=1`；`WAIT` 會解鎖
由使用者 Adventurer's Journal 與公開逐字稿交叉核對的完整繁中手札 5。
答應提爾雪雅後，第一戰是 HELL HOUND `44h`×5＋MARGOYLE `45h`×5。
貝爾哈其後率眾現身；選擇與提爾雪雅並肩時，第二戰載入 RAKSHASA
`43h`×1、地獄犬×6、石像鬼×6，寫 `4CD1=1`。

該分支另關閉一項通用 ECL 戰鬥缺陷：`LOAD CHARACTER 8` 指的是固定保留
八個玩家槽後的第一個怪物，不是 `8-len(activeParty)`。runtime／session
現會把跨 pause 的 team write 投影到第一隻 RAKSHASA 的 `PartyMask`；
State 把他建立成 QuickFight／TemporaryAlly，不再把怪物 MON*CHA 誤解析
成永久玩家 record。raw ECL、單人 adapter 與正常玩家路徑均已通過。
READY spec 399 保存證據；下一步接續倉庫及其餘 block `42h` terrain。

2026-07-30 第四百輪完成外圍遺跡倉庫 terrain `02h／83h`。GEO6 block
`42h` 證明從提爾雪雅 `(1,12)` 到入口 `(3,14)`、再向西進入 `(2,14)`
的合法路線。入口未清時可逃跑，或迎戰 MON6CHA `44h` HELL HOUND ×6
與 `45h` MARGOYLE ×6；勝利後 payload `+0C71h` 寫 `4CD1=1`。第 399 輪
結盟結果使用同一旗標，因此正常玩家路徑不會重打守衛。

倉庫普通踏入只顯示物資堆；只有玩家主動 `SEARCH` 提供 `7ECA=1` 時，
才發出 exact `TREASURE [0,0,0,2000,1500,8,8],82h`，也就是 9,500 gold、
8 gems、8 jewelry 與 ITEM6 block `82h` 的兩件裝備。財寶服務返回後
payload `+0D00h` 寫 `4CD2=1`，重搜不再取得物品。三段文字皆由
game-pack stable ID 提供繁中，地城財寶選單也會保留事件訊息。

同輪修正來源地圖 boundary work `7ED5` 的 transaction 生命週期：
`RunDungeonExitLifecycle` 開始追蹤 boundary attempt，選單分支完成
`NEWECL` 或 `EXIT` 後才清除，避免 Burial Glen 的問題滲入 block `42h`
每一步，又不在普通移動時誤清同 block 流程。raw ECL 與 Standing Stone
起始長玩家路徑均有回歸；READY spec 400 保存證據。下一步接續 terrain
`04h／05h` 逃亡男子及 block `42h` 其餘事件。

2026-07-30 第四百零一輪完成 block `42h` terrain `04h／05h／06h`
逃亡男子與東北藏寶。terrain `04h` 初見寫 `4CD3=1`；救援會建立
MON6CHA `44h` HELL HOUND ×6，戰勝後顯示 `HEAD6 40h + BODY6 40h`，
男子臨終說明東北廢墟藏寶，最後 `4CD4／4CD5=1`。若不救援，地獄犬會
將他撕碎，玩家仍可追擊同一批六犬或任其離開；terrain `05h` 的殘骸由
`4CD4` 控制，只顯示一次且沒有財寶。

terrain `06h` 普通踏入安靜；只有 `4CD5=1` 且主動 `SEARCH`
(`7ECA=1`) 才發出 exact
`TREASURE [0,0,1,0,0,0,0], ItemBlock=43h`。第三欄是一枚 electrum；
ITEM6 `43h` exact 三件物品是 Gauntlets +2、Girdle +1、Long Sword +5。
取得前 `4CD5` 清零，重搜不複製。State 的通用財寶投影新增
`moneyCopperRemainder`，先以 copper 精確累積再按 200 copper／GP
換算，因此單枚 electrum 的 100 copper 不會消失，兩枚可累成一 GP。
跨 `PRESS BUTTON` 的繁中藏寶敘事也會保存，空白 packed text 不再覆蓋。

raw ECL 已覆蓋救援、拒絕後追擊／離開、屍骸重訪與藏寶邊界；Standing
Stone 起始長玩家路徑從第 400 輪倉庫逐格走到 `(3,7)` 救援，再經合法
wrap 路線到 `(14,3)` 搜索並驗證三件真實裝備、幣值餘數與重搜。READY
spec 401 保存完整證據；下一步接續 terrain `07h／08h／09h`。

2026-07-31 第四百零二輪完成 block `42h` terrain `07h／08h／09h`。
terrain `07h` 以原始 `4C06=1` 控制無名者的一次性北方神殿警告，人物圖
是 HEAD `43h`／BODY `46h`。terrain `08h` 首次即寫 `4CD6=1`；拒絕救援
會讓地獄犬叼走受害者，救援則先殺死追犬，再遭 exact
`DAMAGE flags=0Ch,2d8,saveFlags=34h` 落石，最後迎戰五地獄犬、五石像鬼
與一羅剎妖。terrain `09h` 只有可重複的灌木、瓦礫與血跡敘事。

同輪修正 `CALL 2E10h` 的作品中立 State transaction：ECL 在同 block
先寫 `C04B／C04C／C04D=(11,10,2)` 再要求 redraw；VM 保存有 block／PC
的 SAVE 與 CALL trace，State 只在整次 session 未跨 block 且三者均於
CALL 前新寫入時投影為 `(11,10,S)`，不再只發 renderer dirty signal，也
不會覆蓋跨 `NEWECL` 的目的地出生點。Standing Stone 起始長玩家
路徑從上一輪 `(14,3)` 經合法 GEO 看完無名者、救援、落石與十一人戰，
再從真正傳送目的地向南一步到血跡灌木。七段繁中均來自 game-pack stable
IDs；raw ECL、CALL regression 與正常玩家路徑已通過。READY spec 402
保存完整證據；下一步接續 terrain `0Bh／0Ch／8Ah／8Dh`。

2026-07-31 第四百零三輪完成 block `42h` terrain
`0Bh／0Ch／8Ah／8Dh`。羅剎妖居所提供 COMBAT／WAIT／FLEE／PARLAY；
只有 HAUGHTY 交涉成功並解鎖由使用者 Adventurer's Journal PDF 第 13 頁
證實的完整手札 57，其他四種態度與直接戰鬥均建立 HELL HOUND ×5、
MARGOYLE ×5、RAKSHASA ×6。石像鬼門廊陷阱造成全隊 exact
`DAMAGE flags=C0h,3d10,saveFlags=1`，只寫 `C04B=10／C04D=0`，把隊伍
由 `(9,2)` 推到 `(10,2,N)`；裝死可突襲單一 RAKSHASA，不裝死則讓牠
撤退。賭局房拒絕逃跑會迎戰 MARGOYLE ×8、RAKSHASA ×6，勝利取得
1,200 GP、2,000 PP、15 gems、9 jewelry 與 ITEM6 `81h` 一件隨機物品。
下水道石像鬼無論放走或殺死都會露出柵口，經兩次確認後由 `NEWECL 43h`
正式進入昏暗廚房 `(15,15,N)`。

同輪把 `CALL 2E10h` State transaction 修正為欄位選擇性的部分座標提交：
同 session／同 block、CALL 前新寫入方向才構成位置 commit，X／Y 只投影
實際新寫欄位；Filani 對話中只把 `C04B／C04C` 當 scratch values 的真實
反例則不得移動玩家。Standing Stone 起始長玩家路徑以合法 GEO 依序走完
羅剎妖居所、門廊陷阱、賭局十四人戰與下水道入口，最後驗證 block `43h`
出生點。十七段繁中、raw ECL 全分支、手札解鎖、財寶、部分座標 regression
與正式玩家路徑均已通過；READY spec 403 保存完整證據。下一步接續
block `43h` 內部。

2026-07-31 第四百零四輪完成 block `43h` 出生點附近 terrain
`8Ah／8Bh／8Ch`。GEO exact 路線由 `(15,15)` 經 `(13,14)` 廚房、
`(12,12)` 辦公室抵達 `(10,12)` 豪華臥房；每一步都通過雙側牆面檢查。
三段繁中以 stable ID 資料化。臥房在詢問前寫 `4C04=1`，拒絕也會消耗；
同意則透過無怪物 COMBAT 財寶服務取得 5,000 GP、5,000 PP、12 gems、
15 jewelry，ItemBlock `FFh`，共 30,000 gold。

同輪由全 ECL6 writer 稽核發現原始跨 block 旗標碰撞：block
`40h:+0F52h` 與 `43h:+0D8Ch` 共用 `4C05`，block `42h:+0F70h` 與
`43h:+0DEAh` 共用 `4C06`。因此走遍前置支線後，辦公室與廚房 one-shot
原本就會靜默；乾淨 session 則能顯示文字。沒有 `LOAD FILES` 建立
block-local bank 的證據，故 remake 忠實保留 SAVGAM ECL 全域位址，不擅自
清零。raw ECL、YES／NO、重訪、exact 財寶與 Standing Stone 起始玩家路徑
均已回歸；READY spec 404 保存證據。下一步接續犬舍、活動雕像與私人
禮拜堂。

2026-07-31 第四百零五輪完成 block `43h` terrain `87h／88h／89h` 的
raw ECL 與繁中。犬舍於 `+0AEFh` 寫 `4C01=1`，保留原始空文字 PRESS
pause，之後建立 HELL HOUND `44h`×10；活動雕像於 `+0B74h` 寫
`4C02=1`，建立 MARGOYLE `45h`×10。私人禮拜堂於 `+0BD9h` 寫
`4C03=1`，老者說殺死隊伍更能傷害受班恩寵愛的提朗瑟克斯，隨後建立
HIGH PRIEST `48h`×1、PRIEST OF BANE `46h`×4。三場戰鬥、重訪與
MON6CHA 記錄均由 raw regression 鎖定。

Standing Stone 起始長玩家路徑已從臥房走到 `(9,13)`，完成禮拜堂五祭司
戰，再沿合法 GEO 抵達 `(7,10)`。下一步南行 `(7,11)` 是 terrain `83h`；
`82h–85h` 四格共同分派提朗瑟克斯／無名者最終儀式，前往西翼犬舍與活動
雕像的合法最短路線必經其中一格。因此本輪沒有 direct-entry 越過主線，
犬舍／雕像明確仍是 raw branch 完成、player path 未完成。READY spec 405
保存證據；下一步正面接通 terrain `83h` 最終儀式。

2026-08-01 第四百零六輪先回應 GUI fidelity 稽核。IDA Pro 9.4 依 PC-98
Borland symbols 定位 `SHOWHEAD 0176:0091`、`SHOWBODY 0176:00B3`、
`CLEAR3DVIEW 017C:018C`；caller 與 bytes 證明人物先畫 HEAD，再由 BODY
執行 `ADD AX,5` 後走共同 `SHOWPORTRAIT`，DOS overlay 同 offset 也有相同
核心 bytes。DOS runtime 另量出第一人稱／一般 PIC 是獨立灰色內框與
`(24,24)` 起算的 88×88 可見區，不能套用黃色 HEAD／BODY 人物舞台。

renderer 現以原版 raster 建立透明灰框 overlay，一般 PIC／3D 以 2×
nearest-neighbour 畫入 176×176 logical 區；640×480 石框則保留原生前
184 列、插入 40 列訊息側牆、把原版命令帶移到原生 `y=224`。旅店與正式
新遊戲入口第一人稱畫面已重擷取，正文／roster／命令列採倚天粗體 16×15。
READY spec 406 保存 IDA 位址、bytes、圖像雜湊、exact／reconstructed
邊界。這只關閉人物／3D／PIC 舞台及延伸訊息框；角色資訊全頁、所有 PIC、
戰鬥動態仍未完成。下一輪回到 terrain `83h` 最終儀式與 Journal 48。

2026-08-01 第四百零七輪完成 block `43h` terrain `82h–85h` 共用的
提朗瑟克斯／無名者最終儀式。ECL 首次即於 `+0537h` 寫 `4C00=1`，依序
經過十三段文字／pause 與六個 PICTURE：提朗瑟克斯利用枷印操縱隊伍、
演說並解鎖使用者 PDF 核對的完整繁中手札 48、迫使隊伍交出三神器；假扮
祭司的無名者把神器拋回，臨終默念羊皮紙上的解除密語，使控制力消退。
ECL 沒有發出 FIND／DESTROY／TREASURE，故 frontend 不得自行刪除神器。

儀式戰 exact 為 HIGH PRIEST `48h`×2、HELL HOUND `44h`×6、MARGOYLE
`45h`×6。`4CBD=0／1` 兩組 raw session 到戰鬥 boundary 的輸出相同，
只證明此段不因該值改怪，不能擴大到整章。Standing Stone 起始正常玩家
長路徑已從 `(7,10)` 南行觸發儀式、戰勝十四人、續跑回地城，再穿越靜默
`84h／85h`，沿合法 GEO 完成 `(4,9)` 十石像鬼與 `(1,9)` 十地獄犬，
確認 `4C02／4C01=1`；途中龍盔的二樓東北角提示也已資料化繁中。

READY spec 407、deterministic `-inner-ritual` PICTURE 47 截圖、raw regression
與正常玩家長回歸共同保存證據。下一步追 terrain `86h`、東北角、真正
提朗瑟克斯最終戰與結局；目前仍不可聲稱完整可通關。

PICTURE 47 抓圖同時揭露一般 PIC blit 曾把非零 bounds 的 Ebiten sub-image
誤當局部 `(0,0)`，造成事件圖右／下側被裁掉。renderer 現直接以 destination
全域原點做 2× nearest-neighbour blit，並以 scale／全域原點回歸鎖定；screenshot
另固定 PIC frame 0、停用音訊裝置，避免容器啟動時間與 ALSA 影響圖像證據。

同輪依使用者指定，唯讀評估 `/home/anr2/cht/大時代的故事/CLAUDE.md` 的
IDA 經驗，將已驗證且可泛化的非破壞性規則納入 AGENTS.md：保留原始名稱／
位址，語意只附加並帶推論等級；`.i64` xref 優先於 `.asm` grep；直接 xref
不涵蓋指標間接存取；headless IDC 必須寫檔驗證。來源仍在試驗期的自動
語意 dump 方法未升格為本專案硬規則。

2026-08-01 第四百零八輪證明 block `43h` terrain `86h` 不是終戰，而是
`(0,10)` 的共用 redraw。犬舍後沿十五步合法 GEO 路線可抵達 `(10,7)`
terrain `97h`，上樓 transaction 寫入 `(2,5,N)`；二樓再沿十步抵達
`(6,1)` terrain `9Ah`。終戰 raw ECL 三段台詞後 exact 載入 MARGOYLE
`45h`×28、TYRANTHRAXUS `47h`×1、HIGH PRIEST `48h`×8，戰勝後立即
呼叫 `PROGRAM 8`。

Standing Stone 起始累積長回歸已正常走完一樓、西翼、樓梯、二樓與 37 人
終戰，正式 scheduler 勝利後在同一 ECL session 進入勝利存檔選單。途中
盟友旗標使共用 minion encounter 的四組 count 為零；runtime 現只對有真實
dungeon ECL session 與 dungeon return 的零敵人 COMBAT 立即續跑，未知
synthetic case 仍失敗即關閉。`RuntimeState` 也保存跨 UI boundary 的 monster
setup，fresh entry 清理互動 registers 而保留 ECL memory／持續 PRNG。

READY spec 408 保存房間、樓梯、旗標、終戰與限制。直接視覺 checkpoint
因預覽隊伍先被擊敗，只得到戰敗結果頁，故已刪除且未放入 README。下一輪
需建立不影響戰鬥結果、可凍結時間軸的終戰 capture，補齊三類敵人的法術、
特殊能力、AI、音效與 DOS 動態演出；全新隊伍由開場至結局的單一通關仍未
完成。

2026-08-02 第四百零九輪把 remake JSON save 升至 version 6。先稽核發現
只保存 seed 不足：ECL current block、resume PC、shared work memory 與
input offsets 也都會共同決定下一分支。`BlockSession.Snapshot` 因此保存
mutable continuation；code window 只保存相對玩家自備 block 的 runtime
差異，不把完整商業 ECL bytes 寫入 JSON。

獨立 engine 新增 `randomstream`，以 seed＋底層 Source draw count 保持現行
`math/rand` 序列，Restore 設有 replay 上限。合成 PROGRAM boundary、State
SavePartyFile／LoadPartyFile 與真實 `ECL6.DAX` block `40h` terrain `04h`
均證明讀檔後下一批 random values、怪物生成與文字等同不中斷執行。舊 v1–v5
仍可載入，但不能冒稱恢復未曾保存的 RNG 位置。READY spec 409 與雙 repo
繁中知識庫保存證據；任意 UI／戰鬥 frame、音訊位置及 SSI 原版 RNG 仍未完成。

2026-08-02 第四百一十輪以非破壞性 IDA 副本追蹤 PC-98 `EFFPROCS`、
`LOADSAVE` 與 `EFFECTS` overlays。Borland `EFFECTREC` type size 是 9；
`LOADMONSTER 00DC:3AA8` 逐筆複製 9 bytes，將新 node `+5／+7` 清零並重建
linked list，但保持 byte `+4`。因此先前將 `MON6SPC` byte 4 的零值當成
`Active=false` 並靜默丟棄全部天生能力，是 remake adapter 的錯誤。

Battle 新增作品中立 `Innate` 來源標記；角色後天效果仍保留既有 Active
生命週期。提朗瑟克斯 `47h` 六筆真實 `MON6SPC` 現由 Standing Stone 起始
正常長路徑載入，其中 `18h` 偵測隱形會抵消目標隱形 AC bonus；完整路徑仍
戰勝並進入 `PROGRAM 8`。READY spec 410 與可重現
`scripts/ida/pc98_monster_affect_loader_audit.idc` 保存 hashes、位址、xref
及推論等級。下一步優先追 `4Fh／6Ah／70h／84h／87h` 中的閃電或魔法抗性；
HIGH PRIEST `09h／0Ah`、MARGOYLE `77h` 也仍未解。

2026-08-02 第四百一十一輪繼續追蹤 PC-98 EFFPROCS overlay 12。
raw bytes 與先前已保存的 IDA 報告證明 common routine 以
`base+(11-casterLevel)*5` 為門檻，在 current affect 存在或 Magic
damage flag 設定時擲 `1d100`；擲值不大於門檻就呼叫 local `001Bh`
`Protected(0)` 清除傷害。local `23F4h／2404h` 分別是 50／15 base
wrappers。`MON6SPC 6Ah → 15%` 由獨立 DOS 反編譯交叉支持；
由於 TPOV relocation 尚未完整套用，該 mapping 誠實標為
`strong inference`，不把 table raw addend 當直接 handler 位址。

Battle 現以作品中立 `MonsterMagicResistanceBase`與
`MagicResistanceChance` 套用此公式。Magic Missile 維持原順序：先擲完
所有傷害骰，再擲抗性 d100；成功時傷害歸零，但施放格、動畫與
continuation 仍進行。繁中 `combat_magic_resisted` 從正式 locale JSON
解析，測試不複製顯示字串。Standing Stone 起始長路徑證明真實
提朗瑟克斯帶有 operational `6Ah`，並仍可完成 `PROGRAM 8`。
READY spec 411 與 `docs/knowledge/gold-box-combat-effects.md` 是本輪權威。
下一步應完成 TPOV relocation typed decoder，或追蹤 `84h` 閃電施放與
`87h` 電擊保護；不可在未證明 damage boundary 前將魔法抗性全域套到
Fireball、Lightning Bolt 或雲霧。

2026-08-02 第四百一十二輪完成 PC-98 TPOV typed entry／fixup 解碼。
唯讀磁碟抽出的 `GAME.EXE`／`GAME.OVR` hashes 與 36 段 corpus 證明 control
固定頭 `20h`，entry 是 `CD 3F + handler-local u16 + flags`，code 後 relocation
是嚴格遞增的 `u16` fixup offsets。auditor 新增 `-resolve-stub`，並以 signature、
handler bound、fixup bound／排序失敗即關閉。

resident pointer 經 control segment 與 entry index exact 關閉提朗瑟克斯六筆
效果：`18h` 偵測隱形、`4Fh` 2d10 fire、`6Ah` 15% 魔法抗性、`70h` 防火、
`84h` Lightning Bolt、`87h` 防電。`84h` 另由 overlay 22 `SPELLS` slot writer
與 handler push spell `33h` 證實。這是靜態 handler／基本語意完成，不代表
runtime 已完成；下一步仍須追怪物閃電的頻率、目標、terrain reflection、動畫、
聲音，以及將防火／防電接到各自 pre-damage boundary。READY spec 412 是權威。

2026-08-02 第四百一十三輪把 effect `70h／87h` 從靜態 exact handler 推進到
可玩 runtime。Battle 新增作品中立 raw damage flags 與
`MonsterProtectedFromDamage`；Fireball 傳 Fire＋Magic，reflecting line 由
呼叫端傳 Electricity＋Magic。保護命中仍保存 visual impact、法術格及回合
transaction，但 damage 歸零、HP 不變並標記 `Protected`。

正常 memorized Fireball／Lightning Bolt tile-cursor 測試載入正式 zh-TW JSON，
驗證防護摘要 stable IDs；Standing Stone 起始長路徑則用真實 MON6
Tyranthraxus fighter 關閉兩種 damage boundary，之後原 37 人終戰仍進入
`PROGRAM 8`。現行 save draw 先於 protection 的 RNG 順序尚缺 DOS runtime
trace；`4Fh` 天生 fire、`84h` 怪物 Lightning Bolt AI／terrain／演出也仍未
完成。READY spec 413 是本輪權威。

2026-08-02 第四百一十四輪以非破壞性 IDA 副本與 raw bytes 關閉 effect
`4Fh` 的命中後 caller。PC-98 overlay 13 local `18A2h–18B7h` 先確認物理
傷害後目標仍在戰鬥，再把 `attackIndex+1` 連同攻擊者 far pointer 交給
overlay 23 `CHECKFX`；type 2、3 分支均 exact dispatch `4Fh`。overlay 12
local `19B3h` 則對攻擊者現有 target 擲 `2d10` Fire＋Magic，沒有 saving
throw，也不重新選目標。typed resolver 新增 handler-local reverse lookup，
`23:03FE` 可重現 entry 4／resident stub `0034h`，原始位址空間仍分離保存。

Battle 現在只在前兩個物理 attack slots 成功命中且目標存活時投影
operational `4Fh`；miss、物理擊殺、inactive effect 均不觸發。
`AttackEffectResult` 分開原始 effect ID、damage flags、`RolledDamage`、實際
damage 與 protection；`70h` 防火即使清除實際傷害，仍會先消耗兩顆 d10，
保持 PRNG continuation。單擊／多擊／防護繁中訊息全部由正式 locale stable
IDs 驅動。Standing Stone 起始長玩家路徑取得真實 MON6SPC
Tyranthraxus fighter，驗證 `4Fh` boundary 後原 37 人終戰仍完成
`PROGRAM 8`。原版火焰動畫／sound cue／wall-clock timing、轉移目標與自由
攻擊動態 trace、`6Ah` 對 4F 的時序，以及 `84h` 怪物閃電仍未完成。
READY spec 414 是本輪權威。

2026-08-02 第四百一十五輪以非破壞性 IDA 副本與 raw bytes 關閉 effect
`84h` 的怪物 action phase。overlay 9 在一般行動前以 type 14 呼叫
`CHECKFX`；overlay 22 handler exact 比較 `ROUND < 4`、dispatch spell
`33h`，先對初始目標格擲 `16d6` 並處理 Spell save，再獨立擲第二份
`16d6` 處理 range 10 的後續直線／反射，最後清除施法者 actions。
獨立 DOS 重寫只用來交叉確認 battle setup 先設 round 0、每輪 action 前遞增，
所以 remake 第 1–3 回合與 raw gate 對齊。

Battle 的作品中立 line profile 現可分開 initial／path damage pool；零值仍保留
玩家 Lightning Bolt 的單一等級 d6 共用行為。State 在 Magic Missile／物理
攻擊前排程 operational `84h`，正式 Ebiten frontend 注入 DUNGCOM／WILDCOM
terrain，並沿用資料化 Lightning timeline、effect `87h` 防電、聲音階段與
繁中 stable IDs。focused tests 與 Standing Stone 起始正常 GEO／ECL 長路徑
均通過；真實 MON6 提朗瑟克斯在 37 人終戰產生動態閃電事件，之後仍完成
`PROGRAM 8`。READY spec 415 保存 hashes、位址、推論等級與限制。原版
target range／LOS／tie order、終戰牆面逐幀 oracle、精確時間與 `6Ah` 對
effect 84 的時序仍未完成。

2026-08-02 第四百一十六輪以非破壞性 IDA 副本繼續追 effect `84h` 選敵。
Borland `PICKTARGET 00B8:3D7F` 在 overlay 13 exact 設定二十次上限，呼叫
`014A:00C0`；typed resolver 將它落到 overlay 24 handler `285Bh`，再由
overlay 32 footprint／combat-background consumer 建立鄰近敵方候選。
不可見候選會從表內移除後重抽。spell `33h` 的 PC-98 resident record 在
file `13114h`，fixed/per-level range bytes 是 `4／1`；effect line helper
另直接使用 range 10。

Battle 新增作品中立 `TargetSelectionOptions`，依 cardinal／diagonal `2／3`
加權距離、雙方 footprint、range 10 與 `LineTerrain` 牆面建立候選，保留
二十次 PRNG／可見性 callback。State 在一般 living-target selector 前先走
此流程；無 ranged target 時仍消耗 effect `84h` action，不改成物理攻擊。
focused core／State tests 已通過。READY spec 416 保存 exact／strong
inference 邊界；原 combatant-array 同距 tie order、PC-98 invisibility
runtime、無目標 `(0,0)` 動畫、終戰牆面逐幀與 `6Ah` 時序仍未完成。

2026-08-02 第四百一十七輪以 typed TPOV resolver 與非破壞性 IDA 副本關閉
PC-98 status visibility。effect table 的 `008B:00AC／017E` 分別解析為
overlay 12 handlers `06F9h／1713h`；前者只在觀察者缺少 operational `18h`
時設定 `DS:A035h` hidden 並使 attack roll `-4`，後者則無條件執行兩者。
`CHECKTARGET 00B8:11AF` 會清 hidden、分兩 phase dispatch 並暫時交換 action
target，最後只在 hidden 仍為零時回 true。

Battle 新增作品中立 `Fighter.VisibleTo(observer)`，物理 AC 改用同一差異化
投影；State effect `84h` ranged selector 正式注入 visibility callback。一般
怪物／effect `18h`／inactive／`19h`／`47h` 的 core 與 State 正反例，以及
Standing Stone 起始真實 MON6 提朗瑟克斯長路徑均納入回歸。spec 235 已標
SUPERSEDED，spec 410 將 `19h／47h` 一併略過的舊斷言也由 READY spec 417
訂正。blink、invisible-to-animals、完整 effect 生命週期、同距原始順序與
`(0,0)` 動畫仍未完成。

2026-08-02 第四百一十八輪以 typed TPOV resolver 與非破壞性 IDA 副本解析
effect `25h → overlay12:0BDBh`、`45h → 16C2h`。Blink 讀 target Action `+3`；
值 0 時設 `A035h=1` 並把 attack roll `A039h=FFh`。effect `45h` 只在 observer
Player `+11Ah=13h` 時作用，缺 `18h` 才 hidden，但 `-4` 永遠保留。

獨立 dragon-slayer effect `4Bh → 17B5h` 在同一 `+11Ah` 比較 `03h`，六章
MON*CHA corpus 的 `13h` records 則只有 WORG、FIGHTING DOG、MONKEY、OWL BEAR，
因此 parser／Fighter 正式投影 raw MonsterType 與 Animal `13h`。Battle 讓每個
scheduled fighter 的 action delay 維持非零，State 在 action 完成後集中清零；
不新增 PRNG draw，也未宣稱現行 initiative 數值已 exact。core visibility／AC／
natural-20 regression、Tilverton 犬舍真實 FIGHTING DOG 與 Standing Stone phase
spider 長路徑均通過。READY spec 418 保存地址、bytes、推論等級與未完成邊界。

2026-08-02 第四百一十九輪以非破壞性 IDA 副本、typed overlay resolver 與
primary bytes 關閉 PC-98 先攻排程。`DEXRABONUS 014A:1416` 解析為 overlay 24
local `1416h`；overlay 13 local `0000h` exact 寫入
`1d6 + DEX reaction adjustment`，先夾 1、依 `area.field_596` team bit 減 6，
最後把 signed `<0`／`>20` 設零。overlay 8 local `01FBh` 逐一走 TeamList
`+18Ah`，每節點包含 delay 0 都抽 `1d100`；最大 delay 優先、roll 解 tie、
完全 tie 後掃者勝，全零掃描最後回 null。

獨立 engine `4545f0b` 新增 `combat/initiative` 與繁中知識庫，已 push。
CoAB 改以 shared Player `+17h` Dexterity、建構時 TeamList order 與同一 RNG
串流排程，不再使用 d20／fighter ID tie-break；MON combat team offset 也由
錯誤 `+197h` 修正為 `+198h`，`+1A5h` raw byte 不再投影成 initiative bonus。
正式遠端 pseudo-version、Go `h1`、無網路 `GOWORK=off`、明確 Xvfb trap 的
31 套件 gate 全部通過。長玩家路徑另修正單人英雄 attrition 與 Alias
alive/dead fixture 前提，不再依舊排序碰運氣。`area.field_596` writers、玩家
`DELAY` 20→19 同輪重新入列、DOS 等價性及原版底層 PRNG 仍未完成；READY
spec 419 是權威。

2026-08-02 第四百二十輪重新稽核 overlay 08 完整人工 combat input。primary
bytes 證明頂層 `D` 呼叫 local `1028h` 子選單，第二層 `D` 在 local
`1173h–1187h` exact 寫 Action.delay=1 並結束當次人工操作；同一輪 selector
之後會重新選到該角色。先前把 `20→19` 視為一般 Delay 的假說已推翻：它是
ALT+Q 全員 Quick 的 handoff。engine 新增 mutable scheduler，CoAB 改成逐次
TeamList scan，並先把全體初始化 delay 投影回 Fighter，避免尚未行動的 Blink
怪物被誤判成 delay 0。頂層 D 現開啟資料化繁中子選單，D 延後、Q 結束回合；
READY spec 420 保存 bytes、位址、推論等級與未完成邊界。

2026-08-02 第四百二十一輪以非破壞性 IDA 副本稽核 PC-98 overlays
`08／13／18／24`。`CLEAR_ACTIONS` 清除 delay／move／guard 等 action bytes；
GUARD 設 Action `+07=1`，敵人移入鄰接後 consumer 先清旗標再攻擊；BANDAGE
依 TeamList 找第一名同隊 raw status 5，改為 4 並清 Action `+0E`；QUICK 寫
Player `+199=1`；S／F 將 `DS:7F16` 限制在 `0..9`，動畫 consumer 使用
`speed+3` 乘上 frame delay。命令名稱與位址是附加語意，原始 bytes／位置
未改動；spec 421 對每項結論標示 exact 或 strong inference。

Engine 新增作品中立 `combat/action` 狀態與速度比例，CoAB 接通 Guard 反應
攻擊、Bandage、目前角色 Quick、Space 手動收回及資料化繁中速度子選單。
engine 全套與 CoAB `./cmd/... ./gamepack ./internal/...` 31 套件正式 gate 已在
Docker／Xvfb 通過，並改用遠端 engine pseudo-version。`ALT+Q／ALT+M`、敵方
命令 AI、per-action target pointer、原版 wall-clock timing 與可用的正式路徑
截圖仍是明確缺口；本輪 X11 擷取只有黑色緩衝，已刪除且未當成果提交。

2026-08-02 第四百二十二輪以非破壞性 IDA 副本擴大 overlay 08 combat input
範圍。特殊鍵 `10h` exact 將目前 Action delay 寫 `14h`，沿 `DS:9598h`
TeamList 對每個 combatant 呼叫 local `1375h` Quick setter；下次 action entry
把 `14h→13h` 後依 Player `+199h` dispatch AI。Battle／State／Ebiten 接通
全隊 Quick，並讓視覺 timeline 播放中 Space 仍可收回可控制 PC，保留 NPC／
怪物 Quick。

正常 Standing Stone→GEO→紅網長路徑首次接入全隊 Quick 時，揭露 Space 原先
只清 transient Battle，戰後已同步為 Quick 的持久 `State.party` 仍保留舊值，
使下一場羅剎妖戰在 `Continue` 內被自動打完。`CombatManualControl` 現於成功
清除後同步 Battle→party；長路徑驗證四巨蛛自動戰、Space 收回與第二戰人工
停點均通過。

Docker／Xvfb 的 `./cmd/... ./gamepack ./internal/...` 完整套件 gate、locale
JSON parse 與 `git diff --check` 均為本輪提交門檻；原始 PC-98 overlays 全程
唯讀，IDA database／raw report 只留在 `/tmp/coab-ida-422`。

同輪追出特殊鍵 `11h` toggle `DS:A86Ch`、overlay 10 combat init 清零，以及
overlay 09 Quick AI spell selector 的 PC gate 與 consumer。這證明該旗標控制
玩家 Quick AI 是否考慮法術，但完整候選優先序與 suitability predicate 尚未
關閉；因此本輪不實作無效的 `ALT+M` 空開關。READY spec 422 保存 hashes、
位址、bytes、推論等級與未完成邊界。

2026-08-02 第四百二十三輪沿 overlay 09 Quick spell selector 追到全域 spell
table consumer。PC-98 `GAME.EXE` SHA-256
`8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` 的
`07h @ file 012E54h` 與 `0Fh @ 012ED4h` 是不同 16-byte records；selector
直接用 Player spell byte 乘 `10h` 讀 record `+0Dh／+0Eh／+0Fh`，沒有職業
基底轉換。這與 DOS record `+01Eh` 原樣保存 spell bytes、known flags 採全域
一基底索引交叉一致。

玩家 Magic Missile 已由錯誤 class-local `07h` 修成 `0Fh`；camp 顯示、記憶
流程與戰鬥分派一併改用全域 ID。舊 spec 134／142 已標 SUPERSEDED，READY
spec 423 保存 bytes、offset、推論等級與邊界。overlay 09 的三次隨機候選、
priority tier 與 suitability 雖已 exact 定位，但完整 cast consumer 與所有
候選法術在第 423 輪尚未完成，故當輪 `ALT+M` 不開放；第 424 輪已接續，
不能把這段歷史狀態誤讀成目前功能仍關閉。

2026-08-02 第四百二十四輪以合法授權的 IDA Pro 9.4 Docker 映像、唯讀
overlays 與 `/tmp/coab-ida-424` database 接續 ALT+M。overlay 09
`0627h..0754h` exact 證明 `1d7` tiers、每層三次 slot、priority 7 向下與
PC `A86Ch` gate；far `00B8:007Fh` 經 typed resolver 落到 overlay 13 entry 19、
`CASTCOMBATSPELL 27A1h`。該 routine 另證明 casting-time 0 與 pending cast 是
不同 handoff。

Engine 新增 `combat/quickspell` 與 `combat_ai_spells` schema；CoAB JSON 保存
十一筆已抽取 spell record selector metadata，State 不再硬編 priority 表。
Ebiten ALT+M 不與一般 M 移動衝突，每場戰鬥 gate 重設 off。全域 `0Fh` Magic
Missile 已在 focused visual regression 與 Standing Stone→GEO6→紅網四蜘蛛
正常玩家路徑由 ALT+M＋ALT+Q 施放並消耗 slot。非零 MinRange helper 已證明
不是單純距離；當輪 area spell、casting delay、Cure special 與其餘法術在抽中
時 fail-closed 並收回 PC 控制。casting delay／Bless 已由第 425 輪接續；
READY spec 424 只代表當輪邊界。

2026-08-02 第四百二十五輪以唯讀 IDA Pro 9.4 database 關閉
`CASTCOMBATSPELL` 非即時 handoff。overlay 13 exact 將 raw CastingTime
整除 3；非零時保存 Action spell ID，delay 改為 `max(1,delay-units)`；
overlay 08 在同輪重新選中後先清 spell ID 再呼叫 CASTSPELL。engine
`dd99d29` 新增作品中立 pending spell transaction，CoAB JSON 十一筆 record
補 raw casting_time。Quick Bless `01h` raw 10→delay 3 已接通；slot 消耗時點、
當輪手動 CAST、Cure special、其他 Quick 法術與中斷仍未完成；Cure 已由
426–427 接續，手動 CAST delay 已由 428 接續。READY spec 425 只代表當輪邊界。

2026-08-02 第四百二十六輪以 typed TPOV resolver 與唯讀 IDA Pro 9.4
關閉 Quick Cure target handoff。`00B8:0075h` 的正確型別對應是 overlay 13、
entry 17、handler `+1E30h`；entry 編號不可誤讀成 overlay 17。handler 搜尋
施法者周圍九格同隊受傷者，保留低目前 HP 候選，另處理 `Tile_DownPlayer=1Fh`
與 8 HP 門檻，最後寫出 far target pointer 並回傳是否成功。原始 bytes／位置
未修改，所有附加名稱均保存推論等級。

Engine action 新增 opaque `TargetID` 與 targeted spell transaction；CoAB
Quick Cure `03h` 以 stable fighter ID 跨 scheduler delay，結算時才消耗 slot。
focused regression 證明鄰近受傷目標、遠方排除、target 保存與治療；正常
Standing Stone→Myth Drannor→Red Plume 路徑由真實 ECL 箭傷建立目標並完成
七敵戰。equal-HP exact tie、倒地 status predicate、手動 CAST delay、施法中斷
與其餘 Quick 法術仍未完成。READY spec 426 是權威邊界。

2026-08-02 第四百二十七輪以 PC-98 Borland debug table、typed TPOV stubs、
raw bytes 與唯讀 IDA report 關閉 `COMPTARGCURE` 選人順序。`DXDIR／DYDIR`
exact 證明九格為 N→NE→E→SE→S→SW→W→NW→self；strict-lower branch 讓
equal HP 保留先掃到者，自身最後且低於半血時覆蓋先前候選。合法 down-player
只在 active 最佳 HP `>=8` 時取代。

Borland `CHARSTATUS` members 證明 raw `UNCONC=4／DYING=5／DEAD=6／STONED=7／
GONE=8`；overlay 13 `+1E10h` set bytes `C0 01` 與 membership branch exact
排除 `{6,7,8}`。CoAB 不再按全候選 HP 排序，並修正 Stoned 不可 Cure；auditor
新增唯讀 `-members` 模式。四組 focused boundary 與既有 Red Plume 正常路徑
是本輪驗收。READY spec 427 是權威；同格多 corpse ordering、完整 raw status
投影與中斷仍未完成；手動 CAST delay 已由第 428 輪接續。

2026-08-02 第四百二十八輪沿第 425 輪 exact `CASTCOMBATSPELL` 證據，把
手動 CAST 接入與 Quick 相同的 mutable scheduler transaction。engine action
新增 renderer-neutral point target；CoAB Battle／State 分別保存 stable
combatant ID 或 32×16 格點，resume 後才執行既有 spell effect。手動 Bless
證明 pending 階段不先耗 slot／套效果；Fireball 證明 delay=1 即使同一呼叫
重新入列，仍使用玩家選定 `(7,6)`。原始 binary／IDA database 全程未修改，
typed target 欄位只屬 remake projection；正傷害中斷與 slot 時點由第 429 輪
接續，其他中斷仍未完成。

2026-08-02 第四百二十九輪以合法 IDA Pro 9.4 在 Docker 對唯讀 overlay
23／24 產生附加 ledger。`PUTDAMAGE 013E:1FFDh` 只在最終 `DS:A02Eh>0`
且 combat mode 時檢查 Action spell；CP932 short string `1F76h` 是「已無法
繼續吟唱法術」。typed `014A:0070h` 落到 overlay24 entry16 `1739h`，exact
掃描 `Player+1Eh..71h` 並清第一個 matching memorized byte，之後
`PUTDAMAGE 21FCh` 清 pending spell。

Engine action 新增 title-neutral interruption clear；CoAB Battle 所有已接通
正傷害來源共用 event queue，State 依 stable fighter／spell ID 修改正式 roster
並從 locale `combat_spell_interrupted` 顯示繁中。正傷害／零傷害、第一個重複
slot、既有手動 Bless／Fireball regressions 是 focused gate。Cloudkill 直接
死亡、非傷害狀態、monster raw writeback 與原版動態時間仍未完成。

2026-08-02 第四百三十輪沿毒雲術的獨立非傷害路徑繼續。overlay 22 的日文
short string、3×3 tile `1Ch` writer 與 callback exact 加入 raw effect `44h`；
overlay 12 `INITEFFPROX` 把 zero-based slot 67 解析到 resident `008B:016Fh`、
overlay-local `1621h`。該 handler 在 combat mode 檢查 pending Action spell，
顯示「未能完成吟唱法術」，呼叫第 429 輪已證明的 overlay 24 memorized-byte
consumer，再清 pending spell。原始 executable／overlays 全程唯讀，附加 IDC
ledger 保留原 offset、bytes、disassembly 與推論等級。

CoAB Battle 抽出共用 interruption event helper；`CastCloudkill` 對 HD 0–4
自動死亡與 HD 5–6 豁免失敗先建立 event，再進死亡 handoff。State 由正式
roster stable ID 移除第一個 matching slot，顯示文字由正式 locale 取得。HD 7+
正反例保留；共用 engine action API 已足夠，因此本輪沒有 engine source 變更。
沉默、麻痺、睡眠、石化、monster raw writeback、毒雲每回合判定與原版動態
時間仍未完成；READY spec 430 是權威。

提交前以唯讀 overlays 在乾淨 `/tmp/coab-ida-430-final` 重建 IDA ledger：
overlay 12／22 分別為 5,136／11,103 行，並以非空檔案作門檻，不能只信 IDA
exit code。Docker／Xvfb、`--network none`、暫時 local-engine replace 的
`./cmd/... ./gamepack ./internal/...` 正式 gate 共 31 套件通過；Cloudkill
三組 core 與兩組 State／玩家路徑測試另以 `-count=1` 通過。

2026-08-02 第四百三十一輪證明 held effects 與前兩種中斷不同。
`INITEFFPROX` 的 `1Fh／33h／34h／35h` table slots 全寫入 resident
`008B:0039h`；typed resolver 落到 overlay 12 entry 5、local `0075h`，再呼叫
`014A:00CAh → overlay24 entry34 local 2A5Bh`。完整 consumer 清
`Action+03h delay／+00h pending spell／+07h guard／+06h unknown`，函式內沒有
call，也沒有抵達 overlay 24 `1739h` memorized-slot consumer。因此 held
取消 pending cast 但保留法術格，是 exact control flow，不可沿用正傷害／
毒雲術的 slot-consumption event。

State 正常 mutable scheduler 現對四個 raw IDs 先 `ClearAction`、同步 UI
selection，再顯示「無法行動」並跳過回合；四例不建立 interruption event，
正式 roster slot 保留。第 431 輪專用 IDC 由唯讀 overlay 12／24 重建 29＋21
行 ledger。六章 shipped MON*SPC 無這四種 innate effect，Sleep／Hold 動態
writer、豁免、解除、動畫與聲音仍未完成；READY spec 431 是權威。

提交前 focused `internal/combat`／`internal/game` 與完全離線 Docker／Xvfb
31 套件 gate 均以 `-count=1` 通過；第 431 輪沒有 engine source 變更。

2026-08-02 第四百三十二輪由 PC-98 全域 spell record `15h`、overlay 22
`INITSPELLS` 及 typed TPOV resolver 閉合 Sleep dispatch：table slot
`A3E0h` 寫 `0117:00EDh`，落到 entry 41／local `2547h`。專屬 handler exact
擲 `4d4`，依既有 target order 用 HD 成本 `1／2／4／6／10或20／20`
扣除；已有 effect `35h` 或容量不足者清 pointer，但仍繼續掃後續候選。

獨立 engine 新增 `combat/sleep` 與繁中知識庫，只保存上述容量 primitive；
五 HD target `+74h` 尚未命名。第 432 輪 IDC 在唯讀 overlay 副本產生 397 行
ledger。GETSPELLTARGETS 幾何／排序、save、magic resistance、duration、
effect writer 完整參數、手動／Quick cast、動畫與音效仍未完成；下一輪應先
閉合 `GETSPELLTARGETS 112Ch` 對 record `15h` 的上游分支，再接玩家路徑。

Engine commit `73c0144` 已推送。engine `go test -count=1 ./...` 全數通過；
CoAB 第一次兩次 formal gate 因 Xvfb socket／display readiness 配置失敗，
其餘套件通過但不能冒充完整 gate。改用具名 display `:432`、檢查 Xvfb PID
與 socket 後，Docker／Xvfb／`--network none` 的
`go test -count=1 ./cmd/... ./gamepack ./internal/...` 共 31 套件正式通過。

2026-08-02 第四百三十三輪訂正第 432 輪把 `DOSPELLTARGETING`
`0117:0034h` 直接連到錯誤 overlay 的位址空間混用。`INITSPELLS` 的預設值
經 `22:0034h` typed resolution 是 overlay 22 entry 4、local
`112Ch GETSPELLTARGETS`，讀 `AOENONCOMBAT`；combat init 另把 pointer 改成
`00B8:007Ah`，經 `13:007Ah` 落到 overlay 13 entry 18、local `225Fh`，讀
`AOECOMBAT`。Sleep `15h` 的 `AOECOMBAT=09h` 走選格分支，再以 shape 1 呼叫
overlay 31 `SCAN 08D8h`；其三 byte候選表經 local `0035h` 排序，overlay 13
依原順序投影成 `SPELLTARGET`，沒有第二次排序。

共用 writer 對固定 Sleep record 讀 `SAVERESULT=0`，故 exact 不呼叫
saving-throw helper；duration 是 `DURATIONFIXED(0) + FIGCASTERLEVEL ×
DURATIONPERLEVEL(5)`。`SCAN` 第二／三欄的正式幾何名稱、large footprint、
tie order runtime、magic resistance、`PUTEFFECT` 完整 record、解除與演出仍
未閉合，因此仍 fail-closed，不開放手動／Quick Sleep。收斂後的 IDC 以兩份
全新唯讀 overlay 工作副本重建：overlay 13／31 ledger 分別為 959／443 行，
並核對 SHA-256；原始 overlays 與既有 IDA database 均未修改。READY spec 433
是本輪權威，spec 432 的舊 targeting 敘述已明確 supersede。

2026-08-02 第四百三十四輪由 Borland `EFFECTS` symbols、typed TPOV resolver
及唯讀 IDA ledger 訂正 `PUTEFFECT` projection：`013E:0089h` 應以 overlay 23
解析，exact 為 entry 21／local `2325h`；先前 `12:0089h` 的 local `043Ah`
屬另一 control segment，已推翻。新 `/tmp/coab-ida-434` overlay 23 工作副本
輸出 729 行非空報告，SHA-256 為
`a3ea0d9528be57a92c33fc345baa3e27eef375c84822afba0cfbb141c2faabc9`，
原始 binary 與既有 database 未修改。

Sleep 在四次 d4／ordered HD filter 後逐 target 進 `PUTEFFECT`：先寫
`A02D=CASTON(1)`，呼叫 `CHECKFX(target,9)`；operational `6Ah` 依既有 exact
公式擲 d100，成功清 `A02D`、顯示 CP932「には影響がなかった。」並跳過
`ADDEFFECT`，但不退容量。成功 record exact 為 kind `35h`、duration
`5×caster level`、raw `+3=1`、`+4=caster level`。CoAB 新增
`CastSleepOrdered` bounded core 與 raw `+4／Player+74h` 保存；它不猜 SCAN
幾何，也未開手動／Quick UI。focused `internal/combat`／`internal/monster`
測試已通過。private engine pseudo-version 的 go.sum 已透過唯讀 credential
mount 取得；正式 gate 暫以同一 `73c0144` nested engine replace 驗證後移除，
Docker／Xvfb／`--network none` 的 `go test -count=1 ./cmd/... ./gamepack
./internal/...` 全數通過。隊伍／死亡 predicate 尚屬上游 `SCAN`，bounded core
不自行排除。engine 繁中知識庫 commit `7c3acd7` 已推送；CoAB 狀態文件已
同步，待本 repo 集中提交與推送。

2026-08-02 第四百三十五輪以擴大的非破壞性 IDA audit 關閉 overlay 31
`SCAN 08D8h`、`LOSEXISTS 03EAh` 與 local `0035h` 排序。三 byte record
exact 是 object ID、最小成功 LOS 加權距離低 byte、方向 sector；spec 433
先前把第三欄列入 tie 的說法已被推翻。排序器只按第二欄，再以第一欄 ID
與奇偶例外作巢狀交換；第三欄只隨整筆搬移。

Engine `combat/scanorder` commit `b75e169` 已通過 `go test ./...` 並推送；
CoAB `OrderScanTargetIDs` 以一基底 object ID 映射 stable fighter ID，對零、
空值與重複投影失敗即關閉。這仍未建立真實 terrain records，也未開放手動／
Quick Sleep；下一輪應由 PC-98／DOSBox 固定戰場驗證 wall、large footprint
與 cursor，接入既有 `CastSleepOrdered`。READY spec 435 是權威邊界。
全新 `/tmp/coab-ida-435-final` 從唯讀 overlay 重建 1,187 行報告；正式
Docker／Xvfb／`--network none` gate 以 `-buildvcs=false` 明確列出並驗證
31 個 CoAB 套件，避免外置 git-dir 讓不完整測試輸出被誤判成全數通過。

2026-08-02 第四百三十六輪以 Borland 0x52FB symbol／type／member table、
全新非破壞性 IDA raw resident report 與 overlay 31 連續指令閉合
`TDEFTYPE HT／LOS／SYM`、`TACTICALMAP XRAY／TD`、一基底 tile index、
cardinal 2／diagonal 3 metric 及 inclusive `2*range+1` gate。`XRAY` 只略過
`SYM <= source LOS`，不略過距離；TDEF 第四 byte 沒有 Borland member，故
只保存為 `Raw3`。

Engine `combat/scan` commit `9c94ddd` 已通過 `go test ./...` 並推送；CoAB
dependency 已升至 `v0.0.0-20260802035038-9c94dddc1dd5`。CoAB
`BuildScanTargetIDs` 使用 explicit legacy object ID／footprint／direction
callback，並以 bounded `producer → order → CastSleepOrdered` 回歸證明地形
阻擋不會誤施睡眠術。`INARC` sector、COMPOBJ builder、PC-98 wall／corner
動態 trace、正常手動／Quick Sleep、效果生命週期與演出仍未完成；READY
spec 436 是本輪權威。

Engine 全套與 CoAB focused 測試通過。CoAB formal gate 先明確列出 31 個
import paths，再於 Docker／Xvfb／`--network none`／`GOPROXY=off` 逐套件
執行；去重後 31 行 package pass ledger SHA-256 為
`04e7594f05582c1e8a5f16d4c8c8b2b1e532a25e9fe2831c976593ebe5d7bf7b`。
早先 pattern／JSON 嘗試只跑第一 package，均未當作完成證據。

2026-08-02 第四百三十七輪由 PC-98 overlay 31 完整 `INARC 054Ah..08D5h`
關閉八方向 inclusive 半平面、`FFh→8`、相鄰格捷徑與第一命中方向。engine
`combat/scan.Build` 改接原始 arc，不再要求 title callback；50×25 全 source／
target／九 sector 共 14,062,500 組 reference corpus 通過。

Borland symbols exact 命名 `LASTOBJECT 0C29:9740` 與 `OBJECTLIST
0C29:9741`；型別表證明 72×4-byte。新唯讀 IDA audit 在 overlay 10 找到
builder 從 object 1 沿 combatant far-pointer linked list 建立 X／Y／自身 index／
footprint-active 欄位。linked-list→stable fighter ID 尚未閉合，故本輪不開放
正常 Sleep UI。READY spec 437 與 `/tmp/coab-ida-437` 是續作入口。

2026-08-02 第四百三十八輪由 Borland types 補齊身份橋：`CHARACTERLIST
9598h` 是 `CHARRECPTR` head；`IDLIST 9DD3h` 是 72×4-byte `CHARRECPTR`，
overlay 10 builder 把 traversal 中同一 far pointer 寫入 `IDLIST[objectID-1]`，
再沿 `CHARREC.NEXT +18Ah` 前進。這證明 SCAN object ID 以 pointer identity
對應角色，不是顯示名稱或偶然 slice index。

CoAB `Fighter.LegacyObjectID`、`StartCombat` 一基底重建、
`LegacyScanObjects／BuildLegacyScanTargetIDs` 已串成
`CHARACTERLIST→IDLIST→footprint→INARC/LOS/sort→stable ID` transaction，
72 筆上限與身份缺損均 fail-closed。frontend 仍缺目前戰場 `TD/TDEF`，正常
Sleep UI 繼續封鎖；READY spec 438 是續作入口。

2026-08-02 第四百三十九輪閉合 Sleep 的實際 source center 與戰場 terrain
adapter。overlay 13 Pascal 壓棧及 overlay 31 callee consumer 共同證明
`A645／A646` 玩家選定格是 SCAN source，`AOECOMBAT&7=1` 是 range，
`FFh` 是全方向 arc；不得使用 caster footprint。PC-98 65 筆 TDEF 與
`BackgroundTiles[1:66]` 逐 byte 一致，Dungeon／Wilderness floor bytes 是
一基底 TD，不是 renderer TileIndex。frontend provider、State 手動 `Z` 選格、
legacy identity／terrain SCAN、`CastSleepOrdered` 與成功後 slot transaction
已接通；invalid map 不吃 slot。32×16 fallback placement 仍是 reconstructed，
Quick Sleep、wall/corner runtime、effect 解除／save、twinkle／音效仍未完成。
READY spec 439 是續作入口。

2026-08-02 第四百四十輪以全新 `/tmp` IDA 9.4 副本追蹤 overlay 23
`PUTDAMAGE 1FFDh → REMOVEFX 158Ah → SPELLOFF 010Eh`。關鍵訂正是
`REMOVEFX` 的 `[di+159Dh]` 使用 `DS`，不能讀 overlay 同 offset；resident
`DS:159Eh..15B1h` 19-byte 表 exact 含 Sleep `35h`。因此只有實際正傷害會
解除動態睡眠，零傷害不會。`Battle.applyPositiveDamage` 已統一接通，並以
一般攻擊、零傷害及 innate MON*SPC 邊界回歸。duration 遞減 consumer、
combat-end／save、醒來文字、twinkle 與音效仍待續；READY spec 440 是入口。

2026-08-02 第四百四十一輪由 PC-98 Borland `TIMEUNITS`、resident
`MAXCOUNT DS:6804h` 與 `CLOCK_` overlay 20 local `0020h` 關閉
`EFFECTREC+1` duration consumer。原 routine 遍歷 `CHARACTERLIST`／
`Player+F2h` linked list，以十 tick chunk 遞減，duration 零保留、到期經
`SPELLOFF`。engine `combat/effecttime` commit `3142ae0` 已推送；CoAB
Battle 每個新 round 扣一 tick，正常 level 3 手動 Sleep 已驗證
`15→handoff 14→總第15 tick解除`。active battle save 與醒來演出仍待續，
READY spec 441 是入口。

2026-08-02 第四百四十二輪由 `INITEFFPROX` effect slots、typed overlay stub
與 overlay 24 `CLEARACTION／TWINKLE` 完成 Sleep Action 及演出生命週期。
effect `35h` 加入、受傷移除與自然到期都清除完整 Action；傷害路徑先移除
effect，故不會再由一般 interruption 重複消耗 memorized slot。成功目標逐人
播放 runtime 四格 24×6 `TWINKLE`、`SPELLHITFX` 與每人 1440ms delay-only；
魔抗者與醒來 callback 無此演出。renderer 幾何／時間 exact，palette pixels
因尚無 PC-98 runtime capture 標為 layout-reconstructed。active battle save、
Quick Sleep 與實機 palette／draw overhead 仍未完成；READY spec 442 是入口。
正式 Docker／Xvfb／`--network none` gate 已明確驗證
`./cmd/... ./gamepack ./internal/...` 全數通過，marker 為
`ROUND442_FORMAL_EXIT=0`，暫存日誌在 `/tmp/coab-round442-formal.log`。

2026-08-02 第四百四十三輪將 remake JSON save 升至 version 7，新增 bounded
`BattleSnapshot` 與 CoAB `CombatSnapshot`。stable fighter order、effect／Action、
round、persistent areas、attack modifiers、dynamic scheduler selection、pending
interruption、battle randomstream、State turn／target／spell／move cursors及尚未
開始的 visual transaction 均可 round-trip。正常手動 Sleep 以 seed 443 從
選格施法後存檔；原狀態／loaded 狀態完成 TWINKLE handoff 後 snapshot 相同，
自然到期與讀檔後正傷害喚醒兩分支通過。mid-animation save 因 renderer elapsed
未歸 State 所有而明確拒絕；原版 SAVGAM combat layout 仍 unknown。READY
spec 443 是入口。
正式 Docker／Xvfb／`--network none` gate 已驗證
`./cmd/... ./gamepack ./internal/...` 全數通過，marker
`ROUND443_FORMAL_EXIT=0`，日誌位於 `/tmp/coab-round443-formal.log`。

2026-08-02 第四百四十四輪把 save v7 插入既有完整
`TestRealPlayerPathStandingStoneToBurialGlen`。路徑由 Standing Stone 正常世界
旅行、GEO6 合法步行、精靈幽魂、紅網字串輸入抵達四蜘蛛第一戰；party-turn
邊界存檔後，以同一玩家自備 image 建立全新 State，重掛 MON6CHA／MON6SPC／
ITEM6，Battle 與 ECL session snapshot 逐欄相同。後續只使用 loaded state，
完成蜘蛛、Picture 72、羅剎妖第二戰、`4CBF=1`、dungeon return 及原長路徑
後續。高數值英雄只縮短回歸，不支持完整 encounter balance。READY spec 444
是入口。
正式 Docker／Xvfb／`--network none` gate 已驗證
`./cmd/... ./gamepack ./internal/...` 全數通過，marker
`ROUND444_FORMAL_EXIT=0`，日誌位於 `/tmp/coab-round444-formal.log`。

2026-08-02 第四百四十五輪在同一真實 campaign 長路徑繼續至 outer ruins
terrain `01h`。提爾雪雅第一戰後選擇攻擊貝爾哈，ECL 建立 party-side
RAKSHASA `43h` 臨時盟友與 12 名敵人；第二戰 party-turn save v7 後建立全新
State。Battle／ECL session／TemporaryAlly Fighter 逐欄相同，Characters roster
始終只有英雄。loaded state 戰勝、寫 `4CD1=1`、回 dungeon 後，runtime party
與 roster 都只留英雄。READY spec 445 是入口。
正式 Docker／Xvfb／`--network none` gate 已驗證
`./cmd/... ./gamepack ./internal/...` 全數通過，marker
`ROUND445_FORMAL_EXIT=0`，日誌位於 `/tmp/coab-round445-formal.log`。

2026-08-02 第四百四十六輪解除 save v7 的 mid-visual fail-closed。State 新增
權威 `combatVisualElapsed`，snapshot 以納秒保存，frontend 以 saved base 加新
wall-clock delta 續跑；travel／impact／death 已送 cue markers 同步 round-trip。
正常手動 Sleep 在 700ms `TWINKLE` 中段與一擊致死弓箭 death frame 都由全新
State 載入同一 frame，載入、同幀重入及 handoff 不重播已送離散音效。超時
elapsed、越界 marker、無 event 的非零 elapsed 及時間倒退均 fail-closed。
READY spec 446 是權威；播放器 PCM sample offset、BGM driver／synth snapshot
與原版 SAVGAM combat layout 仍未完成。
正式 Docker／Xvfb／`--network none` gate 已驗證
`./cmd/... ./gamepack ./internal/...` 全數通過，marker
`ROUND446_FORMAL_EXIT=0`，日誌位於 `/tmp/coab-round446-formal.log`。

2026-08-02 第四百四十七輪將 remake JSON save 升至 v8。獨立 engine
`f06493f` 暴露 vendored `ymfm_saved_state` 與有理數 resampler snapshot；FM＋
SSG chip 及 resampler 的後續 samples 與不中斷分支逐 sample 相同。CoAB 再保存
七聲道 SequenceMachine、parameter renderer、Timer B、MSCPLAY silence、pending
PCM、stable track ID 與 selector。Ebiten `Position()` 是扣除 device buffer 的
audible frame；TrackPCMStream 以四秒 bounded emitted history補回 decoder
read-ahead，使 synthetic 七聲道 fixture 從第一個尚未聽見的 byte續跑一致。
save JSON 會拒絕未知 version、selector／rate、超大 state／pending及 stack。
READY spec 447 是權威。本機與使用者主目錄找不到 SHA `bddbe63…b12f5` 的完整
MSCDRV，玩家 VFD 又缺 driver sector，故十二首真實曲目 runtime save/load、
active one-shot sample position與原版 SAVGAM audio仍未完成。
engine 全套 gate marker `ENGINE_ROUND447_FORMAL_EXIT=0`，日誌位於
`/tmp/engine-round447-formal.log`；CoAB Docker／Xvfb／`--network none` 全套
marker `ROUND447_FORMAL_EXIT=0`，日誌位於 `/tmp/coab-round447-formal.log`。

2026-08-02 第四百四十八輪將 remake JSON save 升至 v9，補齊第 447 輪留下的
active one-shot sample position。DOS WAV 與 PC-98 software-speaker 以 backend、
stable selector／event、enabled 與 44,100 Hz audible frame保存；不同音效可重疊
續跑，自然結束、停用及舊版未保存的音效不復活。backend／asset／seek 錯誤會先
停止 pre-load voice 並失敗即關閉。frame↔duration 依 Ebiten Position 的 floor
語意做可逆整數換算。READY spec 448 是權威；這不證明原版 SAVGAM audio，也尚缺
實體音訊裝置 loopback 逐 sample oracle。
正式 Docker／Xvfb／`--network none` gate 已驗證
`./cmd/... ./gamepack ./internal/...` 全數通過，marker
`ROUND448_FORMAL_EXIT=0`，日誌位於 `/tmp/coab-round448-formal.log`。

2026-08-02 第四百四十九輪採使用者指定的代表性抽樣，沒有重跑開場至結局
marathon。`-inner-final-battle` 由 block `43h` 正常初始化，經 terrain `97h`
樓梯與 READY spec 408 的十步 GEO 路線抵達 `9Ah`，runtime 斷言原始
MARGOYLE `45h`×28、TYRANTHRAXUS `47h`×1、HIGH PRIEST `48h`×8。
renderer 修正大型 CombatMap 使用全體 bounds 中心的回歸，正式鏡頭恢復依
RuleBook 跟隨主動角色；只有 screenshot capture 啟用 `47h` 首領觀察焦點，
不改 fighter 座標、ECL 或 AI。640×480 圖片已更新 README，標為
`material-exact/layout-reconstructed`。READY spec 449 是權威；正式 gate、
圖片 hash、Docker 清理與 push 狀態待本輪結束補記。

正式 Docker／Xvfb／`--network none` gate 已驗證
`./cmd/... ./gamepack ./internal/...` 全數通過，marker
`ROUND449_FORMAL_EXIT=0`，日誌位於 `/tmp/coab-round449-formal.log`。終戰圖
SHA-256 為 `e7d2fa5763c95fc53478d4ec78cc230779a68e2375e1b1a0adef635e8ce6e6c7`；
原始尺寸經 renderer 輸出與人工檢視確認為 640×480。一次性容器均以 `--rm`
完成，沒有 `coab-round449` 容器殘留；一個 13 日前的未知懸空映像不屬本輪，
為避免清到其他專案而保留。commit／push 狀態在提交後補入下一輪或以實際
HEAD／remote 為準。

2026-08-02 第四百五十輪依「故事與手札不可硬編碼」原則，將 READY spec 315
涵蓋的法師塔庭院、德拉坎德羅斯現身與定身、黑龍塔頂、幻象攻擊、手札 15
及枷印消退遷入 CoAB game-pack。新增九段事件 stable message IDs、兩頁
`journal.15.*` 與十個原 ECL fragment rules；en／zh-TW 各 252 keys 且完全
對齊。State 的九個 switch cases、手札特判及舊 locale 11 筆複本已刪除。
手札仍只在原 ECL 事件後直接寫入遊戲內 `JournalPages`，不要求玩家查 PDF。
聚焦 `./gamepack ./internal/game` Docker 測試已通過；正式 gate、commit 與
push 狀態待本輪結束補記。READY spec 450 是權威。

正式 Docker／Xvfb／`--network none` gate 已驗證
`./cmd/... ./gamepack ./internal/...` 全數通過，marker
`ROUND450_FORMAL_EXIT=0`，日誌位於 `/tmp/coab-round450-formal.log`。聚焦
`./gamepack ./internal/game` 亦先行通過；掃描確認九個舊 ECL catalog IDs 與
兩個 `journal_entry_15_*` 已不再存在於 State／舊 locale。一次性容器均以
`--rm` 完成，無 `coab-round450` 容器殘留；engine repo 本輪無變更。

2026-08-02 第四百五十一輪接續法師塔資料分離：READY spec 316–319 的
ATTACK WIZARD、PARLAY、14 黑龍、龍心與屋頂出口共十段文字，以及三個
法師塔專用選項，均移入 CoAB JSON。State 十個 story cases、三個 option
cases 與舊 locale 十三筆複本已刪；en／zh-TW 各 265 stable IDs。聚焦
`./gamepack ./internal/game` 已通過，正式 gate／commit／push 待本輪收尾。
使用者另要求發行前全面 UI／README 圖片稽核，並將 SSI Gold Box 經驗整理
成繁中知識庫與 skill；`daemon_winter` 僅作待驗證比較樣本，Wasteland 為後續
獨立中文化目標。以上已寫入 AGENTS，不可在 compact 後遺失。

本輪已完成 daemon_winter 唯讀初步比較：該專案 AGENTS 與
`docs/design/engine-extraction-study.md` 的本機證據指出 `DEMON.INT` 是有
3,807 relocations 的原生 MZ 8086，不是 bytecode VM。其 `dwstrings uicheck`、
500／500 coverage、theme 原子切換、唯讀原版資料、sampled A6 與 release
deny-list 可作方法參考；Gold Box ECL／DAX／GEO／combat／save 不可直接外推。
繁中比較與 Wasteland 入口已保存於
`docs/knowledge/ssi-rpg-cross-project-lessons.md`。

正式 Docker／Xvfb／`--network none` gate 已驗證
`./cmd/... ./gamepack ./internal/...` 全數通過，marker
`ROUND451_FORMAL_EXIT=0`，日誌位於 `/tmp/coab-round451-formal.log`。掃描確認
本輪十個舊 ECL catalog IDs 與三個 option catalog IDs 已不再存在於 State／
舊 locale；一次性容器均以 `--rm` 完成，無 `coab-round451` 殘留，engine repo
本輪無變更。

2026-08-02 第四百五十二輪借用冬之魔 `uicheck` 經驗，新增
`cmd/coab-audit`／`internal/sourceaudit`。Go AST scanner 只掃正式非測試 string
literals，以 path／function／SHA-256／occurrence exact baseline 阻止任何漂移；
初始 1,260 signatures／1,315 occurrences，分類為 localization 409、frontend
164、runtime 742。repository regression 已通過；此基線是必須逐輪下降的技術
債，不是豁免額度。正式全套 gate、commit／push 待本輪收尾；READY spec 452
與 `docs/audit/README.md` 是權威。

正式 Docker／Xvfb／`--network none` 驗證先執行 `go run ./cmd/coab-audit`，
再執行 `./cmd/... ./gamepack ./internal/...`，兩者全數通過；marker 為
`ROUND452_AUDIT_EXIT=0`、`ROUND452_FORMAL_EXIT=0`，日誌位於
`/tmp/coab-round452-formal.log`。一次性容器均以 `--rm` 完成，沒有
`coab-round452` 殘留；engine repo 本輪無變更。

2026-08-02 第四百五十三輪選擇訓練場作為 exact baseline 後第一個設施切片。
訓練提示、角色摘要、離場、三種拒絕原因、確認、升級結果、HP 結果、職業名稱、
選法術與學會法術訊息均由 locale stable ID 解析；`trainingSpell` 只保留全域
spell ID／class／level／key，49 筆可學法術中文名稱不再存在於 Go。測試改讀
正式 `zh-TW.json` 並依 stable ID 取得期望值，另逐一驗證所有 UI／class／spell
key coverage。source audit exact baseline 由 1,315 降至 1,251 occurrences，
`runtime_ui_debt` 由 742 降至 678，無新增漢字 literal。既有真實 ECL／GEO
整合路徑另改用正式 locale，驗證 PICTURE 4→PROGRAM 0→選角→確認→升級→
離開並返回 `(5,2)`，不是直接呼叫服務冒充玩家路徑。正式 Docker／Xvfb／
`--network none` 全套 gate 通過，marker `ROUND453_FORMAL_EXIT=0`；commit／push
與容器清理待本輪收尾。READY spec 453 是權威。

2026-08-02 第四百五十四輪接續把剛德神殿完整服務文字移出 Go。十種
`templeCure` 只留 stable locale key／價格；主選單、治療選單、確認、查看、
pool／share／appraise、金幣不足與完成訊息均由正式 `zh-TW.json` 解析。
coverage 測試遍歷 19 個 UI keys 與十個 cure keys；Cure Light Wounds 測試
由 stable ID 取得確認文字，Remove Curse 仍驗證 `24h` effect／cursed item。
真實 ECL／GEO 路徑從 terrain `92h`、PICTURE 6 進 service boundary，治療後
續跑並返回 `(0,7)`。source audit 從 1,251 降至 1,223，runtime 類由 678
降至 650，無新增漢字 literal。正式 Docker／Xvfb／`--network none` 全套
gate 通過，marker `ROUND454_FORMAL_EXIT=0`；commit／push 與容器清理待
收尾。READY spec 454 是權威。

2026-08-02 第四百五十五輪接續 BAR 與酒館資料分離。一般傳聞服務的選單、
六則預設 tale、結束／空結果，以及提爾佛頓 ECL 的酒保、四飲品、特別客人、
紫帶女子、騷動、匕首與手札 17 提示均由 locale stable IDs 解析。英文 option、
prompt、fragment 仍保留作原始來源 identity。coverage 驗證 41 keys；提爾佛頓
terrain `88h` 正常路徑完成飲品→事件→遊戲內手札→返回 `(6,10)`，Ashabenford
Tale 28、Essembra Tale 60 與手札 17 prefix／全文也改讀正式 locale。source
audit 從 1,223 降至 1,169：localization 409→375、runtime 650→630、frontend 164
不變。正式 Docker／Xvfb／`--network none` 全套 gate 通過，marker
`ROUND455_FORMAL_EXIT=0`；commit／push 與容器清理待收尾；READY spec 455
是權威。
