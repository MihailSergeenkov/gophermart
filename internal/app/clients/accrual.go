package clients

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/MihailSergeenkov/gophermart/internal/app/common"
	"github.com/MihailSergeenkov/gophermart/internal/app/config"
	"go.uber.org/zap"
)

const requestTimeout = 1 // in seconds

var ErrOrderRegistered = errors.New("order is not registered")
var ErrServer = errors.New("some server error")
var ErrUnexpectedStatusCode = errors.New("unexpected status code")

type responseOrderAccrual struct {
	Order   string `json:"order"`
	Status  string `json:"status"`
	Accrual int    `json:"accrual,omitempty"`
}

type AccrualClient struct {
	client   http.Client
	settings *config.Settings
	logger   *zap.Logger
}

func newAccrualClient(settings *config.Settings, logger *zap.Logger) *AccrualClient {
	return &AccrualClient{
		client: http.Client{
			Timeout: time.Second * requestTimeout,
		},
		settings: settings,
		logger:   logger,
	}
}

func (ac *AccrualClient) GetOrderAccrual(number string) (string, int, error) {
	const path = "/api/orders/"
	result, err := url.JoinPath("http://", ac.settings.AccrualSystemAddress, path, number)
	if err != nil {
		ac.logger.Error("failed to construct URL", zap.Error(err))
		return "", 0, fmt.Errorf("failed to construct URL: %w", err)
	}

	request, err := http.NewRequest(http.MethodGet, result, nil)
	if err != nil {
		return "", 0, fmt.Errorf("failed to construct request: %w", err)
	}

	request.Header.Set(common.ContentTypeHeader, common.JSONContentType)
	response, err := ac.client.Do(request)
	if err != nil {
		return "", 0, fmt.Errorf("failed to make request: %w", err)
	}

	return parseResponse(response)
}

func parseResponse(response *http.Response) (string, int, error) {
	switch response.StatusCode {
	case 200:
		return decodeResponse(response)
	case 204:
		return "", 0, ErrOrderRegistered
	case 429:
		return "", 0, generateToManyRequestsError(response)
	case 500:
		return "", 0, ErrServer
	default:
		return "", 0, ErrUnexpectedStatusCode
	}
}

func decodeResponse(response *http.Response) (string, int, error) {
	var res responseOrderAccrual

	dec := json.NewDecoder(response.Body)
	if err := dec.Decode(&res); err != nil {
		return "", 0, fmt.Errorf("failed to decode response: %w", err)
	}
	response.Body.Close()

	return res.Status, res.Accrual, nil
}

func generateToManyRequestsError(response *http.Response) error {
	headerRetryAfter := response.Header.Get("Retry-After")
	result, err := strconv.Atoi(headerRetryAfter)
	if err != nil {
		return newToManyRequestsError(60)
	}

	return newToManyRequestsError(result)
}
