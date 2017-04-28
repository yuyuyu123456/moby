package client

import (
	"encoding/json"
	"net/url"

	"moby/api/types"
	"moby/api/types/filters"
	"moby/api/types/swarm"
	"golang.org/x/net/context"
)

// TaskList returns the list of tasks.
func (cli *Client) TaskList(ctx context.Context, options types.TaskListOptions) ([]swarm.Task, error) {
	query := url.Values{}

	if options.Filters.Len() > 0 {
		filterJSON, err := filters.ToParam(options.Filters)
		if err != nil {
			return nil, err
		}

		query.Set("filters", filterJSON)
	}

	resp, err := cli.get(ctx, "/tasks", query, nil)
	if err != nil {
		return nil, err
	}

	var tasks []swarm.Task
	err = json.NewDecoder(resp.body).Decode(&tasks)
	ensureReaderClosed(resp)
	return tasks, err
}
