package gfile

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"ackevin.com/jsonq"
)

//Configdatajsonq config文件内容json格式
var Configdatajsonq *jsonq.JSONQuery

/*
Readconfigfile 读取配置文件内容，返回json格式
*/
func Readconfigfile(filename string) (*jsonq.JSONQuery, error) {
	// 读取配置文件 config.ini 内容
	configfile, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open config file error:%s", err)
	}
	defer configfile.Close() //如果打开文件异常，就关闭文件
	configdata, err := ioutil.ReadAll(configfile)
	if err != nil {
		return nil, fmt.Errorf("read config file error:%s", err)
	}
	if checkUTF8Format(configdata) {
		configdata = configdata[3:]
	}
	configdatastring := string(configdata)
	// 去除配置文件的注释
	configdatastring = dispelAnnotation(configdatastring)
	fmt.Println(configdatastring)
	// //格式化 配置文件 config.ini 内的内容格式化 json
	//将json字符串转为json结构体实例
	mapData := map[string]interface{}{}                                         //初始化一个map[string]interface{}
	err = json.NewDecoder(strings.NewReader(configdatastring)).Decode(&mapData) //将json字符串解析到map
	Configdatajsonq = jsonq.NewQuery(mapData)                                   //创建一个json查询
	//返回
	return Configdatajsonq, err
}

/*
dispelAnnotation 将文本中的 // 注释 以及 / * ... * / // 进行去除
2020-09-09 添加文本中的 ; 注释
*/
func dispelAnnotation(data string) string {
	//去除代码注释部分
	pat := "(\\s+;.*\n*)|(\\s+/{2,}.*\n*)|(/\\*[\\s\\S]*?\\*/)" //正则
	re, _ := regexp.Compile(pat)
	return re.ReplaceAllString(data, "\n")
}

/*
checkUTF8Format 检查是否为UTF-8字符数据
*/
func checkUTF8Format(date []byte) bool {
	if len(date) < 3 {
		return false
	}
	//检查是否存在utf-8 的 bom头
	if date[0] == 239 && date[1] == 187 && date[2] == 191 {
		return true
	}
	return false
}