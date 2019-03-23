package response

import (
	"encoding/json"
	"log"
	"net/http"
)

type HttpResponseWriter interface {
	WriteToStream(w *http.ResponseWriter)
}

type HttpResponse struct {

	HttpResponseWriter		`json:"-"`
	HttpState 	int 		`json:"-"`
	Code 		int
	Message 	string
	Body		interface{}

}

func (r *HttpResponse) WriteToStream(w *http.ResponseWriter) {

	if bs, err := json.Marshal(r); err != nil {

		if _, err := (*w).Write([]byte("Unkown Error.")); err != nil {
			log.Println(err)
		}

	} else {

		if _, err := (*w).Write(bs); err != nil {
			log.Println(err)
		}

	}

}