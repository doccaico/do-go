package golang

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

// Run はメイン（nightup）から呼び出されるエントリーポイントです
func Run(distDir, downloadDir string) {
	// 1. Curl
	url := "https://go.dev/dl/?mode=json"
	cmd := exec.Command("curl", "-sSL", "-A", "Mozilla/5.0", url)

	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	contents := string(output)
	fmt.Println("Download (json) is done")

	// 2. 正規表現で Windows 用の ZIP ファイル名を抽出
	reUrl := regexp.MustCompile(`"filename":\s*"(go[0-9.]+\.windows-amd64\.zip)"`)
	match := reUrl.FindStringSubmatch(contents)

	var filename string
	if len(match) > 1 {
		filename = match[1]
	}

	if filename == "" {
		fmt.Fprintln(os.Stderr, "failed to find ZIP URL for go-windows-amd64")
		os.Exit(1)
	}

	downloadUrl := "https://go.dev/dl/" + filename
	fmt.Println("Download URL:", downloadUrl)

	// 3. 作業用ディレクトリの作成
	workDirName := "go-latest-upgrade-working"
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
	localZip := "go-latest.zip"
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
	fmt.Println("Download (ZIP) is done")

	// 5. 外部コマンド tar の実行（ご指定の方法）
	// Windows10/11標準の bsdtar は tar -xf で .zip も透過的に解凍できます
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
