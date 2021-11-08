/*
 * @Author: sky
 * @Date: 2021-09-16 11:27:29
 * @LastEditors: bill
 * @LastEditTime: 2021-09-16 11:27:29
 * @Description:
 * @FilePath: /device_server/utils/ars_md5_file.go
 */

package utils

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

const hashChunk = 2 << 10 //2KB

func ArsFileHash(path string) (string, error) {
	s, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if s.IsDir() {
		return "", errors.New(fmt.Sprintf("%s is dir", path))
	}
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	size, err := file.Seek(0, os.SEEK_END)
	if err != nil {
		return "", err
	}
	_, err = file.Seek(0, os.SEEK_SET)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	//大于 chunk * 3 则分片哈希 。小于，则全部哈希
	if size > hashChunk*3 {
		//input 最前面一块
		buffer := make([]byte, hashChunk)
		_, err := file.Read(buffer)
		if err != nil {
			return "", err
		}
		h.Write(buffer)
		//input 中间一块
		file.Seek(size>>1, os.SEEK_SET)
		_, err = file.Read(buffer)
		if err != nil {
			return "", err
		}
		h.Write(buffer)
		//input 最后最后一块
		file.Seek(size-hashChunk, os.SEEK_SET)
		_, err = file.Read(buffer)
		if err != nil {
			return "", err
		}
		h.Write(buffer)
		//文件大小  int64 转字节 哈希
		binary.Write(h, binary.BigEndian, size)
	} else {
		io.Copy(h, file)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
