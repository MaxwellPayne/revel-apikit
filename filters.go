package apikit
import (
	"github.com/revel/revel"
	"errors"
	"runtime/debug"
	"net/http"
)

type AuthenticationFunction func(username, password string) User

func CreateRESTControllerInjectionFilter(authFunction AuthenticationFunction) revel.Filter {
	return func(c *revel.Controller, fc []revel.Filter) {
		if embedsRESTController(c.AppController) {
			// use the RESTController only if this controller embeds one

			ctrlAsModelProvider, ok := c.AppController.(RESTController)
			if !ok {
				panic(errors.New("Type Injection: Given Controller does not conform to ModelProvider"))
			}

			restController := getEmbeddedRESTController(c.AppController)
			restController.modelProvider = ctrlAsModelProvider

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
	defer func() {
		if err := recover(); err != nil {
			if revel.DevMode {
				revel.ERROR.Print(err, "\n", string(debug.Stack()))
			}

			c.Result = DefaultInternalServerErrorMessage()
		} else {
			if c.Response.Status == http.StatusNotFound {
				// clobber Revel's html template-based NotFound result
				c.Result = DefaultNotFoundMessage()
			}
		}
	}()

	fc[0](c, fc[1:])
}
