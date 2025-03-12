package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"test/load-test/pkg/errors"
	"test/load-test/pkg/response"
	"test/load-test/presentation"
)

func RunK6Test(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Handler")
	cmd := exec.Command("k6", "run", "script.js")
	output, err := cmd.CombinedOutput()
	// resp := response.StandardResponse{}
	if err != nil {
		log.Printf("Error: %+v", err)
		err = errors.NewInternalServerError(err)
	}
	response.RenderResponse(w, err, nil)
	w.Write(output)
}

func Test(w http.ResponseWriter, r *http.Request) {
	err := func() error {
		var req presentation.TestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			err = errors.NewValidationError(err.Error())
			return err
		}

		if req.Name == "" {
			return errors.NewValidationError("Name is empty, please input the string")
		}

		if req.Name == "test-system-error" {
			return fmt.Errorf("Test")
		}
		return nil
	}()

	if err != nil {
		log.Printf("Error: %+v", err)
	}
	response.RenderResponse(w, err, nil)
}
