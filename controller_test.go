package apikit
import (
	"testing"
	reveltest "github.com/revel/revel/testing"
	"github.com/revel/revel"
	"net/http"
	"fmt"
	"strings"
)

var _ = fmt.Println

// Controller that does not conform to ModelProvider interface
type NonModelProviderConformingController struct {
	*revel.Controller
	RESTController
}


func TestNonConformingController(t *testing.T) {
	suite := reveltest.NewTestSuite()
	suite.Post("/badmodelprovider", "application/json", strings.NewReader("{}"))
	suite.AssertStatus(http.StatusInternalServerError)
}
