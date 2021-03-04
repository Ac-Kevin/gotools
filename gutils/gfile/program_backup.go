package gfile

import (
	"archive/zip"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// CONFIGFILE 默认打包的配置文件
	CONFIGFILE = "config.ini"
	// DEFAULTVERSION 默认版本号
	DEFAULTVERSION = "V1.00"
)

// ProgramBackup functions Author:LC  time:2019-08-06
// ProgramBackup 程序运行时自动备份
// 使用方法：在main函数下调用即可，注意备份文件夹不存在会创建备份文件夹，重新编译文件后运行会生成新的备份
// 打包策略：采用取运行文件名作为程序名，打包全称：程序ID_程序名_版本号_MD5.zip
// 功能：程序运行时备份
// 参数：programID 程序ID；version 程序版本号;
// 返回: error 错误
func ProgramBackup(programID int, version string, customPath string) (err error) {
	var (
		// fileSlice 打包的文件句柄列表
		fileSlice = make([]*os.File, 0)
	)

	// 检查路径的正确性，以及是否要创建路径
	directoryPath := fmt.Sprintf("./%d_ProgramBackup", programID) //备份目录

	// 判断文件夹是否存在，不存在则创建
	if _, err := os.Stat(directoryPath); err != nil {
		os.MkdirAll(directoryPath, 0777)
	}

	// 获取当前文件全路径
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		if strings.HasSuffix(err.Error(), os.ErrNotExist.Error()) {
			return nil
		}
		return errors.New("xfile.ProgramBackup failed, " + err.Error())
	}

	// 打包文件到备份文件夹
	fileSlice = append(fileSlice, openFile(file), openFile(CONFIGFILE)) // 集合打包的文件的句柄
	if customPath != "" {
		fileNameList, err := findCurDirMatchPath(customPath)
		if err != nil {
			fmt.Printf("xfile.ProgramBackup failed, 备份程序 <%d> 正则表达式匹配错误 %s\n", programID, err.Error())
		} else {
			for _, cp := range *fileNameList {
				fileSlice = append(fileSlice, openFile(cp)) // 集合打包的文件的句柄
			}
		}
	}

	// 程序运行结束要关闭掉打开的文件句柄
	defer closeFiles(fileSlice)

	// 获取文件的hash值
	fileMd5, err := fileMD5(fileSlice)
	if err != nil {
		return errors.New("xfile.ProgramBackup MD5 failed, " + err.Error())
	}

	// 组合成打包文件名全称
	fullPath := fmt.Sprintf("%s/%d_%s_%s_%s.zip",
		directoryPath,
		programID,
		filepath.Base(file),       // 取运行文件名作为程序名
		decoratorVersion(version), // 重构 可判断传进来的版本带不带V, 修改字符串拼接方式为模板字符串拼接 lc
		fileMd5)

	// 判断要打包的文件是否在备份文件中(存在则不处理)
	_, err = os.Stat(fullPath)
	if err == nil {
		return nil
	}

	// 打包zip
	resetFiles(fileSlice) // 重置文件句柄，可以再次从文件中读取文件内容
	err = toZip(fileSlice, fullPath)
	if err != nil {
		return errors.New("xfile.ProgramBackup zip failed, " + err.Error())
	}
	return nil
}

// decoratorVersion 处理版本号
func decoratorVersion(version string) string {
	if version == "" {
		return DEFAULTVERSION
	}
	v := strings.ToUpper(version)
	if !strings.HasPrefix(v, "V") {
		v = "V" + v
	}
	return v
}

// resetFiles 重置文件游标到起始位置，使得文件可以再次读取文件内容
func resetFiles(files []*os.File) {
	for _, file := range files {
		if file != nil {
			file.Seek(io.SeekStart, io.SeekStart)
		}
	}
}

// closeFiles 关闭文件句柄
func closeFiles(files []*os.File) {
	for _, file := range files {
		if file != nil {
			file.Close()
		}
	}
}

// fileMD5 对文件进行md5计算
// 参数：files 文件句柄slice
// 返回：hash值；error 错误
func fileMD5(files []*os.File) (string, error) {
	var err error
	h := md5.New()
	for _, file := range files {
		if file == nil {
			continue
		}
		_, err = io.Copy(h, file)
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// toZip 压缩函数
// 参数：files 文件句柄slice
//      destZip 打包压缩到的路径
// 返回：error 错误
func toZip(files []*os.File, destZip string) error {
	zipFile, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	for _, file := range files {
		if file == nil {
			continue
		}
		err = compress(file, "", archive)
		if err != nil {
			return err
		}
	}
	return nil
}

// compress 压缩成zip的方法
// 参数 file 文件句柄，prefix文件名前缀默认为空 zw压缩对象
// 返回 error
func compress(file *os.File, prefix string, zw *zip.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	defer file.Close()

	if info.IsDir() {
		prefix = prefix + "/" + info.Name()
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return err
			}
			err = compress(f, prefix, zw)
			if err != nil {
				return err
			}
		}
	} else {
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = prefix + "/" + header.Name
		header.Method = zip.Deflate // 设置压缩
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, file)
		if err != nil {
			return err
		}
	}
	return nil
}

// openFile 打开文件，获取文件句柄
// 参数 path 文件路径
// 返回 文件句柄
func openFile(path string) *os.File {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}

	return f
}

// findCurDirMatchPath 找到当前文件夹下所有匹配的文件名，如匹配 *.ini 等类型文件
// 参数 customPath 匹配符
// 返回 文件名列表，error
func findCurDirMatchPath(customPath string) (*[]string, error) {
	customPathSlice := make([]string, 0)
	//修改使用正则表达式匹配的方式
	reg, err := regexp.Compile(customPath)
	if err != nil {
		return &customPathSlice, err
	}
	dirList, err := ioutil.ReadDir("./")
	if err != nil {
		return &customPathSlice, err
	}
	for _, v := range dirList {
		if !v.IsDir() && reg.MatchString(v.Name()) {
			customPathSlice = append(customPathSlice, v.Name())
		}
	}
	return &customPathSlice, nil
}