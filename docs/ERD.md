# DebtEase Database ERD

```mermaid
erDiagram
    users ||--o{ refresh_tokens : "has"
    users ||--o{ debts : "owns"
    debts ||--o{ payments_history : "tracks"
    debts ||--o| credit_cards : "extends"

    users {
        uuid id PK
        varchar name
        varchar email UK
        varchar password_hash
        decimal monthly_income
        boolean is_active
        timestamptz created_at
        timestamptz updated_at
    }

    refresh_tokens {
        text token PK
        uuid user_id FK
        timestamptz created_at
        timestamptz updated_at
        timestamptz expires_at
        timestamptz revoked_at
    }

    debts {
        uuid id PK
        uuid user_id FK
        varchar name
        debt_type type "credit_card|personal_loan|student_loan|mortgage|medical|other"
        varchar lender
        decimal principal_decimal
        decimal outstanding_balance
        decimal interest_rate
        decimal min_payment
        date due_date
        payment_freq payment_frequency "monthly|biweekly|weekly"
        boolean is_overdue
        decimal late_fees
        text other_charges
        timestamptz debt_taken
        risk_level risk_level "low|medium|high"
        decimal debt_health_score
        smallint priority
        text notes
        decimal accrued_interest
        date last_accrual_date
        int payment_due_day "1-31 for loans"
        decimal emi_amount "calculated EMI"
        int tenure_months "loan period"
        int months_paid "EMIs paid"
        date moratorium_until "student loans"
        boolean in_moratorium
        timestamptz created_at
        timestamptz updated_at
    }

    credit_cards {
        uuid debt_id PK_FK
        int billing_day "1-31"
        int grace_period_days
        decimal min_payment_percent
        date last_statement_date
        date payment_due_date
        decimal billed_balance "statement balance"
        decimal unbilled_balance "current spending"
        boolean in_grace_period
        timestamptz created_at
        timestamptz updated_at
    }

    payments_history {
        uuid id PK
        uuid debt_id FK
        decimal amount
        date payment_date
        payment_type type "minimum|extra|full|late"
        text notes
        decimal principal_applied
        decimal interest_applied
        varchar payment_method
        timestamptz created_at
    }
```

## Table Relationships

### users → debts (One-to-Many)
- **Type**: One user can have multiple debts
- **FK**: `debts.user_id` references `users.id`
- **Delete**: CASCADE (deleting user deletes all their debts)

### users → refresh_tokens (One-to-Many)
- **Type**: One user can have multiple refresh tokens
- **FK**: `refresh_tokens.user_id` references `users.id`
- **Delete**: CASCADE (deleting user deletes all their tokens)

### debts → credit_cards (One-to-One)
- **Type**: One debt can be extended by one credit card record (optional)
- **FK**: `credit_cards.debt_id` references `debts.id`
- **Delete**: CASCADE (deleting debt deletes credit card details)
- **Note**: Only exists when `debts.type = 'credit_card'`

### debts → payments_history (One-to-Many)
- **Type**: One debt can have multiple payment records
- **FK**: `payments_history.debt_id` references `debts.id`
- **Delete**: CASCADE (deleting debt deletes payment history)

## Enums

### debt_type
- `credit_card`
- `personal_loan`
- `student_loan`
- `mortgage`
- `medical`
- `other`

### payment_freq
- `monthly`
- `biweekly`
- `weekly`

### risk_level
- `low`
- `medium`
- `high`

### payment_type
- `minimum`
- `extra`
- `full`
- `late`

## Indexes

### users
- `email` (UNIQUE)

### refresh_tokens
- `idx_refresh_user` on `user_id`

### debts
- `idx_debts_user` on `user_id`
- `idx_debts_due` on `due_date`
- `idx_debts_payment_due_day` on `payment_due_day` (WHERE payment_due_day IS NOT NULL)

### credit_cards
- `idx_credit_cards_payment_due` on `payment_due_date`

## Key Design Patterns

### Debt Type Differentiation

#### Credit Cards (`type = 'credit_card'`)
- Has a related record in `credit_cards` table
- Uses daily interest accrual (after grace period)
- Two-balance system: `billed_balance` + `unbilled_balance`
- Monthly statement generation on `billing_day`
- Grace period: statement date → payment due date

#### Loans (`type = 'personal_loan' | 'student_loan' | 'mortgage'`)
- No credit_cards record
- Uses monthly interest accrual on `payment_due_day`
- EMI-based repayment
- Fields: `emi_amount`, `tenure_months`, `months_paid`
- Student loans can have `moratorium_until`

#### Other Debts (`type = 'medical' | 'other'`)
- Basic debt tracking
- Uses daily interest accrual
- No special processing

### Interest Calculation

#### debts.accrued_interest
- Tracks interest that has accrued but not yet been paid
- Updated by workers based on debt type
- Paid first before principal in payments

#### debts.last_accrual_date
- Last date interest was calculated
- Prevents duplicate accrual
- Updated with each interest calculation

### Credit Card Billing Cycle

1. **Statement Generation** (on `billing_day`):
   - `unbilled_balance` → `billed_balance`
   - `unbilled_balance` = 0
   - `in_grace_period` = TRUE
   - `payment_due_date` = today + `grace_period_days`

2. **During Grace Period**:
   - No interest accrues
   - Payments reduce `billed_balance`

3. **After Grace Period** (if not paid in full):
   - `in_grace_period` = FALSE
   - Daily interest starts accruing
   - Interest added to `unbilled_balance`

### Loan EMI Processing

1. **Monthly on `payment_due_day`**:
   - Calculate interest: `outstanding * (APR / 1200)`
   - Calculate principal: `emi_amount - interest`
   - Update `outstanding_balance` -= principal
   - Increment `months_paid`

2. **During Moratorium** (`in_moratorium = TRUE`):
   - Interest capitalizes (added to principal)
   - No EMI payment
   - Continues until `moratorium_until`

3. **Exiting Moratorium**:
   - Recalculate `emi_amount` with capitalized principal
   - Resume normal EMI payments

## Business Logic Summary

### Workers (Background Jobs)

1. **Daily Accrual Worker** → Credit cards out of grace only
2. **Statement Worker** → Generates statements on billing_day
3. **EMI Worker** → Processes loan EMIs on payment_due_day

### Payment Flow

1. User makes payment → `payments_history` record created
2. Payment split: Interest first, then principal
3. Update `debts.outstanding_balance` and `debts.accrued_interest`
4. For credit cards: Update `credit_cards.billed_balance`
5. For loans: Increment `months_paid`

