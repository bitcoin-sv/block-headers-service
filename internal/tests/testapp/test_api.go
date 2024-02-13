package testapp

import (
	"net/http"
	"net/http/httptest"
)

// Api exposes functions to easy testing of block headers service endpoints.
type Api struct {
	*TestBlockHeaderService
}

// Call Send request to block headers service api
// req is a Http request that should be made against block headers service server
// errors can be used to provide a way to call this method with the result of http.NewRequest
//
// Example 1:
//
//	api.Call(http.NewRequest(http.MethodGet, "/some/url", nil))
//
// Example 2:
//
//	 req, err = http.NewRequest(http.MethodDelete, prefix+"/access/"+*tokenToDelete, nil)
//		if err != nil {
//			//handle error
//		}
//		api.Call(req)
func (api *Api) Call(req *http.Request, errors ...error) *httptest.ResponseRecorder {
	api.handleErrorsIfPassed(errors)
	res := httptest.NewRecorder()
	api.engine.ServeHTTP(res, req)
	return res
}

func (api *Api) handleErrorsIfPassed(errors []error) {
	if len(errors) == 0 {
		return
	}

	haveNotNilError := false
	for _, err := range errors {
		if err != nil {
			haveNotNilError = true
			break
		}
	}

	if haveNotNilError {
		api.t.Fatalf("%v", errors)
	}
}
