package dockerfile

// internals for handling commands. Covers many areas and a lot of
// non-contiguous functionality. Please read the comments.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/backend"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/builder"
	"github.com/docker/docker/builder/dockerfile/parser"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/httputils"
	"github.com/docker/docker/pkg/ioutils"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/progress"
	"github.com/docker/docker/pkg/streamformatter"
	"github.com/docker/docker/pkg/stringid"
	"github.com/docker/docker/pkg/system"
	"github.com/docker/docker/pkg/tarsum"
	"github.com/docker/docker/pkg/urlutil"
	"github.com/pkg/errors"
	//"docker/vendor/github.com/docker/swarmkit/manager/orchestrator"
)

func (b *Builder) commit(id string, autoCmd strslice.StrSlice, comment string) error {
	if b.disableCommit {
		return nil
	}
	if !b.hasFromImage() {
		return errors.New("Please provide a source image with `from` prior to commit")
	}
	b.runConfig.Image = b.image

	if id == "" {
		cmd := b.runConfig.Cmd
		b.runConfig.Cmd = strslice.StrSlice(append(getShell(b.runConfig), "#(nop) ", comment))
		defer func(cmd strslice.StrSlice) { b.runConfig.Cmd = cmd }(cmd)

		hit, err := b.probeCache()
		if err != nil {
			return err
		} else if hit {
			return nil
		}
		id, err = b.create()
		if err != nil {
			return err
		}
	}

	// Note: Actually copy the struct
	autoConfig := *b.runConfig
	autoConfig.Cmd = autoCmd

	commitCfg := &backend.ContainerCommitConfig{
		ContainerCommitConfig: types.ContainerCommitConfig{
			Author: b.maintainer,
			Pause:  true,
			Config: &autoConfig,
		},
	}

	// Commit the container
	imageID, err := b.docker.Commit(id, commitCfg)
	if err != nil {
		return err
	}

	b.image = imageID
	b.imageContexts.update(imageID, &autoConfig)
	return nil
}

type copyInfo struct {
	builder.FileInfo
	decompress bool
}

type fileCacheInter interface{
     getCopyInfo(origins string) ([]copyInfo,bool)
     setCopyInfo(origins string,copyinfo []copyInfo)(bool,error)
     delCopyInfo(origins []string)(bool,error)
     getFileCacheInfo(origins []string)(fileCacheInfo,bool)
     setFileCacheInfo(origins []string,filecacheinfo fileCacheInfo)(bool,error)
     delFileCacheInfo(origins []string)	(bool,error)
     delAll()
}

var fileca=&fileCache{
	singlefileCacheMap:make(map[string][]copyInfo),
	fileCacheMap:make(map[string]fileCacheInfo),
}
type fileCacheInfo struct {
	infos []copyInfo
        srcHash string
	origPaths string
}

type fileCache struct {
	singlefileCacheMap map[string][]copyInfo
	fileCacheMap map[string]fileCacheInfo
}

func (filecache *fileCache)getCopyInfo(origins string)([]copyInfo,bool)  {

	if  len(origins)==0{
		logrus.Error("getCopyInfo: origins key is nil or length is 0")
		return nil,false
	}

	if filecache.singlefileCacheMap==nil{
		logrus.Error("singlefileCacheMap is not initialized")
		logrus.Debug("initializing singlefileCacheMap")
		filecache.singlefileCacheMap=make(map[string][]copyInfo)
		return nil,false
	}

	copyinfos,exist:=filecache.singlefileCacheMap[origins]
	if !exist{
		logrus.Debug("do not find copyinfos")
		return nil,false
	}
	return copyinfos,true
}


func (filecache *fileCache)setCopyInfo(origins string,copyinfo []copyInfo)(bool,error){

	//if !checkFileCacheInfo(origins,filecacheinfo){
	//	return false,errors.New("filecacheinfo error")
	//}
	if len(copyinfo)==0{
		return false,errors.New("copyinfo is empty")
	}

	filecache.singlefileCacheMap[origins]=copyinfo
	return true,nil
}

func (filecache *fileCache)delCopyInfo(origins []string)(bool,error){

	if origins==nil|| len(origins)==0{
		return false,errors.New("key args is nil")
	}
	for _,orgin:=range origins{
		_,exist:=filecache.getCopyInfo(orgin)
		if !exist{
			logrus.Debug("delCopyInfo:key origin in singlefileCacheMap not found,do nothing")
		}else{
			logrus.Debug("delCopyInfo:key origin in singlefileCacheMap deleting")
			delete(filecache.singlefileCacheMap,orgin)
		}
	}
	return true,nil
}

func (filecache *fileCache)getFileCacheInfo(origins []string)(fileCacheInfo,bool){
	var info fileCacheInfo
	if origins==nil || len(origins)==0||len(origins)==1{
		logrus.Error("getFileCacheInfo: origins key is nil or length is 0 oris not complex-valued")
		return info,false
	}

	sort.Strings(origins)
	for key,value:=range filecache.fileCacheMap{
		keys:=strings.Split(key,",")
		if compareSlice(origins,keys){
			return value,true
		}
	}
        logrus.Debug("getFileCacheInfo:origins key not found")
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

func (filecache *fileCache)setFileCacheInfo(origins []string,filecacheinfo fileCacheInfo)(bool,error){
        if !checkFileCacheInfo(origins,filecacheinfo){
		return false,errors.New("setFileCacheInfo failure")
	}

	sort.Strings(origins)
	key:=strings.Join(origins,",")
	filecache.fileCacheMap[key]=filecacheinfo
	return true,nil
}
func(filecache *fileCache) delFileCacheInfo(origins []string)	(bool,error)  {
	if origins==nil|| len(origins)==0||len(origins)==1 {
		return false,errors.New("key args is nil")
	}

	_, exist := filecache.getFileCacheInfo(origins)
	if !exist {
		logrus.Debug("delFileCacheInfo:key origin in fileCacheMap not found,do nothing")
	} else {
		logrus.Debug("delFileCacheInfo:key origin in fileCacheMap deleting")
		sort.Strings(origins)
		value:=strings.Join(origins,",")
		delete(filecache.fileCacheMap,value)
	}

	return true,nil
}

func checkFileCacheInfo(origins []string,filecacheinfo fileCacheInfo)bool{

	if origins == nil || len(origins) == 0 || len(origins) == 1 {
		logrus.Error("setFileCacheInfo: origins key is nil or length is 0 or is not complex-valued")
		return false
	}

	srcHash := strings.Replace(filecacheinfo.srcHash, " ", "", -1)
	if srcHash == "" {
		logrus.Error("setFileCacheInfo:srcHash is  empty")
		return false
	}
	originPaths := strings.Replace(filecacheinfo.origPaths, " ", "", -1)
	if originPaths == "" {
		logrus.Error("setFileCacheInfo:originPaths is empty")
		return false
	}

	return true

}

func (filecache *fileCache) delAll(){
      for k,_:=range filecache.singlefileCacheMap{
	      delete(filecache.singlefileCacheMap,k)
      }

      for k,_:=range filecache.fileCacheMap{
	      delete(filecache.fileCacheMap,k)
      }

}

func (b *Builder) runContextCommand(args []string, allowRemote bool, allowLocalDecompression bool, cmdName string, imageSource *imageMount) error {
	if len(args) < 2 {
		return fmt.Errorf("Invalid %s format - at least two arguments required", cmdName)
	}

	// Work in daemon-specific filepath semantics
	dest := filepath.FromSlash(args[len(args)-1]) // last one is always the dest

	b.runConfig.Image = b.image

	var infos []copyInfo

	// Loop through each src file and calculate the info we need to
	// do the copy (e.g. hash value if cached).  Don't actually do
	// the copy until we've looked at all src files
	var err error

	for _, orig := range args[0 : len(args)-1] {
		if err=handleFileInfos(orig,b,allowRemote,cmdName,allowLocalDecompression,imageSource,&infos);err!=nil{
			return err
		}
	}

	if len(infos) == 0 {
		return errors.New("No source files were specified")
	}
	if len(infos) > 1 && !strings.HasSuffix(dest, string(os.PathSeparator)) {
		return fmt.Errorf("When using %s with more than one source file, the destination must be a directory and end with a /", cmdName)
	}

	// For backwards compat, if there's just one info then use it as the
	// cache look-up string, otherwise hash 'em all into one
	origPaths,srcHash:=handleSrcHashAndOrigPaths(infos)


	cmd := b.runConfig.Cmd
	b.runConfig.Cmd = strslice.StrSlice(append(getShell(b.runConfig), fmt.Sprintf("#(nop) %s %s in %s ", cmdName, srcHash, dest)))
	defer func(cmd strslice.StrSlice) { b.runConfig.Cmd = cmd }(cmd)

	if hit, err := b.probeCache(); err != nil {
		return err
	} else if hit {
		return nil
	}

	container, err := b.docker.ContainerCreate(types.ContainerCreateConfig{
		Config: b.runConfig,
		// Set a log config to override any default value set on the daemon
		HostConfig: &container.HostConfig{LogConfig: defaultLogConfig},
	})
	if err != nil {
		return err
	}
	b.tmpContainers[container.ID] = struct{}{}

	comment := fmt.Sprintf("%s %s in %s", cmdName, origPaths, dest)

	// Twiddle the destination when it's a relative path - meaning, make it
	// relative to the WORKINGDIR
	if dest, err = normaliseDest(cmdName, b.runConfig.WorkingDir, dest); err != nil {
		return err
	}

	for _, info := range infos {
		if err := b.docker.CopyOnBuild(container.ID, dest, info.FileInfo, info.decompress); err != nil {
			return err
		}
	}

	return b.commit(container.ID, cmd, comment)
}

func handleFileInfos(orig string,b *Builder,allowRemote bool,cmdName string,allowLocalDecompression bool,imageSource *imageMount,copyinfos *[]copyInfo)error{
	// Loop through each src file and calculate the info we need to
	// do the copy (e.g. hash value if cached).  Don't actually do
	// the copy until we've looked at all src files
	var err error
	var cpinfo copyInfo
	if urlutil.IsURL(orig) {
		if !allowRemote {
			return fmt.Errorf("Source can't be a URL for %s", cmdName)
		}
		if !b.options.Usefilecache {
			cpinfo,err=b.getByDownload(orig)
			if err!=nil{
				return  err
			}

		}else{
			copyinfos,hit:=fileca.getCopyInfo(orig)
			cpinfo=copyinfos[0]
			if hit && len(copyinfos)==1{

				//if copyinfo do not have modtime,
				// use cache fileinfo without check
				//otherwise check modtime and update file
				//if !(cpinfo.ModTime().IsZero() ||cpinfo.ModTime().Equal(time.Unix(0, 0))){
				//
				//}
				if ok,err:=b.updateFile(orig,cpinfo);err!=nil{
					logrus.Debug("update file in cache fail")
					if ok{
						logrus.Debug("get the cache after update")
						cpinfos,err:=fileca.getCopyInfo(orig)
						if err==nil{
							cpinfo=cpinfos[0]
						}

					}
				}

			}
			if len(copyinfos)!=1{
				logrus.Error("the cache copyinfo is mistaken")
			}
		}

		*copyinfos = append(*copyinfos,cpinfo)
		return nil
	}
	// not a URL
	subInfos, err := b.calcCopyInfo(cmdName, orig, allowLocalDecompression, true, imageSource)
	if err != nil {
		return err
	}

	*copyinfos = append(*copyinfos, subInfos...)
	return nil
}
//download url resourses and save in the cache
func(b *Builder) getByDownload(orig string)(copyInfo,error){
	var cpinfo copyInfo
	fi, err:= b.download(orig)
	if err != nil {
		return cpinfo,err
	}
	//defer os.RemoveAll(filepath.Dir(fi.Path()))
	cpinfo=copyInfo{
		FileInfo:   fi,
		decompress: false,
	}
	logrus.Debug("setCopyInfo :saving in the cache")
	fileca.setCopyInfo(orig,[]copyInfo{cpinfo})
	return cpinfo,nil
}

func handleSrcHashAndOrigPaths(infos []copyInfo)(origPaths string,srcHash string){
	if len(infos) == 1 {
		fi := infos[0].FileInfo
		origPaths = fi.Name()
		if hfi, ok := fi.(builder.Hashed); ok {
			srcHash = hfi.Hash()
		}
		return origPaths,srcHash
	} else {
		var hashs []string
		var origs []string
		for _, info := range infos {
			fi := info.FileInfo
			origs = append(origs, fi.Name())
			if hfi, ok := fi.(builder.Hashed); ok {
				hashs = append(hashs, hfi.Hash())
			}
		}
		hasher := sha256.New()
		hasher.Write([]byte(strings.Join(hashs, ",")))
		srcHash = "multi:" + hex.EncodeToString(hasher.Sum(nil))
		origPaths = strings.Join(origs, " ")
		return origPaths,srcHash
	}
}
//if the server modified,update filecache return true
//otherwise return false
func(b *Builder) updateFile(srcURL string,cpinfo copyInfo)(bool,error){
	//// get filename from URL
	//u, err := url.Parse(srcURL)
	//if err != nil {
	//	return false,err
	//}
	//path := filepath.FromSlash(u.Path) // Ensure in platform semantics
	//if strings.HasSuffix(path, string(os.PathSeparator)) {
	//	path = path[:len(path)-1]
	//}
	//parts := strings.Split(path, string(os.PathSeparator))
	//filename := parts[len(parts)-1]
	//if filename == "" {
	//	err = fmt.Errorf("cannot determine filename from url: %s", u)
	//	return
	//}
        filename,err:=handleFileName(srcURL)
	if err!=nil{
		return false,err
	}
	if !(cpinfo.ModTime().IsZero() ||cpinfo.ModTime().Equal(time.Unix(0, 0))){
		client:=http.DefaultClient
		req,err:=http.NewRequest("GET",srcURL,nil)
		if err!=nil{
			return false,err
		}
		req.Header.Add("If-Modified-Since",cpinfo.ModTime().String())
		resp,err:=client.Do(req)
		if err!=nil{
			return false,err
		}
		if resp.StatusCode >= 400 {
			return false, fmt.Errorf("Got HTTP status code >= 400: %s", resp.Status)
		}
		if resp.StatusCode==304{
			logrus.Debug("%s server not modified",srcURL)
			return false,nil
		}
		if resp.StatusCode==200{
		    fmt.Fprintf(b.Stdout,"downloading modified file  %s and update cache",srcURL)
                    hashedfileinfo,err:=b.downloadFile(filename,resp)
		    if err!=nil{
			    return false,err
		    }
		    copyinfo:=copyInfo{
				FileInfo:   hashedfileinfo,
				decompress: false,
			}
		    fileca.setCopyInfo(srcURL,[]copyInfo{copyinfo})
		    return true,nil
		}
	}
     return false,nil
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
func (b *Builder)downloadFile (filename string,resp *http.Response)(*builder.HashedFileInfo,error){
	var hashedfileinfo *builder.HashedFileInfo
	// Prepare file in a tmp dir
	tmpDir, err := ioutils.TempDir("", "docker-remote")
	if err != nil {
		return hashedfileinfo,err
	}
	defer func() {
		if err != nil {
			os.RemoveAll(tmpDir)
		}
	}()
	tmpFileName := filepath.Join(tmpDir, filename)
	tmpFile, err := os.OpenFile(tmpFileName, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return hashedfileinfo,err
	}

	stdoutFormatter := b.Stdout.(*streamformatter.StdoutFormatter)
	progressOutput := stdoutFormatter.StreamFormatter.NewProgressOutput(stdoutFormatter.Writer, true)
	progressReader := progress.NewProgressReader(resp.Body, progressOutput, resp.ContentLength, "", "Downloading")
	// Download and dump result to tmp file
	if _, err = io.Copy(tmpFile, progressReader); err != nil {
		tmpFile.Close()
		return hashedfileinfo,err
	}
	fmt.Fprintln(b.Stdout)
	// ignoring error because the file was already opened successfully
	tmpFileSt, err := tmpFile.Stat()
	if err != nil {
		tmpFile.Close()
		return hashedfileinfo,err
	}

	// Set the mtime to the Last-Modified header value if present
	// Otherwise just remove atime and mtime
	mTime := time.Time{}

	lastMod := resp.Header.Get("Last-Modified")
	if lastMod != "" {
		// If we can't parse it then just let it default to 'zero'
		// otherwise use the parsed time value
		if parsedMTime, err := http.ParseTime(lastMod); err == nil {
			mTime = parsedMTime
		}
	}

	tmpFile.Close()

	if err = system.Chtimes(tmpFileName, mTime, mTime); err != nil {
		return hashedfileinfo,err
	}

	// Calc the checksum, even if we're using the cache
	r, err := archive.Tar(tmpFileName, archive.Uncompressed)
	if err != nil {
		return hashedfileinfo,err
	}
	tarSum, err := tarsum.NewTarSum(r, true, tarsum.Version1)
	if err != nil {
		return hashedfileinfo,err
	}
	if _, err = io.Copy(ioutil.Discard, tarSum); err != nil {
		return hashedfileinfo,err
	}
	hash := tarSum.Sum(nil)
	r.Close()
	hashedfileinfo=&builder.HashedFileInfo{FileInfo: builder.PathFileInfo{FileInfo: tmpFileSt, FilePath: tmpFileName}, FileHash: hash}
	return hashedfileinfo,nil
}

func (b *Builder) download(srcURL string) (fileinfo builder.FileInfo,err error) {
	filename,err:=handleFileName(srcURL)
	if err!=nil{
		return
	}
	// Initiate the download
	resp, err := httputils.Download(srcURL)
	if err != nil {
		return
	}

	//// Prepare file in a tmp dir
	//tmpDir, err := ioutils.TempDir("", "docker-remote")
	//if err != nil {
	//	return
	//}
	//defer func() {
	//	if err != nil {
	//		os.RemoveAll(tmpDir)
	//	}
	//}()
	//tmpFileName := filepath.Join(tmpDir, filename)
	//tmpFile, err := os.OpenFile(tmpFileName, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	//if err != nil {
	//	return
	//}
	//
	//stdoutFormatter := b.Stdout.(*streamformatter.StdoutFormatter)
	//progressOutput := stdoutFormatter.StreamFormatter.NewProgressOutput(stdoutFormatter.Writer, true)
	//progressReader := progress.NewProgressReader(resp.Body, progressOutput, resp.ContentLength, "", "Downloading")
	//// Download and dump result to tmp file
	//if _, err = io.Copy(tmpFile, progressReader); err != nil {
	//	tmpFile.Close()
	//	return
	//}
	//fmt.Fprintln(b.Stdout)
	//// ignoring error because the file was already opened successfully
	//tmpFileSt, err := tmpFile.Stat()
	//if err != nil {
	//	tmpFile.Close()
	//	return
	//}
	//
	//// Set the mtime to the Last-Modified header value if present
	//// Otherwise just remove atime and mtime
	//mTime := time.Time{}
	//
	//lastMod := resp.Header.Get("Last-Modified")
	//if lastMod != "" {
	//	// If we can't parse it then just let it default to 'zero'
	//	// otherwise use the parsed time value
	//	if parsedMTime, err := http.ParseTime(lastMod); err == nil {
	//		mTime = parsedMTime
	//	}
	//}
	//
	//tmpFile.Close()
	//
	//if err = system.Chtimes(tmpFileName, mTime, mTime); err != nil {
	//	return
	//}
	//
	//// Calc the checksum, even if we're using the cache
	//r, err := archive.Tar(tmpFileName, archive.Uncompressed)
	//if err != nil {
	//	return
	//}
	//tarSum, err := tarsum.NewTarSum(r, true, tarsum.Version1)
	//if err != nil {
	//	return
	//}
	//if _, err = io.Copy(ioutil.Discard, tarSum); err != nil {
	//	return
	//}
	//hash := tarSum.Sum(nil)
	//r.Close()
	fileinfo,err=b.downloadFile(filename,resp)
	if err!=nil{
		return
	}
	return fileinfo, nil
}

var windowsBlacklist = map[string]bool{
	"c:\\":        true,
	"c:\\windows": true,
}

func (b *Builder) calcCopyInfo(cmdName, origPath string, allowLocalDecompression, allowWildcards bool, imageSource *imageMount) ([]copyInfo, error) {

	// Work in daemon-specific OS filepath semantics
	origPath = filepath.FromSlash(origPath)
	// validate windows paths from other images
	if imageSource != nil && runtime.GOOS == "windows" {
		p := strings.ToLower(filepath.Clean(origPath))
		if !filepath.IsAbs(p) {
			if filepath.VolumeName(p) != "" {
				if p[len(p)-2:] == ":." { // case where clean returns weird c:. paths
					p = p[:len(p)-1]
				}
				p += "\\"
			} else {
				p = filepath.Join("c:\\", p)
			}
		}
		if _, blacklisted := windowsBlacklist[p]; blacklisted {
			return nil, errors.New("copy from c:\\ or c:\\windows is not allowed on windows")
		}
	}

	if origPath != "" && origPath[0] == os.PathSeparator && len(origPath) > 1 {
		origPath = origPath[1:]
	}
	origPath = strings.TrimPrefix(origPath, "."+string(os.PathSeparator))

	context := b.context
	var err error
	if imageSource != nil {
		context, err = imageSource.context()
		if err != nil {
			return nil, err
		}
	}

	if context == nil {
		return nil, errors.Errorf("No context given. Impossible to use %s", cmdName)
	}

	// Deal with wildcards
	if allowWildcards && containsWildcards(origPath) {
		var copyInfos []copyInfo
		if err := context.Walk("", func(path string, info builder.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Name() == "" {
				// Why are we doing this check?
				return nil
			}
			if match, _ := filepath.Match(origPath, path); !match {
				return nil
			}

			// Note we set allowWildcards to false in case the name has
			// a * in it
			subInfos, err := b.calcCopyInfo(cmdName, path, allowLocalDecompression, false, imageSource)
			if err != nil {
				return err
			}
			copyInfos = append(copyInfos, subInfos...)
			return nil
		}); err != nil {
			return nil, err
		}
		return copyInfos, nil
	}

	// Must be a dir or a file
	statPath, fi, err := context.Stat(origPath)
	if err != nil {
		return nil, err
	}

	copyInfos := []copyInfo{{FileInfo: fi, decompress: allowLocalDecompression}}

	hfi, handleHash := fi.(builder.Hashed)
	if !handleHash {
		return copyInfos, nil
	}
	if imageSource != nil {
		// fast-cache based on imageID
		if h, ok := b.imageContexts.getCache(imageSource.id, origPath); ok {
			hfi.SetHash(h.(string))
			return copyInfos, nil
		}
	}

	// Deal with the single file case
	if !fi.IsDir() {
		hfi.SetHash("file:" + hfi.Hash())
		return copyInfos, nil
	}
	// Must be a dir
	var subfiles []string
	err = context.Walk(statPath, func(path string, info builder.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// we already checked handleHash above
		subfiles = append(subfiles, info.(builder.Hashed).Hash())
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(subfiles)
	hasher := sha256.New()
	hasher.Write([]byte(strings.Join(subfiles, ",")))
	hfi.SetHash("dir:" + hex.EncodeToString(hasher.Sum(nil)))
	if imageSource != nil {
		b.imageContexts.setCache(imageSource.id, origPath, hfi.Hash())
	}

	return copyInfos, nil
}

func (b *Builder) processImageFrom(img builder.Image) error {
	if img != nil {
		b.image = img.ImageID()

		if img.RunConfig() != nil {
			b.runConfig = img.RunConfig()
		}
	}

	// Check to see if we have a default PATH, note that windows won't
	// have one as it's set by HCS
	if system.DefaultPathEnv != "" {
		if _, ok := b.runConfigEnvMapping()["PATH"]; !ok {
			b.runConfig.Env = append(b.runConfig.Env,
				"PATH="+system.DefaultPathEnv)
		}
	}

	if img == nil {
		// Typically this means they used "FROM scratch"
		return nil
	}

	// Process ONBUILD triggers if they exist
	if nTriggers := len(b.runConfig.OnBuild); nTriggers != 0 {
		word := "trigger"
		if nTriggers > 1 {
			word = "triggers"
		}
		fmt.Fprintf(b.Stderr, "# Executing %d build %s...\n", nTriggers, word)
	}

	// Copy the ONBUILD triggers, and remove them from the config, since the config will be committed.
	onBuildTriggers := b.runConfig.OnBuild
	b.runConfig.OnBuild = []string{}

	// Reset stdin settings as all build actions run without stdin
	b.runConfig.OpenStdin = false
	b.runConfig.StdinOnce = false

	// parse the ONBUILD triggers by invoking the parser
	for _, step := range onBuildTriggers {
		result, err := parser.Parse(strings.NewReader(step))
		if err != nil {
			return err
		}

		for _, n := range result.AST.Children {
			if err := checkDispatch(n); err != nil {
				return err
			}

			upperCasedCmd := strings.ToUpper(n.Value)
			switch upperCasedCmd {
			case "ONBUILD":
				return errors.New("Chaining ONBUILD via `ONBUILD ONBUILD` isn't allowed")
			case "MAINTAINER", "FROM":
				return errors.Errorf("%s isn't allowed as an ONBUILD trigger", upperCasedCmd)
			}
		}

		if err := dispatchFromDockerfile(b, result); err != nil {
			return err
		}
	}
	return nil
}

// probeCache checks if cache match can be found for current build instruction.
// If an image is found, probeCache returns `(true, nil)`.
// If no image is found, it returns `(false, nil)`.
// If there is any error, it returns `(false, err)`.
func (b *Builder) probeCache() (bool, error) {
	c := b.imageCache
	if c == nil || b.options.NoCache || b.cacheBusted {
		return false, nil
	}
	cache, err := c.GetCache(b.image, b.runConfig)
	if err != nil {
		return false, err
	}
	if len(cache) == 0 {
		logrus.Debugf("[BUILDER] Cache miss: %s", b.runConfig.Cmd)
		b.cacheBusted = true
		return false, nil
	}

	fmt.Fprint(b.Stdout, " ---> Using cache\n")
	logrus.Debugf("[BUILDER] Use cached version: %s", b.runConfig.Cmd)
	b.image = string(cache)
	b.imageContexts.update(b.image, b.runConfig)

	return true, nil
}

func (b *Builder) create() (string, error) {
	if !b.hasFromImage() {
		return "", errors.New("Please provide a source image with `from` prior to run")
	}
	b.runConfig.Image = b.image

	resources := container.Resources{
		CgroupParent: b.options.CgroupParent,
		CPUShares:    b.options.CPUShares,
		CPUPeriod:    b.options.CPUPeriod,
		CPUQuota:     b.options.CPUQuota,
		CpusetCpus:   b.options.CPUSetCPUs,
		CpusetMems:   b.options.CPUSetMems,
		Memory:       b.options.Memory,
		MemorySwap:   b.options.MemorySwap,
		Ulimits:      b.options.Ulimits,
	}

	// TODO: why not embed a hostconfig in builder?
	hostConfig := &container.HostConfig{
		SecurityOpt: b.options.SecurityOpt,
		Isolation:   b.options.Isolation,
		ShmSize:     b.options.ShmSize,
		Resources:   resources,
		NetworkMode: container.NetworkMode(b.options.NetworkMode),
		// Set a log config to override any default value set on the daemon
		LogConfig:  defaultLogConfig,
		ExtraHosts: b.options.ExtraHosts,
	}

	config := *b.runConfig

	// Create the container
	c, err := b.docker.ContainerCreate(types.ContainerCreateConfig{
		Config:     b.runConfig,
		HostConfig: hostConfig,
	})
	if err != nil {
		return "", err
	}
	for _, warning := range c.Warnings {
		fmt.Fprintf(b.Stdout, " ---> [Warning] %s\n", warning)
	}

	b.tmpContainers[c.ID] = struct{}{}
	fmt.Fprintf(b.Stdout, " ---> Running in %s\n", stringid.TruncateID(c.ID))

	// override the entry point that may have been picked up from the base image
	if err := b.docker.ContainerUpdateCmdOnBuild(c.ID, config.Cmd); err != nil {
		return "", err
	}

	return c.ID, nil
}

var errCancelled = errors.New("build cancelled")

func (b *Builder) run(cID string) (err error) {
	errCh := make(chan error)
	go func() {
		errCh <- b.docker.ContainerAttachRaw(cID, nil, b.Stdout, b.Stderr, true)
	}()

	finished := make(chan struct{})
	cancelErrCh := make(chan error, 1)
	go func() {
		select {
		case <-b.clientCtx.Done():
			logrus.Debugln("Build cancelled, killing and removing container:", cID)
			b.docker.ContainerKill(cID, 0)
			b.removeContainer(cID)
			cancelErrCh <- errCancelled
		case <-finished:
			cancelErrCh <- nil
		}
	}()

	if err := b.docker.ContainerStart(cID, nil, "", ""); err != nil {
		close(finished)
		if cancelErr := <-cancelErrCh; cancelErr != nil {
			logrus.Debugf("Build cancelled (%v) and got an error from ContainerStart: %v",
				cancelErr, err)
		}
		return err
	}

	// Block on reading output from container, stop on err or chan closed
	if err := <-errCh; err != nil {
		close(finished)
		if cancelErr := <-cancelErrCh; cancelErr != nil {
			logrus.Debugf("Build cancelled (%v) and got an error from errCh: %v",
				cancelErr, err)
		}
		return err
	}

	if ret, _ := b.docker.ContainerWait(cID, -1); ret != 0 {
		close(finished)
		if cancelErr := <-cancelErrCh; cancelErr != nil {
			logrus.Debugf("Build cancelled (%v) and got a non-zero code from ContainerWait: %d",
				cancelErr, ret)
		}
		// TODO: change error type, because jsonmessage.JSONError assumes HTTP
		return &jsonmessage.JSONError{
			Message: fmt.Sprintf("The command '%s' returned a non-zero code: %d", strings.Join(b.runConfig.Cmd, " "), ret),
			Code:    ret,
		}
	}
	close(finished)
	return <-cancelErrCh
}

func (b *Builder) removeContainer(c string) error {
	rmConfig := &types.ContainerRmConfig{
		ForceRemove:  true,
		RemoveVolume: true,
	}
	if err := b.docker.ContainerRm(c, rmConfig); err != nil {
		fmt.Fprintf(b.Stdout, "Error removing intermediate container %s: %v\n", stringid.TruncateID(c), err)
		return err
	}
	return nil
}

func (b *Builder) clearTmp() {
	for c := range b.tmpContainers {
		if err := b.removeContainer(c); err != nil {
			return
		}
		delete(b.tmpContainers, c)
		fmt.Fprintf(b.Stdout, "Removing intermediate container %s\n", stringid.TruncateID(c))
	}
}

// readAndParseDockerfile reads a Dockerfile from the current context.
func (b *Builder) readAndParseDockerfile() (*parser.Result, error) {
	// If no -f was specified then look for 'Dockerfile'. If we can't find
	// that then look for 'dockerfile'.  If neither are found then default
	// back to 'Dockerfile' and use that in the error message.
	if b.options.Dockerfile == "" {
		b.options.Dockerfile = builder.DefaultDockerfileName
		if _, _, err := b.context.Stat(b.options.Dockerfile); os.IsNotExist(err) {
			lowercase := strings.ToLower(b.options.Dockerfile)
			if _, _, err := b.context.Stat(lowercase); err == nil {
				b.options.Dockerfile = lowercase
			}
		}
	}

	result, err := b.parseDockerfile()
	if err != nil {
		return nil, err
	}

	// After the Dockerfile has been parsed, we need to check the .dockerignore
	// file for either "Dockerfile" or ".dockerignore", and if either are
	// present then erase them from the build context. These files should never
	// have been sent from the client but we did send them to make sure that
	// we had the Dockerfile to actually parse, and then we also need the
	// .dockerignore file to know whether either file should be removed.
	// Note that this assumes the Dockerfile has been read into memory and
	// is now safe to be removed.
	if dockerIgnore, ok := b.context.(builder.DockerIgnoreContext); ok {
		dockerIgnore.Process([]string{b.options.Dockerfile})
	}
	return result, nil
}

func (b *Builder) parseDockerfile() (*parser.Result, error) {
	f, err := b.context.Open(b.options.Dockerfile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("Cannot locate specified Dockerfile: %s", b.options.Dockerfile)
		}
		return nil, err
	}
	defer f.Close()
	if f, ok := f.(*os.File); ok {
		// ignoring error because Open already succeeded
		fi, err := f.Stat()
		if err != nil {
			return nil, fmt.Errorf("Unexpected error reading Dockerfile: %v", err)
		}
		if fi.Size() == 0 {
			return nil, fmt.Errorf("The Dockerfile (%s) cannot be empty", b.options.Dockerfile)
		}
	}
	return parser.Parse(f)
}
