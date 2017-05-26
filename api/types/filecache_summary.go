/**
 * Created by zizhi.yuwenqi on 2017/5/23.
 */

package types

type FileCacheSummary struct {

	Orig string `json:"Orig"`

	FileHash string `json:"FileHash"`

	FileName string `json:"FileName"`

        FilePath string `json:"FilePath"`

	LastMod string `json:"LastMod"`

	JsonFileName string `json:"JsonFileName"`
}

