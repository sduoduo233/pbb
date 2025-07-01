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

	// public routes

	g.GET("/", index)
	g.GET("/servers", indexServers)
	g.GET("/servers/:id", view)
	g.GET("/servers/:id/data", viewData)

	// auth

	g.GET("/login", login)
	g.POST("/login", doLogin)

	// admin dashboard

	g.GET("/dashboard/servers", servers, mustAuth)
	g.GET("/dashboard/servers/add", addServer, mustAuth)
	g.POST("/dashboard/servers/add", doAddServer, mustAuth)
	g.GET("/dashboard/servers/edit/:id", editServer, mustAuth)
	g.POST("/dashboard/servers/edit/:id", doEditServer, mustAuth)
	g.DELETE("/dashboard/servers/:id", deleteServer, mustAuth)
	g.GET("/dashboard/servers/install/:id", installAgent, mustAuth)

	g.GET("/dashboard/groups", groups, mustAuth)
	g.GET("/dashboard/groups/add", addGroup, mustAuth)
	g.POST("/dashboard/groups/add", doAddGroup, mustAuth)
	g.GET("/dashboard/groups/:id", editGroup, mustAuth)
	g.POST("/dashboard/groups/:id", doEditGroup, mustAuth)
	g.DELETE("/dashboard/groups/:id", deleteGroup, mustAuth)

	g.GET("/dashboard/user", user, mustAuth)
	g.POST("/dashboard/user", doChangePassword, mustAuth)

	g.GET("/dashboard/settings", settings, mustAuth)
	g.POST("/dashboard/settings", doUpdateSettings, mustAuth)

	// agent reporting

	g.POST("/agent/metric", metrics)
	g.POST("/agent/info", systemInfo)

}
