# 開発者ガイド

Mini-Kanban の開発に関する情報をまとめています。

## ビルド

```bash
make build
```

## Go で直接実行（ビルド不要）

```bash
# CLI として使用
go run ./cmd [コマンド]

# 例: タスク一覧
go run ./cmd ls

# 例: Web UI を起動
go run ./cmd web
```

## 開発時のビルド＆起動

### フロントエンド開発（推奨）

フロントエンドの変更を即座に反映させるには、以下の手順で開発します。

**ターミナル1（フロントエンドの監視ビルド）**:

```bash
cd web
npm run build -- --watch
```

**ターミナル2（Goサーバーを開発モードで起動）**:

```bash
# --dev フラグをつけると web/dist から直接配信されます（Goの再ビルド不要）
go run ./cmd web --dev
```

## ディレクトリ構成

```
mini-kanban/
├── cmd/mini-kanban/  # メインエントリポイント
├── cli/              # CLI コマンド実装
├── srv/              # Web サーバー・ハンドラー
│   └── templates/    # HTML テンプレート
├── db/               # データベース・マイグレーション
├── ai/               # AI 支援機能
├── config/           # 設定管理
└── search/           # 検索機能
```

## リリース手順

本プロジェクトは GitHub Actions を使用して npm へ自動リリースします。

### 自動リリース（タグプッシュ）

バージョンタグをプッシュすると自動的に npm パッケージが公開されます。

```bash
# バージョンタグを作成してプッシュ
git tag v1.0.0
git push origin v1.0.0
```

### 手動リリース

GitHub Actions の「Release to npm」ワークフローから手動でトリガーすることも可能です。

1. GitHub リポジトリの **Actions** タブを開く
2. 左メニューから **Release to npm** を選択
3. **Run workflow** ボタンをクリック
4. バージョン番号を入力（例: `1.0.0`）して実行

### 必要なシークレット設定

リリースを行うには、リポジトリに以下のシークレットを設定してください。

| シークレット名 | 説明 |
|---------------|------|
| `NPM_TOKEN` | npm の公開用アクセストークン |

> [!NOTE]
> npm トークンは [npm アカウント設定](https://www.npmjs.com/settings/~/tokens) から生成できます。

### 公開されるパッケージ

リリースにより以下のパッケージが npm に公開されます。

| パッケージ名 | 説明 |
|-------------|------|
| `mini-kanban` | メインパッケージ（`npx mini-kanban` で実行可能） |
| `@mini-kanban/linux-x64` | Linux x64 用バイナリ |
| `@mini-kanban/linux-arm64` | Linux ARM64 用バイナリ |
| `@mini-kanban/darwin-x64` | macOS x64 用バイナリ |
| `@mini-kanban/darwin-arm64` | macOS ARM64 (Apple Silicon) 用バイナリ |
| `@mini-kanban/win32-x64` | Windows x64 用バイナリ |
