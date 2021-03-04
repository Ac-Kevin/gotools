package gjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"ackevin.com/jsonq"
	"strings"
)
 
/*
JSONToMap Json格式化为 jsonqMAP
*/
func JSONToMap(bodystring string) *jsonq.JSONQuery {
	datastring := map[string]interface{}{}
	dec := json.NewDecoder(strings.NewReader(bodystring)) //configdatastring 是json的数据字符串
	dec.Decode(&datastring)
	jqdata := jsonq.NewQuery(datastring)
	return jqdata
}

/*MapToJSON map转json */
func MapToJSON(v interface{}) string {
	// 存在html标签转义问题
	// str, _ := json.Marshal(v)
	// return string(str)
	// 官方解释 https://stackoverflow.com/questions/28595664/how-to-stop-json-marshal-from-escaping-and
	// 新增 | json.Marshal 默认 escapeHtml 为true,会转义 <、>、&
	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false) //设为false，不转义
	jsonEncoder.Encode(v)
	// 去掉尾部换行符返回
	return strings.TrimRight(bf.String(), string(rune(10)))

}

/*
JSONToMapString Json格式化为 map[string]string
*/
func JSONToMapString(jsonStr string) map[string]string {
	vdata := make(map[string]interface{}, 10)
	json.Unmarshal([]byte(jsonStr), &vdata)
	data := make(map[string]string, len(vdata))
	for k := range vdata {
		data[k] = fmt.Sprintf("%v", vdata[k])
	}
	return data
}

/*FormatJSONString 按照json协议格式化字符串 已在调用微信中间件 启用*/
func FormatJSONString(s string) string {
	var buffer bytes.Buffer
	rs := []rune(s)
	for i := 0; i < len(rs); i++ {
		switch rs[i] {
		case '"':
			buffer.WriteString("\\\"")
			break
		case '\\':
			buffer.WriteString("\\\\")
			break
		case '/':
			buffer.WriteString("\\/")
			break
		case '\b':
			buffer.WriteString("\\b")
			break
		case '\f':
			buffer.WriteString("\\f")
			break
		case '\n':
			buffer.WriteString("\\n")
			break
		case '\r':
			buffer.WriteString("\\r")
			break
		case '\t':
			buffer.WriteString("\\t")
			break
		default:
			buffer.WriteRune(rs[i])
		}
	}
	return buffer.String()
}

/*JSONToMapInterfaceValue json字符串转 map[string]interface{} 结构 */
func JSONToMapInterfaceValue(s string) (data map[string]interface{}, err error) {
	err = json.Unmarshal([]byte(s), &data)
	return
}

/*JSONToMapInterface json字符串转 map[interface{}]interface{} 结构 */
func JSONToMapInterface(s string) (data map[interface{}]interface{}, err error) {
	data = make(map[interface{}]interface{})
	mapInterfaceValue, err := JSONToMapInterfaceValue(s)
	if err != nil {
		return nil, err
	}
	for k, v := range mapInterfaceValue {
		data[k] = v
	}
	return
}
