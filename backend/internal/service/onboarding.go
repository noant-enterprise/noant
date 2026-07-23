package service

import (
	"context"
	"fmt"

	"noant/config"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"
)

type IndustryTemplate struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Icon       string   `json:"icon"`
	Categories []string `json:"categories"`
}

var IndustryTemplates = []IndustryTemplate{
	{
		ID: "retail", Name: "Retail / E-commerce", Icon: "shopping-bag",
		Categories: []string{
			"Shipping & Delivery", "Returns & Refunds", "Order Status",
			"Payment Methods", "Product Availability", "Store Hours",
			"Discounts & Promotions", "Sizing & Fit",
		},
	},
	{
		ID: "food", Name: "Food & Restaurant", Icon: "utensils-crossed",
		Categories: []string{
			"Menu & Pricing", "Opening Hours", "Reservations",
			"Delivery & Takeout", "Dietary Restrictions", "Catering",
			"Gift Cards", "Events & Parties",
		},
	},
	{
		ID: "health", Name: "Healthcare & Clinic", Icon: "stethoscope",
		Categories: []string{
			"Appointment Booking", "Services & Treatments", "Insurance & Billing",
			"Prescriptions", "Office Hours", "Emergency Care",
			"Patient Portal", "Medical Records",
		},
	},
	{
		ID: "education", Name: "Education & Tutoring", Icon: "graduation-cap",
		Categories: []string{
			"Course Offerings", "Tuition & Fees", "Schedule & Calendar",
			"Registration", "Online Learning", "Homework Help",
			"Admissions", "Scholarships",
		},
	},
	{
		ID: "beauty", Name: "Salon & Spa", Icon: "sparkles",
		Categories: []string{
			"Services & Pricing", "Booking Appointments", "Product Sales",
			"Gift Certificates", "Membership Plans", "Cancellation Policy",
			"Hair & Nail Care", "Skin Treatments",
		},
	},
	{
		ID: "fitness", Name: "Fitness & Gym", Icon: "dumbbell",
		Categories: []string{
			"Membership Plans", "Class Schedule", "Personal Training",
			"Facilities & Amenities", "Pricing", "Opening Hours",
			"Trial Passes", "Group Classes",
		},
	},
	{
		ID: "realestate", Name: "Real Estate", Icon: "building",
		Categories: []string{
			"Property Listings", "Schedule Viewing", "Mortgage & Financing",
			"Rental Applications", "Neighborhood Info", "Property Taxes",
			"Home Inspection", "Closing Process",
		},
	},
	{
		ID: "hospitality", Name: "Hotel & Hospitality", Icon: "hotel",
		Categories: []string{
			"Room Booking", "Amenities", "Check-in / Check-out",
			"Cancellation Policy", "Dining Options", "Local Attractions",
			"Group Rates", "Loyalty Program",
		},
	},
	{
		ID: "auto", Name: "Automotive", Icon: "car",
		Categories: []string{
			"Service Appointments", "Parts & Accessories", "Pricing & Quotes",
			"Financing Options", "Test Drives", "Trade-ins",
			"Warranty Info", "Oil Change & Tires",
		},
	},
	{
		ID: "professional", Name: "Professional Services", Icon: "briefcase",
		Categories: []string{
			"Service Offerings", "Consultation Booking", "Pricing & Packages",
			"Client Portal", "Case Updates", "Billing & Invoicing",
			"Appointments", "FAQ",
		},
	},
}

type OnboardingService struct {
	cfg    *config.Config
	repos  *repository.Repositories
	redis  *infrastructure.RedisClient
	logger *infrastructure.Logger
}

func NewOnboardingService(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger) *OnboardingService {
	return &OnboardingService{
		cfg:    cfg,
		repos:  repos,
		redis:  redis,
		logger: logger,
	}
}

func (s *OnboardingService) GetStatus(ctx context.Context, userID string) (map[string]interface{}, error) {
	user, err := s.repos.User.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	status := "pending"
	if user.OnboardingStatus != nil {
		status = *user.OnboardingStatus
	}

	steps := []map[string]interface{}{
		{"id": "profile", "label": "Company Profile", "completed": status != "pending"},
		{"id": "training", "label": "Train Your AI", "completed": status == "training" || status == "whatsapp" || status == "complete"},
		{"id": "whatsapp", "label": "Connect WhatsApp", "completed": status == "whatsapp" || status == "complete"},
		{"id": "complete", "label": "Go Live", "completed": status == "complete"},
	}

	return map[string]interface{}{
		"status": status,
		"steps":  steps,
		"user": map[string]interface{}{
			"company_name":   user.CompanyName,
			"industry":       user.Industry,
			"first_name":     user.FirstName,
			"last_name":      user.LastName,
			"phone":          user.Phone,
			"avatar":         user.Avatar,
		},
	}, nil
}

func (s *OnboardingService) CompleteStep(ctx context.Context, userID, step string, industry *string) error {
	switch step {
	case "profile":
		if err := s.repos.User.UpdateOnboardingStatus(ctx, userID, "training", industry); err != nil {
			return fmt.Errorf("failed to update step: %w", err)
		}
	case "training":
		if err := s.repos.User.UpdateOnboardingStatus(ctx, userID, "whatsapp", nil); err != nil {
			return fmt.Errorf("failed to update step: %w", err)
		}
	case "whatsapp":
		if err := s.repos.User.UpdateOnboardingStatus(ctx, userID, "complete", nil); err != nil {
			return fmt.Errorf("failed to update step: %w", err)
		}
	default:
		return fmt.Errorf("unknown step: %s", step)
	}
	return nil
}

func (s *OnboardingService) AutoCreateCategories(ctx context.Context, userID, industryID string) ([]*domain.Category, error) {
	var template *IndustryTemplate
	for _, t := range IndustryTemplates {
		if t.ID == industryID {
			template = &t
			break
		}
	}
	if template == nil {
		return nil, fmt.Errorf("unknown industry: %s", industryID)
	}

	var cats []*domain.Category
	for _, name := range template.Categories {
		cat := &domain.Category{
			UserID: userID,
			OrgID:  userID,
			Name:   name,
			Description: fmt.Sprintf("Questions about %s", name),
		}
		if err := s.repos.Category.Create(ctx, cat); err != nil {
			return cats, fmt.Errorf("failed to create category %q: %w", name, err)
		}
		cats = append(cats, cat)
	}
	return cats, nil
}

func (s *OnboardingService) GetIndustryTemplates() []IndustryTemplate {
	return IndustryTemplates
}
