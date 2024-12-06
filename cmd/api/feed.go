package main

import (
	"net/http"

	"github.com/bishop254/bursary/internal/store"
)

func (a *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {

	feedQuery := &store.PaginatedFeedQuery{
		Limit:  10,
		Offset: 10,
		Sort:   "desc",
	}

	feedQuery, err := feedQuery.Parse(r)
	if err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(feedQuery); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	feed, err := a.store.Posts.GetUserFeed(ctx, int64(27), feedQuery)
	if err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusOK, feed); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}
