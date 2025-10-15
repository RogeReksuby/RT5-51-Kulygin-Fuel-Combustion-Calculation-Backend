package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"net/http"
	"repback/internal/app/ds"
	"repback/internal/app/role"
	"strings"
	"time"
)

const jwtPrefix = "Bearer "

func (h *Handler) WithAuthCheck(gCtx *gin.Context) {
	tokenString := gCtx.GetHeader("Authorization")

	if tokenString == "" {
		gCtx.Set("role", role.Guest)
		gCtx.Set("userID", uint(0))
		gCtx.Set("isModerator", false)
		gCtx.Next()
		return
	}

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

	userRole := role.FromUser(claims.UserID, claims.IsModerator)

	// Сохраняем данные пользователя в контекст для использования в handler'ах
	gCtx.Set("userID", claims.UserID)
	gCtx.Set("userLogin", claims.Login)
	gCtx.Set("isModerator", claims.IsModerator)
	gCtx.Set("userName", claims.Name)
	gCtx.Set("role", userRole)

	gCtx.Next()
}

func (h *Handler) RequireAuth() gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		userRole, exists := gCtx.Get("role")
		if !exists {
			gCtx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка проверки прав доступа",
			})
			return
		}

		if !userRole.(role.Role).IsAuthenticated() {
			gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Требуется авторизация",
			})
			return
		}

		gCtx.Next()
	}
}

func (h *Handler) RequireModerator() gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		userRole, exists := gCtx.Get("role")
		if !exists {
			gCtx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка проверки прав доступа",
			})
			return
		}

		if !userRole.(role.Role).HasModeratorAccess() {
			gCtx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Недостаточно прав. Требуется роль модератора",
			})
			return
		}

		gCtx.Next()
	}
}
