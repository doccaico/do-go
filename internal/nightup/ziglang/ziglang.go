package ziglang

import (
	"fmt"
	// "io"
	// "net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

// Run はメイン（nightup）から呼び出されるエントリーポイントです
func Run(distDir, downloadDir string) {
	// 1. 最新バージョンのJSONを取得
	url := "https://ziglang.org/download/index.json"
	cmd := exec.Command("curl", "-sSL", "-A", "Mozilla/5.0", url)

	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	contents := string(output)
	fmt.Println("Download (index.json) is done")

	// 2. 正規表現で master の x86_64-windows 用の URL を抽出
	// Goのregexpでドット（.）を改行にマッチさせるため、(?s) を使用します
	reUrl := regexp.MustCompile(`(?s)"master":\s*\{.*?"x86_64-windows":\s*\{.*?"tarball":\s*"([^"]+)"`)
	match := reUrl.FindStringSubmatch(contents)

	var downloadUrl string
	if len(match) > 1 {
		downloadUrl = match[1]
	}

	if downloadUrl == "" {
		fmt.Fprintln(os.Stderr, "failed to find ZIP URL for x86_64-windows master")
		os.Exit(1)
	}
	fmt.Println("Download URL:", downloadUrl)

	// 3. 作業用ディレクトリの作成
	workDirName := "zig-master-upgrade-working"
	workDirPath := filepath.Join(downloadDir, workDirName)

	if _, err := os.Stat(workDirPath); err == nil {
		if err := os.RemoveAll(workDirPath); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("Removed: \"%s\"\n", workDirPath)
	}

	if err := os.MkdirAll(workDirPath, 0755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Created: \"%s\"\n", workDirPath)

	// 4. ZIPファイルのダウンロード
	localZip := "zig-master-latest.zip"
	localZipPath := filepath.Join(workDirPath, localZip)

	zipCmd := exec.Command("curl", "-fsSL", "-A", "Mozilla/5.0", downloadUrl, "-o", localZip)
	zipCmd.Dir = workDirPath // 作業ディレクトリに cd してから実行
	zipCmd.Stdout = os.Stdout
	zipCmd.Stderr = os.Stderr

	if err := zipCmd.Run(); err != nil {
		_ = os.RemoveAll(workDirPath)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Download (ZIP) is done: %s\n", localZip)

	// 5. 外部コマンド tar の実行
	tarCmd := exec.Command("tar", "-xf", localZip, "--strip-components=1")
	tarCmd.Dir = workDirPath
	tarCmd.Stdout = os.Stdout
	tarCmd.Stderr = os.Stderr

	if err := tarCmd.Run(); err != nil {
		_ = os.RemoveAll(workDirPath)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Extraction is done")

	// 6. 不要になったZIPの削除
	if _, err := os.Stat(localZipPath); err == nil {
		if err := os.Remove(localZipPath); err != nil {
			_ = os.RemoveAll(workDirPath)
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("Removed: \"%s\"\n", localZipPath)
	}

	// 7. 配置（アップデートの適用）
	if _, err := os.Stat(distDir); err == nil {
		if err := os.RemoveAll(distDir); err != nil {
			_ = os.RemoveAll(workDirPath)
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("Removed: \"%s\"\n", distDir)
	}

	// ワークスペースを作業パスから distDir へ移動
	if err := os.Rename(workDirPath, distDir); err != nil {
		_ = os.RemoveAll(workDirPath)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Moved: \"%s\" to \"%s\"\n", workDirPath, distDir)
	fmt.Printf("Updated: \"%s\"\n", distDir)
}
