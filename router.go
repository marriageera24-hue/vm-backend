package main

import (
	"log"
	"os"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func startServer() {
	e := echo.New()

	httpAuth := middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		apiPW := os.Getenv("VM_HTTP_PASSWORD")
		if apiPW == "" {
			return false, nil
		}

		if password == apiPW {
			return true, nil
		}

		return false, nil
	})

	e.GET("/admin/otp", getOTP, httpAuth)

	jwtAuth := middleware.JWT([]byte(os.Getenv("JWT_SECRET")))

	objects := []string{"users", "email_templates"}
	for _, s := range objects {
		e.POST("/"+s, apiCreateHandler, jwtAuth)
		e.POST("/"+s+"/search", apiSearchHandler, jwtAuth)
		e.POST("/"+s+"/bulk", apiBulkReadHandler, jwtAuth)
		e.GET("/"+s+"/:uuid", apiReadHandler, jwtAuth)
		e.PATCH("/"+s+"/:uuid", apiUpdateHandler, jwtAuth)
		e.DELETE("/"+s+"/:uuid", apiDeleteHandler, jwtAuth)
		e.POST("/"+s+"/:uuid/restore", apiRestoreHandler, jwtAuth)
	}

	// Users
	e.POST("/users/register", userRegisterHandler)                   // Open endpoint
	e.POST("/users/login/password", sessionsLoginPasswordHandler)    // Open endpoint
	e.POST("/users/login/send-email", sessionsLoginSendEmailHandler) // Open endpoint
	e.POST("/users/login/email", sessionsLoginEmailHandler)          // Open endpoint
	e.GET("/users/exists/:info", userExistsHandler)                  // Open endpoint
	e.POST("/users/search/count", (User{}).Count, jwtAuth)

	e.POST("/users/forgot", passwordForgotHandler)                        // Open endpoint
	e.POST("/users/reset", passwordResetHandler)                          // Open endpoint
	e.GET("/users/me", userGetHandler, jwtAuth)                           // Open endpoint
	e.PUT("/users/me", userUpdateHandler, jwtAuth)                        // Open endpoint
	e.PATCH("/users/me", userPatchHandler, jwtAuth)                       // Open endpoint
	e.POST("/users/verify", userVerifyHandler, jwtAuth)                   // Open endpoint
	e.POST("/users/me/send-verify/:type", userSendVerifyHandler, jwtAuth) // Open endpoint // type (otp|email)
	e.POST("/users/me/logout", sessionsLogoutHandler, jwtAuth)            // Open endpoint
	// e.POST("/users/me/subscriptions", userSubscriptionsHandler) // Open endpoint
	// e.POST("/users/me/resubscribe", userResubscribeHandler) // Open endpoint

	e.GET("/users/me/interests", interests, jwtAuth)
	e.GET("/users/me/interested", interested, jwtAuth)
	e.POST("/users/me/interest", addInterest, jwtAuth)

	e.POST("/users/media/:type/:uuid", upload, jwtAuth)
	e.GET("/users/media/:uuid", download, jwtAuth)
	// e.GET("/users/media/:id", download, jwtAuth)

	e.GET("/ping", pingHandler)
	e.GET("/lists", listAPIHandler, jwtAuth)

	e.GET("/*", redirectHandler)

	origins := strings.Split(os.Getenv("VM_ALLOWED_ORIGINS"), ",")

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     origins,
		AllowMethods:     []string{echo.OPTIONS, echo.POST},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowCredentials: true,
		MaxAge:           10,
	}))

	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(middleware.HTTPSRedirect())

	if os.Getenv("VM_ENVIRONMENT") != "production" {
		e.Use(middleware.Logger())
	}

	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Use(middleware.Secure())

	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	// ref https://support.microfocus.com/kb/doc.php?id=7013103 to get .pem files
	e.Logger.Fatal(e.StartTLS(":"+port, getEnvFilePath("cert.pem"), getEnvFilePath("key.pem")))
	return

	// e.Logger.Fatal(e.Start(":" + port))
}
