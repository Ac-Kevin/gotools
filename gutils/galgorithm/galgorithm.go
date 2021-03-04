package galgorithm


import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

/*
MD5ToUpper32 将字符串,转为32位md5加密，返回大写字母
*/
func MD5ToUpper32(str string) string {
	w := md5.New()
	io.WriteString(w, str)                  //将str写入到w中
	md5Str := fmt.Sprintf("%x", w.Sum(nil)) //w.Sum(nil)将w的hash转成[]byte格式
	return strings.ToUpper(md5Str)
}

/*
MD5ToLower32 将字符串,转为32位md5加密，返回小写字母
*/
func MD5ToLower32(str string) string {
	w := md5.New()
	io.WriteString(w, str)                  //将str写入到w中
	md5Str := fmt.Sprintf("%x", w.Sum(nil)) //w.Sum(nil)将w的hash转成[]byte格式
	return md5Str
}

/*
AESEncrypt 用于AES加密
*/
func AESEncrypt(plantText, keyText string) (string, error) {
	key := []byte(keyText)
	plantTextBytes := []byte(plantText)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "error", err
	}
	plantTextBytes = zeroPadding(plantTextBytes, block.BlockSize())
	blockModel := cipher.NewCBCEncrypter(block, key)

	ciphertext := make([]byte, len(plantTextBytes))

	blockModel.CryptBlocks(ciphertext, plantTextBytes)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

//AES加密CBC 填充算法，是为了加密补位
//默认的 blockSize=16(即采用16*8=128, AES-128长的密钥)，如密钥使用 256位(32个字符)，那么 blockSize=32
//所以当 ciphertext len 刚好被blockSize 整除，就不需要做填充
func zeroPadding(ciphertext []byte, blockSize int) []byte {
	if len(ciphertext)%blockSize == 0 {
		return ciphertext
	}
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(ciphertext, padtext...)

}

/*
AESDecrypt 用于AES解密
*/
func AESDecrypt(cipherText, key string) (string, error) {
	keyBytes := []byte(key)
	cipherTextBytes, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(keyBytes) //选择加密算法
	if err != nil {
		return "", err
	}
	blockModel := cipher.NewCBCDecrypter(block, keyBytes)
	plantText := make([]byte, len(cipherTextBytes))
	blockModel.CryptBlocks(plantText, cipherTextBytes)
	plantText = bytes.TrimRight(plantText, "\x00") //Tim 2018.09.01 去除加密时补位的空字符NULL
	return string(plantText), nil
}
