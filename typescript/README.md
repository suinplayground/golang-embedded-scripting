# VPC別ConfigMapグループ化・マージ処理（TypeScript + Goja）

KubernetesのConfigMapをVPC ID別にグループ化してマージするサンプルです。

## 概要

このプログラムは以下の処理を行います：

1. **VPC IDでグループ化**: `metadata.labels["vpc-id"]`が同じConfigMapをグループ化
2. **サブネットデータの抽出**: 各ConfigMapから`subnet-id`キーのみを取り出す
3. **データのマージ**: 同じVPC内のサブネット情報をマージ
4. **新しいConfigMap作成**: マージ結果を`name`がVPC IDの新しいConfigMapとして生成

## 処理フロー

```
入力:
  ConfigMap[](各サブネット情報)
    ↓
TypeScript処理:
  - vpc-idラベルでグループ化
  - subnet-idデータのみ抽出
  - VPC単位でマージ
    ↓
出力:
  ConfigMap[](VPC単位にマージ済み)
```

## TypeScript処理ロジック

```typescript
function groupByVpcAndMerge(configMaps: ConfigMap[]): ConfigMap[] {
    // VPC IDでグループ化
    const vpcGroups = new Map<string, ConfigMap[]>();

    for (const configMap of configMaps) {
        const vpcId = configMap.metadata.labels?.["vpc-id"];
        if (vpcId) {
            vpcGroups.get(vpcId)?.push(configMap);
        }
    }

    // グループごとにマージ
    const mergedConfigMaps: ConfigMap[] = [];
    
    for (const [vpcId, configMapsInVpc] of vpcGroups) {
        const mergedData: { [key: string]: string } = {};
        
        // subnet-idのみを抽出
        for (const cm of configMapsInVpc) {
            for (const [key, value] of Object.entries(cm.data)) {
                if (key === "subnet-id") {
                    mergedData[cm.metadata.name + ".subnet-id"] = value;
                }
            }
        }
        
        // VPC IDを名前とする新しいConfigMapを作成
        mergedConfigMaps.push({
            apiVersion: "v1",
            kind: "ConfigMap",
            metadata: {
                name: vpcId,
                labels: { "vpc-id": vpcId, "merged": "true" }
            },
            data: mergedData
        });
    }

    return mergedConfigMaps;
}
```

## セットアップ

```bash
cd k8s-typescript
go mod tidy
```

## 実行

```bash
go run main.go
```

## 入力データ例

```yaml
# VPC vpc-12345 のサブネット
- name: subnet-az1a
  labels:
    vpc-id: vpc-12345
  data:
    subnet-id: subnet-aaa111
    cidr-block: 10.0.1.0/24

- name: subnet-az1c
  labels:
    vpc-id: vpc-12345
  data:
    subnet-id: subnet-ccc333
    cidr-block: 10.0.3.0/24

# VPC vpc-67890 のサブネット
- name: subnet-vpc2-az1a
  labels:
    vpc-id: vpc-67890
  data:
    subnet-id: subnet-bbb222
    cidr-block: 192.168.1.0/24
```

## 出力例

```
=== VPC別ConfigMapグループ化・マージ処理 ===

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

## ユースケース

- **VPC別サブネット一覧の集約**: 複数のサブネット情報をVPC単位で集約
- **インフラ構成の可視化**: VPCごとのリソース構成を把握
- **Terraform/Pulumi連携**: インフラ定義からKubernetesリソース生成
- **マルチクラスタ管理**: 複数クラスタのネットワーク情報を統合

## 特徴

✨ **型安全なグループ化**: TypeScriptの型定義でVPC別グループ化を安全に実装  
✨ **柔軟なデータ抽出**: 特定キー（subnet-id）のみを抽出してマージ  
✨ **sourcemap対応**: エラー時にTypeScriptの行番号を表示  
✨ **実用的**: 実際のKubernetesリソース構造に基づいた実装

## 技術スタック

- **goja**: JavaScriptランタイム
- **esbuild**: TypeScript→JavaScriptトランスパイラ
- **go-sourcemap**: エラー位置のマッピング