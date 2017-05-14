/**
 * Created by zizhi.yuwenqi on 2017/5/12.
 */

package builder

import (
	"container/list"
	"encoding/hex"
	"path/filepath"
	"os"
	"sort"
	"strings"
	"errors"
	"crypto/sha256"
	"github.com/docker/docker/pkg/ioutils"
	"encoding/json"
	"github.com/docker/docker/pkg/urlutil"
	"net/url"
	"fmt"

	"github.com/Sirupsen/logrus"
	//"time"
)

type CopyInfo struct {
	FileInfo
	//HashedFileInfo
	Decompress bool
}
type CopyHashedFileInfo struct{
      *HashedPathFileInfo
      Decompress bool
}
type CopyHashedFileInfoAndLastMod struct {
	Infos   []CopyHashedFileInfo
	LastMod string
}
//type FileStat struct {
//	Name1    string
//	Size1    int64
//	Mode1   os.FileMode
//	ModTime1 time.Time
//	Sys1     interface{}
//}
//func (fs *FileStat) Size() int64        { return fs.Size1 }
//func (fs *FileStat) Mode() os.FileMode     { return fs.Mode1 }
//func (fs *FileStat) ModTime() time.Time { return fs.ModTime1 }
//func (fs *FileStat) Sys() interface{}   { return fs.Sys1 }
type FileCacheInter interface{
	GetCopyInfo(origins string)(CopyInfoAndLastMod,bool,error)
	SetCopyInfo(origins string,copyinfoandlastmod CopyInfoAndLastMod,todisk bool)(bool,error)
	DelCopyInfo(origins []string)(bool,error)
	GetFileCacheInfo(origins []string)(FileCacheInfo,bool)//TODO
	SetFileCacheInfo(origins []string,filecacheinfo FileCacheInfo)(bool,error)//TODO
	DelFileCacheInfo(origins []string)	(bool,error)
	DelAll()
}

//var fileca=&FileCache{
//	SingleFileCacheMap:make(map[string]CopyInfoAndLastMod),
//	FileCacheMap:make(map[string]FileCacheInfo),
//}

func NewFileCache()FileCacheInter{
	return &FileCache{
		SingleFileCacheMap:CopyInfoAndLastModMap{NewDefaultLruCache()},
		FileCacheMap:FileCacheInfoMap{NewDefaultLruCache()},
	}
}
type FileCacheInfo struct {
	//infos []copyInfo
	SrcHash   string
	OrigPaths string
}
type CopyInfoAndLastMod struct{
	Infos   []CopyInfo
	LastMod string
}
//type CopyHashedFileInfoAndLastMod struct {
//	Infos []CopyHashedFileInfo
//	LastMod string
//}
/*
list 保存数据维护顺序
map 查找
 */
type LruCache struct{
	capacity int
	list *list.List
	cacheMap map[string]*list.Element
}
type CopyInfoAndLastModMap struct{
	Copyinfolrucache *LruCache
}
type FileCacheInfoMap struct {
	FileCacheInfolrucache *LruCache
}
type FileCache struct {
	SingleFileCacheMap CopyInfoAndLastModMap
	FileCacheMap       FileCacheInfoMap
}
type FileMetaData struct {
	Orig               string
	Copyinfoandlastmod CopyInfoAndLastMod
	Filecacheinfo      FileCacheInfo
}
type FileMetaDataJson struct {
	Orig string
	CopyInfoAndLastMod CopyHashedFileInfoAndLastMod
	Filecacheinfo FileCacheInfo
}
//type HashedFileMetaData struct {
//	Orig               string
//	Copyinfoandlastmod CopyHashedFileInfoAndLastMod
//	Filecacheinfo      FileCacheInfo
//}
type LruCacheNode struct {
	key string
	value interface{}
}
const DefaultCapacity =2
func NewDefaultLruCache()(*LruCache){
	return &LruCache{
		capacity:DefaultCapacity,
		list:list.New(),
		cacheMap: make(map[string]*list.Element),
	}
}


func (lruCache *LruCache) Size() int {
	return lruCache.list.Len()
}

func (lruCache *LruCache)Set(k string,value interface{})(error){

	if lruCache.list == nil {
		return errors.New("LruCache结构体未初始化.")
	}

	if element,ok := lruCache.cacheMap[k]; ok {
		lruCache.list.MoveToFront(element)
		element.Value.(*LruCacheNode).value= value
		return nil
	}

	newElement := lruCache.list.PushFront( &LruCacheNode{k,value} )
	lruCache.cacheMap[k] = newElement

	if lruCache.list.Len() > lruCache.capacity {
		lastElement := lruCache.list.Back()
		if lastElement == nil {
			return nil
		}
		cacheNode := lastElement.Value.(*LruCacheNode)
		delete(lruCache.cacheMap,cacheNode.key)
		lruCache.list.Remove(lastElement)
	}
	return nil
}

func (lruCache *LruCache)Get(k string)(v interface{},ret bool,err error){

	if lruCache.cacheMap == nil {
		return v,false,errors.New("LRUCache结构体未初始化.")
	}

	if element,ok := lruCache.cacheMap[k]; ok {
		lruCache.list.MoveToFront(element)
		return element.Value.(*LruCacheNode).value,true,nil
	}
	return v,false,nil
}
func (lruCache *LruCache)Remove(k string)(bool){

	if lruCache.cacheMap == nil {
		return false
	}

	if pElement,ok := lruCache.cacheMap[k]; ok {
		delete(lruCache.cacheMap,k)
		lruCache.list.Remove(pElement)
		return true
	}
	return false
}
const filecachejsonpath  = "/var/lib/docker/filecachejson"
func (fileMetaData *FileMetaData)ToDisk()error{
	if err:=fileMetaData.checkFileMetaData();err!=nil{
		return err
	}
	hash := sha256.New()
	hash.Write([]byte(fileMetaData.Orig))
	md := hash.Sum(nil)
	mdStr := hex.EncodeToString(md)
	pth:=filepath.Join(filecachejsonpath,mdStr)
	jsonSource, err := ioutils.NewAtomicFileWriter(pth, 0644)
	if err != nil {
		return err
	}
	defer jsonSource.Close()

	enc := json.NewEncoder(jsonSource)
	logrus.Debug("save filecache json to disk")
	infos:=make([]CopyHashedFileInfo,len(fileMetaData.Copyinfoandlastmod.Infos))
	for i,v:=range fileMetaData.Copyinfoandlastmod.Infos{
		info:=(v.FileInfo).(*HashedFileInfo)
		info1:=(info.FileInfo).(PathFileInfo)
		infos[i].Decompress=v.Decompress
		infos[i].HashedPathFileInfo=&HashedPathFileInfo{
			FileHash:info.FileHash,
			PathFileInfoWithoutFileInfo:PathFileInfoWithoutFileInfo{FileName:info1.FileName,FilePath:info1.FilePath}}
	}
	filemetadatajson:=&FileMetaDataJson{
		Orig:fileMetaData.Orig,
		Filecacheinfo:fileMetaData.Filecacheinfo,
		CopyInfoAndLastMod:CopyHashedFileInfoAndLastMod{
			LastMod:fileMetaData.Copyinfoandlastmod.LastMod,
			Infos:infos},
	}

	// Save filecache settings
	if err := enc.Encode(filemetadatajson); err != nil {
		return err
	}

	return nil
}

func (fileMetaData *FileMetaData)FromDisk()(error,bool){
	hash := sha256.New()
	hash.Write([]byte(fileMetaData.Orig))
	md := hash.Sum(nil)
	mdStr := hex.EncodeToString(md)
	pth:=filepath.Join(filecachejsonpath,mdStr)
	jsonSource, err := os.Open(pth)
	if err != nil {
		if os.IsNotExist(err){
			return nil,false
		}
		return err,false
	}
	defer jsonSource.Close()

	dec := json.NewDecoder(jsonSource)
	filemetadatajson:=&FileMetaDataJson{}
	// Load container settings
	if err := dec.Decode(filemetadatajson); err != nil {
		return err,false
	}
	fileMetaData.Orig=filemetadatajson.Orig
	copyinfos:=make([]CopyInfo,len(filemetadatajson.CopyInfoAndLastMod.Infos))
	for i,v:=range filemetadatajson.CopyInfoAndLastMod.Infos{
		copyinfos[i].Decompress=v.Decompress
		var fileinfo os.FileInfo
		//if urlutil.IsURL(filemetadata.Orig){
		logrus.Debug("get fileinfo of filepath:",v.FilePath)
		fileinfo, err = os.Stat(v.FilePath)
		if err != nil {
			return err,false
		}
		copyinfos[i].FileInfo=&HashedFileInfo{FileInfo:PathFileInfo{FilePath: v.FilePath,FileName:v.FileName,FileInfo:fileinfo}, FileHash: v.FileHash}
	}
	fileMetaData.Copyinfoandlastmod=CopyInfoAndLastMod{
		LastMod:filemetadatajson.CopyInfoAndLastMod.LastMod,
		Infos:copyinfos,
	}
	fileMetaData.Filecacheinfo=filemetadatajson.Filecacheinfo
	if err:=fileMetaData.checkFileMetaData();err!=nil{
		return err,false
	}
	return nil,true
}
func removeDiskFile(orig string)(err error){
	hash := sha256.New()
	hash.Write([]byte(orig))
	md := hash.Sum(nil)
	mdStr := hex.EncodeToString(md)
	pth:=filepath.Join(filecachejsonpath,mdStr)
	//filemetadata:=&FileMetaData{Orig:orig}
	//if err=filemetadata.FromDisk();err!=nil{
	//	return
	//}
	if err=os.Remove(pth);err!=nil{
		return
	}
	//remove json file ,and must remove content file
	//remote url must remove local file
	if urlutil.IsURL(orig){
		filename,err1:=handleFileName(orig)
		if err1!=nil{
			return err1
		}
		tmpDir:="/var/lib/docker/remotefile"
		tmpFileName:= filepath.Join(tmpDir, filename)
		logrus.Debug("remove file from disk tmpfilename is ", tmpFileName)
		if err1=os.Remove(tmpFileName);err1!=nil{
			return err1
		}

	}
	return


}
func handleFileName(srcURL string)(string,error){
	// get filename from URL
	u, err := url.Parse(srcURL)
	if err != nil {
		return "",err
	}
	path := filepath.FromSlash(u.Path) // Ensure in platform semantics
	if strings.HasSuffix(path, string(os.PathSeparator)) {
		path = path[:len(path)-1]
	}
	parts := strings.Split(path, string(os.PathSeparator))
	filename := parts[len(parts)-1]
	if filename == "" {
		err = fmt.Errorf("cannot determine filename from url: %s", u)
		return "",err
	}
	return filename,nil

}
func (fileMetaData *FileMetaData)checkFileMetaData()error{
	if fileMetaData.Orig==""{
		return errors.New("FileMetaData orig is empty")
	}
	if fileMetaData.Copyinfoandlastmod.Infos==nil &&
		(fileMetaData.Filecacheinfo.SrcHash==""&&fileMetaData.Filecacheinfo.OrigPaths==""){
		return errors.New("FileMetaData coppyinfoandlastmod and filecacheinfo is empty")
	}
	return nil
}
func (filecache *FileCache)GetCopyInfo(origins string)(CopyInfoAndLastMod,bool,error)  {
	var copyinfoandlastmod CopyInfoAndLastMod
	if  len(origins)==0{
		logrus.Error("GetCopyInfo: origins key is nil or length is 0")
		return copyinfoandlastmod,false,errors.New("arg is error")
	}
	//if filecache.SingleFileCacheMap ==nil{
	//	logrus.Error("singleFileCacheMap is not initialized")
	//	logrus.Debug("initializing singleFileCacheMap")
	//	filecache.SingleFileCacheMap =make(map[string]CopyInfoAndLastMod)
	//	return copyinfoandlastmod,false
	//}

	//copyinfoandlastmod,exist:=filecache.SingleFileCacheMap[origins]
	v,exist,err:=filecache.SingleFileCacheMap.Copyinfolrucache.Get(origins)
	if err!=nil{
		return copyinfoandlastmod,false,err
	}
	if !exist{
		logrus.Debug("do not find copyinfos in memory")
		logrus.Debug("trying find copyinfos in disk")
		fileMetaData:=&FileMetaData{Orig:origins}
		if err,b:=fileMetaData.FromDisk();err!=nil{
			return copyinfoandlastmod,false,err
		}else if !b{
			logrus.Debug("copyinfos do not find in disk")
			return copyinfoandlastmod,false,nil
		}
		logrus.Debug("copyinfos find in disk and trying set data from disk ")
		copyinfoandlastmod=fileMetaData.Copyinfoandlastmod
		if _,err=filecache.SetCopyInfo(origins,copyinfoandlastmod,false);err!=nil{
			logrus.Debug("copyinfos find in disk but setcopyinfo fail err:%v",err)
			return copyinfoandlastmod,true,err
		}
		return copyinfoandlastmod,true,nil
	}
	if v,ok:=v.(CopyInfoAndLastMod);!ok{
		logrus.Warn("filecache.SingleFileCacheMap get origins is  not type CopyInfoAndLastMod")
	}else{
		copyinfoandlastmod=v
	}

	return copyinfoandlastmod,true,nil
}


func (filecache *FileCache)SetCopyInfo(origins string,copyinfoandlastmod CopyInfoAndLastMod,todisk bool)(bool,error){

	//if !checkFileCacheInfo(origins,filecacheinfo){
	//	return false,errors.New("filecacheinfo error")
	//}
	if len(copyinfoandlastmod.Infos)==0{
		return false,errors.New("copyinfo is empty")
	}

	//filecache.SingleFileCacheMap[origins]=copyinfoandlastmod
	if err:=filecache.SingleFileCacheMap.Copyinfolrucache.Set(origins,copyinfoandlastmod);err!=nil{
		logrus.Debug("setcopyinfo :SingleFileCacheMap.Set error",err)
		return false,err
	}
	if todisk{
		filemetadata:=&FileMetaData{
			Orig:origins,
			Copyinfoandlastmod:copyinfoandlastmod,
		}
		err:=filemetadata.ToDisk()
		if err!=nil{
			logrus.Debug("setcopyinfo :filemetadata Todisk err :",err)
			return false,err
		}
	}
	return true,nil
}

func (filecache *FileCache)DelCopyInfo(origins []string)(bool,error){
	var b bool
	var err error
	if origins==nil|| len(origins)==0{
		err=errors.New("key args is nil")
		return b,err
	}
	for _,orgin:=range origins{
		_,exist,err:=filecache.GetCopyInfo(orgin)
		if err!=nil{
			return b,err
		}
		if !exist{
			logrus.Debug("DelCopyInfo:key origin in singleFileCacheMap not found,do nothing")
		}else{
			logrus.Debug("DelCopyInfo:key origin in singleFileCacheMap deleting")
			//delete(filecache.SingleFileCacheMap,orgin)
			b=filecache.SingleFileCacheMap.Copyinfolrucache.Remove(orgin)
			if err=removeDiskFile(orgin);err!=nil{
				logrus.Warn("remove", orgin,"local file err",err)
			}
		}
	}
	return b,err
}

func (filecache *FileCache)GetFileCacheInfo(origins []string)(FileCacheInfo,bool){
	var info FileCacheInfo
	if origins==nil || len(origins)==0||len(origins)==1{
		logrus.Error("GetFileCacheInfo: origins key is nil or length is 0 oris not complex-valued")
		return info,false
	}

	sort.Strings(origins)
	for key,value:=range filecache.FileCacheMap.FileCacheInfolrucache.cacheMap{
		keys:=strings.Split(key,",")
		if compareSlice(origins,keys){
			v:=value.Value
			if info,ok:=v.(FileCacheInfo);ok{
				return info,true
			}

		}
	}
	logrus.Debug("GetFileCacheInfo:origins key not found")
	return info,false
}

func compareSlice(sli1 []string,sli2 []string)bool{
	if len(sli1)!=len(sli2){
		return false
	}
	for i,k:=range sli1{
		if sli2[i]!=k{
			return false
		}
	}
	return true
}

func (filecache *FileCache)SetFileCacheInfo(origins []string,filecacheinfo FileCacheInfo)(bool,error){
	if !checkFileCacheInfo(origins,filecacheinfo){
		return false,errors.New("SetFileCacheInfo failure")
	}

	sort.Strings(origins)
	key:=strings.Join(origins,",")
	//filecache.FileCacheMap[key]=filecacheinfo
	filecache.FileCacheMap.FileCacheInfolrucache.Set(key,filecacheinfo)
	return true,nil
}
func(filecache *FileCache) DelFileCacheInfo(origins []string)	(bool,error)  {
	if origins==nil|| len(origins)==0||len(origins)==1 {
		return false,errors.New("key args is nil")
	}

	_, exist := filecache.GetFileCacheInfo(origins)
	if !exist {
		logrus.Debug("DelFileCacheInfo:key origin in fileCacheMap not found,do nothing")
	} else {
		logrus.Debug("DelFileCacheInfo:key origin in fileCacheMap deleting")
		sort.Strings(origins)
		value:=strings.Join(origins,",")
		//delete(filecache.FileCacheMap,value)
		filecache.FileCacheMap.FileCacheInfolrucache.Remove(value)
	}

	return true,nil
}

func checkFileCacheInfo(origins []string,filecacheinfo FileCacheInfo)bool{

	if origins == nil || len(origins) == 0 || len(origins) == 1 {
		logrus.Error("SetFileCacheInfo: origins key is nil or length is 0 or is not complex-valued")
		return false
	}

	srcHash := strings.Replace(filecacheinfo.SrcHash, " ", "", -1)
	if srcHash == "" {
		logrus.Error("SetFileCacheInfo:srcHash is  empty")
		return false
	}
	originPaths := strings.Replace(filecacheinfo.OrigPaths, " ", "", -1)
	if originPaths == "" {
		logrus.Error("SetFileCacheInfo:originPaths is empty")
		return false
	}

	return true

}

func (filecache *FileCache) DelAll(){
	for k,_:=range filecache.SingleFileCacheMap.Copyinfolrucache.cacheMap {
		//delete(filecache.SingleFileCacheMap,k)
		filecache.SingleFileCacheMap.Copyinfolrucache.Remove(k)
	}

	for k,_:=range filecache.FileCacheMap.FileCacheInfolrucache.cacheMap{
		// delete(filecache.FileCacheMap,k)
		filecache.FileCacheMap.FileCacheInfolrucache.Remove(k)
	}

}

