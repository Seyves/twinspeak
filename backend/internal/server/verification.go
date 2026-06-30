package server

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/twinspeak/backend/internal/email"
)

func (r *RestApi) MountVerifyRouter(router fiber.Router) {
	router.Get("/verification/verify", r.verifyEmail)
	router.Post("/verification/resend", r.resendVerification)
}

func (r *RestApi) verifyEmail(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return fiber.NewError(fiber.StatusBadRequest, "token required")
	}
	userId, err := r.service.VerifyEmail(c.Context(), token)
	if errors.Is(err, email.ErrInvalidVerificationToken) {
		return fiber.NewError(fiber.StatusForbidden, "invalid verification token")
	} else if err != nil {
		log.Errorf("Error verifying email: %s", err.Error())
		return fiber.NewError(fiber.StatusBadRequest, "invalid or expired verification token")
	}
	// Clear email_unverified cookie
	c.Cookie(getEmailUnverifiedCookie(false, time.Now().Add(time.Minute*10)))

	return c.JSON(fiber.Map{
		"success": true,
		"userId":  userId.String(),
	})
}

func (r *RestApi) resendVerification(c *fiber.Ctx) error {
	userId := c.Locals("userId").(uuid.UUID)
	err := r.service.ResendVerificationEmail(c.Context(), userId)
	if err != nil {
		log.Errorf("Error sending verification email: %s", err.Error())
		return fiber.NewError(fiber.StatusInternalServerError, "failed to send verification email")
	}
	return c.SendStatus(fiber.StatusOK)
}
