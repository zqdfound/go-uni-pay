package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zqdfound/go-uni-pay/internal/payment"
	paymentService "github.com/zqdfound/go-uni-pay/internal/service/payment"
)

// MockPaymentService 模拟支付服务
type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) CreatePayment(ctx context.Context, req *paymentService.CreatePaymentRequest) (*paymentService.CreatePaymentResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paymentService.CreatePaymentResponse), args.Error(1)
}

func (m *MockPaymentService) QueryPayment(ctx context.Context, userID uint64, orderNo string) (interface{}, error) {
	args := m.Called(ctx, userID, orderNo)
	return args.Get(0), args.Error(1)
}

func (m *MockPaymentService) HandleNotify(ctx context.Context, provider string, req *payment.NotifyRequest) ([]byte, error) {
	args := m.Called(ctx, provider, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockPaymentService) GetConfigByID(ctx context.Context, configID uint64) (map[string]interface{}, error) {
	args := m.Called(ctx, configID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// TestHandleNotify_Alipay 测试支付宝支付通知
func TestHandleNotify_Alipay(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testData, err := os.ReadFile("testdata/alipay_notify.txt")
	assert.NoError(t, err)

	formData := parseFormData(string(testData))

	mockService := new(MockPaymentService)
	// Mock 配置查询
	mockService.On("GetConfigByID", mock.Anything, uint64(1)).Return(map[string]interface{}{
		"app_id":      "test_app_id",
		"private_key": "test_private_key",
		"public_key":  "test_public_key",
	}, nil)
	mockService.On("HandleNotify", mock.Anything, "alipay", mock.Anything).Return([]byte("success"), nil)

	handler := NewPaymentHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/notify/alipay/1", bytes.NewBufferString(string(testData)))
	c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.Request.Form = formData
	c.Params = gin.Params{
		{Key: "provider", Value: "alipay"},
		{Key: "config_id", Value: "1"},
	}

	handler.HandleNotify(c)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "success", w.Body.String())
	mockService.AssertExpectations(t)
}

// TestHandleNotify_Wechat 测试微信支付通知
func TestHandleNotify_Wechat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testData, err := os.ReadFile("testdata/wechat_notify.xml")
	assert.NoError(t, err)

	mockService := new(MockPaymentService)
	// Mock 配置查询
	mockService.On("GetConfigByID", mock.Anything, uint64(2)).Return(map[string]interface{}{
		"mch_id":     "test_mch_id",
		"api_v3_key": "test_api_v3_key",
	}, nil)
	mockService.On("HandleNotify", mock.Anything, "wechat", mock.Anything).Return([]byte(`{"code": "SUCCESS", "message": "成功"}`), nil)

	handler := NewPaymentHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/notify/wechat/2", bytes.NewBuffer(testData))
	c.Request.Header.Set("Content-Type", "application/xml")
	c.Params = gin.Params{
		{Key: "provider", Value: "wechat"},
		{Key: "config_id", Value: "2"},
	}

	handler.HandleNotify(c)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "SUCCESS")
	mockService.AssertExpectations(t)
}

// TestHandleNotify_Stripe 测试 Stripe 支付通知
func TestHandleNotify_Stripe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testData, err := os.ReadFile("testdata/stripe_notify.json")
	assert.NoError(t, err)

	mockService := new(MockPaymentService)
	// Mock 配置查询
	mockService.On("GetConfigByID", mock.Anything, uint64(3)).Return(map[string]interface{}{
		"api_key":        "test_api_key",
		"webhook_secret": "test_webhook_secret",
	}, nil)
	mockService.On("HandleNotify", mock.Anything, "stripe", mock.Anything).Return([]byte(`{"received": true}`), nil)

	handler := NewPaymentHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/notify/stripe/3", bytes.NewBuffer(testData))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Stripe-Signature", "t=1234567890,v1=test_signature")
	c.Params = gin.Params{
		{Key: "provider", Value: "stripe"},
		{Key: "config_id", Value: "3"},
	}

	handler.HandleNotify(c)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "received")
	mockService.AssertExpectations(t)
}

// TestHandleNotify_PayPal 测试 PayPal 支付通知
func TestHandleNotify_PayPal(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testData, err := os.ReadFile("testdata/paypal_notify.json")
	assert.NoError(t, err)

	mockService := new(MockPaymentService)
	// Mock 配置查询
	mockService.On("GetConfigByID", mock.Anything, uint64(4)).Return(map[string]interface{}{
		"client_id":     "test_client_id",
		"client_secret": "test_client_secret",
		"webhook_id":    "test_webhook_id",
	}, nil)
	mockService.On("HandleNotify", mock.Anything, "paypal", mock.Anything).Return([]byte(`{"status": "success"}`), nil)

	handler := NewPaymentHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/notify/paypal/4", bytes.NewBuffer(testData))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Paypal-Transmission-Id", "test-transmission-id")
	c.Request.Header.Set("Paypal-Transmission-Time", "2024-01-01T00:00:00Z")
	c.Request.Header.Set("Paypal-Transmission-Sig", "test-signature")
	c.Request.Header.Set("Paypal-Cert-Url", "https://api.paypal.com/cert")
	c.Request.Header.Set("Paypal-Auth-Algo", "SHA256withRSA")
	c.Params = gin.Params{
		{Key: "provider", Value: "paypal"},
		{Key: "config_id", Value: "4"},
	}

	handler.HandleNotify(c)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	mockService.AssertExpectations(t)
}

// TestHandleNotify_InvalidProvider 测试无效的支付提供商
func TestHandleNotify_InvalidProvider(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockPaymentService)
	handler := NewPaymentHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/notify//1", nil)
	c.Params = gin.Params{
		{Key: "provider", Value: ""},
		{Key: "config_id", Value: "1"},
	}

	handler.HandleNotify(c)

	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), "provider is required")
}

// TestHandleNotify_InvalidConfigID 测试无效的配置ID
func TestHandleNotify_InvalidConfigID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockPaymentService)
	handler := NewPaymentHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/notify/stripe/abc", nil)
	c.Params = gin.Params{
		{Key: "provider", Value: "stripe"},
		{Key: "config_id", Value: "abc"},
	}

	handler.HandleNotify(c)

	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), "invalid config_id")
}

// TestHandleNotify_ConfigNotFound 测试配置不存在
func TestHandleNotify_ConfigNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockPaymentService)
	mockService.On("GetConfigByID", mock.Anything, uint64(999)).Return(nil, assert.AnError)

	handler := NewPaymentHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/notify/stripe/999", nil)
	c.Params = gin.Params{
		{Key: "provider", Value: "stripe"},
		{Key: "config_id", Value: "999"},
	}

	handler.HandleNotify(c)

	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), "config not found")
	mockService.AssertExpectations(t)
}

// TestHandleNotify_ServiceError 测试服务错误
func TestHandleNotify_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testData := []byte(`{"test": "data"}`)

	mockService := new(MockPaymentService)
	mockService.On("GetConfigByID", mock.Anything, uint64(3)).Return(map[string]interface{}{
		"api_key": "test_api_key",
	}, nil)
	mockService.On("HandleNotify", mock.Anything, "stripe", mock.Anything).Return(nil, assert.AnError)

	handler := NewPaymentHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/notify/stripe/3", bytes.NewBuffer(testData))
	c.Params = gin.Params{
		{Key: "provider", Value: "stripe"},
		{Key: "config_id", Value: "3"},
	}

	handler.HandleNotify(c)

	assert.Equal(t, 500, w.Code)
	assert.Contains(t, w.Body.String(), "error")
	mockService.AssertExpectations(t)
}

// parseFormData 解析表单数据
func parseFormData(data string) url.Values {
	values := url.Values{}
	pairs := strings.Split(data, "&")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			key, _ := url.QueryUnescape(kv[0])
			value, _ := url.QueryUnescape(kv[1])
			values.Add(key, value)
		}
	}
	return values
}

// BenchmarkHandleNotify_Alipay 基准测试
func BenchmarkHandleNotify_Alipay(b *testing.B) {
	gin.SetMode(gin.TestMode)

	testData, _ := os.ReadFile("testdata/alipay_notify.txt")
	mockService := new(MockPaymentService)
	mockService.On("GetConfigByID", mock.Anything, uint64(1)).Return(map[string]interface{}{
		"app_id": "test_app_id",
	}, nil)
	mockService.On("HandleNotify", mock.Anything, "alipay", mock.Anything).Return([]byte("success"), nil)

	handler := NewPaymentHandler(mockService)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/notify/alipay/1", bytes.NewBuffer(testData))
		c.Params = gin.Params{
			{Key: "provider", Value: "alipay"},
			{Key: "config_id", Value: "1"},
		}
		handler.HandleNotify(c)
	}
}

// ExamplePaymentHandler_HandleNotify 示例
func ExamplePaymentHandler_HandleNotify() {
	gin.SetMode(gin.TestMode)

	mockService := new(MockPaymentService)
	mockService.On("GetConfigByID", mock.Anything, uint64(3)).Return(map[string]interface{}{
		"api_key": "test_api_key",
	}, nil)
	mockService.On("HandleNotify", mock.Anything, "stripe", mock.Anything).Return([]byte(`{"received": true}`), nil)

	handler := NewPaymentHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	payloadJSON := map[string]interface{}{
		"type": "checkout.session.completed",
		"data": map[string]interface{}{
			"object": map[string]interface{}{
				"id":     "cs_test_123",
				"amount": 10000,
			},
		},
	}
	payload, _ := json.Marshal(payloadJSON)

	c.Request = httptest.NewRequest("POST", "/notify/stripe/3", bytes.NewBuffer(payload))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{
		{Key: "provider", Value: "stripe"},
		{Key: "config_id", Value: "3"},
	}

	handler.HandleNotify(c)

	response, _ := io.ReadAll(w.Body)
	fmt.Print(string(response))
	// Output: {"received": true}
}
