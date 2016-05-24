package controllers

import (
	"github.com/revel/revel"
	"github.com/MaxwellPayne/revel-apikit"
	"github.com/MaxwellPayne/revel-apikit/example/app/models"
)

// Controller for Users
type UserController struct {
	*revel.Controller
	apikit.GenericRESTController
}

// Implementation of ModelProvider interface
func (c *UserController) ModelFactory() apikit.RESTObject {
	return &models.User{}
}

func (c *UserController) GetModelByID(id uint64) apikit.RESTObject {
	if u := models.GetUserByID(id); u == nil {
		return nil
	} else {
		return u
	}
}

func (c *UserController) EnableGET() bool {
	return true
}

func (c *UserController) EnablePOST() bool {
	return true
}

func (c *UserController) EnablePUT() bool {
	return true
}

func (c *UserController) EnableDELETE() bool {
	return true
}
