package service

import (
	"context"
	"fmt"
	"time"

	"noant/config"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"
)

type CreateCampaignRequest struct {
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type CampaignService struct {
	cfg    *config.Config
	repos  *repository.Repositories
	redis  *infrastructure.RedisClient
	logger *infrastructure.Logger
	credit *CreditService
}

func NewCampaignService(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger, credit *CreditService) *CampaignService {
	return &CampaignService{
		cfg:    cfg,
		repos:  repos,
		redis:  redis,
		logger: logger,
		credit: credit,
	}
}

// Create creates a new campaign schedule
func (s *CampaignService) Create(ctx context.Context, userID string, req CreateCampaignRequest) (*domain.CampaignSchedule, error) {
	// Validate dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %v", err)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %v", err)
	}
	if endDate.Before(startDate) {
		return nil, fmt.Errorf("end date must be after start date")
	}
	if startDate.Before(time.Now()) {
		return nil, fmt.Errorf("start date cannot be in the past")
	}

	// Create campaign record
	campaign := &domain.CampaignSchedule{
		UserID:   userID,
		Name:     req.Name,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Status:    "draft",
	}

	if err := s.repos.Campaign.Create(ctx, campaign); err != nil {
		return nil, err
	}

	return campaign, nil
}

// List returns all campaigns for a user
func (s *CampaignService) List(ctx context.Context, userID string) ([]domain.CampaignSchedule, error) {
	return s.repos.Campaign.ListByUser(ctx, userID)
}

// Cancel cancels a campaign by ID
func (s *CampaignService) Cancel(ctx context.Context, id string, userID string) error {
	// First verify the campaign belongs to the user
	campaigns, err := s.repos.Campaign.ListByUser(ctx, userID)
	if err != nil {
		return err
	}
	for _, c := range campaigns {
		if c.ID == id {
			return s.repos.Campaign.UpdateStatus(ctx, id, "cancelled")
		}
	}
	return fmt.Errorf("campaign not found or access denied")
}

// ProcessStarting is a background job that finds campaigns starting today and activates them
func (s *CampaignService) ProcessStarting(ctx context.Context) error {
	// Find campaigns starting today that are in draft status
	campaigns, err := s.repos.Campaign.GetScheduledForToday(ctx)
	if err != nil {
		return err
	}

	// For each campaign, activate it by purchasing credits
	for _, campaign := range campaigns {
		s.logger.Info("Activating campaign", "campaignID", campaign.ID, "userID", campaign.UserID, "name", campaign.Name)

		// Calculate duration in days
		startDate, _ := time.Parse("2006-01-02", campaign.StartDate)
		endDate, _ := time.Parse("2006-01-02", campaign.EndDate)
		_ = int(endDate.Sub(startDate).Hours()/24) + 1 // inclusive

		// For simplicity, we'll purchase a medium pack for demonstration
		// In a real implementation, this would be configurable or based on user selection
		packType := "medium"

		// Purchase the pack via Polar (this would normally be done by the user)
		// For activation, we simulate a successful purchase
		checkoutID := fmt.Sprintf("campaign_%s_%d", campaign.ID, time.Now().Unix())
		
		// Activate the purchase (add credits and set expiry)
		if err := s.credit.ActivatePurchase(ctx, checkoutID, campaign.UserID, packType); err != nil {
			s.logger.Error("Failed to activate campaign purchase", "error", err, "campaignID", campaign.ID)
			continue // Continue with other campaigns
		}

		// Update campaign status to active
		if err := s.repos.Campaign.UpdateStatus(ctx, campaign.ID, "active"); err != nil {
			s.logger.Error("Failed to update campaign status", "error", err, "campaignID", campaign.ID)
		}

		s.logger.Info("Campaign activated successfully", "campaignID", campaign.ID)
	}

	return nil
}

// ProcessEnding is a background job that finds campaigns ending today and completes them
func (s *CampaignService) ProcessEnding(ctx context.Context) error {
	// Find campaigns ending today that are active
	campaigns, err := s.repos.Campaign.GetEndingToday(ctx)
	if err != nil {
		return err
	}

	// For each campaign, mark as completed
	for _, campaign := range campaigns {
		s.logger.Info("Completing campaign", "campaignID", campaign.ID, "userID", campaign.UserID, "name", campaign.Name)

		if err := s.repos.Campaign.UpdateStatus(ctx, campaign.ID, "completed"); err != nil {
			s.logger.Error("Failed to update campaign status", "error", err, "campaignID", campaign.ID)
			continue
		}

		s.logger.Info("Campaign completed successfully", "campaignID", campaign.ID)
	}

	return nil
}