# Debtease — Debt Classification Domain (v1)

**Status:** v1 (Foundational)
**Owner:** Core Decision Intelligence
**Non-goal:** Credit scoring, default prediction, lending decisions

---

## 1. Why this domain exists (and why it must be separate)

### Problem Statement

Most debt tools treat all debts as equivalent numeric objects.
In reality, **the danger of a debt emerges only in context**.

A ₹10,000 BNPL can be more dangerous than a ₹10L home loan **for the same user**.

### Purpose of this domain

> To interpret *what a debt means right now* for a user, given their financial context.

This domain:

* does **not** compute EMIs (repayment domain does that)
* does **not** predict default (banks do that)
* does **not** judge users

It **classifies debts into treatment categories** that guide safer decisions.

---

## 2. User Risk vs Bank Risk (Critical distinction)

### Bank / NBFC Risk

**Question banks ask:**

> “Will this borrower repay *this* loan?”

**Characteristics**

* Portfolio-level
* Probability-based (PD, LGD, EAD)
* Optimized for lender losses
* Backward-looking (credit history)

**Outcome**

* Approve / reject / price the loan

---

### User Risk (Debtease’s scope)

**Question Debtease asks:**

> “Does this debt make the user’s financial life fragile?”

**Characteristics**

* Individual-level
* Structural, not probabilistic
* Forward-looking (cashflow + stress)
* Focused on regret and lock-in

**Outcome**

* Avoid / delay / restructure / prioritize

**Key rule**

> Debtease never predicts default.
> It detects *fragility conditions*.

This distinction is non-negotiable for:

* ethics
* compliance
* long-term trust

---

## 3. Inputs to the Classification Domain

### 3.1 Objective Inputs (from other domains)

From **Repayment Domain**

* EMI
* Remaining tenure
* Interest type (fixed v1)
* Secured / unsecured
* Debt instrument type

From **Financial State Domain**

* Monthly net income
* Fixed obligations
* Existing EMIs (portfolio)
* Count of short-term obligations (BNPL, 0-cost EMI, LOC)

---

### 3.2 Subjective / Policy Inputs

* Policy thresholds (defined below)
* Risk tolerance profile (optional later)
* Country-specific assumptions (India v1)

---

## 4. Outputs of the Classification Domain

### Primary Output

A **treatment classification**, not a risk score.

```text
SAFE_TO_HOLD
NEEDS_MONITORING
AVOID_ADDING
ACTIVELY_REDUCE
```

### Secondary Outputs

* Context metrics (computed)
* Flags / reasons (explainability)
* Confidence level (low/medium/high)

---

## 5. Context Metrics (v1) — Computable & Explicit

These metrics **define context**.
They are derived, not stored.

---

### 5.1 EMI Ratio (Cashflow Load)

**What it captures:** monthly rigidity

```
EMI_RATIO = Total_Monthly_EMI / Monthly_Net_Income
```

**Interpretation**

* < 30% → manageable
* 30–45% → tight
* > 45% → fragile

---

### 5.2 Fragmentation Index (Debt Stacking)

**What it captures:** cognitive + operational overload

```
FRAGMENTATION_INDEX =
(Number_of_Active_Debts + BNPL_Count * 0.5)
```

BNPL weighted lower individually, but dangerous in aggregate.

---

### 5.3 Rigidity Score (Lock-in Risk)

**What it captures:** inability to adjust cashflow

```
RIGIDITY_SCORE =
Σ (EMI × Remaining_Months) / Monthly_Income
```

Higher = longer unavoidable commitment.

---

### 5.4 Shock Absorption Capacity

**What it captures:** buffer against income shock

```
SHOCK_BUFFER_MONTHS =
(Income - Fixed_Obligations - EMIs) / EMIs
```

Represents how many months of EMI user can absorb if income dips.

---

### 5.5 Unsecured Exposure Ratio

```
UNSECURED_RATIO =
Unsecured_EMI / Total_EMI
```

High unsecured exposure amplifies downside risk.

---

## 6. Classification Pipeline (End-to-End)

### Step 1: Normalize Debts

Convert all debts into a common internal form:

* EMI
* Tenure
* Secured / unsecured
* Instrument category

No judgement here.

---

### Step 2: Compute Context Metrics

Calculate:

* EMI Ratio
* Fragmentation Index
* Rigidity Score
* Shock Buffer
* Unsecured Ratio

This step is deterministic.

---

### Step 3: Apply Policy Thresholds

Policies define **what is considered dangerous**.

Example:

* EMI_RATIO > 45% → cashflow stress
* FRAGMENTATION_INDEX ≥ 5 → debt loop risk
* SHOCK_BUFFER < 1 → fragile

Policies are versioned and auditable.

---

### Step 4: Assign Treatment Category

Mapping logic (simplified):

| Condition                         | Treatment        |
| --------------------------------- | ---------------- |
| Low EMI ratio + low fragmentation | SAFE_TO_HOLD     |
| Moderate stress, stable structure | NEEDS_MONITORING |
| High fragmentation OR low buffer  | AVOID_ADDING     |
| High unsecured + high rigidity    | ACTIVELY_REDUCE  |

---

### Step 5: Generate Flags (Explainability)

Each decision produces **reasons**:

* `HIGH_CASHFLOW_RIGIDITY`
* `SHORT_TERM_DEBT_STACKING`
* `LOW_SHOCK_BUFFER`
* `UNSECURED_OVEREXPOSURE`

Flags power Smart Debt insights.

---

## 7. Policy Definitions (v1)

Policies are **explicit assumptions**, not hidden logic.

### Core Policy Thresholds (India v1)

```text
MAX_SAFE_EMI_RATIO = 0.35
MAX_FRAGMENTATION_INDEX = 4
MIN_SHOCK_BUFFER_MONTHS = 1.5
MAX_UNSECURED_RATIO = 0.6
```

### Why policies exist

* To allow iteration without breaking logic
* To explain decisions clearly
* To adapt regionally later

---

## 8. Invariants (Must Always Hold)

These protect the system from lying.

1. Classification must be **recomputable**
2. Same inputs → same output
3. No probabilistic language
4. No irreversible labels
5. User can override planner decisions

Violating these breaks trust.

---

## 9. End-to-End Example

### User Context

* Income: ₹60,000
* Existing EMIs: ₹22,000
* BNPL: 3 active
* New loan considered: ₹3L personal loan

---

### Computed Metrics

* EMI Ratio: 36%
* Fragmentation Index: 4.5
* Shock Buffer: ~0.7 months
* Unsecured Ratio: 100%

---

### Classification Output

```text
Treatment: AVOID_ADDING
Flags:
- LOW_SHOCK_BUFFER
- SHORT_TERM_DEBT_STACKING
- HIGH_UNSECURED_EXPOSURE
```

---

### Interpretation (UX Layer)

> “Even though the EMI looks affordable, your cashflow is fragile.
> Adding this loan increases the chance of getting stuck in short-term rollovers.”

No prediction.
No judgement.
Just structure.

---

## 10. Legal, Ethical & Positioning Safeguards

### What Debtease explicitly does NOT do

* No credit scoring
* No default prediction
* No lending decision
* No data selling

### Allowed positioning

* “Decision support”
* “Financial wellness”
* “Debt awareness & planning”

### Language rules

❌ “You will default”
❌ “High-risk borrower”

✅ “Increases financial stress”
✅ “Reduces flexibility”

---

## 11. Why this domain is future-proof

Because later you can:

* Add new metrics without touching repayment
* Adjust policy thresholds
* Expose this domain as B2B API
* Localize for other countries

All without breaking UX.

---

## 12. Things people usually miss (you didn’t, but worth stating)

* Risk ≠ probability
* Context ≠ raw data
* Classification ≠ advice
* Transparency > accuracy early
* Versioning beats “smart” logic

---

## Final Principle (North Star)

> **Debtease classifies debt to protect users from regret, not to judge their creditworthiness.**

If this principle holds, every future feature stays aligned.
