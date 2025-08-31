package health

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// HealthStatus はヘルスチェックの状態を表します
type HealthStatus struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime"`
	Version   string    `json:"version"`
	Services  Services  `json:"services"`
}

// Services は各サービスの状態を表します
type Services struct {
	Discord  ServiceStatus `json:"discord"`
	Database ServiceStatus `json:"database"`
	Gemini   ServiceStatus `json:"gemini"`
}

// ServiceStatus は個別サービスの状態を表します
type ServiceStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// HealthChecker はヘルスチェック機能を提供します
type HealthChecker struct {
	startTime time.Time
	version   string
	services  map[string]func() ServiceStatus
	mu        sync.RWMutex
}

// NewHealthChecker は新しいHealthCheckerインスタンスを作成します
func NewHealthChecker(version string) *HealthChecker {
	return &HealthChecker{
		startTime: time.Now(),
		version:   version,
		services:  make(map[string]func() ServiceStatus),
	}
}

// RegisterService はサービスを登録します
func (h *HealthChecker) RegisterService(name string, checker func() ServiceStatus) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.services[name] = checker
}

// GetStatus は現在のヘルスステータスを取得します
func (h *HealthChecker) GetStatus() HealthStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()

	services := Services{}

	// 各サービスの状態をチェック
	if discordChecker, exists := h.services["discord"]; exists {
		services.Discord = discordChecker()
	}

	if dbChecker, exists := h.services["database"]; exists {
		services.Database = dbChecker()
	}

	if geminiChecker, exists := h.services["gemini"]; exists {
		services.Gemini = geminiChecker()
	}

	// 全体の状態を判定
	overallStatus := "healthy"
	if services.Discord.Status == "unhealthy" ||
		services.Database.Status == "unhealthy" ||
		services.Gemini.Status == "unhealthy" {
		overallStatus = "unhealthy"
	}

	return HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Uptime:    time.Since(h.startTime).String(),
		Version:   h.version,
		Services:  services,
	}
}

// StartHealthCheckFile はヘルスチェックファイルを定期的に更新します
func (h *HealthChecker) StartHealthCheckFile(filePath string) {
	// シグナルハンドリング
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status := h.GetStatus()
			h.writeHealthFile(filePath, status)
		case <-sigChan:
			log.Println("ヘルスチェックファイル更新を停止します")
			return
		}
	}
}

// writeHealthFile はヘルスチェックファイルを書き込みます
func (h *HealthChecker) writeHealthFile(filePath string, status HealthStatus) {
	content := fmt.Sprintf(`{
		"status": "%s",
		"timestamp": "%s",
		"uptime": "%s",
		"version": "%s"
	}`, status.Status, status.Timestamp.Format(time.RFC3339), status.Uptime, status.Version)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		log.Printf("ヘルスチェックファイルの書き込みに失敗: %v", err)
	}
}

// CreateSimpleHealthChecker はシンプルなヘルスチェック関数を作成します
func CreateSimpleHealthChecker(checkFunc func() error) func() ServiceStatus {
	return func() ServiceStatus {
		if err := checkFunc(); err != nil {
			return ServiceStatus{
				Status:  "unhealthy",
				Message: "サービスが利用できません",
				Error:   err.Error(),
			}
		}
		return ServiceStatus{
			Status:  "healthy",
			Message: "サービスが正常に動作しています",
		}
	}
}
