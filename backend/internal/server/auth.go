package server

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/twinspeak/backend/internal/auth"
	"github.com/twinspeak/backend/internal/googleauth"
)

const (
	accessTokenCookie     = "access_token"
	refreshTokenCookie    = "refresh_token"
	emailUnverifiedCookie = "email_unverified"
	sessionStateCookie    = "session_state"
)

func (r *RestApi) MountAuthRoutes(router fiber.Router) {
	router.Post("/auth/sign-in", r.signIn)
	router.Post("/auth/sign-up", r.signUp)
	router.Post("/auth/refresh", r.refresh)
	router.Post("/auth/logout", r.logout)
	router.Get("/auth/google/sign-in", r.googleSignIn)
	router.Post("/auth/google/callback", r.googleCallback)
}

func (r *RestApi) signIn(c *fiber.Ctx) error {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req request
	err := json.Unmarshal(c.Request().Body(), &req)
	if err != nil {
		log.Errorf("Error unmarshalling request body: %s", err.Error())
		return fiber.NewError(fiber.StatusBadRequest, invalidRequestBody)
	}

	// userAgent := string(c.Context().UserAgent())
	// ip, _ := netip.ParseAddr(c.IP())

	accessToken, refreshToken, userId, err := r.users.SignIn(c.Context(), time.Now(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
		} else {
			log.Errorf("Error during sign in: %s", err.Error())
			return fiber.NewError(fiber.StatusInternalServerError, internalServerError)
		}
	}

	// Get user to check email verification status
	user, err := r.users.GetCurrentUser(c.Context(), userId)
	if err != nil {
		log.Errorf("Error getting user: %s", err.Error())
		return fiber.NewError(fiber.StatusInternalServerError, internalServerError)
	}

	c.Cookie(getSecureCookie(refreshTokenCookie, refreshToken.Value, refreshToken.ExpiresAt))
	c.Cookie(getSecureCookie(accessTokenCookie, accessToken.Value, accessToken.ExpiresAt))
	c.Cookie(getEmailUnverifiedCookie(!user.EmailVerified, accessToken.ExpiresAt))

	return c.SendStatus(fiber.StatusOK)
}

func (r *RestApi) signUp(c *fiber.Ctx) error {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req request
	err := json.Unmarshal(c.Request().Body(), &req)
	if err != nil {
		log.Errorf("Error unmarshalling request body: %s", err.Error())
		return fiber.NewError(fiber.StatusBadRequest, invalidRequestBody)
	}

	// userAgent := string(c.Context().UserAgent())
	// ip, _ := netip.ParseAddr(c.IP())

	accessToken, refreshToken, err := r.users.SignUp(c.Context(), time.Now(), req.Email, req.Password)
	if errors.Is(err, auth.ErrEmailAlreadyTaken) {
		return fiber.NewError(fiber.StatusConflict, "email already taken")
	} else if err != nil {
		log.Errorf("Error during sign up: %s", err.Error())
		return fiber.NewError(fiber.StatusInternalServerError, internalServerError)
	}
	c.Cookie(getSecureCookie(refreshTokenCookie, refreshToken.Value, refreshToken.ExpiresAt))
	c.Cookie(getSecureCookie(accessTokenCookie, accessToken.Value, accessToken.ExpiresAt))
	c.Cookie(getEmailUnverifiedCookie(true, accessToken.ExpiresAt)) // Always true for new signups

	return c.SendStatus(fiber.StatusOK)
}

func (r *RestApi) refresh(c *fiber.Ctx) error {
	refreshTokenStr := c.Cookies(refreshTokenCookie)
	if refreshTokenStr == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid refresh token")
	}

	// userAgent := string(c.Context().UserAgent())
	// ip, _ := netip.ParseAddr(c.IP())
	//
	accessToken, refreshToken, userId, err := r.users.RotateSession(
		c.Context(),
		time.Now(),
		refreshTokenStr,
	)
	if err != nil {
		if errors.Is(err, auth.ErrMaliciousSuspicion) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid refresh token")
		} else {
			log.Errorf("Error during refreshing token: %s", err.Error())
			return fiber.NewError(fiber.StatusInternalServerError, internalServerError)
		}
	}

	// Get user to check email verification status
	user, err := r.users.GetCurrentUser(c.Context(), userId)
	if err != nil {
		log.Errorf("Error getting user: %s", err.Error())
		return fiber.NewError(fiber.StatusInternalServerError, internalServerError)
	}

	c.Cookie(getSecureCookie(refreshTokenCookie, refreshToken.Value, refreshToken.ExpiresAt))
	c.Cookie(getSecureCookie(accessTokenCookie, accessToken.Value, accessToken.ExpiresAt))
	c.Cookie(getEmailUnverifiedCookie(!user.EmailVerified, accessToken.ExpiresAt))

	return c.SendStatus(fiber.StatusOK)
}

func (r *RestApi) logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies(refreshTokenCookie)
	if refreshToken != "" {
		_ = r.users.Logout(c.Context(), refreshToken)
	}

	// Expire cookies by setting them with a past expiration date
	expiredTime := time.Unix(0, 0)
	c.Cookie(getSecureCookie(accessTokenCookie, "", expiredTime))
	c.Cookie(getSecureCookie(refreshTokenCookie, "", expiredTime))

	return c.SendStatus(fiber.StatusOK)
}

func (r *RestApi) googleSignIn(c *fiber.Ctx) error {
	url, state, err := r.users.GoogleRedirect()
	if err != nil {
		log.Errorf("Error generation sign-in url: %s", err.Error())
		return fiber.NewError(fiber.StatusInternalServerError, internalServerError)
	}
	stateExpiresAt := time.Now().Add(time.Minute * 10)
	c.Cookie(getSecureCookie("session_state", state, stateExpiresAt))
	return c.Redirect(url, fiber.StatusTemporaryRedirect)
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
		return fiber.NewError(fiber.StatusBadRequest, invalidRequestBody)
	}

	sessionState := c.Cookies("session_state")
	// userAgent := string(c.Context().UserAgent())
	// ip, _ := netip.ParseAddr(c.IP())

	accessToken, refreshToken, err := r.users.GoogleCallback(c.Context(), time.Now(), req.Code, sessionState, req.State)
	if errors.Is(err, googleauth.ErrGoogleInvalidState) ||
		errors.Is(err, googleauth.ErrGoogleCannotExchange) ||
		errors.Is(err, googleauth.ErrGoogleInvalidIdToken) {
		log.Errorf("Error while processing redirect: %s", err.Error())
		return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired creds")
	} else if err != nil {
		log.Errorf("Error while processing redirect: %s", err.Error())
		return fiber.NewError(fiber.StatusInternalServerError, internalServerError)
	}

	c.Cookie(getSecureCookie(refreshTokenCookie, refreshToken.Value, refreshToken.ExpiresAt))
	c.Cookie(getSecureCookie(accessTokenCookie, accessToken.Value, accessToken.ExpiresAt))
	c.Cookie(getEmailUnverifiedCookie(false, accessToken.ExpiresAt)) // Google users are auto-verified
	// Killing session state cookie
	c.Cookie(getSecureCookie(sessionStateCookie, "", time.Unix(0, 0)))

	return c.SendStatus(fiber.StatusOK)
}

func getSecureCookie(key string, value string, expiresAt time.Time) *fiber.Cookie {
	cookie := new(fiber.Cookie)
	cookie.Name = key
	cookie.Value = value
	cookie.HTTPOnly = true
	cookie.Secure = true
	cookie.SameSite = fiber.CookieSameSiteLaxMode
	cookie.Expires = expiresAt
	cookie.Path = "/"
	return cookie
}

func getEmailUnverifiedCookie(unverified bool, expiresAt time.Time) *fiber.Cookie {
	value := "false"
	if unverified {
		value = "true"
	}

	cookie := new(fiber.Cookie)
	cookie.Name = emailUnverifiedCookie
	cookie.Value = value
	cookie.HTTPOnly = false // Frontend needs to read this
	cookie.Secure = true
	cookie.SameSite = fiber.CookieSameSiteLaxMode
	cookie.Expires = expiresAt
	cookie.Path = "/"
	return cookie
}
