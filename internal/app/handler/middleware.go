package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"net/http"
	"repback/internal/app/ds"
	"strings"
	"time"
)

const jwtPrefix = "Bearer "

func (h *Handler) WithAuthCheck(gCtx *gin.Context) {
	tokenString := gCtx.GetHeader("Authorization")
	if !strings.HasPrefix(tokenString, jwtPrefix) {
		gCtx.AbortWithStatus(http.StatusForbidden)
		return
	}

	jwtConfig := h.Config.GetJWTConfig()

	tokenString = tokenString[len(jwtPrefix):]
	claims := &ds.JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Проверяем что алгоритм подписи тот который мы используем
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(jwtConfig.SecretKey), nil
	})

	// Проверяем ошибки парсинга
	if err != nil {
		gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Неверный токен: " + err.Error(),
		})
		return
	}

	// Проверяем что токен валиден
	if !token.Valid {
		gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Токен невалиден",
		})
		return
	}

	// Проверяем что токен не истек
	if claims.ExpiresAt < time.Now().Unix() {
		gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Токен истек",
		})
		return
	}

	// Сохраняем данные пользователя в контекст для использования в handler'ах
	gCtx.Set("userID", claims.UserID)
	gCtx.Set("userLogin", claims.Login)
	gCtx.Set("isModerator", claims.IsModerator)
	gCtx.Set("userName", claims.Name)

	gCtx.Next()
}
