package user

import (
	"io"
	"net/http"
	"strconv"

	userv1 "github.com/S1FFFkA/user-mgz/pkg/api/user/v1"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	marshaler = protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: false,
	}
	unmarshaler = protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
)

type Handler struct {
	client *Client
}

func NewHandler(client *Client) *Handler {
	return &Handler{client: client}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	g := e.Group("/v1/users")

	g.POST("", h.createUser)
	g.GET("", h.listUsers)
	g.GET("/:user_id", h.getUser)
	g.PATCH("/:user_id", h.updateUser)
	g.DELETE("/:user_id", h.deleteUser)
	g.POST("/:user_id/photos/upload-url", h.getPhotoUploadURL)
	g.POST("/:user_id/photos/confirm", h.confirmPhotoUpload)
	g.DELETE("/:user_id/photos/:photo_id", h.deletePhoto)
	g.GET("/:user_id/photos/download-url", h.getPhotoDownloadURL)
}

func grpcErrToHTTP(c echo.Context, err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return c.JSON(http.StatusInternalServerError, errResp("internal error"))
	}
	switch st.Code() {
	case codes.NotFound:
		return c.JSON(http.StatusNotFound, errResp("not found"))
	case codes.AlreadyExists:
		return c.JSON(http.StatusConflict, errResp("already exists"))
	case codes.InvalidArgument:
		return c.JSON(http.StatusBadRequest, errResp(st.Message()))
	case codes.Unavailable:
		return c.JSON(http.StatusServiceUnavailable, errResp("service unavailable"))
	default:
		return c.JSON(http.StatusInternalServerError, errResp("internal error"))
	}
}

func errResp(msg string) map[string]string {
	return map[string]string{"error": msg}
}

func writeProto(c echo.Context, msg proto.Message) error {
	data, err := marshaler.Marshal(msg)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errResp("marshal error"))
	}
	return c.JSONBlob(http.StatusOK, data)
}

func readBody(c echo.Context) ([]byte, error) {
	return io.ReadAll(c.Request().Body)
}

func (h *Handler) createUser(c echo.Context) error {
	body, err := readBody(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errResp("cannot read body"))
	}
	req := &userv1.CreateUserRequest{}
	if err := unmarshaler.Unmarshal(body, req); err != nil {
		return c.JSON(http.StatusBadRequest, errResp("invalid json: "+err.Error()))
	}
	// Forward the auth-service UUID so user-mgz uses it as the profile ID,
	// ensuring both services share the same identity across the platform.
	ctx := c.Request().Context()
	if authID := c.Request().Header.Get("X-User-Id"); authID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-auth-user-id", authID)
	}
	resp, err := h.client.api.CreateUser(ctx, req)
	if err != nil {
		return grpcErrToHTTP(c, err)
	}
	return writeProto(c, resp)
}

func (h *Handler) listUsers(c echo.Context) error {
	req := &userv1.ListUsersRequest{
		Limit:  20,
		Offset: 0,
	}
	if v := c.QueryParam("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			req.Limit = int32(n)
		}
	}
	if v := c.QueryParam("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			req.Offset = int32(n)
		}
	}
	if v := c.QueryParam("city_id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			req.CityId = n
		}
	}
	resp, err := h.client.api.ListUsers(c.Request().Context(), req)
	if err != nil {
		return grpcErrToHTTP(c, err)
	}
	return writeProto(c, resp)
}

func (h *Handler) getUser(c echo.Context) error {
	req := &userv1.GetUserRequest{
		UserId: c.Param("user_id"),
	}
	resp, err := h.client.api.GetUser(c.Request().Context(), req)
	if err != nil {
		return grpcErrToHTTP(c, err)
	}
	return writeProto(c, resp)
}

func (h *Handler) updateUser(c echo.Context) error {
	body, err := readBody(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errResp("cannot read body"))
	}
	req := &userv1.UpdateUserRequest{}
	if err := unmarshaler.Unmarshal(body, req); err != nil {
		return c.JSON(http.StatusBadRequest, errResp("invalid json: "+err.Error()))
	}
	req.UserId = c.Param("user_id")
	resp, err := h.client.api.UpdateUser(c.Request().Context(), req)
	if err != nil {
		return grpcErrToHTTP(c, err)
	}
	return writeProto(c, resp)
}

func (h *Handler) deleteUser(c echo.Context) error {
	req := &userv1.DeleteUserRequest{
		UserId: c.Param("user_id"),
	}
	resp, err := h.client.api.DeleteUser(c.Request().Context(), req)
	if err != nil {
		return grpcErrToHTTP(c, err)
	}
	return writeProto(c, resp)
}

func (h *Handler) getPhotoUploadURL(c echo.Context) error {
	body, err := readBody(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errResp("cannot read body"))
	}
	req := &userv1.GetUserPhotoUploadUrlRequest{}
	if err := unmarshaler.Unmarshal(body, req); err != nil {
		return c.JSON(http.StatusBadRequest, errResp("invalid json: "+err.Error()))
	}
	req.UserId = c.Param("user_id")
	resp, err := h.client.api.GetUserPhotoUploadUrl(c.Request().Context(), req)
	if err != nil {
		return grpcErrToHTTP(c, err)
	}
	return writeProto(c, resp)
}

func (h *Handler) confirmPhotoUpload(c echo.Context) error {
	body, err := readBody(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errResp("cannot read body"))
	}
	req := &userv1.ConfirmUserPhotoUploadRequest{}
	if err := unmarshaler.Unmarshal(body, req); err != nil {
		return c.JSON(http.StatusBadRequest, errResp("invalid json: "+err.Error()))
	}
	req.UserId = c.Param("user_id")
	resp, err := h.client.api.ConfirmUserPhotoUpload(c.Request().Context(), req)
	if err != nil {
		return grpcErrToHTTP(c, err)
	}
	return writeProto(c, resp)
}

func (h *Handler) deletePhoto(c echo.Context) error {
	photoID, err := strconv.ParseInt(c.Param("photo_id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errResp("invalid photo_id"))
	}
	req := &userv1.DeleteUserPhotoRequest{
		UserId:  c.Param("user_id"),
		PhotoId: photoID,
	}
	resp, err := h.client.api.DeleteUserPhoto(c.Request().Context(), req)
	if err != nil {
		return grpcErrToHTTP(c, err)
	}
	return writeProto(c, resp)
}

func (h *Handler) getPhotoDownloadURL(c echo.Context) error {
	req := &userv1.GetUserPhotoDownloadUrlRequest{
		UserId: c.Param("user_id"),
	}
	if v := c.QueryParam("photo_type"); v != "" {
		if pt, ok := userv1.PhotoType_value[v]; ok {
			req.PhotoType = userv1.PhotoType(pt)
		} else {
			return c.JSON(http.StatusBadRequest, errResp("invalid photo_type"))
		}
	}
	if v := c.QueryParam("photo_id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			req.PhotoId = n
		}
	}
	resp, err := h.client.api.GetUserPhotoDownloadUrl(c.Request().Context(), req)
	if err != nil {
		return grpcErrToHTTP(c, err)
	}
	return writeProto(c, resp)
}
