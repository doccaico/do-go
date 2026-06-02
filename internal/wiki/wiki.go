package wiki

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

const helpMsg = `
USAGE:
    do.exe wiki [OPTION] COUNT
OPTION:
    -h, --help                 ヘルプメッセージを表示`

// \uXXXX 形式の Unicode エスケープ文字を元の文字にデコードする関数
func decodeUnicodeEscape(s string) string {
	re := regexp.MustCompile(`\\u([0-9a-fA-F]{4})`)
	return re.ReplaceAllStringFunc(s, func(m string) string {
		// m は "\\u1234" なので、ダブルクォーテーションで囲って strconv.Unquote に渡す
		unquoted, err := strconv.Unquote(`"` + m + `"`)
		if err != nil {
			return m
		}
		return unquoted
	})
}

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

	url := fmt.Sprintf(
		"https://ja.wikipedia.org/w/api.php"+
			"?format=json"+
			"&action=query"+
			"&list=random"+
			"&rnnamespace=0"+
			"&rnfilterredir=nonredirects"+
			"&rnlimit=%s",
		args[0],
	)

	// 1. Curl
	cmd := exec.Command("curl", "-sSL", "-A", "Mozilla/5.0", url)

	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	responseText := string(output)

	// 2. 正規表現による ID と Title の抽出
	// Goのregexpでドット（.）を改行にマッチさせるため、(?s) を使用します
	re := regexp.MustCompile(`(?s)"id":\s*(\d+).*?"title":\s*"([^"]+?)"`)

	var articles []string
	matches := re.FindAllStringSubmatch(responseText, -1)
	for idx, match := range matches {
		id := match[1]
		title := match[2]
		cleanTitle := decodeUnicodeEscape(title)

		// 色付きの出力文字列を生成
		article := fmt.Sprintf(
			"\x1b[35m%d\x1b[0m:\x1b[36m%s\x1b[0m: \x1b[32mhttps://ja.wikipedia.org/?curid=%s\x1b[0m",
			idx+1, cleanTitle, id,
		)
		articles = append(articles, article)
	}

	if len(articles) == 0 {
		fmt.Fprintln(os.Stderr, "No random articles found. Raw response might be unexpected.")
		fmt.Fprintf(os.Stderr, "Raw: %s\n", responseText)
		os.Exit(1)
	}

	// 3. less コマンドの実行準備
	lessCmd := exec.Command("less", "-R", "-i", "--silent")

	lessStdin, err := lessCmd.StdinPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	lessCmd.Stdout = os.Stdout
	lessCmd.Stderr = os.Stderr

	if err := lessCmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// 4. less の標準入力にデータを書き込む
	for _, article := range articles {
		_, writeErr := fmt.Fprintln(lessStdin, article)
		if writeErr != nil {
			fmt.Fprintln(os.Stderr, err)
			_ = lessCmd.Process.Kill()
			_ = lessCmd.Wait()
			os.Exit(1)
		}
	}

	// 書き込み終了を less に通知
	lessStdin.Close()

	// less の終了を待つ
	if err := lessCmd.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
