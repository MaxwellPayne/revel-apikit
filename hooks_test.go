package apikit
import (
	"github.com/revel/revel"
	reveltest "github.com/revel/revel/testing"
	"errors"
	"net/http"
	"testing"
	"fmt"
	"encoding/json"
)

type Fish struct {
	ID         uint64       `json:"id"`
	FinCount   int          `json:"fin_count"`
	Color      string       `json:"color"`
	IsImmortal bool         `json:"is_immortal"`
	Owner      *ExampleUser `json:"owner"`
}

func (fish *Fish) CanBeViewedBy(user User) bool {
	return true
}

func (fish *Fish) CanBeModifiedBy(user User) bool {
	return true
}

func (fish *Fish) Validate(v *revel.Validation) {
	v.Min(fish.FinCount, 2).Message("Fish must have at least 2 fins")
}

func (fish *Fish) Delete() error {
	return nil
}

func (fish *Fish) Save() error {
	v := new(revel.Validation)
	fish.Validate(v)
	if v.HasErrors() {
		return errors.New(v.Errors[0].String())
	}
	return nil
}

const (
	fishDeleteFailureMessage = "Foolish mortal, you cannot kill an immortal fish."
	fishDeleteSuccessMessage = "Uh oh, owner. Looks like you killed your own fish."
)

type FishHookerController struct {
	*revel.Controller
	GenericRESTController
}

func (c *FishHookerController) ModelFactory() RESTObject {
	return &Fish{}
}

func (c *FishHookerController) GetModelByID(id uint64) RESTObject {
	for _, f := range pond {
		if f.ID == id {
			return &f
		}
	}
	return nil
}

func (c *FishHookerController) EnableGET() bool {
	return true
}

func (c *FishHookerController) EnablePOST() bool {
	return true
}

func (c *FishHookerController) EnablePUT() bool {
	return true
}

func (c *FishHookerController) EnableDELETE() bool {
	return true
}

func (c *FishHookerController) PreDELETEHook(model RESTObject, authUser User) revel.Result {
	fish := model.(*Fish)
	if fish.IsImmortal {
		return ApiMessage{
			StatusCode: http.StatusUnauthorized,
			Message: fishDeleteFailureMessage,
		}
	}
	return nil
}

func (c *FishHookerController) PostDELETEHook(model RESTObject, authUser User, err error) revel.Result {
	if owner := model.(*Fish).Owner; owner != nil {
		return ApiMessage{
			StatusCode: http.StatusOK,
			Message: fishDeleteSuccessMessage,
		}
	}
	return nil
}

var pond []Fish = []Fish{
	Fish{
		ID: 7,
		FinCount: 2,
		Color: "Red",
		IsImmortal: false,
		Owner: usersDB[0],
	},
	Fish{
		ID: 8,
		FinCount: 1200,
		Color: "Rainbow",
		IsImmortal: true,
		Owner: usersDB[0],
	},
}

func TestPreDELETEHook(t *testing.T) {
	fish := pond[1]
	endpoint := fmt.Sprint("/fish/", fish.ID)

	suite := reveltest.NewTestSuite()
	suite.Assert(fish.IsImmortal)

	suite.Delete(endpoint)
	suite.AssertStatus(http.StatusUnauthorized)

	msg := ApiMessage{}
	err := json.Unmarshal(suite.ResponseBody, &msg)
	suite.Assert(err == nil)
	suite.AssertEqual(msg.Message, fishDeleteFailureMessage)


}

func TestPostDELETEHook(t *testing.T) {
	fish := pond[0]
	endpoint := fmt.Sprint("/fish/", fish.ID)
	suite := reveltest.NewTestSuite()
	suite.Assert(!fish.IsImmortal)

	suite.Delete(endpoint)
	suite.AssertOk()

	msg := ApiMessage{}
	err := json.Unmarshal(suite.ResponseBody, &msg)
	suite.Assert(err == nil)
	suite.AssertEqual(msg.Message, fishDeleteSuccessMessage)
}
