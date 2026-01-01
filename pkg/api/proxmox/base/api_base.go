package base

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

// BaseReq holds the HTTP request.
type BaseReq struct {
	*http.Request
}

// BaseAPI defines the interface for making API requests.
type BaseAPI interface {
	NewRequest(path, method string, body io.Reader) (*BaseReq, error)
}

func (bReq *BaseReq) Execute() (res []byte, err error) {
	method := strings.ToLower(bReq.Method)

	// Replace method arg with "create" or "set" accordingly
	var cmdMethod string
	switch method {
	case "post":
		cmdMethod = "create"
	case "put":
		cmdMethod = "set"
	default:
		cmdMethod = method
	}

	args := []string{cmdMethod, bReq.URL.Path, "--output-format", "json"}

	if method == "post" || method == "put" {
		bodyBytes := []byte{}
		if bReq.Body != nil {
			bodyBytesv, err := io.ReadAll(bReq.Body)
			if err != nil && err != io.EOF {
				return nil, err
			}

			if bodyBytesv != nil {
				bodyBytes = bodyBytesv
			}

		}

		if len(bodyBytes) > 0 {
			var bodyMap map[string]interface{}
			if err = json.Unmarshal(bodyBytes, &bodyMap); err != nil {
				return nil, err
			}

			for k, v := range bodyMap {
				valStr := fmt.Sprintf("%v", v)
				args = append(args, fmt.Sprintf("--%s=%s", k, valStr))

			}

		}
	}

	if method == "get" || method == "delete" {
		if bReq.URL.RawQuery != "" {
			params := strings.Split(bReq.URL.RawQuery, "&")
			args = append(args, params...)
		}
	}

	cmd := exec.Command("pvesh", args...)
	log.Debug().
		Str("cmd", "pvesh "+strings.Join(args, " ")).
		Msg("Executing Proxmox API command")

	res, err = cmd.Output()
	if err != nil {
		if res != nil {
			err = fmt.Errorf("%s - %s", string(res), err.Error())
		}
		return nil, err
	}

	return
}
