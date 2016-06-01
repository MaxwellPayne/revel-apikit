package apikit
import (
	"github.com/revel/revel"
	reveltest "github.com/revel/revel/testing"
	"errors"
	"net/http"
	"testing"
	"fmt"
	"encoding/json"
	"net"
	"strconv"
	"bytes"
	"time"
	"sync"
)

type Fish struct {
	ID         uint64       `json:"id"`
	CreateDate time.Time    `apikit:"immutable"`
	FinCount   int          `json:"fin_count"`
	Color      string       `json:"color"`
	IsImmortal bool         `json:"is_immortal"`
	Owner      *ExampleUser `json:"owner"`
}

func (fish *Fish) CanBeViewedBy(user User) bool {
	return true
}

func (fish *Fish) CanBeCreatedBy(user User) bool {
	return true
}

func (fish *Fish) CanBeDeletedBy(user User) bool {
	return true
}

func (fish *Fish) CanBeModifiedBy(user User) bool {
	return true
}

func (fish *Fish) UniqueID() uint64 {
	return fish.ID
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
	// Use these constants for GETHooker tests
	luckyFishID = 23
	luckyFishMessage = "Hey look! You got the lucky fish."

	// Use these constants for DELETEHooker tests
	fishDeleteFailureMessage = "Foolish mortal, you cannot kill an immortal fish."
	fishDeleteSuccessMessage = "Uh oh, owner. Looks like you killed your own fish."
)

var (
	// Use these for POSTHooker tests
	prePOSTHookerChan = make(chan bool)
	postPOSTHookerChan = make(chan bool)

	// Use these for PUTHooker tests
	prePUTHookerChan = make(chan bool)
	postPUTHookerChan = make(chan bool)
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

// GETHooker interface implementation
func (c *FishHookerController) PreGETHook(id uint64, authUser User) revel.Result {
	if id == luckyFishID {
		return ApiMessage{
			StatusCode: http.StatusOK,
			Message: luckyFishMessage,
		}
	}
	return nil
}

func (c *FishHookerController) PostGETHook(model RESTObject, authUser User) revel.Result {
	if authUser != nil {
		return ApiMessage{
			StatusCode: http.StatusTeapot,
		}
	}
	return nil
}

// POSTHooker interface implementation
func (c *FishHookerController) PrePOSTHook(model RESTObject, authUser User) revel.Result {
	go func() {
		prePOSTHookerChan <- true
	}()
	return nil
}

func (c *FishHookerController) PostPOSTHook(model RESTObject, authUser User, err error) revel.Result {
	go func() {
		postPOSTHookerChan <- true
	}()
	return nil
}

// PUTHooker interface implementation
func (c *FishHookerController) PrePUTHook(newInstance, existingInstance RESTObject, authUser User) revel.Result {
	go func() {
		prePUTHookerChan <- true
	}()
	return nil
}

func (c *FishHookerController) PostPUTHook(newInstance, existingInstance RESTObject, authUser User, err error) revel.Result {
	go func() {
		postPUTHookerChan <- true
	}()
	return nil
}

// DELETEHooker interface implementation
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
		CreateDate: time.Now(),
	},
	Fish{
		ID: 8,
		FinCount: 1200,
		Color: "Rainbow",
		IsImmortal: true,
		Owner: usersDB[0],
		CreateDate: time.Now(),
	},
}

func TestPreGETHook(t *testing.T) {
	endpoint := fmt.Sprint("/fish/", luckyFishID)
	suite := reveltest.NewTestSuite()
	suite.Get(endpoint)
	suite.AssertOk()

	msg := ApiMessage{}
	err := json.Unmarshal(suite.ResponseBody, &msg)
	suite.Assert(err == nil)
	suite.AssertEqual(msg.Message, luckyFishMessage)

	// now get a non-lucky fish, hook should have no effect here
	regularFish := pond[1]
	endpoint = fmt.Sprint("/fish/", regularFish.ID)
	suite.Get(endpoint)
	suite.AssertOk()
	result := Fish{}
	err = json.Unmarshal(suite.ResponseBody, &result)
	suite.Assert(err == nil)
	suite.AssertEqual(regularFish.Color, result.Color)
}

func TestPostGETHook(t *testing.T) {
	suite := reveltest.NewTestSuite()
	user := usersDB[0]
	username, password := user.Username, user.Password

	// PostGETHook + authUser should trigger a Teapot status
	endpoint := fmt.Sprint("/fish/", pond[0].ID)
	url := "http://" + net.JoinHostPort("localhost", strconv.Itoa(testPort)) + endpoint
	req := suite.GetCustom(url)
	req.SetBasicAuth(username, password)
	req.MakeRequest()
	suite.AssertStatus(http.StatusTeapot)

	// PostGETHook should never be called when RESTObject does not exist
	badFishId := 12345
	endpoint = fmt.Sprint("/fish/", badFishId)
	url = "http://" + net.JoinHostPort("localhost", strconv.Itoa(testPort)) + endpoint
	req = suite.GetCustom(url)
	req.SetBasicAuth(username, password)
	req.MakeRequest()
	suite.AssertStatus(http.StatusNotFound)

	// Make sure this 404 did not come from a routing error
	suite.AssertContains(strconv.Itoa(badFishId))
}

func TestPrePOSTHook(t *testing.T) {
	endpoint := "/fish"
	suite := reveltest.NewTestSuite()
	body, _ := json.Marshal(pond[0])
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		select {
		case <- prePOSTHookerChan:
			break
		case <- time.After(time.Second * 1):
			t.Error("PrePOSTHook did not trigger")
		}
		wg.Done()
	}()

	suite.Post(endpoint, "application/json", bytes.NewReader(body))
	wg.Wait()
	suite.AssertOk()
}

func TestPostPOSTHook(t *testing.T) {
	endpoint := "/fish"
	suite := reveltest.NewTestSuite()
	invalidFish := Fish{
		FinCount: 1,
	}
	body, _ := json.Marshal(&invalidFish)
	suite.Post(endpoint, "application/json", bytes.NewReader(body))
	suite.AssertStatus(http.StatusBadRequest)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		select {
		case <- postPOSTHookerChan:
			break
		case <- time.After(time.Second * 1):
			t.Error("PostPOSTHook did not trigger")
		}
		wg.Done()
	}()

	validFish := Fish{
		FinCount: 2,
	}
	body, _ = json.Marshal(&validFish)
	suite.Post(endpoint, "application/json", bytes.NewReader(body))
	wg.Wait()
	suite.AssertOk()
}

func TestPrePUTHook(t *testing.T) {
	endpoint := "/fish"
	suite := reveltest.NewTestSuite()
	body, _ := json.Marshal(pond[0])
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		select {
		case <- prePUTHookerChan:
			break
		case <- time.After(time.Second * 1):
			t.Error("PrePUTHook did not trigger")
		}
		wg.Done()
	}()

	suite.Put(endpoint, "application/json", bytes.NewReader(body))
	wg.Wait()
	suite.AssertOk()
}

func TestPostPUTHook(t *testing.T) {
	endpoint := "/fish"
	suite := reveltest.NewTestSuite()
	fish := pond[0]
	fish.FinCount = 1
	body, _ := json.Marshal(&fish)
	suite.Post(endpoint, "application/json", bytes.NewReader(body))
	suite.AssertStatus(http.StatusBadRequest)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		select {
		case <- postPUTHookerChan:
			break
		case <- time.After(time.Second * 1):
			t.Error("PostPUTHook did not trigger")
		}
		wg.Done()
	}()

	fish.FinCount = 2
	body, _ = json.Marshal(&fish)
	suite.Post(endpoint, "application/json", bytes.NewReader(body))
	wg.Wait()
	suite.AssertOk()
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
