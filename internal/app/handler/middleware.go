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

func (h *Handler) WithAuthCheck(allowedRoles ...role.Role) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		jwtStr := gCtx.GetHeader("Authorization")
		if !strings.HasPrefix(jwtStr, jwtPrefix) {
			gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Требуется авторизация",
			})
			return
		}

		// Отрезаем префикс
		jwtStr = jwtStr[len(jwtPrefix):]

		claims := &ds.JWTClaims{}
		token, err := jwt.ParseWithClaims(jwtStr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
			}
			return []byte(h.Config.JWTSecretKey), nil
		})

		if err != nil {
			gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Неверный токен: " + err.Error(),
			})
			return
		}

		if !token.Valid {
			gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Токен невалиден",
			})
			return
		}

		if claims.ExpiresAt < time.Now().Unix() {
			gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Токен истек",
			})
			return
		}

		// 🔥 ИСПРАВЛЕННАЯ ЛОГИКА ПРОВЕРКИ РОЛЕЙ:
		// Если указаны allowedRoles, проверяем что роль пользователя входит в список разрешенных
		if len(allowedRoles) > 0 {
			roleAllowed := false
			for _, allowedRole := range allowedRoles {
				if claims.Role == allowedRole {
					roleAllowed = true
					break
				}
			}

			if !roleAllowed {
				gCtx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": fmt.Sprintf("Недостаточно прав. Ваша роль: %s, требуемые: %v", claims.Role, allowedRoles),
				})
				return
			}
		}

		// Сохраняем данные пользователя в контекст
		gCtx.Set("userID", claims.UserID)
		gCtx.Set("userLogin", claims.Login)
		gCtx.Set("isModerator", claims.IsModerator)
		gCtx.Set("userName", claims.Name)
		gCtx.Set("role", claims.Role)

		gCtx.Next()
	}
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
