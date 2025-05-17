package main

import (
	"github.com/energye/energy/v2/cef"
)

type MyResourceHandler struct {
	cef.ICefResourceHandler
}

func NewMyResourceHandler() *MyResourceHandler {
	h := &MyResourceHandler{}

	// Create instance from energy binding
	h.ICefResourceHandler = cef.ICefResourceHandler{}

	// Register GetResponseHeaders callback
	h.GetResponseHeaders(h.onGetResponseHeaders)

	// You probably also want to implement Read() and ProcessRequest()
	// to actually serve content.

	return h
}

func (h *MyResourceHandler) onGetResponseHeaders(response *cef.ICefResponse) (responseLength int64, redirectUrl string) {
	response.SetStatus(200)
	response.SetStatusText("OK")

	return responseLength, redirectUrl
}
