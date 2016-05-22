package apikit
import (
	"github.com/revel/revel"
	"reflect"
	"fmt"
	"errors"
)

var _ = fmt.Println

type AuthenticationFunction func(username, password string) User

func CreateRESTControllerInjectionFilter(authFunction AuthenticationFunction) revel.Filter {
	return func(c *revel.Controller, fc []revel.Filter) {
		if embedsRESTController(c.AppController) {
			// perform reflection only if this is a RESTController
			var controllerValue reflect.Value = reflect.ValueOf(c.AppController)

			implementsModelProvider := controllerValue.Type().Implements(reflect.TypeOf((*ModelProvider)(nil)).Elem())
			if !implementsModelProvider {
				panic(errors.New("Type Injection: Given Controller does not conform to ModelProvider"))
			}

			restController := getEmbeddedRESTController(c.AppController)

			// attach the type factory to RESTController
			modelFactoryMethodValue := controllerValue.MethodByName(ModelFactoryFunc)
			modelFactory := modelFactoryMethodValue.Call([]reflect.Value{})[0].Interface().(func() RESTObject)
			restController.modelFactory = modelFactory

			// attach the get unique function to RESTController
			getUniqueMethodValue := controllerValue.MethodByName(GetModelByIDMethodName)
			getUniqueFunc := getUniqueMethodValue.Call([]reflect.Value{})[0].Interface().(func(id uint64) RESTObject)
			restController.getUniqueFunc = getUniqueFunc

			if username, pass, ok := c.Request.BasicAuth(); ok {
				restController.authenticatedUser = authFunction(username, pass)
			}
			restController.Request = c.Request

			setEmbeddedRESTController(c.AppController, *restController)
		}

		fc[0](c, fc[1:]) // Execute the next filter stage.
	}
}
