
# 📘 Loan Engine Documentation

**Version:** v1 (Locked)
**Scope:** Fixed-rate term loans (no credit cards)

---

## 1. What Problem This Engine Solves

Most people have **multiple loans**:

* Home loan
* Car loan
* Personal loan
* Education loan

Each loan:

* Has a different interest rate
* Has a different tenure
* Charges interest on a **reducing balance**

The goal of this engine is to:

1. Accurately simulate **how banks calculate EMIs**
2. Generate a **month-by-month schedule**
3. Support **multiple loans together**
4. Show **exact cash outflow and balance changes**

No assumptions. No shortcuts.

---

## 2. Core Concepts (No Finance Background Required)

### 2.1 What Is EMI?

**EMI (Equated Monthly Installment)** is a fixed amount you pay every month.

Each EMI has **two parts**:

* **Interest** → cost of borrowing
* **Principal** → repayment of the loan itself

At the start:

* Interest is high
* Principal repayment is low

Over time:

* Interest reduces
* Principal repayment increases

---

## 3. How Banks Actually Calculate EMI

### 3.1 Interest Rate Conversion

Banks quote interest annually, but charge monthly.

If:

* Annual rate = 9%

Then:

```
Monthly rate = 9 / 12 / 100 = 0.0075
```

---

### 3.2 EMI Formula (Industry Standard)

For a loan with:

* Principal = P
* Monthly rate = r
* Tenure = n months

```
EMI = P × r × (1+r)^n / ((1+r)^n − 1)
```

This formula ensures:

* EMI stays constant
* Loan fully repaid in `n` months
* Interest is correctly distributed over time

✅ This is the **same formula banks use**

---

## 4. Monthly Loan Lifecycle (Critical Section)

Every month, **banks follow this exact order**:

### Step 1: Start with opening balance

```
Opening Balance = Outstanding loan from last month
```

### Step 2: Interest accrues

```
Interest = Opening Balance × Monthly Rate
```

### Step 3: EMI is paid

```
Principal Component = EMI − Interest
```

### Step 4: Balance reduces

```
Closing Balance = Opening Balance − Principal Component
```

⚠️ **Interest is always calculated BEFORE EMI payment**

This order is **non-negotiable** and is enforced in the engine.

---

## 5. Why Reducing Balance Matters

Interest is **not calculated on original loan amount**, but on:

```
Current outstanding balance
```

That’s why:

* Early payments are interest-heavy
* Prepayments save a lot of interest

This engine strictly follows **reducing balance logic**.

---

## 6. Moratorium (Loan Pause) Handling

Sometimes banks allow a **moratorium**:

* EMIs are paused
* Interest continues to accrue

### What happens during moratorium?

Each moratorium month:

```
Interest = Balance × Monthly Rate
New Balance = Balance + Interest
```

No EMI is paid, but debt grows.

This engine:

* Capitalizes interest correctly
* Extends loan naturally

---

## 7. Loan Closure & Last EMI Adjustment

Banks **do not overcharge** on the final EMI.

If:

```
Remaining Balance + Interest < EMI
```

Then:

* EMI is reduced
* Loan closes exactly at zero

This engine automatically:

* Adjusts the last EMI
* Prevents overpayment

---

## 8. Single Loan Engine (File 1)

### `loan_engine.py`

**Purpose:**
Simulate one loan exactly like a bank ledger.

**Key outputs per month:**

* Opening balance
* EMI
* Interest
* Principal repaid
* Closing balance

**Guarantees:**

* Deterministic results
* Bank-accurate math
* Safe edge handling

---

## 9. Multiple Loans & Portfolio View (File 2)

### `loan_portfolio_engine.py`

Real users don’t have one loan — they have **many**.

This layer:

* Simulates each loan independently
* Merges them month-by-month
* Shows combined cash outflow

### Why loans are simulated independently

Because:

* Banks don’t care about your other loans
* Interest rules differ
* Tenures differ

So:

```
Loan A math ≠ Loan B math
```

Then we aggregate.

---

## 10. Portfolio-Level Monthly View

For each month, we compute:

* Total EMI paid
* Total interest paid
* Total principal reduction

This answers:

* “How much cash do I need every month?”
* “How much is going to interest vs repayment?”

---

## 11. Example: Two Loans Running Together

**Month 5**

```
Home Loan EMI     = ₹26,000
Car Loan EMI      = ₹10,200
-------------------------
Total EMI         = ₹36,200
Interest Paid     = ₹19,800
Principal Paid    = ₹16,400
```

This is exactly what users want to see.

---

## 12. What This Engine Does NOT Do (By Design)

❌ Floating rate changes
❌ Credit card interest
❌ Daily compounding
❌ Strategy logic (snowball/avalanche)

These are **separate layers**, not bugs.

---

## 13. Why This Engine Is Safe to Build On

✔ No hidden assumptions
✔ Matches bank statements
✔ Deterministic outputs
✔ Strategy-ready
✔ Audit-friendly

This is a **foundation**, not a hack.

---

## 14. Next Logical Extensions (Future)

Once this engine is locked:

1. Add **prepayment strategies**
2. Add **RBI-compliant constraints**
3. Add **visual timelines**
4. Add **real statement validation**

But **this core stays untouched**.

---

## Final Note (Important)

Most fintech products fail because:

* They mix strategy logic with math
* They approximate interest
* They hide assumptions
