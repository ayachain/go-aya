package response

import (
	"encoding/json"
	"github.com/labstack/echo"
	"net/http"
)

type HttpResponseWriter interface {
	WriteToEchoContext(c *echo.Context) error
}

type HttpResponse struct {

	HttpResponseWriter		`json:"-"`
	HttpState 	int 		`json:"-"`
	Code 		int
	Message 	string
	Body		interface{}

}

func (r *HttpResponse) WriteToEchoContext(c *echo.Context) error {

	if bs, err := json.Marshal(r); err != nil {

		return err

	} else {

		if err := (*c).String(http.StatusOK, string(bs)); err != nil {
			return err
		}

		return nil
	}

}