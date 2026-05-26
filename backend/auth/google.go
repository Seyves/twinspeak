package auth

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/twinspeak/backend/db"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/idtoken"
	oauth2api "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

type GoogleOauthConfig struct {
	ClientId     string `mapstructure:"client-id"`
	ClientSecret string `mapstructure:"client-secret"`
	RedirectUrl  string `mapstructure:"redirect-url"`
}

type GoogleOauth struct {
	config     *oauth2.Config
	hmacSecret string
	queries    *db.Queries
}

func (g *GoogleOauth) GetSignInUrl() string {
	return g.config.AuthCodeURL(
		// TODO add generate state
		"todo",
		oauth2.AccessTypeOffline,
	)
}

func (g *GoogleOauth) CreateNewUser(ctx context.Context, userInfo *oauth2api.Userinfo) (uuid.UUID, error) {
	userId, err := g.queries.CreateAccountFromGoogle(ctx, db.CreateAccountFromGoogleParams{
		GoogleSub:      &userInfo.Id,
		Email:          userInfo.Email,
		ProfilePicture: &userInfo.Picture,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("cannot insert user in db: %w", err)
	}
	return userId, nil
}

func (g *GoogleOauth) ProcessRedirect(
	ctx context.Context,
	code string,
	state string,
	userAgent string,
	ip *netip.Addr,
) (accessToken *token, refreshToken *token, err error) {
	if state != "todo" {
		return nil, nil, fmt.Errorf("invalid state value")
	}
	token, err := g.config.Exchange(context.Background(), code)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot exchange code: %w", err)
	}

	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, nil, errors.New("id token is not a string")
	}

	_, err = idtoken.Validate(ctx, idToken, g.config.ClientID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid id token: %w", err)
	}

	client := g.config.Client(context.Background(), token)

	oauth2Service, err := oauth2api.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create service: %w", err)
	}

	userInfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot get openid info: %w", err)
	}

	userId, err := g.queries.FindAccountFromGoogle(ctx, &userInfo.Id)
	if errors.Is(err, pgx.ErrNoRows) {
		userId, err = g.CreateNewUser(ctx, userInfo)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot insert user into db: %w", err)
		}
	} else if err != nil {
		return nil, nil, fmt.Errorf("cannot get user from db: %w", err)
	}

	refreshToken, err = createRefreshToken(ctx, time.Now(), g.queries, userId, userAgent, ip)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create refresh token: %w", err)
	}

	accessToken, err = createAccessToken(ctx, time.Now(), g.hmacSecret, userId)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create access token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func NewGoogleOauth(config GoogleOauthConfig, queries *db.Queries, hmacSecret string) *GoogleOauth {
	return &GoogleOauth{
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
		hmacSecret: hmacSecret,
		queries:    queries,
	}
}
