package store_handler

import (
	// "encoding/json"
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/labstack/echo/v4"
	"github.com/rohankmr414/arima/fsm"
	"github.com/rohankmr414/arima/utils"
	"net/http"
	"strings"
	"time"
)

// Delete handling remove data from raft cluster. Delete will invoke raft.Apply to make this deleted in all cluster
// with acknowledge from n quorum. Delete must be done in raft leader, otherwise return error.
func (h handler) Delete(eCtx echo.Context) error {
	var key = strings.TrimSpace(eCtx.Param("key"))
	if key == "" {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": "key is empty",
		})
	}

	keyByte := []byte(key)

	if h.raft.State() != raft.Leader {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": "not the leader",
		})
	}

	payload := fsm.CommandPayload{
		Operation: "delete",
		Key:       keyByte,
		Value:     nil,
	}

	data, err := utils.EncodeMsgPack(payload)
	if err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error preparing remove data payload: %s", err.Error()),
		})
	}

	applyFuture := h.raft.Apply(data.Bytes(), 500*time.Millisecond)
	if err := applyFuture.Error(); err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error removing data in raft cluster: %s", err.Error()),
		})
	}

	_, ok := applyFuture.Response().(*fsm.ApplyResponse)
	if !ok {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": "error response is not match apply response",
		})
	}

	return eCtx.JSON(http.StatusOK, map[string]interface{}{
		"message": "success removing data",
		"data": map[string]interface{}{
			"key":   key,
			"value": nil,
		},
	})
}
