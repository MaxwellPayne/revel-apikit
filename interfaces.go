package apikit

import (
	"github.com/revel/revel"
)

type RESTObject interface {
	CanBeViewedBy(user User) bool
	CanBeModifiedBy(user User) bool

	Validate(v *revel.Validation)

	Delete() error
	Save() error

	EnableGET() bool
	EnablePOST() bool
	EnablePUT() bool
	EnableDELETE() bool
}

type User interface {
	RESTObject
	UserID() uint64
}
