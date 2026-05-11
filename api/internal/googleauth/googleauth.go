package googleauth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/oauth2"
)

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
}

type UserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type Client struct {
	cfg         oauth2.Config
	userInfoURL string
}

func New(cfg Config) *Client {
	return &Client{
		cfg: oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Endpoint: oauth2.Endpoint{
				AuthURL:  cfg.AuthURL,
				TokenURL: cfg.TokenURL,
			},
			Scopes: []string{"openid", "email"},
		},
		userInfoURL: cfg.UserInfoURL,
	}
}

func (c *Client) AuthCodeURL(state string) string {
	return c.cfg.AuthCodeURL(state)
}

// IsSecure reports whether the OAuth redirect URL is HTTPS, used to set the
// Secure flag on the state cookie so it's only sent over TLS in production.
func (c *Client) IsSecure() bool {
	return strings.HasPrefix(c.cfg.RedirectURL, "https://")
}

func (c *Client) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := c.cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchanging code: %w", err)
	}
	return token, nil
}

func (c *Client) GetUserInfo(ctx context.Context, token *oauth2.Token) (UserInfo, error) {
	resp, err := c.cfg.Client(ctx, token).Get(c.userInfoURL)
	if err != nil {
		return UserInfo{}, fmt.Errorf("fetching userinfo: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return UserInfo{}, fmt.Errorf("userinfo: unexpected status %d", resp.StatusCode)
	}
	var info UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return UserInfo{}, fmt.Errorf("decoding userinfo: %w", err)
	}
	return info, nil
}

// FetchUser exchanges an auth code for a token and fetches the user's profile in one step.
func (c *Client) FetchUser(ctx context.Context, code string) (id, email string, err error) {
	token, err := c.Exchange(ctx, code)
	if err != nil {
		return "", "", err
	}
	info, err := c.GetUserInfo(ctx, token)
	if err != nil {
		return "", "", err
	}
	return info.ID, info.Email, nil
}
