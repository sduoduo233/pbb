package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Route(g *echo.Group) {
	g.Use(csrf)
	g.Use(auth)

	g.RouteNotFound("/*", func(c echo.Context) error {
		return c.HTML(http.StatusNotFound, "404 not found")
	})

	g.GET("/", index)

	g.GET("/login", login)
	g.POST("/login", doLogin)

	g.GET("/dashboard/servers", servers, mustAuth)
	g.GET("/dashboard/servers/add", addServer, mustAuth)
	g.POST("/dashboard/servers/add", doAddServer, mustAuth)
	g.GET("/dashboard/servers/edit/:id", editServer, mustAuth)
	g.POST("/dashboard/servers/edit/:id", doEditServer, mustAuth)
	g.DELETE("/dashboard/servers/:id", deleteServer, mustAuth)

	g.GET("/dashboard/groups", groups, mustAuth)
	g.GET("/dashboard/groups/add", addGroup, mustAuth)
	g.POST("/dashboard/groups/add", doAddGroup, mustAuth)
	g.GET("/dashboard/groups/:id", editGroup, mustAuth)
	g.POST("/dashboard/groups/:id", doEditGroup, mustAuth)
	g.DELETE("/dashboard/groups/:id", deleteGroup, mustAuth)

	g.POST("/agent/metric", metrics)

}
