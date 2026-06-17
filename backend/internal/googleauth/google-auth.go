package googleauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/twinspeak/backend/internal/db"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/idtoken"
	oauth2api "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

type Config struct {
	ClientId     string `mapstructure:"client-id"`
	ClientSecret string `mapstructure:"client-secret"`
	RedirectUrl  string `mapstructure:"redirect-url"`
}

type Module struct {
	config *oauth2.Config
}

var ErrGoogleAccountNotFound = errors.New("google account not found")
var ErrGoogleInvalidState = errors.New("google invalid state")
var ErrGoogleCannotExchange = errors.New("google cannot exchange")
var ErrGoogleInvalidIdToken = errors.New("google invalid id token")
var ErrEmailAlreadyTaken = errors.New("email already taken")

func (m *Module) Redirect() (url string, state string, err error) {
	state, err = generateState()
	if err != nil {
		return "", "", fmt.Errorf("cannot generate state: %w", err)
	}
	url = m.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return url, state, nil
}

func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (m *Module) CreateUser(ctx context.Context, tx *db.Queries, userInfo *oauth2api.Userinfo) (uuid.UUID, error) {
	if _, err := tx.GetUserByEmail(ctx, userInfo.Email); err == nil {
		return uuid.Nil, ErrEmailAlreadyTaken
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, fmt.Errorf("cannot select user from db by email: %w", err)
	}
	userId, err := tx.CreateAccountFromGoogle(ctx, db.CreateAccountFromGoogleParams{
		GoogleSub:      &userInfo.Id,
		Email:          userInfo.Email,
		ProfilePicture: &userInfo.Picture,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("cannot insert user in db: %w", err)
	}
	return userId, nil
}

func (m *Module) LinkExistingUser(ctx context.Context, tx *db.Queries, userInfo *oauth2api.Userinfo) (uuid.UUID, error) {
	userId, err := tx.LinkAccountToGoogle(ctx, db.LinkAccountToGoogleParams{
		GoogleSub:      &userInfo.Id,
		Email:          userInfo.Email,
		ProfilePicture: &userInfo.Picture,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("cannot update user in db: %w", err)
	}
	return userId, nil
}

func (m *Module) Callback(ctx context.Context, code string, sessionState string, state string) (*oauth2api.Userinfo, error) {
	if sessionState != state {
		return nil, ErrGoogleInvalidState
	}
	token, err := m.config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGoogleCannotExchange, err)
	}

	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, ErrGoogleInvalidIdToken
	}

	_, err = idtoken.Validate(ctx, idToken, m.config.ClientID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGoogleInvalidIdToken, err)
	}

	client := m.config.Client(context.Background(), token)

	oauth2Service, err := oauth2api.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("cannot create service: %w", err)
	}

	userInfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		return nil, fmt.Errorf("cannot get openid info: %w", err)
	}
	return userInfo, nil
}

func (m *Module) FindGoogleAccount(ctx context.Context, tx *db.Queries, googleSub string) (uuid.UUID, error) {
	userId, err := tx.FindAccountFromGoogle(ctx, &googleSub)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrGoogleAccountNotFound
	} else if err != nil {
		return uuid.Nil, fmt.Errorf("cannot select user from db: %w", err)
	}
	return userId, nil
}

func New(config Config) *Module {
	return &Module{
		config: &oauth2.Config{
			ClientID:     config.ClientId,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.RedirectUrl,
			Scopes: []string{
				oauth2api.OpenIDScope,
				oauth2api.UserinfoEmailScope,
				oauth2api.UserinfoProfileScope,
			},
			Endpoint: google.Endpoint,
		},
	}
}
