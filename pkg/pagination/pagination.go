package pagination

import (
	"math"
	"net/http"
	"strconv"

	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/constants"
)

type Pagination struct {
	Page   int `json:"page"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func FromRequest(r *http.Request) Pagination {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = constants.DefaultPage
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 1 {
		limit = constants.DefaultLimit
	}
	if limit > constants.MaxLimit {
		limit = constants.MaxLimit
	}

	return Pagination{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}

func TotalPages(totalItems int64, limit int) int {
	return int(math.Ceil(float64(totalItems) / float64(limit)))
}
