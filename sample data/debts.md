<think>
The user wants sample debt data to test the payments. They've already added the payment requests to Postman, so they just need the JSON body for creating a debt that they can use for testing payments.

I should provide:
1. A Create Debt request body with realistic sample data
2. Maybe a few variations for different scenarios
3. Make sure the data makes sense for testing payments (e.g., has outstanding balance, interest rate, min payment, etc.)

Let me check what fields are required for creating a debt first.
</think>
Creating sample debt data for payment testing:


[1 tool called]

<think>
The user needs sample debt data. I should provide realistic JSON examples that can be used for testing payments. Let me create a few variations that will be good for testing different payment scenarios:

1. Credit card debt - good for testing minimum and extra payments
2. Personal loan - good for testing regular payments
3. Student loan - another option

I should make sure the data has:
- Reasonable outstanding balance
- Interest rate (as decimal, e.g., 0.1999 for 19.99%)
- Minimum payment
- Due date in the future (or past to test overdue scenarios)
- Payment frequency

Let me provide a few options with different scenarios.
</think>
Sample debt data for payment testing:

## Sample Debt Data #1 - Credit Card (High Interest)

```json
{
  "name": "Visa Credit Card",
  "type": "credit_card",
  "lender": "Chase Bank",
  "principal": "5000.00",
  "outstanding_balance": "3850.75",
  "interest_rate": "0.2499",
  "min_payment": "95.00",
  "due_date": "2025-12-15T00:00:00Z",
  "payment_frequency": "monthly",
  "debt_taken": "2024-01-15T00:00:00Z",
  "risk_level": "high",
  "priority": 8,
  "notes": "High APR credit card - prioritize paying off"
}
```

## Sample Debt Data #2 - Personal Loan (Medium Interest)

```json
{
  "name": "Personal Loan",
  "type": "personal_loan",
  "lender": "Wells Fargo",
  "principal": "10000.00",
  "outstanding_balance": "7250.00",
  "interest_rate": "0.1299",
  "min_payment": "350.00",
  "due_date": "2025-11-20T00:00:00Z",
  "payment_frequency": "monthly",
  "debt_taken": "2023-06-01T00:00:00Z",
  "risk_level": "medium",
  "priority": 5,
  "notes": "Consolidation loan - fixed rate"
}
```

## Sample Debt Data #3 - Student Loan (Lower Interest)

```json
{
  "name": "Federal Student Loan",
  "type": "student_loan",
  "lender": "Navient",
  "principal": "25000.00",
  "outstanding_balance": "22000.50",
  "interest_rate": "0.0650",
  "min_payment": "275.00",
  "due_date": "2025-11-10T00:00:00Z",
  "payment_frequency": "monthly",
  "debt_taken": "2020-09-01T00:00:00Z",
  "risk_level": "low",
  "priority": 3,
  "notes": "Federal loan - lower priority"
}
```

## Sample Debt Data #4 - Small Medical Debt (Quick Test)

```json
{
  "name": "Medical Bill",
  "type": "medical",
  "lender": "City Hospital",
  "principal": "1200.00",
  "outstanding_balance": "1200.00",
  "interest_rate": "0.0000",
  "min_payment": "100.00",
  "due_date": "2025-11-30T00:00:00Z",
  "payment_frequency": "monthly",
  "risk_level": "low",
  "priority": 4,
  "notes": "Interest-free medical debt"
}
```

## Testing Scenarios

After creating a debt, test payments with:

1. Minimum payment: Use "Record Payment - Minimum" (auto-calculates from min_payment)
2. Full payment: Use "Record Payment - Full" (pays off entire balance)
3. Extra payment: Use "Record Payment - Extra" with an amount like `"amount": "200.00"`

Use Sample #1 or #2 for realistic interest accrual scenarios. Sample #4 is good for quick tests without interest calculations.