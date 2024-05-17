package encryption_test

//func TestSwag(t *testing.T) {
//	// 读取第一个OpenAPI文件
//	file1, err := ioutil.ReadFile("openapi1.yaml")
//	if err != nil {
//		fmt.Println("Error reading file1:", err)
//		os.Exit(1)
//	}
//
//	// 读取第二个OpenAPI文件
//	file2, err := ioutil.ReadFile("openapi2.yaml")
//	if err != nil {
//		fmt.Println("Error reading file2:", err)
//		os.Exit(1)
//	}
//
//	// 解析第一个OpenAPI文件
//	swagger1, err := openapi3.NewLoader().LoadFromData(file1)
//	if err != nil {
//		fmt.Println("Error parsing file1:", err)
//		os.Exit(1)
//	}
//
//	// 解析第二个OpenAPI文件
//	swagger2, err := openapi3.NewLoader().LoadFromData(file2)
//	if err != nil {
//		fmt.Println("Error parsing file2:", err)
//		os.Exit(1)
//	}
//
//	// 合并两个OpenAPI文件
//	swagger1.Merge(swagger2)
//
//	// 将合并后的OpenAPI写入文件
//	mergedData, err := swagger1.MarshalJSON()
//	if err != nil {
//		fmt.Println("Error marshaling merged OpenAPI:", err)
//		os.Exit(1)
//	}
//
//	err = ioutil.WriteFile("merged_openapi.yaml", mergedData, 0644)
//	if err != nil {
//		fmt.Println("Error writing merged OpenAPI to file:", err)
//		os.Exit(1)
//	}
//
//	fmt.Println("Merged OpenAPI written to merged_openapi.yaml")
//}
