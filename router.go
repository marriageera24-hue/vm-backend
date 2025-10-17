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
		e.POST("/api/"+s, apiCreateHandler, jwtAuth)
		e.POST("/api/"+s+"/search", apiSearchHandler, jwtAuth)
		e.POST("/api/"+s+"/bulk", apiBulkReadHandler, jwtAuth)
		e.GET("/api/"+s+"/:uuid", apiReadHandler, jwtAuth)
		e.PATCH("/api/"+s+"/:uuid", apiUpdateHandler, jwtAuth)
		e.DELETE("/api/"+s+"/:uuid", apiDeleteHandler, jwtAuth)
		e.POST("/api/"+s+"/:uuid/restore", apiRestoreHandler, jwtAuth)
	}
	e.GET("/api/admin/users/search", apiAdminSearchHandler, httpAuth)
	e.DELETE("/api/admin/users/delete", apiAdminDeleteHandler, httpAuth)
	e.POST("/api/admin/users/verify", userVerifyHandler, httpAuth)
	e.POST("/api/admin/users/unverify", userUnverifyHandler, httpAuth)   // Open endpoint
	e.PATCH("/api/admin/users/update", userAdminUpdateHandler, httpAuth) // Open endpoint

	// Users
	e.POST("/api/users/register", userRegisterHandler)                   // Open endpoint
	e.POST("/api/users/login/password", sessionsLoginPasswordHandler)    // Open endpoint
	e.POST("/api/users/login/send-email", sessionsLoginSendEmailHandler) // Open endpoint
	e.POST("/api/users/login/email", sessionsLoginEmailHandler)          // Open endpoint
	e.GET("/api/users/exists/:info", userExistsHandler)                  // Open endpoint
	e.POST("/api/users/search/count", (User{}).Count, jwtAuth)

	e.POST("/api/users/forgot", passwordForgotHandler) // Open endpoint
	e.POST("/api/users/reset", passwordResetHandler)   // Open endpoint
	e.GET("/api/users/me", userGetHandler, jwtAuth)
	e.POST("/api/users/delete", userDeleteHandler, jwtAuth)                   // Open endpoint
	e.PUT("/api/users/me", userUpdateHandler, jwtAuth)                        // Open endpoint
	e.PATCH("/api/users/me", userPatchHandler, jwtAuth)                       // Open endpoint
	e.POST("/api/users/me/send-verify/:type", userSendVerifyHandler, jwtAuth) // Open endpoint // type (otp|email)
	e.POST("/api/users/me/logout", sessionsLogoutHandler, jwtAuth)            // Open endpoint
	// e.POST("/users/me/subscriptions", userSubscriptionsHandler) // Open endpoint
	// e.POST("/users/me/resubscribe", userResubscribeHandler) // Open endpoint

	e.GET("/api/users/me/interests", interests, jwtAuth)
	e.GET("/api/users/me/interested", interested, jwtAuth)
	e.POST("/api/users/me/interest", addInterest, jwtAuth)

	e.POST("/api/users/media/:type/:uuid", upload, jwtAuth)
	e.GET("/api/users/media/:uuid", download, jwtAuth)
	e.POST("/api/users/payments", savePayments, jwtAuth)
	e.GET("/api/users/balance", getBalance, jwtAuth)
	// e.GET("/users/media/:id", download, jwtAuth)

	e.GET("/api/ping", pingHandler)
	e.GET("/api/lists", listAPIHandler, jwtAuth)

	e.GET("/*", redirectHandler)

	origins := strings.Split(os.Getenv("VM_ALLOWED_ORIGINS"), ",")

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     origins,
		AllowMethods:     []string{echo.OPTIONS, echo.POST, echo.DELETE, echo.PATCH},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
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
	print("port: ", port)

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	// ref https://support.microfocus.com/kb/doc.php?id=7013103 to get .pem files
	e.Logger.Fatal(e.StartTLS(":"+port, getEnvFilePath("cert.pem"), getEnvFilePath("key.pem")))
	return

	// e.Logger.Fatal(e.Start(":" + port))
}
