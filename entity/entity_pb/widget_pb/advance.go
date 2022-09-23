package widget_pb

func NewWidgetErr(code int32, message, detail string) *WidgetErrorMessage {
	return &WidgetErrorMessage{
		Code:    code,
		Message: message,
		Detail:  detail,
	}
}
