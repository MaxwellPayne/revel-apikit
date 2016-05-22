package apikit

import (
	"time"
	"github.com/revel/revel"
	"fmt"
)

var _ = fmt.Println

type Model struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RESTObject interface {
	CanBeViewedBy(user *User) bool
	CanBeModifiedBy(user *User) bool

	Validate(v *revel.Validation)
	UniqueID() uint64

	Delete() error
	Save() error

	EnableGET() bool
	EnablePOST() bool
	EnablePUT() bool
	EnableDELETE() bool
}

type User interface {
	RESTObject
}
