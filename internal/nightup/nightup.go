package nightup

import (
	"fmt"
	"gopkg.in/ini.v1"
	"os"
	"path/filepath"

	golang "github.com/doccaico/do-go/internal/nightup/golang"
	odinlang "github.com/doccaico/do-go/internal/nightup/odinlang"
	vim "github.com/doccaico/do-go/internal/nightup/vim"
	vlang "github.com/doccaico/do-go/internal/nightup/vlang"
	ziglang "github.com/doccaico/do-go/internal/nightup/ziglang"
)

const helpMsg = `
Usage:
    do.exe nightup go
    do.exe nightup odin
    do.exe nightup v
    do.exe nightup zig
    do.exe nightup vim`

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

	// 1. ホームディレクトリの取得
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Impossible to get your home dir!")
		os.Exit(1)
	}

	// パスの結合 ($HOME/.nightup)
	iniPath := filepath.Join(homeDir, ".nightup")

	// 2. INIファイルのロード (Rustの load_from_file_noescape 相当)
	// go-ini ライブラリはデフォルトでバックスラッシュ（\）をエスケープシーケンスとして処理しないためそのままでOK
	cfg, err := ini.Load(iniPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open: %s\n", iniPath)
		os.Exit(1)
	}

	// 3. Windows セクションの取得
	section := cfg.Section("Windows")
	if section == nil {
		fmt.Fprintln(os.Stderr, `nightup ini: not found section: "Windows"`)
		os.Exit(1)
	}

	// 4. 一時保存ディレクトリの設定 (TEMP環境変数。無ければ現在のディレクトリ ".")
	downloadDir := os.Getenv("TEMP")
	if downloadDir == "" {
		downloadDir = "."
	}

	// 5. 各言語のアップデート処理への振り分け
	switch args[0] {
	case "zig":
		if !section.HasKey("zig") {
			fmt.Fprintln(os.Stderr, `nightup ini: not found path: "zig"`)
			os.Exit(1)
		}
		distDir := section.Key("zig").String()
		ziglang.Run(distDir, downloadDir)

	case "odin":
		if !section.HasKey("odin") {
			fmt.Fprintln(os.Stderr, `nightup ini: not found path: "odin"`)
			os.Exit(1)
		}
		distDir := section.Key("odin").String()
		odinlang.Run(distDir, downloadDir)

	case "v":
		if !section.HasKey("v") {
			fmt.Fprintln(os.Stderr, `nightup ini: not found path: "v"`)
			os.Exit(1)
		}
		distDir := section.Key("v").String()
		vlang.Run(distDir, downloadDir)

	case "go":
		if !section.HasKey("go") {
			fmt.Fprintln(os.Stderr, `nightup ini: not found path: "go"`)
			os.Exit(1)
		}
		distDir := section.Key("go").String()
		golang.Run(distDir, downloadDir) // パッケージ名は go_up

	case "vim":
		vim.Run()

	default:
		fmt.Fprintf(os.Stderr, "nightup: unknown command '%s'\n%s\n", args[0], helpMsg)
		os.Exit(1)
	}
}
