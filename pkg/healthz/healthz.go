package healthz

import (
	"fmt"
	"net/http"
	"path"
)

// Handler is an http.Handler that aggregates the results of the given
// checkers to the root path, and supports calling individual checkers on
// subpaths of the name of the checker.
//
// Adding checks on the fly is *not* threadsafe -- use a wrapper.
type Handler struct {
	Checks map[string]Checker
}

// Checker knows how to perform a health check.
type Checker func(req *http.Request) error

// Ping returns true automatically when checked.
var Ping Checker = func(_ *http.Request) error { return nil }

func (h *Handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	// clean up the request (duplicating the internal logic of http.ServeMux a bit)
	// clean up the path a bit
	reqPath := req.URL.Path
	if reqPath == "" || reqPath[0] != '/' {
		reqPath = "/" + reqPath
	}
	// path.Clean removes the trailing slash except for root for us
	// (which is fine, since we're only serving one layer of sub-paths)
	reqPath = path.Clean(reqPath)

	// either serve the root endpoint...
	if reqPath == "/" {
		h.serveAggregated(resp, req)
		return
	}

	// ...the default check (if nothing else is present)...
	if len(h.Checks) == 0 && reqPath[1:] == "ping" {
		CheckHandler{Checker: Ping}.ServeHTTP(resp, req)
		return
	}

	// ...or an individual checker
	checkName := reqPath[1:] // ignore the leading slash
	checker, known := h.Checks[checkName]
	if !known {
		http.NotFoundHandler().ServeHTTP(resp, req)
		return
	}

	CheckHandler{Checker: checker}.ServeHTTP(resp, req)
}

func (h *Handler) serveAggregated(resp http.ResponseWriter, req *http.Request) {
	//failed := false
	//excluded := getExcludedChecks(req)
	//
	//parts := make([]checkStatus, 0, len(h.Checks))
	//
	//// calculate the results...
	//for checkName, check := range h.Checks {
	//	// no-op the check if we've specified we want to exclude the check
	//	if excluded.Has(checkName) {
	//		excluded.Delete(checkName)
	//		parts = append(parts, checkStatus{name: checkName, healthy: true, excluded: true})
	//		continue
	//	}
	//	if err := check(req); err != nil {
	//		log.V(1).Info("healthz check failed", "checker", checkName, "error", err)
	//		parts = append(parts, checkStatus{name: checkName, healthy: false})
	//		failed = true
	//	} else {
	//		parts = append(parts, checkStatus{name: checkName, healthy: true})
	//	}
	//}
	//
	//// ...default a check if none is present...
	//if len(h.Checks) == 0 {
	//	parts = append(parts, checkStatus{name: "ping", healthy: true})
	//}
	//
	//for _, c := range excluded.UnsortedList() {
	//	log.V(1).Info("cannot exclude health check, no matches for it", "checker", c)
	//}
	//
	//// ...sort to be consistent...
	//sort.Slice(parts, func(i, j int) bool { return parts[i].name < parts[j].name })
	//
	//// ...and write out the result
	//// TODO(directxman12): this should also accept a request for JSON content (via a accept header)
	//_, forceVerbose := req.URL.Query()["verbose"]
	//writeStatusesAsText(resp, parts, excluded, failed, forceVerbose)
}

// CheckHandler is an http.Handler that serves a health check endpoint at the root path,
// based on its checker.
type CheckHandler struct {
	Checker
}

func (h CheckHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if err := h.Checker(req); err != nil {
		http.Error(resp, fmt.Sprintf("internal server error: %v", err), http.StatusInternalServerError)
	} else {
		fmt.Fprint(resp, "ok")
	}
}
