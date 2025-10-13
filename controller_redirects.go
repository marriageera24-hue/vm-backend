package main

import (
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo"
)

func redirectHandler(ctx echo.Context) error {
	url := os.Getenv("404_REDIRECT_URL")

	uri := ctx.Request().RequestURI
	log.Print("/public" + uri)

	// if IsOneOf(uri, objects) {
	// 	return ctx.File("public/index.html")
	// 	// ctx.Echo().Static("/", "/public/index.html")
	// 	// return nil
	// }

	if _, err := os.Stat("public" + uri); err == nil {
		ctx.Echo().Static("/", "public")
		return nil
	}

	if uri == "/" {
		// ctx.Echo().Static("/", "public/index.html")
		return ctx.File("public/index.html")
	}

	return ctx.Redirect(http.StatusMovedPermanently, url)
}

func frontEndRouteHandler(ctx echo.Context) error {
	return ctx.File("public/index.html")
}
