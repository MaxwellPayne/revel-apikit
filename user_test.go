package apikit
import (
	"github.com/revel/revel"
	reveltest "github.com/revel/revel/testing"
	"testing"
	"encoding/json"
	"bytes"
	"fmt"
	"net/http"
	"net"
	"strconv"
)

var _ = fmt.Println

type ExampleUser struct {
	ID            uint64    `json:"id"`
	Username      string    `json:"username"`
	FavoriteColor string    `json:"favorite_color"`
	Password      string    `json:"-"`
}

func (u *ExampleUser) CanBeViewedBy(other User) bool {
	return true
}

func (u *ExampleUser) CanBeModifiedBy(other User) bool {
	return other != nil && u.UserID() == other.UserID()
}

func (u *ExampleUser) Validate(v *revel.Validation) {

}

func (u *ExampleUser) UserID() uint64 {
	return u.ID
}

func (u *ExampleUser) Delete() error {
	return nil
}

func (u *ExampleUser) Save() error {
	return nil
}

// Controller for ExampleUsers
type ExampleUserController struct {
	*revel.Controller
	GenericRESTController
}

func (c *ExampleUserController) ModelFactory() RESTObject {
	return &ExampleUser{}
}

func (c *ExampleUserController) GetModelByID(id uint64) RESTObject {
	for _, u := range usersDB {
		if u.ID == id {
			return u
		}
	}
	return nil
}

func (c *ExampleUserController) EnableGET() bool {
	return true
}

func (c *ExampleUserController) EnablePOST() bool {
	return true
}

func (c *ExampleUserController) EnablePUT() bool {
	return true
}

func (c *ExampleUserController) EnableDELETE() bool {
	return true
}

var usersDB []*ExampleUser = []*ExampleUser{
	&ExampleUser{
		ID: 1,
		Username: "MaxwellPayne",
		FavoriteColor: "Red",
		Password: "banana",
	},
	&ExampleUser{
		ID: 2,
		Username: "SmokeyTheBear",
		FavoriteColor: "Blue",
		Password: "orange",
	},
}

func TestGetExampleUser(t *testing.T) {
	mockUser := usersDB[0]

	suite := reveltest.NewTestSuite()
	suite.Get(fmt.Sprint("/user/", mockUser.ID))
	suite.AssertOk()

	u := ExampleUser{}
	err := json.Unmarshal(suite.ResponseBody, &u)
	suite.Assert(err == nil)
	suite.Assert(u.ID == mockUser.ID)
	suite.Assert(u.Username == mockUser.Username)

	suite.Get(fmt.Sprint("/user/", 1234))
	suite.AssertStatus(http.StatusNotFound)
}

func TestPostExampleUser(t *testing.T) {
	mockUser := usersDB[0]

	newUserData, _ := json.Marshal(mockUser)
	suite := reveltest.NewTestSuite()
	suite.Post("/user", "application/json", bytes.NewReader(newUserData))
	suite.AssertOk()

	u := ExampleUser{}
	err := json.Unmarshal(suite.ResponseBody, &u)
	suite.Assert(err == nil)
	suite.Assert(u.ID == mockUser.ID)
	suite.Assert(u.Username == mockUser.Username)
}

func TestPutExampleUser(t *testing.T) {
	endpoint := "/user"
	putUrl := "http://" + net.JoinHostPort("localhost", strconv.Itoa(testPort)) + endpoint
	me := usersDB[0]
	me.FavoriteColor = "Purple"

	modifiedUserData, _ := json.Marshal(me)
	suite := reveltest.NewTestSuite()

	// should fail without authentication
	suite.Put(endpoint, "application/json", bytes.NewReader(modifiedUserData))
	suite.AssertStatus(http.StatusUnauthorized)

	// should fail with wrong person's authentication
	somebodyElse := usersDB[1]
	req := suite.PutCustom(putUrl, "application/json", bytes.NewReader(modifiedUserData))
	req.SetBasicAuth(somebodyElse.Username, somebodyElse.Password)
	req.MakeRequest()
	suite.AssertStatus(http.StatusUnauthorized)

	// should succeed with authentication
	req = suite.PutCustom(putUrl, "application/json", bytes.NewReader(modifiedUserData))
	req.SetBasicAuth(me.Username, me.Password)
	req.MakeRequest()
	suite.AssertOk()

	updatedUser := ExampleUser{}
	err := json.Unmarshal(suite.ResponseBody, &updatedUser)
	suite.Assert(err == nil)
	suite.Assert(me.ID == updatedUser.ID)
	suite.Assert(me.FavoriteColor == updatedUser.FavoriteColor)
}

func TestDeleteExampleUser(t *testing.T) {
	me := usersDB[0]
	endpoint := fmt.Sprint("/user/", me.ID)
	//deleteUrl := "http://" + net.JoinHostPort("localhost", strconv.Itoa(testPort)) + endpoint

	suite := reveltest.NewTestSuite()
	suite.Delete(endpoint)
	suite.AssertStatus(http.StatusUnauthorized)
}

func TestGetCustomMethod(t *testing.T) {
	suite := reveltest.NewTestSuite()
	suite.Get("/userscustomroute")
	suite.AssertStatus(http.StatusNotFound)
}