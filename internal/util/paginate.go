package util

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Brix101/psgc-tool/internal/domain"
	"github.com/go-playground/validator/v10"
)

const (
	DefaultPage    = 1
	DefaultPerPage = 100
)

type PaginateCtx struct{}

func Paginate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the "page", "perPage", and "keyword" query parameters from the URL
		pageParam := r.URL.Query().Get("page")
		perPageParam := r.URL.Query().Get("per_page")
		keywordParam := r.URL.Query().Get("keyword")

		// Parse the "page", "perPage", and "keyword" query parameters
		page, err := strconv.Atoi(pageParam)
		if err != nil || page <= 0 {
			page = DefaultPage
		}

		perPage, err := strconv.Atoi(perPageParam)
		if err != nil || perPage <= 0 {
			perPage = DefaultPerPage
		}

		params := domain.PaginationParams{
			Page:    page,
			PerPage: perPage,
			Keyword: keywordParam,
		}

		validate := validator.New()
		if err := validate.Struct(params); err != nil {
			validationErr, isValidationErr := err.(validator.ValidationErrors)
			if isValidationErr {
				fieldName := validationErr[0].Namespace()
				fieldName = strings.ToLower(fieldName[strings.LastIndex(fieldName, ".")+1:])
				message := fmt.Sprintf(
					"%s should be less than %s.",
					fieldName,
					validationErr[0].Param(),
				)

				http.Error(w, message, http.StatusBadRequest)
				return
			}

			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Create a context with pagination information and pass it down the chain
		ctx := context.WithValue(r.Context(), PaginateCtx{}, params)

		// Serve the request with the modified context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
