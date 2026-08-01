# 第四百一十輪：PC-98 怪物效果載入與提朗瑟克斯偵測隱形（READY）

## 範圍與結論

本規格只關閉 `MON*SPC` 九 byte 紀錄的 runtime linked-list 邊界，以及
提朗瑟克斯效果 `18h` 的「偵測隱形」戰鬥投影。它不代表 `4Fh／6Ah／70h／
84h／87h` 已解，也不代表提朗瑟克斯的閃電、魔法抗性、AI、死亡演出或整場
終戰 fidelity 已完成。

- `EFFECTREC` 是 9 bytes：kind `+0`、word `+1`、byte `+3`、byte `+4`、
  runtime next pointer `+5..+8`。
- `MON6SPC` 保存的 bytes `+5..+8` 是舊 linked-list pointer，不是能力參數。
- PC-98 `LOADMONSTER` 每次配置 9 bytes、複製整筆紀錄，再把新 node 的
  `+5／+7` 兩個 words 清零並依序配置下一個 node。它不改寫 byte `+4`。
- 因此 template 中 byte `+4=0` 不能被 remake 解讀成「怪物天生效果停用」。
  Battle 以作品中立的 `Innate` 來源標記讓 `MON*SPC` 效果可運作；角色後天
  effect 原有的 `Active` 生命週期仍保留。
- 提朗瑟克斯 `MON6CHA 47h` 的六筆 `MON6SPC` 皆進入正常最終戰，其中
  kind `18h` 依既有效果表為 Detect Invisibility。spec 417 已以 PC-98
  `CHECKTARGET` 與兩個原始 handler 修正本輪過度斷言：`18h` 只抵消目標
  `19h` 的隱藏與 +4 AC；`47h` 仍無條件生效。

## 原始資料

| 輸入 | SHA-256 |
|---|---|
| DOS image `curseoftheazurebonds.zip` | `c98698a6271c17177dfdb27f34b0389b7d34f58ef206e92575393f4655f5b26d` |
| PC-98 `GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` |
| PC-98 `GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` |
| overlay 12 `EFFPROCS` | `7e1e8394652b02986da00944769331e0be7db7fb8476054fbe0369dbca05b9d9` |
| overlay 16 `LOADSAVE` | `62146faf1a984170d72ea01f30e3c077d2298b86310436aec39a64cf2d7fb050` |
| overlay 23 `EFFECTS` | `a3ea0d9528be57a92c33fc345baa3e27eef375c84822afba0cfbb141c2faabc9` |

提朗瑟克斯六筆 raw records：

```text
18 0000 FF 00 04001B1D
70 0000 FF 00 0B001A1D
4F 0000 FF 00 02001A1D
6A 0000 FF 00 0900191D
84 0000 FF 00 0000191D
87 0000 FF 00 00000000
```

最後四 bytes 形成載入前的舊鏈結值；`LOADMONSTER` 在 runtime 副本清除並
重建它們，故不得把 `04001B1D` 等數字命名成法術強度、目標或傷害。

## IDA Pro 9.4 證據

可重現入口是
`scripts/ida/pc98_monster_affect_loader_audit.idc`。原始 binary、TPOV 與
基準 database 唯讀，只在 Docker `/work` 分析 code-only 副本。

Borland symbols：

- `EFFPROCS 008B:2ED4 INITEFFPROX`；
- `LOADSAVE 00DC:3AA8 LOADMONSTER`；
- `EFFECTS 013E:00C9 CALLEFFECT`、`013E:13D7 ADDEFFECT`；
- type `EFFECTREC` index 838，size `0009h`。

`ADDEFFECT` exact writes：

```text
013E:145F  mov es:[di],al      ; +0
013E:1468  mov es:[di+1],ax    ; +1 word
013E:1472  mov es:[di+4],al    ; +4 byte
013E:147C  mov es:[di+3],al    ; +3 byte
013E:144F  mov es:[di+5],0
013E:1455  mov es:[di+7],0
```

`LOADMONSTER` overlay-local `3C2Fh..3C87h` 每次複製 9 bytes，於 `3C4Ch／
3C50h` 清除 `+5／+7`，以 `ADD [bp-8],9` 前進來源，再由 `+5` 配置下一
node。這是 linked list 的直接資料流證據，推論等級 `exact`。

`CALLEFFECT` 以傳入 byte 乘 4 後呼叫 `ds:A040h` 的 far-procedure table；
IDA xref graph 找到 overlay 23 內五個 caller：`0184h／03F5h／0F9Ch／
104Eh／24DAh`。`INITEFFPROX` 對 `18h／4Fh／6Ah／70h／87h` 的 table slot
有直接 writes；第 412 輪另在 overlay 22 `SPELLS` 初始化找到 `84h` slot
writer，並以 typed TPOV entry resolver 關閉六筆 handler。

第 412 輪證明 resident control 的五 byte entry stub 是
`CD 3F + handler-local u16 + flags`，並將 raw far pointer 經 segment／stub
index 解析到指定 overlay local handler。舊的 `hypothesis／unknown` 位址
斷言由 spec 412 supersede；原始 pointer 與位址空間仍須並列保存。

## 實作與驗證

- `combat.MonsterAffect.Innate` 記錄來源，不覆寫 raw byte `+4`。
- 隱形、加速／緩慢與定身判定共用 `Active || Innate`；只有從
  `BuildEnemiesWithAffects` 載入的 `MON*SPC` 會自動標成 innate。
- `MonsterCanDetectInvisible` 只認 kind `18h`；其精確可見性與命中投影由
  spec 417 supersede 本規格原先把 `19h／47h` 合併處理的斷言。
- `TestResolveAttackInnateDetectInvisibleBypassesInvisibilityACBonus` 鎖定
  命中邊界，並以非 innate、inactive 反例避免所有 `18h` 無條件生效。
- `TestRealPlayerPathStandingStoneToBurialGlen` 現在載入真實 `MON6SPC.DAX`；
  正常 Standing Stone 長路徑抵達 37 人終戰，確認 `47h` 唯一 fighter 有
  六筆 innate effects 與可運作的 `18h`，之後仍由正式 scheduler 勝利並進入
  `PROGRAM 8`。

## 尚未完成

- byte `+4` 在所有 effect phase 的完整正式名稱；目前只證明它不能當
  `MON*SPC` template 的總啟用 gate。
- `4Fh／6Ah／70h／84h／87h` handler 已由 spec 412 靜態關閉；仍缺其完整
  runtime boundary、AI、頻率與動態演出。
- 提朗瑟克斯閃電、魔法抗性、三神器修正、AI、聲音及 DOS 動態演出。
- HIGH PRIEST 的 `09h／0Ah`、MARGOYLE 的 `77h` 特殊能力。
