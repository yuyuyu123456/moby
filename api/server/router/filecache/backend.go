/**
 * Created by zizhi.yuwenqi on 2017/5/25.
 */

package filecache

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

type Backend interface {
	filecacheBackend
}

type filecacheBackend interface {
	FileCaches(filecachesfilters filters.Args,withExtraAttrs bool)([]*types.FileCacheSummary,error)
	//FileCacheDelete(Orig string)(error)
}
