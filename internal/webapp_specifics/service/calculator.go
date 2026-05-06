package service

import (
	repayment_v1 "DebtEase/internal/domain/repayment/v1"
	"DebtEase/internal/webapp_specifics/models"
	"context"

	"github.com/shopspring/decimal"
)

const DefaultEngineVersion = repayment_v1.EngineVersion

func CalculateBaseline(ctx context.Context, loans []models.LoanInput, engineVersion string) (*models.BaselineCalculateResponse, error) {
	if engineVersion == "" {
		engineVersion = DefaultEngineVersion
	}

	// Convert API loans to domain loans
	domainLoans := make([]repayment_v1.Loan, len(loans))
	for i, loan := range loans {
		domainLoans[i] = repayment_v1.Loan{
			ID:               loan.ID,
			Principal:        loan.Principal,
			AnnualRate:       loan.AnnualRate,
			TenureMonths:     loan.TenureMonths,
			MoratoriumMonths: loan.MoratoriumMonths,
		}
	}

	// Generate portfolio schedule
	schedule := repayment_v1.GeneratePortfolioSchedule(domainLoans)

	// Summarize portfolio
	summary := repayment_v1.SummarizePortfolio(schedule)

	// Convert domain schedule to API format
	apiSchedule := convertScheduleToAPI(schedule)

	// Convert domain summary to API format
	apiSummary := models.PortfolioSummary{
		Months:         summary.Months,
		TotalPaid:      summary.TotalPaid,
		TotalInterest:  summary.TotalInterest,
		TotalPrincipal: summary.TotalPrincipal,
	}

	return &models.BaselineCalculateResponse{
		EngineVersion: engineVersion,
		Summary:       apiSummary,
		Schedule:      apiSchedule,
	}, nil
}

func CalculateStrategy(ctx context.Context, loans []models.LoanInput, strategy string, monthlyExtra decimal.Decimal, engineVersion string) (*models.StrategyCalculateResponse, error) {
	if engineVersion == "" {
		engineVersion = DefaultEngineVersion
	}

	// Convert API loans to domain loans
	domainLoans := make([]repayment_v1.Loan, len(loans))
	for i, loan := range loans {
		domainLoans[i] = repayment_v1.Loan{
			ID:               loan.ID,
			Principal:        loan.Principal,
			AnnualRate:       loan.AnnualRate,
			TenureMonths:     loan.TenureMonths,
			MoratoriumMonths: loan.MoratoriumMonths,
		}
	}

	// Convert strategy string to domain StrategyType
	var domainStrategy repayment_v1.StrategyType
	switch strategy {
	case "AVALANCHE":
		domainStrategy = repayment_v1.StrategyAvalanche
	case "SNOWBALL":
		domainStrategy = repayment_v1.StrategySnowball
	default:
		domainStrategy = repayment_v1.StrategyAvalanche
	}

	// Calculate baseline
	baselineSchedule := repayment_v1.GeneratePortfolioSchedule(domainLoans)
	baselineSummary := repayment_v1.SummarizePortfolio(baselineSchedule)

	// Calculate strategy
	strategySchedule := repayment_v1.ApplyStrategy(domainLoans, domainStrategy, monthlyExtra)
	strategySummary := repayment_v1.SummarizePortfolio(strategySchedule)

	// Calculate delta
	delta := models.Delta{
		InterestSaved: baselineSummary.TotalInterest.Sub(strategySummary.TotalInterest),
		MonthsSaved:   baselineSummary.Months - strategySummary.Months,
	}

	// Convert strategy schedule to API format
	strategyAPISchedule := convertScheduleToAPI(strategySchedule)

	return &models.StrategyCalculateResponse{
		EngineVersion: engineVersion,
		Strategy:      strategy,
		BaselineSummary: models.PortfolioSummary{
			Months:         baselineSummary.Months,
			TotalPaid:      baselineSummary.TotalPaid,
			TotalInterest:  baselineSummary.TotalInterest,
			TotalPrincipal: baselineSummary.TotalPrincipal,
		},
		StrategySummary: models.PortfolioSummary{
			Months:         strategySummary.Months,
			TotalPaid:      strategySummary.TotalPaid,
			TotalInterest:  strategySummary.TotalInterest,
			TotalPrincipal: strategySummary.TotalPrincipal,
		},
		Delta:    delta,
		Schedule: strategyAPISchedule,
	}, nil
}

func convertScheduleToAPI(schedule map[int][]repayment_v1.EMIRecord) map[int][]models.EMIRecord {
	apiSchedule := make(map[int][]models.EMIRecord)
	for month, records := range schedule {
		apiRecords := make([]models.EMIRecord, len(records))
		for i, record := range records {
			apiRecords[i] = models.EMIRecord{
				LoanID:         record.LoanID,
				Month:          record.Month,
				OpeningBalance: record.OpeningBalance,
				EMI:            record.EMI,
				Interest:       record.Interest,
				Principal:      record.Principal,
				ClosingBalance: record.ClosingBalance,
			}
		}
		apiSchedule[month] = apiRecords
	}
	return apiSchedule
}
