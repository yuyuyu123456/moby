/**
 * Created by zizhi.yuwenqi on 2017/5/22.
 */

package filecache
import (
	"net/http"
	"github.com/docker/docker/api/server/httputils"
	"github.com/docker/docker/api/types/filters"
	"golang.org/x/net/context"
	"strings"
	"fmt"
)
func (s *filecacheRouter) getFilecachesJSON(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	filecacheFilters, err := filters.FromParam(r.Form.Get("filters"))
	if err != nil {
		return err
	}

	filterParam := r.Form.Get("filter")

	if filterParam != "" {
		filecacheFilters.Add("reference", filterParam)
	}

	filecaches, err := s.backend.FileCaches(filecacheFilters,false)
	if err != nil {
		return err
	}

	return httputils.WriteJSON(w, http.StatusOK, filecaches)
}

func (s *filecacheRouter) deleteFileCaches(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	name := vars["Orig"]

	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("filecache Orig name cannot be blank")
	}


	list, err := s.backend.FileCacheDelete(name)
	if err != nil {
		return err
	}

	return httputils.WriteJSON(w, http.StatusOK, list)
}