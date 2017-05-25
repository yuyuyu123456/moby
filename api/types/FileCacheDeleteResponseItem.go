/**
 * Created by zizhi.yuwenqi on 2017/5/25.
 */

package types

type FileCacheDeleteResponseItem struct {

	// The filecache orig  that was deleted
	Orig string `json:"Orig,omitempty"`

	//if  the filecache orig  is not exist,return notexist
	Notexist string `json:"Notexist,omitempty"`
}
