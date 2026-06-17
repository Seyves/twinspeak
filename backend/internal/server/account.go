package server

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/twinspeak/backend/internal/db"
)

func (r *RestApi) MountAccountRoutes(router fiber.Router) {
	router.Get("/account", r.getAccount)
	router.Get("/account/preferences", r.getPreferences)
	router.Put("/account/preferences", r.updatePreferences)
	router.Get("/account/messages", r.getMessages)
	router.Get("/account/credits", r.getCredits)
}

func (r *RestApi) getAccount(c *fiber.Ctx) error {
	userID := c.Locals("userId").(uuid.UUID)
	user, err := r.users.GetCurrentUser(c.Context(), userID)
	if err != nil {
		log.Errorf("Error getting current user: %s", err.Error())
		return fiber.NewError(fiber.StatusInternalServerError, internalServerError)
	}

	type response struct {
		Email          string    `json:"email"`
		ProfilePicture *string   `json:"profilePicture"`
		CreatedAt      time.Time `json:"createdAt"`
	}

	return c.JSON(response{
		Email:          user.Email,
		ProfilePicture: user.ProfilePicture,
		CreatedAt:      user.CreatedAt,
	})
}

func (r *RestApi) getPreferences(c *fiber.Ctx) error {
	userID := c.Locals("userId").(uuid.UUID)

	prefs, err := r.users.GetPreferences(c.Context(), userID)
	if err != nil {
		log.Errorf("Error getting user preferences: %s", err.Error())
		return fiber.NewError(fiber.StatusInternalServerError, internalServerError)
	}

	type response struct {
		ChatMsgSize db.Size  `json:"chatMessageSize"`
		Theme       db.Theme `json:"theme"`
		InLang      string   `json:"inLang"`
		OutLang     string   `json:"outLang"`
	}

	return c.JSON(response{
		ChatMsgSize: prefs.ChatMessageSize,
		Theme:       prefs.Theme,
		InLang:      prefs.InLang,
		OutLang:     prefs.OutLang,
	})
}

func (r *RestApi) updatePreferences(c *fiber.Ctx) error {
	type request struct {
		ChatMsgSize db.Size  `json:"chatMessageSize"`
		Theme       db.Theme `json:"theme"`
		InLang      string   `json:"inLang"`
		OutLang     string   `json:"outLang"`
	}

	var req request
	err := json.Unmarshal(c.Body(), &req)
	if err != nil {
		log.Errorf("Error unmarshalling request body: %s", err.Error())
		return fiber.NewError(fiber.StatusBadRequest, invalidRequestBody)
	}

	userID := c.Locals("userId").(uuid.UUID)

	err = r.users.UpdatePreferences(c.Context(), db.UpdateUserPrefsParams{
		UserID:          userID,
		ChatMessageSize: req.ChatMsgSize,
		Theme:           req.Theme,
		InLang:          req.InLang,
		OutLang:         req.OutLang,
	})
	if err != nil {
		log.Errorf("Error updating user preferences: %s", err.Error())
		return fiber.NewError(fiber.StatusInternalServerError, internalServerError)
	}

	return c.SendStatus(fiber.StatusOK)
}

func (r *RestApi) getMessages(c *fiber.Ctx) error {
	userID := c.Locals("userId").(uuid.UUID)

	speeches, err := r.users.GetSpeeches(c.Context(), userID)
	if err != nil {
		log.Errorf("Error getting speeches: %s", err.Error())
		return fiber.NewError(fiber.StatusInternalServerError, internalServerError)
	}

	type respItem struct {
		ID            uuid.UUID   `json:"id"`
		SendedFrom    db.ChatSide `json:"sendedFrom"`
		Status        string      `json:"status"`
		Transcription string      `json:"transcription"`
		Translation   string      `json:"translation"`
	}

	resp := make([]respItem, 0, len(speeches))
	for _, speech := range speeches {
		resp = append(resp, respItem{
			ID:            speech.ID,
			SendedFrom:    speech.ChatSide,
			Status:        "processed",
			Transcription: speech.Transcription,
			Translation:   speech.Translation,
		})
	}

	return c.JSON(resp)
}

func (r *RestApi) getCredits(c *fiber.Ctx) error {
	userID := c.Locals("userId").(uuid.UUID)
	now := time.Now()
	grants, err := r.users.GetCreditGrants(c.Context(), userID, now)
	if err != nil {
		log.Errorf("Error getting credit grants: %s", err.Error())
		return fiber.NewError(fiber.StatusInternalServerError, internalServerError)
	}

	type grantResponse struct {
		ID              string     `json:"id"`
		Amount          int32      `json:"amount"`
		RemainingAmount int32      `json:"remainingAmount"`
		Type            string     `json:"type"`
		ExpiresAt       *time.Time `json:"expiresAt"`
		CreatedAt       time.Time  `json:"createdAt"`
	}

	response := make([]grantResponse, len(grants))
	for i, g := range grants {
		response[i] = grantResponse{
			ID:              g.ID.String(),
			Amount:          g.Amount,
			RemainingAmount: g.RemainingAmount,
			Type:            string(g.Type),
			ExpiresAt:       g.ExpiresAt,
			CreatedAt:       g.CreatedAt,
		}
	}

	return c.JSON(response)
}
