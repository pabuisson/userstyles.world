package style

import (
	"github.com/gofiber/fiber/v2"

	"userstyles.world/database"
	"userstyles.world/handlers/jwt"
	"userstyles.world/models"
)

func StylePromote(c *fiber.Ctx) error {
	u, _ := jwt.User(c)
	p := c.Params("id")

	// Only moderator and above have permissions to promote styles.
	if u.Role < models.Moderator {
		return c.Render("err", fiber.Map{
			"Title": "You don't have enough permission for this.",
			"User":  u,
		})
	}

	// TODO: Make it possible to remove promotion.
	err := database.DB.
		Model(models.Style{}).
		Where("id = ?", p).
		Update("featured", true).
		Error

	if err != nil {
		return c.Render("err", fiber.Map{
			"Title": "Failed to promote a style",
			"User":  u,
		})
	}

	return c.Redirect("/style/"+p, fiber.StatusSeeOther)
}
