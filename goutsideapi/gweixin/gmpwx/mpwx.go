package gmpwx

import (
	"ackevin.com/gutils/ghttp"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

const domain string = "https://api.weixin.qq.com"

//UserInfo 用户信息
type UserInfo struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	Unionid    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

//Code2Session code换取用户信息
//传入参数  appid 小程序应用id, secret 小程序应用secret, code 用户code
func Code2Session(appid, secret, code string) (UserInfo, error) {
	var userinfo UserInfo
	var err error
	urlPath := domain + "/sns/jscode2session"
	params := fmt.Sprintf("appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", appid, secret, code)
	var result string
	result, err = ghttp.SendGET(urlPath, params, 10)
	if err != nil {
		return userinfo, fmt.Errorf("Http error : %s", err.Error())
	}
	err = json.Unmarshal([]byte(result), &userinfo)
	if err != nil {
		return userinfo, fmt.Errorf("json Unmarshal error : %s", err.Error())
	}
	if userinfo.ErrCode != 0 {
		return userinfo, fmt.Errorf("api error :%d-%s", userinfo.ErrCode, userinfo.ErrMsg)
	}
	return userinfo, err
}

//DecodeEncryptedData 加密加密的数据库
func DecodeEncryptedData(encryptedData, sessionKey, iv string) (string, error) {
	aesKey, _ := base64.StdEncoding.DecodeString(sessionKey)
	ivKey, _ := base64.StdEncoding.DecodeString(iv)
	data, _ := base64.StdEncoding.DecodeString(encryptedData)
	//AES CBC - plainData
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	blockMode := cipher.NewCBCDecrypter(block, ivKey)
	plainData := make([]byte, len(data))
	blockMode.CryptBlocks(plainData, data)
	//PKCS7UnPadding -  length  - unpadding
	length := len(plainData)
	unpadding := int(plainData[length-1])
	return string(plainData[:(length - unpadding)]), nil
}

//PhoneNumber 手机号码数据
type PhoneNumber struct {
	PhoneNumber     string `json:"phoneNumber"`
	PurePhoneNumber string `json:"purePhoneNumber"`
	CountryCode     string `json:"countryCode"`
	Watermark       struct {
		Timestamp int    `json:"timestamp"`
		Appid     string `json:"appid"`
	} `json:"watermark"`
}

//DecodePhoneNumber 解密手机号码数据
func DecodePhoneNumber(encryptedData, sessionKey, iv string) (PhoneNumber, error) {
	var pn PhoneNumber
	dataStr, err := DecodeEncryptedData(encryptedData, sessionKey, iv)
	if err != nil {
		return pn, err
	}
	err = json.Unmarshal([]byte(dataStr), &pn)
	if err != nil {
		return pn, err
	}
	return pn, nil
}

//DecodeUserInfo 被加密的用户信息
type DecodeUserInfo struct {
	OpenID    string `json:"openId"`    //用户Openid
	NickName  string `json:"nickName"`  //昵称
	Gender    int    `json:"gender"`    //性别
	Language  string `json:"language"`  //语言
	City      string `json:"city"`      //城市
	Province  string `json:"province"`  //省份
	Country   string `json:"country"`   //国家
	AvatarURL string `json:"avatarUrl"` //头像
	UnionID   string `json:"unionId"`   //unionid
	Watermark struct {
		Timestamp int    `json:"timestamp"`
		Appid     string `json:"appid"`
	} `json:"watermark"`
}

//DecodeGetUserInfo 解密获取用户书序
func DecodeGetUserInfo(encryptedData, sessionKey, iv string) (DecodeUserInfo, error) {
	var pn DecodeUserInfo
	dataStr, err := DecodeEncryptedData(encryptedData, sessionKey, iv)
	if err != nil {
		return pn, err
	}
	err = json.Unmarshal([]byte(dataStr), &pn)
	if err != nil {
		return pn, err
	}
	return pn, nil
}
