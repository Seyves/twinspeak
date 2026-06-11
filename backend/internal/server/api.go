package server

import (
	"encoding/json"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/google/uuid"
	"github.com/twinspeak/backend/internal/db"
	"github.com/twinspeak/backend/internal/metrics"
	"github.com/twinspeak/backend/internal/speechpipeline"
	"github.com/twinspeak/backend/internal/users"
)

const (
	internalServerError = "internal server error"
	invalidRequestBody  = "invalid request body"
)

type RestApi struct {
	host     string
	fiber    *fiber.App
	pipeline speechpipeline.Pipeline
	users    *users.Service
}

func (r *RestApi) Start() error {
	// Add logger middleware
	r.fiber.Use(logger.New())

	r.fiber.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))

	api := r.fiber.Group("/api/v1")
	api.Use(r.requestIdMiddleware)
	api.Use(r.metricsMiddleware)

	api.Post("/auth/sign-in", r.signIn)
	api.Post("/auth/sign-up", r.signUp)
	api.Get("/auth/google/sign-in", r.googleSignIn)
	api.Post("/auth/google/callback", r.googleCallback)
	api.Post("/auth/refresh", r.refresh)
	api.Post("/auth/logout", r.logout)

	api.Get("/supported-languages", r.authMiddleware, r.supportedLanguages)
	api.Get("/ws-ticket", r.authMiddleware, r.getWSTiket)
	api.Get("/ping", r.authMiddleware, r.ping)
	api.Get("/preferences", r.authMiddleware, r.getPreferences)
	api.Put("/preferences", r.authMiddleware, r.updatePreferences)
	api.Get("/messages", r.authMiddleware, r.getMessages)
	api.Get("/me", r.authMiddleware, r.me)
	api.Get("/me/credits", r.authMiddleware, r.getCredits)

	api.Get("/ws/session", r.wsAuthMiddleware, websocket.New(r.startSession))

	return r.fiber.Listen(r.host)
}

func (r *RestApi) ping(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

func (r *RestApi) supportedLanguages(c *fiber.Ctx) error {
	languages := r.pipeline.SupportedLanguages(c.Context())
	return c.JSON(languages)
}

func (r *RestApi) me(c *fiber.Ctx) error {
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

func NewRestApi(
	host string,
	pipeline speechpipeline.Pipeline,
	metrics *metrics.Module,
	users *users.Service,
) *RestApi {
	server := fiber.New(fiber.Config{
		AppName: "TwinspeakBackend",
	})

	return &RestApi{
		host:     host,
		fiber:    server,
		users:    users,
		pipeline: pipeline,
	}
}
