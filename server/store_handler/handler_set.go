package store_handler

import (
	"fmt"
	// "encoding/json"
	"github.com/rohankmr414/arima/utils"
	"github.com/rohankmr414/arima/fsm"
	"github.com/hashicorp/raft"
	"github.com/labstack/echo/v4"
	"net/http"
	// "strings"
	"time"
)

type requestSet struct {
	Key   string
	Value string
}

// Store handling save to raft cluster. Store will invoke raft.Apply to make this stored in all cluster
// with acknowledge from n quorum. Store must be done in raft leader, otherwise return error.
func (h handler) Set(eCtx echo.Context) error {
	var form = requestSet{}
	if err := eCtx.Bind(&form); err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error binding: %s", err.Error()),
		})
	}

	// if form.Key == "" {
	// 	return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
	// 		"error": "key is empty",
	// 	})
	// }
	

	
	if h.raft.State() != raft.Leader {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": "not the leader",
		})
	}

	payload := fsm.CommandPayload{
		Operation: "set",
		Key:       []byte(form.Key),
		Value:     []byte(form.Value),
	}

	data, err := utils.EncodeMsgPack(payload)
	if err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error preparing saving data payload: %s", err.Error()),
		})
	}

	applyFuture := h.raft.Apply(data.Bytes(), 500*time.Millisecond)
	if err := applyFuture.Error(); err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error persisting data in raft cluster: %s", err.Error()),
		})
	}

	_, ok := applyFuture.Response().(*fsm.ApplyResponse)
	if !ok {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": "error response is not match apply response",
		})
	}

	return eCtx.JSON(http.StatusOK, map[string]interface{}{
		"message": "success persisting data",
		"data":    form,
	})
}
