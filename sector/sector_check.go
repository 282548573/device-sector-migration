/*
 * @Author: bill
 * @Date: 2021-11-04 16:07:37
 * @LastEditors: bill
 * @LastEditTime: 2021-11-08 10:35:18
 * @Description:
 * @FilePath: /device-sector-migration/sector/sector_check.go
 */
package sector

import (
	"device-sector-migration/configs"
	"device-sector-migration/global"
	"device-sector-migration/utils"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/snapcore/snapd/osutil"
)

func SectorCheck() {
	// 1. 读取扇区文件
	var sectorList configs.SectorMaps
	if err := sectorList.ReadConfig(); err != nil {
		global.Logger.Error(err.Error())
		return
	}

	global.Logger.Info("==============================检查结果=====================================")

	result := "./result.log"
	if utils.FileOrDirIsExists(result) {
		os.RemoveAll(result)
	}
	f, err := os.Create(result)
	defer func() {
		if err := f.Close(); err != nil {
			global.Logger.Error(err.Error())
		}
	}()
	if err != nil {
		fmt.Println(err.Error())
	}

	for number, info := range sectorList.Maps {
		sectorNumber := fmt.Sprintf("s-%s-%d", info.MinerId, number)
		if info.IsFinish {
			if sourFiles, err := getMd5ListNew(info.SourcePath, info.MinerId, number); err != nil {
				global.Logger.Error(err.Error())
			} else {
				info.SourceFiles = sourFiles
			}
			if targetFiles, err := getMd5ListNew(info.TargetPath, info.MinerId, number); err != nil {
				global.Logger.Error(err.Error())
			} else {
				info.TargetFiles = targetFiles
			}

			sourMaps := make(map[string]int, len(info.SourceFiles))
			for index, file := range info.SourceFiles {
				sourMaps[file.FileName] = index
			}

			noEqualCount := 0
			for _, file := range info.TargetFiles {
				if i, ok := sourMaps[file.FileName]; ok {
					if file.Hash != info.SourceFiles[i].Hash {
						noEqualCount++
					}
				} else {
					noEqualCount++
				}
			}

			if noEqualCount == 0 {
				info.CheckResult = "迁移成功"
			} else {
				info.CheckResult = "迁移失败: 扇区Hash对比不同"
			}
		} else {
			msg := fmt.Sprintf("扇区(%s), 迁移失败,  扇区拷贝失败", sectorNumber)
			global.Logger.Info(msg)
			global.Logger.Info("===================================================================\n")
			info.CheckResult = "迁移失败: 扇区拷贝失败"
		}

		msg := fmt.Sprintf("扇区%d: 源目录%s   ====>   目标目录%s,%s", number, info.SourcePath, info.TargetPath, info.CheckResult)
		global.Logger.Info(msg)
		if err := utils.AppendToFile("./result.log", msg+"\n"); err != nil {
			global.Logger.Error(err.Error())
		}
	}

	global.Logger.Info(fmt.Sprintf("检查结果查看%s文件", result))

	if err := sectorList.WriteConfig(); err != nil {
		global.Logger.Error(err.Error())
	}
}

type HashFilesNew struct {
	FileName string
	FilePath string
}

func getMd5ListNew(diskMountpoint string, minerId string, sectorId int64) ([]configs.SectorFile, error) {
	md5InfoFiles := make([]HashFilesNew, 0)
	fpath := fmt.Sprintf("%s/.lotusworker/cache/s-%s-%d", diskMountpoint, minerId, sectorId)
	files, err := ioutil.ReadDir(fpath)
	if err != nil {
		message := fmt.Sprintf("获取%s下的文件失败,err:%s", fpath, err.Error())
		global.Logger.Error(message)
		return nil, errors.New(message)
	}
	for _, file := range files {
		md5InfoFiles = append(md5InfoFiles, HashFilesNew{
			FileName: file.Name(),
			FilePath: fmt.Sprintf("%s/.lotusworker/cache/s-%s-%d/%s", diskMountpoint, minerId, sectorId, file.Name()),
		})
	}

	sealFileName := fmt.Sprintf("s-%s-%d", minerId, sectorId)
	md5InfoFiles = append(md5InfoFiles, HashFilesNew{
		FileName: sealFileName,
		FilePath: fmt.Sprintf("%s/.lotusworker/sealed/%s", diskMountpoint, sealFileName),
	})

	return arsHashNew(md5InfoFiles, sectorId, diskMountpoint)
}

func arsHashNew(files []HashFilesNew, sectorId int64, path string) ([]configs.SectorFile, error) {
	sectorMd5s := make([]configs.SectorFile, len(files))
	for i := 0; i < len(files); i++ {
		// hash
		hash, err := utils.ArsFileHash(files[i].FilePath)
		if err != nil {
			global.Logger.Error(err.Error())
			return sectorMd5s, fmt.Errorf(fmt.Sprintf("get file %s hash err", files[i].FilePath))
		}

		// size
		size, err := utils.GetFileSize(files[i].FilePath)
		if err != nil {
			global.Logger.Error(err.Error())
			return sectorMd5s, fmt.Errorf(fmt.Sprintf("get file %s size err", files[i].FilePath))
		}

		// lsattr i
		file, err := os.Open(files[i].FilePath)
		if err != nil {
			global.Logger.Error(err.Error())
			return sectorMd5s, fmt.Errorf(fmt.Sprintf("open file %s err", files[i].FilePath))
		}

		attr, err := osutil.GetAttr(file)
		if err != nil {
			global.Logger.Error(err.Error())
			return sectorMd5s, fmt.Errorf(fmt.Sprintf("get file %s lsattr err", files[i].FilePath))
		}
		defer func() {
			if ferr := file.Close(); ferr != nil {
				global.Logger.Error(ferr.Error())
			}
		}()

		attrs := attr & osutil.FS_IMMUTABLE_FL
		immutable := false
		if attrs == osutil.FS_IMMUTABLE_FL {
			immutable = true
		}

		sectorMd5s[i] = configs.SectorFile{
			SectorID:  sectorId,
			FilePath:  path,
			FileName:  files[i].FileName,
			FileSize:  size,
			Hash:      hash,
			Immutable: immutable,
		}
	}
	return sectorMd5s, nil
}
