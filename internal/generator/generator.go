package generator

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"sync/atomic"

	"github.com/Brix101/psgc-tool/internal/domain"
	"github.com/Brix101/psgc-tool/internal/repository"
	"github.com/gocarina/gocsv"
	"go.uber.org/zap"
)

const (
	Regions    = "regions"
	Provinces  = "provinces"
	Cities     = "cities"
	Barangays  = "barangays"
	Masterlist = "masterlist"
	JsonFolder = "files/json"
	CsvFolder  = "files/csv"
)

type Generator struct {
	Filename string

	regRepo      domain.RegionRepository
	provRepo     domain.ProvinceRepository
	cityMuniRepo domain.CityMuniRepository
	bgyRepo      domain.BarangayRepository
}

func NewGenerator(Filename string, db *sql.DB) *Generator {
	regRepo := repository.NewDBRegion(db)
	provRepo := repository.NewDBProvince(db)
	cityMuniRepo := repository.NewDBCityMuni(db)
	brgyRepo := repository.NewDBBarangay(db)

	return &Generator{
		Filename: Filename,

		regRepo:      regRepo,
		provRepo:     provRepo,
		cityMuniRepo: cityMuniRepo,
		bgyRepo:      brgyRepo,
	}
}

func (g *Generator) GenerateData(ctx context.Context, logger *zap.Logger) error {
	file, err := os.Open(g.Filename)
	if err != nil {
		return err
	}
	defer file.Close()

	psgcData := []*domain.Masterlist{}

	if err := gocsv.Unmarshal(file, &psgcData); err != nil {
		return err
	}
	// Create a channel for errors during record creation
	errCh := make(chan error, len(psgcData))

	// Use a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Counter to keep track of processed items
	var processedCount int32

	for i, data := range psgcData {
		wg.Add(1)
		go func(i int, data *domain.Masterlist) {
			defer wg.Done()
			err := g.createRecord(ctx, data)

			if err != nil {
				logger.Error(
					"Create error",
					zap.Error(err),
					zap.String("Level", data.Level),
					zap.Int("Index", i),
				)
				errCh <- err
			} else {
				atomic.AddInt32(&processedCount, 1)
			}
		}(i, data)
	}

	// Close the error channel when all goroutines are done
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Collect errors from the error channel
	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		// You can decide how to handle errors here, e.g., return the first error encountered
		return errors[0]
	}
	// Log the total number of items processed
	logger.Info("Total items processed", zap.Int32("Count", processedCount))

	os.Exit(0)
	return nil
}

func (g *Generator) createRecord(ctx context.Context, data *domain.Masterlist) error {
	switch data.Level {
	case "Reg":
		return g.regRepo.Create(ctx, data)
	case "Prov":
		return g.provRepo.Create(ctx, data)
	case "City", "Mun":
		return g.cityMuniRepo.Create(ctx, data)
	case "Bgy":
		return g.bgyRepo.Create(ctx, data)
	default:
		return nil
	}
}
