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
	resp.WriteHeader(http.StatusOK, "application/json")
	body, _ := json.Marshal(result.body)
	resp.Out.Write(body)
}
