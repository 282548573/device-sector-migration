/*
 * @Author: bill
 * @Date: 2021-10-28 18:27:16
 * @LastEditors: bill
 * @LastEditTime: 2021-11-04 16:19:05
 * @Description:
 * @FilePath: /device-sector-migration/configs/config.go
 */
package configs

import (
	"errors"
	"io/ioutil"
	"log"

	jsoniter "github.com/json-iterator/go"
)

func GetJsonCfg(filePath string, val interface{}) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Unmarshal(data, &val)
}

func SetJsonCfg(filePath string, val interface{}) error {
	if val == nil {
		log.Println("SetJsonCfg val is nil")
		return errors.New("SetJsonCfg val is nil")
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b, err := json.MarshalIndent(&val, "", "  ")
	if err != nil {
		log.Println(err.Error())
		return err
	}

	err = ioutil.WriteFile(filePath, b, 0644)
	return err
}
