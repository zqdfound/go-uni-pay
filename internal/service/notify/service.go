package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zqdfound/go-uni-pay/internal/domain/entity"
	"github.com/zqdfound/go-uni-pay/internal/domain/repository"
	"github.com/zqdfound/go-uni-pay/pkg/logger"
	"go.uber.org/zap"
)

// Service 通知服务
type Service struct {
	queueRepo   repository.NotifyQueueRepository
	workerCount int
	retryInterval time.Duration
	maxRetry    int
	stopCh      chan struct{}
}

// NewService 创建通知服务
func NewService(queueRepo repository.NotifyQueueRepository, workerCount int, retryInterval time.Duration, maxRetry int) *Service {
	return &Service{
		queueRepo:     queueRepo,
		workerCount:   workerCount,
		retryInterval: retryInterval,
		maxRetry:      maxRetry,
		stopCh:        make(chan struct{}),
	}
}

// AddNotify 添加通知任务
func (s *Service) AddNotify(ctx context.Context, orderID uint64, orderNo, notifyURL string, notifyData map[string]interface{}) error {
	queue := &entity.NotifyQueue{
		OrderID:    orderID,
		OrderNo:    orderNo,
		NotifyURL:  notifyURL,
		NotifyData: notifyData,
		RetryCount: 0,
		MaxRetry:   s.maxRetry,
		Status:     entity.NotifyStatusPending,
	}

	return s.queueRepo.Create(ctx, queue)
}

// Start 启动通知服务
func (s *Service) Start() {
	logger.Info("notify service started", zap.Int("worker_count", s.workerCount))

	for i := 0; i < s.workerCount; i++ {
		go s.worker(i)
	}
}

// Stop 停止通知服务
func (s *Service) Stop() {
	logger.Info("notify service stopping...")
	close(s.stopCh)
}

// worker 工作协程
func (s *Service) worker(id int) {
	ticker := time.NewTicker(s.retryInterval)
	defer ticker.Stop()

	logger.Info("notify worker started", zap.Int("worker_id", id))

	for {
		select {
		case <-s.stopCh:
			logger.Info("notify worker stopped", zap.Int("worker_id", id))
			return
		case <-ticker.C:
			s.processPendingTasks(context.Background())
		}
	}
}

// processPendingTasks 处理待处理的任务
func (s *Service) processPendingTasks(ctx context.Context) {
	// 获取待处理的任务
	tasks, err := s.queueRepo.GetPendingTasks(ctx, 10)
	if err != nil {
		logger.Error("failed to get pending tasks", zap.Error(err))
		return
	}

	for _, task := range tasks {
		s.processTask(ctx, task)
	}
}

// processTask 处理单个任务
func (s *Service) processTask(ctx context.Context, task *entity.NotifyQueue) {
	logger.Info("processing notify task",
		zap.Uint64("task_id", task.ID),
		zap.String("order_no", task.OrderNo),
		zap.Int("retry_count", task.RetryCount))

	// 更新任务状态为处理中
	task.Status = entity.NotifyStatusProcessing
	if err := s.queueRepo.Update(ctx, task); err != nil {
		logger.Error("failed to update task status", zap.Error(err))
		return
	}

	// 发送通知
	err := s.sendNotify(ctx, task.NotifyURL, task.NotifyData)

	if err != nil {
		// 通知失败
		task.RetryCount++
		task.LastError = err.Error()

		if task.RetryCount >= task.MaxRetry {
			// 超过最大重试次数，标记为失败
			task.Status = entity.NotifyStatusFailed
			logger.Error("notify task failed after max retries",
				zap.Uint64("task_id", task.ID),
				zap.String("order_no", task.OrderNo),
				zap.Int("retry_count", task.RetryCount))
		} else {
			// 计算下次重试时间（指数退避）
			nextRetryTime := time.Now().Add(s.calculateRetryDelay(task.RetryCount))
			task.NextRetryTime = &nextRetryTime
			task.Status = entity.NotifyStatusPending

			logger.Warn("notify task failed, will retry",
				zap.Uint64("task_id", task.ID),
				zap.String("order_no", task.OrderNo),
				zap.Int("retry_count", task.RetryCount),
				zap.Time("next_retry_time", nextRetryTime))
		}
	} else {
		// 通知成功
		now := time.Now()
		task.Status = entity.NotifyStatusSuccess
		task.SuccessTime = &now

		logger.Info("notify task succeeded",
			zap.Uint64("task_id", task.ID),
			zap.String("order_no", task.OrderNo))
	}

	// 更新任务
	if err := s.queueRepo.Update(ctx, task); err != nil {
		logger.Error("failed to update task", zap.Error(err))
	}
}

// sendNotify 发送通知
func (s *Service) sendNotify(ctx context.Context, notifyURL string, notifyData map[string]interface{}) error {
	// 将通知数据转换为JSON
	jsonData, err := json.Marshal(notifyData)
	if err != nil {
		return fmt.Errorf("failed to marshal notify data: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, notifyURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// calculateRetryDelay 计算重试延迟（指数退避）
func (s *Service) calculateRetryDelay(retryCount int) time.Duration {
	// 基础延迟时间（秒）
	delays := []int{60, 120, 300, 600, 1800} // 1分钟, 2分钟, 5分钟, 10分钟, 30分钟

	if retryCount >= len(delays) {
		return time.Duration(delays[len(delays)-1]) * time.Second
	}

	return time.Duration(delays[retryCount]) * time.Second
}
