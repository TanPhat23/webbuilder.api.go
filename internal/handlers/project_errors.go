package handlers

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v3"

	"my-go-app/internal/authz"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"
)

// projectAccessFailure translates the project repository's access errors into
// an HTTP status and a message that is safe to show a client. ok is false for
// anything else, which callers must treat as an internal error.
func projectAccessFailure(err error) (status int, message string, ok bool) {
	switch {
	case errors.Is(err, authz.ErrUnauthorized):
		return http.StatusUnauthorized, "Unauthorized", true
	case errors.Is(err, repositories.ErrProjectUnauthorized), errors.Is(err, authz.ErrForbidden):
		return http.StatusForbidden, "You do not have access to this project", true
	case errors.Is(err, repositories.ErrProjectNotFound), errors.Is(err, authz.ErrNotFound):
		return http.StatusNotFound, "Project not found", true
	case errors.Is(err, repositories.ErrSubdomainTaken):
		return http.StatusConflict, "This subdomain is already taken", true
	case errors.Is(err, services.ErrInvalidSubdomain):
		return http.StatusBadRequest, err.Error(), true
	default:
		return 0, "", false
	}
}

// handleProjectError answers a failed project operation: access errors become
// 403/404, a missing record 404, and everything else a 500 that carries only
// internalMsg (the cause is logged, not returned).
func handleProjectError(c fiber.Ctx, err error, internalMsg string) error {
	if status, message, ok := projectAccessFailure(err); ok {
		return utils.SendError(c, status, message, err)
	}
	return utils.HandleRepoError(c, err, "Project not found", internalMsg)
}

// projectAccessFiberError is projectAccessFailure for handlers that return
// fiber errors instead of writing the response.
func projectAccessFiberError(err error, internalMsg string) error {
	if status, message, ok := projectAccessFailure(err); ok {
		return fiber.NewError(status, message)
	}
	return fiber.NewError(http.StatusInternalServerError, internalMsg)
}
