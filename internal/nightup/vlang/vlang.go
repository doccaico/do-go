package vlang

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

// Run はメイン（nightup）から呼び出されるエントリーポイントです
func Run(distDir, downloadDir string) {
	// 1. 最新バージョンのJSONを取得
	req, err := http.NewRequest("GET", "https://api.github.com/repos/vlang/v/releases/latest", nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	contents := string(bodyBytes)
	fmt.Println("Download (json) is done")

	// 2. 正規表現で Windows 用の ZIP ファイルのブラウザダウンロード URL を抽出
	reUrl := regexp.MustCompile(`"browser_download_url":\s*"(https://[^"]+v_windows\.zip)"`)
	match := reUrl.FindStringSubmatch(contents)

	var downloadUrl string
	if len(match) > 1 {
		downloadUrl = match[1]
	}

	if downloadUrl == "" {
		fmt.Fprintln(os.Stderr, "failed to find ZIP URL for v-windows")
		os.Exit(1)
	}
	fmt.Println("Download URL:", downloadUrl)

	// 3. 作業用ディレクトリの作成
	workDirName := "v-latest-upgrade-working"
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

	// 正常移動が完了しなかった時だけフォルダを消すクリーンアップ処理
	cleanupActive := true
	defer func() {
		if cleanupActive {
			_ = os.RemoveAll(workDirPath)
		}
	}()

	// 4. ZIPファイルのダウンロード
	zipReq, err := http.NewRequest("GET", downloadUrl, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	zipReq.Header.Set("User-Agent", "Mozilla/5.0")

	zipResp, err := client.Do(zipReq)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer zipResp.Body.Close()

	if zipResp.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	localZip := "v-latest.zip"
	localZipPath := filepath.Join(workDirPath, localZip)
	out, err := os.Create(localZipPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if _, err = io.Copy(out, zipResp.Body); err != nil {
		out.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	out.Close()
	fmt.Println("Download (ZIP) is done")

	// 5. 外部コマンド tar の実行
	tarCmd := exec.Command("tar", "-xf", localZip, "--strip-components=1")
	tarCmd.Dir = workDirPath
	tarCmd.Stdout = os.Stdout
	tarCmd.Stderr = os.Stderr

	if err := tarCmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Extraction is done")

	// 6. 不要になったZIPの削除
	if _, err := os.Stat(localZipPath); err == nil {
		if err := os.Remove(localZipPath); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("Removed: \"%s\"\n", localZipPath)
	}

	// 7. 配置（アップデートの適用）
	if _, err := os.Stat(distDir); err == nil {
		if err := os.RemoveAll(distDir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("Removed: \"%s\"\n", distDir)
	}

	// ワークスペースを作業パスから distDir へ移動
	if err := os.Rename(workDirPath, distDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// 移動成功のためクリーンアップフラグを解除
	cleanupActive = false

	fmt.Printf("Moved: \"%s\" to \"%s\"\n", workDirPath, distDir)
	fmt.Printf("Updated: \"%s\"\n", distDir)
}
