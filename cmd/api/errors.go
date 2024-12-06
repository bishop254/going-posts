package main

import (
	"net/http"
)

func (a *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	a.logger.Errorw("internal server error", "method", r.Method, "path", r.URL.Path, "err", err.Error())
	writeJSONError(w, http.StatusInternalServerError, err.Error())
}
func (a *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	a.logger.Warnf("bad request error", "method", r.Method, "path", r.URL.Path, "body", r.Body, "err", err.Error())
	writeJSONError(w, http.StatusBadRequest, err.Error())
}
func (a *application) notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	a.logger.Errorw("resource not found error", "method", r.Method, "path", r.URL.Path, "err", err.Error())
	writeJSONError(w, http.StatusNotFound, err.Error())
}
