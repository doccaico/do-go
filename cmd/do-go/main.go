package main

import (
	"fmt"
	delete_duplicate_path "github.com/doccaico/do-go/internal/delete-duplicate-path"
	diary_search "github.com/doccaico/do-go/internal/diary-search"
	gitup "github.com/doccaico/do-go/internal/gitup"
	nightup "github.com/doccaico/do-go/internal/nightup"
	shitaraba "github.com/doccaico/do-go/internal/shitaraba"
	verse "github.com/doccaico/do-go/internal/verse"
	wiki "github.com/doccaico/do-go/internal/wiki"
	"os"
)

const helpMsg = `
Usage:
    do.exe [OPTION] COMMAND [ARGS...]
OPTION:
    -h, --help                 ヘルプメッセージを表示
COMMAND:
    diary_search                環境変数(DIARY_DIR)にある日記を検索
    gitup                       GithubにPush
    shitaraba                   Shitarabaを閲覧
    delete_duplicate_path       環境変数PATHの重複を解消して表示
    verse                       聖書(新共同訳)を表示
    wiki                        ランダムWIKIのリストを表示
    nightup                     ソフトウェアアップデーター`

func main() {
	args := os.Args

	if len(args) == 1 {
		fmt.Fprintln(os.Stderr, helpMsg)
		os.Exit(1)
	}

	if len(args) == 2 && (args[1] == "-h" || args[1] == "--help") {
		fmt.Println(helpMsg)
		os.Exit(0)
	}

	// サブコマンドの判定
	switch args[1] {
	case "diary_search":
		diary_search.Run(args[2:])
	case "gitup":
		gitup.Run(args[2:])
	case "delete_duplicate_path":
		delete_duplicate_path.Run()
	case "shitaraba":
		shitaraba.Run(args[2:])
	case "verse":
		verse.Run(args[2:])
	case "wiki":
		wiki.Run(args[2:])
	case "nightup":
		nightup.Run(args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command '%s'\n", args[1])
		fmt.Fprintln(os.Stderr, helpMsg)
		os.Exit(1)
	}
}
