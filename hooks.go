package apikit
import "github.com/revel/revel"

type GETHooker interface {
	RESTController
	PreGETHook(id uint64, authUser User) revel.Result
	PostGETHook(model RESTObject, authUser User) revel.Result
}

type POSTHooker interface {
	RESTController
	PrePOSTHook(model RESTObject, authUser User) revel.Result
	PostPOSTHook(model RESTObject, authUser User, err error) revel.Result
}

type PUTHooker interface {
	RESTController
	PrePUTHook(model RESTObject, authUser User) revel.Result
	PostPUTHook(model RESTObject, authUser User, err error) revel.Result
}

type DELETEHooker interface {
	RESTController
	PreDELETEHook(model RESTObject, authUser User) revel.Result
	PostDELETEHook(model RESTObject, authUser User, err error) revel.Result
}
