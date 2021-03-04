package gwxofficialaccount

import (
	"ackevin.com/gdb/gdbwx"
	"ackevin.com/glog"
	"ackevin.com/gutils/ghttp"
	"ackevin.com/gutils/gjson"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"
)

/*GetWXAccessToken 获取微信token
功能：获取微信token
参数：
			Appid 公众号唯一ID  "wx3166b12e65274c9b"
			Appsecret 公众号的appsecret	"add0ddcbae5e7ae064a829cfd4c9338a"
参考：https://mp.weixin.qq.com/wiki?t=resource/res_main&id=mp1421140183
返回：
			accessToken  公用Token
			accessTokenTime 过期时间
			err 错误
*/
func GetWXAccessToken(appid, appSecret string) (accessToken string, accessTokenTime time.Time, err error) {
	errtime := time.Now()
	url := `https://api.weixin.qq.com/cgi-bin/token`
	parameter := `grant_type=client_credential&appid=` + appid + `&secret=` + appSecret
	_,responseData, err := ghttp.HTTPRequestData(url, parameter, "GET", 5)
	if err != nil {
		return "", errtime, err
	}

	dataJSON := gjson.JSONToMap(responseData)

	//如果错误会返回: {"errcode":40164,"errmsg":"invalid ip 183.15.176.217, not in whitelist hint: [CznWNA09822974]"}
	//如果正确会返回：{"access_token":"13_CD-3i84w3sAUjwYpxCzJgaJ3UAOT2QcX1BJU2rNsG30y3TvlN-7oZJPiTFyPTPUOWV5QEA0H2aIwBOW-JeGmgIfiIdx19zPar-zKxOaGG06kuE8froR7CXlAnsm0q4HJK0f8zYowdrXCxjXYHGNaABATSZ","expires_in":7200}
	if dataJSON.GetString("errmsg") != "" { //如果微信返回了错误
		return "", errtime, fmt.Errorf(dataJSON.GetString("errmsg"))
	}
	accessTokenTime = time.Now().Add(time.Second * time.Duration(dataJSON.GetInt("expires_in")-60))
	return dataJSON.GetString("access_token"), accessTokenTime, nil
}

/*GetOpenidList 获取微信公众账号 openid 列表
功能： 获取微信公众账号 openid 列表
参数：
			access_token  调用接口凭证
			next_openid 第一个拉取的OPENID，不填默认从头开始拉取
参考：https://developers.weixin.qq.com/doc/offiaccount/User_Management/Getting_a_User_List.html
返回：
			openid  openid list
			next_openid 第一个拉取的OPENID，不填默认从头开始拉取
			err 错误
*/
func GetOpenidList(accesstoken string, oldnextOpenid string) (openidList OpenidListData, err error) {
	url := `https://api.weixin.qq.com/cgi-bin/user/get`
	parameter := `access_token=` + accesstoken + `&next_openid=` + oldnextOpenid
	_,responseData, err := ghttp.HTTPRequestData(url, parameter, "GET", 5)
	if err != nil {
		return openidList, err
	}
	dataJSON := gjson.JSONToMap(responseData)

	//如果错误会返回: {"errcode":40164,"errmsg":"invalid ip 183.15.176.217, not in whitelist hint: [CznWNA09822974]"}
	//如果正确会返回：{
	//     "total":2,
	//     "count":2,
	//     "data":{
	//     "openid":["OPENID1","OPENID2"]},
	//     "next_openid":"NEXT_OPENID"
	// }
	if dataJSON.GetString("errmsg") != "" { //如果微信返回了错误
		return openidList, fmt.Errorf(dataJSON.GetString("errmsg"))
	}
	err = json.Unmarshal([]byte(responseData), &openidList)
	if err != nil {
		return openidList, err
	}
	return openidList, nil

}

//OpenidListData 微信openid数据
type OpenidListData struct {
	Total int `json:"total"`
	Count int `json:"count"`
	Data  struct {
		Openid []string `json:"openid"`
	} `json:"data"`
	NextOpenid string `json:"next_openid"`
}

/*GetUserInfoWithAccessToken 使用Openid获取获取用户资料(普通accessToken)
功能：使用Openid获取获取用户资料(可获得用户是否关注)
参数：微信普通accessToken,用户openid
参考：https://mp.weixin.qq.com/wiki?t=resource/res_main&id=mp1421140839
返回：微信返回的json字符串，错误信息
*/
func GetUserInfoWithAccessToken(accessToken string, openid string) (string, error) {
	var parameter string
	requesturl := "https://api.weixin.qq.com/cgi-bin/user/info"
	parameter = parameter + "access_token=" + accessToken //公众号普通accesstoken
	parameter = parameter + "&openid=" + openid           //用户的Openid
	parameter = parameter + "&lang=" + "zh_CN"            //固定在值
	//sbjlog.Printfer("110", "GetUserInfoWithAccessToken:accessToken%s，openid:%s ", accessToken, openid)
	_,bodyData, err := ghttp.HTTPRequestData(requesturl, parameter, "GET", 5)
	if err != nil {
		return "", fmt.Errorf("getUserInfoWithAccessToken accessToken:%s Openid:%s \n	Err:%v ", accessToken, openid, err)
	}
	return bodyData, nil
}

/*获取用户微信资料例子
{
    "subscribe": 1,
    "openid": "o8jbut-xxxxxxxxxxxxxxxxxx",
    "nickname": "xxxx",
    "sex": 1,
    "language": "zh_CN",
    "city": "深圳",
    "province": "广东",
    "country": "中国",
    "headimgurl": "http://thirdwx.qlogo.cn/mmopen/fc6vOB4amw40OuFWIyK2fzicuL2KIwuibm9u4viaUFktOhNoDF2EZtc83WPhtTV3qWpGLiafUyoT06QibX7kmSm6wABYb3V3Qudoe/132",
    "subscribe_time": 1588813719,
    "unionid": "xxxxxxxxxxxxxxxxxxxxx",
    "remark": "",
    "groupid": 0,
    "tagid_list": [],
    "subscribe_scene": "ADD_SCENE_PROFILE_CARD",
    "qr_scene": 0, //二维码扫码场景
    "qr_scene_str": "" //二维码扫码场景描述
}
*/

//WxUserInfo 微信用户资料
type WxUserInfo struct {
	Subscribe      int           `json:"subscribe"`
	Openid         string        `json:"openid"`
	Nickname       string        `json:"nickname"`
	Sex            int           `json:"sex"`
	Language       string        `json:"language"`
	City           string        `json:"city"`
	Province       string        `json:"province"`
	Country        string        `json:"country"`
	Headimgurl     string        `json:"headimgurl"`
	SubscribeTime  int           `json:"subscribe_time"`
	Unionid        string        `json:"unionid"`
	Remark         string        `json:"remark"`
	Groupid        int           `json:"groupid"`
	TagidList      []interface{} `json:"tagid_list"`
	SubscribeScene string        `json:"subscribe_scene"`
	QrScene        int           `json:"qr_scene"`
	QrSceneStr     string        `json:"qr_scene_str"`
}

//DecodeUserInfo 解析用户信息
func DecodeUserInfo(bodyString string) (*WxUserInfo, error) {
	var item WxUserInfo
	err := json.Unmarshal([]byte(bodyString), &item)
	return &item, err
}

//====================  网页授权 相关 开始====================================

/*GetUserAccessTokenWithCode 根据Code换取WebAccessToken 等资料
功能：使用Code调用微信网页授权换取用户的openID及 WebAccessToken 等资料
			换取获得例子：{"access_token":"13_M2tGrbypGzp8UqpXOE5CJymfJQfg6BBdGYTDAQOwJ_M_7FTS9tVORA93bceVKZM6BXgoUslNxMvOEnQbeZBrbXn0zGRxb-FF5u9X3tB8_BY","expires_in":7200,"refresh_token":"13_fK5W0swn9iKcuD4xt2IHxG3h6XmPlghWjH5iKY42yv1diEepEcjyYFZKWKzQ8H85-TnyUTgcSzK0L-ap088vaM7ejhzEphK2Je8kifaWV3E","openid":"os7Cm1I-dqeQoBvXOgG0W9F-_5eE","scope":"snsapi_base"}
参数：
			Appid 公众号唯一ID  "wx3166b12e65274c9b"
			Appsecret 公众号的appsecret	"add0ddcbae5e7ae064a829cfd4c9338a"
			code 获得的Code
返回：微信返回的json字符串，错误信息
*/
func GetUserAccessTokenWithCode(appid, appsecret, code string) (string, error) {
	var parameter string
	requesturl := "https://api.weixin.qq.com/sns/oauth2/access_token"
	parameter = parameter + "appid=" + appid                      // "wx3166b12e65274c9b"                 //公众号唯一ID
	parameter = parameter + "&secret=" + appsecret                // "add0ddcbae5e7ae064a829cfd4c9338a" //公众号的appsecret
	parameter = parameter + "&code=" + code                       //获得的Code
	parameter = parameter + "&grant_type=" + "authorization_code" //固定在值
	_,bodyData, err := ghttp.HTTPRequestData(requesturl, parameter, "GET", 5)
	if err != nil {
		return "", fmt.Errorf("getUserAccessTokenWithCode appid:%s appsecret:%s code:%s\n	Err:%v ", appid, appsecret, code, err)
	}
	return bodyData, nil
}

/*GetUserInfoWithWebaccessToken 使用Openid获取获取用户资料（网页授权accessToken，这个accessToken获取需要用户同意）
功能：使用Openid获取获取用户资料(可获得用户是否关注)
参数：微信网页授权accessToken,用户openid
参考：https://mp.weixin.qq.com/wiki?t=resource/res_main&id=mp1421140842
返回：微信返回的json字符串，错误信息
*/
func GetUserInfoWithWebaccessToken(webaccessToken string, openid string) (string, error) {
	var parameter string
	requesturl := "https://api.weixin.qq.com/sns/userinfo"
	parameter = parameter + "access_token=" + webaccessToken //网页accesstoken
	parameter = parameter + "&openid=" + openid              //用户的Openid
	parameter = parameter + "&lang=" + "zh_CN"               //固定在值
	_,bodyData, err := ghttp.HTTPRequestData(requesturl, parameter, "GET", 5)
	if err != nil {
		return "", fmt.Errorf("getUserInfoWithWebAccessToken webaccessToken:%s Openid:%s\n Err:%v ", err, webaccessToken, openid)
	}
	return bodyData, nil
}

//WXMessage 微信消息
type WXMessage struct {
	ToUserName   string  //	开发者微信号
	FromUserName string  //	发送方帐号（一个OpenID）
	CreateTime   int     //	消息创建时间 （整型）
	MsgType      string  //	消息类型，event(事件) CLICK（点击菜单）text(文本消息)  image（图片消息）
	Event        string  //	事件类型，subscribe(订阅)、unsubscribe(取消订阅)
	EventKey     string  //	事件KEY值，qrscene_为前缀，后面为二维码的参数值
	Latitude     float64 //	地理位置纬度
	Longitude    float64 //	地理位置经度
	Precision    float64 //	地理位置精度
	Ticket       string  //	二维码的ticket，可用来换取二维码图片
	Content      string  //	文本消息内容
	MsgID        int64   `xml:"MsgId"`   // 消息id，64位整型 (微信官方为 MsgId	注意ID大小写)
	PicURL       string  `xml:"PicUrl"`  // 图片链接（由微信系统生成）(微信官方为 PicUrl	注意URL大小写)
	MediaID      string  `xml:"MediaId"` // 图片消息媒体id，可以调用多媒体文件下载接口拉取数据。 (微信官方为 MediaId		注意ID大小写)
}

//WXEncryptionMessage 微信加密消息
type WXEncryptionMessage struct {
	ToUserName string //	开发者微信号
	Encrypt    string //	加密信息内容
}

type xmldata struct { //用户消息的推送数据xml格式
	ToUserName string `xml:"ToUserName"` //微信开发者账号
	Encrypt    string `xml:"Encrypt"`    //加密数据
}

/*DecryptMsg 解密用户消息
参数 WXID 微信ID, sMsgSignature 签名, sTimeStamp 时间戳, sNonce 随机字符串, sPostData POST获取的内容
*/
func DecryptMsg(token *gdbwx.TbWeixinAccount, sMsgSignature, sTimeStamp, sNonce, sPostData string) (string, error) {
	//把xml数据解析成xmldata对象
	var m xmldata
	err := xml.Unmarshal([]byte(sPostData), &m)
	if err != nil {
		return "", fmt.Errorf("解析XML消息失败:%s", err.Error())
	}
	sEncryptMsg := m.Encrypt                                                             //获取加密里面的内容
	if GenarateSinature(token.Token, sTimeStamp, sNonce, sEncryptMsg) != sMsgSignature { //进行签名校验
		return "", fmt.Errorf("加密消息签名校验失败")
	}
	// Decode base64
	cipherData, errs := base64.StdEncoding.DecodeString(sEncryptMsg) //用base64 编码转成数组
	// AES Decrypt
	plainData, errs := aesDecrypt(token.Encodingaeskey, cipherData) //解密
	if errs != nil {
		return "", fmt.Errorf("解密数据出现错误:%s", errs.Error())
	}
	// Read length
	buf := bytes.NewBuffer(plainData[16:20]) //取16-20中4个字节的长度int32数据数组
	var length int32
	binary.Read(buf, binary.BigEndian, &length) //根据大端编码 读取数值 （有效数据的长度）
	//加密前内容格式  随机字符串+数据长度（4个字节）+数据+Appid
	// appID validation
	appIDstart := 20 + length
	id := plainData[appIDstart : int(appIDstart)+len(token.Appid)] //取appid长度对应的数据出来进行对比
	if string(id) != token.Appid {                                 //如果Appid不相同 说明数据不正确校验失败
		return "", errors.New("Appid is invalid")
	}
	return string(plainData[20 : 20+length]), nil //取数据 [20:20+数据长度]
}

/*EncryptMsg 加密用户消息 得到XML的包
参数 WXID 微信ID, sMsgSignature 签名, sTimeStamp 时间戳, sNonce 随机字符串, sPostData POST获取的内容
*/
func EncryptMsg(token *gdbwx.TbWeixinAccount, sTimeStamp, sNonce, Data string) (string, error) {
	//random(16B) + msg_len(4B) + msg + appid
	var MsgByte []byte
	var NonceByte = []byte(sNonce)
	var AppidByte = []byte(token.Appid)
	var DataByte = []byte(Data)
	var _datalength = make([]byte, 4)
	datalen := uint32(len(DataByte))
	binary.BigEndian.PutUint32(_datalength, datalen)
	MsgByte = append(MsgByte, NonceByte...)   //random(16B)
	MsgByte = append(MsgByte, _datalength...) //msg_len(4B)
	MsgByte = append(MsgByte, DataByte...)    //msg
	MsgByte = append(MsgByte, AppidByte...)   //AESKey
	EncryptBytes, errs := AesEncryptPKCS7(MsgByte, token.Encodingaeskey)
	if errs != nil {
		return "", fmt.Errorf("加密数据出现错误:%s", errs.Error())
	}
	EncryptStr := base64.StdEncoding.EncodeToString(EncryptBytes)
	sMsgSignature := GenarateSinature(token.Token, sTimeStamp, sNonce, EncryptStr) //生成签名
	var sEncryptMsg string
	var EncryptLabelHead = "<Encrypt><![CDATA["
	var EncryptLabelTail = "]]></Encrypt>"
	var MsgSigLabelHead = "<MsgSignature><![CDATA["
	var MsgSigLabelTail = "]]></MsgSignature>"
	var TimeStampLabelHead = "<TimeStamp><![CDATA["
	var TimeStampLabelTail = "]]></TimeStamp>"
	var NonceLabelHead = "<Nonce><![CDATA["
	var NonceLabelTail = "]]></Nonce>"
	sEncryptMsg = sEncryptMsg + "<xml>" + EncryptLabelHead + EncryptStr + EncryptLabelTail
	sEncryptMsg = sEncryptMsg + MsgSigLabelHead + sMsgSignature + MsgSigLabelTail
	sEncryptMsg = sEncryptMsg + TimeStampLabelHead + sTimeStamp + TimeStampLabelTail
	sEncryptMsg = sEncryptMsg + NonceLabelHead + sNonce + NonceLabelTail
	sEncryptMsg += "</xml>"
	return sEncryptMsg, nil
}

/*GenarateSinature 获取签名(将所有的String数据进行字典排序，进行哈希sha1算法计算)*/
func GenarateSinature(token string, args ...string) string {
	var SortString []string
	SortString = append(SortString, token)
	SortString = append(SortString, args...)
	sort.Strings(SortString) //进行字典排序
	hx := ""
	for _, value := range SortString {
		hx += value
	}
	//哈希进行计算
	Sha1 := sha1.New()
	Sha1.Write([]byte(hx))
	bs := Sha1.Sum(nil)
	sha1Str := hex.EncodeToString(bs) //十六进制数据
	return strings.ToLower(sha1Str)
}

func aesDecrypt(Encodingaeskey string, cipherData []byte) ([]byte, error) {
	aesKey, _ := base64.StdEncoding.DecodeString(Encodingaeskey + "=") //添加=号
	k := len(aesKey)                                                   //PKCS#7
	if len(cipherData)%k != 0 {
		return nil, errors.New("crypto/cipher: ciphertext size is not multiple of aes key length")
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	plainData := make([]byte, len(cipherData))
	blockMode.CryptBlocks(plainData, cipherData)
	return plainData, nil
}

//DecodeMsg 解析消息 | xml字符串 -> 消息结构体
func DecodeMsg(xmlStr string) (WXMessage, error) {
	var item WXMessage
	err := xml.Unmarshal([]byte(xmlStr), &item)
	return item, err
}

//DecodeEncryptionMsg 解析加密的微信消息 （不含解密，只是为了拿到 ToUserName ）
func DecodeEncryptionMsg(xmlStr string) (*WXEncryptionMessage, error) {
	var item WXEncryptionMessage
	err := xml.Unmarshal([]byte(xmlStr), &item)
	return &item, err
}

//GetJSTicket 获取jsticket
func GetJSTicket(accessToken string) (JSTicket, error) {
	var item JSTicket
	//发送get请求
	respBody, err := ghttp.SendGET("https://api.weixin.qq.com/cgi-bin/ticket/getticket", fmt.Sprintf("access_token=%s&type=jsapi", accessToken), 3)
	if err != nil {
		return item, err
	}
	err = json.Unmarshal([]byte(respBody), &item)
	return item, err
}

/*QrSceneStr 生成二维码 字符串形式*/
func QrSceneStr(AccessToken string, ExpireSeconds int, sceneID string) (string, error) {
	strData := "{\"expire_seconds\": " + strconv.Itoa(ExpireSeconds) + ", \"action_name\": \"QR_STR_SCENE\", \"action_info\": {\"scene\": {\"scene_str\":  \"" + gjson.FormatJSONString(sceneID) + "\"}}}"
	return ghttp.SendPostJSON(`https://api.weixin.qq.com/cgi-bin/qrcode/create?access_token=`+AccessToken, strData, 10)
}

//JSTicket 微信jsticket
type JSTicket struct {
	Errcode   int    `json:"errcode"`
	Errmsg    string `json:"errmsg"`
	Ticket    string `json:"ticket"`
	ExpiresIn int    `json:"expires_in"`
}

//CreateSignature 构造微信签名
func CreateSignature(args map[string]string) string {
	//对所有待签名参数按照字段名的ASCII 码从小到大排序（字典序）
	keys := make([]string, 0)
	for k := range args {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	//按照排序好的key来构造url参数
	urlParams := ""
	for _, key := range keys {
		urlParams += fmt.Sprintf("&%s=%s", key, args[key])
	}
	urlParams = urlParams[1:] //去掉第一个&

	//创建一个hash变量
	hashItem := sha1.New()
	//写入数据
	hashItem.Write([]byte(urlParams))
	//将hash加密后的数据（[]byte）以字符串格式输出
	signature := hex.EncodeToString(hashItem.Sum([]byte(nil)))

	return signature
}

//templateMsgParams 模板消息参数
type templateMsgParams struct {
	Touser      string      `json:"touser"`
	TemplateID  string      `json:"template_id"`
	URL         string      `json:"url"`
	Data        interface{} `json:"data"`
	Miniprogram interface{} `json:"miniprogram"`
}

//Miniprogram  小程序参数
type Miniprogram struct {
	Appid    string `json:"appid"`
	Pagepath string `json:"pagepath"`
}

//templateMsgResult 模板消息返回结果
type templateMsgResult struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	Msgid   int64  `json:"msgid"`
}

//SendTemplateMsg 推送模板消息
func SendTemplateMsg(accessToken, templateID, urlStr, touser string, msgData interface{}, otherParams ...interface{}) (msgID int64, err error) {
	defer func() {
		if err != nil { //err不为空，封装err，注明err出处
			err = fmt.Errorf("xwxofficialaccount.SendTemplateMsg-%s", err)
		}
	}()
	var params templateMsgParams
	params.TemplateID = templateID //模板ID
	params.Touser = touser         //用户Openid
	params.URL = urlStr            //点击跳转URL
	params.Data = msgData          //消息渲染参数 - 具体根据消息模板定义
	//需要跳转小程序
	if len(otherParams) > 0 {
		params.Miniprogram = otherParams[0]
	}
	resp, err := ghttp.SendPostForm("https://api.weixin.qq.com/cgi-bin/message/template/send?access_token="+accessToken, gjson.MapToJSON(params), 5)
	if err != nil {
		return 0, err
	}
	//json解析返回数据
	var item templateMsgResult
	json.Unmarshal([]byte(resp), &item)
	if err != nil {
		return 0, fmt.Errorf("json解析失败:%s", resp)
	}
	//判断errcode
	if item.Errcode != 0 {
		return 0, errors.New(item.Errmsg) //返回错误消息
	}
	return item.Msgid, nil //返回消息ID
}

/*KfOnlineListData 客服列表*/
type KfOnlineListData struct {
	KfOnlineList []KfOnlineList `json:"kf_online_list"`
}

/*KfOnlineList 客服状态*/
type KfOnlineList struct {
	KfAccount    string `json:"kf_account"`
	Status       int    `json:"status"`
	KfID         int    `json:"kf_id"`
	AcceptedCase int    `json:"accepted_case"`
}

//QueryCustomerServiceStatus 查询所有客服状态
//参数 accessToken 微信公众号AccessToken
func QueryCustomerServiceStatus(accessToken string) ([]KfOnlineList, error) {
	defer func() {
		if err := recover(); err != nil {
			//打印错误日志
			glog.Debug("QueryCustomerServiceStatus 查询所有客服状态 Err:%s\n", err)
		}
	}()
	url := "https://api.weixin.qq.com/cgi-bin/customservice/getonlinekflist?access_token=" + accessToken
	result, err := ghttp.SendPostForm(url, "", 5)
	fmt.Println(result)
	var kfOnlineListData KfOnlineListData
	//客服完整账号
	err = json.Unmarshal([]byte(result), &kfOnlineListData)
	if err != nil {
		return nil, err
	}
	return kfOnlineListData.KfOnlineList, err
}

//CreateConversation 创建会话
//参数 accessToken 微信公众号AccessToken, kfAccount 客户账号, openID Openid
func CreateConversation(accessToken, kfAccount, openID string) (bool, string, error) {
	defer func() {
		if err := recover(); err != nil {
			//打印错误日志
			glog.Debug("CreateConversation 创建会话 Err:%s\n", err)
		}
	}()
	url := "https://api.weixin.qq.com/customservice/kfsession/create?access_token=" + accessToken
	data := map[string]interface{}{
		"kf_account": kfAccount,
		"openid":     openID,
	}
	result, err := ghttp.SendPostForm(url, gjson.MapToJSON(data), 5)
	if err != nil {
		return false, "", err
	}
	return true, result, err
}


//AesDecryptPKCS7 解密方法（微信消息验证通用）
func AesDecryptPKCS7(cipherData []byte, Encodingaeskey string) ([]byte, error) {
	aesKey, _ := base64.StdEncoding.DecodeString(Encodingaeskey + "=") //添加=号
	k := len(aesKey)                                                   //PKCS#7
	if len(cipherData)%k != 0 {
		return nil, errors.New("crypto/cipher: ciphertext size is not multiple of aes key length")
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	plainData := make([]byte, len(cipherData))
	blockMode.CryptBlocks(plainData, cipherData)
	return plainData, nil
}

//AesEncryptPKCS7 内部使用 用户消息 AES加密方法
func AesEncryptPKCS7(plainData []byte, Encodingaeskey string) ([]byte, error) {
	aesKey, _ := base64.StdEncoding.DecodeString(Encodingaeskey + "=")
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	iv := aesKey[:aes.BlockSize]
	k := len(aesKey)
	if len(plainData)%k != 0 {
		plainData = pkcs7Padding(plainData, k)
	}
	cipherData := make([]byte, len(plainData))
	blockMode := cipher.NewCBCEncrypter(block, iv)
	blockMode.CryptBlocks(cipherData, plainData)
	return cipherData, nil
}

func pkcs7Padding(message []byte, blocksize int) (padded []byte) {
	padlen := blocksize - len(message)%blocksize // calculate padding length
	if padlen == 0 {
		padlen = blocksize
	}
	padding := bytes.Repeat([]byte{byte(padlen)}, padlen) // define PKCS7 padding block
	padded = append(message, padding...)                  // apply padding
	return padded
}
