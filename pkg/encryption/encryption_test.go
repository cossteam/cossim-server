package encryption_test

//func getDefaultGateway() (string, error) {
//	routes, err := net.InterfaceAddrs()
//	if err != nil {
//		return "", err
//	}
//	fmt.Println("网卡列表:", routes)
//
//	for _, route := range routes {
//		if ipNet, ok := route.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
//			if ipNet.IP.String() == "0.0.0.0" {
//				fmt.Println("默认路由:", ipNet.IP.String())
//				return ipNet.IP.String(), nil
//			}
//		}
//	}
//
//	return "", fmt.Errorf("找不到默认路由")
//}
//
//func getInterfaceIP(defaultGateway string) (string, error) {
//	interfaces, err := net.Interfaces()
//	if err != nil {
//		return "", err
//	}
//	fmt.Println("网卡列表:", interfaces)
//	for _, iface := range interfaces {
//		addrs, err := iface.Addrs()
//		if err != nil {
//			return "", err
//		}
//
//		for _, addr := range addrs {
//			ipNet, ok := addr.(*net.IPNet)
//			if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
//				if strings.HasPrefix(ipNet.IP.String(), defaultGateway) {
//					return ipNet.IP.String(), nil
//				}
//			}
//		}
//	}
//
//	return "", fmt.Errorf("找不到默认路由网卡的IP地址")
//}
//
//func TestIpadd(t *testing.T) {
//	defaultGateway, err := getDefaultGateway()
//	if err != nil {
//		fmt.Println("获取默认路由错误:", err)
//		return
//	}
//
//	interfaceIP, err := getInterfaceIP(defaultGateway)
//	if err != nil {
//		fmt.Println("获取默认路由网卡的IP地址错误:", err)
//		return
//	}
//
//	fmt.Println("默认路由网卡的IP地址:", interfaceIP)
//}
//
//
//
//func TestDns(t *testing.T) {
//	ip, err := GetOutBoundIP()
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Println(ip)
//}
