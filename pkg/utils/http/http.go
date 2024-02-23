package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type IPInfo struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`     // 国家
	CountryCode string  `json:"countryCode"` // 国家代码
	Region      string  `json:"region"`      // 省份
	RegionName  string  `json:"regionName"`  // 省份名称
	City        string  `json:"city"`        // 城市
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"` // 时区
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
	QueryIp     string  `json:"query"`
}

// OnlineIpInfo 通过ip-api.com接口查询IP信息
// 返回：IP地址的信息（格式：字符串的json）
func OnlineIpInfo(ip string) *IPInfo {
	url := "http://ip-api.com/json/" + ip + "?lang=zh-CN"
	resp, err := http.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	var result IPInfo
	if err := json.Unmarshal(out, &result); err != nil {
		return nil
	}
	return &result
}

func GetMyPublicIP() string {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	return string(content)
}
