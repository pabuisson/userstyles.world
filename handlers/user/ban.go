package user

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"userstyles.world/handlers/jwt"
	"userstyles.world/models"
	"userstyles.world/modules/config"
	"userstyles.world/modules/log"
	"userstyles.world/utils"
)

func Ban(c *fiber.Ctx) error {
	u, _ := jwt.User(c)
	id := c.Params("id") // TODO: Switch to int type.

	if !u.IsModOrAdmin() {
		return c.Render("err", fiber.Map{
			"Title": "Unauthorized",
			"User":  u,
		})
	}

	user, err := models.FindUserByID(id)
	if err != nil {
		return c.Render("err", fiber.Map{
			"Title": "User ID doesn't exist",
			"User":  u,
		})
	}

	if u.ID == user.ID {
		return c.Render("err", fiber.Map{
			"Title": "You can't ban yourself",
			"User":  u,
		})
	}

	return c.Render("user/ban", fiber.Map{
		"Title":  "Ban user",
		"User":   u,
		"Params": user,
	})
}

func sendBanEmail(c *fiber.Ctx, baseURL string, user *models.User, modLogID uint, reason string) error {
	modLogEntry := baseURL + "/modlog#id-" + strconv.Itoa(int(modLogID))

	args := fiber.Map{
		"Reason": reason,
		"Link":   modLogEntry,
	}

	var bufText bytes.Buffer
	var bufHTML bytes.Buffer
	errText := c.App().Config().Views.Render(&bufText, "email/userbanned.text", args)
	errHTML := c.App().Config().Views.Render(&bufHTML, "email/userbanned.html", args)
	if errText != nil || errHTML != nil {
		return c.Status(fiber.StatusInternalServerError).Render("err", fiber.Map{
			"Title": "Internal server error",
			"Error": "Failed to render email templates.",
		})
	}

	err := utils.NewEmail().
		SetTo(user.Email).
		SetSubject("You have been banned").
		AddPart(*utils.NewPart().SetBody(bufText.String())).
		AddPart(*utils.NewPart().SetBody(bufHTML.String()).HTML()).
		SendEmail(config.IMAPServer)
	if err != nil {
		return err
	}
	return nil
}

func ConfirmBan(c *fiber.Ctx) error {
	u, _ := jwt.User(c)
	stringID := c.Params("id")
	id, _ := strconv.Atoi(stringID)

	if !u.IsModOrAdmin() {
		return c.Render("err", fiber.Map{
			"Title": "Unauthorized",
			"User":  u,
		})
	}

	if u.ID == uint(id) {
		return c.Render("err", fiber.Map{
			"Title": "You can't ban yourself",
			"User":  u,
		})
	}

	// Check if user exists.
	targetUser, err := models.FindUserByID(stringID)
	if err != nil {
		return c.Render("err", fiber.Map{
			"Title": "User ID doesn't exist",
			"User":  u,
		})
	}

	// Delete from database.
	user := new(models.User)
	if err := user.DeleteWhereID(targetUser.ID); err != nil {
		log.Warn.Printf("Failed to ban user %d: %s\n", id, err.Error())
		return c.Render("err", fiber.Map{
			"Title": "Internal server error",
			"User":  u,
		})
	}

	// Delete user's styles.
	styles := new(models.Style)
	if err := styles.BanWhereUserID(targetUser.ID); err != nil {
		log.Warn.Printf("Failed to ban styles from user %d: %s\n", id, err.Error())
		return c.Render("err", fiber.Map{
			"Title": "Internal server error",
			"User":  u,
		})
	}

	// Initialize modlog data.
	logEntry := models.Log{
		UserID:         u.ID,
		Username:       u.Username,
		Kind:           models.LogBanUser,
		TargetUserName: targetUser.Username,
		Reason:         strings.TrimSpace(c.FormValue("reason")),
		Censor:         c.FormValue("censor") == "on",
	}

	// Add banned user log entry.
	modlog := new(models.Log)
	if err := modlog.AddLog(&logEntry); err != nil {
		log.Warn.Printf("Failed to add user %d to ModLog: %s\n", targetUser.ID, err.Error())
		return c.Render("err", fiber.Map{
			"Title": "Internal server error",
			"User":  u,
		})
	}

	go func(baseURL string, user *models.User, modLogID uint, reason string) {
		// Send a email about they've been banned.
		if err := sendBanEmail(c, baseURL, targetUser, modLogID, reason); err != nil {
			log.Warn.Printf("Failed to send an email to user %d: %s", user.ID, err.Error())
		}
	}(c.BaseURL(), targetUser, logEntry.ID, logEntry.Reason)

	return c.Redirect("/dashboard")
}
