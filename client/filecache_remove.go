/**
 * Created by zizhi.yuwenqi on 2017/5/25.
 */

package client

import (
	"net/url"
	"github.com/docker/docker/api/types"
	"golang.org/x/net/context"
	"encoding/json"
)
// ImageRemove removes an image from the docker host.
func (cli *Client) FileCacheRemove(ctx context.Context,filecache string)(deleteresponses []types.FileCacheDeleteResponseItem, err error) {
	query := url.Values{}

	resp, err := cli.delete(ctx, "/filecaches/"+filecache, query, nil)
	if err != nil {
		return
	}
	err = json.NewDecoder(resp.body).Decode(&deleteresponses)
	ensureReaderClosed(resp)
	return
}
