# git-smartmsg

🤖 OpenAI APIを使用したAI搭載Gitコミットメッセージ改善ツール

**git-smartmsg** は、Gitのコミット履歴を解析してAIを使用し改善されたコミットメッセージを生成するコマンドラインツールです。プランニング（コミット分析と提案生成）と適用（新しいブランチで改善されたメッセージを使って履歴を書き換え）の2段階で動作します。

[English README is here](./README.md) | [英語版READMEはこちら](./README.md)

## 特徴

- 🎯 **AI搭載**: OpenAI APIを使用して意味のあるコミットメッセージを生成
- 🔒 **安全設計**: 新しいブランチを作成し、現在のブランチは変更しません
- 📋 **2段階プロセス**: まずプランを作成、確認後に適用
- 🎨 **絵文字モード**: コミット分類のための視覚的な絵文字プレフィックス（オプション）
- ⚡ **Conventional Commits**: conventional commit形式をサポート
- 🛡️ **履歴保持**: 作成者情報とタイムスタンプを維持
- 🔄 **柔軟な範囲指定**: 特定のコミット範囲または最近のコミットを処理

## インストール

### 前提条件

- Go 1.25以上
- Git
- OpenAI APIキー

### ソースからビルド

```bash
git clone https://github.com/yourusername/git-smartmsg
cd git-smartmsg
go build -o git-smartmsg main.go
```

### PATHに追加（オプション）

```bash
# PATHの通ったディレクトリに移動
sudo mv git-smartmsg /usr/local/bin/
# またはシンボリックリンクを作成
ln -s $(pwd)/git-smartmsg /usr/local/bin/git-smartmsg
```

## 環境変数

### 必須

```bash
export OPENAI_API_KEY="your-openai-api-key"
```

### オプション

```bash
# カスタムAPIエンドポイント（OpenAI互換サービス用）
export OPENAI_API_BASE="https://api.openai.com/v1"

# デフォルトモデル（デフォルト: gpt-5-nano）
export OPENAI_MODEL="gpt-4o"
```

## クイックスタート

1. **Gitリポジトリに移動**
   ```bash
   cd your-git-repo
   ```

2. **改善されたコミットメッセージを生成**
   ```bash
   ./git-smartmsg plan --limit 5
   ```

3. **生成されたプランを確認**
   ```bash
   cat plan.json
   ```

4. **改善されたメッセージを新しいブランチに適用**
   ```bash
   ./git-smartmsg apply --branch improved-messages
   ```

## 使用方法

### コマンド概要

```bash
git-smartmsg <サブコマンド> [オプション]
```

### サブコマンド

#### `plan` - AIコミットメッセージ生成

```bash
git-smartmsg plan [オプション]
```

**オプション:**
- `--limit <n>`: HEADから含めるコミット数（デフォルト: 20）
- `--range <範囲>`: 明示的なgit範囲指定（例: `HEAD~10..HEAD`）
- `--model <モデル>`: 使用するLLMモデル（デフォルト: 環境変数または`gpt-5-nano`）
- `--emoji`: 絵文字スタイルのコミットメッセージを使用
- `--allow-merges`: マージコミットを含める（非推奨）
- `--out <ファイル>`: プランファイルの出力先（デフォルト: `plan.json`）
- `--timeout <期間>`: コミット毎のAIタイムアウト（デフォルト: 25秒）

#### `apply` - プランを新しいブランチに適用

```bash
git-smartmsg apply [オプション]
```

**オプション:**
- `--branch <名前>`: 新しいブランチ名（必須）
- `--in <ファイル>`: プランファイルのパス（デフォルト: `plan.json`）
- `--allow-merges`: マージコミットの保持を試行（実験的機能）

## 使用例

### 基本的な使用方法

```bash
# 最新の10コミットを改善
./git-smartmsg plan --limit 10

# 特定のモデルを使用
./git-smartmsg plan --limit 5 --model gpt-4o

# 新しいブランチに適用
./git-smartmsg apply --branch feature/improved-commits
```

### 高度な使用方法

```bash
# 特定の範囲を処理
./git-smartmsg plan --range v1.0.0..HEAD

# 絵文字モードを使用
./git-smartmsg plan --emoji --limit 15

# マージコミットを含める（実験的）
./git-smartmsg plan --allow-merges --limit 20
./git-smartmsg apply --allow-merges --branch with-merges
```

### ワークフロー例

```bash
# 1. 改善したいコミットを確認
git log --oneline -10

# 2. 絵文字モードでプランを生成
./git-smartmsg plan --emoji --limit 10

# 3. 提案を確認
cat plan.json | jq '.items[] | {old: .old_message, new: .new_message}'

# 4. 新しいブランチに適用
./git-smartmsg apply --branch feature/ai-improved-messages

# 5. 新しいブランチを確認
git log --oneline -10

# 6. 満足したらプッシュ（オプション）
git push --force-with-lease origin feature/ai-improved-messages
```

## 絵文字モード

`--emoji` フラグを使用すると、コミットメッセージにコンテキストに応じた絵文字が付加されます：

| 絵文字 | コード | 用途 |
|-------|------|------|
| 🎨 | `:art:` | コード構造・フォーマット改善 |
| 🐛 | `:bug:` | バグ修正 |
| 🔥 | `:fire:` | コードやファイルの削除 |
| 📝 | `:memo:` | ドキュメント作成 |
| ⚡ | `:zap:` | パフォーマンス改善 |
| ✅ | `:white_check_mark:` | テスト追加 |
| 🔒 | `:lock:` | セキュリティ修正 |
| ⬆️ | `:arrow_up:` | 依存関係のアップグレード |

**出力例:**
```
🎨 ユーザー認証モジュールをリファクタリング
🐛 データパーサーのnullポインタ例外を修正
📝 v2エンドポイント向けAPIドキュメントを更新
✅ 支払い処理のユニットテストを追加
```

## 安全性とベストプラクティス

### 安全機能

- **クリーンワークツリー必須**: 未コミット変更がないことを確認（`plan.json`は無視）
- **新しいブランチ作成**: 現在のブランチは決して変更しません
- **作成者情報の保持**: 元の作成者情報とタイムスタンプを維持
- **バックアップ推奨**: 元のコミットは引き続きアクセス可能

### ベストプラクティス

1. **適用前の確認**: `apply`実行前に必ず`plan.json`をチェック
2. **小さなバッチで処理**: より良い結果のために10-20コミットずつ処理
3. **ブランチのテスト**: マージ前に生成されたブランチを確認
4. **チーム調整**: 履歴書き換えのフォースプッシュ前にチームと調整
5. **バックアップ**: 大規模な書き換え前にバックアップブランチの作成を検討

### 安全なフォースプッシュ

```bash
# より安全なフォースプッシュのために --force-with-lease を使用
git push --force-with-lease origin your-branch-name
```

## ファイル構造

```
.
├── main.go           # 完全なアプリケーション
├── go.mod            # Go依存関係
├── go.sum            # 依存関係チェックサム
├── plan.json         # 生成されたプラン（gitで無視）
├── CLAUDE.md         # Claude Codeガイダンス
├── README.md         # 英語版README
└── README.ja.md      # このファイル
```

## トラブルシューティング

### よくある問題

**"worktree is not clean"**
- 最初に変更をコミットまたはstash
- `plan.json`は自動的に無視されます

**"AI failed for commit"**
- OpenAI APIキーを確認
- APIクォータ/制限を確認
- より小さなバッチサイズを試す

**"cherry-pick failed"**
- 複雑な競合は手動解決が必要な場合があります
- デフォルト設定でマージコミットを除外することを検討

### ヘルプの取得

```bash
# 使用可能なコマンドを表示
./git-smartmsg

# コマンド固有のヘルプを表示
./git-smartmsg plan --help
```

## 貢献

1. リポジトリをフォーク
2. 機能ブランチを作成（`git checkout -b feature/amazing-feature`）
3. 変更をコミット（`git commit -m 'Add amazing feature'`）
4. ブランチにプッシュ（`git push origin feature/amazing-feature`）
5. プルリクエストを作成

## ライセンス

このプロジェクトはMITライセンスの下でライセンスされています - 詳細は[LICENSE](LICENSE)ファイルを参照してください。

## 謝辞

- [OpenAI API](https://openai.com/api/) - AI搭載メッセージ生成のため
- [Conventional Commits](https://www.conventionalcommits.org/) - コミットメッセージ標準のため
- [gitmoji](https://gitmoji.dev/) - 絵文字のインスピレーションのため