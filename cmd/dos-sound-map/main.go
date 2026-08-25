// Command dos-sound-map 在 **DOS** 版找出音效常式與那張音效描述子表，並和
// PC-98 逐格對起來。
//
// ★ 存在的理由：音效的語意（哪一格是揮擊、哪一格是法術沒中）全部來自 PC-98 的
// Borland 除錯符號，而**行為 oracle 是 DOS**。DOS 沒有符號，所以那一層語意在
// DOS 這一側原本是空的——spec 1186 的每一條結論都還沒有在 oracle 上驗過。
//
// 認法**不靠符號，靠形狀**，兩個互相獨立的訊號：
//
//	呼叫點分佈  哪些 overlay 各叫幾次。SOUNDFX 的分佈跨 8 個模組，
//	            是一組很難撞號的指紋。
//	表的位置    描述子在資料段連續排列，順序就是選擇子順序。
//
// 兩個訊號各自認出來的名字要一致才算數。只有分佈的話，兩個分佈相同的效果
// （例如都只在 overlay-13 出現兩次）分不開；只有位置的話，整張表平移一格
// 也會**每一列都看起來合理**。
//
// ⚠ 不要用 far-call 對照表：它只收 IDA 認成程式碼的呼叫點，也不涵蓋常駐
// （spec 1186 就是被這個坑咬過）。這裡兩邊都是位元組直掃。
//
// 用法：
//
//	go run ./cmd/dos-sound-map -output docs/audit/dos-sound-map.md
package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98sfx"
)

// distribution 是「模組 → 呼叫點數」。
type distribution map[string]int

func (d distribution) total() int {
	sum := 0
	for _, count := range d {
		sum += count
	}
	return sum
}

func (d distribution) String() string {
	modules := make([]string, 0, len(d))
	for module := range d {
		modules = append(modules, module)
	}
	sort.Strings(modules)
	parts := make([]string, 0, len(modules))
	for _, module := range modules {
		parts = append(parts, fmt.Sprintf("%s×%d", module, d[module]))
	}
	return strings.Join(parts, "、")
}

// difference 是兩個分佈的對稱差；0 表示逐模組完全相同。
func difference(left, right distribution) int {
	keys := map[string]bool{}
	for key := range left {
		keys[key] = true
	}
	for key := range right {
		keys[key] = true
	}
	total := 0
	for key := range keys {
		delta := left[key] - right[key]
		if delta < 0 {
			delta = -delta
		}
		total += delta
	}
	return total
}

// module 是一個掃描面：overlay 或常駐。
type module struct {
	name string
	data []byte
}

func loadModules(root, resident string) []module {
	modules := make([]module, 0, 40)
	entries, _ := os.ReadDir(filepath.Join(root, "overlays"))
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, "overlay-") || !strings.HasSuffix(name, ".bin") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, "overlays", name))
		if err != nil {
			continue
		}
		modules = append(modules, module{name: strings.TrimSuffix(name, ".bin"), data: data})
	}
	sort.Slice(modules, func(left, right int) bool { return modules[left].name < modules[right].name })
	// 常駐排最後：報表裡它是「不是 overlay 的那一塊」。
	if data, err := os.ReadFile(filepath.Join(root, resident)); err == nil {
		modules = append(modules, module{name: tooltext.Text("h.4671035ec213"), data: data})
	}
	return modules
}

// callSite 是一處 far call 與它前面推的描述子位址（沒有就是 −1）。
type callSite struct {
	module     string
	ea         int
	descriptor int
}

// farCalls 找出所有打到 `segment:offset` 的 `call far`，並往前找描述子。
func farCalls(modules []module, segment, offset uint16) []callSite {
	pattern := []byte{0x9A, byte(offset), byte(offset >> 8), byte(segment), byte(segment >> 8)}
	sites := make([]callSite, 0, 64)
	for _, item := range modules {
		for at := 0; ; {
			found := bytes.Index(item.data[at:], pattern)
			if found < 0 {
				break
			}
			ea := at + found
			sites = append(sites, callSite{module: item.name, ea: ea, descriptor: descriptorAt(item.data, ea)})
			at = ea + 1
		}
	}
	return sites
}

// descriptorAt 往前 20 bytes 找 `push word [位址]`。
func descriptorAt(data []byte, ea int) int {
	low := ea - 20
	if low < 0 {
		low = 0
	}
	for index := ea - 3; index >= low; index-- {
		if data[index] == 0xFF && data[index+1] == 0x36 {
			return int(binary.LittleEndian.Uint16(data[index+2:]))
		}
	}
	return -1
}

// candidate 是一個 far-call 目標與它的呼叫點分佈。
type candidate struct {
	segment, offset uint16
	dist            distribution
}

// scanTargets 統計每個 far-call 目標的呼叫點分佈。**只掃 overlay**：
// 常駐呼叫自己的段用的是同一套編碼，但它的分佈不該影響「哪個目標是 SOUNDFX」
// 的判斷——指紋是拿 overlay 側比的。
func scanTargets(modules []module) map[uint32]*candidate {
	targets := map[uint32]*candidate{}
	for _, item := range modules {
		if item.name == tooltext.Text("h.4671035ec213") {
			continue
		}
		for ea := 0; ea+4 < len(item.data); ea++ {
			if item.data[ea] != 0x9A {
				continue
			}
			offset := binary.LittleEndian.Uint16(item.data[ea+1:])
			segment := binary.LittleEndian.Uint16(item.data[ea+3:])
			key := uint32(segment)<<16 | uint32(offset)
			if targets[key] == nil {
				targets[key] = &candidate{segment: segment, offset: offset, dist: distribution{}}
			}
			targets[key].dist[item.name]++
		}
	}
	return targets
}

// judgedPlatformGaps 是「兩版對不上、但已經判定是**平台差異**而不是漏掉」的音效。
//
// ★ 存在的理由與 `cmd/sound-trigger-compare` 的 `judgedGaps` 相同：判過的結論要留在
// 產生報表的程式碼裡。只留一個「1」的話，下一輪會把它當待辦重查一次。
var judgedPlatformGaps = map[string]string{
	"CRASHFX": tooltext.Text("h.b8d284553d9e") +
		tooltext.Text("h.3629d25a5aa1") +
		tooltext.Text("h.a2dbcbee0b7c"),
}

func main() {
	dosRoot := flag.String("dos", "workplace/re-sweep/dos", tooltext.Text("h.00e3ccc00e4c"))
	dosResident := flag.String("dos-resident", "START.EXE", tooltext.Text("h.a5242f33d486"))
	pc98Root := flag.String("pc98", "workplace/re-sweep/pc98", tooltext.Text("h.a51d2488a6c1"))
	pc98Resident := flag.String("pc98-resident", "PC98-GAME.EXE", tooltext.Text("h.6709b9cff31d"))
	output := flag.String("output", "", tooltext.Text("h.78eb014c7900"))
	flag.Parse()

	pc98Modules := loadModules(*pc98Root, *pc98Resident)
	dosModules := loadModules(*dosRoot, *dosResident)
	if len(pc98Modules) == 0 || len(dosModules) == 0 {
		log.Fatal(tooltext.Text("h.b1407f24e720"))
	}

	// PC-98 那一側：段 0893h 位移 0000h 是 SOUNDFX（符號表直接讀出來的）。
	const pc98SoundSegment, pc98SoundOffset = 0x0893, 0x0000
	pc98Sites := farCalls(pc98Modules, pc98SoundSegment, pc98SoundOffset)
	pc98Overlay := distribution{}
	pc98ByEffect := map[string]distribution{}
	pc98Resident2 := distribution{}
	for _, site := range pc98Sites {
		info, known := pc98sfx.SelectorForDescriptor(site.descriptor)
		name := tooltext.Format("h.1a9a9b2900a6", site.descriptor)
		if known {
			name = info.Symbol
		}
		if site.module == tooltext.Text("h.4671035ec213") {
			pc98Resident2[name]++
			continue
		}
		pc98Overlay[site.module]++
		if pc98ByEffect[name] == nil {
			pc98ByEffect[name] = distribution{}
		}
		pc98ByEffect[name][site.module]++
	}

	// DOS 那一側：目標未知，用 PC-98 的 overlay 分佈當指紋去找。
	targets := scanTargets(dosModules)
	ranked := make([]*candidate, 0, len(targets))
	for _, item := range targets {
		ranked = append(ranked, item)
	}
	sort.Slice(ranked, func(left, right int) bool {
		leftScore := difference(ranked[left].dist, pc98Overlay)
		rightScore := difference(ranked[right].dist, pc98Overlay)
		if leftScore != rightScore {
			return leftScore < rightScore
		}
		return ranked[left].dist.total() > ranked[right].dist.total()
	})
	best := ranked[0]

	var report strings.Builder
	fmt.Fprint(&report, tooltext.Format("h.1db040ca5da2"))
	fmt.Fprint(&report, tooltext.Format("h.7d222ffc4098"))
	fmt.Fprint(&report, tooltext.Text("h.44ee0db72be1")+
		tooltext.Text("h.7a290a0dcefa"))

	fmt.Fprint(&report, tooltext.Format("h.bfc28e6a742f"))
	fmt.Fprint(&report, tooltext.Format("h.f11e338d1195", len(pc98Overlay), pc98Overlay.total()))
	fmt.Fprintf(&report, "> %s\n\n", pc98Overlay)
	fmt.Fprint(&report, tooltext.Format("h.9f71c7628809", len(targets)))
	fmt.Fprint(&report, tooltext.Format("h.5fb769376c23"))
	for index := 0; index < 5 && index < len(ranked); index++ {
		item := ranked[index]
		fmt.Fprintf(&report, "| %d | `%04X:%04X` | %d | %d | %s |\n",
			index+1, item.segment, item.offset, difference(item.dist, pc98Overlay),
			item.dist.total(), truncate(item.dist.String(), 90))
	}
	runnerUp := 0
	if len(ranked) > 1 {
		runnerUp = difference(ranked[1].dist, pc98Overlay)
	}
	fmt.Fprintf(&report, tooltext.Text("h.806660e43527")+
		tooltext.Text("h.abeffed39b5b"),
		best.segment, best.offset, difference(best.dist, pc98Overlay), runnerUp)

	// DOS 描述子。
	dosSites := farCalls(dosModules, best.segment, best.offset)
	dosByDescriptor := map[int]distribution{}
	dosResidentByDescriptor := map[int]int{}
	dosUnresolved := 0
	for _, site := range dosSites {
		if site.descriptor < 0 {
			dosUnresolved++
			continue
		}
		if site.module == tooltext.Text("h.4671035ec213") {
			dosResidentByDescriptor[site.descriptor]++
			continue
		}
		if dosByDescriptor[site.descriptor] == nil {
			dosByDescriptor[site.descriptor] = distribution{}
		}
		dosByDescriptor[site.descriptor][site.module]++
	}
	descriptors := make([]int, 0, len(dosByDescriptor))
	for address := range dosByDescriptor {
		descriptors = append(descriptors, address)
	}
	for address := range dosResidentByDescriptor {
		if dosByDescriptor[address] == nil {
			descriptors = append(descriptors, address)
		}
	}
	sort.Ints(descriptors)
	if len(descriptors) == 0 {
		log.Fatal(tooltext.Text("h.c62d847e6923"))
	}

	// 表的基底不是用猜的。 兩版的描述子都是 `基底 ＋ 選擇子×2`，所以只要定出
	// 一格就定出整張表；問題是**平移一格之後每一列還是說得通**（位址仍然連續、
	// 名字仍然一一對應），只有分佈會整排錯開。
	//
	// ⚠ 所以這裡把**每一個觀察到的描述子**都試一次當基底，用「有幾格的 overlay
	// 分佈與 PC-98 逐模組相同」計分，取最高分並印出與第二名的差距。取最低位址
	// 當基底會在「最低那一格剛好沒有 DOS 呼叫點」時整排偏掉，而且看不出來。
	pc98Base := pc98sfx.HaltDescriptor
	dosBase, bestBaseScore, secondBaseScore := pickBase(descriptors, dosByDescriptor, pc98ByEffect)
	// 對不上一半就不要印一張看起來很合理的表。
	if bestBaseScore*2 < len(descriptors) {
		log.Fatal(tooltext.Format("h.9e487de640f0", dosBase, bestBaseScore, len(descriptors)))
	}

	fmt.Fprint(&report, tooltext.Format("h.b0736c40c381"))
	fmt.Fprintf(&report, tooltext.Text("h.7971e279e5fa")+
		tooltext.Text("h.01b65629678e"),
		dosBase, pc98Base, pc98Base-dosBase)
	fmt.Fprintf(&report, tooltext.Text("h.65ca4d3dd247")+
		tooltext.Text("h.8e8bc709fa8a")+
		tooltext.Text("h.4243c2b05ae2")+
		tooltext.Text("h.1aaa32e78ee1"),
		dosBase, bestBaseScore, secondBaseScore)
	fmt.Fprint(&report, tooltext.Text("h.0de1cfc05f9b")+
		tooltext.Text("h.2d6dce111721")+
		tooltext.Text("h.a90dc78e0250"))
	fmt.Fprint(&report, tooltext.Format("h.347975fec7ef"))
	fmt.Fprintf(&report, "|---|---|---|---|---|---|\n")

	agree, disagree, dosOnly := 0, 0, 0
	judged := 0
	for _, info := range pc98sfx.Selectors() {
		dosAddress := dosBase + (info.Descriptor - pc98Base)
		dosDist := dosByDescriptor[dosAddress]
		pc98Dist := pc98ByEffect[info.Symbol]
		if dosDist == nil {
			dosDist = distribution{}
		}
		if pc98Dist == nil {
			pc98Dist = distribution{}
		}
		verdict := "✅"
		switch {
		case dosDist.total() == 0 && pc98Dist.total() == 0:
			verdict = tooltext.Text("h.7299ee1f3676")
		case difference(dosDist, pc98Dist) == 0:
			agree++
		case dosDist.total() == 0:
			if reason, ok := judgedPlatformGaps[info.Symbol]; ok {
				verdict = tooltext.Text("h.fed8783c191e") + reason
				judged++
				break
			}
			verdict = tooltext.Text("h.aea12fe06c96")
			disagree++
		case pc98Dist.total() == 0:
			verdict = tooltext.Text("h.5651fb26d0c3")
			dosOnly++
		default:
			verdict = tooltext.Format("h.b1184e3b06c3", difference(dosDist, pc98Dist))
			disagree++
		}
		fmt.Fprintf(&report, "| `%04Xh` | `%04Xh` | %s | %s | %s | %s |\n",
			info.Descriptor, dosAddress, info.Symbol,
			blank(dosDist.String()), blank(pc98Dist.String()), verdict)
	}
	fmt.Fprintf(&report, "\n")

	fmt.Fprint(&report, tooltext.Format("h.13c83a8a875e"))
	fmt.Fprint(&report, tooltext.Format("h.e6a2acd59bfe", agree))
	fmt.Fprint(&report, tooltext.Format("h.e2482603babb", disagree+dosOnly))
	fmt.Fprint(&report, tooltext.Format("h.fcaf5cdee2d1", judged))
	fmt.Fprint(&report, tooltext.Format("h.6ca7133d0e2e", len(dosSites)-residentCount(dosSites)))
	fmt.Fprint(&report, tooltext.Format("h.384460dd29fe", residentCount(dosSites)))
	fmt.Fprint(&report, tooltext.Format("h.151f324913c7", dosUnresolved))

	// 常駐那一半：兩版都只放引擎內務那三個。
	fmt.Fprint(&report, tooltext.Format("h.b450823f6dc0"))
	fmt.Fprint(&report, tooltext.Format("h.26f435678c9e"))
	for _, info := range pc98sfx.Selectors() {
		dosAddress := dosBase + (info.Descriptor - pc98Base)
		dosCount := dosResidentByDescriptor[dosAddress]
		pc98Count := pc98Resident2[info.Symbol]
		if dosCount == 0 && pc98Count == 0 {
			continue
		}
		fmt.Fprintf(&report, "| %s | %d | %d |\n", info.Symbol, dosCount, pc98Count)
	}
	fmt.Fprint(&report, tooltext.Text("h.f141c38802a0")+
		tooltext.Text("h.452d76a1be29")+
		tooltext.Text("h.b993b4afad77"))

	text := report.String()
	if *output == "" {
		fmt.Print(text)
	} else if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "dos-target=%04X:%04X margin=%d..%d agree=%d disagree=%d dos-sites=%d pc98-sites=%d\n",
		best.segment, best.offset, difference(best.dist, pc98Overlay), runnerUp,
		agree, disagree+dosOnly, len(dosSites), len(pc98Sites))
}

func residentCount(sites []callSite) int {
	count := 0
	for _, site := range sites {
		if site.module == tooltext.Text("h.4671035ec213") {
			count++
		}
	}
	return count
}

func blank(text string) string {
	if text == "" {
		return "—"
	}
	return text
}

func truncate(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + "…"
}

// scoreBase 算「拿 base 當錨點時，有幾格的 overlay 分佈與 PC-98 逐模組相同」。
//
// ⚠ DOS 那一格沒有呼叫點就**不計分也不扣分**：那種格子在任何錨點下都長一樣，
// 讓它參與計分只會把所有候選拉平。
func scoreBase(base int, dosByDescriptor map[int]distribution, pc98ByEffect map[string]distribution) int {
	matched := 0
	for _, info := range pc98sfx.Selectors() {
		dosDist := dosByDescriptor[base+(info.Descriptor-pc98sfx.HaltDescriptor)]
		pc98Dist := pc98ByEffect[info.Symbol]
		if dosDist == nil || pc98Dist == nil || dosDist.total() == 0 {
			continue
		}
		if difference(dosDist, pc98Dist) == 0 {
			matched++
		}
	}
	return matched
}

// pickBase 回傳最佳錨點、它的分數，以及第二名的分數。
//
// ★ 第二名的分數是**要印出來給人看的**：整張表平移一格之後位址仍然連續、名字
// 仍然一一對應，光看表看不出偏移。兩個分數的差距才是「沒有偏移」的證據。
func pickBase(descriptors []int, dosByDescriptor map[int]distribution, pc98ByEffect map[string]distribution) (int, int, int) {
	ranking := append([]int(nil), descriptors...)
	sort.Slice(ranking, func(left, right int) bool {
		leftScore := scoreBase(ranking[left], dosByDescriptor, pc98ByEffect)
		rightScore := scoreBase(ranking[right], dosByDescriptor, pc98ByEffect)
		if leftScore != rightScore {
			return leftScore > rightScore
		}
		return ranking[left] < ranking[right]
	})
	if len(ranking) == 0 {
		return 0, 0, 0
	}
	second := 0
	if len(ranking) > 1 {
		second = scoreBase(ranking[1], dosByDescriptor, pc98ByEffect)
	}
	return ranking[0], scoreBase(ranking[0], dosByDescriptor, pc98ByEffect), second
}
