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

			methodValue := controllerValue.MethodByName(GetModelMethodName)

			model := methodValue.Call([]reflect.Value{})[0].Interface().(RESTObject)
			var _ = model

			restController := getEmbeddedRESTController(c.AppController)


			if username, pass, ok := c.Request.BasicAuth(); ok {
				restController.authenticatedUser = authFunction(username, pass)
			}
			restController.typeFactory = func() RESTObject {
				return model
			}

			setEmbeddedRESTController(c.AppController, *restController)
		}

		fc[0](c, fc[1:]) // Execute the next filter stage.
	}
}