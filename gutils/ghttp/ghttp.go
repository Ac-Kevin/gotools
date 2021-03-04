package ghttp

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

/*
HTTPRequestData 通用请求页面数据方法 POST or GET
传入参数说明：
url 访问地址,
parameter 访问参数,
method 请求方式 POST , GET
timeout http请求超时秒数（s）

返回参数说明：
int http响应码
string http响应消息内容（string）
error 错误返回
*/
func HTTPRequestData(url, parameter, method string, timeout int) (int, string, error) {
	httpweixin := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	var resp *http.Response
	var err error

	if method == "POST" {
		resp, err = httpweixin.Post(url, "application/x-www-form-urlencoded", strings.NewReader(parameter))
	} else {
		resp, err = httpweixin.Get(url + "?" + parameter)
	}
	if err != nil {
		return 0, "", err
	}
	//读取返回信息
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return 0, "", err
	}
	return resp.StatusCode, string(body), nil
}

/*
HTTPRemoteIP 获取客户端IP | 内部反向代理程序会将真实客户端IP写入 Remote_addr,方法先读 Remote_addr,没有的话再读取HTTP自己携带的客户端IP RemoteAddr
输入参数 http.Request HTTP请求
输出 客户端ip
*/
func HTTPRemoteIP(req *http.Request) string {

	//2019-08-15 tkw edit 修复因其他反向代理程序导致获取的IP不对的问题
	//获取 Remote_addr(内部golang程序) 获取 X-Forwarded-For 内反向代理路径
	//对比  Remote_addr 和 X-Forwarded-For ， 如果  X-Forwarded-For 不为空，则取 X-Forwarded-For 第一个IP
	//如果 X-Forwarded-For 为空 则取 Remote_addr
	//如果 X-Forwarded-For，Remote_addr 都 为空 则取HTTP协议自带 Remote_Addr
	xForwardedFor := req.Header.Get("X-Forwarded-For")
	goRemoteAddr := req.Header.Get("Remote_addr")
	remoteAddr := ""        //返回的IP (不带端口号)
	if goRemoteAddr != "" { //第一读取IP golang后台自己写进去的用户实际IP地址
		remoteAddr = goRemoteAddr
	} else if xForwardedFor != "" { //读取代理IP地址
		remoteAddr = strings.Split(xForwardedFor, ",")[0]
	} else { //读取原生IP地址
		remoteAddr = req.RemoteAddr
	}
	// 用冒号分割 取[0]作为IP地址
	remoteAddr = strings.Split(remoteAddr, ":")[0]
	// 空则默认为本机IP地址
	if remoteAddr == "" {
		remoteAddr = "127.0.0.1"
	}
	//返回IP地址
	return remoteAddr
}

/*
GetformatURLData 解析http get请求参数 例：data:=formatURLData(r.Form)
*/
func GetformatURLData(data map[string][]string) map[string]string {
	m := make(map[string]string)
	for k := range data {
		if len(data[k]) > 0 {
			m[k] = data[k][0]
		}
	}
	return m
}

//============================================= 以下为整理后 http 辅助函数 =========================================================

/*SendPostForm 发送postform请求
传入参数：url 请求主机地址
				 params 参数
				 timeout 超时设置 单位 s
*/
func SendPostForm(url, params string, timeout int, headers ...map[string]string) (string, error) {
	return HTTPBaseRequest(url, params, "application/x-www-form-urlencoded", "POST", timeout, headers...)
}

/*SendPostJSON 发送postform请求
传入参数：url 请求主机地址
				 params 参数
				 timeout 超时设置 单位 s
*/
func SendPostJSON(url, params string, timeout int, headers ...map[string]string) (string, error) {
	return HTTPBaseRequest(url, params, "application/json;charset=UTF-8", "POST", timeout, headers...)
}

/*SendGET 发送GET http请求 */
func SendGET(url, params string, timeout int, headers ...map[string]string) (string, error) {
	return HTTPBaseRequest(url, params, "", "GET", timeout, headers...)
}

//CertConfig 微信支付，证书配置文件，微信文档地址：https://pay.weixin.qq.com/wiki/doc/api/jsapi_sl.php?chapter=9_4
type CertConfig struct {
	WechatPayCert string // 证书路径
	WechatPayKey  string // 证书秘钥路径
	RootCa        string // 根证书路径，根证书文件是需要自己另外下载的，下载地址：https://pay.weixin.qq.com/wiki/doc/api/micropay.php?chapter=23_4
}

//HTTPPostWithCert 提交post请求带证书
//参考博文：https://blog.csdn.net/mario08/article/details/86243266
func HTTPPostWithCert(url string, contentType string, body io.Reader, certConfig *CertConfig) (*http.Response, error) {
	var wechatPayCert = "./cert/apiclient_cert.pem"
	var wechatPayKey = "./cert/apiclient_key.pem"
	var rootCa = "./cert/rootca.pem"
	if certConfig != nil {
		if certConfig.WechatPayCert != "" {
			wechatPayCert = certConfig.WechatPayCert
		}
		if certConfig.WechatPayKey != "" {
			wechatPayKey = certConfig.WechatPayKey
		}
		if certConfig.RootCa != "" {
			rootCa = certConfig.RootCa
		}
	}
	var tr *http.Transport
	// 微信提供的API证书,证书和证书密钥 .pem格式
	certs, err := tls.LoadX509KeyPair(wechatPayCert, wechatPayKey)
	if err != nil {
		//sbjlog.Debug("HTTPPostWithCert %v \n", err)
	} else {
		// 微信支付HTTPS服务器证书的根证书  .pem格式
		rootCa, err := ioutil.ReadFile(rootCa)
		if err != nil {
			//sbjlog.Debug("HTTPPostWithCert ioutil.ReadFile %v \n", err)
		} else {
			pool := x509.NewCertPool()
			pool.AppendCertsFromPEM(rootCa)

			tr = &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:      pool,
					Certificates: []tls.Certificate{certs},
				},
			}
		}

	}
	client := &http.Client{Transport: tr}
	return client.Post(url, contentType, body)

}

/*HTTPBaseRequest 发送http请求
传入参数：url 请求主机地址
				 params 参数
				 contentType 参数格式
				 method 请求方法 "POST" OR "GET" 注意要大写
				 timeout 超时设置 单位 s
				 headers map[string]string 非必填参数 | 可用于添加实际用户IP地址 例子：map[string]string{"Remote_addr": "用户IP地址"}
*/
func HTTPBaseRequest(url, params, contentType, method string, timeout int, headers ...map[string]string) (string, error) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	var req *http.Request
	var err error
	switch method {
	case "POST":
		req, err = http.NewRequest("POST", url, strings.NewReader(params))
		if contentType == "" {
			contentType = "application/x-www-form-urlencoded"
		}
		req.Header.Set("Content-Type", contentType)
	case "GET":
		req, err = http.NewRequest("GET", url+"?"+params, nil)
	default:
		err = fmt.Errorf("不可识别method：'%s'", method)
	}
	if err != nil {
		return "", fmt.Errorf("HTTP-Request-Err :%s", err)
	}
	//设置heads
	if len(headers) > 0 {
		for key, value := range headers[0] {
			req.Header.Add(key, value)
		}
	}
	//http请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP-Request-Err :%s", err)
	}
	//读取返回信息
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}
	//添加404错误检查
	bodyString := string(body)
	if len(bodyString) > 17 && bodyString[:18] == "404 page not found" {
		return bodyString, errors.New("404 page not found")
	}
	return bodyString, nil
}

/*
HTTPRequestDataV2 通用请求页面数据方法 POST or GET
传入参数：url 请求主机地址
				 params 参数
				 contentType 参数格式
				 method 请求方法 "POST" OR "GET" 注意要大写
				 timeout 超时设置 单位 s

返回参数说明：
int http响应码
string http响应消息内容（string）
error 错误返回
*/
func HTTPRequestDataV2(url, params, contentType, method string, timeout int) (int, string, error) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	var resp *http.Response
	var err error
	if method == "POST" {
		resp, err = client.Post(url, contentType, strings.NewReader(params))
	} else if method == "GET" {
		resp, err = client.Get(url + "?" + params)
	} else {
		return 0, "", fmt.Errorf("不可识别method：'%s'", method)
	}
	if err != nil {
		return 0, "", fmt.Errorf("HTTP-Request-Err :%s", err)
	}
	//读取返回信息
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return 0, "", err
	}
	return resp.StatusCode, string(body), nil
}

/*HTTPBaseRequestWithHeads 发送http请求
传入参数：	 url 请求主机地址
			params 参数
			contentType 参数格式
			method 请求方法 "POST" OR "GET" 注意要大写
			timeout 超时设置 单位 s
			heads 报文头
*/
func HTTPBaseRequestWithHeads(url, params, contentType, method string, timeout int, heads map[string]string) (string, error) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	var req *http.Request
	var err error
	switch method {
	case "POST":
		req, err = http.NewRequest("POST", url, strings.NewReader(params))
		if contentType == "" {
			contentType = "application/x-www-form-urlencoded"
		}
		req.Header.Set("Content-Type", contentType)
	case "GET":
		req, err = http.NewRequest("GET", url+"?"+params, nil)
	default:
		err = fmt.Errorf("不可识别method：'%s'", method)
	}
	if err != nil {
		return "", fmt.Errorf("HTTP-Request-Err :%s", err)
	}
	//设置heads
	if heads != nil {
		for key, value := range heads {
			req.Header.Add(key, value)
		}
	}
	//fmt.Println(req.Header)
	//http请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP-Request-Err :%s", err)
	}
	//读取返回信息
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}
	return string(body), nil
}

//HTTPRequestWithToken 发送带token的http请求
func HTTPRequestWithToken(url, params, method string, timeout int, tokens map[string]string) (string, error) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	var req *http.Request
	var err error
	switch method {
	case "POST":
		req, err = http.NewRequest("POST", url, strings.NewReader(params))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	case "GET":
		req, err = http.NewRequest("GET", url+"?"+params, nil)
	default:
		err = fmt.Errorf("不可识别method：'%s'", method)
	}
	if err != nil {
		return "", fmt.Errorf("HTTP-Request-Err :%s", err)
	}

	//设置cookies
	for key, value := range tokens {
		req.AddCookie(&http.Cookie{Name: key, Value: value, HttpOnly: true})
		req.Header.Add(key, value)
	}

	//http请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP-Request-Err :%s", err)
	}

	//读取返回信息
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}
	return string(body), nil
}
