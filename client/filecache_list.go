/**
 * Created by zizhi.yuwenqi on 2017/5/23.
 */

package client
import (
	//"encoding/json"
	//"net/url"

	"github.com/docker/docker/api/types"
	//"github.com/docker/docker/api/types/filters"
	//"github.com/docker/docker/api/types/versions"
	"golang.org/x/net/context"
)

// FilecahceList returns a list of filecaches in the docker host.
func (cli *Client) FileCacheList(ctx context.Context, options types.FileCachesOptions) (filecaches []types.FileCacheSummary, err error) {
	//var images []types.ImageSummary
	//query := url.Values{}
	//
	//optionFilters := options.Filters
	//referenceFilters := optionFilters.Get("reference")
	//if versions.LessThan(cli.version, "1.25") && len(referenceFilters) > 0 {
	//	query.Set("filter", referenceFilters[0])
	//	for _, filterValue := range referenceFilters {
	//		optionFilters.Del("reference", filterValue)
	//	}
	//}
	//if optionFilters.Len() > 0 {
	//	filterJSON, err := filters.ToParamWithVersion(cli.version, optionFilters)
	//	if err != nil {
	//		return images, err
	//	}
	//	query.Set("filters", filterJSON)
	//}
	//
	//serverResp, err := cli.get(ctx, "/images/json", query, nil)
	//if err != nil {
	//	return images, err
	//}
	//
	//err = json.NewDecoder(serverResp.body).Decode(&images)
	//ensureReaderClosed(serverResp)
	//return images, err
	filecaches=types.FileCacheSummary{
		Orig:"dir/",
		FileHash:"yuwneiq",
		FilePath:"/var/lib/docker",
		FileName:"dir/aaa",
		LastMod:"2017",
	}
	return
}

