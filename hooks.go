package apikit
import (
	"github.com/revel/revel"
	"encoding/json"
	"net/http"
)

type GETHooker interface {
	RESTController
	PreGETHook(id uint64, authUser User) revel.Result
	PostGETHook(model RESTObject, authUser User) revel.Result
}

type POSTHooker interface {
	RESTController
	PrePOSTHook(model RESTObject, authUser User) revel.Result
	PostPOSTHook(model RESTObject, authUser User, err error) revel.Result
}

type PUTHooker interface {
	RESTController
	PrePUTHook(newInstance, existingInstance RESTObject, authUser User) revel.Result
	PostPUTHook(newInstance, existingInstance RESTObject, authUser User, err error) revel.Result
}

type DELETEHooker interface {
	RESTController
	PreDELETEHook(model RESTObject, authUser User) revel.Result
	PostDELETEHook(model RESTObject, authUser User, err error) revel.Result
}

// Result that can be returned from a hook
type HookJsonResult struct {
	Body interface{}
}

func (result HookJsonResult) Apply(req *revel.Request, resp *revel.Response) {
	if body, err := json.Marshal(result.Body); err != nil {
		ApiMessage{
			StatusCode: http.StatusInternalServerError,
			Message: "Something went wrong",
		}.Apply(req, resp)
	} else {
		resp.WriteHeader(http.StatusOK, "application/json")
		resp.Out.Write(body)
	}
}