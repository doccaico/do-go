package diarysearch

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

const helpMsg = `
Usage:
    do.exe diary_search [OPTION] 検索キーワード
OPTION:
    -h, --help                 ヘルプメッセージを表示
REQUIRED:
    環境変数(DIARY_DIR)に日記が入っているディレクトリを設定すること`

// Run はメインから呼び出されるエントリーポイントです
func Run(args []string) {
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Println(helpMsg)
		os.Exit(0)
	}
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, helpMsg)
		os.Exit(1)
	}

	// 1. 環境変数の取得
	diaryDirRaw := os.Getenv("DIARY_DIR")
	if diaryDirRaw == "" {
		panic("not found 'DIARY_DIR' in env variable")
	}

	// 2. ディレクトリの存在確認
	info, err := os.Stat(diaryDirRaw)
	if os.IsNotExist(err) || !info.IsDir() {
		panic(fmt.Sprintf("'%s' does not exist or is a file", diaryDirRaw))
	}

	// 3. rip-grep (rg) コマンドの実行
	rgCmd := exec.Command("rg",
		"--color", "always",
		"--heading",
		"--line-number",
		"--ignore-case",
		"--sort=path",
		args[0],
		diaryDirRaw,
	)

	// rgの標準出力をまとめて取得する
	output, err := rgCmd.Output()

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			code := exitErr.ExitCode()

			if code == 1 {
				fmt.Printf("No matches found for '%s'.\n", args[0])
				os.Exit(0)
			} else if code >= 2 {
				panic(fmt.Sprintf("'rg' failed with exit code %d", code))
			}
		}
		// コマンドが見つからない、起動できないなどの致命的なエラーはここでキャッチする
		panic(err)
	}

	// 4. less コマンドの実行準備
	lessCmd := exec.Command("less", "-R", "-i", "--silent")

	// less の標準入力をパイプで取得
	lessStdin, err := lessCmd.StdinPipe()
	if err != nil {
		panic(fmt.Sprintf("failed to open stdin: %v", err))
	}

	// less の出力を現在のターミナル（標準出力・標準エラー）に紐付ける
	lessCmd.Stdout = os.Stdout
	lessCmd.Stderr = os.Stderr

	// less コマンドを開始（非同期）
	if err := lessCmd.Start(); err != nil {
		panic(fmt.Sprintf("failed to spawn 'less': %v", err))
	}

	// less の標準入力に rg の結果を書き込む
	_, writeErr := lessStdin.Write(output)

	// 書き込みが終わったら明示的に閉じて less に通知（Rustの drop(stdin) 相当）
	lessStdin.Close()

	if writeErr != nil {
		fmt.Fprintln(os.Stderr, "failed to write to less stdin")
		_ = lessCmd.Process.Kill()
		_ = lessCmd.Wait()
		os.Exit(1)
	}

	// less の終了を待つ
	if err := lessCmd.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, "less exited with an error")
		os.Exit(1)
	}
}
