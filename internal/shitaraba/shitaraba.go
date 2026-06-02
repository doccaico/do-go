package shitaraba

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

const helpMsg = `
Usage:
    do.exe shitaraba [OPTION] GENRE ID NUMBER
OPTION:
    -h, --help                 ヘルプメッセージを表示`

// HTMLエンティティ（&#12345;）を絵文字/文字に変換する関数
func convertCp(codePoint string) string {
	cpInt, err := strconv.ParseUint(codePoint, 10, 32)
	if err != nil {
		return codePoint
	}
	return string(rune(cpInt))
}

// Run はメインから呼び出されるエントリーポイントです
func Run(args []string) {
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Println(helpMsg)
		os.Exit(0)
	}
	if len(args) != 3 {
		fmt.Fprintln(os.Stderr, helpMsg)
		os.Exit(1)
	}

	genre, id, number := args[0], args[1], args[2]
	url := fmt.Sprintf("https://jbbs.shitaraba.net/bbs/read.cgi/%s/%s/%s/l50", genre, id, number)

	// 1. HTTPリクエストの送信
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// 2. EUC-JP から UTF-8 へのデコード（busybox64u iconv の代わり）
	utf8Reader := transform.NewReader(resp.Body, japanese.EUCJP.NewDecoder())
	bodyBytes, err := io.ReadAll(utf8Reader)
	if err != nil {
		panic(err)
	}
	contents := string(bodyBytes)

	// 3. 正規表現のコンパイル
	re := regexp.MustCompile(`(?s)<dt\b[^>]*?>.+?<b>(.*?)</b>.+?：\s*(.*?)</dt>\s*<dd>(.*?)</dd>`)
	reEmoji := regexp.MustCompile(`&#(\d+?);`)
	reTag := regexp.MustCompile(`<[^>]*?>`)

	type datum struct {
		name, date, post string
	}
	var dataList []datum

	// 4. マッチングとデータ構築
	matches := re.FindAllStringSubmatch(contents, -1)
	for _, match := range matches {
		name := match[1]
		date := match[2]
		post := match[3]

		// 名前のタグ削除とトリム
		name = reTag.ReplaceAllString(name, "")
		name = strings.TrimSpace(name)

		// 日付のトリム（右側の空白を削除）
		date = strings.TrimRight(date, " \t\r\n")

		// 本文の改行変換とタグ削除
		post = strings.TrimSpace(post)
		post = strings.ReplaceAll(post, "<br>          <br>", "\n")
		post = strings.ReplaceAll(post, "<br>", "")
		post = reTag.ReplaceAllString(post, "")

		// HTMLエンティティ（絵文字）の置換
		finalPost := reEmoji.ReplaceAllStringFunc(post, func(m string) string {
			// m は "&#12345;" 全体なので、数字部分だけを抜き出す
			subMatch := reEmoji.FindStringSubmatch(m)
			if len(subMatch) > 1 {
				emoji := convertCp(subMatch[1])
				return emoji
			}
			return m
		})

		dataList = append(dataList, datum{name: name, date: date, post: finalPost})
	}

	// 5. less コマンドの実行準備
	lessCmd := exec.Command("less", "-R", "-i", "--silent")

	lessStdin, err := lessCmd.StdinPipe()
	if err != nil {
		panic(fmt.Sprintf("failed to open stdin: %v", err))
	}

	lessCmd.Stdout = os.Stdout
	lessCmd.Stderr = os.Stderr

	if err := lessCmd.Start(); err != nil {
		panic(fmt.Sprintf("failed to spawn 'less': %v", err))
	}

	// lessの標準入力にデータを書き込む
	for _, d := range dataList {
		_, writeErr := fmt.Fprintf(lessStdin, "\x1b[36m%s\x1b[0m: \x1b[32m%s\x1b[0m\n%s\n", d.name, d.date, d.post)
		if writeErr != nil {
			fmt.Fprintln(os.Stderr, "failed to write to less stdin")
			_ = lessCmd.Process.Kill()
			_ = lessCmd.Wait()
			os.Exit(1)
		}
	}

	// 書き込み終了を less に通知
	lessStdin.Close()

	// less の終了を待つ
	if err := lessCmd.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, "less exited with an error")
		os.Exit(1)
	}
}
