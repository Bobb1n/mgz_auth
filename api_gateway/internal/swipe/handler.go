package swipe

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"api_gateway/pkg/grpcerr"
	swipev1 "swipe-mgz/pkg/api/swipe/v1"

	"github.com/labstack/echo/v4"
)

const headerUserID = "X-User-Id"

type Handler struct {
	client *Client
}

func NewHandler(client *Client) *Handler {
	return &Handler{client: client}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	v1 := e.Group("/v1")
	v1.POST("/swipes", h.swipe)
	v1.GET("/matches", h.listMatches)
	v1.PUT("/location", h.updateLocation)
	v1.GET("/candidates", h.getCandidates)
}

type swipeRequest struct {
	SwipeeID  string `json:"swipee_id"`
	Direction string `json:"direction"`
}

type updateLocationRequest struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

func (h *Handler) swipe(c echo.Context) error {
	userID, ok := currentUserID(c)
	if !ok {
		return errJSON(c, http.StatusUnauthorized, "missing user id")
	}

	var req swipeRequest
	if err := c.Bind(&req); err != nil {
		return errJSON(c, http.StatusBadRequest, "invalid json")
	}

	dir, err := parseDirection(req.Direction)
	if err != nil {
		return errJSON(c, http.StatusBadRequest, err.Error())
	}

	resp, err := h.client.api.Swipe(c.Request().Context(), &swipev1.SwipeRequest{
		SwiperId:  userID,
		SwipeeId:  req.SwipeeID,
		Direction: dir,
	})
	if err != nil {
		return writeGRPCErr(c, err)
	}
	return c.JSON(http.StatusOK, swipeResponseJSON(resp))
}

func (h *Handler) listMatches(c echo.Context) error {
	userID, ok := currentUserID(c)
	if !ok {
		return errJSON(c, http.StatusUnauthorized, "missing user id")
	}

	limit, offset := int32(20), int32(0)
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

	resp, err := h.client.api.ListMatches(c.Request().Context(), &swipev1.ListMatchesRequest{
		UserId: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return writeGRPCErr(c, err)
	}

	matches := make([]map[string]interface{}, 0, len(resp.GetMatches()))
	for _, m := range resp.GetMatches() {
		matches = append(matches, matchJSON(m))
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"matches": matches,
		"limit":   resp.GetLimit(),
		"offset":  resp.GetOffset(),
	})
}

func (h *Handler) updateLocation(c echo.Context) error {
	userID, ok := currentUserID(c)
	if !ok {
		return errJSON(c, http.StatusUnauthorized, "missing user id")
	}

	var req updateLocationRequest
	if err := c.Bind(&req); err != nil {
		return errJSON(c, http.StatusBadRequest, "invalid json")
	}

	_, err := h.client.api.UpdateLocation(c.Request().Context(), &swipev1.UpdateLocationRequest{
		UserId:    userID,
		Longitude: req.Longitude,
		Latitude:  req.Latitude,
	})
	if err != nil {
		return writeGRPCErr(c, err)
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) getCandidates(c echo.Context) error {
	userID, ok := currentUserID(c)
	if !ok {
		return errJSON(c, http.StatusUnauthorized, "missing user id")
	}
	resp, err := h.client.api.GetCandidates(c.Request().Context(), &swipev1.GetCandidatesRequest{UserId: userID})
	if err != nil {
		return writeGRPCErr(c, err)
	}

	cands := make([]map[string]interface{}, 0, len(resp.GetCandidates()))
	for _, cd := range resp.GetCandidates() {
		cands = append(cands, map[string]interface{}{
			"user_id":     cd.GetUserId(),
			"longitude":   cd.GetLongitude(),
			"latitude":    cd.GetLatitude(),
			"distance_km": cd.GetDistanceKm(),
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"candidates": cands})
}

func currentUserID(c echo.Context) (string, bool) {
	id := c.Request().Header.Get(headerUserID)
	return id, id != ""
}

func errJSON(c echo.Context, code int, msg string) error {
	return c.JSON(code, map[string]string{"error": msg})
}

func writeGRPCErr(c echo.Context, err error) error {
	code, msg := grpcerr.HTTPStatus(err)
	return errJSON(c, code, msg)
}

func parseDirection(s string) (swipev1.Direction, error) {
	switch strings.ToLower(s) {
	case "like":
		return swipev1.Direction_DIRECTION_LIKE, nil
	case "dislike":
		return swipev1.Direction_DIRECTION_DISLIKE, nil
	default:
		return swipev1.Direction_DIRECTION_UNSPECIFIED, errors.New("direction must be 'like' or 'dislike'")
	}
}

func directionString(d swipev1.Direction) string {
	switch d {
	case swipev1.Direction_DIRECTION_LIKE:
		return "like"
	case swipev1.Direction_DIRECTION_DISLIKE:
		return "dislike"
	default:
		return ""
	}
}

func swipeJSON(s *swipev1.Swipe) map[string]interface{} {
	if s == nil {
		return nil
	}
	return map[string]interface{}{
		"id":         s.GetId(),
		"swiper_id":  s.GetSwiperId(),
		"swipee_id":  s.GetSwipeeId(),
		"direction":  directionString(s.GetDirection()),
		"created_at": s.GetCreatedAt().AsTime(),
	}
}

func matchJSON(m *swipev1.Match) map[string]interface{} {
	if m == nil {
		return nil
	}
	return map[string]interface{}{
		"id":         m.GetId(),
		"user1_id":   m.GetUser1Id(),
		"user2_id":   m.GetUser2Id(),
		"created_at": m.GetCreatedAt().AsTime(),
	}
}

func swipeResponseJSON(r *swipev1.SwipeResponse) map[string]interface{} {
	out := map[string]interface{}{"swipe": swipeJSON(r.GetSwipe())}
	if r.GetMatch() != nil {
		out["match"] = matchJSON(r.GetMatch())
	}
	return out
}
