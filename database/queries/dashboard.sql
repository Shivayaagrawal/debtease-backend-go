-- name: GetDashboardSummary :one
-- Comprehensive dashboard summary in a single query (more efficient)
SELECT 
    -- Total Outstanding
    COALESCE(SUM(outstanding_balance + accrued_interest), 0) as total_outstanding,
    
    -- Repayment Progress
    COALESCE(SUM(principal_decimal), 0) as total_principal,
    COALESCE(SUM(principal_decimal - outstanding_balance), 0) as total_paid,
    CASE 
        WHEN SUM(principal_decimal) > 0 
        THEN CAST(ROUND((SUM(principal_decimal - outstanding_balance) / SUM(principal_decimal) * 10000)::numeric, 0) AS INTEGER)
        ELSE 0
    END as repayment_percentage,
    
    -- Debt Health (handle NULL debt_health_score by treating as 100)
    CASE 
        WHEN SUM(outstanding_balance) > 0
        THEN CAST(ROUND((SUM(COALESCE(debt_health_score, 100) * outstanding_balance) / SUM(outstanding_balance))::numeric, 0) AS INTEGER)
        ELSE 100
    END as overall_health_score,
    
    -- Monthly Minimum Payment
    COALESCE(SUM(min_payment), 0) as total_minimum_payment,
    
    -- Average Interest
    CASE 
        WHEN SUM(outstanding_balance) > 0 
        THEN CAST(ROUND((SUM(interest_rate * outstanding_balance) / SUM(outstanding_balance) * 100000)::numeric, 0) AS INTEGER)
        ELSE 0
    END as weighted_avg_interest_rate,
    
    -- Counts
    COUNT(*) as total_active_debts,
    COUNT(CASE WHEN is_overdue THEN 1 END) as overdue_count,
    COUNT(CASE WHEN risk_level = 'high' THEN 1 END) as high_risk_count,
    
    -- Expected Debt Free
    CASE 
        WHEN SUM(min_payment) > 0 
        THEN CEIL((SUM(outstanding_balance + accrued_interest) / SUM(min_payment))::numeric)
        ELSE NULL
    END as months_to_debt_free
    
FROM debts
WHERE user_id = $1 
    AND outstanding_balance > 0;