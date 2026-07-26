# Curse-of-the-Azure-Bonds SSI 青色枷的詛咒 中文化 & remake

## GOLDEN_BOX 金盒子系統重製評估

@./GOLDEN_BOX_RE.md


## 原始目標

1. 我想透過反組譯還原青色枷的詛咒遊戲引擎, 建立類似 scummvm-like VM bytes code, 建立冬之魔引擎  
 - 希望在 golang /ebitain 環境下可以還原執行
 - 音樂 / 音效 都要還原
2. 基於先前 VM bytecode引擎 完成 中文化, 讓界面與 message 都能用中文顯示, 讓經典用中文說明
3. 以前遊戲需要參考遊戲手札才能知道劇情, 手札都紀錄在說明書內, 希望這次 remake能把手札訊息一起放入遊戲 不用在查詢手冊
4. 中文說明書也要整理成 markdown 保存中文金盒子遊戲歷史
5. 可以上網先蒐集相關資料 遊戲畫面作為 orcale

請先評估建立 PLAN.md , 簡易的工作請安排 subagent 執行


## 工作方式：SDD

反組譯 → 收攏成規格（`docs/spec/`）→ 才實作。
**只有標 READY 的規格可以動手。** 目前刻意還沒開始寫引擎本體。

每一輪都要：更新 markdown → **清掉被推翻的斷言** → commit + push → 更新 CONTEXT.md 現況。

# 遊戲 image
- @./'curseoftheazurebonds.zip'

# github repo
- https://github.com/wicanr2/Curse-of-the-Azure-Bonds-cht.git

# 遊戲手冊
- @./珍020-青色枷的詛咒.rar
- @Curse-of-the-Azure-Bonds_Manual_DOS_EN.pdf
- @Curse-of-the-Azure-Bonds_Misc_DOS_EN_Adventurers-Journal.pdf
- @Curse-of-the-Azure-Bonds_Misc_DOS_EN_Clue-Book.pdf

# 工作目錄
 - @./workplace

# 參考專案 
 - @~/cht/daemon_winter (冬之魔 中文化)

# 反組譯工具
 - @~/.claude/knowledge-base/retro-cht
 - @~/ida_94_official/dist
