package server

import (
	"net/http"

	"github.com/rancher/naok/pkg/attributes"

	"github.com/gorilla/mux"
	"github.com/rancher/norman/pkg/types"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type APIFunc func(*types.APIRequest)

func (a *apiServer) routes() {
	a.Path("/api/{version}/{resource}").Handler(a.handle(a.k8sAPI))
	a.Path("/api/{version}/{resource}/{nameorns}").Handler(a.handle(a.k8sAPI))
	a.Path("/api/{version}/{resource}/{namespace}/{name}").Handler(a.handle(a.k8sAPI))

	a.Path("/apis/{group}/{version}/{resource}").Handler(a.handle(a.k8sAPI))
	a.Path("/apis/{group}/{version}/{resource}/{nameorns}").Handler(a.handle(a.k8sAPI))
	a.Path("/apis/{group}/{version}/{resource}/{namespace}/{name}").Handler(a.handle(a.k8sAPI))

	a.Path("/v1/{type}").Handler(a.handle(nil))
	a.Path("/v1/{type}/{name}").Handler(a.handle(nil))
}

func (a *apiServer) handle(apiFunc APIFunc) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		a.api(rw, req, apiFunc)
	})
}

func (a *apiServer) api(rw http.ResponseWriter, req *http.Request, apiFunc APIFunc) {
	apiOp, ok := a.common(rw, req)
	if ok {
		if apiFunc != nil {
			apiFunc(apiOp)
		}
		a.server.Handle(apiOp)
	}
}

func (a *apiServer) k8sAPI(apiOp *types.APIRequest) {
	vars := mux.Vars(apiOp.Request)
	apiOp.Name = vars["name"]
	apiOp.Type = a.sf.ByGVR(schema.GroupVersionResource{
		Version:  vars["version"],
		Group:    vars["group"],
		Resource: vars["resource"],
	})

	nOrN := vars["nameorns"]
	if nOrN != "" {
		schema := apiOp.Schemas.Schema(apiOp.Type)
		if attributes.Namespaced(schema) {
			vars["namespace"] = nOrN
		} else {
			vars["name"] = nOrN
		}
	}

	if namespace := vars["namespace"]; namespace != "" {
		apiOp.Namespaces = []string{namespace}
	}
}