package apikit
import (
	"github.com/revel/revel"
	"reflect"
	"errors"
	"runtime/debug"
	"net/http"
)

type AuthenticationFunction func(username, password string) User

func CreateRESTControllerInjectionFilter(authFunction AuthenticationFunction) revel.Filter {
	return func(c *revel.Controller, fc []revel.Filter) {
		if embedsRESTController(c.AppController) {
			// perform reflection only if this is a RESTController
			var controllerValue reflect.Value = reflect.ValueOf(c.AppController)

			if !implementsModelProvider(c.AppController) {
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

func APIPanicFilter(c *revel.Controller, fc []revel.Filter) {
	const defaultErrMsg string = "An unexpected error ocurred"
	defer func() {
		if err := recover(); err != nil {
			if revel.DevMode {
				revel.ERROR.Print(err, "\n", string(debug.Stack()))
			}

			c.Result = ApiMessage{
				StatusCode: http.StatusInternalServerError,
				Message: revel.Config.StringDefault("apikit.internalservererror", defaultErrMsg),
			}
		} else {
			if c.Response.Status == http.StatusNotFound {
				// clobber Revel's html template-based NotFound result
				c.Result = DefaultNotFoundMessage()
			}
		}
	}()

	fc[0](c, fc[1:])
}
