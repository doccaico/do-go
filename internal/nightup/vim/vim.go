package vim

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

// Run はメイン（nightup）から呼び出されるエントリーポイントです
func Run() {
	// 1. 最新リリースのJSONを取得
	url := "https://api.github.com/repos/vim/vim-win32-installer/releases/latest"
	cmd := exec.Command("curl", "-sSL", "-A", "Mozilla/5.0", url)

	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	contents := string(output)
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

	// 4. curl を使ってインストーラーをダウンロード (net/http を完全排除)
	// -fsSOL オプションを使用して、リモート名に合わせたファイル名で安全に保存します
	exeCmd := exec.Command("curl", "-fsSOL", "-A", "Mozilla/5.0", downloadUrl)
	exeCmd.Dir = userDownloadDir // ユーザーの Downloads フォルダに cd してから実行
	exeCmd.Stdout = os.Stdout
	exeCmd.Stderr = os.Stderr

	if err := exeCmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Download (ZIP) is done")

	// 5. 外部コマンド cmd /C start explorer . の実行
	// ダウンロードしたディレクトリを基点にしてエクスプローラーを開きます
	explorerCmd := exec.Command("cmd", "/C", "start", "explorer", ".")
	explorerCmd.Dir = userDownloadDir

	if err := explorerCmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Opened EXPLORER.EXE")
}
