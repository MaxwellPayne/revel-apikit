package controllers

import (
	"github.com/revel/revel"
	"github.com/MaxwellPayne/revel-apikit"
	"github.com/MaxwellPayne/revel-apikit/example/app/models"
)

// Controller for Users
type UserController struct {
	*revel.Controller
	apikit.RESTController
}

// Implementation of ModelProvider interface
func (c *UserController) ModelFactoryFunc() func() apikit.RESTObject {
	return func() apikit.RESTObject {
		return &models.User{}
	}
}

func (c *UserController) GetModelByIDFunc() func(id uint64) apikit.RESTObject {
	return func(id uint64) apikit.RESTObject {
		if u := models.GetUserByID(id); u == nil {
			return nil
		} else {
			return u
		}
	}
}
