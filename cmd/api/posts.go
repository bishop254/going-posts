package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bishop254/bursary/internal/store"
	"github.com/go-chi/chi/v5"
)

type postKey string

const postCtx postKey = "post"

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

func (a *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	user := getUserFromCtx(r)

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  int64(user.ID),
	}

	ctx := r.Context()

	if err := a.store.Posts.Create(ctx, post); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusCreated, post); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

type CreatePostCommentPayload struct {
	Content string `json:"content" validate:"required,max=1000"`
	PostID  int64  `json:"post_id" validate:"required"`
}

func (a *application) createPostCommentHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostCommentPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	user := getUserFromCtx(r)

	comment := &store.Comment{
		Content: payload.Content,
		PostID:  payload.PostID,
		UserID:  int64(user.ID),
	}

	ctx := r.Context()

	if err := a.store.Comments.Create(ctx, comment); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusCreated, comment); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

func (a *application) getOnePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	comments, err := a.store.Comments.GetPostWithCommentsByID(r.Context(), post.ID)
	if err != nil {
		a.internalServerError(w, r, err)
		return
	}

	post.Comments = comments

	if err := jsonResponse(w, http.StatusOK, post); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

func (a *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idParam := chi.URLParam(r, "postId")
	postId, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		a.internalServerError(w, r, err)
		return
	}

	err = a.store.Posts.Delete(ctx, postId)
	if err != nil {
		switch {
		case errors.Is(err, errors.New("post not found")):
			a.notFoundError(w, r, err)
		default:
			a.internalServerError(w, r, err)
		}
		return
	}

	type response struct {
		message string
		code    string
	}

	resp := response{
		message: "Deleted post successfully",
		code:    "00",
	}

	if err := jsonResponse(w, http.StatusOK, resp); err != nil {
		a.internalServerError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type UpdatePostPayload struct {
	Title   *string  `json:"title" validate:"omitempty,max=100"`
	Content *string  `json:"content" validate:"omitempty,max=1000"`
	Tags    []string `json:"tags" validate:"omitempty,max=10000"`
}

func (a *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	var payload UpdatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		a.badRequestError(w, r, err)
		return
	}

	if payload.Content != nil {
		post.Content = *payload.Content
	}

	if payload.Title != nil {
		post.Title = *payload.Title
	}

	// if payload.Tags != nil {
	// 	post.Tags = payload.Tags
	// }

	if err := a.store.Posts.Update(r.Context(), post); err != nil {
		a.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusOK, post); err != nil {
		a.internalServerError(w, r, err)
		return
	}
}

func getPostFromCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value(postCtx).(*store.Post)
	return post
}
