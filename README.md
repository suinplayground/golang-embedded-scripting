# Golang Embedded Scripting

Goアプリケーションに**スクリプト言語を組み込む**3つのアプローチを比較するサンプルプロジェクトです。

## 📚 概要

このリポジトリでは、同じタスク（Kubernetes ConfigMapのVPC別グループ化・マージ）を**3つの異なるスクリプト言語**で実装し、それぞれの特徴と使い分けを学べます：

| 言語 | 特徴 | 用途 |
|------|------|------|
| **[TypeScript](./typescript/)** | 型安全・汎用的 | 複雑なビジネスロジック |
| **[Starlark](./starlark/)** | 決定性・安全 | ビルド設定・設定管理 |
| **[CUE](./cuelang/)** | 制約ベース・検証 | データバリデーション・生成 |

## 🎯 実装タスク

すべての実装で以下の処理を行います：

1. **VPC IDでグループ化**: Kubernetes ConfigMapを`vpc-id`ラベルでグループ化
2. **データ抽出**: 各ConfigMapから`subnet-id`のみを抽出
3. **マージ**: VPC単位でサブネット情報をマージ
4. **新規作成**: VPC IDを名前とする新しいConfigMapを生成

### 入力例

```yaml
# VPC vpc-12345 のサブネット
- name: subnet-az1a
  labels: { vpc-id: vpc-12345 }
  data: { subnet-id: subnet-aaa111 }

- name: subnet-az1c
  labels: { vpc-id: vpc-12345 }
  data: { subnet-id: subnet-ccc333 }

# VPC vpc-67890 のサブネット  
- name: subnet-vpc2-az1a
  labels: { vpc-id: vpc-67890 }
  data: { subnet-id: subnet-bbb222 }
```

### 出力例

```yaml
# VPC単位にマージ
- name: vpc-12345
  labels: { vpc-id: vpc-12345, merged: "true" }
  data:
    subnet-az1a.subnet-id: subnet-aaa111
    subnet-az1c.subnet-id: subnet-ccc333

- name: vpc-67890
  labels: { vpc-id: vpc-67890, merged: "true" }
  data:
    subnet-vpc2-az1a.subnet-id: subnet-bbb222
```

## 🔧 各実装の詳細

### 1. TypeScript + Goja

**ディレクトリ**: [`typescript/`](./typescript/)

- **ランタイム**: [goja](https://github.com/dop251/goja) (Go実装のJavaScript VM)
- **トランスパイラ**: [esbuild](https://esbuild.github.io/)
- **特徴**:
  - ✅ 型安全な開発（TypeScript）
  - ✅ 豊富なエコシステム
  - ✅ sourcemap対応でエラー行番号を正確に表示
  - ❌ トランスパイルが必要

```bash
cd typescript
go run main.go
```

### 2. Starlark

**ディレクトリ**: [`starlark/`](./starlark/)

- **ランタイム**: [starlark-go](https://github.com/google/starlark-go)
- **特徴**:
  - ✅ 決定性（同じ入力→同じ出力）
  - ✅ サンドボックス化（ファイルI/O・ネットワーク禁止）
  - ✅ Pythonライクで学習コストが低い
  - ✅ 並列実行可能
  - ❌ 標準ライブラリ制限

```bash
cd starlark
go run main.go
```

### 3. CUE

**ディレクトリ**: [`cuelang/`](./cuelang/)

- **ランタイム**: [cuelang.org/go](https://cuelang.org/)
- **特徴**:
  - ✅ コンパイル時の型チェック・バリデーション
  - ✅ 制約ベースのデータ定義
  - ✅ Kubernetesネイティブ対応
  - ✅ JSONスーパーセットで学習が容易
  - ❌ 手続き型処理には不向き

```bash
cd cuelang
go run main.go
```

## 📊 比較表

### 言語特性

| 項目 | TypeScript | Starlark | CUE |
|------|-----------|----------|-----|
| **構文** | JavaScript/TypeScript | Python風 | JSON拡張 |
| **型システム** | オプショナル静的型付け | 動的型付け | 静的型付け（制約） |
| **実行** | トランスパイル→VM | 直接実行 | 評価 |
| **エラー検出** | 実行時 | 実行時 | コンパイル時 |
| **決定性** | ❌ | ✅ | ✅ |
| **サンドボックス** | ❌ | ✅ | ✅ |

### ユースケース

| 用途 | 推奨 | 理由 |
|------|------|------|
| **複雑なビジネスロジック** | TypeScript | 型安全・デバッグ容易・豊富なライブラリ |
| **ビルド設定** | Starlark | 決定性・並列実行・安全性 |
| **データバリデーション** | CUE | コンパイル時チェック・スキーマ定義 |
| **設定ファイル生成** | CUE | 制約ベース・テンプレート機能 |
| **プラグインシステム** | TypeScript/Starlark | 柔軟性・拡張性 |

### パフォーマンス

| 項目 | TypeScript | Starlark | CUE |
|------|-----------|----------|-----|
| **起動時間** | 遅い（トランスパイル） | 速い | 速い |
| **実行速度** | 速い（JIT） | 中程度 | 速い |
| **メモリ使用量** | 多い | 少ない | 中程度 |

## 🛠️ セットアップ

### 前提条件

- Go 1.23以上
- （オプション）[devbox](https://www.jetpack.io/devbox)

### インストール

```bash
# リポジトリをクローン
git clone https://github.com/suinplayground/golang-embedded-scripting.git
cd golang-embedded-scripting

# 各ディレクトリで依存関係をインストール
cd typescript && go mod tidy
cd ../starlark && go mod tidy
cd ../cuelang && go mod tidy
```

### devbox使用時

```bash
# 開発環境を起動
devbox shell

# 各実装を実行
cd typescript && go run main.go
cd ../starlark && go run main.go
cd ../cuelang && go run main.go
```

## 📖 各実装の詳細ドキュメント

- **[TypeScript版 README](./typescript/README.md)** - Goja + esbuild + sourcemap
- **[Starlark版 README](./starlark/README.md)** - 決定性・サンドボックス
- **[CUE版 README](./cuelang/README.md)** - 制約ベース・バリデーション

## 💡 使い分けガイド

### TypeScriptを選ぶべき場合

- 複雑なビジネスロジックを実装したい
- 型安全性が重要
- JavaScriptエコシステムを活用したい
- デバッグのしやすさを重視

### Starlarkを選ぶべき場合

- 決定性が必要（再現性のある処理）
- サンドボックス化が必須（セキュリティ）
- ビルド設定やCI/CD定義を書きたい
- 並列実行が必要

### CUEを選ぶべき場合

- データのバリデーションが主目的
- スキーマ定義から設定を生成したい
- Kubernetes/Helm/Terraformと統合したい
- コンパイル時のエラー検出を重視

## 🔗 参考リンク

### 公式ドキュメント

- **TypeScript**: https://www.typescriptlang.org/
- **Goja**: https://github.com/dop251/goja
- **esbuild**: https://esbuild.github.io/
- **Starlark**: https://github.com/bazelbuild/starlark
- **starlark-go**: https://github.com/google/starlark-go
- **CUE**: https://cuelang.org/
- **CUE Go API**: https://pkg.go.dev/cuelang.org/go/cue

### 関連プロジェクト

- **Bazel**: Starlarkを使ったビルドシステム
- **Buck2**: MetaのStarlarkベースビルドツール
- **Isopod**: StarlarkベースのKubernetes管理ツール
- **kustomize**: Kubernetesマニフェストのカスタマイズツール（CUE対応）

## 📝 ライセンス

このプロジェクトはサンプルコードです。自由にご利用ください。

## 🤝 コントリビューション

Issue・PRを歓迎します！

---

**作成者**: [@suinplayground](https://github.com/suinplayground)

