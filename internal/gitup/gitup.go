package gitup

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

const helpMsg = `
Usage:
    do.exe gitup [OPTION] DIR MESSAGE
    do.exe gitup [OPTION]     MESSAGE
OPTION:
    -h, --help                 ヘルプメッセージを表示`

// Run はメインから呼び出されるエントリーポイントです
func Run(args []string) {
	if len(args) == 0 || len(args) > 2 {
		fmt.Fprintln(os.Stderr, helpMsg)
		os.Exit(1)
	}
	if args[0] == "-h" || args[0] == "--help" {
		fmt.Println(helpMsg)
		os.Exit(0)
	}

	var dirPath string
	var commitMsg string

	if len(args) == 2 {
		info, err := os.Stat(args[0])
		if os.IsNotExist(err) || !info.IsDir() {
			fmt.Fprintf(os.Stderr, "'%s' does not exist or is a file\n", args[0])
			os.Exit(1)
		}
		dirPath = args[0]
		commitMsg = args[1]
	} else {
		dirPath = "."
		commitMsg = args[0]
	}

	// 1. git status --porcelain の実行
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = dirPath // 実行ディレクトリを指定

	output, err := statusCmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "'%s' is not a git repository (or git error)\n", dirPath)
		os.Exit(1)
	}

	// 空白や改行を削って変更があるかチェック
	if len(bytes.TrimSpace(output)) == 0 {
		fmt.Println("There is no need to update")
		return
	}

	// 2. git add . の実行
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = dirPath
	addCmd.Stdout = os.Stdout // 標準出力を現在のターミナルに紐付け (標準出力を得るため)
	addCmd.Stderr = os.Stderr // 標準エラー出力を現在のターミナルに紐付け (標準出力を得るため)
	if err := addCmd.Run(); err != nil {
		fmt.Println("failed to run 'git add'")
		os.Exit(1)
	}

	// 3. git commit -m "MESSAGE" の実行
	commitCmd := exec.Command("git", "commit", "-m", commitMsg)
	commitCmd.Dir = dirPath
	commitCmd.Stdout = os.Stdout
	commitCmd.Stderr = os.Stderr
	if err := commitCmd.Run(); err != nil {
		fmt.Println("failed to run 'git commit'")
		os.Exit(1)
	}

	// 4. git push の実行
	pushCmd := exec.Command("git", "push")
	pushCmd.Dir = dirPath
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr
	if err := pushCmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to run 'git push'")
		os.Exit(1)
	}
}
