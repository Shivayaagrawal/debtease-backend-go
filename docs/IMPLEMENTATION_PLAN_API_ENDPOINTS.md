# Implementation Plan: API Endpoints (Finance Steps, Session, Debt Health, Loan Simulation, Stress)

This document maps the required features to **RESTful API endpoints** and implementation steps for the Debtease backend.

---

## 1. Overview: Feature → Endpoint Mapping

| Feature | Purpose | REST Endpoints |
|--------|---------|----------------|
| **Finance steps / manual input** | Save and load user-entered finance data (loans, income, obligations) per session | `POST` create/update, `GET` retrieve |
| **Session storage** | Persist and retrieve calculation session and its state | `GET` session (by cookie/session_id) |
| **Debt health** | Per-debt score + portfolio-level health/classification | `GET` health summary, `PATCH` score (existing), **classification** endpoint |
| **Test your loan** | Run loan simulation (baseline/strategy) and return results | Reuse/align with existing `POST /api/v1/calculate/*` |
| **Loan simulation input** | REST resource for “simulation input” (loans + params) | `POST` submit input, `GET` last input by session |
| **Classification (debt health)** | Return treatment category (SAFE_TO_HOLD, NEEDS_MONITORING, etc.) + metrics | New `POST /api/v1/classification` (or under stress) |
| **Stress folder** | New package with stress metrics + classification APIs | New `internal/stress` (or `internal/domain/stress`) + REST routes |

---

## 2. Current State (Already in Codebase)

- **Sessions:** `calculation_sessions` table; `GetOrCreateSession`, `SessionExists`; cookie `calc_session`.
- **Calculator:** `POST /api/v1/calculate/baseline`, `POST /api/v1/calculate/strategy` (loan list in body); session from cookie.
- **Events:** `POST /api/events` (event_type + metadata); tied to session.
- **Debt health (auth layer):** `PATCH /api/debts/{id}/health` (handler exists, commented out in `main.go`).
- **Classification domain:** Documented in `docs/domains/v1/loan_classifications_user_risk.md`; **no Go implementation yet**.
- **No “stress” package** and no dedicated “finance steps” or “simulation input” storage table.

---

## 3. REST Endpoint Specification

### 3.1 Finance Steps / Manual Input (POST, GET)

**Intent:** Persist and retrieve the user’s manual finance inputs (e.g. loan list, monthly income, fixed obligations) keyed by session.

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/finance-steps` | Create or update manual finance input for the current session (session from cookie or body). |
| `GET`  | `/api/v1/finance-steps` | Get the latest finance steps / manual input for the current session. |

**Request (POST):**

- **Session:** From cookie `calc_session` (create session if missing, same as calculator).
- **Body (example):**
  - `session_id` (optional, UUID) – if no cookie.
  - `monthly_income` (decimal, optional).
  - `fixed_obligations` (decimal, optional).
  - `loans` (array of loan input objects, same shape as baseline/strategy).
  - `engine_version` (string, optional).

**Response:**

- **POST:** `200 OK` or `201 Created` with saved payload (e.g. `{ "session_id", "saved_at", "finance_steps": { ... } }`).
- **GET:** `200 OK` with `{ "session_id", "finance_steps": { ... } }` or `404` if no saved steps.

**Implementation:**

- Add table (e.g. `finance_steps` or `session_finance_input`): `session_id` (FK to `calculation_sessions`), `monthly_income`, `fixed_obligations`, `loans` (JSONB), `engine_version`, `created_at`, `updated_at`. One row per session (upsert by `session_id`).
- Add sqlc queries: UpsertFinanceSteps, GetFinanceStepsBySessionID.
- Handlers: `internal/webapp_specifics/handlers/finance_steps.go` – POST/GET using session from cookie (and optional body), call service → DB.
- Register in `cmd/main.go`: `POST /api/v1/finance-steps`, `GET /api/v1/finance-steps` with logging middleware.

---

### 3.2 Session Storage (GET)

**Intent:** Let the client read the current session and optionally minimal state (e.g. session_id, engine_version, created_at).

| Method | Path | Description |
|--------|------|-------------|
| `GET`  | `/api/v1/session` | Return current session info (from cookie or query `session_id`). |

**Request:**

- Cookie `calc_session` (preferred) or query `?session_id=<uuid>`.

**Response:**

- `200 OK`: `{ "session_id", "engine_version", "created_at" }`.
- `400` if no session; `404` if session_id not found in DB.

**Implementation:**

- Reuse `GetCalculationSessionByID` (or add GetSessionByID if needed). Handler in `internal/webapp_specifics/handlers/session.go`: read session from cookie/query, validate exists, return JSON.
- Register: `GET /api/v1/session`.

---

### 3.3 Debt Health (GET summary, PATCH score, Classification)

**Intent:** (1) Per-debt health score (PATCH already defined). (2) Portfolio-level health summary. (3) Classification (treatment category + metrics).

| Method | Path | Description |
|--------|------|-------------|
| `PATCH` | `/api/debts/{id}/health` | Update debt health score (existing handler; uncomment in main when auth is on). |
| `GET`   | `/api/v1/debt-health` | Portfolio-level debt health summary (e.g. weighted score, counts). Input: session (cookie) or body with loan list. |
| `POST`  | `/api/v1/classification` | Return treatment classification + context metrics (see § 3.6). |

**GET /api/v1/debt-health:**

- Input: session (cookie) → load finance_steps / last simulation input; or accept body with `loans` + `monthly_income` (stateless).
- Output: e.g. `overall_health_score`, `debt_count`, optional per-debt scores if available.

**Implementation:**

- **GET debt-health:** Handler that uses session → finance_steps (or request body); if loans exist, optionally compute a simple aggregate score (or delegate to stress/classification). No new table if reusing finance_steps + existing dashboard-style logic.
- **Classification:** Implement in stress package and expose via `POST /api/v1/classification` (see § 3.6).

---

### 3.4 Test Your Loan / Loan Simulation (REST)

**Intent:** “Test your loan” = run a simulation. “Loan simulation input” = REST resource for the input (loans + strategy params).

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/calculate/baseline` | Already exists – run baseline simulation; body = loans (+ engine_version). |
| `POST` | `/api/v1/calculate/strategy` | Already exists – run strategy simulation; body = loans, strategy, monthly_extra. |
| `POST` | `/api/v1/simulation/input` | Save simulation input (loans + params) for current session (optional; can merge with finance-steps). |
| `GET`  | `/api/v1/simulation/input` | Get last simulation input for current session. |

**Options:**

- **A)** Treat “loan simulation input” as the same as “finance steps”: use `POST/GET /api/v1/finance-steps` for both (one stored blob per session: loans + income + obligations).
- **B)** Separate resource: add `POST/GET /api/v1/simulation/input` that store only `{ loans, strategy?, monthly_extra?, engine_version }` (e.g. in `session_simulation_input` or same table with a type flag).

**Recommendation:** Start with **A** (single “finance steps” store). If later you need a distinct “last simulation” (e.g. without income), add **B** with a small table or extra JSONB column.

**Implementation:**

- If A: no extra endpoints; document that “loan simulation input” is persisted via `POST /api/v1/finance-steps` and read via `GET /api/v1/finance-steps`. Calculator can POST finance-steps after each run, or client can GET finance-steps and prefill the “test your loan” form.
- If B: add `POST /api/v1/simulation/input`, `GET /api/v1/simulation/input`; DB and handlers similar to finance-steps but scoped to simulation payload only.

---

### 3.5 Classification (Debt Health) – Output

**Intent:** Return treatment classification (SAFE_TO_HOLD, NEEDS_MONITORING, AVOID_ADDING, ACTIVELY_REDUCE) and context metrics (EMI ratio, fragmentation, shock buffer, etc.) as per `docs/domains/v1/loan_classifications_user_risk.md`.

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/classification` | Request body: loans (+ optional monthly_income, fixed_obligations). Response: classification + metrics + flags. |

**Request body (example):**

```json
{
  "loans": [ { "id", "principal", "annual_rate", "tenure_months", "moratorium_months" } ],
  "monthly_income": "150000.00",
  "fixed_obligations": "20000.00",
  "engine_version": "v1"
}
```

**Response (example):**

```json
{
  "classification": "NEEDS_MONITORING",
  "confidence": "medium",
  "metrics": {
    "emi_ratio": 0.38,
    "fragmentation_index": 3,
    "rigidity_score": 0.42,
    "shock_buffer_months": 2.1,
    "unsecured_ratio": 0.55
  },
  "flags": [ "HIGH_CASHFLOW_RIGIDITY" ]
}
```

**Implementation:** Implement inside the new **stress** package and expose via this endpoint (see § 4).

---

### 3.6 Stress Folder (New Package + REST)

**Intent:** New package that owns stress/classification logic and exposes REST endpoints.

**Suggested layout:**

```
internal/
  stress/                    # NEW
    handlers/
      classification.go      # POST /api/v1/classification
      stress_metrics.go      # POST /api/v1/stress/metrics (optional)
    service/
      classification.go      # classification pipeline (normalize → metrics → policy → category)
      metrics.go            # EMI ratio, fragmentation, rigidity, shock buffer, unsecured ratio
    models/
      request.go            # ClassificationRequest, StressRequest
      response.go           # ClassificationResponse, StressMetricsResponse
```

**REST endpoints:**

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/classification` | Full classification + metrics + flags (see § 3.5). |
| `POST` | `/api/v1/stress/metrics`  | Optional: return only stress metrics (no treatment category). |

**Implementation steps:**

1. **Create `internal/stress`** (or `internal/domain/stress`): models for request/response (use `decimal.Decimal`, `time.Time` per workspace rules).
2. **Implement metrics (from classification doc):**
   - EMI ratio = Total_Monthly_EMI / Monthly_Net_Income
   - Fragmentation index = (Number_of_Active_Debts + BNPL_Count * 0.5)
   - Rigidity score = Σ(EMI × Remaining_Months) / Monthly_Income
   - Shock buffer months = (Income - Fixed_Obligations - EMIs) / EMIs
   - Unsecured ratio = Unsecured_EMI / Total_EMI
3. **Implement policy thresholds (v1):** e.g. MAX_SAFE_EMI_RATIO=0.35, MAX_FRAGMENTATION_INDEX=4, MIN_SHOCK_BUFFER_MONTHS=1.5, MAX_UNSECURED_RATIO=0.6.
4. **Classification pipeline:** Normalize debts → compute metrics → apply thresholds → assign treatment (SAFE_TO_HOLD, NEEDS_MONITORING, AVOID_ADDING, ACTIVELY_REDUCE) + flags.
5. **Handlers:** Parse JSON (loans, monthly_income, fixed_obligations), call service, return JSON (classification + metrics + flags).
6. **Register in `cmd/main.go`:**
   - `POST /api/v1/classification` → stress handler
   - `POST /api/v1/stress/metrics` → stress handler (optional)

Use existing repayment domain only for EMI math (or duplicate minimal EMI calc in stress to keep dependency one-way if preferred).

---

## 4. Implementation Order (Recommended)

1. **Session (GET)**  
   - Add `GET /api/v1/session` handler; register in main.  
   - No new DB table; use existing `calculation_sessions` and cookie.

2. **Finance steps (POST, GET)**  
   - Migration: add `finance_steps` (or `session_finance_input`) table.  
   - sqlc: UpsertFinanceSteps, GetFinanceStepsBySessionID.  
   - Handlers: POST/GET `/api/v1/finance-steps`; register in main.

3. **Stress package + classification**  
   - Create `internal/stress` with models, metrics, policy, classification pipeline.  
   - Handlers: `POST /api/v1/classification` (and optionally `POST /api/v1/stress/metrics`).  
   - Register routes in main.

4. **Debt health GET**  
   - Add `GET /api/v1/debt-health` that uses session → finance_steps (or body) and optionally calls stress to get aggregate score / classification.  
   - Register in main.

5. **Loan simulation input (if not merged with finance-steps)**  
   - If you chose separate simulation input: table + POST/GET `/api/v1/simulation/input` and register.

6. **PATCH /api/debts/{id}/health**  
   - Uncomment in main when auth is enabled; no change to handler.

---

## 5. Summary Table: New/Updated Endpoints

| Method | Path | New? | Notes |
|--------|------|------|--------|
| `GET`  | `/api/v1/session` | Yes | Session info from cookie/query. |
| `POST` | `/api/v1/finance-steps` | Yes | Save manual finance input (loans, income, obligations) per session. |
| `GET`  | `/api/v1/finance-steps` | Yes | Get saved finance steps for current session. |
| `GET`  | `/api/v1/debt-health` | Yes | Portfolio debt health summary (session or body). |
| `POST` | `/api/v1/classification` | Yes | Classification + metrics + flags (stress package). |
| `POST` | `/api/v1/stress/metrics` | Optional | Stress metrics only (stress package). |
| `POST` | `/api/v1/simulation/input` | Optional | Only if not merged with finance-steps. |
| `GET`  | `/api/v1/simulation/input` | Optional | Only if not merged with finance-steps. |
| `POST` | `/api/v1/calculate/baseline` | No | Existing; “test your loan” baseline. |
| `POST` | `/api/v1/calculate/strategy` | No | Existing; “test your loan” strategy. |
| `PATCH` | `/api/debts/{id}/health` | No | Exists; uncomment in main when auth is on. |

---

## 6. Consistency with Existing Conventions

- **Money:** `decimal.Decimal` everywhere (no float for money).  
- **Dates:** `time.Time` in Go; ISO/RFC3339 in JSON.  
- **Handlers:** Use `r.Context()`, JSON helpers, proper status codes; stack with `middleware.Logging`.  
- **Validation:** Validate request bodies (e.g. required fields, loan array non-empty); return 400 with clear error messages.  
- **Session:** Prefer cookie `calc_session`; support optional `session_id` in body/query where specified.  
- **Docs:** Update `docs/API.md` and `docs/ARCHITECTURE.md` when adding routes.

This plan gives you a clear path to implement the API surface between finance steps (manual input), session storage, debt health, test-your-loan simulation, loan simulation input, classification debt health, and the new stress folder with RESTful endpoints.

---

## 7. Testing

- **Unit tests (no DB):** `go test ./internal/stress/...` runs stress metrics and classification tests.
- **Integration tests (DB required):**
  1. Apply migrations (including `20260308120000_create_finance_steps.sql`).
  2. Set `DB_URL` (e.g. `postgres://user:pass@localhost/debtease?sslmode=disable`).
  3. Run: `go test ./cmd -tags=integration -v`.
  Tests cover: `GET /api/v1/session`, `POST`/`GET /api/v1/finance-steps`, `GET /api/v1/debt-health`, `POST /api/v1/classification`, `POST /api/v1/stress/metrics`. If `DB_URL` is not set, integration tests are skipped.
