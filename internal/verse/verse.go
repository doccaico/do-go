package verse

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const helpMsg = `
USAGE:
    do.exe verse [OPTION] 書物 章
OPTION:
    -h, --help                 ヘルプメッセージを表示

      ` + "\x1b[36m" + `- 旧約 (Old Testament) -` + "\x1b[0m" + `
      創世記:GEN(1:50)
      出エジプト記:EXO(1:40)
      レビ記:LEV(1:27)
      民数記:NUM(1:36)
      申命記:DEU(1:34)
      ヨシュア記:JOS(1:24)
      士師記:JDG(1:21)
      ルツ記:RUT(1:4)
      サムエル記(上):1SA(1:31)
      サムエル記(下):2SA(1:24)
      列王記(上):1KI(1:22)
      列王記(下):2KI(1:25)
      歴代誌(上):1CH(1:29)
      歴代誌(下):2CH(1:36)
      エズラ記:EZR(1:10)
      ネヘミヤ記:NEH(1:13)
      エステル記:EST(1:10)
      ヨブ記:JOB(1:42)
      詩編:PSA(1:150)
      箴言:PRO(1:31)
      コヘレトの言葉:ECC(1:12)
      雅歌:SNG(1:8)
      イザヤ書:ISA(1:66)
      エレミヤ書:JER(1:52)
      哀歌:LAM(1:5)
      エゼキエル書:EZK(1:48)
      ダニエル書:DAN(1:12)
      ホセア書:HOS(1:14)
      ヨエル書:JOL(1:4)
      アモス書:AMO(1:9)
      オバデヤ書:OBA(1)
      ヨナ書:JON(1:4)
      ミカ書:MIC(1:7)
      ナホム書:NAM(1:3)
      ハバクク書:HAB(1:3)
      ゼファニヤ書:ZEP(1:3)
      ハガイ書:HAG(1:2)
      ゼカリヤ書:ZEC(1:14)
      マラキ書:MAL(1:3)
      ユディト記:JDT(1:16)
      知恵の書:WIS(1:19)
      トビト記:TOB(1:14)
      シラ:SIR(1:51)
      バルク書:BAR(1:5)
      エレミヤの手紙:LJE(1)
      マカバイ記(一):1MA(1:16)
      マカバイ記(二 書簡):2MA(1:15)
      エステル記(ギリシア語):ESG(1:10 + 10_1)
      ダニエル書補遺 スザンナ:SUS(1)
      ダニエル書補遺 ベルと竜:BEL(1)
      ダニエル書補遺 アザルヤの祈りと三人の若者の賛歌:S3Y(1)
      エズラ記(ギリシア語):1ES(1:9)
      エズラ記(ラテン語):2ES(1:16)
      マナセの祈り:MAN(1)
      ` + "\x1b[36m" + `- 新約 (New Testament) -` + "\x1b[0m" + `
      マタイによる福音書:MAT(1:28)
      マルコによる福音書:MRK(1:16)
      ルカによる福音書:LUK(1:24)
      ヨハネによる福音書:JHN(1:21)
      使徒言行録:ACT(1:28)
      ローマの信徒への手紙:ROM(1:16)
      コリントの信徒への手紙(一):1CO(1:16)
      コリントの信徒への手紙(二):2CO(1:13)
      ガラテヤの信徒への手紙:GAL(1:6)
      エフェソの信徒への手紙:EPH(1:6)
      フィリピの信徒への手紙:PHP(1:4)
      コロサイの信徒への手紙:COL(1:4)
      テサロニケの信徒への手紙(一):1TH(1:5)
      テサロニケの信徒への手紙(二):2TH(1:3)
      テモテへの手紙(一):1TI(1:6)
      テモテへの手紙(二):2TI(1:4)
      テトスへの手紙:TIT(1:3)
      フィレモンへの手紙:PHM(1)
      ペトロの手紙(一):1PE(1:5)
      ペトロの手紙(二):2PE(1:3)
      ヨハネの手紙(一):1JN(1:5)
      ヨハネの手紙(二):2JN(1)
      ヨハネの手紙(三):3JN(1)
      ヘブライ人への手紙:HEB(1:13)
      ヤコブの手紙:JAS(1:5)
      ユダの手紙:JUD(1)
      ヨハネの黙示録:REV(1:22)`

// Run はメインから呼び出されるエントリーポイントです
func Run(args []string) {
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Println(helpMsg)
		os.Exit(0)
	}
	if len(args) != 2 {
		fmt.Println(helpMsg)
		os.Exit(1)
	}

	book, chapter := args[0], args[1]
	url := fmt.Sprintf("https://bible.com/ja/bible/1819/%s.%s/", book, chapter)

	// 1. HTTPリクエストの送信
	req, err := http.NewRequest("GET", url, nil)
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

	// 2. 正規表現のコンパイルとマッチング
	// Goのregexpでドット（.）を改行にマッチさせるため、(?s) を使用します
	re := regexp.MustCompile(`(?s)content">(.+?)</span>`)

	var datum []string
	matches := re.FindAllStringSubmatch(contents, -1)
	for _, match := range matches {
		// match[1] がキャプチャグループの「(.+?)」に相当します
		body := strings.TrimLeft(match[1], " \t\r\n")
		if body != "" {
			datum = append(datum, body)
		}
	}

	if len(datum) == 0 {
		fmt.Fprintln(os.Stderr, "no verses found or failed to parse the page")
		fmt.Fprintf(os.Stderr, "check the page: %s\n", url)
		os.Exit(1)
	}

	// 3. less コマンドの実行準備
	lessCmd := exec.Command("less", "-R", "--silent")

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
	// タイトルの書き込み (Rustの writeln! 相当)
	_, _ = fmt.Fprintf(lessStdin, "\x1b[36m%s\x1b[0m:\x1b[32m%s\x1b[0m\n", book, chapter)

	for _, data := range datum {
		_, writeErr := fmt.Fprintln(lessStdin, data)
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
