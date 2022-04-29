package raft_handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// StatsRaftHandler get raft status
func (h handler) StatsRaftHandler(eCtx echo.Context) error {
	return eCtx.JSON(http.StatusOK, map[string]interface{}{
		"message": "Here is the raft status",
		"data":    h.raft.Stats(),
	})
}
