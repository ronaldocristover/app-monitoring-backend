package scheduler

import (
	"context"
	"net/http"
	"time"

	"github.com/ronaldocristover/app-monitoring/internal/model"
	"github.com/ronaldocristover/app-monitoring/internal/repository"
	"go.uber.org/zap"
)

type MonitoringWorker struct {
	monitoringConfigRepo repository.MonitoringConfigRepository
	serviceRepo          repository.ServiceRepository
	monitoringLogRepo    repository.MonitoringLogRepository
	scheduler            *Scheduler
	logger               *zap.SugaredLogger
	checkInterval        time.Duration
	stopChan             chan struct{}
}

func NewMonitoringWorker(
	monitoringConfigRepo repository.MonitoringConfigRepository,
	serviceRepo repository.ServiceRepository,
	monitoringLogRepo repository.MonitoringLogRepository,
	sched *Scheduler,
	logger *zap.SugaredLogger,
	checkInterval time.Duration,
) *MonitoringWorker {
	if checkInterval == 0 {
		checkInterval = 30 * time.Second
	}
	return &MonitoringWorker{
		monitoringConfigRepo: monitoringConfigRepo,
		serviceRepo:          serviceRepo,
		monitoringLogRepo:    monitoringLogRepo,
		scheduler:            sched,
		logger:               logger,
		checkInterval:        checkInterval,
		stopChan:             make(chan struct{}),
	}
}

func (w *MonitoringWorker) Start() {
	go w.run()
	w.logger.Info("Monitoring worker started")
}

func (w *MonitoringWorker) Stop() {
	close(w.stopChan)
	w.logger.Info("Monitoring worker stopped")
}

func (w *MonitoringWorker) run() {
	ticker := time.NewTicker(w.checkInterval)
	defer ticker.Stop()

	w.checkAll()
	for {
		select {
		case <-ticker.C:
			w.checkAll()
		case <-w.stopChan:
			return
		}
	}
}

func (w *MonitoringWorker) checkAll() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var configs []model.MonitoringConfig
	if err := w.monitoringConfigRepo.FindEnabled(ctx, &configs); err != nil {
		w.logger.Errorw("failed to fetch enabled monitoring configs", "error", err)
		return
	}

	for _, cfg := range configs {
		cfg := cfg
		w.scheduler.Enqueue(&pingJob{
			config:            cfg,
			serviceRepo:       w.serviceRepo,
			monitoringLogRepo: w.monitoringLogRepo,
			logger:            w.logger,
		})
	}
}

type pingJob struct {
	config            model.MonitoringConfig
	serviceRepo       repository.ServiceRepository
	monitoringLogRepo repository.MonitoringLogRepository
	logger            *zap.SugaredLogger
}

func (j *pingJob) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	svc, err := j.serviceRepo.GetByID(ctx, j.config.ServiceID)
	if err != nil {
		j.logger.Warnw("service not found for monitoring", "service_id", j.config.ServiceID, "error", err)
		return nil
	}

	if svc.URL == "" {
		return nil
	}

	timeout := time.Duration(j.config.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	logEntry := &model.MonitoringLog{
		ServiceID: j.config.ServiceID,
		CheckedAt: time.Now(),
	}

	client := &http.Client{Timeout: timeout}
	start := time.Now()
	resp, err := client.Get(svc.URL)
	elapsed := time.Since(start)
	logEntry.ResponseTimeMs = int(elapsed.Milliseconds())

	if err != nil {
		logEntry.Status = "down"
		logEntry.ErrorMessage = err.Error()
	} else {
		defer resp.Body.Close()
		logEntry.StatusCode = resp.StatusCode
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			logEntry.Status = "up"
		} else {
			logEntry.Status = "down"
			logEntry.ErrorMessage = http.StatusText(resp.StatusCode)
		}
	}

	if err := j.monitoringLogRepo.Create(ctx, logEntry); err != nil {
		j.logger.Errorw("failed to save monitoring log", "service_id", j.config.ServiceID, "error", err)
		return err
	}

	j.logger.Infow("monitoring ping completed",
		"service_id", j.config.ServiceID,
		"service_name", svc.Name,
		"status", logEntry.Status,
		"response_ms", logEntry.ResponseTimeMs,
	)
	return nil
}
