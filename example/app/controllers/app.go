package controllers

import (
	"github.com/revel/revel"
)

type App struct {
	*revel.Controller
}

func (c *App) Hello() revel.Result {
	return c.RenderJson(map[string]interface{}{
		"greeting": "Hello, world!",
	})
}
