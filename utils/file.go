/*
 * @Author: bill
 * @Date: 2021-10-28 18:27:16
 * @LastEditors: bill
 * @LastEditTime: 2021-11-08 09:58:00
 * @Description:
 * @FilePath: /device-sector-migration/utils/file.go
 */
package utils

import (
	"device-sector-migration/global"
	"fmt"
	"os"
)

func FileOrDirIsExists(path string) (existed bool) {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// fileName:文件名字(带全路径)
// content: 写入的内容
func AppendToFile(fileName string, content string) error {
	// 以只写的模式，打开文件
	f, err := os.OpenFile(fileName, os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("cacheFileList.yml file create failed. err: " + err.Error())
	} else {
		// 查找文件末尾的偏移量
		n, err := f.Seek(0, os.SEEK_END)
		if err != nil {
			global.Logger.Error(err.Error())
		}
		// 从末尾的偏移量开始写入内容
		_, err = f.WriteAt([]byte(content), n)
		if err != nil {
			global.Logger.Error(err.Error())
		}
	}
	defer f.Close()
	return err
}
