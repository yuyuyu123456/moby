/**
 * Created by zizhi.yuwenqi on 2017/5/25.
 */

package client

import (
	"net/url"
	"github.com/docker/docker/api/types"
	"golang.org/x/net/context"
	"encoding/json"
	"github.com/Sirupsen/logrus"
)
// ImageRemove removes an image from the docker host.
func (cli *Client) FileCacheRemove(ctx context.Context,filecache string)(deleteresponses []types.FileCacheDeleteResponseItem, err error) {
	query := url.Values{}

	resp, err := cli.delete(ctx, "/filecaches/"+filecache, query, nil)
	if err != nil {
		logrus.Debug("FileCacheRemove get response error:",err)
		return
	}
	err = json.NewDecoder(resp.body).Decode(&deleteresponses)
	ensureReaderClosed(resp)
	return
}
