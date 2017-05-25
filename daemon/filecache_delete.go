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
	"Sirupsen/logrus"
)

func (daemon *Daemon) FileCacheDelete(Orig string)(deleteresponses []*types.FileCacheDeleteResponseItem,err error){
	hash := sha256.New()
	hash.Write([]byte(Orig))
	md := hash.Sum(nil)
	mdStr := hex.EncodeToString(md)
	pth:=filepath.Join(builder.Filecachejsonpath,mdStr)
	_,err=os.Stat(pth)
	if os.IsNotExist(err){
		deleteresponse:=&types.FileCacheDeleteResponseItem{
			Notexist:pth+"is not exist",
		}
		deleteresponses=append(deleteresponses,deleteresponse)
		return
	}
	logrus.Debug("remove orig is ",Orig,"filename is ",pth)
	err=os.Remove(pth)
	if err!=nil{
		return
	}
	deleteresponse:=&types.FileCacheDeleteResponseItem{
		Orig:Orig,
	}
	deleteresponses=append(deleteresponses,deleteresponse)
	return

}
