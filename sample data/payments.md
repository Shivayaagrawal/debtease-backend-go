<think>
The user wants sample data for testing the payment API endpoints. I need to provide JSON request bodies for the payment endpoints:
1. Record Payment - Minimum
2. Record Payment - Full
3. Record Payment - Extra
4. List Payments (no body needed)

Let me provide realistic payment test data based on the payment model structure.
</think>
Sample payment request data:

## Payment Test Data

### 1. Record Payment - Minimum Payment

```json
{
  "payment_date": "2025-11-15T00:00:00Z",
  "type": "minimum",
  "notes": "Monthly minimum payment processed"
}
```

### 2. Record Payment - Full Payment

```json
{
  "payment_date": "2025-11-15T00:00:00Z",
  "type": "full",
  "payment_method": "bank_transfer",
  "notes": "Paid off debt in full"
}
```

### 3. Record Payment - Extra Payment (Small)

```json
{
  "payment_date": "2025-11-15T00:00:00Z",
  "type": "extra",
  "amount": "50.00",
  "payment_method": "credit_card",
  "notes": "Small extra payment to reduce principal"
}
```

### 4. Record Payment - Extra Payment (Large)

```json
{
  "payment_date": "2025-11-15T00:00:00Z",
  "type": "extra",
  "amount": "500.00",
  "payment_method": "bank_transfer",
  "notes": "Large extra payment to accelerate payoff"
}
```

### 5. Record Payment - Minimum (Past Date)

```json
{
  "payment_date": "2025-11-01T00:00:00Z",
  "type": "minimum",
  "payment_method": "auto_pay",
  "notes": "Previous month payment"
}
```

### 6. Record Payment - Extra (Future Date)

```json
{
  "payment_date": "2025-11-20T00:00:00Z",
  "type": "extra",
  "amount": "200.00",
  "payment_method": "check",
  "notes": "Scheduled extra payment for later this month"
}
```

### 7. Record Payment - Minimum (Minimal Fields)

```json
{
  "payment_date": "2025-11-15T00:00:00Z",
  "type": "minimum"
}
```

### 8. Record Payment - Full (No Notes)

```json
{
  "payment_date": "2025-11-15T00:00:00Z",
  "type": "full",
  "payment_method": "cash"
}
```

## Payment Method Options
Use any of these for `payment_method`:
- `"bank_transfer"`
- `"credit_card"`
- `"debit_card"`
- `"check"`
- `"auto_pay"`
- `"cash"`
- `"online_payment"`
- `"mobile_payment"`

## Testing Sequence

1. Create a debt (use previous sample data)
2. Record Minimum Payment → checks auto-calculation
3. Record Extra Payment → uses specified amount
4. Record Full Payment → pays remaining balance
5. List Payments → see all recorded payments

## Important Notes

- `payment_date`: ISO 8601 format (e.g., `2025-11-15T00:00:00Z`)
- `type`: Must be `"minimum"`, `"full"`, or `"extra"`
- `amount`: Required for `"extra"` type, optional for others
- Amounts are auto-calculated for `"minimum"` and `"full"` types
- All fields except `payment_date` and `type` are optional

Start with samples #1, #3, and #2 to cover minimum, extra, and full scenarios.