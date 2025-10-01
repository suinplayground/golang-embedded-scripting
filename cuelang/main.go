package main

import (
	"encoding/json"
	"fmt"
	"log"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

// Kubernetes ConfigMap構造体
type ConfigMap struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   Metadata          `json:"metadata"`
	Data       map[string]string `json:"data"`
}

// Kubernetes Metadata
type Metadata struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

func main() {
	// サンプルConfigMapデータ（VPC別のサブネット情報）
	configMaps := []ConfigMap{
		{
			APIVersion: "v1",
			Kind:       "ConfigMap",
			Metadata: Metadata{
				Name:      "subnet-az1a",
				Namespace: "default",
				Labels: map[string]string{
					"vpc-id": "vpc-12345",
					"az":     "ap-northeast-1a",
				},
			},
			Data: map[string]string{
				"subnet-id":   "subnet-aaa111",
				"cidr-block":  "10.0.1.0/24",
				"description": "Subnet in AZ 1a",
			},
		},
		{
			APIVersion: "v1",
			Kind:       "ConfigMap",
			Metadata: Metadata{
				Name:      "subnet-az1c",
				Namespace: "default",
				Labels: map[string]string{
					"vpc-id": "vpc-12345",
					"az":     "ap-northeast-1c",
				},
			},
			Data: map[string]string{
				"subnet-id":   "subnet-ccc333",
				"cidr-block":  "10.0.3.0/24",
				"description": "Subnet in AZ 1c",
			},
		},
		{
			APIVersion: "v1",
			Kind:       "ConfigMap",
			Metadata: Metadata{
				Name:      "subnet-az1d",
				Namespace: "default",
				Labels: map[string]string{
					"vpc-id": "vpc-12345",
					"az":     "ap-northeast-1d",
				},
			},
			Data: map[string]string{
				"subnet-id":   "subnet-ddd444",
				"cidr-block":  "10.0.4.0/24",
				"description": "Subnet in AZ 1d",
			},
		},
		{
			APIVersion: "v1",
			Kind:       "ConfigMap",
			Metadata: Metadata{
				Name:      "subnet-vpc2-az1a",
				Namespace: "default",
				Labels: map[string]string{
					"vpc-id": "vpc-67890",
					"az":     "ap-northeast-1a",
				},
			},
			Data: map[string]string{
				"subnet-id":   "subnet-bbb222",
				"cidr-block":  "192.168.1.0/24",
				"description": "Subnet in VPC2 AZ 1a",
			},
		},
		{
			APIVersion: "v1",
			Kind:       "ConfigMap",
			Metadata: Metadata{
				Name:      "subnet-vpc2-az1c",
				Namespace: "default",
				Labels: map[string]string{
					"vpc-id": "vpc-67890",
					"az":     "ap-northeast-1c",
				},
			},
			Data: map[string]string{
				"subnet-id":   "subnet-eee555",
				"cidr-block":  "192.168.2.0/24",
				"description": "Subnet in VPC2 AZ 1c",
			},
		},
	}

	fmt.Println("=== VPC別ConfigMapグループ化・マージ処理（CUE版）===")
	fmt.Println()

	fmt.Printf("入力ConfigMap数: %d個\n", len(configMaps))
	for i, cm := range configMaps {
		vpcID := cm.Metadata.Labels["vpc-id"]
		fmt.Printf("%d. %s (vpc-id: %s, subnet-id: %s)\n",
			i+1, cm.Metadata.Name, vpcID, cm.Data["subnet-id"])
	}
	fmt.Println()

	// CUEで処理
	mergedConfigMaps, err := processWithCUE(configMaps)
	if err != nil {
		log.Fatalf("エラー: %v\n", err)
	}

	fmt.Println("✅ 処理完了")
	fmt.Println()

	// 結果を表示
	fmt.Printf("マージ済ConfigMap: %d個\n", len(mergedConfigMaps))
	fmt.Println()

	for i, cm := range mergedConfigMaps {
		fmt.Printf("--- ConfigMap %d ---\n", i+1)
		fmt.Printf("Name: %s\n", cm.Metadata.Name)
		fmt.Printf("Namespace: %s\n", cm.Metadata.Namespace)
		fmt.Printf("Labels:\n")
		for k, v := range cm.Metadata.Labels {
			fmt.Printf("  %s: %s\n", k, v)
		}
		fmt.Printf("Data:\n")
		for k, v := range cm.Data {
			fmt.Printf("  %s: %s\n", k, v)
		}
		fmt.Println()
	}
}

// CUEでConfigMapを処理（VPC別にグループ化してマージ）
func processWithCUE(configMaps []ConfigMap) ([]ConfigMap, error) {
	ctx := cuecontext.New()

	// CUEスクリプト（VPC別グループ化とマージのロジック）
	cueScript := `
package process

// 入力ConfigMap
#ConfigMap: {
	apiVersion: string
	kind:       string
	metadata: {
		name:      string
		namespace: string
		labels: [string]: string
	}
	data: [string]: string
}

// 入力データ
inputConfigMaps: [...#ConfigMap]

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

// ConfigMapをグループに追加
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

// マージ処理
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
`

	// CUEスクリプトをコンパイル
	value := ctx.CompileString(cueScript)
	if value.Err() != nil {
		return nil, fmt.Errorf("CUEコンパイルエラー: %w", value.Err())
	}

	// ConfigMapをJSONに変換してCUEに渡す
	configMapsJSON, err := json.Marshal(configMaps)
	if err != nil {
		return nil, fmt.Errorf("JSON変換エラー: %w", err)
	}

	var configMapsInterface []interface{}
	json.Unmarshal(configMapsJSON, &configMapsInterface)

	// inputConfigMapsに値を設定
	configMapsValue := ctx.Encode(configMapsInterface)
	filled := value.FillPath(cue.ParsePath("inputConfigMaps"), configMapsValue)
	if filled.Err() != nil {
		return nil, fmt.Errorf("CUE Fill エラー: %w", filled.Err())
	}

	// enrichedGroupsを取得してログ出力
	enrichedGroupsValue := filled.LookupPath(cue.ParsePath("enrichedGroups"))
	if enrichedGroupsValue.Err() != nil {
		return nil, fmt.Errorf("enrichedGroups取得エラー: %w", enrichedGroupsValue.Err())
	}

	// enrichedGroupsをデコードしてログ出力
	iter, _ := enrichedGroupsValue.Fields()
	for iter.Next() {
		vpcId := iter.Label()
		group := iter.Value()

		configMapsField := group.LookupPath(cue.ParsePath("configMaps"))
		var groupConfigMaps []map[string]interface{}
		configMapsField.Decode(&groupConfigMaps)

		fmt.Printf("📦 VPC ID: %s - ConfigMap数: %d\n", vpcId, len(groupConfigMaps))

		for _, cm := range groupConfigMaps {
			metadata := cm["metadata"].(map[string]interface{})
			data := cm["data"].(map[string]interface{})
			name := metadata["name"].(string)

			if subnetId, ok := data["subnet-id"]; ok {
				fmt.Printf("  ✓ 追加: %s.subnet-id = %s\n", name, subnetId)
			}
		}
	}

	// mergedConfigMapsを取得
	mergedPath := cue.ParsePath("mergedConfigMaps")
	mergedValue := filled.LookupPath(mergedPath)
	if mergedValue.Err() != nil {
		return nil, fmt.Errorf("mergedConfigMaps取得エラー: %w", mergedValue.Err())
	}

	// 結果をデコード
	var mergedInterface []interface{}
	if err := mergedValue.Decode(&mergedInterface); err != nil {
		return nil, fmt.Errorf("デコードエラー: %w", err)
	}

	// JSONを経由してGoの構造体に変換
	resultJSON, err := json.Marshal(mergedInterface)
	if err != nil {
		return nil, fmt.Errorf("JSON変換エラー: %w", err)
	}

	var mergedConfigMaps []ConfigMap
	if err := json.Unmarshal(resultJSON, &mergedConfigMaps); err != nil {
		return nil, fmt.Errorf("JSON Unmarshalエラー: %w", err)
	}

	fmt.Printf("\n✅ 合計 %d 個のVPCグループを作成\n", len(mergedConfigMaps))

	return mergedConfigMaps, nil
}
