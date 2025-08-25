package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/minh6824pro/nxrGO/internal/dto"
	"github.com/minh6824pro/nxrGO/internal/models"
	"github.com/minh6824pro/nxrGO/internal/repositories"
	"github.com/minh6824pro/nxrGO/internal/services"
	"net/http"
	"net/url"
	"time"
)

type merchantService struct {
	repo repositories.MerchantRepository
}

func NewMerchantService(repo repositories.MerchantRepository) services.MerchantService {
	return &merchantService{repo}
}

func (merchantService *merchantService) Create(ctx context.Context, m *dto.CreateMerchantInput) (*models.Merchant, error) {

	lat, long, err := GetGPSLocation(ctx, m.Location)
	if err != nil {
		return nil, err
	}
	return merchantService.repo.Create(ctx, CreateMerchantInputDtoMapper(m, lat, long))
}

func GetGPSLocation(ctx context.Context, locStr string) (string, string, error) {
	type geocodeResult struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}
	baseURL := "https://nominatim.openstreetmap.org/search"
	q := url.QueryEscape(locStr)
	apiURL := fmt.Sprintf("%s?q=%s&format=json&limit=1", baseURL, q)

	// tạo HTTP client có timeout
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", "", err
	}
	// Nominatim yêu cầu User-Agent hợp lệ
	req.Header.Set("User-Agent", "nxrGO/1.0(minhnd6824@gmail.com)")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var results []geocodeResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return "", "", err
	}

	if len(results) == 0 {
		return "", "", fmt.Errorf("không tìm thấy vị trí cho: %s", locStr)
	}

	return results[0].Lat, results[0].Lon, nil
}

func (merchantService *merchantService) GetByID(ctx context.Context, id uint) (*models.Merchant, error) {

	return merchantService.repo.GetByID(ctx, id)
}

//func (merchantService *merchantService) UpdateOrderStatus(ctx context.Context, m *models.Merchant) error {
//	existing, err := merchantService.repo.GetByID(ctx, m.ID)
//
//	if err != nil {
//		if errors.Is(err, gorm.ErrRecordNotFound) {
//			return gorm.ErrRecordNotFound
//		}
//		return err
//	}
//
//	existing.Name = m.Name
//
//	return merchantService.repo.UpdateOrderStatus(ctx, existing)
//}

func (merchantService *merchantService) Delete(ctx context.Context, id uint) error {
	_, err := merchantService.repo.GetByID(ctx, id)

	if err != nil {
		return err
	}

	return merchantService.repo.Delete(ctx, id)
}

func (merchantService *merchantService) List(ctx context.Context) ([]models.Merchant, error) {
	return merchantService.repo.List(ctx)
}

func (merchantService *merchantService) Patch(ctx context.Context, id uint, input *dto.UpdateMerchantInput) (*models.Merchant, error) {
	existing, err := merchantService.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		existing.Name = input.Name
	}
	if err := merchantService.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func CreateMerchantInputDtoMapper(m *dto.CreateMerchantInput, lat, long string) *models.Merchant {
	return &models.Merchant{
		Name:      m.Name,
		Location:  m.Location,
		Latitude:  lat,
		Longitude: long,
	}
}

func UpdateMerchantInputDtoMapper(m *dto.CreateMerchantInput) *models.Merchant {
	return &models.Merchant{
		Name: m.Name,
	}
}
