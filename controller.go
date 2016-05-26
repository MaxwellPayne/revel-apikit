package apikit

import (
	"github.com/revel/revel"
	"fmt"
	"net/http"
	"reflect"
	"encoding/json"
)

type GenericRESTController struct {
	authenticatedUser User
	Request           *revel.Request
	modelProvider     RESTController
}

const (
	RESTControllerName string = "GenericRESTController"
)

func (c *GenericRESTController) Get(id uint64) revel.Result {
	if !c.modelProvider.EnableGET() {
		return DefaultBadRequestMessage()
	}
	if hooker, ok := c.modelProvider.(GETHooker); ok {
		if prematureResult := hooker.PreGETHook(id, c.authenticatedUser); prematureResult != nil {
			return prematureResult
		}
	}
	if found := c.modelProvider.GetModelByID(id); found == nil {
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
		if hooker, ok := c.modelProvider.(GETHooker); ok {
			if prematureResult := hooker.PostGETHook(found, c.authenticatedUser); prematureResult != nil {
				return prematureResult
			}
		}
		return HookJsonResult{
			Body: found,
		}
	}
}

func (c *GenericRESTController) Post() revel.Result {
	instance := c.modelProvider.ModelFactory()
	if !c.modelProvider.EnablePOST() {
		return DefaultNotFoundMessage()
	}
	return c.unmarshalRequestBody(&instance, func() revel.Result {
		if hooker, ok := c.modelProvider.(POSTHooker); ok {
			if prematureResult := hooker.PrePOSTHook(instance, c.authenticatedUser); prematureResult != nil {
				return prematureResult
			}
		}
		if !instance.CanBeModifiedBy(c.authenticatedUser) {
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
			if hooker, ok := c.modelProvider.(POSTHooker); ok {
				if prematureResult := hooker.PostPOSTHook(instance, c.authenticatedUser, err); prematureResult != nil {
					return prematureResult
				}
			}
			return HookJsonResult{
				Body: instance,
			}
		}
	})
}

func (c *GenericRESTController) Put() revel.Result {
	instance := c.modelProvider.ModelFactory()
	if !c.modelProvider.EnablePUT() {
		return DefaultNotFoundMessage()
	}
	return c.unmarshalRequestBody(&instance, func() revel.Result {
		if hooker, ok := c.modelProvider.(PUTHooker); ok {
			if prematureResult := hooker.PrePUTHook(instance, c.authenticatedUser); prematureResult != nil {
				return prematureResult
			}
		}
		// ensure that this is a pre-existing record
		if preExisting := c.modelProvider.GetModelByID(instance.UniqueID()); preExisting == nil {
			return ApiMessage{
				StatusCode: http.StatusBadRequest,
				Message: fmt.Sprint(c.modelName(), " with ID ", instance.UniqueID(), " does not exist"),
			}
		}
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
			if hooker, ok := c.modelProvider.(PUTHooker); ok {
				if prematureResult := hooker.PostPUTHook(instance, c.authenticatedUser, err); prematureResult != nil {
					return prematureResult
				}
			}
			return HookJsonResult{
				Body: instance,
			}
		}
	})
}

func (c *GenericRESTController) Delete(id uint64) revel.Result {
	if !c.modelProvider.EnableDELETE() {
		return DefaultNotFoundMessage()
	}
	if found := c.modelProvider.GetModelByID(id); found == nil {
		return ApiMessage{
			StatusCode: http.StatusNotFound,
			Message: fmt.Sprint(c.modelName(), " with ID ", id, " not found"),
		}
	} else {
		if hooker, ok := c.modelProvider.(DELETEHooker); ok {
			if prematureResult := hooker.PreDELETEHook(found, c.authenticatedUser); prematureResult != nil {
				return prematureResult
			}
		}
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
			if hooker, ok := c.modelProvider.(DELETEHooker); ok {
				if prematureResult := hooker.PostDELETEHook(found, c.authenticatedUser, err); prematureResult != nil {
					return prematureResult
				}
			}
			return ApiMessage{
				StatusCode: http.StatusOK,
				Message: "Success",
			}
		}
	}
}

func (c *GenericRESTController) modelName() string {
	instance := c.modelProvider.ModelFactory()
	return reflect.TypeOf(instance).Elem().Name()
}

func (c *GenericRESTController) unmarshalRequestBody(o interface{}, next func() revel.Result) revel.Result {
	err := json.NewDecoder(c.Request.Body).Decode(o)
	if err != nil {
		return DefaultBadRequestMessage()
	}
	return next()
}

func DefaultBadRequestMessage() ApiMessage {
	return ApiMessage{
		StatusCode: http.StatusBadRequest,
		Message: "Improperly formatted request body",
	}
}

func DefaultNotFoundMessage() ApiMessage {
	return ApiMessage{
		StatusCode: http.StatusNotFound,
		Message: "Not Found",
	}
}
