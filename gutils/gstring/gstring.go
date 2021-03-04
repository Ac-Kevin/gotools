package gstring


import (
	"crypto/rand"
	"strings"
)

//RandomString 生成随机字符串
//参数：
//strSize 需要生产多少位
//randType 返回字符串
func RandomString(strSize int, randType string) string {
	//默认为数字加字母
	dictionary := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	if randType == "alpha" { //字母
		dictionary = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	} else if randType == "number" { //数字
		dictionary = "0123456789"
	}
	var bytes = make([]byte, strSize)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bytes)
}

/*AESStringSend 发送方aes字符串处理 */
func AESStringSend(str string) string {
	str = strings.Replace(str, "+", `%2B`, -1)
	str = strings.Replace(str, "=", `%3D`, -1)
	str = strings.Replace(str, "/", `%2F`, -1)
	return str
}

/*AESStringRecv 接收方aes字符串处理 */
func AESStringRecv(str string) string {
	//兼容 IOS 编码符 大写
	str = strings.Replace(str, `%2B`, "+", -1)
	str = strings.Replace(str, `%3D`, "=", -1)
	str = strings.Replace(str, `%2F`, "/", -1)
	return str
}

//GetInitPowerString 获取初始权限字符串 q传 "0" 或者 "1"
func GetInitPowerString(q string, count int) string {
	power := ""
	for i := 0; i < count; i++ {
		power += string(q)
	}
	return power
}

//SubString 截取字符串 按字节长度截取 中文字符串也可使用
func SubString(str string, byteSize int) string {
	if len(str) <= byteSize {
		return str
	}
	substr := ""
	for _, c := range []rune(str) {
		if byteSize = byteSize - len(string(c)); byteSize < 0 {
			return substr
		}
		substr += string(c)
	}
	return str
}
