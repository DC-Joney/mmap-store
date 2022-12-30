package statics

import (
	"net"
	"net/http"
	"time"
)

var httpClient *http.Client

// createHTTPClient 创建http client 连接池
func createHTTPClient() *http.Client {

	client := &http.Client{
		Transport: &http.Transport{
			//DisableKeepAlives:false,// 是否开启http keepalive功能，也即是否重用连接，默认开启(false)
			Proxy: http.ProxyFromEnvironment,
			// 通过设置tls.Config的InsecureSkipVerify为true，client将不再对服务端的证书进行校验
			//TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext, // 该函数用于创建http（非https）连接，通常需要关注Timeout和KeepAlive参数
			MaxIdleConns:        100, // 连接池对所有host的最大链接数量，host也即dest-ip，默认为无穷大（0）
			MaxIdleConnsPerHost: 100, // 连接池对每个host的最大链接数量
			// 空闲timeout设置，也即socket在该时间内没有交互则自动关闭连接
			// （注意：该timeout起点是从每次空闲开始计时，若有交互则重置为0）,
			// 该参数通常设置为分钟级别，例如：90秒
			IdleConnTimeout:     time.Duration(5) * time.Second,
		},
		Timeout: 30 * time.Second,
	}
	return client
}

func init()  {
	//创建Http连接池
	httpClient = createHTTPClient()

}
