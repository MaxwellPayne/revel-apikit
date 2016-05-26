package apikit

import (
	"github.com/revel/revel"
)

// A server-side data model that can be served by RESTControllers
type RESTObject interface {
	UniqueID() uint64

	CanBeViewedBy(user User) bool
	CanBeSavedBy(user User) bool

	Validate(v *revel.Validation)

	Delete() error
	Save() error
}

// A Controller that concerns itself with managing one type of RESTObject
type RESTController interface {
	ModelFactory() RESTObject
	GetModelByID(id uint64) RESTObject

	EnableGET() bool
	EnablePOST() bool
	EnablePUT() bool
	EnableDELETE() bool
}

// A RESTObject that can be authenticated by RESTControllers
type User interface {
	RESTObject
	HasAdminPrivileges() bool
}

