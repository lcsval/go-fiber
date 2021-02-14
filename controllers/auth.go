package controllers

import (
	"fmt"
	"go-fiber/models"
	"go-fiber/repository"
	"go-fiber/security"
	"go-fiber/util"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gopkg.in/asaskevich/govalidator.v9"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
	var newUser models.User
	err := ctx.BodyParser(&newUser)
	if err != nil {
		return ctx.Status(http.StatusUnprocessableEntity).JSON(util.NewJError(err))
	}

	newUser.Email = util.NormalizeEmail(newUser.Email)
	if !govalidator.IsEmail(newUser.Email) {
		return ctx.Status(http.StatusBadRequest).JSON(util.NewJError(util.ErrInvalidEmail))
	}

	exists, err := c.usersRepo.GetByEmail(newUser.Email)
	if err == mgo.ErrNotFound {
		if strings.TrimSpace(newUser.Password) == "" {
			return ctx.Status(http.StatusBadRequest).JSON(util.NewJError(util.ErrEmptyPassword))
		}

		newUser.Password, err = security.EncryptPassword(newUser.Password)
		if err != nil {
			return ctx.Status(http.StatusBadRequest).JSON(util.NewJError(err))
		}

		newUser.CreatedAt = time.Now()
		newUser.UpdatedAt = newUser.CreatedAt
		newUser.Id = bson.NewObjectId()
		err = c.usersRepo.Save(&newUser)
		if err != nil {
			return ctx.Status(http.StatusBadRequest).JSON(util.NewJError(err))
		}
		return ctx.Status(http.StatusCreated).JSON(newUser)
	}

	if exists != nil {
		err = util.ErrEmailAlreadyExists
	}

	return ctx.Status(http.StatusBadRequest).JSON(util.NewJError(err))
}

func (c *authController) SignIn(ctx *fiber.Ctx) error {
	var input models.User
	err := ctx.BodyParser(&input)
	if err != nil {
		return ctx.Status(http.StatusUnprocessableEntity).JSON(util.NewJError(err))
	}

	input.Email = util.NormalizeEmail(input.Email)
	user, err := c.usersRepo.GetByEmail(input.Email)
	if err != nil {
		log.Printf("%s signin failed: %v\n", input.Email, err.Error())
		return ctx.Status(http.StatusUnprocessableEntity).JSON(util.NewJError(util.ErrInvalidCredentials))
	}

	err = security.VerifyPassword(user.Password, input.Password)
	if err != nil {
		log.Printf("%s signin failed: %v\n", input.Email, err.Error())
		return ctx.Status(http.StatusUnprocessableEntity).JSON(util.NewJError(err))
	}

	token, err := security.NewToken(user.Id.Hex())
	if err != nil {
		log.Printf("%s signin failed: %v\n", input.Email, err.Error())
		return ctx.Status(http.StatusUnprocessableEntity).JSON(util.NewJError(err))
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"user":  user,
		"token": fmt.Sprintf("Bearer %s", token),
	})
}

func (c *authController) GetUser(ctx *fiber.Ctx) error {
	payload, err := AuthRequestWithWith(ctx)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(util.NewJError(err))
	}
	user, err := c.usersRepo.GetById(payload.id)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(util.NewJError(err))
	}
	return ctx.Status(http.StatusOK).JSON(user)
}

func AuthRequestWithWith(ctx *fiber.Ctx) (*jwt.StandardClaims, error) {
	id := ctx.Params("id")
	if !bson.IsObjectIdHex(id) {
		return nil, util.ErrUnauthorized
	}
	token := ctx.Locals("user").(*jwt.Token)
	payload, err := security.ParseToken(token.Raw)
	if err != nil {
		return nil, err
	}
	if payload.Id != id || payload.Issuer != id {
		return nil, util.ErrUnauthorized
	}
	return payload, nil
}
