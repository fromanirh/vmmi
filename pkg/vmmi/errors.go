package vmmi

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

const (
	ErrorCodeNone = iota
	ErrorCodeUnknown
	ErrorCodeMalformedParameters
	ErrorCodeMissingParameters
	ErrorCodeMigrationFailed
	ErrorCodeMigrationAborted
	ErrorCodeVMDisappeared
)

type ErrorData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}

type ErrorMessage struct {
	Header
	Timestamp int64      `json:"timestamp"`
	Error     *ErrorData `json:"error"`
}

func Strerror(code int) string {
	switch code {
	case ErrorCodeNone:
		return "none"
	case ErrorCodeUnknown:
		return "unexpected error"
	case ErrorCodeMalformedParameters:
		return "malformed parameters"
	case ErrorCodeMissingParameters:
		return "missing parameters"
	case ErrorCodeMigrationFailed:
		return "libvirt migration failed"
	case ErrorCodeMigrationAborted:
		return "migration aborted"
	case ErrorCodeVMDisappeared:
		return "VM disappeared"
	}
	return "unknown"
}

func Report(w io.Writer, code int, details string) {
	msg := ErrorMessage{
		Header: Header{
			Version: Version,
			ContentType: MessageError,
		},
		Timestamp: time.Now().Unix(),
		Error: &ErrorData{
			Code:    code,
			Message: Strerror(code),
			Details: details,
		},
	}
	// skip errors: we have no place to report them!
	enc := json.NewEncoder(w)
	enc.Encode(msg)
}

func (pc *PluginContext) Abort(code int, details string) {
	Report(pc.Out, code, details)
	os.Exit(1)
}