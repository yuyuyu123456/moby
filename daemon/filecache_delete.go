/**
 * Created by zizhi.yuwenqi on 2017/5/25.
 */

package daemon

import (
	"github.com/docker/docker/api/types"
	"encoding/hex"
	"path/filepath"
	"crypto/sha256"
	"github.com/docker/docker/builder"
	"os"
	"github.com/Sirupsen/logrus"
)

func (daemon *Daemon) FileCacheDelete(Orig string)(deleteresponses []*types.FileCacheDeleteResponseItem,err error){
	hash := sha256.New()
	hash.Write([]byte(Orig))
	md := hash.Sum(nil)
	mdStr := hex.EncodeToString(md)
	pth:=filepath.Join(builder.Filecachejsonpath,mdStr)
	//_,err=os.Stat(pth)
	//if os.IsNotExist(err){
	//	logrus.Debug("FileCacheDelete file is not exist ",Orig)
	//	deleteresponse:=&types.FileCacheDeleteResponseItem{
	//		Notexist:pth+"is not exist",
	//	}
	//	deleteresponses=append(deleteresponses,deleteresponse)
	//	return
	//}
	//logrus.Debug("remove orig is ",Orig,"filename is ",pth)
	//var filemetadata *builder.FileMetaData
	//filemetadata,err=daemon.FromDisk(mdStr)
	//if err!=nil{
	//	logrus.Debug("FileCacheDelete fromdisk error: ",err)
	//	return
	//}
	//_,err =daemon.filecache.SetCopyInfo(filemetadata.Orig,filemetadata.Copyinfoandlastmod,false)
	_,err=daemon.filecache.DelCopyInfo([]string{Orig})
	logrus.Debug("filecachedelete err: ",err)
	if err!=nil{
		logrus.Errorf("Failed to DelFile error:%v",err)
	}
	//err=os.Remove(pth)
	//if err!=nil{
	//	logrus.Debug("FileCacheDelete remove file error ",err)
	//	return
	//}
	if os.IsNotExist(err){
			logrus.Debug("FileCacheDelete file is not exist ",Orig)
			deleteresponse:=&types.FileCacheDeleteResponseItem{
				Notexist:pth+"  Orig file is not exist",
			}
			deleteresponses=append(deleteresponses,deleteresponse)
		        err=nil
			return
		}
	deleteresponse:=&types.FileCacheDeleteResponseItem{
		Orig:Orig+":filepath "+pth,
	}
	deleteresponses=append(deleteresponses,deleteresponse)
	return

}
