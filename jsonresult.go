package apikit
import (
	"github.com/revel/revel"
	"encoding/json"
	"net/http"
)

type jsonResult struct {
	body interface{}
}

func (result jsonResult) Apply(req *revel.Request, resp *revel.Response) {
	if body, err := json.Marshal(result.body); err != nil {
		ApiMessage{
			StatusCode: http.StatusInternalServerError,
			Message: "Something went wrong",
		}.Apply(req, resp)
	} else {
		resp.WriteHeader(http.StatusOK, "application/json")
		resp.Out.Write(body)
	}
}
