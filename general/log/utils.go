// Package log is an wrapper package that exposes underlying logger instance.
package log

func getMessage(msg any) string {
	switch msg := msg.(type) {
	case error:
		return msg.Error()
	case string:
		return msg
	default:
		return ""
	}
}
