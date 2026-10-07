package api

import (
	"context"
)

type PKCETokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type PKCEExchangeRequest struct {
	Code         string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
}

type PKCERefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (c *Client) PKCEExchange(ctx context.Context, req PKCEExchangeRequest) (PKCETokenResponse, error) {
	var out PKCETokenResponse
	if err := c.DoJSON(ctx, "POST", "auth/pkce/exchange", req, &out); err != nil {
		return PKCETokenResponse{}, err
	}
	return out, nil
}

func (c *Client) PKCERefresh(ctx context.Context, refreshToken string) (PKCETokenResponse, error) {
	var out PKCETokenResponse
	req := PKCERefreshRequest{RefreshToken: refreshToken}
	if err := c.DoJSON(ctx, "POST", "auth/pkce/refresh", req, &out); err != nil {
		return PKCETokenResponse{}, err
	}
	return out, nil
}
