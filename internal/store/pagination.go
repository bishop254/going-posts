package store

import (
	"net/http"
	"strconv"
	"time"
)

type PaginatedFeedQuery struct {
	Limit          int    `json:"limit" validate:"gte=1,lte=170"`
	Offset         int    `json:"offset" validate:"gte=0"`
	Sort           string `json:"sort" validate:"oneof=asc desc"`
	Search         string `json:"search" validate:"max=50"`
	AllocationType string `json:"allocation_type" validate:"max=50"`
	Since          string `json:"since"`
	Until          string `json:"until"`
}

type PaginatedAdminUserQuery struct {
	Limit     int    `json:"limit" validate:"gte=1,lte=170"`
	Offset    int    `json:"offset" validate:"gte=0"`
	Sort      string `json:"sort" validate:"oneof=asc desc"`
	Search    string `json:"search" validate:"max=50"`
	Blocked   string `json:"blocked" validate:"max=50"`
	Role      string `json:"role" validate:"max=50"`
	RoleCode  string `json:"role_code" validate:"max=50"`
	Activated string `json:"activated" validate:"max=50"`
	Since     string `json:"since"`
	Until     string `json:"until"`
}

func (fq *PaginatedFeedQuery) Parse(r *http.Request) (*PaginatedFeedQuery, error) {

	queryString := r.URL.Query()

	limit := queryString.Get("limit")
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return nil, err
		}

		fq.Limit = l
	}

	offset := queryString.Get("offset")
	if offset != "" {
		f, err := strconv.Atoi(offset)
		if err != nil {
			return nil, err
		}

		fq.Offset = f
	}

	sort := queryString.Get("sort")
	if sort != "" {
		fq.Sort = sort
	} else {
		fq.Sort = "desc"
	}

	search := queryString.Get("search")
	if search != "" {
		fq.Search = search
	}

	allocationType := queryString.Get("allocation_type")
	if allocationType != "" {
		fq.AllocationType = allocationType
	}

	since := queryString.Get("since")
	if since != "" {
		fq.Since = since + " 00:00:00+00"
	} else {
		fq.Since = "1900-01-01 00:00:00+00"
	}

	until := queryString.Get("until")
	if until != "" {
		fq.Until = until + " 00:00:00+00"
	} else {
		fq.Until = "9999-12-31 00:00:00+00"
	}

	return fq, nil
}

func (fq *PaginatedAdminUserQuery) ParseAdminUser(r *http.Request) (*PaginatedAdminUserQuery, error) {

	queryString := r.URL.Query()

	limit := queryString.Get("limit")
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return nil, err
		}

		fq.Limit = l
	}

	offset := queryString.Get("offset")
	if offset != "" {
		f, err := strconv.Atoi(offset)
		if err != nil {
			return nil, err
		}

		fq.Offset = f
	}

	sort := queryString.Get("sort")
	if sort != "" {
		fq.Sort = sort
	} else {
		fq.Sort = "desc"
	}

	search := queryString.Get("search")
	if search != "" {
		fq.Search = search
	}

	blocked := queryString.Get("blocked")
	if blocked != "" {
		fq.Blocked = blocked
	}

	activated := queryString.Get("activated")
	if activated != "" {
		fq.Activated = activated
	}

	role := queryString.Get("role")
	if role != "" {
		fq.Role = role
	}

	role_code := queryString.Get("role_code")
	if role_code != "" {
		fq.RoleCode = role_code
	}

	since := queryString.Get("since")
	if since != "" {
		fq.Since = since + " 00:00:00+00"
	} else {
		fq.Since = "1900-01-01 00:00:00+00"
	}

	until := queryString.Get("until")
	if until != "" {
		fq.Until = until + " 00:00:00+00"
	} else {
		fq.Until = "9999-12-31 00:00:00+00"
	}

	return fq, nil
}

func parseTime(s string) string {
	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return ""
	}
	return t.Format(time.DateOnly)
}
