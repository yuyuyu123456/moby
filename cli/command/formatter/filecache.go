/**
 * Created by zizhi.yuwenqi on 2017/5/24.
 */

package formatter

import (
	"github.com/docker/docker/api/types"

)

const(
	defaultFilecacheQuietFormat  = "{{.Orig}}"
	defaultFilecacheTableFormatWithDigest="table {{.Orig}}\t{{.FileHash}}\t{{.FileName}}\t{{.FilePath}}\t{{.LastMod}}"
        defaultFilecacheTableFormat="table {{.Orig}}\t{{.FileName}}\t{{.FilePath}}\t{{.LastMod}}"
	filecacheOrigHeader="Orig"
	filecacheFileNameHeader="FileName"
	filecacheFilePathHeader="FilePath"
	filecacheLastModHeader="LastMod"
	filecacheFileHashHeader="FileHash"
)

// NewfileCacheFormat returns a format for rendering an FileCacheContext
func NewFileCacheFormat(source string, quiet bool, digest bool) Format {
	switch source {
	case TableFormatKey:
		switch {
		case quiet:
			return defaultFilecacheQuietFormat
		case digest:
			return defaultFilecacheTableFormatWithDigest
		default:
			return defaultFilecacheTableFormat
		}
//	case RawFormatKey:
//		switch {
//		case quiet:
//			return `image_id: {{.ID}}`
//		case digest:
//			return `repository: {{ .Repository }}
//tag: {{.Tag}}
//digest: {{.Digest}}
//image_id: {{.ID}}
//created_at: {{.CreatedAt}}
//virtual_size: {{.Size}}
//`
//		default:
//			return `repository: {{ .Repository }}
//tag: {{.Tag}}
//image_id: {{.ID}}
//created_at: {{.CreatedAt}}
//virtual_size: {{.Size}}
//`
//		}
	}

	format := Format(source)
	if format.IsTable() && digest && !format.Contains("{{.FileHash}}") {
		format += "\t{{.FileHash}}"
	}
	return format
}
type FileCacheContext struct {
	Context
	Digest bool
}
type filecacheContext struct {
	HeaderContext
	f     types.FileCacheSummary
}

func newFilecacheContext() *filecacheContext{
	filecacheCtx := filecacheContext{}
	filecacheCtx.header = map[string]string{
		"Orig":           filecacheOrigHeader,
		"FileName":       filecacheFileNameHeader,
		"FilePath":       filecacheFilePathHeader,
		"FileHash":       filecacheFileHashHeader,
		"LastMod":        filecacheLastModHeader,
	}
	return &filecacheCtx
}
func (f *filecacheContext) MarshalJSON() ([]byte, error) {
	return marshalJSON(f)
}
// FileCacheWrite writes the formatter filecaches using the FileCacheContext
func FileCacheWrite(ctx FileCacheContext, filecaches []types.FileCacheSummary) error {
	render := func(format func(subContext subContext) error) error {
		for _, filecache := range filecaches {
			err := format(&filecacheContext{f: filecache})
			if err != nil {
				return err
			}
		}
		return nil
	}
	return ctx.Write(newFilecacheContext(), render)
}

func (f *filecacheContext) Orig() string {
	return f.f.Orig
}
func (f *filecacheContext) FileName() string{
	return  f.f.FileName
}
func (f *filecacheContext) FilePath()string{
	return f.f.FilePath
}
func (f *filecacheContext)FileHash()string{
	return f.f.FileHash
}
func (f *filecacheContext)LastMod()string{
	if f.f.LastMod==""{
		return "<NONE>"
	}
	return f.f.LastMod
}
