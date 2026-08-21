# `29h ENCOUNTER MENU` 的旁白有沒有接上譯文

由 `cmd/ecl-encounter-text` 產生，不要手改。

`29h` 帶三句旁白（運算元 9、10、11），原作依距離挑一句。remake 目前只取
**第一句非空的**（`internal/ecl/runtime.go` 的 `0x29`）。所以缺口分兩種：

- **演得到但沒接**：玩家真的會看到原文，是中文化缺口。
- **演不到且沒接**：remake 現在演不到那一句，是還原度的缺口。

⚠ 一句一句比對 `all_contains`，不是把三句合起來比——合起來比會讓沒接上的
那兩句被第一句的規則蓋掉。

⚠ `cmd/ecl-text-coverage` 的分母裡**沒有這個 opcode**，所以那份報告的
「未接上 0 群」與這一份不衝突：那裡從來沒數過這批文字。

| 段 | 位移 | 旁白 | remake 演得到 | 規則 | 原文 |
|---|---|---:|---|---|---|
| `ECL3/0x10` | `0x0d99` | 9 | 是 | `yulash.shambling-mounds-drag-cleric` | SHAMBLING MOUNDS ATTEMPT TO DRAG A CLERICS BODY AWAY. |
| `ECL3/0x10` | `0x0d99` | 10 | 否 | **沒有規則** | SHAMBLING MOUNDS SEEM RELUCTANT TO TOUCH AN OBJECT NEAR A DEAD C… |
| `ECL3/0x10` | `0x0d99` | 11 | 否 | **沒有規則** | YOU SEE A LARGE GROUP OF SHAMBLING MOUNDS IN THE DISTANCE. |
| `ECL3/0x10` | `0x10c2` | 9 | 是 | `myth-drannor.rubble-scavengers` | A FILTHY GROUP HAS BEEN PICKING THROUGH THE RUBBLE.  THEY STARE … |
| `ECL3/0x10` | `0x10c2` | 10 | 否 | **沒有規則** | THE PEOPLE ARE DRESSED IN RAGS. |
| `ECL3/0x10` | `0x10c2` | 11 | 否 | **沒有規則** | YOU SEE A GROUP OF PEOPLE IN THE DISTANCE. |
| `ECL3/0x10` | `0x184c` | 9 | 是 | `yulash.red-plume-pit-refusal` | A RED PLUME GUARD GROWLS, 'NOBODY'S GOING TO MAKE US GO BACK INT… |
| `ECL3/0x10` | `0x184c` | 10 | 否 | **沒有規則** | THESE RED PLUMES LOOK SOMEWHAT SHIFTY AND DIRTY. |
| `ECL3/0x10` | `0x184c` | 11 | 否 | **沒有規則** | YOU NOTICE SOME RED PLUME GUARDS AHEAD. |
| `ECL3/0x10` | `0x19fe` | 9 | 是 | `yulash.dirty-robed-people-appear` | SOME DIRTY ROBED PEOPLE APPEAR AHEAD. |
| `ECL3/0x10` | `0x19fe` | 10 | 否 | **沒有規則** | YOU SEE CLERICS WITH THE MOUTH IN HAND SIGIL OF MOANDER. |
| `ECL3/0x10` | `0x19fe` | 11 | 否 | **沒有規則** | SOME VERY SKITTISH CLERICS OF MOANDER EYE YOU SUSPICIOUSLY. |
| `ECL3/0x11` | `0x11a8` | 9 | 是 | `pit.cultists-push-you-out` | CULTISTS TRY TO PUSH YOU OUT THE DOOR. |
| `ECL3/0x11` | `0x11a8` | 10 | 否 | **沒有規則** | THE FANATICS OF MOADER BESEECH YOU TO GO BACK TO THE CORRIDOR. |
| `ECL3/0x11` | `0x11a8` | 11 | 否 | **沒有規則** | A GROUP OF MOADERITES POINT BACK TO THE DOOR. |
| `ECL3/0x12` | `0x03ae` | 9 | 是 | `pit.vegepygmies-retreat` | YOU COME UPON SOME VEGEPYGMIES. THEY START TO BACK AWAY, POINTIN… |
| `ECL3/0x12` | `0x03ae` | 10 | 否 | **沒有規則** | SOME VEGEPYGMIES SEEM TO BE MOTIONING YOU BACK OUT INTO THE CORR… |
| `ECL3/0x12` | `0x03ae` | 11 | 否 | **沒有規則** | YOU SEE VEGEPYGMIES IN THE DISTANCE. |
| `ECL3/0x12` | `0x0523` | 9 | 是 | `pit.shambling-mounds-push-corridor` | SOME SHAMBLING MOUNDS ATTEMPT TO PUSH YOU BACK INTO THE CORRIDOR… |
| `ECL3/0x12` | `0x0523` | 10 | 否 | **沒有規則** | SHAMBLING MOUNDS ATTEMPT TO GET AWAY. |
| `ECL3/0x12` | `0x0523` | 11 | 否 | **沒有規則** | YOU SEE SHAMBLING MOUNDS IN THE DISTANCE. |
| `ECL4/0x20` | `0x04e4` | 9 | 是 | `zhentil.keep-troops-stare-down` | ZHENTIL KEEP TROOPS STARE YOU DOWN. |
| `ECL4/0x20` | `0x04e4` | 10 | 否 | **沒有規則** | SOLDIERS ARE APPROACHING THE PARTY. |
| `ECL4/0x20` | `0x04e4` | 11 | 否 | **沒有規則** | YOU SEE SOLDIERS. |
| `ECL4/0x20` | `0x06a0` | 9 | 是 | `zhentil.mages-and-bodyguards-upon-you` | MAGES AND THEIR BODYGUARDS ARE UPON YOU. |
| `ECL4/0x20` | `0x06a0` | 10 | 否 | **沒有規則** | YOU SEE MAGES AND SOLDIERS. |
| `ECL4/0x20` | `0x06a0` | 11 | 否 | **沒有規則** | YOU SEE MAGES. |
| `ECL4/0x20` | `0x0814` | 9 | 是 | `zhentil.priests-of-bane-upon-you` | PRIESTS OF BANE AND THEIR BODYGUARDS ARE UPON YOU. |
| `ECL4/0x20` | `0x0814` | 10 | 否 | **沒有規則** | SOME PRIESTS OF BANE APPROACH. |
| `ECL4/0x20` | `0x0814` | 11 | 否 | **沒有規則** | YOU SEE CLERICS. |
| `ECL4/0x20` | `0x189f` | 9 | 是 | `zhentil.fzoul-summons` | YOU MUST COME WITH US TO SEE OUR LORD FZOUL CHEMBRYL! |
| `ECL4/0x21` | `0x0a9f` | 9 | 是 | `zhentil.unwashed-minions-of-bane` | YOU ARE FACE TO FACE WITH THE UNWASHED MINIONS OF BANE. |
| `ECL4/0x21` | `0x0a9f` | 10 | 否 | **沒有規則** | THE PRIESTS APPROACH CAUTIOUSLY. |
| `ECL4/0x21` | `0x0a9f` | 11 | 否 | **沒有規則** | YOU SEE A PATROL OF PRIESTS AND TROOPS IN THE DISTANCE. |
| `ECL5/0x31` | `0x0133` | 9 | 是 | `hap.dark-elf-patrol-arrives` | A DARK ELF PATROL ARRIVES |
| `ECL5/0x32` | `0x01b7` | 9 | 是 | `encounter.salamanders-investigate` | THE SALAMANDERS COME UP TO INVESTIGATE |
| `ECL5/0x32` | `0x0624` | 9 | 是 | `encounter.patrol-watches` | THE PATROL WATCHES YOU |
| `ECL6/0x40` | `0x038e` | 9 | 是 | `encounter.insects-watch` | INSECTS WATCH YOU |
| `ECL6/0x40` | `0x038e` | 10 | 否 | **沒有規則** | INSECTS PREPARE |
| `ECL6/0x40` | `0x038e` | 11 | 否 | **沒有規則** | INSECTS APPEAR |
| `ECL6/0x40` | `0x0d40` | 9 | 是 | `myth-drannor.red-plume.gesture` | HE MAKES A GESTURE OF FRIENDSHIP |
| `ECL6/0x40` | `0x0d40` | 10 | 否 | **沒有規則** | A RED PLUME APPROACHES IN A FRIENDLY MANNER |
| `ECL6/0x40` | `0x0d40` | 11 | 否 | **沒有規則** | YOU SPOT A LONE RED PLUME |
| `ECL6/0x40` | `0x15e0` | 9 | 是 | `myth-drannor.clan-figure.greeting` | A FIGURE APPEARS FROM THE SHADOWS. 'HAIL BONDED ONES!' |
| `ECL6/0x42` | `0x0347` | 9 | 是 | `myth-drannor.spotted-by-patrol` | YOU ARE SPOTTED BY A PATROL |
| `ECL6/0x42` | `0x0347` | 10 | 否 | **沒有規則** | A PATROL ADVANCES TOWARD YOU |
| `ECL6/0x42` | `0x0347` | 11 | 否 | **沒有規則** | A PATROL CONFRONTS YOU |
| `ECL6/0x42` | `0x1611` | 9 | 是 | `myth-drannor.outer.rakshasa-residence` | A RAKSHASA RESIDES HERE IN SPLENDOR |

## 摘要（語系 `zh-TW`）

| 項目 | 數 |
|---|---:|
| `29h` 的處數 | 20 |
| 旁白總句數 | 48 |
| remake 演得到的句數 | 20 |
| **演得到但沒接上譯文** | **0** |
| remake 演不到的句數 | 28 |
| 其中沒接上譯文 | 28 |
