package vim

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
// 使われていない引数の頭にはアンダースコアを付けて明示しています
func Run() {
	// 1. 最新リリースのJSONを取得
	req, err := http.NewRequest("GET", "https://api.github.com/repos/vim/vim-win32-installer/releases/latest", nil)
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

	// 2. 正規表現でインストーラー（.exe）のダウンロードURLを抽出
	reUrl := regexp.MustCompile(`"browser_download_url":\s*"(https://[^"]+_x64_signed\.exe)"`)
	match := reUrl.FindStringSubmatch(contents)

	var downloadUrl string
	if len(match) > 1 {
		downloadUrl = match[1]
	}

	if downloadUrl == "" {
		fmt.Fprintln(os.Stderr, "failed to find ZIP URL for gvim-x64-signed")
		os.Exit(1)
	}
	fmt.Println("Download URL:", downloadUrl)

	// 3. ユーザーの「ダウンロード」フォルダパスを構築 ($HOME/Downloads)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Impossible to get your home dir!")
		os.Exit(1)
	}
	userDownloadDir := filepath.Join(homeDir, "Downloads")

	// 4. URLの末尾からファイル名（gvim_xxxx_x64_signed.exe）を抽出して保存先パスを作る
	fileName := filepath.Base(downloadUrl)
	localExePath := filepath.Join(userDownloadDir, fileName)

	// 5. インストーラーをダウンロード
	exeReq, err := http.NewRequest("GET", downloadUrl, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	exeReq.Header.Set("User-Agent", "Mozilla/5.0")

	exeResp, err := client.Do(exeReq)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer exeResp.Body.Close()

	if exeResp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "download failed with status: %s\n", exeResp.Status)
		os.Exit(1)
	}

	out, err := os.Create(localExePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if _, err = io.Copy(out, exeResp.Body); err != nil {
		out.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	out.Close()
	fmt.Println("Download (ZIP) is done")

	// 6. 外部コマンド cmd /C start explorer . の実行
	// ダウンロードしたディレクトリを基点にしてエクスプローラーを開きます
	explorerCmd := exec.Command("cmd", "/C", "start", "explorer", ".")
	explorerCmd.Dir = userDownloadDir

	if err := explorerCmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Opened EXPLORER.EXE")
}
