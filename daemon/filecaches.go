/**
 * Created by zizhi.yuwenqi on 2017/5/25.
 */

package daemon

import (
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"github.com/docker/docker/builder"
)

func(daemon *Daemon) FileCaches(filecachesfilters filters.Args,withExtraAttrs bool)(filecacheSummarys []*types.FileCacheSummary,err error){
	logrus.Info("Loading filecache: start.")

	dir, err := ioutil.ReadDir(daemon.filecachedir)
	if err != nil {
		return
	}
	//load DefaultCapacity filecache when start a daemon
	for _, v := range dir {
		filename := v.Name()
		filemetadata, err := daemon.FromDisk(filename)
		if err != nil {
			logrus.Errorf("Failed to load filecache %v: %v", filename, err)
			continue
		}
		logrus.Debug("load file cache",filename)
		//_,err =daemon.filecache.SetCopyInfo(filemetadata.Orig,filemetadata.Copyinfoandlastmod,false)
		for _,vv:=range filemetadata.Copyinfoandlastmod.Infos {
			fileinfo:=vv.FileInfo.(*builder.HashedFileInfo)
			fileinfo1:=(fileinfo.FileInfo).(builder.PathFileInfo)
			filecacheSummary := &types.FileCacheSummary{
				Orig:filemetadata.Orig,
				FileHash:fileinfo.FileHash,
				FileName:fileinfo1.FileName,
				FilePath:fileinfo1.FilePath,
				LastMod:filemetadata.Copyinfoandlastmod.LastMod,
			}
			filecacheSummarys = append(filecacheSummarys,filecacheSummary)
		}
		//if err!=nil{
		//	logrus.Errorf("Failed to SetCopyInfo %v:%v",filename,err)
		//	continue
		//}
		logrus.Debug("list filecaches load end")
	}
	return

}
