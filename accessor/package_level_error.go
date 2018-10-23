package accessor

import "net/http"

type Error struct {
	Code            int // http коды из "${GOROOT}/src/net/http/status.go"
	UnderlyingError error
}

func (e *Error) Error() (string) {
	return "Error '" + http.StatusText(e.Code) + "': " + e.UnderlyingError.Error()
}
