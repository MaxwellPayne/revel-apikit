package apikit

import (
	"github.com/revel/revel"
	"fmt"
	"net/http"
	"reflect"
	"encoding/json"
)

var (
	_ = fmt.Println
)

type RESTController struct {
	authenticatedUser User
	Request           *revel.Request
	modelFactory      func() RESTObject
	getUniqueFunc     func(id uint64) RESTObject
}

type ModelProvider interface {
	ModelFactoryFunc() func() RESTObject
	GetModelByIDFunc() func(id uint64) RESTObject
}

const (
	ModelFactoryFunc string = "ModelFactoryFunc"
	GetModelByIDMethodName string = "GetModelByIDFunc"
	RESTControllerName string = "RESTController"
	RESTControllertypeFactoryFieldName string = "typeFactory"
)

func (c *RESTController) Get(id uint64) revel.Result {
	if !c.modelFactory().EnableGET() {
		return defaultBadRequestMessage()
	}
	if found := c.getUniqueFunc(id); found == nil {
		return ApiMessage{
			StatusCode: http.StatusNotFound,
			Message: fmt.Sprint(c.modelName(), " with ID ", id, " not found"),
		}
	} else if !found.CanBeViewedBy(c.authenticatedUser) {
		return ApiMessage{
			StatusCode: http.StatusUnauthorized,
			Message: fmt.Sprint("Unauthorized to view ", c.modelName(), " with ID ", id),
		}
	} else {
		return jsonResult{
			body: found,
		}
	}
}

func (c *RESTController) Post() revel.Result {
	instance := c.modelFactory()
	if !instance.EnablePOST() {
		return defaultNotFoundMessage()
	}
	return c.unmarshalRequestBody(&instance, func() revel.Result {
		// Users should not be subject to modification checks during Post
		if (!instance.CanBeModifiedBy(c.authenticatedUser)) && (!implementsUser(instance)) {
			return ApiMessage{
				StatusCode: http.StatusUnauthorized,
				Message: "Not authorized to post this " + c.modelName(),
			}
		}
		if err := instance.Save(); err != nil {
			return ApiMessage{
				StatusCode: http.StatusBadRequest,
				Message: err.Error(),
			}
		} else {
			return jsonResult{
				body: instance,
			}
		}
	})
}

func (c *RESTController) Put() revel.Result {
	instance := c.modelFactory()
	if !instance.EnablePUT() {
		return defaultNotFoundMessage()
	}
	return c.unmarshalRequestBody(&instance, func() revel.Result {
		if !instance.CanBeModifiedBy(c.authenticatedUser) {
			return ApiMessage{
				StatusCode: http.StatusUnauthorized,
				Message: "Not authorized to modify this " + c.modelName(),
			}
		}
		if err := instance.Save(); err != nil {
			return ApiMessage{
				StatusCode: http.StatusBadRequest,
				Message: err.Error(),
			}
		} else {
			return jsonResult{
				body: instance,
			}
		}
	})
}

func (c *RESTController) Delete(id uint64) revel.Result {
	if !c.modelFactory().EnableDELETE() {
		return defaultNotFoundMessage()
	}
	if found := c.getUniqueFunc(id); found == nil {
		return ApiMessage{
			StatusCode: http.StatusNotFound,
			Message: fmt.Sprint(c.modelName(), " with ID ", id, " not found"),
		}
	} else {
		if !found.CanBeModifiedBy(c.authenticatedUser) {
			return ApiMessage{
				StatusCode: http.StatusUnauthorized,
				Message: "Not authorized to delete this " + c.modelName(),
			}
		}
		if err := found.Delete(); err != nil {
			return ApiMessage{
				StatusCode: http.StatusBadRequest,
				Message: err.Error(),
			}
		} else {
			return ApiMessage{
				StatusCode: http.StatusOK,
				Message: "Success",
			}
		}
	}
}

func (c *RESTController) modelName() string {
	instance := c.modelFactory()
	return reflect.TypeOf(instance).Name()
}

func (c *RESTController) unmarshalRequestBody(o interface{}, next func() revel.Result) revel.Result {
	err := json.NewDecoder(c.Request.Body).Decode(o)
	if err != nil {
		return defaultBadRequestMessage()
	}
	return next()
}

func defaultBadRequestMessage() ApiMessage {
	return ApiMessage{
		StatusCode: http.StatusBadRequest,
		Message: "Improperly formatted request body",
	}
}

func defaultNotFoundMessage() ApiMessage {
	return ApiMessage{
		StatusCode: http.StatusNotFound,
		Message: "Not Found",
	}
}
