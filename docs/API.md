# DebtEase API Documentation

## Table of Contents
- [Base URL](#base-url)
- [Authentication](#authentication)
- [Data Types](#data-types)
- [Error Responses](#error-responses)
- [Endpoints](#endpoints)
  - [Health Check](#health-check)
  - [Authentication](#authentication-endpoints)
  - [Users](#user-endpoints)
  - [Dashboard](#dashboard-endpoints)
  - [Debts](#debt-endpoints)
  - [Payments](#payment-endpoints)
  - [Admin](#admin-endpoints)

## Base URL

```
Development: http://localhost:8080
Production: [Your production URL]
```

All API endpoints are prefixed with `/api/` except health check and admin endpoints.

## Authentication

The API uses JWT (JSON Web Tokens) for authentication. Most endpoints require an `Authorization` header with a Bearer token.

### Authentication Flow

1. **Register** a new user account via `POST /api/register`
2. **Login** with credentials via `POST /api/login` to receive:
   - `token` (access token, valid for 1 hour)
   - `refresh_token` (valid for 60 days)
3. Include the access token in subsequent requests:
   ```
   Authorization: Bearer <access_token>
   ```

### Token Refresh

Access tokens expire after 1 hour. Use the refresh token to obtain a new access token (refresh token endpoint not yet implemented in main.go, but handler exists).

## Data Types

### Currency
**All monetary values are in Indian Rupees (INR, ₹).** The backend stores amounts as decimal numbers without currency symbols. Currency formatting (₹ symbol, comma separators) should be handled in the frontend/Flutter app.

**Important:** The API is currency-agnostic at the core - all calculations work with decimal numbers. Currently, all values are assumed to be in INR. For multi-currency support in the future, currency codes would need to be added to the schema.

### Decimal Values
All monetary values (balances, payments, interest rates) are returned as strings in JSON to maintain precision. Use `decimal.Decimal` in Go or parse as string in Flutter.

**Flutter Example:**
```dart
import 'package:decimal/decimal.dart';
import 'package:intl/intl.dart';

final balance = Decimal.parse(json['outstanding_balance']);

// Format as Indian Rupees
final formatter = NumberFormat.currency(locale: 'en_IN', symbol: '₹');
final formatted = formatter.format(balance.toDouble());
// Output: ₹1,90,055.50
```

### Date/Time Formats
- **Request**: ISO 8601 format (RFC3339) - `"2025-12-01T00:00:00Z"` or date-only `"2025-12-01"`
- **Response**: 
  - Dates: `"YYYY-MM-DD"` format (e.g., `"2025-12-01"`)
  - Timestamps: RFC3339 format (e.g., `"2025-11-15T10:30:00Z"`)

**Flutter Example:**
```dart
final dueDate = DateTime.parse(json['due_date']); // For date-only, add time
```

### UUIDs
All IDs are UUIDs returned as strings in JSON.

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error message description"
}
```

### Common Status Codes

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `204 No Content` - Request successful, no response body
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Missing or invalid authentication token
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict (e.g., duplicate email)
- `500 Internal Server Error` - Server error

---

## Endpoints

## Health Check

### GET /api/healthz

Check if the API is running.

**Authentication:** None required

**Response:**
- **Status Code:** `200 OK`
- **Content-Type:** `text/plain`
- **Body:** `OK`

**Example:**
```bash
curl http://localhost:8080/api/healthz
```

---

## Authentication Endpoints

### POST /api/login

Authenticate a user and receive access tokens.

**Authentication:** None required

**Request Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123"
}
```

**Response:**
- **Status Code:** `200 OK`
- **Body:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "created_at": "2025-11-01T10:00:00Z",
  "updated_at": "2025-11-01T10:00:00Z",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
}
```

**Error Responses:**
- `400 Bad Request` - Missing email or password
- `401 Unauthorized` - Incorrect email or password
- `500 Internal Server Error` - Server error

**Flutter Example:**
```dart
final response = await http.post(
  Uri.parse('$baseUrl/api/login'),
  headers: {'Content-Type': 'application/json'},
  body: jsonEncode({
    'email': 'user@example.com',
    'password': 'securePassword123',
  }),
);

final data = jsonDecode(response.body);
final token = data['token'];
// Store token for subsequent requests
```

---

### POST /api/register

Create a new user account.

**Authentication:** None required

**Request Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "securePassword123"
}
```

**Field Descriptions:**
- `name` (string, required) - User's full name (min 1 character)
- `email` (string, required) - Valid email address
- `password` (string, required) - Password (min 8 characters)

**Response:**
- **Status Code:** `201 Created`
- **Body:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "John Doe",
  "email": "john@example.com",
  "created_at": "2025-11-01T10:00:00Z",
  "updated_at": "2025-11-01T10:00:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Missing required fields or invalid data
- `409 Conflict` - Email already exists
- `500 Internal Server Error` - Server error

**Flutter Example:**
```dart
final response = await http.post(
  Uri.parse('$baseUrl/api/register'),
  headers: {'Content-Type': 'application/json'},
  body: jsonEncode({
    'name': 'John Doe',
    'email': 'john@example.com',
    'password': 'securePassword123',
  }),
);
```

---

## User Endpoints

### PUT /api/users

Update user profile information.

**Authentication:** Required (Bearer token)

**Request Headers:**
```
Content-Type: application/json
Authorization: Bearer <access_token>
```

**Request Body:**
All fields are optional. Only include fields you want to update.
```json
{
  "name": "John Updated",
  "email": "john.updated@example.com",
  "password": "newSecurePassword123",
  "monthly_income": "50000.00"
}
```

**Field Descriptions:**
- `name` (string, optional) - Updated name
- `email` (string, optional) - Updated email address
- `password` (string, optional) - New password (min 8 characters)
- `monthly_income` (string, optional) - Monthly income in INR (₹) as decimal string

**Response:**
- **Status Code:** `200 OK`
- **Body:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "John Updated",
  "email": "john.updated@example.com",
  "monthly_income": "50000.00",
  "created_at": "2025-11-01T10:00:00Z",
  "updated_at": "2025-11-15T14:30:00Z"
}
```

**Note:** All monetary values are in INR (₹).

**Error Responses:**
- `400 Bad Request` - Invalid data or no fields provided
- `401 Unauthorized` - Missing or invalid token
- `500 Internal Server Error` - Server error

**Flutter Example:**
```dart
final response = await http.put(
  Uri.parse('$baseUrl/api/users'),
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer $token',
  },
  body: jsonEncode({
    'name': 'John Updated',
    'monthly_income': '50000.00', // INR: ₹50,000
  }),
);
```

---

## Dashboard Endpoints

### GET /api/dashboard

Get comprehensive dashboard summary including debt overview, repayment progress, and upcoming payments.

**Authentication:** Required (Bearer token)

**Request Headers:**
```
Authorization: Bearer <access_token>
```

**Response:**
- **Status Code:** `200 OK`
- **Body:**
```json
{
  "total_outstanding": "125000.50",
  "repayment_progress": {
    "total_principal": "150000.00",
    "total_paid": "24999.50",
    "repayment_percentage": "16.66"
  },
  "debt_health": {
    "overall_health_score": "72.50",
    "total_active_debts": 3,
    "overdue_count": 1,
    "high_risk_count": 1
  },
  "expected_debt_free": "2026-08-15T00:00:00Z",
  "months_to_debt_free": 9,
  "monthly_minimum_payment": "4500.00",
  "average_interest_rate": "0.15500",
  "upcoming_emis": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440000",
      "name": "HDFC Credit Card",
      "type": "credit_card",
      "lender": "HDFC Bank",
      "outstanding_balance": "25000.00",
      "accrued_interest": "125.50",
      "payment_amount": "1250.00",
      "due_date": "2025-12-01",
      "payment_due_day": 1,
      "is_overdue": false,
      "late_fees": null,
      "days_until_due": 15
    }
  ],
  "payment_history": {
    "paid_last_30_days": "12500.00",
    "paid_last_90_days": "35000.00",
    "paid_last_year": "120000.00",
    "payments_last_30_days": 8
  }
```

**Note:** All monetary values in the dashboard response are in INR (₹).

**Field Descriptions:**
- `total_outstanding` - Total outstanding debt across all accounts
- `repayment_progress` - Progress tracking for debt repayment
- `debt_health` - Overall debt health metrics
- `upcoming_emis` - List of upcoming payments/EMIs
- `payment_history` - Payment statistics for different time periods

**Error Responses:**
- `401 Unauthorized` - Missing or invalid token
- `500 Internal Server Error` - Server error

**Flutter Example:**
```dart
final response = await http.get(
  Uri.parse('$baseUrl/api/dashboard'),
  headers: {'Authorization': 'Bearer $token'},
);

final dashboard = jsonDecode(response.body);
final totalOutstanding = Decimal.parse(dashboard['total_outstanding']);
```

---

## Debt Endpoints

### POST /api/debts

Create a new debt entry.

**Authentication:** Required (Bearer token)

**Request Headers:**
```
Content-Type: application/json
Authorization: Bearer <access_token>
```

**Request Body:**

The request body varies based on debt type. Common fields:

```json
{
  "name": "HDFC Credit Card",
  "type": "credit_card",
  "lender": "HDFC Bank",
  "principal": "50000.00",
  "outstanding_balance": "35000.00",
  "interest_rate": "0.3600",
  "min_payment": "1750.00",
  "due_date": "2025-12-01T00:00:00Z",
  "payment_frequency": "monthly",
  "debt_taken": "2024-05-01T00:00:00Z",
  "risk_level": "medium",
  "priority": 5,
  "notes": "Credit card debt"
}
```

**Note:** All amounts are in INR (₹). Interest rates are annual (e.g., 0.3600 = 36% APR, typical for Indian credit cards).

**Debt Type-Specific Fields:**

#### Credit Card (`type: "credit_card"`)
```json
{
  "name": "HDFC Credit Card",
  "type": "credit_card",
  "lender": "HDFC Bank",
  "principal": "50000.00",
  "outstanding_balance": "35000.00",
  "interest_rate": "0.3600",
  "min_payment": "1750.00",
  "due_date": "2025-12-01T00:00:00Z",
  "payment_frequency": "monthly",
  "billing_day": 15,
  "grace_period_days": 20,
  "min_payment_percent": "5.00",
  "risk_level": "medium",
  "priority": 5
}
```

**Indian Context:**
- Credit card interest rates in India typically range from 24% to 48% APR (0.24 to 0.48)
- Minimum payment is usually 5% of outstanding balance
- Grace period is typically 18-20 days in India

**Required for credit_card:**
- `billing_day` (int32, 1-31) - Day of month when statement is generated
- `grace_period_days` (int32) - Number of grace period days
- `min_payment_percent` (decimal, optional) - Minimum payment percentage (default: 5.00)

#### Personal Loan (`type: "personal_loan"`)
```json
{
  "name": "Personal Loan",
  "type": "personal_loan",
  "lender": "ICICI Bank",
  "principal": "500000.00",
  "outstanding_balance": "425000.00",
  "interest_rate": "0.1200",
  "min_payment": "17500.00",
  "due_date": "2025-12-15T00:00:00Z",
  "payment_frequency": "monthly",
  "payment_due_day": 15,
  "emi_amount": "17500.00",
  "tenure_months": 36,
  "months_paid": 5,
  "moratorium_until": null
}
```

**Indian Context:**
- Personal loan interest rates in India typically range from 10% to 24% APR
- Loan amounts usually range from ₹50,000 to ₹50,00,000
- Typical tenure: 12 to 60 months

**Loan-specific fields:**
- `payment_due_day` (int32, 1-31, optional) - Day of month payment is due
- `emi_amount` (decimal, optional) - EMI amount (auto-calculated if not provided)
- `tenure_months` (int32, optional) - Loan tenure in months
- `moratorium_until` (datetime, optional) - Moratorium end date

#### Student Loan (`type: "student_loan"`)
```json
{
  "name": "Education Loan",
  "type": "student_loan",
  "lender": "SBI",
  "principal": "1000000.00",
  "outstanding_balance": "850000.00",
  "interest_rate": "0.0850",
  "min_payment": "12000.00",
  "due_date": "2025-12-20T00:00:00Z",
  "payment_frequency": "monthly",
  "debt_taken": "2020-09-01T00:00:00Z",
  "risk_level": "low",
  "priority": 2,
  "payment_due_day": 20,
  "tenure_months": 120,
  "moratorium_until": null
}
```

**Indian Context:**
- Education loan interest rates in India typically range from 7% to 12% APR
- Loan amounts can range from ₹50,000 to ₹1,00,00,000+
- Typical tenure: 5 to 15 years (60 to 180 months)
- Moratorium period is common during course duration + 6-12 months

#### Mortgage (`type: "mortgage"`)
```json
{
  "name": "Home Loan",
  "type": "mortgage",
  "lender": "HDFC Bank",
  "principal": "5000000.00",
  "outstanding_balance": "4500000.00",
  "interest_rate": "0.0850",
  "min_payment": "45000.00",
  "due_date": "2025-12-01T00:00:00Z",
  "payment_frequency": "monthly",
  "debt_taken": "2020-01-15T00:00:00Z",
  "risk_level": "low",
  "priority": 1,
  "payment_due_day": 1,
  "emi_amount": "45000.00",
  "tenure_months": 240
}
```

**Indian Context:**
- Home loan interest rates in India typically range from 7% to 10% APR (as of 2025)
- Loan amounts usually range from ₹5,00,000 to ₹5,00,00,000+
- Typical tenure: 15 to 30 years (180 to 360 months)
- Interest rates may be fixed or floating (repo rate linked)

**Field Descriptions:**
- `name` (string, required) - Debt name/description
- `type` (string, required) - One of: `credit_card`, `personal_loan`, `student_loan`, `mortgage`, `medical`, `other`
- `lender` (string, optional) - Lender name
- `principal` (decimal, required) - Original debt amount
- `outstanding_balance` (decimal, required) - Current outstanding balance
- `interest_rate` (decimal, required) - Annual interest rate as decimal (e.g., 0.3600 = 36% APR, typical for Indian credit cards; 0.0850 = 8.5% APR, typical for home loans)
- `min_payment` (decimal, required) - Minimum payment amount
- `due_date` (datetime, required) - Next payment due date
- `payment_frequency` (string, optional) - One of: `monthly`, `biweekly`, `weekly`
- `debt_taken` (datetime, optional) - Date when debt was taken
- `risk_level` (string, optional) - One of: `low`, `medium`, `high`
- `priority` (int16, optional) - Priority level (1-10)
- `notes` (string, optional) - Additional notes

**Response:**
- **Status Code:** `201 Created`
- **Body:**
```json
{
  "id": "770e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "HDFC Credit Card",
  "type": "credit_card",
  "lender": "HDFC Bank",
  "principal": "50000.00",
  "outstanding_balance": "35000.00",
  "interest_rate": "0.3600",
  "min_payment": "1750.00",
  "due_date": "2025-12-01",
  "payment_frequency": "monthly",
  "is_overdue": false,
  "late_fees": null,
  "other_charges": null,
  "debt_taken": "2024-05-01T00:00:00Z",
  "risk_level": "medium",
  "debt_health_score": null,
  "priority": 5,
  "notes": "Credit card debt",
  "accrued_interest": "0.00",
  "last_accrual_date": "2025-11-15T00:00:00Z",
  "created_at": "2025-11-15T10:00:00Z",
  "updated_at": "2025-11-15T10:00:00Z",
  "payment_due_day": null,
  "emi_amount": null,
  "tenure_months": null,
  "months_paid": null,
  "moratorium_until": null,
  "in_moratorium": null
}
```

**Error Responses:**
- `400 Bad Request` - Missing required fields or invalid data
- `401 Unauthorized` - Missing or invalid token
- `500 Internal Server Error` - Server error

**Flutter Example:**
```dart
final response = await http.post(
  Uri.parse('$baseUrl/api/debts'),
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer $token',
  },
  body: jsonEncode({
    'name': 'HDFC Credit Card',
    'type': 'credit_card',
    'lender': 'HDFC Bank',
    'principal': '50000.00', // ₹50,000
    'outstanding_balance': '35000.00', // ₹35,000
    'interest_rate': '0.3600', // 36% APR
    'min_payment': '1750.00', // ₹1,750 (5% of ₹35,000)
    'due_date': '2025-12-01T00:00:00Z',
    'payment_frequency': 'monthly',
    'billing_day': 15,
    'grace_period_days': 20,
    'min_payment_percent': '5.00',
  }),
);
```

---

### GET /api/debts

List all debts for the authenticated user.

**Authentication:** Required (Bearer token)

**Request Headers:**
```
Authorization: Bearer <access_token>
```

**Response:**
- **Status Code:** `200 OK`
- **Body:** Array of debt objects
```json
[
  {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Visa Card",
    "type": "credit_card",
    "lender": "Visa",
    "principal": "2500.00",
    "outstanding_balance": "1900.55",
    "interest_rate": "0.1999",
    "min_payment": "40.00",
    "due_date": "2025-12-01",
    "payment_frequency": "monthly",
    "is_overdue": false,
    "accrued_interest": "12.50",
    "last_accrual_date": "2025-11-15T00:00:00Z",
    "created_at": "2025-11-15T10:00:00Z",
    "updated_at": "2025-11-15T10:00:00Z"
  }
]
```

**Error Responses:**
- `401 Unauthorized` - Missing or invalid token
- `500 Internal Server Error` - Server error

---

### GET /api/debts/{id}

Get a specific debt by ID.

**Authentication:** Required (Bearer token)

**Request Headers:**
```
Authorization: Bearer <access_token>
```

**Path Parameters:**
- `id` (UUID, required) - Debt ID

**Response:**
- **Status Code:** `200 OK`
- **Body:** Debt object (same structure as POST /api/debts response)

**Error Responses:**
- `400 Bad Request` - Invalid ID format
- `401 Unauthorized` - Missing or invalid token
- `404 Not Found` - Debt not found
- `500 Internal Server Error` - Server error

---

### PATCH /api/debts/{id}/balance

Update the outstanding balance and overdue status of a debt.

**Authentication:** Required (Bearer token)

**Request Headers:**
```
Content-Type: application/json
Authorization: Bearer <access_token>
```

**Path Parameters:**
- `id` (UUID, required) - Debt ID

**Request Body:**
```json
{
  "outstanding_balance": "1850.25",
  "is_overdue": false
}
```

**Field Descriptions:**
- `outstanding_balance` (decimal, required) - New outstanding balance
- `is_overdue` (boolean, required) - Whether the debt is overdue

**Response:**
- **Status Code:** `204 No Content`

**Error Responses:**
- `400 Bad Request` - Invalid data
- `401 Unauthorized` - Missing or invalid token
- `404 Not Found` - Debt not found
- `500 Internal Server Error` - Server error

---

### PATCH /api/debts/{id}/health

Update the debt health score.

**Authentication:** Required (Bearer token)

**Request Headers:**
```
Content-Type: application/json
Authorization: Bearer <access_token>
```

**Path Parameters:**
- `id` (UUID, required) - Debt ID

**Request Body:**
```json
{
  "debt_health_score": "72.50"
}
```

**Field Descriptions:**
- `debt_health_score` (decimal, required) - Health score (0-100)

**Response:**
- **Status Code:** `204 No Content`

**Error Responses:**
- `400 Bad Request` - Invalid data
- `401 Unauthorized` - Missing or invalid token
- `404 Not Found` - Debt not found
- `500 Internal Server Error` - Server error

---

### POST /api/debts/{id}/overdue

Mark a debt as overdue.

**Authentication:** Required (Bearer token)

**Request Headers:**
```
Authorization: Bearer <access_token>
```

**Path Parameters:**
- `id` (UUID, required) - Debt ID

**Response:**
- **Status Code:** `204 No Content`

**Error Responses:**
- `400 Bad Request` - Invalid ID format
- `401 Unauthorized` - Missing or invalid token
- `404 Not Found` - Debt not found
- `500 Internal Server Error` - Server error

---

## Payment Endpoints

### POST /api/debts/{id}/payments

Record a payment for a debt.

**Authentication:** Required (Bearer token)

**Request Headers:**
```
Content-Type: application/json
Authorization: Bearer <access_token>
```

**Path Parameters:**
- `id` (UUID, required) - Debt ID

**Request Body:**

#### Minimum Payment
```json
{
  "payment_date": "2025-11-15T00:00:00Z",
  "type": "minimum",
  "notes": "Monthly minimum payment"
}
```

#### Full Payment
```json
{
  "payment_date": "2025-11-15T00:00:00Z",
  "type": "full",
  "payment_method": "bank_transfer",
  "notes": "Paid off in full"
}
```

#### Extra Payment
```json
{
  "payment_date": "2025-11-15T00:00:00Z",
  "type": "extra",
  "amount": "5000.00",
  "payment_method": "upi",
  "notes": "Extra payment to reduce principal"
}
```

**Note:** All amounts are in INR (₹). Payment methods in India commonly include: `bank_transfer`, `upi`, `credit_card`, `debit_card`, `cash`.

**Field Descriptions:**
- `payment_date` (datetime, required) - Date of payment
- `type` (string, required) - One of: `minimum`, `full`, `extra`
- `amount` (decimal, optional) - Payment amount (required for `extra` type, auto-calculated for `minimum` and `full`)
- `payment_method` (string, optional) - Payment method (e.g., `bank_transfer`, `credit_card`, `cash`)
- `notes` (string, optional) - Payment notes

**Response:**
- **Status Code:** `201 Created`
- **Body:**
```json
{
  "payment": {
    "id": "880e8400-e29b-41d4-a716-446655440000",
    "debt_id": "770e8400-e29b-41d4-a716-446655440000",
    "amount": "1750.00",
    "payment_date": "2025-11-15",
    "type": "minimum",
    "notes": "Monthly minimum payment",
    "principal_applied": "1225.00",
    "interest_applied": "525.00",
    "payment_method": "upi",
    "created_at": "2025-11-15T10:30:00Z"
  },
  "debt": {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    "outstanding_balance": "33250.00",
    "accrued_interest": "0.00",
    "last_accrual_date": "2025-11-15T00:00:00Z"
  }
}
```

**Note:** All amounts are in INR (₹).

**Error Responses:**
- `400 Bad Request` - Invalid data or missing required fields
- `401 Unauthorized` - Missing or invalid token
- `404 Not Found` - Debt not found
- `500 Internal Server Error` - Server error

**Flutter Example:**
```dart
final response = await http.post(
  Uri.parse('$baseUrl/api/debts/$debtId/payments'),
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer $token',
  },
  body: jsonEncode({
    'payment_date': DateTime.now().toIso8601String(),
    'type': 'minimum',
    'notes': 'Monthly minimum payment',
    'payment_method': 'upi', // Common in India: upi, bank_transfer, etc.
  }),
);
```

---

### GET /api/debts/{id}/payments

List all payments for a specific debt.

**Authentication:** Required (Bearer token)

**Request Headers:**
```
Authorization: Bearer <access_token>
```

**Path Parameters:**
- `id` (UUID, required) - Debt ID

**Response:**
- **Status Code:** `200 OK`
- **Body:** Array of payment objects
```json
[
  {
    "id": "880e8400-e29b-41d4-a716-446655440000",
    "debt_id": "770e8400-e29b-41d4-a716-446655440000",
    "amount": "1750.00",
    "payment_date": "2025-11-15",
    "type": "minimum",
    "notes": "Monthly minimum payment",
    "principal_applied": "1225.00",
    "interest_applied": "525.00",
    "payment_method": "upi",
    "created_at": "2025-11-15T10:30:00Z"
  }
]
```

**Note:** All amounts are in INR (₹).

**Error Responses:**
- `400 Bad Request` - Invalid ID format
- `401 Unauthorized` - Missing or invalid token
- `404 Not Found` - Debt not found
- `500 Internal Server Error` - Server error

---

## Admin Endpoints

### GET /admin/metrics

Get admin metrics (file server hit count).

**Authentication:** None required (public endpoint)

**Response:**
- **Status Code:** `200 OK`
- **Content-Type:** `text/html; charset=utf-8`
- **Body:** HTML page with metrics

**Note:** This is primarily for monitoring and may not be needed for Flutter app integration.

---

## Flutter Integration Tips

### 1. HTTP Client Setup
```dart
import 'package:http/http.dart' as http;
import 'dart:convert';
import 'package:decimal/decimal.dart';
import 'package:intl/intl.dart';

class DebtEaseAPI {
  final String baseUrl;
  String? accessToken;

  DebtEaseAPI({this.baseUrl = 'http://localhost:8080'});

  Map<String, String> get headers => {
    'Content-Type': 'application/json',
    if (accessToken != null) 'Authorization': 'Bearer $accessToken',
  };

  // Format INR currency
  static String formatINR(Decimal amount) {
    final formatter = NumberFormat.currency(
      locale: 'en_IN',
      symbol: '₹',
      decimalDigits: 2,
    );
    return formatter.format(amount.toDouble());
  }
}
```

### 2. Error Handling
```dart
void handleError(http.Response response) {
  if (response.statusCode >= 400) {
    final error = jsonDecode(response.body);
    throw Exception(error['error'] ?? 'An error occurred');
  }
}
```

### 3. Decimal Parsing
```dart
Decimal parseDecimal(dynamic value) {
  if (value is String) {
    return Decimal.parse(value);
  } else if (value is num) {
    return Decimal.fromInt(value.toInt());
  }
  throw FormatException('Cannot parse decimal from $value');
}
```

### 4. Date Parsing
```dart
DateTime parseDate(String dateString) {
  // Handle date-only format (YYYY-MM-DD)
  if (dateString.length == 10) {
    return DateTime.parse('${dateString}T00:00:00Z');
  }
  return DateTime.parse(dateString);
}
```

### 5. Currency Formatting (INR)
```dart
import 'package:intl/intl.dart';

String formatINR(double amount) {
  final formatter = NumberFormat.currency(
    locale: 'en_IN', // Indian locale for proper formatting
    symbol: '₹',
    decimalDigits: 2,
  );
  return formatter.format(amount);
  // Example: 50000.50 -> "₹50,000.50"
  // Example: 1000000 -> "₹10,00,000.00" (Indian numbering system)
}
```

---

## Rate Limiting

Currently, the API does not implement rate limiting. However, it's recommended to:
- Implement client-side request throttling
- Cache dashboard and debt list responses when appropriate
- Avoid making excessive requests in short time periods

---

## Best Practices

1. **Token Management**: Store access tokens securely and refresh before expiration
2. **Error Handling**: Always check response status codes and handle errors gracefully
3. **Decimal Precision**: Use `decimal.Decimal` or equivalent for all monetary values
4. **Date Handling**: Be consistent with date formats (ISO 8601)
5. **Request Validation**: Validate data on the client side before sending requests
6. **Offline Support**: Consider caching data for offline access
7. **Currency Display**: Always format monetary values as INR (₹) in the UI. Use Indian locale (`en_IN`) for proper number formatting (e.g., ₹10,00,000.50 instead of ₹1,000,000.50)
8. **Interest Rates**: Be aware of typical Indian interest rates:
   - Credit Cards: 24-48% APR
   - Personal Loans: 10-24% APR
   - Home Loans: 7-10% APR
   - Education Loans: 7-12% APR

---

## Support

For issues or questions, please contact the development team or refer to the project repository.

