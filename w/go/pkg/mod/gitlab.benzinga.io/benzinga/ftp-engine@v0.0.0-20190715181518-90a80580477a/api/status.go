package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type statusResponse struct {
	Status string `json:"status"`
	Build  string `json:"build"`
}

func (h *H) getStatus(c *gin.Context) {

	// Ok
	c.JSON(http.StatusOK, statusResponse{
		Status: "OK",
		Build:  h.config.AppBuild,
	})
}
