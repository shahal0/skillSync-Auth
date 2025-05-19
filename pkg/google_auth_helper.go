package pkg

import (
	"context"
	"errors"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googleapi "google.golang.org/api/oauth2/v2"
)

func GetGoogleOAuthConfig(redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  redirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

func GetGoogleUserInfo(conf *oauth2.Config, code string) (*googleapi.Userinfo, error) {
	ctx := context.Background()
	token, err := conf.Exchange(ctx, code)
	if err != nil {
		return nil, errors.New("failed to exchange code for token: " + err.Error())
	}
	oauth2Service, err := googleapi.New(conf.Client(ctx, token))
	if err != nil {
		return nil, errors.New("failed to create oauth2 service: " + err.Error())
	}
	userinfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		return nil, errors.New("failed to get user info: " + err.Error())
	}
	return userinfo, nil
}
