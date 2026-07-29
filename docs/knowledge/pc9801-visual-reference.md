# PC-9801《Curse of the Azure Bonds》視覺參考

更新日期：2026-07-29

## 可驗證來源

- [MobyGames PC-98 gallery](https://www.mobygames.com/game/503/curse-of-the-azure-bonds/screenshots/pc98/)
  收錄 25 張 640×400 畫面。
- [冒險／敘事畫面 464680](https://www.mobygames.com/game/503/curse-of-the-azure-bonds/screenshots/pc98/464680/)
  顯示六人 roster、AC／HP、時間、日文敘事與 Return 提示。
- [戰鬥畫面 464695](https://www.mobygames.com/game/503/curse-of-the-azure-bonds/screenshots/pc98/464695/)
  顯示左戰場與右側日文敵人狀態欄。
- [PC-9801.com title capture](https://pc-9801.com/%EF%BD%81%EF%BD%84%EF%BC%86%EF%BD%84-%E3%82%AB%E3%83%BC%E3%82%B9-%E3%82%AA%E3%83%96-%E3%82%A2%E3%82%B8%E3%83%A5%E3%82%A2%E3%83%9C%E3%83%B3%E3%83%89/)
  確認日本 PC-9801 發行版。

外部畫面只作研究連結，不在 repo 重複散布。若未取得原始 PNG，以下觀察標為
gallery-level，不能冒充逐像素量測。

## 對 remake 的意義

- PC-98 原生 640×400 已有足夠解析度顯示雙位元組文字，但一般敘事、名單、
  數值與戰鬥 HUD 仍採緊密點陣字。這直接反駁「640 寬就把正文放成 24px」
  的早期假設。
- DOS 與 PC-98 都保留 Gold Box 的左視覺區、右 roster/status、下方文字區
  骨架；高解析日文版沒有改造成現代全螢幕對話框。
- CoAB 繁中版因此以粗體倚天 16×15 作主要遊戲字，24px 僅保留給標題與
  特別強調。640×480 多出的高度用於更舒適的行距、分頁和命令提示，而不是
  無條件放大每個字。
- DOS 本機實機圖仍負責邊框材質、區塊幾何與 EGA 素材校準；PC-98 主要負責
  CJK 字級、行距及資訊密度判斷。

## 後續蒐集

- 取得可合法本機執行的 PC-98 media 後，以 Neko Project II 或同級 emulator
  擷取 title、角色建立、第一人稱、事件長文、AREA map、戰鬥、camp、journal。
- 每張保留 emulator、解析度、遊戲狀態與來源；量測 glyph bounding box、
  baseline、行距、每行日文字數及各 panel rectangle。
- 將量測結果寫入共用 engine 的 typography/layout JSON，而不是在 CoAB
  renderer 中散落 magic numbers。

