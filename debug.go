package debugmonitor

// debug は開発時にldflagsで設定するデバッグフラグ
// ビルド時に -ldflags="-X github.com/kohkimakimoto/echo-debugmonitor.debug=true" で設定可能
// 小文字にすることで外部パッケージから直接アクセスできないようにしている
var debug string

// isDebug はデバッグモードが有効かどうかを返す
// 小文字にすることで外部パッケージから直接アクセスできないようにしている
func isDebug() bool {
	return debug == "true"
}
