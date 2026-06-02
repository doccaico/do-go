package delete_duplicate_path

import (
	"fmt"
	"os"
	"strings"
)

// Run はメインから呼び出されるエントリーポイントです
func Run() {
	// 1. 環境変数 PATH の取得
	paths := os.Getenv("PATH")
	if paths == "" {
		panic("couldn't interpret 'PATH'")
	}

	// 2. セミコロン（;）で分割
	pathList := strings.Split(paths, ";")

	// 重複チェック用のマップ
	seen := make(map[string]struct{})
	// 結果を格納するスライス
	var pathVec []string

	// 3. ループで重複を排除しながら追加
	for _, path := range pathList {
		// 空文字は除外
		if path == "" {
			continue
		}

		// マップに存在しない（まだ追加していない）場合
		if _, exists := seen[path]; !exists {
			seen[path] = struct{}{}         // マップに記録
			pathVec = append(pathVec, path) // スライスに追加
		}
	}

	// 4. 再びセミコロン（;）で結合して出力
	fmt.Print(strings.Join(pathVec, ";"))
}
