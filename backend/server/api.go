package server

import (
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/google/uuid"
	"github.com/twinspeak/backend/auth"
	"github.com/twinspeak/backend/metrics"
	"github.com/twinspeak/backend/pipeline"
)

const (
	internalServerError = "internal server error"
	invalidRequestBody  = "invalid request body"
)

type RestApi struct {
	host        string
	fiber       *fiber.App
	googleOauth *auth.GoogleOauth
	auth        *auth.Auth
	pipeline    pipeline.SpeechPipeline
	metrics     *metrics.Metrics
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
	api.Get("/me", r.authMiddleware, r.me)

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
	user, err := r.auth.GetCurrentUser(c.Context(), userID)
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

func NewRestApi(
	host string,
	googleOauth *auth.GoogleOauth,
	auth *auth.Auth,
	pipeline pipeline.SpeechPipeline,
	metrics *metrics.Metrics,
) *RestApi {
	server := fiber.New(fiber.Config{
		AppName: "TwinspeakBackend",
	})

	return &RestApi{
		host:        host,
		fiber:       server,
		googleOauth: googleOauth,
		auth:        auth,
		pipeline:    pipeline,
		metrics:     metrics,
	}
}
