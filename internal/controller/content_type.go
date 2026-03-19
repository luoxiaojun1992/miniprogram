package controller

func isInteractionContentType(contentType int8) bool {
	return contentType == 1 || contentType == 2
}
