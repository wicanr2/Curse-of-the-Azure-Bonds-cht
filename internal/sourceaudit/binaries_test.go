package sourceaudit

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// buildProductMagics 是「這是編出來的執行檔」的前幾個位元組。
// ELF 涵蓋本機 build 的產物，Mach-O／PE 涵蓋交叉編出來的。
var buildProductMagics = [][]byte{
	{0x7F, 'E', 'L', 'F'},             // ELF
	{0xFE, 0xED, 0xFA, 0xCE},          // Mach-O 32-bit
	{0xFE, 0xED, 0xFA, 0xCF},          // Mach-O 64-bit
	{0xCA, 0xFE, 0xBA, 0xBE},          // Mach-O universal
	{'M', 'Z'},                        // PE／MZ
}

// skipDirs 是**不歸這個 repo 管**的目錄：本地工作產物、獨立 history、
// 以及 Go 自己的 build cache。
var skipDirs = map[string]bool{
	"workplace":                true,
	"golden-box-remake-engine": true,
	"pc98":                     true,
	".git":                     true,
}

// TestRepositoryTracksNoBuiltBinaries 擋住「`go build ./cmd/x` 的執行檔被
// `git add -A` 收進版控」。
//
// ★ 為什麼要用測試擋，而不是只靠 `.gitignore`。 原本的 `.gitignore` 是**逐一
// 列出每個 cmd 的名字**，而那份清單本質上要靠人記得維護——它漏過至少六次，
// 最後有五個執行檔共 28 MB 進了 repo。**清單會漏，掃描不會**：這條測試直接看
// 檔案的前幾個位元組，新的 cmd 不必登記也擋得住。
//
// ⚠ 這條測試看的是**工作目錄**不是索引：容器裡沒有 git，而且這個 repo 的
// `.git` 在 `workplace/azure-bonds-git`，`git ls-files` 在測試裡跑不起來。
// 所以它自己算一遍「這個路徑會不會被 `.gitignore` 擋掉」——**被擋掉的只記一行、
// 不算失敗**（`go build` 之後根目錄躺著十個執行檔是正常的），沒被擋掉的才失敗。
//
// ⇒ 根目錄那條白名單補上之後，這條測試真正還在守的是**子目錄**：在 `cmd/foo/`
// 底下跑 `go build` 會把執行檔丟在 `cmd/foo/foo`，那裡沒有白名單擋。
func TestRepositoryTracksNoBuiltBinaries(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	ignored := loadIgnoredRootNames(t, root)

	var found, blocked []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		if entry.IsDir() {
			if relative != "." && skipDirs[relative] {
				return fs.SkipDir
			}
			return nil
		}
		if !isBuildProduct(t, path) {
			return nil
		}
		size := int64(0)
		if info, statErr := entry.Info(); statErr == nil {
			size = info.Size()
		}
		line := fmt.Sprintf("%s %.1f MB", relative, float64(size)/(1<<20))
		if ignored[relative] {
			blocked = append(blocked, line)
			return nil
		}
		found = append(found, line)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocked) > 0 {
		t.Logf("`.gitignore` 擋著的 build 產物 %d 個（正常，不算失敗）：%s",
			len(blocked), strings.Join(blocked, "、"))
	}
	for _, item := range found {
		t.Errorf("build 產物沒被 `.gitignore` 擋住：%s", item)
	}
	if len(found) > 0 {
		t.Fatalf("%d 個 build 產物會被 `git add -A` 收進版控；輸出改到 workplace/，"+
			"已經進版控的用 `git rm --cached` 移掉", len(found))
	}
}

// isBuildProduct 只讀前幾個位元組；副檔名不看，因為 Go 的產物本來就沒有副檔名。
func isBuildProduct(t *testing.T, path string) bool {
	t.Helper()
	handle, err := os.Open(path)
	if err != nil {
		return false
	}
	defer handle.Close()
	header := make([]byte, 4)
	read, err := io.ReadFull(handle, header)
	if err != nil && read < 2 {
		return false
	}
	header = header[:read]
	for _, magic := range buildProductMagics {
		if len(header) >= len(magic) && bytes.Equal(header[:len(magic)], magic) {
			// ⚠ `MZ` 兩個位元組太短，純文字檔開頭剛好是 `MZ` 就會誤判。
			// PE 一定有 `PE\0\0`，所以再確認一次。
			if bytes.Equal(magic, []byte{'M', 'Z'}) {
				return looksLikePE(path)
			}
			return true
		}
	}
	return false
}

func looksLikePE(path string) bool {
	payload, err := os.ReadFile(path)
	if err != nil || len(payload) < 0x40 {
		return false
	}
	return bytes.Contains(payload[:min(len(payload), 1024)], []byte{'P', 'E', 0, 0})
}

// loadIgnoredRootNames 算出「根目錄有哪些檔名被 `.gitignore` 擋掉」。
//
// **只認得這個 repo 實際用到的兩種寫法**：`/*` 那條白名單，加上 `!/…` 的例外。
// 沒有 `/*` 就回空的（沒有白名單就沒有「擋住了但還留著」這個類別），
// 寧可誤報也不要漏報。子目錄一律當成沒被擋——白名單管不到那裡。
func loadIgnoredRootNames(t *testing.T, root string) map[string]bool {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		return map[string]bool{}
	}
	blanket := false
	allowed := map[string]bool{}
	for _, line := range strings.Split(string(payload), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case line == "/*":
			blanket = true
		case strings.HasPrefix(line, "!/"):
			allowed[strings.TrimPrefix(line, "!/")] = true
		}
	}
	if !blanket {
		return map[string]bool{}
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return map[string]bool{}
	}
	ignored := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if allowed[name] || (allowed["*.md"] && strings.HasSuffix(name, ".md")) {
			continue
		}
		ignored[name] = true
	}
	return ignored
}
