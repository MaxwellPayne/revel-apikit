package apikit
import (
	"reflect"
	"fmt"
)

var _ = fmt.Println

func implementsRESTObject(obj interface{}) bool {
	t := reflect.TypeOf(obj)
	return t.Implements(reflect.TypeOf((*RESTObject)(nil)).Elem())
}

func embedsRESTController(obj interface{}) bool {
	return getEmbeddedRESTController(obj) != nil
}

func getEmbeddedRESTController(obj interface{}) *RESTController {
	var theStruct reflect.Value

	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Struct {
		theStruct = reflect.ValueOf(obj)
	} else if t.Kind() == reflect.Ptr {
		dereferenced := reflect.Indirect(reflect.ValueOf(obj))
		if dereferenced.Kind() == reflect.Struct {
			theStruct = dereferenced
		} else {
			return nil
		}
	} else {
		return nil
	}

	for fieldIdx := 0; fieldIdx < theStruct.NumField(); fieldIdx ++ {
		field := theStruct.Type().Field(fieldIdx)
		if field.Name == RESTControllerName && field.Anonymous {
			ctrl := theStruct.FieldByIndex(field.Index).Interface().(RESTController)
			return &ctrl
		}
	}
	return nil
}

func setEmbeddedRESTController(obj interface{}, ctrl RESTController) (ok bool) {
	var theStruct reflect.Value

	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Struct {
		theStruct = reflect.ValueOf(obj)
	} else if t.Kind() == reflect.Ptr {
		dereferenced := reflect.Indirect(reflect.ValueOf(obj))
		if dereferenced.Kind() == reflect.Struct {
			theStruct = dereferenced
		} else {
			return false
		}
	} else {
		return false
	}

	for fieldIdx := 0; fieldIdx < theStruct.NumField(); fieldIdx ++ {
		field := theStruct.Type().Field(fieldIdx)
		if field.Name == RESTControllerName && field.Anonymous {
			theStruct.FieldByIndex(field.Index).Set(reflect.ValueOf(ctrl))
			return true
		}
	}
	return false
}
