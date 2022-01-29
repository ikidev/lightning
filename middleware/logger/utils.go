package logger

import (
	"github.com/ikidev/lightning"
)

func methodColor(method string) string {
	switch method {
	case lightning.MethodGet:
		return cCyan
	case lightning.MethodPost:
		return cGreen
	case lightning.MethodPut:
		return cYellow
	case lightning.MethodDelete:
		return cRed
	case lightning.MethodPatch:
		return cWhite
	case lightning.MethodHead:
		return cMagenta
	case lightning.MethodOptions:
		return cBlue
	default:
		return cReset
	}
}

func statusColor(code int) string {
	switch {
	case code >= lightning.StatusOK && code < lightning.StatusMultipleChoices:
		return cGreen
	case code >= lightning.StatusMultipleChoices && code < lightning.StatusBadRequest:
		return cBlue
	case code >= lightning.StatusBadRequest && code < lightning.StatusInternalServerError:
		return cYellow
	default:
		return cRed
	}
}
