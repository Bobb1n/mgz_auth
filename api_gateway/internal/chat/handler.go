package chat

import (
	"net/http"
	"strconv"

	"api_gateway/internal/middleware"
	"api_gateway/pkg/grpcerr"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	client *Client
}

func NewHandler(client *Client) *Handler {
	return &Handler{client: client}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	g := e.Group("/v1")

	g.POST("/chats/direct", h.createDirectChat)
	g.POST("/messages", h.sendMessage)
	g.GET("/chats", h.listUserChats)
}

type createDirectChatRequest struct {
	OtherUserID string `json:"other_user_id"`
}

func (h *Handler) createDirectChat(c echo.Context) error {
	var req createDirectChatRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	if req.OtherUserID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "other_user_id is required"})
	}

	currentUserID := c.Request().Header.Get(middleware.HeaderUserID)
	if currentUserID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing user id"})
	}

	chat, err := h.client.CreateDirectChat(c.Request().Context(), currentUserID, req.OtherUserID)
	if err != nil {
		code, msg := grpcerr.HTTPStatus(err)
		return c.JSON(code, map[string]string{"error": msg})
	}
	return c.JSON(http.StatusOK, chat)
}

type sendMessageRequest struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

func (h *Handler) sendMessage(c echo.Context) error {
	var req sendMessageRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	if req.ChatID == "" || req.Text == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "chat_id and text are required"})
	}

	currentUserID := c.Request().Header.Get(middleware.HeaderUserID)
	if currentUserID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing user id"})
	}

	msg, err := h.client.SendMessage(c.Request().Context(), req.ChatID, currentUserID, req.Text)
	if err != nil {
		code, m := grpcerr.HTTPStatus(err)
		return c.JSON(code, map[string]string{"error": m})
	}
	return c.JSON(http.StatusOK, msg)
}

func (h *Handler) listUserChats(c echo.Context) error {
	currentUserID := c.Request().Header.Get(middleware.HeaderUserID)
	if currentUserID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing user id"})
	}

	limit := int32(15)
	offset := int32(0)

	if v := c.QueryParam("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = int32(n)
		}
	}
	if v := c.QueryParam("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = int32(n)
		}
	}

	chats, err := h.client.ListUserChats(c.Request().Context(), currentUserID, limit, offset)
	if err != nil {
		code, m := grpcerr.HTTPStatus(err)
		return c.JSON(code, map[string]string{"error": m})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"chats":  chats,
		"limit":  limit,
		"offset": offset,
	})
}

