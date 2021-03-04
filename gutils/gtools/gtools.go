package gtools

import (
	"ackevin.com/gutils/ghttp"
	"ackevin.com/gutils/gjson"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"regexp"
	"strconv"
	"time"
)

//AbsInt64 取Int64绝对值
func AbsInt64(i int64) int64 {
	if i < 0 {
		return int64(-i)
	}
	return i
}

/*VerifyMobile  验证手机号*/
func VerifyMobile(mobile string) bool {
	if num, _ := strconv.ParseInt(mobile, 10, 64); len(mobile) != 11 || num <= 10000000000 || num >= 20000000000 {
		return false
	}
	return true
}

//EncodeMoblie 手机号编码
func EncodeMoblie(mobile int64) string {
	str := fmt.Sprintf("%d", mobile)
	if len(str) < 11 {
		return ""
	}
	return str[:3] + "*****" + str[len(str)-3:]
}

/*VerifyEmail  验证邮箱*/
func VerifyEmail(email string) bool {
	pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

//VerifyDate 日期验证
func VerifyDate(date string) bool {
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}

// VerifyDateTime 验证日期时间
func VerifyDateTime(str string) bool {
	_, err := time.Parse("2006-01-02 15:04:05", str)
	return err == nil
}

/*TurnONThread 开启线程
参数: f 线程方法 ， threadNum 线程数
*/
func TurnONThread(threadMethod func(), threadNum int) {
	for i := 0; i < threadNum; i++ {
		go threadMethod()
	}
}

/*GetTimeStamp 获取时间戳
传入参数：url 访问地址 | params url参数 | tpInterval tp时间差地址
返回参数：
*/
func GetTimeStamp(url, params string, tpInterval *int64) (int64, error) {
	currTpInterval := *tpInterval
	if currTpInterval == 0 {
		tBegin := time.Now().UnixNano()
		respBody, err := ghttp.SendGET(url, params, 5)
		if err != nil || respBody == "" {
			return 0, errors.New("请求时间戳失败")
		}
		//计算单程请求耗费的时间 假设请求和响应时间比例为 1:1 ，（当前时间-开始请求时间）/ 2 (单位 ns )/ 1000000( ns 转 ms)
		reqInterval := (time.Now().UnixNano() - tBegin) / 2 / 1000000
		//解析时间戳 | 例子：{"type":"10","status":"True","errcode":"","data":"1565595752538","msg":"","return":""}
		result := gjson.JSONToMapString(respBody) // json字符串 转 map[string]string
		tp := int64(0)
		if tp, _ = strconv.ParseInt(result["data"], 10, 64); tp <= 0 { //新版api 读取data字段
			if tp, _ = strconv.ParseInt(result["msg"], 10, 64); tp <= 0 { //兼容旧版api 读取msg字段
				if tp = decodeMQTTAPITimestamp(respBody); tp <= 0 { // 2020.11.19 zmm 添加解析mqtt返回的时间戳
					return 0, errors.New("解析时间戳失败")
				}
			}
		}
		//计算两台服务器时间差 = 服务器时间（时间戳 去掉 单程请求耗费时间） -  开始请求时间/1000000(ns 转 ms)
		if currTpInterval = (tp - reqInterval) - tBegin/1000000; currTpInterval == 0 {
			currTpInterval = 1 //避免 时间差为 0 ms 的临界情况 ，这里修改时间差为 1 ms
		}
		*tpInterval = currTpInterval //更新tp时间差
	}
	return time.Now().UTC().UnixNano()/1000000 + currTpInterval, nil
}

// 解析mqtt api 返回的时间戳
func decodeMQTTAPITimestamp(resp string) int64 {
	var data struct {
		Type   string `json:"type"`
		Status string `json:"status"`
		Msg    string `json:"msg"`
		Data   struct {
			Timestamp string `json:"timestamp"`
		} `json:"data"`
	}
	json.Unmarshal([]byte(resp), &data)
	tp, _ := strconv.ParseInt(data.Data.Timestamp, 10, 64)
	return tp
}

//GetCurrentMillisecond 获取当前时间戳 毫秒
func GetCurrentMillisecond() string {
	return fmt.Sprintf("%d", time.Now().UTC().UnixNano()/1000000)
}

//GetCurrentNanosecond 获取当前时间戳 纳秒
func GetCurrentNanosecond() string {
	return fmt.Sprintf("%d", time.Now().UTC().UnixNano())
}

//DeclareStruct 声明结构体
func DeclareStruct(obj interface{}) {
	obj = nil
}

//GetLocalIP 获取本机IP地址
func GetLocalIP() string {
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

//GetParamsKey 根据value获取参数map的key | 适用于value唯一的map数据
func GetParamsKey(mData *map[int]string, value string) int {
	for k, v := range *mData {
		if v == value {
			return k
		}
	}
	return 0
}

//SysParams 系统参数结构体
type SysParams struct {
	Key   int    `json:"Key"`
	Value string `json:"Value"`
}

//GetParamsArray 获取系统参数数组 | map转数组形式
func GetParamsArray(mData *map[int]string) *[]SysParams {
	data := []SysParams{}
	for key, value := range *mData {
		data = append(data, SysParams{Key: key, Value: value})
	}
	return &data
}

/*GetEAN13CheckCode 获取检验位*/
//参数 12位数字的字符串
//返回值 -1 = 生成失败（包括提供的字符无效） >=0 成功
func GetEAN13CheckCode(snStr string) int {
	if sn, _ := strconv.ParseInt(snStr, 10, 64); len(snStr) != 12 || sn <= 100000000000 {
		return -1
	}
	//EAN13 编码校验 从右边开始排序
	numMap := map[rune]int{'0': 0, '1': 1, '2': 2, '3': 3, '4': 4, '5': 5, '6': 6, '7': 7, '8': 8, '9': 9}
	arr12 := []rune(snStr)
	//奇数位和
	oddNum := numMap[arr12[0]] + numMap[arr12[2]] + numMap[arr12[4]] + numMap[arr12[6]] + numMap[arr12[8]] + numMap[arr12[10]]
	//偶数位和
	evenNum := (numMap[arr12[1]] + numMap[arr12[3]] + numMap[arr12[5]] + numMap[arr12[7]] + numMap[arr12[9]] + numMap[arr12[11]]) * 3
	return (10 - (oddNum+evenNum)%10) % 10
}

//ChangePrecisionFloat64 浮点数精度修改
func ChangePrecisionFloat64(format string, num float64) float64 {
	//2020-06-30 因golang存在逢5不进1的情况 如 70.2250 保留两位小数后是 70.22
	//与 MSSQL 不相符，MSSQL 是逢5必进1，不会检查前前位是否奇数偶数
	//根据 精度位后不为0则逢5进1的特性，在修改精度时，默认添加  0.00000000000001
	num += 0.00000000000001
	numStr := fmt.Sprintf(format, num)
	num, _ = strconv.ParseFloat(numStr, 64)
	return num
}

//GetMoney 字符串转金额
func GetMoney(money string) float64 {
	f, _ := strconv.ParseFloat(money, 64)
	return f
}

//
/*GetGpsDistance 获取GPS两个坐标点的距离
传入参数：
				A点(lngA-经度, latA-纬度)
				B点(lngB-经度, latB-纬度)
返回参数: 距离(单位：米)
*/
func GetGpsDistance(lngA, latA, lngB, latB float64) float64 {
	//地球平均半径(单位m) | 地球赤道半径6378.137千米,极半径6356.752千米,平均半径约6371千米
	earthRadius := float64(6371000)
	//计算弧度
	rad := math.Pi / 180.0
	latARadian := latA * rad
	lngARadian := lngA * rad
	latBRadian := latB * rad
	lngBRadian := lngB * rad
	if latARadian < 0 {
		latARadian = math.Pi/2 + math.Abs(latARadian)
	}
	if latARadian > 0 {
		latARadian = math.Pi/2 - math.Abs(latARadian)
	}
	if lngARadian < 0 {
		lngARadian = math.Pi*2 - math.Abs(lngARadian)
	}
	if latBRadian < 0 {
		latBRadian = math.Pi/2 + math.Abs(latBRadian)
	}
	if latBRadian > 0 {
		latBRadian = math.Pi/2 - math.Abs(latBRadian)
	}
	if lngBRadian < 0 {
		lngBRadian = math.Pi*2 - math.Abs(lngBRadian)
	}
	//计算A点坐标
	x1 := earthRadius * math.Cos(lngARadian) * math.Sin(latARadian)
	y1 := earthRadius * math.Sin(lngARadian) * math.Sin(latARadian)
	z1 := earthRadius * math.Cos(latARadian)
	//计算B点坐标
	x2 := earthRadius * math.Cos(lngBRadian) * math.Sin(latBRadian)
	y2 := earthRadius * math.Sin(lngBRadian) * math.Sin(latBRadian)
	z2 := earthRadius * math.Cos(latBRadian)
	//按公式计算AB两点弧度距离
	d := math.Sqrt((x1-x2)*(x1-x2) + (y1-y2)*(y1-y2) + (z1-z2)*(z1-z2))
	theta := math.Acos((earthRadius*earthRadius + earthRadius*earthRadius - d*d) / (2 * earthRadius * earthRadius))
	distance := theta * earthRadius
	//返回距离
	return distance
}

//DecodeMsSQLTimeToUnix 解析数据库时间为时间戳（秒），返回int64
func DecodeMsSQLTimeToUnix(sqlTime time.Time) int64 {
	loc, _ := time.LoadLocation("Local") //获取当前时区
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", sqlTime.Format("2006-01-02 15:04:05"), loc)
	return t.Unix()
}

//DecodeMsSQLTimeToUnixString 解析数据库时间为时间戳（秒），返回字符串
func DecodeMsSQLTimeToUnixString(sqlTime time.Time) string {
	return fmt.Sprintf("%d", DecodeMsSQLTimeToUnix(sqlTime))
}