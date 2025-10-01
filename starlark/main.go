package main

import (
	"encoding/json"
	"fmt"
	"log"

	"go.starlark.net/starlark"
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

	fmt.Println("=== VPC別ConfigMapグループ化・マージ処理（Starlark版）===")
	fmt.Println()

	fmt.Printf("入力ConfigMap数: %d個\n", len(configMaps))
	for i, cm := range configMaps {
		vpcID := cm.Metadata.Labels["vpc-id"]
		fmt.Printf("%d. %s (vpc-id: %s, subnet-id: %s)\n",
			i+1, cm.Metadata.Name, vpcID, cm.Data["subnet-id"])
	}
	fmt.Println()

	// Starlarkで処理
	mergedConfigMaps, err := processWithStarlark(configMaps)
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

// StarlarkでConfigMapを処理（VPC別にグループ化してマージ）
func processWithStarlark(configMaps []ConfigMap) ([]ConfigMap, error) {
	// Starlarkスクリプト
	starlarkScript := `
# VPC別にConfigMapをグループ化してマージする関数
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
        print("📦 VPC ID:", vpc_id, "- ConfigMap数:", len(config_maps_in_vpc))
        
        # subnet-idのみを抽出してマージ
        merged_data = {}
        namespace = "default"
        
        for cm in config_maps_in_vpc:
            # namespaceを取得（最初のものを使用）
            if cm.get("metadata", {}).get("namespace"):
                namespace = cm["metadata"]["namespace"]
            
            # subnet-idキーのみを抽出
            cm_name = cm.get("metadata", {}).get("name", "")
            cm_data = cm.get("data", {})
            
            for key, value in cm_data.items():
                if key == "subnet-id":
                    # 元のConfigMap名をキー名として使用
                    new_key = cm_name + "." + key
                    merged_data[new_key] = value
                    print("  ✓ 追加:", new_key, "=", value)
        
        # マージ済みConfigMapを作成
        merged_config_map = {
            "apiVersion": "v1",
            "kind": "ConfigMap",
            "metadata": {
                "name": vpc_id,
                "namespace": namespace,
                "labels": {
                    "vpc-id": vpc_id,
                    "merged": "true"
                }
            },
            "data": merged_data
        }
        
        merged_config_maps.append(merged_config_map)
    
    print("\n✅ 合計", len(merged_config_maps), "個のVPCグループを作成")
    
    return merged_config_maps

# グローバルスコープで関数を定義
result = group_by_vpc_and_merge(input_config_maps)
`

	// Starlarkスレッドとグローバル環境を作成
	thread := &starlark.Thread{
		Name: "vpc-processor",
		Print: func(_ *starlark.Thread, msg string) {
			fmt.Println(msg)
		},
	}

	// ConfigMapをStarlarkの値に変換
	configMapsJSON, err := json.Marshal(configMaps)
	if err != nil {
		return nil, fmt.Errorf("JSON変換エラー: %w", err)
	}

	// ConfigMapをStarlarkの値に変換
	var configMapsInterface interface{}
	json.Unmarshal(configMapsJSON, &configMapsInterface)
	configMapsValue := goToStarlark(configMapsInterface)

	// グローバル変数を設定
	globals := starlark.StringDict{
		"input_config_maps": configMapsValue,
	}

	// Starlarkスクリプトを実行
	result, err := starlark.ExecFile(thread, "vpc-processor.star", starlarkScript, globals)
	if err != nil {
		return nil, fmt.Errorf("Starlark実行エラー: %w", err)
	}

	// 結果を取得
	resultValue, ok := result["result"]
	if !ok {
		return nil, fmt.Errorf("結果が見つかりません")
	}

	// Starlarkの値をGoの値に変換
	resultGo := starlarkToGo(resultValue)

	// JSONを経由してGoの構造体にデコード
	resultJSON, err := json.Marshal(resultGo)
	if err != nil {
		return nil, fmt.Errorf("JSON変換エラー: %w", err)
	}

	var mergedConfigMaps []ConfigMap
	err = json.Unmarshal(resultJSON, &mergedConfigMaps)
	if err != nil {
		return nil, fmt.Errorf("JSON デコードエラー: %w", err)
	}

	return mergedConfigMaps, nil
}

// GoのinterfaceをStarlarkの値に変換
func goToStarlark(v interface{}) starlark.Value {
	switch v := v.(type) {
	case nil:
		return starlark.None
	case bool:
		return starlark.Bool(v)
	case int:
		return starlark.MakeInt(v)
	case int64:
		return starlark.MakeInt64(v)
	case float64:
		return starlark.Float(v)
	case string:
		return starlark.String(v)
	case []interface{}:
		elems := make([]starlark.Value, len(v))
		for i, elem := range v {
			elems[i] = goToStarlark(elem)
		}
		return starlark.NewList(elems)
	case map[string]interface{}:
		dict := &starlark.Dict{}
		for key, val := range v {
			dict.SetKey(starlark.String(key), goToStarlark(val))
		}
		return dict
	default:
		return starlark.None
	}
}

// Starlarkの値をGoのinterfaceに変換
func starlarkToGo(v starlark.Value) interface{} {
	switch v := v.(type) {
	case starlark.NoneType:
		return nil
	case starlark.Bool:
		return bool(v)
	case starlark.Int:
		i, _ := v.Int64()
		return i
	case starlark.Float:
		return float64(v)
	case starlark.String:
		return string(v)
	case *starlark.List:
		result := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			result[i] = starlarkToGo(v.Index(i))
		}
		return result
	case *starlark.Dict:
		result := make(map[string]interface{})
		for _, item := range v.Items() {
			key := starlarkToGo(item[0]).(string)
			result[key] = starlarkToGo(item[1])
		}
		return result
	default:
		return nil
	}
}
