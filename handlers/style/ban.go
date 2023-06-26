package style

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"userstyles.world/handlers/jwt"
	"userstyles.world/models"
	"userstyles.world/modules/config"
	"userstyles.world/modules/database"
	"userstyles.world/modules/email"
	"userstyles.world/modules/log"
	"userstyles.world/modules/search"
)

func BanGet(c *fiber.Ctx) error {
	u, _ := jwt.User(c)

	// Check if logged-in user has permissions.
	if !u.IsModOrAdmin() {
		c.Status(fiber.StatusUnauthorized)
		return c.Render("err", fiber.Map{
			"Title": "You are not authorized to perform this action",
			"User":  u,
		})
	}

	// Check if style exists.
	s, err := models.GetStyleByID(c.Params("id"))
	if err != nil {
		c.Status(fiber.StatusNotFound)
		return c.Render("err", fiber.Map{
			"Title": "Style not found",
			"User":  u,
		})
	}

	return c.Render("style/ban", fiber.Map{
		"Title": "Confirm ban",
		"User":  u,
		"Style": s,
	})
}

func sendBanEmail(user *models.User, entry *models.Log) error {
	modLogEntry := config.BaseURL + "/modlog#id-" + strconv.Itoa(int(entry.ID))

	args := fiber.Map{
		"Reason":  entry.Reason,
		"Message": entry.Message,
		"Link":    modLogEntry,
	}

	title := "Your style has been removed"
	return email.Send("styleremoved", user.Email, title, args)
}

func BanPost(c *fiber.Ctx) error {
	u, _ := jwt.User(c)
	id := c.Params("id")

	// Check if logged-in user has permissions.
	if !u.IsModOrAdmin() {
		c.Status(fiber.StatusUnauthorized)
		return c.Render("err", fiber.Map{
			"Title": "You are not authorized to perform this action",
			"User":  u,
		})
	}

	// Check if style exists.
	s, err := models.GetStyleByID(id)
	if err != nil {
		c.Status(fiber.StatusNotFound)
		return c.Render("err", fiber.Map{
			"Title": "Style not found",
			"User":  u,
		})
	}

	// Initialize modlog data.
	logEntry := models.Log{
		UserID:         u.ID,
		Username:       u.Username,
		Kind:           models.LogRemoveStyle,
		TargetUserName: s.Username,
		TargetData:     s.Name,
		Reason:         strings.TrimSpace(c.FormValue("reason")),
		Message:        strings.TrimSpace(c.FormValue("message")),
		Censor:         c.FormValue("censor") == "on",
	}

	// Add banned style log entry.
	modlog := new(models.Log)
	if err := modlog.AddLog(&logEntry); err != nil {
		log.Warn.Printf("Failed to add style %d to ModLog: %s", s.ID, err.Error())
		return c.Render("err", fiber.Map{
			"Title": "Internal server error",
			"User":  u,
		})
	}

	// Delete style from database.
	q := new(models.Style)
	if err = database.Conn.Delete(q, "styles.id = ?", s.ID).Error; err != nil {
		log.Warn.Printf("Failed to delete style %d: %s\n", s.ID, err.Error())
		c.Status(fiber.StatusInternalServerError)
		return c.Render("err", fiber.Map{
			"Title": "Internal server error",
			"User":  u,
		})
	}

	// Delete stats from database.
	if err = new(models.Stats).Delete(s.ID); err != nil {
		log.Warn.Printf("Failed to delete stats for style %d: %s\n", s.ID, err.Error())
		c.Status(fiber.StatusInternalServerError)
		return c.Render("err", fiber.Map{
			"Title": "Internal server error",
			"User":  u,
		})
	}

	if err = models.RemoveStyleCode(strconv.Itoa(int(s.ID))); err != nil {
		log.Warn.Printf("kind=removecode id=%v err=%q\n", s.ID, err)
	}

	if err = search.DeleteStyle(s.ID); err != nil {
		log.Warn.Printf("Failed to delete style %d from index: %s", s.ID, err)
	}

	go func(style *models.APIStyle, entry models.Log) {
		// Add notification to database.
		notification := models.Notification{
			Seen:     false,
			Kind:     models.KindBannedStyle,
			TargetID: int(entry.ID),
			UserID:   int(u.ID),
			StyleID:  int(s.ID),
		}

		if err := notification.Create(); err != nil {
			log.Warn.Printf("Failed to create a notification for ban removal %d: %v\n", style.ID, err)
		}

		targetUser, err := models.FindUserByID(strconv.Itoa(int(style.UserID)))
		if err != nil {
			log.Warn.Printf("Failed to find user %d: %s", style.UserID, err.Error())
			return
		}

		// Notify the author about style removal.
		if err := sendBanEmail(targetUser, &logEntry); err != nil {
			log.Warn.Printf("Failed to email author for style %d: %s\n", style.ID, err)
		}
	}(s, logEntry)

	return c.Redirect("/modlog", fiber.StatusSeeOther)
}
