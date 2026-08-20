# 驗證報告：`SEG-03` 段落 id 與標籤

- 日期：2026-08-20
- 範圍：主線分段驗證計畫的 `SEG-03`（段落註冊表的前半：id 規則與人類可讀標籤）
- 結論：**通過**。25 個段的 id 與標籤都定了，標籤逐條有原作敘述為證。

## 一、段的 id 不用地圖名

計畫原本寫「9 個佔位地圖名字沒定，段的 id 也定不了」。**那個相依是自己加的。**

段的 id 一律是 **`ECL{成員}/0x{block}`**——機械、穩定、與 game pack 的命名無關。
人類可讀的名字是**標籤**，放
[`docs/plan/segment-labels.json`](segment-labels.json)，改標籤不會動到 id。

順帶省掉一次大搬風：`original.geo*` 這組佔位 id 在程式、測試、game pack 與歷史
文件裡有 **42 處引用**，為了好看去改它們是純粹的 churn。

## 二、標籤的證據

用 `cmd/ecl-block-text` 從每個 block 撈原作敘述判定。逐條：

| 段 | 標籤 | 原作敘述 |
|---|---|---|
| `ECL3/0x10` | 猶拉什：地面與指揮部 | `SMOKE RISES FROM BEHIND THE RUINED WALLS`／`TROOPS COME BURSTING OUT OF THE COMMANDER'S` |
| `ECL3/0x11` | 猶拉什：地下第一層（教團）| `YOU SEE THREE CULTISTS LYING DEAD ON THE FLOOR.` |
| `ECL3/0x12` | 猶拉什：地下第二層（散提爾）| `THE MANGLED REMAINS OF A DEAD ZHENTRIM`／`SCROLL WITH THE SEAL OF ZHENTIL` |
| `ECL3/0x15` | 猶拉什：戰火街區 | `YOU FIND A WAR BLASTED SECTION OF THE CITY.` |
| `ECL4/0x20` | 散提爾堡內城 | `THE GUARDS TAKE YOU ASIDE FOR QUESTIONING.` |
| `ECL4/0x21` | 散提爾堡：神殿與牢房 | `YOU ARE DRAGGED THROUGH THE TEMPLE.`／`YOU WAKE UP IN A DARK, DREARY CELL.` |
| `ECL4/0x22` | 散提爾堡：迪姆斯華特與脫身 | `THE OLD MAN INTRODUCES HIMSELF AS DIMSWART THE SAGE` |
| `ECL4/0x23` | 散提爾堡：法庭 | `YOU ARE DRAGGED INTO THE COURTROOM.`／`THE MAGISTRATE` |
| `ECL4/0x25` | 眼魔洞穴 | `'THE MULMASTER BEHOLDER CORPS IS RUMORED TO BE` |
| `ECL5/0x30` | 黑暗精靈章尾聲 | `YOUR DARK ELF WEAPONS AND ARMOR DECAY TO USELESSNESS.`／`AKABAR SPEAKS` |
| `ECL5/0x31` | 哈普村 | `THIS RUN DOWN VILLAGE IS STRANGELY QUIET.` |
| `ECL5/0x32` | 古熔岩洞 | `YOU HAVE ENTERED AN ANCIENT LAVA TUBE.` |
| `ECL5/0x33` | 巫師塔 | `YOU HAVE COME OUT INTO THE COURTYARD OF A FIVE STORY TOWER.` |
| `ECL5/0x35` | 深層地城入口 | `'IT'S PRETTY FIERCE DOWN THERE. THE DEEPER YOU GO THE NASTIER THE CREATURES.` |
| `ECL6/0x45` | 密斯卓諾：外圍遺跡區 | `YOU HAVE FOUND AN EERIE BLOCK OF RUINS.` |

其餘十個段（提爾佛頓一族、密斯卓諾三張、世界地圖 hub、開場）沿用 game pack
既有的命名。

## 三、⚠ 一個宣告與內容不符

`SEG-02` 標出 `zhentil-keep.beholder-cave` 宣告 `script_block 0x22` ＋
`geometry_block 0x25`，與 `LOAD FILES` 對不上。這一輪的文字證據把它講清楚了：

- **眼魔的內容在 `ECL4/0x25`**（`THE MULMASTER BEHOLDER CORPS`），而 `0x25` 的
  `LOAD FILES` 正是 `25`。
- **`ECL4/0x22` 是迪姆斯華特那一段**（`DIMSWART THE SAGE`），`LOAD FILES` 是 `21`
  ——也就是神殿／牢房那張地圖，與「醒過來、老人自我介紹」的敘述一致。

⇒ `beholder-cave` 的 `script_block` 指到的是**別段的劇本**。

**這一輪不改。** 現行的眼魔洞穴 checkpoint 走得通，改欄位是行為變更，要先確認
那條路徑實際吃的是哪一個 block。列進 `SEG-10` 拆段時一併驗。

## 四、量錯一次，記下來

第一版撈地名用 `[A-Z][a-z]{2,}`（首字大寫）——**十個 block 全部回空**。
原作的敘述**整段都是大寫**，撈不到不是因為沒有文字。

正對照（拿已知有文字的 `ECL2/0x03` 跑同一支）當場戳破：261 條候選，
第三條就是 `YOU ARE ENTERING THE FOUL SMELLING, SLIME COVERED`。

`cmd/ecl-block-text` 的檔頭把這件事寫成警告。

## 五、還沒做的（`SEG-03` 的後半）

段落註冊表的另外三欄——**進入方式、正常結束條件、產出快照**——要等 `SEG-04`
（統一 `-segment` 旗標）與 `SEG-11`（快照交接）才填得出來。目前有的是
id、標籤、地圖、進出邊，已經足夠開始拆段。
