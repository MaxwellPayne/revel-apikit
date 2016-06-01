package apikit

import (
	"encoding/json"
	"bytes"
	"time"
	"net/http"
	"strconv"
	"net"
	"testing"
	reveltest "github.com/revel/revel/testing"
	"github.com/revel/revel"
)

type EmbeddedFish struct {
	Fish
}

type EmbeddedFishController struct {
	*revel.Controller
	GenericRESTController
}

func (c *EmbeddedFishController) ModelFactory() RESTObject {
	return &EmbeddedFish{}
}

func (c *EmbeddedFishController) GetModelByID(id uint64) RESTObject {
	for _, f := range pond {
		if f.ID == id {
			return &EmbeddedFish{f}
		}
	}
	return nil
}

func (c *EmbeddedFishController) EnableGET() bool {
	return true
}

func (c *EmbeddedFishController) EnablePOST() bool {
	return true
}

func (c *EmbeddedFishController) EnablePUT() bool {
	return true
}

func (c *EmbeddedFishController) EnableDELETE() bool {
	return true
}

func TestCopyImmutableAttributesCustom(t *testing.T) {
	suite := reveltest.NewTestSuite()
	endpoint := "/user"
	putUrl := "http://" + net.JoinHostPort("localhost", strconv.Itoa(testPort)) + endpoint

	admin := *usersDB[2]
	suite.Assert(admin.IsAdmin)
	adminData, _ := json.Marshal(&admin)

	req := suite.PutCustom(putUrl, "application/json", bytes.NewReader(adminData))
	req.SetBasicAuth(admin.Username, admin.Password)
	req.MakeRequest()
	// should have gotten a 500 error from CopyImmutableAttributes
	suite.AssertStatus(http.StatusInternalServerError)

	nonAdmin := *usersDB[0]
	originalCreateDate := nonAdmin.DateCreated
	suite.Assert(!originalCreateDate.IsZero())

	// this is an immutable attribute, should not change
	newCreateDate := originalCreateDate.Add(time.Hour * 100)
	nonAdmin.DateCreated = newCreateDate
	nonAdminData, _ := json.Marshal(&nonAdmin)

	req = suite.PutCustom(putUrl, "application/json", bytes.NewReader(nonAdminData))
	req.SetBasicAuth(nonAdmin.Username, nonAdmin.Password)
	req.MakeRequest()
	suite.AssertOk()

	updatedNonAdmin := ExampleUser{}
	err := json.Unmarshal(suite.ResponseBody, &updatedNonAdmin)
	suite.Assert(err == nil)
	// make sure that attributed did not change
	suite.Assert(originalCreateDate.Equal(updatedNonAdmin.DateCreated))
	suite.Assert(!newCreateDate.Equal(updatedNonAdmin.DateCreated))
}

func TestCopyImmutableAttributesFromStructTag(t *testing.T) {
	endpoint := "/fish"
	fish := pond[0]
	suite := reveltest.NewTestSuite()
	originalCreateDate := fish.CreateDate
	suite.Assert(!originalCreateDate.IsZero())

	// this is an immutable attribute, should not change
	newCreateDate := originalCreateDate.Add(time.Hour * 100)
	fish.CreateDate = newCreateDate
	body, _ := json.Marshal(&fish)

	suite.Put(endpoint, "application/json", bytes.NewReader(body))
	suite.AssertOk()

	updatedFish := Fish{}
	err := json.Unmarshal(suite.ResponseBody, &updatedFish)
	suite.Assert(err == nil)
	suite.Assert(originalCreateDate.Equal(updatedFish.CreateDate))
	suite.Assert(!newCreateDate.Equal(updatedFish.CreateDate))
}

func TestCopyImmutableAttributesEmbedded(t *testing.T) {
	endpoint := "/embeddedfish"
	fish := pond[0]
	suite := reveltest.NewTestSuite()
	originalCreateDate := fish.CreateDate
	suite.Assert(!originalCreateDate.IsZero())

	// this is an immutable attribute, should not change
	newCreateDate := originalCreateDate.Add(time.Hour * 100)
	fish.CreateDate = newCreateDate
	body, _ := json.Marshal(&fish)

	suite.Put(endpoint, "application/json", bytes.NewReader(body))
	suite.AssertOk()

	updatedFish := EmbeddedFish{}
	err := json.Unmarshal(suite.ResponseBody, &updatedFish)
	suite.Assert(err == nil)
	suite.Assert(originalCreateDate.Equal(updatedFish.CreateDate))
	suite.Assert(!newCreateDate.Equal(updatedFish.CreateDate))
}