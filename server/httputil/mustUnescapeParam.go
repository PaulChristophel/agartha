package httputil

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

func MustUnescapeParam(c *gin.Context, name string) (string, bool) {
	raw := c.Param(name)
	val, err := url.PathUnescape(raw)
	if err != nil {
		NewError(c, http.StatusBadRequest, fmt.Sprintf("Invalid %s path component: %q", name, raw))
		return "", false
	}
	return val, true
}
