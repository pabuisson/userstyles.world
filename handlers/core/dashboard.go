package core

import (
	"sort"

	"github.com/gofiber/fiber/v2"

	"userstyles.world/handlers/jwt"
	"userstyles.world/models"
)

func Dashboard(c *fiber.Ctx) error {
	u, _ := jwt.User(c)

	// Don't allow regular users to see this page.
	if u.Role < models.Moderator {
		return c.Render("err", fiber.Map{
			"Title": "Page not found",
			"User":  u,
		})
	}

	// Get styles.
	styles, err := models.GetAllAvailableStyles()
	if err != nil {
		return c.Render("err", fiber.Map{
			"Title": "Styles not found",
			"User":  u,
		})
	}

	sort.Slice(styles, func(i, j int) bool {
		return styles[i].ID > styles[j].ID
	})

	// Get users.
	users, err := models.FindAllUsers()
	if err != nil {
		return c.Render("err", fiber.Map{
			"Title": "Users not found",
			"User":  u,
		})
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].ID > users[j].ID
	})

	return c.Render("dashboard", fiber.Map{
		"Title":  "Dashboard",
		"User":   u,
		"Styles": styles,
		"Users":  users,
	})
}