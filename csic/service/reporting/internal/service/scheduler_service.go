package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"csic-platform/service/reporting/internal/domain"
	"csic-platform/service/reporting/internal/repository"
)

// SchedulerService handles automated report scheduling
type SchedulerService struct {
	scheduleRepo   repository.ReportScheduleRepository
	templateRepo   repository.ReportTemplateRepository
	reportService  *ReportGenerationService
	kafkaProducer  KafkaProducer
	cron           *cron.Cron
	running        bool
	mu             sync.Mutex
	activeSchedules map[string]uuid.UUID // cronEntryID -> scheduleID
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService(
	scheduleRepo repository.ReportScheduleRepository,
	templateRepo repository.ReportTemplateRepository,
	reportService *ReportGenerationService,
	kafkaProducer KafkaProducer,
) *SchedulerService {
	return &SchedulerService{
		scheduleRepo:   scheduleRepo,
		templateRepo:   templateRepo,
		reportService:  reportService,
		kafkaProducer:  kafkaProducer,
		cron:           cron.New(cron.WithSeconds()),
		activeSchedules: make(map[string]uuid.UUID),
	}
}

// Start starts the scheduler
func (s *SchedulerService) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}
	s.running = true
	s.mu.Unlock()

	// Load and start active schedules
	schedules, err := s.scheduleRepo.GetActiveSchedules(ctx)
	if err != nil {
		return fmt.Errorf("failed to load active schedules: %w", err)
	}

	for _, schedule := range schedules {
		if err := s.scheduleRepo.UpdateNextRun(ctx, schedule.ID, time.Now()); err != nil {
			log.Printf("Failed to update next run for schedule %s: %v", schedule.ID, err)
		}
		if err := s.addSchedule(ctx, schedule); err != nil {
			log.Printf("Failed to add schedule %s: %v", schedule.ID, err)
		}
	}

	s.cron.Start()
	log.Println("Report scheduler started")

	return nil
}

// Stop stops the scheduler
func (s *SchedulerService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	// Stop all active schedules
	entries := s.cron.Entries()
	for _, entry := range entries {
		s.cron.Remove(entry.ID)
	}

	s.cron.Stop()
	s.activeSchedules = make(map[string]uuid.UUID)
	s.running = false
	log.Println("Report scheduler stopped")
}

// AddSchedule adds a new schedule
func (s *SchedulerService) AddSchedule(ctx context.Context, schedule *domain.ReportSchedule) error {
	// Validate the cron expression
	if _, err := cron.ParseStandard(schedule.CronExpression); err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	// Create the schedule
	if err := s.scheduleRepo.Create(ctx, schedule); err != nil {
		return fmt.Errorf("failed to create schedule: %w", err)
	}

	// Add to cron if active
	if schedule.IsActive {
		return s.addSchedule(ctx, schedule)
	}

	return nil
}

// addSchedule adds a schedule to the cron scheduler
func (s *SchedulerService) addSchedule(ctx context.Context, schedule *domain.ReportSchedule) error {
	// Calculate next run time
	nextRun := s.calculateNextRun(schedule)
	if err := s.scheduleRepo.UpdateNextRun(ctx, schedule.ID, nextRun); err != nil {
		return fmt.Errorf("failed to update next run: %w", err)
	}

	// Create a unique ID for the cron job
	cronID := fmt.Sprintf("report_schedule_%s", schedule.ID.String())

	// Add to cron
	entryID, err := s.cron.AddFunc(schedule.CronExpression, s.createJobFunc(ctx, schedule))
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	s.mu.Lock()
	s.activeSchedules[entryID.String()] = schedule.ID
	s.mu.Unlock()

	log.Printf("Added schedule %s with cron expression %s, next run at %s", schedule.ID, schedule.CronExpression, nextRun)

	return nil
}

// RemoveSchedule removes a schedule
func (s *SchedulerService) RemoveSchedule(ctx context.Context, scheduleID uuid.UUID) error {
	// Remove from cron
	s.mu.Lock()
	for entryID, schedID := range s.activeSchedules {
		if schedID == scheduleID {
			s.cron.Remove(cron.EntryID(entryID))
			delete(s.activeSchedules, entryID)
			break
		}
	}
	s.mu.Unlock()

	// Deactivate in database
	if err := s.scheduleRepo.Deactivate(ctx, scheduleID); err != nil {
		return fmt.Errorf("failed to deactivate schedule: %w", err)
	}

	log.Printf("Removed schedule %s", scheduleID)

	return nil
}

// UpdateSchedule updates an existing schedule
func (s *SchedulerService) UpdateSchedule(ctx context.Context, schedule *domain.ReportSchedule) error {
	// Remove existing cron job
	s.RemoveSchedule(ctx, schedule.ID)

	// Update in database
	if err := s.scheduleRepo.Update(ctx, schedule); err != nil {
		return fmt.Errorf("failed to update schedule: %w", err)
	}

	// Re-add if active
	if schedule.IsActive {
		return s.addSchedule(ctx, schedule)
	}

	return nil
}

// TriggerSchedule immediately triggers a schedule
func (s *SchedulerService) TriggerSchedule(ctx context.Context, scheduleID uuid.UUID) error {
	schedule, err := s.scheduleRepo.GetByID(ctx, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to get schedule: %w", err)
	}

	if schedule == nil {
		return fmt.Errorf("schedule not found: %s", scheduleID)
	}

	// Trigger report generation
	return s.triggerReportGeneration(ctx, schedule)
}

// createJobFunc creates a job function for a schedule
func (s *SchedulerService) createJobFunc(ctx context.Context, schedule *domain.ReportSchedule) func() {
	return func() {
		if err := s.triggerReportGeneration(context.Background(), schedule); err != nil {
			log.Printf("Failed to trigger report generation for schedule %s: %v", schedule.ID, err)
		}
	}
}

// triggerReportGeneration triggers report generation for a schedule
func (s *SchedulerService) triggerReportGeneration(ctx context.Context, schedule *domain.ReportSchedule) error {
	template, err := s.templateRepo.GetByID(ctx, schedule.TemplateID)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	if template == nil {
		return fmt.Errorf("template not found: %s", schedule.TemplateID)
	}

	// Calculate period based on frequency
	periodStart, periodEnd := s.calculatePeriod(schedule.Frequency)

	// Generate the report
	req := &GenerateReportRequest{
		TemplateID:       schedule.TemplateID,
		PeriodStart:      periodStart,
		PeriodEnd:        periodEnd,
		GeneratedBy:      "scheduler",
		GenerationMethod: "scheduled",
		TriggeredBy:      fmt.Sprintf("schedule:%s", schedule.ID),
		OutputFormats:    schedule.OutputFormats,
	}

	report, err := s.reportService.GenerateReport(ctx, req)
	if err != nil {
		// Increment failure count
		s.scheduleRepo.IncrementRunCount(ctx, schedule.ID, false)
		return fmt.Errorf("failed to generate report: %w", err)
	}

	// Publish Kafka event
	if err := s.kafkaProducer.PublishReportScheduled(ctx, schedule); err != nil {
		log.Printf("Failed to publish schedule event: %v", err)
	}

	// Update schedule statistics
	nextRun := s.calculateNextRun(schedule)
	if err := s.scheduleRepo.UpdateLastRun(ctx, schedule.ID, time.Now(), nextRun); err != nil {
		log.Printf("Failed to update last run: %v", err)
	}

	s.scheduleRepo.IncrementRunCount(ctx, schedule.ID, true)

	log.Printf("Successfully triggered report %s for schedule %s", report.ID, schedule.ID)

	return nil
}

// calculateNextRun calculates the next run time for a schedule
func (s *SchedulerService) calculateNextRun(schedule *domain.ReportSchedule) time.Time {
	// Parse the cron expression
	scheduleParser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	parser, err := scheduleParser.Parse(schedule.CronExpression)
	if err != nil {
		// Fall back to current time
		return time.Now()
	}

	now := time.Now()
	nextRun := parser.Next(now)

	// Apply timezone if specified
	if schedule.TimeZone != "" {
		location, err := time.LoadLocation(schedule.TimeZone)
		if err == nil {
			nextRun = nextRun.In(location)
		}
	}

	// Skip weekends if configured
	if schedule.SkipWeekends {
		for nextRun.Weekday() == time.Saturday || nextRun.Weekday() == time.Sunday {
			nextRun = parser.Next(nextRun)
		}
	}

	// Skip holidays if configured
	if schedule.SkipHolidays {
		for s.isHoliday(nextRun, schedule.Holidays) {
			nextRun = parser.Next(nextRun)
		}
	}

	return nextRun
}

// calculatePeriod calculates the period for a report based on frequency
func (s *SchedulerService) calculatePeriod(frequency domain.ReportFrequency) (time.Time, time.Time) {
	now := time.Now()
	var periodEnd, periodStart time.Time

	switch frequency {
	case domain.FrequencyDaily:
		periodEnd = now
		periodStart = now.AddDate(0, 0, -1)
	case domain.FrequencyWeekly:
		periodEnd = now
		periodStart = now.AddDate(0, 0, -7)
	case domain.FrequencyMonthly:
		periodEnd = now
		periodStart = now.AddDate(0, -1, 0)
	case domain.FrequencyQuarterly:
		periodEnd = now
		periodStart = now.AddDate(0, -3, 0)
	case domain.FrequencyAnnually:
		periodEnd = now
		periodStart = now.AddDate(-1, 0, 0)
	default:
		periodEnd = now
		periodStart = now.AddDate(0, 0, -30)
	}

	return periodStart, periodEnd
}

// isHoliday checks if a date is a holiday
func (s *SchedulerService) isHoliday(date time.Time, holidays []string) bool {
	dateStr := date.Format("2006-01-02")
	for _, holiday := range holidays {
		if holiday == dateStr {
			return true
		}
	}
	return false
}

// GetSchedule retrieves a schedule by ID
func (s *SchedulerService) GetSchedule(ctx context.Context, scheduleID uuid.UUID) (*domain.ReportSchedule, error) {
	return s.scheduleRepo.GetByID(ctx, scheduleID)
}

// ListSchedules lists all schedules
func (s *SchedulerService) ListSchedules(ctx context.Context, page, pageSize int) ([]*domain.ReportSchedule, int, error) {
	return s.scheduleRepo.List(ctx, page, pageSize)
}

// GetDueSchedules retrieves schedules that are due
func (s *SchedulerService) GetDueSchedules(ctx context.Context, now time.Time) ([]*domain.ReportSchedule, error) {
	return s.scheduleRepo.GetDueSchedules(ctx, now)
}

// ActivateSchedule activates a schedule
func (s *SchedulerService) ActivateSchedule(ctx context.Context, scheduleID uuid.UUID) error {
	schedule, err := s.scheduleRepo.GetByID(ctx, scheduleID)
	if err != nil {
		return err
	}

	if schedule == nil {
		return fmt.Errorf("schedule not found: %s", scheduleID)
	}

	// Activate in database
	if err := s.scheduleRepo.Activate(ctx, scheduleID); err != nil {
		return err
	}

	// Add to cron
	return s.addSchedule(ctx, schedule)
}

// DeactivateSchedule deactivates a schedule
func (s *SchedulerService) DeactivateSchedule(ctx context.Context, scheduleID uuid.UUID) error {
	// Remove from cron
	s.RemoveSchedule(ctx, scheduleID)

	// Deactivate in database
	return s.scheduleRepo.Deactivate(ctx, scheduleID)
}

// GetSchedulerStats retrieves scheduler statistics
func (s *SchedulerService) GetSchedulerStats(ctx context.Context) (SchedulerStats, error) {
	activeSchedules, err := s.scheduleRepo.GetActiveSchedules(ctx)
	if err != nil {
		return SchedulerStats{}, err
	}

	dueSchedules, err := s.scheduleRepo.GetDueSchedules(ctx, time.Now())
	if err != nil {
		return SchedulerStats{}, err
	}

	nextRunSchedules, err := s.scheduleRepo.GetNextRunSchedules(ctx, time.Now(), 10)
	if err != nil {
		return SchedulerStats{}, err
	}

	return SchedulerStats{
		TotalActiveSchedules: len(activeSchedules),
		DueSchedules:         len(dueSchedules),
		NextScheduledRuns:    nextRunSchedules,
		Running:              s.running,
	}, nil
}

// SchedulerStats represents scheduler statistics
type SchedulerStats struct {
	TotalActiveSchedules int
	DueSchedules         int
	NextScheduledRuns    []*domain.ReportSchedule
	Running              bool
}
