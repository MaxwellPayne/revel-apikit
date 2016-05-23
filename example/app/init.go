package app

import (
	"github.com/revel/revel"
	"github.com/MaxwellPayne/revel-apikit"
	"github.com/MaxwellPayne/revel-apikit/example/app/models"
	"github.com/MaxwellPayne/revel-apikit/example/app/controllers"
)

func init() {
	revel.Filters = []revel.Filter{

		// Replace revel.PanicFilter with apikit.APIPanicFilter
		// This will return JSON apikit.ApiMessage responses instead of templated
		// responses for 404 and 500 errors
		apikit.APIPanicFilter,
		//revel.PanicFilter,

		revel.RouterFilter,
		revel.FilterConfiguringFilter,
		revel.ParamsFilter,
		revel.SessionFilter,
		revel.FlashFilter,
		revel.ValidationFilter,
		revel.I18nFilter,
		HeaderFilter,
		revel.InterceptorFilter,
		revel.CompressFilter,

		// Add this to ensure that RESTControllers are invoked by requests
		// Must come somewhere after revel.RouterFilter
		RESTControllerFilter,

		revel.ActionInvoker,
	}


	// Register RESTControllers OnAppStart
	controllers := []apikit.ModelProvider{
		(*controllers.UserController)(nil),
	}
	revel.OnAppStart(func() {
		apikit.RegisterRESTControllers(controllers)
	})
}

var HeaderFilter = func(c *revel.Controller, fc []revel.Filter) {
	c.Response.Out.Header().Add("X-Frame-Options", "SAMEORIGIN")
	c.Response.Out.Header().Add("X-XSS-Protection", "1; mode=block")
	c.Response.Out.Header().Add("X-Content-Type-Options", "nosniff")

	fc[0](c, fc[1:])
}

var RESTControllerFilter = apikit.CreateRESTControllerInjectionFilter(models.AuthenticationHandler)
