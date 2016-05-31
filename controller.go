package apikit

import (
	"github.com/revel/revel"
	"fmt"
	"net/http"
	"reflect"
	"encoding/json"
	"errors"
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
		if !instance.CanBeCreatedBy(c.authenticatedUser) {
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
		// ensure that this is a pre-existing record
		preExisting := c.modelProvider.GetModelByID(instance.UniqueID())
		if err := c.copyImmutableAttributes(instance, preExisting); err != nil {
			return DefaultInternalServerErrorMessage()
		}
		if hooker, ok := c.modelProvider.(PUTHooker); ok {
			if prematureResult := hooker.PrePUTHook(instance, preExisting, c.authenticatedUser); prematureResult != nil {
				return prematureResult
			}
		}

		if preExisting == nil {
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
				if prematureResult := hooker.PostPUTHook(instance, preExisting, c.authenticatedUser, err); prematureResult != nil {
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
		if !found.CanBeDeletedBy(c.authenticatedUser) {
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

func (c *GenericRESTController) copyImmutableAttributes(newInstance, existingInstance RESTObject) error {
	if newInstance == nil {
		return errors.New("Given a nil destination object")
	}
	if existingInstance == nil {
		return errors.New("Given a nil source object")
	}
	if copier, ok := existingInstance.(ImmutableAttributeCopier); ok {
		// use the custom implementation if exists
		return copier.CopyImmutableAttributes(newInstance)
	} else {
		// copy immutable attributes based on struct tags
		var vOld reflect.Value = reflect.ValueOf(existingInstance)
		if vOld.Type().Kind() == reflect.Ptr {
			if vOld.Elem().Type().Kind() == reflect.Struct {
				vOld = vOld.Elem()
			} else {
				return errors.New("Source is not a pointer to a struct")
			}
		} else {
			return errors.New("Source is not a pointer to a struct")
		}

		var vNew reflect.Value = reflect.ValueOf(newInstance)
		if vNew.Type().Kind() == reflect.Ptr {
			if vNew.Elem().Type().Kind() == reflect.Struct {
				vNew = vNew.Elem()
			} else {
				return errors.New("Destination is not a pointer to a struct")
			}
		} else {
			return errors.New("Destination is not a pointer to a struct")
		}


		for i := 0; i < vOld.NumField(); i++ {
			oldField := vOld.Type().Field(i)
			if oldField.Tag.Get("apikit") == "immutable" {
				// this field was marked as immutable
				fieldName := oldField.Name
				if newField := vNew.FieldByName(fieldName); newField.IsValid() && newField.CanSet() {
					newField.Set(vOld.FieldByName(fieldName))
				}
			}
		}
		return nil
	}
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

func DefaultInternalServerErrorMessage() ApiMessage {
	const defaultErrMsg string = "An unexpected error ocurred"
	return ApiMessage{
		StatusCode: http.StatusInternalServerError,
		Message: revel.Config.StringDefault("apikit.internalservererror", defaultErrMsg),
	}
}
