package server

import (
	"bytes"
	"encoding/json"
	"net/netip"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/google/uuid"
	"github.com/twinspeak/backend/auth"
	"github.com/twinspeak/backend/providers"
)

type RestApi struct {
	host        string
	fiber       *fiber.App
	googleOauth *auth.GoogleOauth
	auth        *auth.Auth
	transcriber providers.Transcriber
	translater  providers.Translater
}

func (r *RestApi) Start() error {
	// Add logger middleware
	r.fiber.Use(logger.New())

	r.fiber.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))

	api := r.fiber.Group("/api/v1")

	auth := api.Group("/auth")
	auth.Post("/refresh", r.refresh)
	auth.Get("/google/sign-in", r.googleSignIn)
	auth.Post("/google/callback", r.googleCallback)

	papi := api.Group("/", r.authMiddleware)
	papi.Post("/process-speech", r.processSpeech)
	papi.Get("/ping", r.ping)

	return r.fiber.Listen(r.host)
}

func getSecureCookie(key string, value string) *fiber.Cookie {
	cookie := new(fiber.Cookie)
	cookie.Name = key
	cookie.Value = value
	cookie.HTTPOnly = true
	cookie.Secure = true
	cookie.SameSite = fiber.CookieSameSiteStrictMode
	return cookie
}

func (r *RestApi) authMiddleware(c *fiber.Ctx) error {
	accessToken := c.Cookies("access_token")
	token, err := r.auth.ValidateAccessToken(c.Context(), time.Now(), accessToken)
	if err != nil {
		log.Errorf("Error in auth middleware: %s", err.Error())
		return fiber.NewError(fiber.StatusUnauthorized, "invalid JWT")
	}
	c.Locals("token", token)
	return c.Next()
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

func (r *RestApi) refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return fiber.NewError(fiber.StatusUnauthorized)
	}

	userAgent := string(c.Context().UserAgent())
	ip, _ := netip.ParseAddr(c.IP())

	accessToken, refreshToken, err := r.auth.RotateAccessToken(
		c.Context(),
		time.Now(),
		refreshToken,
		userAgent,
		&ip,
	)
	if err != nil {
		log.Errorf("Error while rotating access token: %s", err.Error())
		return fiber.NewError(fiber.StatusUnauthorized)
	}

	c.Cookie(getSecureCookie("refresh_token", refreshToken))
	c.Cookie(getSecureCookie("access_token", accessToken))

	return c.SendStatus(fiber.StatusOK)
}

func (r *RestApi) ping(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

func (r *RestApi) googleSignIn(c *fiber.Ctx) error {
	return c.Redirect(r.googleOauth.GetSignInUrl(), fiber.StatusTemporaryRedirect)
}

func (r *RestApi) googleCallback(c *fiber.Ctx) error {
	type request struct {
		Code  string `json:"code"`
		State string `json:"state"`
	}

	var req request
	err := json.Unmarshal(c.Request().Body(), &req)
	if err != nil {
		log.Errorf("Error unmarshalling request body: %s", err.Error())
		return fiber.NewError(fiber.StatusBadRequest)
	}

	userAgent := string(c.Context().UserAgent())
	ip, _ := netip.ParseAddr(c.IP())

	accessToken, refreshToken, err := r.googleOauth.ProcessRedirect(c.Context(), req.Code, req.State, userAgent, &ip)
	if err != nil {
		log.Errorf("Error while processing redirect: %s", err.Error())
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	c.Cookie(getSecureCookie("refresh_token", refreshToken))
	c.Cookie(getSecureCookie("access_token", accessToken))

	return c.SendStatus(fiber.StatusOK)
}

func NewRestApi(
	host string,
	googleOauth *auth.GoogleOauth,
	auth *auth.Auth,
	transcriber providers.Transcriber,
	translater providers.Translater,
) *RestApi {
	server := fiber.New(fiber.Config{
		AppName: "TwinspeakBackend",
	})

	return &RestApi{
		host:        host,
		fiber:       server,
		googleOauth: googleOauth,
		auth:        auth,
		transcriber: transcriber,
		translater:  translater,
	}
}
