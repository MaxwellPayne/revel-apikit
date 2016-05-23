package apikit
import (
	"testing"
	"github.com/revel/revel"
	reveltest "github.com/revel/revel/testing"
	"errors"
	"encoding/json"
	"net/http"
)

const (
	// the panic filter message stored in app.conf
	apiPanicFilterMessage string = "Oh no, we blew it. Here's an internal server error."
)

type PanicFilterTestController struct {
	*revel.Controller
}

func (c *PanicFilterTestController) CausePanic() revel.Result {
	panic(errors.New("Well, you asked for it"))
}

func TestAPIPanicFilter(t *testing.T) {
	suite := reveltest.NewTestSuite()
	serverErrMsg := revel.Config.StringDefault("apikit.internalservererror", "")
	suite.AssertEqual(serverErrMsg, apiPanicFilterMessage)

	suite.Get("/causepanic")
	suite.AssertStatus(http.StatusInternalServerError)

	result := ApiMessage{}
	err := json.Unmarshal(suite.ResponseBody, &result)
	suite.Assert(err == nil)
	suite.AssertEqual(result.Message, serverErrMsg)
}