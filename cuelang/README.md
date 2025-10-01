# VPC別ConfigMapグループ化・マージ処理（CUE版）

[CUE](https://cuelang.org/) を使ってKubernetesのConfigMapをVPC ID別にグループ化してマージするサンプルです。

## 概要

このプログラムは以下の処理を**全てCUE言語で実装**しています：

1. **VPC IDでグループ化**: `metadata.labels["vpc-id"]`が同じConfigMapをCUEでグループ化
2. **サブネットデータの抽出**: 各ConfigMapから`subnet-id`キーのみをCUEで取り出す
3. **データのマージ**: 同じVPC内のサブネット情報をCUEでマージ
4. **CUEスキーマでバリデーション**: CUEの型システムによる自動検証
5. **新しいConfigMap作成**: マージ結果を`name`がVPC IDの新しいConfigMapとしてCUEで生成

## CUEとは

[CUE (Configure, Unify, Execute)](https://cuelang.org/) は、データのスキーマ定義、バリデーション、生成を行うためのオープンソース言語です：

- **型安全**: 静的型チェックによるデータ検証
- **統合**: JSONスーパーセット、YAMLとの相互変換可能
- **制約ベース**: 値ではなく制約を定義
- **データ生成**: テンプレートからデータを生成
- **バリデーション**: スキーマに基づく自動検証

KubernetesやHelm、Terraform などの設定管理で使用されています。

## セットアップ

```bash
cd cuelang
go mod tidy
```

## 実行

```bash
go run main.go
```

## 出力例

```
=== VPC別ConfigMapグループ化・マージ処理（CUE版）===

入力ConfigMap数: 5個
1. subnet-az1a (vpc-id: vpc-12345, subnet-id: subnet-aaa111)
2. subnet-az1c (vpc-id: vpc-12345, subnet-id: subnet-ccc333)
3. subnet-az1d (vpc-id: vpc-12345, subnet-id: subnet-ddd444)
4. subnet-vpc2-az1a (vpc-id: vpc-67890, subnet-id: subnet-bbb222)
5. subnet-vpc2-az1c (vpc-id: vpc-67890, subnet-id: subnet-eee555)

📦 VPC ID: vpc-12345 - ConfigMap数: 3
  ✓ 追加: subnet-az1a.subnet-id = subnet-aaa111
  ✓ 追加: subnet-az1c.subnet-id = subnet-ccc333
  ✓ 追加: subnet-az1d.subnet-id = subnet-ddd444
📦 VPC ID: vpc-67890 - ConfigMap数: 2
  ✓ 追加: subnet-vpc2-az1a.subnet-id = subnet-bbb222
  ✓ 追加: subnet-vpc2-az1c.subnet-id = subnet-eee555

✅ 合計 2 個のVPCグループを作成
✅ 処理完了

マージ済ConfigMap: 2個

--- ConfigMap 1 ---
Name: vpc-12345
Namespace: default
Labels:
  vpc-id: vpc-12345
  merged: true
Data:
  subnet-az1a.subnet-id: subnet-aaa111
  subnet-az1c.subnet-id: subnet-ccc333
  subnet-az1d.subnet-id: subnet-ddd444

--- ConfigMap 2 ---
Name: vpc-67890
Namespace: default
Labels:
  vpc-id: vpc-67890
  merged: true
Data:
  subnet-vpc2-az1a.subnet-id: subnet-bbb222
  subnet-vpc2-az1c.subnet-id: subnet-eee555
```

## CUE処理ロジックの詳細

### 1. VPC IDでグループ化

```cue
// VPC IDごとにグループ化してConfigMapを集約
vpcGroups: {
	for cm in inputConfigMaps {
		let vid = cm.metadata.labels["vpc-id"]
		if vid != _|_ {
			"\(vid)": {
				vpcId:      vid
				namespace:  cm.metadata.namespace
				configMaps: [...#ConfigMap]
			}
		}
	}
}
```

- `for cm in inputConfigMaps`: 全ConfigMapをループ
- `let vid = ...`: VPC IDを変数に格納
- `if vid != _|_`: VPC IDが存在する場合のみ処理
- `"\(vid)": {...}`: VPC IDをキーとする動的フィールド

### 2. ConfigMapをグループに集約

```cue
enrichedGroups: {
	for vid, group in vpcGroups {
		"\(vid)": {
			vpcId:     group.vpcId
			namespace: group.namespace
			configMaps: [
				for cm in inputConfigMaps
				if cm.metadata.labels["vpc-id"] == vid {cm}
			]
		}
	}
}
```

- 内部forループでフィルタリング
- `if cm.metadata.labels["vpc-id"] == vid`: 同じVPC IDのConfigMapのみ

### 3. subnet-idのマージ

```cue
mergedConfigMaps: [
	for vid, group in enrichedGroups {
		{
			apiVersion: "v1"
			kind:       "ConfigMap"
			metadata: {
				name:      vid
				namespace: group.namespace
				labels: {
					"vpc-id": vid
					merged:   "true"
				}
			}
			data: {
				for cm in group.configMaps
				for key, value in cm.data
				if key == "subnet-id" {
					"\(cm.metadata.name).\(key)": value
				}
			}
		}
	},
]
```

- 二重ループ: ConfigMapとdataフィールド
- `if key == "subnet-id"`: subnet-idのみ抽出
- `"\(cm.metadata.name).\(key)"`: 動的キー生成

## CUEの特徴

### ✅ メリット

1. **型安全性**: コンパイル時の型チェック
2. **制約ベース**: 値の範囲や形式を制約で定義
3. **統合と検証**: データの統合と自動検証
4. **JSONスーパーセット**: JSON互換で学習コストが低い
5. **Kubernetes統合**: kustomize、helmとの統合が容易

### 🎯 ユースケース

- **Kubernetes設定**: マニフェストのバリデーションと生成
- **CI/CD**: パイプライン設定の検証
- **API定義**: OpenAPIスキーマの管理
- **設定管理**: アプリケーション設定の型安全な管理

## TypeScript vs Starlark vs CUE 比較

| 項目 | TypeScript (goja) | Starlark (starlark-go) | CUE (cuelang) |
|------|------------------|----------------------|---------------|
| **構文** | JavaScript/TypeScript | Python風 | JSON拡張 |
| **型システム** | オプショナル静的型付け | 動的型付け | 静的型付け |
| **主な用途** | 汎用スクリプト | ビルド設定 | データ検証・生成 |
| **バリデーション** | 実行時 | 実行時 | コンパイル時 |
| **決定性** | なし | あり | あり |
| **安全性** | JavaScript実行 | サンドボックス | 制約ベース |
| **Kubernetes** | 手動統合 | 手動統合 | ネイティブ対応 |

## CUE vs 他言語の特徴比較

### データ定義方法

**JSON:**
```json
{
  "name": "vpc-12345",
  "namespace": "default"
}
```

**YAML:**
```yaml
name: vpc-12345
namespace: default
```

**CUE:**
```cue
name:      "vpc-12345"
namespace: "default"
namespace: string  // 型定義も可能
```

### スキーマ定義

**TypeScript:**
```typescript
interface ConfigMap {
  name: string;
  namespace: string;
}
```

**CUE:**
```cue
ConfigMap: {
  name:      string & !=""  // 空文字列禁止
  namespace: string | *"default"  // デフォルト値
}
```

## 実装のポイント

### 1. CUEコンテキストの作成

```go
ctx := cuecontext.New()
```

### 2. CUEスクリプトのコンパイル

```go
value := ctx.CompileString(cueScript)
if value.Err() != nil {
    return fmt.Errorf("CUEコンパイルエラー: %w", value.Err())
}
```

### 3. データの設定とバリデーション

```go
// Goデータをエンコード
configMapsValue := ctx.Encode(configMapsInterface)

// CUEスキーマと統合
unified := schemaValue.Unify(cmValue)

// バリデーション
if err := validated.Validate(); err != nil {
    return fmt.Errorf("バリデーションエラー: %w", err)
}
```

### 4. 結果の取得

```go
resultValue := unified.LookupPath(cue.ParsePath("result"))
var result []ConfigMap
resultValue.Decode(&result)
```

## ユースケース

- **Kubernetes設定管理**: ConfigMap/Secretの統合管理
- **Helmチャート**: バリューファイルのバリデーション
- **CI/CD設定**: パイプライン定義の型安全性確保
- **マルチクラスタ管理**: 複数環境の設定統合

## 参考リンク

- [CUE公式サイト](https://cuelang.org/)
- [CUE GitHub](https://github.com/cue-lang/cue)
- [CUE Go API](https://pkg.go.dev/cuelang.org/go/cue)
- [CUE Tutorial](https://cuetorials.com/)
- [Kubernetes + CUE](https://cuelang.org/docs/integrations/kubernetes/)

## 技術スタック

- **[cuelang.org/go](https://pkg.go.dev/cuelang.org/go)**: CUE Go API
- **Go標準ライブラリ**: JSON処理
