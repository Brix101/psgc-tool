package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Brix101/psgc-tool/internal/domain"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const (
	CityCtx = "CityCtx"
)

type cityResource apiResource

// Routes creates a REST router for the cities resource
func (rs cityResource) Routes() chi.Router {
	r := chi.NewRouter()
	// r.Use() // some middleware..

	r.With(paginate).Get("/", rs.List) // GET /cities - read a list of cities

	r.Route("/{psgcCode}", func(r chi.Router) {
		r.Use(rs.CitiesCtx) // lets have a cities map, and lets actually load/manipulate
		r.Get("/", rs.Get)  // GET /cities/{psgcCode} - read a single todo by :id
	})

	return r
}

func (rs cityResource) CitiesCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		psgcCode := chi.URLParam(r, "psgcCode") // Get the {psgcCode} from the route

		item, err := rs.Repo.GetRegionById(ctx, psgcCode)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		ctx = context.WithValue(ctx, CityCtx, item)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ShowCities godoc
//
//	@Summary		Show list of cities
//	@Description	get cities
//	@Tags			cities
//	@Accept			json
//	@Produce		json
//	@Param			query	query		PaginationParams	false	"Pagination and filter parameters"
//	@Success		200		{object}	PaginatedResponse
//	@Failure		400		{object}	string	"Bad Request"
//	@Failure		500		{object}	string	"Internal Server Error"
//	@Router			/cities [get]
func (rs cityResource) List(w http.ResponseWriter, r *http.Request) {
	// Get the context from the request
	ctx := r.Context()

	pageParams, ok := ctx.Value(PaginationParamsKey).(domain.PaginationParams)
	if !ok {
		// Handle the case where pagination information is not found in the context
		// You can choose to use default values or return an error response.
		http.Error(w, "Pagination information not found", http.StatusBadRequest)
		return
	}

	data, err := rs.Repo.GetCityList(ctx, pageParams)
	if err != nil {
		rs.logger.Error("failed to fetch cities from database", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Marshal and send the response
	res, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Error marshaling response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

// ShowCities godoc
//
//	@Summary		Show a city
//	@Description	get string by PsgcCode
//	@Tags			cities
//	@Accept			json
//	@Produce		json
//	@Param			psgcCode	path		string true	"City PsgcCode"
//	@Success		200			{object}	domain.Masterlist
//	@Failure		400			{object}	string	"Bad Request"
//	@Failure		400			{object}	string	"Item Not Found"
//	@Failure		500			{object}	string	"Internal Server Error"
//	@Router			/cities/{psgcCode} [get]
func (rs cityResource) Get(w http.ResponseWriter, r *http.Request) {
	// Get the context from the request
	ctx := r.Context()

	item, ok := ctx.Value(CityCtx).(domain.Masterlist)
	if !ok {
		// Handle the case where item is not found in the context
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	// Marshal and send the response
	res, err := json.Marshal(item)
	if err != nil {
		http.Error(w, "Error marshaling response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}
