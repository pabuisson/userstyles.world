package user

import (
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"userstyles.world/config"
	"userstyles.world/database"
	"userstyles.world/handlers/jwt"
	"userstyles.world/models"
	"userstyles.world/utils"
)

func LoginGet(c *fiber.Ctx) error {
	if u, ok := jwt.User(c); ok {
		log.Printf("User %d has set session, redirecting.", u.ID)
		return c.Redirect("/account", fiber.StatusSeeOther)
	}

	return c.Render("login", fiber.Map{
		"Title": "Login",
	})
}

func LoginPost(c *fiber.Ctx) error {
	form := models.User{
		Email:    c.FormValue("email"),
		Password: c.FormValue("password"),
	}
	remember := c.FormValue("remember") == "on"

	err := utils.Validate().StructPartial(form, "Email", "Password")
	if err != nil {
		errors := err.(validator.ValidationErrors)
		log.Println("Validation errors:", errors)

		return c.Render("login", fiber.Map{
			"Title":  "Login failed",
			"Errors": errors,
		})
	}

	user, err := models.FindUserByEmail(database.DB, form.Email)
	if err != nil {
		log.Printf("Failed to find %s, error: %s", form.Email, err)

		c.SendStatus(fiber.StatusUnauthorized)
		return c.Render("login", fiber.Map{
			"Title": "Login failed",
			"Error": "Invalid credentials.",
		})
	}
	if user.OAuthProvider != "none" {
		c.SendStatus(fiber.StatusUnauthorized)
		return c.Render("login", fiber.Map{
			"Title": "Login failed",
			"Error": "Login via OAuth provider",
		})
	}

	match := utils.CompareHashedPassword(user.Password, form.Password)
	if match != nil {
		log.Printf("Failed to match hash for user: %#+v\n", user.Email)

		c.SendStatus(fiber.StatusInternalServerError)
		return c.Render("login", fiber.Map{
			"Title": "Login failed",
			"Error": "Invalid credentials.",
		})
	}

	var expiration time.Time
	if remember {
		// 2 weeks
		expiration = time.Now().Add(time.Hour * 24 * 14)
	}
	t, err := utils.NewJWTToken().
		SetClaim("id", user.ID).
		SetClaim("name", user.Username).
		SetClaim("email", user.Email).
		SetClaim("role", user.Role).
		SetExpiration(expiration).
		GetSignedString(nil)

	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		return c.Render("err", fiber.Map{
			"Title": "Internal server error.",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     fiber.HeaderAuthorization,
		Value:    t,
		Path:     "/",
		Expires:  expiration,
		Secure:   config.DB != "dev.db",
		HTTPOnly: true,
		SameSite: "strict",
	})

	return c.Redirect("/account", fiber.StatusSeeOther)
}
