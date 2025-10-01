package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/go-sourcemap/sourcemap"
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

	fmt.Println("=== VPC別ConfigMapグループ化・マージ処理 ===")
	fmt.Println()

	fmt.Printf("入力ConfigMap数: %d個\n", len(configMaps))
	for i, cm := range configMaps {
		vpcID := cm.Metadata.Labels["vpc-id"]
		fmt.Printf("%d. %s (vpc-id: %s, subnet-id: %s)\n",
			i+1, cm.Metadata.Name, vpcID, cm.Data["subnet-id"])
	}
	fmt.Println()

	// TypeScriptで処理
	mergedConfigMaps, err := processWithTypeScript(configMaps)
	if err != nil {
		fmt.Printf("エラー: %v\n", err)
		return
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

// TypeScriptでConfigMapを処理（VPC別にグループ化してマージ）
func processWithTypeScript(configMaps []ConfigMap) ([]ConfigMap, error) {
	// TypeScriptコード
	tsCode := `
		// Kubernetes ConfigMapの型定義
		interface ConfigMap {
			apiVersion: string;
			kind: string;
			metadata: Metadata;
			data: { [key: string]: string };
		}

		// Metadataの型定義
		interface Metadata {
			name: string;
			namespace?: string;
			labels?: { [key: string]: string };
		}

		// VPC別にConfigMapをグループ化してマージする関数
		function groupByVpcAndMerge(configMaps: ConfigMap[]): ConfigMap[] {
			// VPC IDでグループ化
			const vpcGroups = new Map<string, ConfigMap[]>();

			for (const configMap of configMaps) {
				const vpcId = configMap.metadata.labels?.["vpc-id"];
				
				if (!vpcId) {
					console.log("⚠ vpc-idラベルがありません:", configMap.metadata.name);
					continue;
				}

				if (!vpcGroups.has(vpcId)) {
					vpcGroups.set(vpcId, []);
				}
				vpcGroups.get(vpcId)!.push(configMap);
			}

			// グループごとにマージ
			const mergedConfigMaps: ConfigMap[] = [];

			for (const [vpcId, configMapsInVpc] of vpcGroups) {
				console.log("📦 VPC ID:", vpcId, "- ConfigMap数:", configMapsInVpc.length);

				// subnet-idのみを抽出してマージ
				const mergedData: { [key: string]: string } = {};
				let namespace = "default";

				for (const cm of configMapsInVpc) {
					// namespaceを取得（最初のものを使用）
					if (cm.metadata.namespace) {
						namespace = cm.metadata.namespace;
					}

					// subnet-idキーのみを抽出
					for (const [key, value] of Object.entries(cm.data)) {
						if (key === "subnet-id") {
							// 元のConfigMap名をキー名として使用
							const newKey = cm.metadata.name + "." + key;
							mergedData[newKey] = value;
							console.log("  ✓ 追加:", newKey, "=", value);
						}
					}
				}

				// マージ済みConfigMapを作成
				const mergedConfigMap: ConfigMap = {
					apiVersion: "v1",
					kind: "ConfigMap",
					metadata: {
						name: vpcId,
						namespace: namespace,
						labels: {
							"vpc-id": vpcId,
							"merged": "true"
						}
					},
					data: mergedData
				};

				mergedConfigMaps.push(mergedConfigMap);
			}

			console.log("\n✅ 合計", mergedConfigMaps.length, "個のVPCグループを作成");

			return mergedConfigMaps;
		}

		// メイン処理
		(function() {
			const result = groupByVpcAndMerge(inputConfigMaps);
			return result;
		})();
	`

	// TypeScript→JavaScriptトランスパイル
	jsCode, sourceMapData, err := transpileTypeScriptWithSourceMap(tsCode)
	if err != nil {
		return nil, fmt.Errorf("トランスパイルエラー: %w", err)
	}

	// sourcemapをパース
	smap, err := sourcemap.Parse("", []byte(sourceMapData))
	if err != nil {
		return nil, fmt.Errorf("sourcemapパースエラー: %w", err)
	}

	// gojaで実行
	vm := goja.New()

	// console.logを実装
	console := vm.NewObject()
	console.Set("log", func(args ...interface{}) {
		fmt.Println(args...)
	})
	vm.Set("console", console)

	// ConfigMapをJSON経由でJavaScriptに渡す
	configMapsJSON, _ := json.Marshal(configMaps)
	var configMapsJS interface{}
	json.Unmarshal(configMapsJSON, &configMapsJS)
	vm.Set("inputConfigMaps", configMapsJS)

	// JavaScriptを実行
	result, err := vm.RunString(jsCode)
	if err != nil {
		// エラーをTypeScriptの行番号に変換
		mappedErr := mapErrorToTypeScript(err, smap, "vpc-processor.ts", tsCode)
		return nil, mappedErr
	}

	// 結果をJSONに変換してGoの構造体にデコード
	resultJSON, err := json.Marshal(result.Export())
	if err != nil {
		return nil, fmt.Errorf("結果のJSON変換エラー: %w", err)
	}

	var mergedConfigMaps []ConfigMap
	err = json.Unmarshal(resultJSON, &mergedConfigMaps)
	if err != nil {
		return nil, fmt.Errorf("結果のデコードエラー: %w", err)
	}

	return mergedConfigMaps, nil
}

// TypeScriptをJavaScriptに変換（sourcemap付き）
func transpileTypeScriptWithSourceMap(tsCode string) (jsCode string, sourceMap string, err error) {
	result := api.Transform(tsCode, api.TransformOptions{
		Loader:    api.LoaderTS,
		Sourcemap: api.SourceMapInline,
		Target:    api.ES2020,
	})

	if len(result.Errors) > 0 {
		var errMsgs []string
		for _, err := range result.Errors {
			errMsgs = append(errMsgs, err.Text)
		}
		return "", "", fmt.Errorf("esbuildエラー: %s", strings.Join(errMsgs, "; "))
	}

	jsCode = string(result.Code)

	// インラインsourcemapを抽出
	re := regexp.MustCompile(`//# sourceMappingURL=data:application/json;base64,(.+)`)
	matches := re.FindStringSubmatch(jsCode)
	if len(matches) < 2 {
		return "", "", fmt.Errorf("sourcemapが見つかりません")
	}

	sourceMapBase64 := matches[1]
	sourceMapBytes, err := decodeBase64(sourceMapBase64)
	if err != nil {
		return "", "", fmt.Errorf("base64デコードエラー: %w", err)
	}

	return jsCode, string(sourceMapBytes), nil
}

// エラーをTypeScriptの位置にマッピング
func mapErrorToTypeScript(err error, smap *sourcemap.Consumer, filename, tsCode string) error {
	errStr := err.Error()

	re := regexp.MustCompile(`at.*?:(\d+):(\d+)`)
	matches := re.FindStringSubmatch(errStr)

	if len(matches) < 3 {
		return fmt.Errorf("%s\n(sourcemapでの位置特定不可)", errStr)
	}

	jsLine, _ := strconv.Atoi(matches[1])
	jsCol, _ := strconv.Atoi(matches[2])

	_, _, line, col, ok := smap.Source(jsLine, jsCol)
	if !ok {
		return fmt.Errorf("%s\n(sourcemap変換失敗: JS %d:%d)", errStr, jsLine, jsCol)
	}

	lines := strings.Split(tsCode, "\n")
	var contextLines []string

	start := max(0, line-3)
	end := min(len(lines), line+2)

	for i := start; i < end; i++ {
		lineNum := i + 1
		prefix := "  "
		if lineNum == line {
			prefix = "→ "
		}
		contextLines = append(contextLines, fmt.Sprintf("%s%4d | %s", prefix, lineNum, lines[i]))
		if lineNum == line {
			spaces := strings.Repeat(" ", col+8)
			contextLines = append(contextLines, spaces+"^")
		}
	}

	return fmt.Errorf(`
%s

ファイル: %s
位置: %d行%d列

%s
`, errStr, filename, line, col, strings.Join(contextLines, "\n"))
}

// ヘルパー関数
func decodeBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
