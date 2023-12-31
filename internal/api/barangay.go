package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Brix101/psgc-tool/internal/domain"
	"github.com/Brix101/psgc-tool/internal/util"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type (
	BrgyCtx     struct{}
	bryResource struct {
		logger  *zap.Logger
		bgyRepo domain.BarangayRepository
	}
)

// Routes creates a REST router for the barangays resource
func (rs bryResource) Routes() chi.Router {
	r := chi.NewRouter()
	// r.Use() // some middleware..

	r.With(util.Paginate).Get("/", rs.List) // GET /barangays - read a list of barangays

	r.Route("/{psgc_code}", func(r chi.Router) {
		r.Use(rs.BarangayCtx) // lets have a barangays map, and lets actually load/manipulate
		r.Get("/", rs.Get)    // GET /barangays/{psgc_code} - read a single todo by :id
	})

	return r
}

func (rs bryResource) BarangayCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		psgcCode := chi.URLParam(r, "psgc_code") // Get the {psgc_code} from the route

		item, err := rs.bgyRepo.GetById(ctx, psgcCode)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		ctx = context.WithValue(ctx, BrgyCtx{}, item)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ShowBarangays godoc
//
//	@Summary		Show list of Barangays
//	@Description	get Barangays
//	@Tags			Barangays
//	@Accept			json
//	@Produce		json
//	@Param			query	query		PaginationParams	false	"Pagination and filter parameters"
//	@Success		200		{object}	PaginatedBarangay
//	@Failure		400		{object}	string	"Bad Request"
//	@Failure		500		{object}	string	"Internal Server Error"
//	@Router			/barangays [get]
func (rs bryResource) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pageParams, ok := ctx.Value(util.PaginateCtx{}).(domain.PaginationParams)
	if !ok {
		http.Error(w, "Pagination information not found", http.StatusBadRequest)
		return
	}

	data, err := rs.bgyRepo.GetAll(ctx, pageParams)
	if err != nil {
		rs.logger.Error("failed to fetch barangays from database", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Error marshaling response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

// ShowBarangay godoc
//
//	@Summary		Show a Barangay
//	@Description	get string by PsgcCode
//	@Tags			Barangays
//	@Accept			json
//	@Produce		json
//	@Param			psgc_code	path		string true	"Barangay psgcCode"
//	@Success		200			{object}	domain.Barangay
//	@Failure		400			{object}	string	"Bad Request"
//	@Failure		404			{object}	string	"Item Not Found"
//	@Failure		500			{object}	string	"Internal Server Error"
//	@Router			/barangays/{psgc_code} [get]
func (rs bryResource) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	item, ok := ctx.Value(BrgyCtx{}).(domain.Barangay)
	if !ok {

		http.Error(w, domain.ErrNotFound.Error(), http.StatusNotFound)
		return
	}

	res, err := json.Marshal(item)
	if err != nil {
		http.Error(w, "Error marshaling response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}
