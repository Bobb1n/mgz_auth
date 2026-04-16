package auth

import (
	"io"
	"net/http"

	"api_gateway/pkg/grpcerr"
	authv1 "auth_service/pkg/api/auth/v1"

	"github.com/labstack/echo/v4"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	marshaler   = protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: false}
	unmarshaler = protojson.UnmarshalOptions{DiscardUnknown: true}
)

type Handler struct {
	client *Client
}

func NewHandler(client *Client) *Handler {
	return &Handler{client: client}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	g := e.Group("/api/v1/auth")
	g.POST("/register", h.register)
	g.POST("/login", h.login)
	g.POST("/refresh", h.refresh)
	g.POST("/logout", h.logout)
}

func (h *Handler) register(c echo.Context) error {
	body, err := readBody(c)
	if err != nil {
		return errJSON(c, http.StatusBadRequest, "cannot read body")
	}
	req := &authv1.RegisterRequest{}
	if err := unmarshaler.Unmarshal(body, req); err != nil {
		return errJSON(c, http.StatusBadRequest, "invalid json: "+err.Error())
	}
	resp, err := h.client.api.Register(c.Request().Context(), req)
	if err != nil {
		return writeGRPCErr(c, err)
	}
	return writeProto(c, http.StatusCreated, resp)
}

func (h *Handler) login(c echo.Context) error {
	body, err := readBody(c)
	if err != nil {
		return errJSON(c, http.StatusBadRequest, "cannot read body")
	}
	req := &authv1.LoginRequest{}
	if err := unmarshaler.Unmarshal(body, req); err != nil {
		return errJSON(c, http.StatusBadRequest, "invalid json: "+err.Error())
	}
	resp, err := h.client.api.Login(c.Request().Context(), req)
	if err != nil {
		return writeGRPCErr(c, err)
	}
	return writeProto(c, http.StatusOK, resp)
}

func (h *Handler) refresh(c echo.Context) error {
	body, err := readBody(c)
	if err != nil {
		return errJSON(c, http.StatusBadRequest, "cannot read body")
	}
	req := &authv1.RefreshRequest{}
	if err := unmarshaler.Unmarshal(body, req); err != nil {
		return errJSON(c, http.StatusBadRequest, "invalid json: "+err.Error())
	}
	resp, err := h.client.api.Refresh(c.Request().Context(), req)
	if err != nil {
		return writeGRPCErr(c, err)
	}
	return writeProto(c, http.StatusOK, resp)
}

func (h *Handler) logout(c echo.Context) error {
	body, err := readBody(c)
	if err != nil {
		return errJSON(c, http.StatusBadRequest, "cannot read body")
	}
	req := &authv1.LogoutRequest{}
	if err := unmarshaler.Unmarshal(body, req); err != nil {
		return errJSON(c, http.StatusBadRequest, "invalid json: "+err.Error())
	}
	if _, err := h.client.api.Logout(c.Request().Context(), req); err != nil {
		return writeGRPCErr(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func readBody(c echo.Context) ([]byte, error) { return io.ReadAll(c.Request().Body) }

func writeProto(c echo.Context, code int, msg proto.Message) error {
	data, err := marshaler.Marshal(msg)
	if err != nil {
		return errJSON(c, http.StatusInternalServerError, "marshal error")
	}
	return c.JSONBlob(code, data)
}

func errJSON(c echo.Context, code int, msg string) error {
	return c.JSON(code, map[string]string{"error": msg})
}

func writeGRPCErr(c echo.Context, err error) error {
	code, msg := grpcerr.HTTPStatus(err)
	return errJSON(c, code, msg)
}
