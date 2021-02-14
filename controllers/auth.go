package controllers

import (
	"go-fiber/models"
	"go-fiber/repository"
	"go-fiber/util"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gopkg.in/mgo.v2"
)

type AuthController interface {
}

type authController struct {
	usersRepo repository.UsersRepository
}

func NewAuthController(usersRepo repository.UsersRepository) AuthController {
	return &authController{usersRepo}
}

func (c *authController) SignUp(ctx *fiber.Ctx) error {
	var input models.User
	err := ctx.BodyParser(&input)
	if err != nil {
		return ctx.Status(http.StatusUnprocessableEntity).JSON(util.NewJError(err))
	}

	input.Email = util.NormalizeEmail(input.Email)
	if !govalidator.IsEmail(input.Email) {
		return ctx.Status(http.StatusBadRequest).JSON(util.NewJError(util.ErrInvalidEmail))
	}

	exists, err := c.usersRepo.GetByEmail(input.Email)
	if err == mgo.ErrNotFound {
		if strings.TrimSpace(input.Password) == "" {
			return ctx.Status(http.StatusBadRequest).JSON(util.NewJError(util.ErrEmptyPassword))
		}
	}

	if exists != nil {
		err = util.ErrEmailAlreadyExists
	}

	return ctx.Status(http.StatusBadRequest).JSON(util.NewJError(err))
}
