package server

import (
	"bytes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/google/uuid"
	"github.com/twinspeak/backend/providers"
)

type RestApi struct {
	transcriber providers.Transcriber
	translater  providers.Translater
	host        string
	fiber       *fiber.App
}

func (r *RestApi) Start() error {
	// Add logger middleware
	r.fiber.Use(logger.New())

	r.fiber.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))

	r.fiber.Get("/healthz", r.healthcheck)
	r.fiber.Post("/process-speech", r.processSpeech)

	return r.fiber.Listen(r.host)
}

func (r *RestApi) healthcheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

func (r *RestApi) processSpeech(c *fiber.Ctx) error {
	speechId := uuid.New()

	type response struct {
		Id            uuid.UUID `json:"id"`
		Transcription string    `json:"transcription"`
		Translation   string    `json:"translation"`
	}

	inputLang := c.Query("inputLang")
	if inputLang == "" {
		return fiber.NewError(400, "No inputLang query provided")
	}

	outputLang := c.Query("outputLang")
	if outputLang == "" {
		return fiber.NewError(400, "No outputLang query provided")
	}

	body := c.Request().Body()

	transcription, err := r.transcriber.Transcribe(inputLang, c.Get("Content-Type"), c.Get("Content-Length"), bytes.NewReader(body))
	if err != nil {
		log.Errorf("Error while transcribing: %s", err.Error())
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	if transcription == "" {
		return c.JSON(response{
			Id:            speechId,
			Transcription: transcription,
			Translation:   "",
		})
	}

	translation, err := r.translater.Translate(inputLang, outputLang, transcription)
	if err != nil {
		log.Errorf("Error while translating: %s", err.Error())
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	return c.JSON(response{
		Id:            speechId,
		Transcription: transcription,
		Translation:   translation,
	})
}

func NewRestApi(
	transcriber providers.Transcriber,
	translater providers.Translater,
	host string,
) *RestApi {
	server := fiber.New(fiber.Config{
		AppName: "TwinspeakBackend",
	})

	return &RestApi{
		transcriber: transcriber,
		translater:  translater,
		fiber:       server,
		host:        host,
	}
}
