package i18n

// messagesJA contains Japanese translations.
var messagesJA = map[string]string{
	// Root command
	"root.short": "軽量・高速なタスク管理CLIおよびローカルWebアプリ",
	"root.long":  "mini-kanbanは軽量で高速なタスク管理ツールです。CLIとローカルWebインターフェースを提供します。",
	"root.flag.project": "プロジェクト名（環境変数/設定を上書き）",
	"root.flag.lang":    "ヘルプの言語 (en, ja)",

	// init command
	"init.use":   "init",
	"init.short": "設定とデータベースを初期化",
	"init.long":  "mini-kanbanの設定ファイルとデータベースを初期化して利用を開始します。",

	// add command
	"add.use":       "add <タイトル>",
	"add.short":     "新しいタスクを追加",
	"add.long":      "現在のプロジェクトに新しいタスクを追加します。",
	"add.flag.body": "タスクの本文/説明",
	"add.flag.tag":  "タグ（複数指定可）",
	"add.flag.due":  "期限 (YYYY-MM-DD または YYYY-MM-DD HH:MM)",
	"add.flag.ai":   "AI支援を有効にする",
	"add.flag.noai": "AI支援を無効にする",

	// ls command
	"ls.use":        "ls",
	"ls.short":      "タスク一覧を表示",
	"ls.long":       "現在のプロジェクトのタスクをフィルタ付きで一覧表示します。",
	"ls.flag.open":  "未完了のタスクのみ表示",
	"ls.flag.done":  "完了済みのタスクのみ表示",
	"ls.flag.tag":   "タグでフィルタ",
	"ls.flag.q":     "検索クエリ",
	"ls.flag.sort":  "ソート順: created, updated, due",
	"ls.flag.limit": "最大件数",
	"ls.flag.json":  "NDJSON形式で出力",

	// show command
	"show.use":   "show <タスクID>",
	"show.short": "タスクの詳細を表示",
	"show.long":  "指定したタスクの詳細情報を表示します。",

	// edit command
	"edit.use":           "edit <タスクID>",
	"edit.short":         "タスクを編集",
	"edit.long":          "既存のタスクのプロパティを編集します。",
	"edit.flag.title":    "新しいタイトル",
	"edit.flag.body":     "新しい本文/説明",
	"edit.flag.status":   "新しいステータス (open, done)",
	"edit.flag.due":      "新しい期限 (YYYY-MM-DD または YYYY-MM-DD HH:MM)",
	"edit.flag.cleardue": "期限をクリア",
	"edit.flag.addtag":   "タグを追加",
	"edit.flag.rmtag":    "タグを削除",

	// done command
	"done.use":   "done <タスクID>",
	"done.short": "タスクを完了にする",
	"done.long":  "タスクを完了済みとしてマークします。",

	// undo command
	"undo.use":   "undo <タスクID>",
	"undo.short": "完了したタスクを再開",
	"undo.long":  "完了済みのタスクを未完了に戻します。",

	// rm command
	"rm.use":   "rm <タスクID>",
	"rm.short": "タスクを削除",
	"rm.long":  "タスクを完全に削除します。",

	// project command
	"project.use":        "project",
	"project.short":      "プロジェクトを管理",
	"project.long":       "タスクを整理するためのプロジェクトを管理します。",
	"project.list.use":   "list",
	"project.list.short": "全プロジェクトを一覧表示",

	// config command
	"config.use":        "config",
	"config.short":      "設定を管理",
	"config.long":       "mini-kanbanの設定を表示・管理します。",
	"config.show.use":   "show",
	"config.show.short": "現在の設定を表示",
	"config.init.use":   "init",
	"config.init.short": "設定ファイルを初期化",

	// db command
	"db.use":          "db",
	"db.short":        "データベース操作",
	"db.long":         "データベースのメンテナンス操作を行います。",
	"db.path.use":     "path",
	"db.path.short":   "データベースのパスを表示",

	// web command
	"web.use":       "web",
	"web.short":     "ローカルWeb UIを起動",
	"web.long":      "mini-kanban WebインターフェースのローカルWebサーバーを起動します。",
	"web.flag.port": "待ち受けポート",
	"web.flag.host": "バインドするホスト",

	// Web Interface
	"web.project":      "プロジェクト:",
	"web.search.ph":    "タスクを検索...",
	"web.show_done":    "完了済みを表示",
	"web.add.title":    "新しいタスク...",
	"web.add.tags":     "タグ（カンマ区切り）",
	"web.add.btn":      "追加",
	"web.list.empty":   "タスクが見つかりません。",
	"web.status.open":  "未完了",
	"web.status.done":  "完了",
	"web.overdue":      "期限切れ",
	"web.created":      "作成日",
	"web.footer":       "mini-kanban • 軽量タスク管理",
	"web.page.title":   "mini-kanban",

	// Kanban columns
	"web.kanban.todo":  "Todo",
	"web.kanban.doing": "作業中",
	"web.kanban.done":  "完了",

	// Due date presets
	"web.due.today":       "今日まで",
	"web.due.tomorrow":    "明日中",
	"web.due.this_week":   "今週中",
	"web.due.this_month":  "今月まで",
	"web.due.next_month":  "来月まで",
	"web.due.none":        "期限なし",
}
