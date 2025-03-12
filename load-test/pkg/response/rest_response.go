package response

import (
	"encoding/json"
	"net/http"
	"test/load-test/pkg/errors"
)

type StandardResponse struct {
	Data         interface{} `json:"data"`
	StatusCode   int32       `json:"status_code"`
	ErrorMessage string      `json:"error_message"`
}

func RenderResponse(w http.ResponseWriter, err error, data interface{}) {
	resp := StandardResponse{
		StatusCode: http.StatusOK,
		Data:       data,
	}
	if err != nil {
		resp.StatusCode = http.StatusInternalServerError
		resp.ErrorMessage = err.Error()
		if errors.IsValidationError(err) {
			resp.StatusCode = http.StatusBadRequest
		}
	}
	w.WriteHeader(int(resp.StatusCode))
	json.NewEncoder(w).Encode(resp)
}
