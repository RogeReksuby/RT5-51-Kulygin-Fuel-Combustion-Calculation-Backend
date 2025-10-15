package ds

import (
	"github.com/golang-jwt/jwt"
	"repback/internal/app/role"
)

type JWTClaims struct {
	UserID      uint      `json:"user_id"`
	Login       string    `json:"login"`
	IsModerator bool      `json:"is_moderator"`
	Name        string    `json:"name"`
	Role        role.Role `json:"role"`
	jwt.StandardClaims
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	User        *Users `json:"user"`
}
