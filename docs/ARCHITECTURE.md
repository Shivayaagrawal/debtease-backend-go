# DebtEase Backend - High Level Architecture

## Webapp Implementation Architecture (High Level)

This section describes the **full-stack webapp**: domain logic, Go backend, and Next.js frontend. The webapp offers **standalone calculators** (baseline + strategy) for debt payoff, with no auth required. It is distinct from the **app API** (debts, payments, dashboard, workers), which is documented in the rest of this file.

### System Overview: Domain, Backend, Frontend

```mermaid
graph TB
    subgraph "Frontend (Next.js)"
        WEB[debtease-webapp]
        CALC[Calculator Pages<br/>Baseline / Strategy]
        API_CLIENT[lib/api.ts]
        WEB --> CALC
        CALC --> API_CLIENT
    end

    subgraph "Backend (Go)"
        ROUTER[ServeMux Router]
        MW[Logging / Metrics]
        WEBAPP_H[Webapp Handlers<br/>Calculator, Events, Leads]
        CALC_SVC[Calculator Service]
        ROUTER --> MW
        MW --> WEBAPP_H
        WEBAPP_H --> CALC_SVC
    end

    subgraph "Domain"
        DOMAIN[repayment/v1]
        ENGINE[EMI Engine<br/>Portfolio Schedule<br/>Avalanche / Snowball]
        DOMAIN --> ENGINE
    end

    subgraph "Persistence"
        PG[PostgreSQL]
        SESSIONS[calculation_sessions]
        EVENTS[calculation_events]
        LEADS[lead_emails]
        PG --> SESSIONS
        PG --> EVENTS
        PG --> LEADS
    end

    API_CLIENT -->|POST /api/v1/calculate/*, /api/events, /api/waitlist, /api/pdf-request| ROUTER
    CALC_SVC --> DOMAIN
    WEBAPP_H --> PG
```

### Domain Layer

The **repayment domain** (`internal/domain/repayment/v1`) is a **pure, finance-focused** engine. It owns EMI math, schedule generation, and strategy application. No HTTP, no DB.

| Concept | Description |
|--------|-------------|
| **Loan** | `Principal`, `AnnualRate`, `TenureMonths`, `MoratoriumMonths` |
| **EMIRecord** | Per-month: `OpeningBalance`, `EMI`, `Interest`, `Principal`, `ClosingBalance` |
| **PortfolioSummary** | `Months`, `TotalPaid`, `TotalInterest`, `TotalPrincipal` |
| **StrategyType** | `BASELINE`, `AVALANCHE`, `SNOWBALL` |

**Key functions:**

- `CalculateEMI(principal, annualRate, tenureMonths)` — reducing-balance formula.
- `GenerateLoanSchedule(loan)` — month-by-month schedule per loan (moratorium → repayment).
- `GeneratePortfolioSchedule(loans)` — combined portfolio schedule.
- `SummarizePortfolio(schedule)` — totals and debt-free month.
- `ApplyStrategy(loans, strategy, monthlyExtra)` — Avalanche or Snowball with extra payments.

Engine version is fixed (`EngineVersion`). See `docs/loan_engine.md` for formulae and behaviour.

### Backend (Go)

**Entrypoint:** `cmd/main.go` — `net/http` ServeMux, middleware (logging, metrics), then route handlers.

**Webapp-specific surface (no auth):**

| Route | Handler | Purpose |
|-------|---------|---------|
| `POST /api/v1/calculate/baseline` | `HandleBaselineCalculate` | Portfolio baseline (EMI-only) schedule + summary |
| `POST /api/v1/calculate/strategy` | `HandleStrategyCalculate` | Avalanche/Snowball + extra payment vs baseline |
| `POST /api/events` | `HandleCreateEvent` | Analytics events (e.g. `input_modified_after_result`) |
| `POST /api/pdf-request` | `HandlePdfRequest` | Request PDF plan; stores lead |
| `POST /api/waitlist` | `HandleWaitlist` | Waitlist signup; stores lead |

**Flow:** Handler → **Calculator Service** (`internal/webapp_specifics/service/calculator.go`) → **Domain** (`repayment/v1`). Service maps API DTOs ↔ domain types, then returns JSON.

**Persistence (webapp):** SQLC + PostgreSQL. Tables:

- **calculation_sessions** — `session_id`, `engine_version` (session lifecycle).
- **calculation_events** — `session_id`, `event_type`, `metadata` (JSONB) for analytics.
- **calculation_session_metrics** — aggregated signals (runs, strategy switches, input changes, PDF/waitlist).
- **lead_emails** — `email`, `source` (`waitlist` \| `pdf`), optional `session_id`.

App API (debts, payments, dashboard, auth, workers) is implemented but currently **disabled** in `main.go` (commented out). When enabled, those use separate handlers, middleware (e.g. auth), and DB tables (e.g. `debts`, `payments_history`).

### Frontend (Next.js)

**Stack:** Next.js 16, React 19, Tailwind CSS 4. App router, no auth.

**Structure:**

```
debtease-webapp/
├── src/
│   ├── app/
│   │   ├── (marketing)/          # Marketing + calculators
│   │   │   ├── page.tsx          # Landing
│   │   │   ├── blog/
│   │   │   └── calculator/       # Calculator hub
│   │   │       ├── page.tsx      # Baseline vs Strategy links
│   │   │       ├── baseline/     # Baseline calculator
│   │   │       └── strategy/     # Strategy calculator
│   │   └── layout.tsx
│   ├── components/ui/            # Button, Card, Container, Footer, Navbar
│   ├── features/calculator/      # LoanListEditor, ScheduleTable, SummaryCards, PdfModal
│   ├── hooks/                    # e.g. useSessionID
│   └── lib/
│       ├── api.ts                # API client (calculate, events, waitlist, pdf)
│       └── blog.ts
```

**API client (`lib/api.ts`):** `fetch` wrapper with `credentials: "include"`. Calls:

- `calculateBaseline`, `calculateStrategy` → `/api/v1/calculate/baseline`, `/api/v1/calculate/strategy`
- `createEvent` → `/api/events`
- `submitWaitlistEmail` → `/api/waitlist`
- `requestPdf` → `/api/pdf-request`

**Next.js rewrites:** `/api/*` and `/api/v1/*` are proxied to `BACKEND_ORIGIN` (env, default `http://localhost:8080`). The webapp and Go server can run separately; frontend talks to backend via this proxy.

**Calculators:** Two **independent** flows:

1. **Baseline** — Add/edit loans → POST baseline → show summary + schedule table. Optional events when user changes inputs after a result.
2. **Strategy** — Same loan inputs + strategy (Avalanche/Snowball) + monthly extra → POST strategy → show baseline vs strategy, delta (interest/months saved), schedule. Optional PDF request and waitlist.

### Request Flow (Calculator)

```mermaid
sequenceDiagram
    participant User
    participant Next
    participant Backend
    participant Domain
    participant DB

    User->>Next: Use calculator (baseline / strategy)
    Next->>Next: Build request (loans, strategy, extra)
    Next->>Backend: POST /api/v1/calculate/baseline | strategy
    Backend->>Backend: Validate body, parse DTOs
    Backend->>Domain: GeneratePortfolioSchedule / ApplyStrategy
    Domain->>Domain: EMI, schedule, summary
    Domain-->>Backend: schedule, summary
    Backend->>Backend: Map to API response
    Backend-->>Next: JSON (summary, schedule)
    Next->>User: Render SummaryCards, ScheduleTable

    opt Events / leads
        Next->>Backend: POST /api/events | /api/waitlist | /api/pdf-request
        Backend->>DB: calculation_events | lead_emails
    end
```

### Webapp vs App API

| Aspect | Webapp | App API |
|--------|--------|---------|
| **Auth** | None | JWT (login, refresh) |
| **Domain** | `repayment/v1` (calculators) | Payments, accrual, statements, EMI workers |
| **Storage** | Sessions, events, leads | Users, debts, credit_cards, payments_history |
| **Clients** | Next.js webapp | Flutter app (future), other API clients |

The **domain** is shared only in the sense that both use rigorous finance logic; the **repayment** engine is used by the webapp calculators. The **app** backend uses its own services (e.g. payment service, workers) and schema.

### Technology Stack (Webapp)

| Layer | Technologies |
|-------|--------------|
| **Domain** | Go (`internal/domain/repayment/v1`), no external deps |
| **Backend** | Go 1.22+, `net/http` ServeMux, SQLC, PostgreSQL, Goose migrations |
| **Frontend** | Next.js 16, React 19, Tailwind CSS 4, TypeScript |
| **API contract** | JSON over HTTP; `lib/api.ts` types align with backend DTOs |

---

## System Overview

```mermaid
graph TB
    subgraph "Client Layer"
        API[API Requests]
    end

    subgraph "HTTP Server"
        Router[Go ServeMux Router]
        Auth[Auth Middleware]
        Logging[Logging Middleware]
    end

    subgraph "Handler Layer"
        DebtHandler[Debt Handlers]
        PaymentHandler[Payment Handlers]
        UserHandler[User Handlers]
    end

    subgraph "Service Layer"
        PaymentService[Payment Service<br/>- Interest Accrual<br/>- Payment Processing]
    end

    subgraph "Background Workers"
        DailyWorker[Daily Accrual Worker<br/>Credit Cards Only]
        StatementWorker[Statement Worker<br/>Generate Bills]
        EMIWorker[EMI Worker<br/>Process Loans]
    end

    subgraph "Database Layer"
        PostgreSQL[(PostgreSQL)]
        SQLC[SQLC Generated Code]
    end

    API --> Router
    Router --> Auth
    Auth --> Logging
    Logging --> DebtHandler
    Logging --> PaymentHandler
    Logging --> UserHandler
    
    DebtHandler --> PaymentService
    PaymentHandler --> PaymentService
    
    PaymentService --> SQLC
    DebtHandler --> SQLC
    PaymentHandler --> SQLC
    UserHandler --> SQLC
    
    DailyWorker --> SQLC
    StatementWorker --> SQLC
    EMIWorker --> SQLC
    
    SQLC --> PostgreSQL

    style DailyWorker fill:#e1f5ff
    style StatementWorker fill:#e1f5ff
    style EMIWorker fill:#e1f5ff
    style PaymentService fill:#ffe1e1
```

## Debt Processing Flow

```mermaid
flowchart TD
    Start([User Creates Debt]) --> CheckType{Debt Type?}
    
    CheckType -->|Credit Card| CC1[Create Debt Record]
    CheckType -->|Loan| L1[Create Debt Record]
    CheckType -->|Other| O1[Create Debt Record]
    
    CC1 --> CC2[Calculate EMI: NO]
    CC2 --> CC3[Create credit_cards Record]
    CC3 --> CC4[Set billing_day & grace_period]
    CC4 --> CC5[Initial: in_grace_period = TRUE]
    
    L1 --> L2[Auto-calculate EMI<br/>using Reducing Balance Formula]
    L2 --> L3{Has Moratorium?}
    L3 -->|Yes| L4[Set moratorium_until<br/>in_moratorium = TRUE]
    L3 -->|No| L5[Set payment_due_day<br/>Ready for EMI]
    L4 --> L5
    
    O1 --> O2[Standard debt tracking]
    
    CC5 --> Done([Debt Created])
    L5 --> Done
    O2 --> Done
    
    style CC1 fill:#ffebcc
    style CC2 fill:#ffebcc
    style CC3 fill:#ffebcc
    style CC4 fill:#ffebcc
    style CC5 fill:#ffebcc
    style L1 fill:#ccf2ff
    style L2 fill:#ccf2ff
    style L3 fill:#ccf2ff
    style L4 fill:#ccf2ff
    style L5 fill:#ccf2ff
```

## Interest Accrual Logic

```mermaid
flowchart TD
    Start([Daily Worker Runs]) --> GetDebts[Get All Active Debts]
    GetDebts --> CheckType{Debt Type?}
    
    CheckType -->|Credit Card| CC1{In Grace Period?}
    CheckType -->|Loan| L1[Skip - No Daily Accrual]
    CheckType -->|Other| O1[Apply Daily Interest]
    
    CC1 -->|Yes| CC2[Skip - No Interest]
    CC1 -->|No| CC3[Calculate Days Since Last Accrual]
    CC3 --> CC4[Daily Rate = APR / 36500]
    CC4 --> CC5[Interest = Balance × Daily Rate × Days]
    CC5 --> CC6[Add to debts.accrued_interest]
    CC6 --> CC7[Add to credit_cards.unbilled_balance]
    
    O1 --> O2[Daily Rate = APR / 36500]
    O2 --> O3[Interest = Balance × Days]
    O3 --> O4[Add to accrued_interest]
    
    L1 --> Done([Continue])
    CC2 --> Done
    CC7 --> Done
    O4 --> Done
    
    style CC1 fill:#ffebcc
    style CC2 fill:#ffebcc
    style CC3 fill:#ffebcc
    style CC4 fill:#ffebcc
    style CC5 fill:#ffebcc
    style CC6 fill:#ffebcc
    style CC7 fill:#ffebcc
    style L1 fill:#ccf2ff
```

## Credit Card Statement Generation

```mermaid
flowchart TD
    Start([Statement Worker Runs Daily]) --> Check{Today = billing_day?}
    
    Check -->|No| Skip[Skip]
    Check -->|Yes| Gen1[Move unbilled → billed]
    
    Gen1 --> Gen2[Reset unbilled = 0]
    Gen2 --> Gen3[Set in_grace_period = TRUE]
    Gen3 --> Gen4[Calculate payment_due_date<br/>= today + grace_period_days]
    Gen4 --> Gen5[Update last_statement_date]
    Gen5 --> Notify[Log Statement Generated]
    
    Notify --> Wait([Wait for Payment])
    Wait --> PayCheck{Payment Made?}
    
    PayCheck -->|Full Payment| Pay1[billed_balance = 0]
    PayCheck -->|Partial Payment| Pay2[billed_balance -= amount]
    
    Pay1 --> Grace1[in_grace_period = TRUE]
    Pay2 --> Grace2{billed_balance > threshold?}
    
    Grace2 -->|Yes| Grace3[in_grace_period = FALSE<br/>Interest starts accruing]
    Grace2 -->|No| Grace4[in_grace_period = TRUE]
    
    Grace1 --> Done([Done])
    Grace3 --> Done
    Grace4 --> Done
    Skip --> Done
    
    style Gen1 fill:#ffebcc
    style Gen2 fill:#ffebcc
    style Gen3 fill:#ffebcc
    style Gen4 fill:#ffebcc
    style Gen5 fill:#ffebcc
```

## Loan EMI Processing

```mermaid
flowchart TD
    Start([EMI Worker Runs Daily]) --> Check{Today = payment_due_day?}
    
    Check -->|No| Skip[Skip]
    Check -->|Yes| Mor{In Moratorium?}
    
    Mor -->|Yes| MorCheck{Past moratorium_until?}
    Mor -->|No| Calc1[Calculate Monthly Interest<br/>Rate = APR / 1200]
    
    MorCheck -->|Yes| Exit1[Exit Moratorium]
    MorCheck -->|No| Cap1[Calculate Interest<br/>= Balance × APR/1200]
    
    Cap1 --> Cap2[Capitalize Interest<br/>Balance += Interest]
    Cap2 --> Cap3[Log: Moratorium Interest Capitalized]
    
    Exit1 --> Exit2[Remaining Months<br/>= tenure - months_paid]
    Exit2 --> Exit3[Recalculate EMI<br/>with Capitalized Balance]
    Exit3 --> Exit4[Set in_moratorium = FALSE]
    Exit4 --> Calc1
    
    Calc1 --> Calc2[Interest = Balance × Monthly Rate]
    Calc2 --> Calc3[Principal = EMI - Interest]
    Calc3 --> Calc4{Principal >= Balance?}
    
    Calc4 -->|Yes| Last1[Final Payment<br/>Principal = Balance]
    Calc4 -->|No| Update1[Balance -= Principal]
    
    Last1 --> Update2[Balance = 0]
    Update1 --> Update2
    Update2 --> Update3[months_paid += 1]
    Update3 --> Update4[Update last_accrual_date]
    Update4 --> Log[Log: EMI Processed]
    
    Log --> Done([Done])
    Cap3 --> Done
    Skip --> Done
    
    style Mor fill:#ccf2ff
    style MorCheck fill:#ccf2ff
    style Cap1 fill:#ffe6e6
    style Cap2 fill:#ffe6e6
    style Exit1 fill:#e6ffe6
    style Exit2 fill:#e6ffe6
    style Exit3 fill:#e6ffe6
    style Calc1 fill:#ccf2ff
    style Calc2 fill:#ccf2ff
    style Calc3 fill:#ccf2ff
```

## Payment Processing Flow

```mermaid
flowchart TD
    Start([User Makes Payment]) --> Get1[Get Debt Details]
    Get1 --> Accrue[Accrue Interest to Payment Date]
    
    Accrue --> Type{Debt Type?}
    
    Type -->|Credit Card| CC1{In Grace?}
    Type -->|Loan| L1[No Daily Accrual Needed]
    Type -->|Other| O1[Accrue Daily Interest]
    
    CC1 -->|Yes| CC2[No Accrual]
    CC1 -->|No| CC3[Accrue Daily Interest]
    
    CC2 --> Split
    CC3 --> Split
    L1 --> Split
    O1 --> Split
    
    Split[Calculate Payment Split] --> Split1[Total Due = Balance + Accrued Interest]
    Split1 --> Split2[Interest Paid = min amount accrued_interest]
    Split2 --> Split3[Principal Paid = Amount - Interest Paid]
    
    Split3 --> Update[Update Debt]
    Update --> Update1[Balance -= Principal Paid]
    Update1 --> Update2[Accrued Interest -= Interest Paid]
    Update2 --> Update3[last_accrual_date = Today]
    
    Update3 --> TypeUpdate{Debt Type?}
    
    TypeUpdate -->|Credit Card| CCU1[Update credit_cards.billed_balance]
    TypeUpdate -->|Loan| LU1[Increment months_paid if EMI]
    TypeUpdate -->|Other| OU1[Standard Update]
    
    CCU1 --> Record
    LU1 --> Record
    OU1 --> Record
    
    Record[Create Payment Record] --> Record1[Save to payments_history]
    Record1 --> Record2[Store: amount principal interest]
    Record2 --> Done([Payment Complete])
    
    style Split1 fill:#fff3cd
    style Split2 fill:#fff3cd
    style Split3 fill:#fff3cd
```

## Worker Schedule

```mermaid
gantt
    title Daily Background Workers Schedule
    dateFormat HH:mm
    axisFormat %H:%M
    
    section Workers
    Daily Accrual (Credit Cards)    :active, w1, 00:00, 24h
    Statement Generation (Credit Cards) :active, w2, 00:00, 24h
    EMI Processing (Loans)          :active, w3, 00:00, 24h
    
    section Triggers
    Check billing_day match          :milestone, m1, 00:00, 0h
    Check payment_due_day match      :milestone, m2, 00:00, 0h
    Process credit cards out of grace :milestone, m3, 00:00, 0h
```

## Key Implementation Decisions

### 1. Interest Calculation Methods

| Debt Type | Method | Formula | Frequency |
|-----------|--------|---------|-----------|
| Credit Card | Daily (after grace) | APR / 36500 | Every day |
| Loan | Monthly | APR / 1200 | On due day |
| Other | Daily | APR / 36500 | Every day |

### 2. EMI Formula

**Reducing Balance Method:**
```
EMI = P × r × (1+r)^n / ((1+r)^n - 1)

Where:
P = Principal amount
r = Monthly interest rate (APR / 1200)
n = Number of months
```

### 3. Payment Priority

Payments are always split in this order:
1. **Interest First** - Pay off accrued interest
2. **Principal Second** - Remaining amount reduces principal

### 4. Credit Card Grace Period

```
Statement Date (billing_day)
         ↓
    [Grace Period]  ← No interest accrues
         ↓
Payment Due Date (billing_day + grace_period_days)
         ↓
   [Interest Accrual Starts]  ← If not paid in full
```

### 5. Moratorium Logic (Student Loans)

```
Loan Start
     ↓
[Moratorium Period]
  - Interest accrues monthly
  - Added to principal (capitalized)
  - No EMI payment required
     ↓
Moratorium End
     ↓
EMI Recalculated with:
  - New Principal = Original + Capitalized Interest
  - Remaining Tenure
     ↓
[Normal EMI Payments]
```

## Component Responsibilities

### Handlers Layer
- ✅ Validate API requests
- ✅ Create debts (credit cards + loans)
- ✅ Auto-calculate EMI for loans
- ✅ Record payments
- ✅ Return formatted responses

### Service Layer
- ✅ Interest accrual logic
- ✅ Route by debt type (credit card vs loan)
- ✅ Payment splitting (interest/principal)
- ✅ Transaction management

### Workers Layer
- ✅ **Daily Accrual Worker**: Credit cards out of grace
- ✅ **Statement Worker**: Generate monthly statements
- ✅ **EMI Worker**: Process loan payments & moratorium

### Database Layer
- ✅ SQLC type-safe queries
- ✅ Separate tables for credit cards
- ✅ Loan fields in debts table
- ✅ Payment history tracking

## Technology Stack

```mermaid
graph LR
    subgraph "Backend"
        Go[Go 1.22+]
        ServeMux[net/http ServeMux]
        SQLC[SQLC v1.30]
    end
    
    subgraph "Database"
        PG[PostgreSQL]
        Goose[Goose Migrations]
    end
    
    subgraph "Libraries"
        Decimal[shopspring/decimal]
        UUID[google/uuid]
        Zap[uber-go/zap]
    end
    
    Go --> ServeMux
    ServeMux --> SQLC
    SQLC --> PG
    Goose --> PG
    Go --> Decimal
    Go --> UUID
    Go --> Zap
```

## API Endpoints

### Debt Management
- `POST /api/debts` - Create debt (credit card or loan)
- `GET /api/debts` - List user's debts
- `GET /api/debts/{id}` - Get debt details
- `PATCH /api/debts/{id}/balance` - Update balance
- `PATCH /api/debts/{id}/health` - Update health score
- `POST /api/debts/{id}/overdue` - Mark overdue

### Payment Management
- `POST /api/debts/{id}/payments` - Record payment
- `GET /api/debts/{id}/payments` - List payment history

### Authentication
- `POST /api/register` - Create user
- `POST /api/login` - Authenticate user
- `PUT /api/users` - Update user profile

## Performance Optimizations

1. **Indexes**:
   - `debts.user_id` - Fast user debt lookup
   - `debts.due_date` - Quick overdue checks
   - `debts.payment_due_day` - Efficient EMI scheduling
   - `credit_cards.payment_due_date` - Statement processing
   - `refresh_tokens.user_id` - Auth performance

2. **Worker Efficiency**:
   - Target specific records (out of grace, due today)
   - Batch processing where possible
   - Minimal database queries

3. **Type Safety**:
   - SQLC generates Go structs
   - Compile-time SQL validation
   - No ORM overhead

## Future Enhancements

- [ ] Payment reminders (email/SMS)
- [ ] Debt payoff strategies (avalanche/snowball)
- [ ] Auto-payment scheduling
- [ ] Credit score tracking
- [ ] Budget recommendations
- [ ] Debt consolidation suggestions
- [ ] Mobile app integration
- [ ] Analytics dashboard

