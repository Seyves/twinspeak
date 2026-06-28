package server

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twinspeak/backend/internal/db"
	"github.com/twinspeak/backend/internal/email"
	"github.com/twinspeak/backend/internal/metrics"
	"github.com/twinspeak/backend/internal/service"
	"github.com/twinspeak/backend/internal/speechpipeline"
)

const (
	internalServerError = "internal server error"
	invalidRequestBody  = "invalid request body"
)

type RestApi struct {
	host     string
	fiber    *fiber.App
	pipeline speechpipeline.Pipeline
	service  *service.Service
	email    *email.Module
	db       *pgxpool.Pool
	queries  *db.Queries
}

func (r *RestApi) Start() error {
	// Add logger middleware
	r.fiber.Use(logger.New())

	r.fiber.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))

	// Public routes
	public := r.fiber.Group("/api/v1")
	public.Use(r.requestIdMiddleware)
	public.Use(r.metricsMiddleware)
	r.MountAuthRoutes(public)

	// Protected routes
	protected := public.Group("", r.authMiddleware)
	r.MountVerifyRouter(protected)

	// Email verified routes
	verified := protected.Group("", r.emailVerifiedMiddleware)
	verified.Get("/ping", r.ping)
	verified.Get("/supported-languages", r.supportedLanguages)
	r.MountAccountRoutes(verified)
	r.MountWsRoutes(verified)

	return r.fiber.Listen(r.host)
}

func (r *RestApi) Shutdown(ctx context.Context) error {
	return r.fiber.ShutdownWithContext(ctx)
}

func (r *RestApi) supportedLanguages(c *fiber.Ctx) error {
	languages := r.pipeline.SupportedLanguages(c.Context())
	return c.JSON(languages)
}

func (r *RestApi) ping(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

func NewRestApi(
	host string,
	pipeline speechpipeline.Pipeline,
	metrics *metrics.Module,
	service *service.Service,
	emailModule *email.Module,
	dbPool *pgxpool.Pool,
	queries *db.Queries,
) *RestApi {
	server := fiber.New(fiber.Config{
		AppName: "TwinspeakBackend",
	})

	return &RestApi{
		host:     host,
		fiber:    server,
		service:  service,
		pipeline: pipeline,
		email:    emailModule,
		db:       dbPool,
		queries:  queries,
	}
}
