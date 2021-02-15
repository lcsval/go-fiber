package routes

import (
	"go-fiber/controllers"

	"github.com/gofiber/fiber/v2"
)

type authRoutes struct {
	authController controllers.AuthController
}

func NewAuthRoutes(authController controllers.AuthController) Routes {
	return &authRoutes{authController}
}

func (r *authRoutes) Install(app *fiber.App) {
	app.Post("/signup", r.authController.SignUp)
	app.Post("/signin", r.authController.SignIn)
	app.Get("/users/:id", AuthRequired, r.authController.GetUser)
}
