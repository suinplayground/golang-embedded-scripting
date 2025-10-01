# VPC別ConfigMapグループ化・マージ処理（Starlark版）

[Starlark](https://github.com/google/starlark-go) を使ってKubernetesのConfigMapをVPC ID別にグループ化してマージするサンプルです。

## 概要

このプログラムは以下の処理を行います：

1. **VPC IDでグループ化**: `metadata.labels["vpc-id"]`が同じConfigMapをグループ化
2. **サブネットデータの抽出**: 各ConfigMapから`subnet-id`キーのみを取り出す
3. **データのマージ**: 同じVPC内のサブネット情報をマージ
4. **新しいConfigMap作成**: マージ結果を`name`がVPC IDの新しいConfigMapとして生成

## Starlarkとは

[Starlark](https://github.com/bazel-build/starlark) は、Googleが開発したPython風の設定言語です：

- **決定性**: 同じ入力に対して常に同じ出力を生成
- **安全**: サンドボックス化され、ファイルI/Oやネットワークアクセスなし
- **シンプル**: Pythonのサブセットで学習が容易
- **並列実行可能**: 独立したスレッド間で並列実行可能

Bazel、Buck2、Isopod などで採用されています。

## TypeScript版との比較

### TypeScript版
```typescript
function groupByVpcAndMerge(configMaps: ConfigMap[]): ConfigMap[] {
    const vpcGroups = new Map<string, ConfigMap[]>();
    // ...
}
```

### Starlark版
```python
def group_by_vpc_and_merge(config_maps):
    vpc_groups = {}
    # ...
```

**主な違い:**
- Starlarkは型注釈不要（動的型付け）
- Pythonライクな構文で読みやすい
- `Map` の代わりに `dict` を使用
- トランスパイル不要（直接実行）

## セットアップ

```bash
cd starlark
go mod tidy
```

## 実行

```bash
go run main.go
```

## 出力例

```
=== VPC別ConfigMapグループ化・マージ処理（Starlark版）===

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

## Starlarkスクリプトの詳細

```python
def group_by_vpc_and_merge(config_maps):
    # VPC IDでグループ化
    vpc_groups = {}
    
    for config_map in config_maps:
        vpc_id = config_map.get("metadata", {}).get("labels", {}).get("vpc-id")
        
        if not vpc_id:
            print("⚠ vpc-idラベルがありません:", config_map.get("metadata", {}).get("name"))
            continue
        
        if vpc_id not in vpc_groups:
            vpc_groups[vpc_id] = []
        
        vpc_groups[vpc_id].append(config_map)
    
    # グループごとにマージ
    merged_config_maps = []
    
    for vpc_id, config_maps_in_vpc in vpc_groups.items():
        merged_data = {}
        
        # subnet-idのみを抽出
        for cm in config_maps_in_vpc:
            cm_name = cm.get("metadata", {}).get("name", "")
            cm_data = cm.get("data", {})
            
            for key, value in cm_data.items():
                if key == "subnet-id":
                    merged_data[cm_name + "." + key] = value
        
        # VPC IDを名前とする新しいConfigMapを作成
        merged_config_maps.append({
            "apiVersion": "v1",
            "kind": "ConfigMap",
            "metadata": {
                "name": vpc_id,
                "namespace": "default",
                "labels": {
                    "vpc-id": vpc_id,
                    "merged": "true"
                }
            },
            "data": merged_data
        })
    
    return merged_config_maps
```

## Starlarkの特徴

### ✅ メリット

1. **決定性**: 副作用なし、同じ入力→同じ出力
2. **安全性**: サンドボックス化、ファイルI/Oなし
3. **シンプル**: Pythonライクで学習コストが低い
4. **並列実行**: スレッドセーフ
5. **トランスパイル不要**: 直接実行可能

### ❌ 制限

1. **I/O禁止**: ファイル読み書き、ネットワークアクセス不可
2. **標準ライブラリ制限**: CPythonの標準ライブラリは使えない
3. **型チェックなし**: 実行時エラーのみ
4. **while文なし**: 無限ループ防止のため（forループのみ）

## TypeScript vs Starlark 比較表

| 項目 | TypeScript (goja) | Starlark |
|------|------------------|----------|
| 型システム | 静的型付け（オプション） | 動的型付けのみ |
| トランスパイル | 必要（esbuild） | 不要 |
| エラー表示 | sourcemapで元の行番号 | 直接実行なので明確 |
| 構文 | JavaScript/TypeScript | Python風 |
| 安全性 | JavaScript実行リスク | 完全サンドボックス |
| 並列実行 | 制限あり | 完全並列可能 |
| 用途 | 汎用スクリプト | 設定・ビルド定義 |

## ユースケース

- **Bazelビルド設定**: BUILD ファイル
- **インフラ定義**: Kubernetes リソース生成
- **設定管理**: 複雑な設定ロジック
- **データ変換**: 決定性が必要な処理

## 技術スタック

- **[starlark-go](https://github.com/google/starlark-go)**: Starlark インタプリタ
- **Go標準ライブラリ**: JSON処理

## 参考リンク

- [Starlark言語仕様](https://github.com/bazelbuild/starlark/blob/master/spec.md)
- [starlark-go GitHub](https://github.com/google/starlark-go)
- [Bazel: Starlark言語](https://bazel.build/rules/language)
