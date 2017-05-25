/**
 * Created by zizhi.yuwenqi on 2017/5/22.
 */

package filecache

import (
"github.com/docker/docker/api/server/router"
)

// filecacheRouter is a router to talk with the filecache controller
type filecacheRouter struct {
	backend Backend
	routes  []router.Route
}

// NewRouter initializes a new image router
func NewRouter(backend Backend) router.Router {
	r := &filecacheRouter{
		backend: backend,
	}
	r.initRoutes()
	return r
}

// Routes returns the available routes to the image controller
func (r *filecacheRouter) Routes() []router.Route {
	return r.routes
}

// initRoutes initializes the routes in the image router
func (r *filecacheRouter) initRoutes() {
	r.routes = []router.Route{
		// GET
		router.NewGetRoute("/filecache/json", r.getFilecachesJSON),

	}
}
