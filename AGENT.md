# Mini-Kanban Project Overview for Agents

このドキュメントは、AIエージェントが本プロジェクトを迅速に理解するための普遍的な概要をまとめたものです。

## 1. プロジェクトの目的と概要
**Mini-Kanban** は、開発者向けの軽量・高速なタスク管理ツールです。
「ローカルファースト」を原則とし、オフラインで動作し、CLIとWeb UIの両方から同じデータを操作できます。

## 2. 技術スタック (Universal)
開発が進んでも変更されにくい、コアとなる技術選定です。

- **言語**: Go (Latest stable)
- **データベース**: SQLite (CGO不要の `modernc.org/sqlite` を使用)
- **CLIフレームワーク**: `spf13/cobra`
- **Web UI**: Server-Side Rendering (Go html/template) + `htmx`
- **ORM/DAL**: `sqlc` による型安全な SQL 生成
- **構成管理**: `toml`

## 3. アーキテクチャ構成
プロジェクトは「モジュラーモノリス」構成を採用しており、CLIとWebサーバーがロジックを共有しています。

- **cmd/**: エントリーポイント。`main.go` が CLI とサーバーの起動を振り分けます。
- **cli/**: CLI コマンド (`add`, `ls`, `web` 等) の定義。
- **srv/**: Webサーバーのロジック。HTMX を前提とした HTML を返します。
- **db/**: データベースアクセス層。`sqlc` で生成されたコードが含まれます。
- **core/** (または概念的なドメイン層): データベースモデルに強く依存しています。

## 4. ドメインモデル (Core Entities)
主要なデータモデルとその関係性です。

### Project (プロジェクト)
タスクをまとめるコンテナです。
- デフォルトで `default` プロジェクトが存在します。
- **制約**: 名前はユニークである必要があります。

### Task (タスク)
管理対象の最小単位です。
- **Status (ステータス)**: 以下の3状態を持ちます。
  - `todo` (未着手)
  - `doing` (進行中)
  - `done` (完了)
- **属性**: タイトル、本文、期限 (`due_at`)、作成日時、更新日時。
- **検索**: SQLite FTS5 による全文検索が可能です。

### Tag (タグ)
タスクに付与できるラベルです。
- 多対多の関係 (`task_tags` テーブル) で管理されます。

## 5. 重要な原則と制約
- **シングルバイナリ**: 全てが1つの実行ファイルに含まれます（Webアセット含む）。
- **XDG準拠**: 設定ファイルやデータファイルは XDG Base Directory Specification に従います。
  - Config: `~/.config/mini-kanban/`
  - Data: `~/.local/share/mini-kanban/`
- **Local First**: 外部APIへの依存を極力排除します。
