package apikit
import (
	"github.com/revel/revel"
	"encoding/json"
)

// A revel.Result renderable object used to convey a status code and error message
type ApiMessage struct {
	StatusCode int    `json:"code"`
	Message    string `json:"message"`
}

func (msg ApiMessage) Apply(req *revel.Request, resp *revel.Response) {
	resp.WriteHeader(msg.StatusCode, "application/json")
	body, _ := json.Marshal(&msg)
	resp.Out.Write(body)
}
