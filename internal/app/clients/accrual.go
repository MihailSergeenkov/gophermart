package clients

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/MihailSergeenkov/gophermart/internal/app/config"
	"go.uber.org/zap"
)

const baseRetryTimeout = 60 // in seconds

var (
	ErrOrderRegistered      = errors.New("order is not registered")
	ErrServer               = errors.New("some server error")
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
)

type responseOrderAccrual struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual,omitempty"`
}

type AccrualClient struct {
	systemAddress string
	logger        *zap.Logger
	client        http.Client
}

func newAccrualClient(settings *config.Settings, logger *zap.Logger) *AccrualClient {
	return &AccrualClient{
		client: http.Client{
			Timeout: settings.AccrualRequestTimeout,
		},
		systemAddress: settings.AccrualSystemAddress,
		logger:        logger,
	}
}

func (ac *AccrualClient) GetOrderAccrual(number string) (string, float32, error) {
	const path = "/api/orders/"
	result, err := url.JoinPath(ac.systemAddress, path, number)
	if err != nil {
		return "", 0, fmt.Errorf("failed to construct URL: %w", err)
	}

	request, err := http.NewRequest(http.MethodGet, result, http.NoBody)
	if err != nil {
		return "", 0, fmt.Errorf("failed to construct request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	response, err := ac.client.Do(request)
	if err != nil {
		return "", 0, fmt.Errorf("failed to make request: %w", err)
	}
	defer closeBody(ac, response)

	return parseResponse(ac, response)
}

func parseResponse(ac *AccrualClient, response *http.Response) (string, float32, error) {
	switch response.StatusCode {
	case http.StatusOK:
		return decodeResponse(response)
	case http.StatusNoContent:
		return "", 0, ErrOrderRegistered
	case http.StatusTooManyRequests:
		return "", 0, generateTooManyRequestsError(ac, response)
	case http.StatusInternalServerError:
		return "", 0, ErrServer
	default:
		return "", 0, ErrUnexpectedStatusCode
	}
}

func decodeResponse(response *http.Response) (string, float32, error) {
	var res responseOrderAccrual

	dec := json.NewDecoder(response.Body)
	if err := dec.Decode(&res); err != nil {
		return "", 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return res.Status, res.Accrual, nil
}

func generateTooManyRequestsError(ac *AccrualClient, response *http.Response) error {
	headerRetryAfter := response.Header.Get("Retry-After")

	result, err := strconv.Atoi(headerRetryAfter)
	if err != nil {
		ac.logger.Error("too many request for accrual parsing error", zap.Error(err))
		return newToManyRequestsError(baseRetryTimeout)
	}

	return newToManyRequestsError(result)
}

func closeBody(ac *AccrualClient, r *http.Response) {
	err := r.Body.Close()

	if err != nil {
		ac.logger.Error("failed to close accrual client response body", zap.Error(err))
	}
}
