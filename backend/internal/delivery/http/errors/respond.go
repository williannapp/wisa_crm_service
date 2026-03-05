package errors

import (
	"github.com/gin-gonic/gin"
)

// RespondWithError maps the given error to AppError and sends a standardized
// JSON response. Handlers MUST use this helper for all error responses;
// never use c.JSON with err.Error() to avoid leaking internal details.
func RespondWithError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	appErr := MapToAppError(err)
	c.JSON(appErr.HTTPStatus, appErr)
}
