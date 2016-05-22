package apikit
import (
	"testing"
	reveltest "github.com/revel/revel/testing"
	"github.com/revel/revel"
	"net/http"
	"fmt"
	"encoding/json"
)

var _ = fmt.Println

// Controller for ExampleUsers
type ExampleUserController struct {
	*revel.Controller
	RESTController
}

func (c *ExampleUserController) GetModel() RESTObject {
	return &ExampleUser{}
}

// Controller that does not conform to ModelProvider interface
type NonModelProviderConformingController struct {
	*revel.Controller
	RESTController
}

// Controller that incorrectly conforms to ModelProvider interface
type BadModelProviderController struct {
	*revel.Controller
	RESTController
}


func TestInstantiateControllers(t *testing.T) {
	suite := reveltest.NewTestSuite()
	suite.Get("/user")
	suite.AssertOk()

	asMap := make(map[string]interface{})
	err := json.Unmarshal(suite.ResponseBody, &asMap)
	suite.Assert(err == nil)
	_, ok := asMap["id"]
	suite.Assert(ok)
	_, ok = asMap["username"]
	suite.Assert(ok)

	suite.Get("/badmodelprovider")
	suite.AssertStatus(http.StatusInternalServerError)
}
