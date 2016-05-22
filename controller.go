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
	authenticatedUser  User
	typeFactory        func() RESTObject
	getUniqueFunc      func(id uint64) RESTObject
}

type ModelProvider interface {
	GetModel() RESTObject
}

const (
	GetModelMethodName string = "GetModel"
	RESTControllerName string = "RESTController"
	RESTControllertypeFactoryFieldName string = "typeFactory"

)

func (c *RESTController) Ok() revel.Result {
	return jsonResult{
		body: c.typeFactory(),
	}
}

func (c *RESTController) Get(id uint64) revel.Result {
	if found := c.getUniqueFunc(id); found == nil {
		return ApiMessage{
			StatusCode: http.StatusNotFound,
			Message: fmt.Sprint(c.typeName(), " with ID ", id, " not found"),
		}
	} else if !found.CanBeViewedBy(&c.authenticatedUser) {
		return ApiMessage{
			StatusCode: http.StatusUnauthorized,
			Message: fmt.Sprint("Unauthorized to view ", c.typeName(), " with ID ", id),
		}
	} else {
		fmt.Println(found)
		asMap := make(map[string]interface{})
		var _ = json.Unmarshal

		/*data, _ := json.Marshal(found)
		json.Unmarshal(data, &asMap)*/
		return jsonResult{
			body: asMap,
		}
	}
}

func (c *RESTController) typeName() string {
	instance := c.typeFactory()
	return reflect.TypeOf(instance).Name()
}

func (c *RESTController) Authenticate() {
	c.authenticatedUser = &ExampleUser{
		ID: 100,
		Username: "Bilbo",
	}
}

/*
var registeredRestObjects = make(map[string]interface{})

func RegisterRESTObject(obj interface{}) {
	if !implementsRESTObject(obj) {
		panic(errors.New("does not conform to RESTObject"))
	}

	var t reflect.Type = reflect.TypeOf(obj)
	var v reflect.Value = reflect.Indirect(reflect.ValueOf(obj))
	name := v.Type().Name()

	fmt.Println("name of registerd is", name)
	registeredRestObjects[name] = t
}*/
