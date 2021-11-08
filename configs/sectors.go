/*
 * @Author: bill
 * @Date: 2021-11-04 16:17:29
 * @LastEditors: bill
 * @LastEditTime: 2021-11-08 10:49:15
 * @Description:
 * @FilePath: /device-sector-migration/configs/sectors.go
 */
package configs

import "device-sector-migration/global"

type SectorMaps struct {
	Maps map[int64]*SectorInfo `json:"maps"` // key: 扇区号
}

type SectorInfo struct {
	MinerId string `json:"miner"`
	//SectorNumber uint64 `json:"sector_number"`
	SourcePath  string       `json:"source_path"`  // 原始磁盘路径
	SourceFiles []SectorFile `json:"source_files"` // 原始磁盘扇区文件列表
	TargetPath  string       `json:"target_path"`  // 目标磁盘路径
	TargetFiles []SectorFile `json:"target_files"` // 目标磁盘扇区文件列表
	IsFinish    bool         `json:"is_finish"`    // 是否拷贝完成
	CheckResult string       `json:"check_result"` // 拷贝后扇区校验结果
}

type SectorFile struct {
	FileName  string `json:"file_name"`
	FileSize  int64  `json:"file_size"`
	Hash      string `json:"hash"`
	Immutable bool   `json:"immutable"`
	FilePath  string `json:"file_path"`
	SectorID  int64  `json:"sector_id"`
}

func (s *SectorMaps) ReadConfig() error {
	if err := GetJsonCfg("./configs/sectors.json", &s); err != nil {
		global.Logger.Error(err.Error())
		return err
	}
	return nil
}

func (s *SectorMaps) WriteConfig() error {
	if err := SetJsonCfg("./configs/sectors.json", &s); err != nil {
		global.Logger.Error(err.Error())
		return err
	}
	return nil
}
