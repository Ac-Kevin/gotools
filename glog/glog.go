package glog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

/*
===================
 log handlers
===================
*/

//RotatingHandler 日志处理结构体
type RotatingHandler struct {
	ID             string //应用ID
	Version        string //应用版本
	ProgramModTime string //程序修改时间
	RunEnvironment string //服务器运行环境  10正式服务器 20测试服务器
	Dir            string //目录
	Filename       string //log文件名
	MaxSize        int64  //文件最大尺寸
	SaveDay        int    //文件保存时间
	screenStatus   bool   //屏幕输出状态
	httpStatus     bool   //post/get到错误日志服务器状态

	CloudLogStatus bool //云端日志启动状态

	logip string //调用log的项目的ip地址

	logfile        *os.File      //日志文件
	errMsgChannel  chan *string  //错误消息通道
	httpMsgChannel chan *logInfo //http消息通道
	HTTPMsgmethod  string        //http日志发送模式
	HTTPMsgURL     string        //http日志接收地址

}

//LogHandler RotatingHandler结构体对应的 全局变量
var LogHandler = &RotatingHandler{
	ID:             "1000",                     //应用ID
	Version:        "0.01",                     //应用版本
	ProgramModTime: GetProgramModTime(),        //获取程序修改时间
	RunEnvironment: "20",                       //服务器运行环境  10正式服务器 20测试服务器
	Dir:            "./log",                    //初始化后，外部再设置目录，不可用。
	Filename:       "err.log",                  //错误文件名称（修改名称的时候才会生效）
	MaxSize:        4 * 1024 * 1024,            //一个文件最大尺寸 默认 4Mb
	SaveDay:        60,                         //保存日志文件60天
	httpStatus:     false,                      //http日志发送开关
	CloudLogStatus: false,                      //云端日志默认不启动
	HTTPMsgURL:     "",                         //默认http日志地址
	HTTPMsgmethod:  "POST",                     //默认http日志方法方法
	errMsgChannel:  make(chan *string, 10000),  //设置 1w 个写缓存的通道
	httpMsgChannel: make(chan *logInfo, 10000), //post比 写文件速度慢，所以缓存通道多一些
}

var msgTotalLen int64 //现有log文件大小

/*
StartLogHandler 外部调用 log日志 初始化

参数：
		appID	应用ID
		appVersion 应用版本
		runEnvironment 运行环境  10正式服务器 20测试服务器
		screenStatus 是否屏幕输出
		httpStatus 是否开启 http发送错误消息到 服务器做记录
返回：
		内存指针（可用于在外部设置，对应参数）
注意：
		如果需要额外参数，必须运行在sbjlog.StartLogHandler之前
例子：
		sbjlog.LogHandler.HTTPMsgURL =  httpURL   //http日志发送地址
		sbjlog.LogHandler.HTTPMsgmethod = method //http日志发送模式
------------------
*/
func StartLogHandler(appID int, appVersion string, runEnvironment int, screenStatus, httpStatus bool) *RotatingHandler {
	//目录不存在则创建
	LogHandler.ID = strconv.Itoa(appID) //应用ID必填
	LogHandler.Dir = "./" + LogHandler.ID + "_log"
	if _, err := os.Stat(LogHandler.Dir); err != nil {
		os.MkdirAll(LogHandler.Dir, 0777) //原来 0711权限 可能会导致其它线程，读取文件夹内内容出错
	}
	LogHandler.Version = appVersion                          //应用版本 必填
	LogHandler.screenStatus = screenStatus                   //是否屏幕输出
	LogHandler.logip = getLocalIP()                          //获得当前服务器ip地址
	LogHandler.RunEnvironment = strconv.Itoa(runEnvironment) //服务器运行环境

	startTimer(updateOleFileName) //启动零点计时器（重命名旧的test.log文件）
	//开启线程 判断目录下，是否有过期的文件有就删除
	go checkFileTime(LogHandler.SaveDay) //(参数：过期时间)只删除日志文件（.log）
	go writeMsgHandle()                  //开启线程 做写入消息处理

	//如果开启 http 发送模式
	if httpStatus == true {
		LogHandler.httpStatus = true
		go StarupLogHTTPParameter() //启动http日志
	}
	Printfer("100", "sbjlog日志线程已启动！")
	return LogHandler //返回内存指针
}

//writeErrMsgHandle：写入本地log文件错误消息处理
func writeMsgHandle() {
	var errMsg *string
	var logBuffer bytes.Buffer
	//打开文件 (如果文件不存在，那么就创建文件)
	LogHandler.logfile, _ = os.OpenFile(LogHandler.Dir+"/"+LogHandler.Filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	//关闭文件
	if LogHandler.logfile != nil {
		LogHandler.logfile.Close()
	}
	msgTotalLen = fileSize(LogHandler.Dir + "/" + LogHandler.Filename) //将原来文件大小赋值给 合计总文件大小
	timer := time.NewTicker(1 * time.Second)                           //默认1秒判断是否需要写入 | 注：频繁的stdout或者stderr输出 会导致supervisor处理变慢
	for {
		select {
		case <-timer.C: //每1秒检查是否有需要写入
			if logBuffer.Len() > 0 {
				logWriteBytes(&logBuffer) //写入文件(字符串)，如果文件不存在就创建文件
				logBuffer.Reset()         //清空buffer
			}
		case errMsg = <-LogHandler.errMsgChannel:
			msgTotalLen = msgTotalLen + int64(len(*errMsg)) //获得新总写入字节数
			logBuffer.WriteString(*errMsg)                  //将要写的数据放入缓存Buffer
			if msgTotalLen > LogHandler.MaxSize {           //如果合计总文件，大于设置的文件大小，就执行
				LogHandler.rename()                  //改名
				logWriteBytes(&logBuffer)            //写入文件(字符串)，如果文件不存在就创建文件
				msgTotalLen = int64(logBuffer.Len()) //重置msgTotalLen为最后写入的字符串大小）
				logBuffer.Reset()                    //清空buffer
			}
		}
	}
}

//logWriteBytes 将字符串写入日志 --测试用WriteBytes
func logWriteBytes(errBytes *bytes.Buffer) {
	//打开文件 (如果文件不存在，那么就创建文件)
	LogHandler.logfile, _ = os.OpenFile(LogHandler.Dir+"/"+LogHandler.Filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if LogHandler.logfile != nil {
		LogHandler.logfile.Write(errBytes.Bytes()) //写入文件(字符串)
		LogHandler.logfile.Close()                 //关闭文件
	}
}

// 删除过期文件
func checkFileTime(saveDay int) {
	var expiretime time.Time
	var err error
	var listfile []os.FileInfo
	var file os.FileInfo
	for {
		//遍历文件夹下所有的文件
		listfile, err = ioutil.ReadDir(LogHandler.Dir)
		if err == nil {
			for _, file = range listfile {
				//筛选.log文件
				if path.Ext(file.Name()) == ".log" {
					expiretime = file.ModTime().AddDate(0, 0, saveDay)
					//过期文件 删除
					if time.Now().After(expiretime) {
						os.Remove(LogHandler.Dir + "/" + file.Name())
					}
				}
			}
		}
		time.Sleep(1 * 24 * time.Hour) //每天清理一次
	}
}

//方法： 改名
func (h *RotatingHandler) rename() {
	if h.logfile != nil { //关闭打开的文件
		h.logfile.Close()
	}
	newpath := fmt.Sprintf("%s/%s_%d.log", h.Dir, time.Now().Format("2006-01-02"), time.Now().UTC().UnixNano()/1000000) //格式化，返回字符串
	if isExist(newpath) {
		os.Remove(newpath) //如果文件存在，那么就删除文件
	}
	filepath := h.Dir + "/" + h.Filename
	os.Rename(filepath, newpath)
	msgTotalLen = 0 //重置 文件大小，为0
}

// 过零点则将前一天的test.log文件重命名，避免不同日期的日志写在一个文件中
func updateOleFileName() {
	fileInfo, err := os.Stat(LogHandler.Dir + "/" + LogHandler.Filename)
	if err == nil || os.IsExist(err) { //err没有错误，或者，err 的错误为 文件已经存在
		modTime := fileInfo.ModTime() //获取文件的修改时间
		//判断是否为旧log
		if modTime.Format("2006-01-02") < time.Now().Format("2006-01-02") {
			if LogHandler.logfile != nil { //判断文件是否已经被打开
				LogHandler.logfile.Close() //被打开，就关闭
			}
			// 不使用LogHandler.suffix，避免程序重启导致误删,日志后面以时间撮结尾
			newpath := fmt.Sprintf("%s/%s_%d.log", LogHandler.Dir, modTime.Format("2006-01-02"), time.Now().UTC().UnixNano()/1000000) //格式化，返回字符串
			if isExist(newpath) {
				os.Remove(newpath) //如果文件存在，那么就删除文件
			}
			filepath := LogHandler.Dir + "/" + LogHandler.Filename
			os.Rename(filepath, newpath) //改名
			msgTotalLen = 0              //重置 文件大小，为0
		}
	}
}

/*errInfo  日志消息结构体*/
type logInfo struct {
	msgType string
	errCode string
	content string
}

//日志消息处理函数
func logDeal(item *logInfo) {

	logString := fmt.Sprintf("%s: V%s %s code[%s] %s\n", item.msgType, LogHandler.Version, time.Now().Format("2006-01-02 15:04:05.000"), item.errCode, item.content)

	//屏幕打印
	if LogHandler.screenStatus {
		fmt.Println(logString)
	}

	//本地记录通道
	LogHandler.errMsgChannel <- &logString
	item.errCode = LogHandler.ID + item.errCode
	//http发送通道
	if LogHandler.httpStatus && !(!LogHandler.CloudLogStatus && item.msgType == "Log") {
		LogHandler.httpMsgChannel <- item
	}
}

//-----------------------------外部调用函数-----------------------------------------

//RbwLog 启动警告日志
func RbwLog(format string, v ...interface{}) {
	content := fmt.Sprintf(format, v...)
	modTime, _ := time.Parse("2006-01-02 15:04:05.000", LogHandler.ProgramModTime)
	reqParams := fmt.Sprintf("type=95&code=999&msgtype=Rbw&tid=%s&version=%s&runEnvironment=%s&programModTime=%d&err=%s",
		LogHandler.ID,
		LogHandler.Version,
		LogHandler.RunEnvironment,
		modTime.UTC().UnixNano()/1000000,
		url.QueryEscape(content))
	fmt.Println(reqParams)
	httpRequestData(LogHandler.HTTPMsgURL+"/rebooterr", reqParams, "POST", 3)

	logString := fmt.Sprintf("Rbw：V%s %s code[999] %s\n", LogHandler.Version, time.Now().Format("2006-01-02 15:04:05.000"), content)
	//屏幕打印
	if LogHandler.screenStatus {
		fmt.Println(&logString)
	}
	LogHandler.errMsgChannel <- &logString //本地记录通道
}

//Printf 函数用于输出日志
func Printf(format string, v ...interface{}) {
	item := &logInfo{
		msgType: "Log",
		errCode: "0" + "000", //默认普通日志：应用ID+"0"+"000" 兼容旧版，默认000
		content: fmt.Sprintf(format, v...),
	}
	logDeal(item)
}

//Printfer 函数用于输出日志-V2
func Printfer(code, format string, v ...interface{}) {
	item := &logInfo{
		msgType: "Log",
		errCode: "0" + code, //普通日志：应用ID+"0"+code
		content: fmt.Sprintf(format, v...),
	}
	logDeal(item)
}

//Debug 函数用于输出错误
func Debug(format string, v ...interface{}) {
	item := &logInfo{
		msgType: "Bug",
		errCode: "1" + "000", //默认错误日志：应用ID+"0"+"000" 兼容旧版，默认000
		content: fmt.Sprintf(format, v...),
	}
	logDeal(item)
}

/*Debuger 打印错误日志-V2
参数说明：code为错误代码 */
func Debuger(code, format string, v ...interface{}) {
	item := &logInfo{
		msgType: "Bug",
		errCode: "1" + code, //错误日志：应用ID+"1"+code
		content: fmt.Sprintf(format, v...),
	}
	logDeal(item)
}

/*ExcLog 打印异常日志*/
func ExcLog(code, format string, v ...interface{}) {
	item := &logInfo{
		msgType: "Exc",
		errCode: "2" + code, //异常日志：应用ID+"2"+code
		content: fmt.Sprintf(format, v...),
	}
	logDeal(item)
}

/*
StarupLogHTTPParameter 用于设置http参数
参数说明：
method http方法（POST OR GET）,默认为POST
url 要post/get的目标服务端地址
httpid 应用ID
*/
func StarupLogHTTPParameter() {
	if LogHandler.HTTPMsgURL != "" && LogHandler.HTTPMsgmethod != "" {
		starupLogHandle() //发送应用启动日志
		//启动post线程（发现 http的请求 go底层其实自己实现了多线程）
		for i := 0; i < 10; i++ {
			go donormalHTTPRequest()
		}
	} else {
		Printfer("1001", "Err ： http日志发送地址 或者 http日志发送模式 为空！")
	}
}

//-----------------------------外部调用函数-----------------------------------------

/*--------------------------
调用单元
----------------------------*/

//读取指定路径的文件尺寸  返回文件大小
func fileSize(file string) int64 {
	//fmt.Println("fileSize", file)
	f, e := os.Stat(file)
	if e != nil {
		fmt.Println(e.Error())
		return 0
	}
	return f.Size()
}

//GetProgramModTime 获取程序修改时间 返回 时间
func GetProgramModTime() string {
	file, _ := exec.LookPath(os.Args[0])
	f, err := os.Stat(file)
	if err != nil {
		return "1970-01-01 00:00:01.000"
	}
	return f.ModTime().Format("2006-01-02 15:04:05.000")
}

//判断文件是否存在  存在返回 true ，不存在返回  false
func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// 零点定时器
func startTimer(f func()) {
	go func() {
		var now, next time.Time
		var t *time.Timer
		for {
			f()
			now = time.Now()
			// 计算下一个零点
			next = now.Add(time.Hour * 24)
			next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
			t = time.NewTimer(next.Sub(now))
			<-t.C
		}
	}()
}

//获取ip地址
func getLocalIP() string {
	var ipAddr string
	addrSlice, err := net.InterfaceAddrs()
	if err != nil {
		ipAddr = "0.0.0.0"
		return ipAddr
	}
	for _, addr := range addrSlice {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipAddr = ipnet.IP.String()
				return ipAddr
			}
		}
	}
	return "0.0.0.0"
}

//普通的post 请求
func donormalHTTPRequest() {
	httplog := &http.Client{
		Timeout: 2 * time.Second, //为了http 请求线程安全设置2秒超时|终止本次请求
	}
	var posterrString string
	var body []byte
	var resp *http.Response
	var err error
	args := ""
	var item *logInfo
	for {
		item = <-LogHandler.httpMsgChannel
		//没有指定，type为消息类型默认为POST
		if LogHandler.HTTPMsgmethod == "GET" {
			args = fmt.Sprintf("type=10&ip=%s&tid=%s&version=%s&runEnvironment=%s&msgtype=%s&err=%s&code=%s",
				LogHandler.logip,
				LogHandler.ID,
				LogHandler.Version,
				LogHandler.RunEnvironment,
				item.msgType,
				item.content,
				item.errCode)
			resp, err = httplog.Get(LogHandler.HTTPMsgURL + "/errlog?" + args)
		} else {
			body, _ = json.Marshal(map[string]string{
				"type":           "10",
				"ip":             LogHandler.logip,
				"tid":            LogHandler.ID,
				"version":        LogHandler.Version,
				"msgtype":        item.msgType,
				"err":            item.content,
				"runEnvironment": LogHandler.RunEnvironment,
				"code":           item.errCode})
			resp, err = httplog.Post(LogHandler.HTTPMsgURL+"/errlog", "application/json", strings.NewReader(string(body)))
		}

		//如果HTTPRequest的请求有错误，就把错误，写入错误放入 错误通道等待写入错误文件
		if err != nil {
			posterrString = fmt.Sprintf("Sbjlog Debug Time:%s Http Request err :%s \n , Post Data:%s", time.Now().Format("2006-01-02 15:04:05.000"), err, string(body))
			fmt.Println(posterrString)
			LogHandler.errMsgChannel <- &posterrString //将错误消息，放入错误消息通道，用于写入错误日志到文件
		} else {
			resp.Body.Close() //报错情况关闭会导致内存指针错误，简言之，接收端关了，发送端就挂了
		}
	}
}

//发送http应用启动日志
func starupLogHandle() {
	httplog := &http.Client{
		Timeout: 10 * time.Second,
	}
	var resp *http.Response
	var err error
	if LogHandler.HTTPMsgmethod == "GET" {
		u, _ := url.Parse(LogHandler.HTTPMsgURL + "/startuplog")
		q := u.Query()
		q.Set("type", "90")
		q.Set("tid", LogHandler.ID)
		q.Set("version", LogHandler.Version)
		q.Set("programModTime", LogHandler.ProgramModTime)
		q.Set("runEnvironment", LogHandler.RunEnvironment)
		u.RawQuery = q.Encode()
		resp, err = httplog.Get(u.String())
	} else {
		body, _ := json.Marshal(map[string]string{
			"type":           "90",
			"tid":            LogHandler.ID,      //tid 程序id
			"version":        LogHandler.Version, //version 程序版本
			"programModTime": LogHandler.ProgramModTime,
			"runEnvironment": LogHandler.RunEnvironment}) //ProgramModTime 程序的最后修改时间
		bodystr := string(body)
		resp, err = httplog.Post(LogHandler.HTTPMsgURL+"/startuplog", "application/json", strings.NewReader(bodystr))
	}
	var posterrString string
	if err != nil {
		posterrString = fmt.Sprintf("http发送应用启动日志失败:%s \n", err)
		fmt.Println(posterrString)
		posterrString = fmt.Sprintf("%s：V%s %s code[1%s] %s\n", "Bug", LogHandler.Version, time.Now().Format("2006-01-02 15:04:05.000"), "000", posterrString)
		LogHandler.errMsgChannel <- &posterrString
	} else {
		posterrString = fmt.Sprintf("%s：V%s %s code[1%s] %s\n", "Log", LogHandler.Version, time.Now().Format("2006-01-02 15:04:05.000"), "000", "发送启动日志成功")
		LogHandler.errMsgChannel <- &posterrString
		resp.Body.Close()
	}
}

/*
httpRequestData 通用请求页面数据方法 POST or GET
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
func httpRequestData(url, parameter, method string, timeout int) (int, string, error) {
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
