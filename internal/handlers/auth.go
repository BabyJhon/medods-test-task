package handlers

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

const (
	authoriationHeader = "Authorization"
)

// Auth godoc
// @Summary      Получить access и refresh токены
// @Description  Генерирует access и refresh токены для пользователя по guid. Токены возвращаются в httpOnly cookie.
// @Tags         auth
// @Param        guid query string true "GUID пользователя" example(1e44baa9-e04b-4739-89f3-3d86b9a272ce)
// @Success      200  {string}  string "Токены успешно созданы"
// @Header       200  {string}  Set-Cookie "access_token=eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJzZXNzaW9uX2lkIjoiOTJiYmRkMTUtNmE3Ny00MWM2LTg2MzAtN2NkYWVjYmRmOGVjIiwidXNlcl9hZ2VudCI6IlBvc3RtYW5SdW50aW1lLzcuNDQuMSIsImlwIjoiMTcyLjE4LjAuMSIsImV4cCI6MTc1MTg2NDc1NCwiaWF0IjoxNzUxODYzODU0fQ.IdFs5M0cYIa1w7iBqvj5sgtvB8ufuHGpsGVUACnJxCT4MVg2nmFPPxBOamZU4nLVQOjFy9bmBqKzcNxdAb5z5Q;refresh_token=FbI-aH-FhxcYoZaZ6mdb9pBDzi--OcKL7sJT5Du04Xk"
// @Failure      400  {object}  Error "Неверный запрос"
// @Failure      500  {object}  Error "Внутренняя ошибка сервера"
// @Router       /auth [get]
func (h *Handler) auth(c *gin.Context) {
	userAgent := c.GetHeader("User-Agent")
	clientIP := c.ClientIP()

	guidString, exist := c.GetQuery("guid")
	if !exist {
		newErrorResponse(c, http.StatusBadRequest, "Empty guid query param")
		return
	}
	guid, err := uuid.FromString(guidString)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	accessToken, refreshToken, err := h.services.Auth.CreateTokens(c, guid, userAgent, net.ParseIP(clientIP))
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.SetCookie(
		"access_token",
		accessToken,
		12341000,
		"/",
		"",
		true,
		true,
	)
	c.SetCookie(
		"refresh_token",
		refreshToken,
		12341000,
		"/",
		"",
		true,
		true,
	)
}

// User godoc
// @Summary Получить UUID пользователя
// @Description Возвращает UUID пользователя по JWT токену из заголовка Authorization
// @Tags user
// @Security ApiKeyAuth
// @Produce json
// @Param Authorization header string true "Токен доступа в формате: Bearer <token>"
// @Success 200 {string} string "UUID пользователя в формате строки"
// @Failure 401 {object} Error "Неавторизованный доступ"
// @Failure 500 {object} Error "Внутренняя ошибка сервера"
// @Router /user [get]
func (h *Handler) user(c *gin.Context) {
	header := c.GetHeader(authoriationHeader)
	if header == "" {
		newErrorResponse(c, http.StatusUnauthorized, "empty request header")
		return
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 {
		newErrorResponse(c, http.StatusUnauthorized, "invalid header size")
		return
	}

	tokenClaimes, err := h.services.Auth.Parsetoken(headerParts[1])
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	session, err := h.services.GetSession(c, *tokenClaimes)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, session.UserId)
}

// Revoke godoc
// @Summary Отзыв токенов
// @Description Отзывает текущую сессию пользователя по JWT токену. Удаляет сессию из системы.
// @Tags auth
// @Security ApiKeyAuth
// @Produce json
// @Param Authorization header string true "Токен доступа в формате: Bearer <JWT>"
// @Success 200 {string} string "Пустой ответ при успешном отзыве"
// @Failure 401 {object} Error "Неавторизованный доступ: отсутствует/неверный заголовок Authorization, невалидный токен"
// @Failure 500 {object} Error "Внутренняя ошибка сервера при отзыве токена"
// @Router /revoke [post]
func (h *Handler) revoke(c *gin.Context) {
	header := c.GetHeader(authoriationHeader)
	if header == "" {
		newErrorResponse(c, http.StatusUnauthorized, "empty request header")
		return
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 {
		newErrorResponse(c, http.StatusUnauthorized, "invalid header size")
		return
	}

	tokenClaimes, err := h.services.Auth.Parsetoken(headerParts[1])
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	err = h.services.RevokeToken(c, tokenClaimes.SessionID)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, "")
}

// Refresh godoc
// @Summary Обновление токенов
// @Description Обновляет access и refresh токены с помощью текущего refresh токена из cookie
// @Tags auth
// @Accept json
// @Produce json
// @Param Cookie header string false "Токены в cookie (необходимы refresh_token и access_token)"
// @Success 200 {string} string "Токены успешно обновлены"
// @Header 200 {string} Set-Cookie "access_token=<новый_access_token>"
// @Header 200 {string} Set-Cookie "refresh_token=<новый_refresh_token>"
// @Failure 400 {object} Error "Неверные или отсутствующие токены"
// @Failure 401 {object} Error "Невалидные или просроченные токены"
// @Router /refresh [post]
func (h *Handler) refresh(c *gin.Context) {
	userAgent := c.GetHeader("User-Agent")
	clientIP := c.ClientIP()

	refreshCookie, err := c.Request.Cookie("refresh_token")
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}
	base64RefreshToken := refreshCookie.Value

	accessCookie, err := c.Request.Cookie("access_token")
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	accessToken := accessCookie.Value
	accessCookieClaims, err := h.services.Parsetoken(accessToken)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	accessToken, refreshToken, err := h.services.RefreshTokens(c, *accessCookieClaims, base64RefreshToken, userAgent, net.ParseIP(clientIP))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.SetCookie(
		"access_token",
		accessToken,
		12341,
		"/",
		"",
		true,
		true,
	)
	c.SetCookie(
		"refresh_token",
		refreshToken,
		12341,
		"/",
		"",
		true,
		true,
	)
}
