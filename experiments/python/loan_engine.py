from dataclasses import dataclass
from typing import List, Optional


@dataclass
class EMIRecord:
    month: int
    opening_balance: float
    emi: float
    interest: float
    principal: float
    closing_balance: float


def calculate_emi(principal: float, annual_rate: float, tenure_months: int) -> float:
    """
    Standard EMI formula (monthly compounding)
    """
    r = annual_rate / 12 / 100
    if r == 0:
        return principal / tenure_months

    emi = principal * r * (1 + r) ** tenure_months / ((1 + r) ** tenure_months - 1)
    return round(emi, 2)


def generate_amortization_schedule(
    principal: float,
    annual_rate: float,
    tenure_months: int,
    moratorium_months: int = 0,
    prepayments: Optional[dict] = None,
    reduce: str = "TENURE",  # or "EMI"
) -> List[EMIRecord]:
    """
    Generates a bank-accurate amortization schedule.

    prepayments: {month_number: amount}
    reduce: TENURE (default) or EMI
    """

    assert reduce in ("TENURE", "EMI")

    r = annual_rate / 12 / 100
    emi = calculate_emi(principal, annual_rate, tenure_months)
    balance = principal

    schedule: List[EMIRecord] = []
    month = 1
    prepayments = prepayments or {}

    # ---- MORATORIUM PERIOD ----
    for _ in range(moratorium_months):
        interest = round(balance * r, 2)
        balance = round(balance + interest, 2)

        schedule.append(
            EMIRecord(
                month=month,
                opening_balance=balance - interest,
                emi=0.0,
                interest=interest,
                principal=0.0,
                closing_balance=balance,
            )
        )
        month += 1

    # ---- REPAYMENT PERIOD ----
    while balance > 0:
        opening_balance = balance
        interest = round(balance * r, 2)

        effective_emi = emi
        if effective_emi > balance + interest:
            effective_emi = round(balance + interest, 2)

        principal_component = round(effective_emi - interest, 2)
        balance = round(balance - principal_component, 2)

        # ---- PREPAYMENT ----
        if month in prepayments:
            balance = round(balance - prepayments[month], 2)
            if balance < 0:
                balance = 0

            if reduce == "EMI" and balance > 0:
                remaining_months = max(1, tenure_months - month + 1)
                emi = calculate_emi(balance, annual_rate, remaining_months)

        schedule.append(
            EMIRecord(
                month=month,
                opening_balance=opening_balance,
                emi=effective_emi,
                interest=interest,
                principal=principal_component,
                closing_balance=balance,
            )
        )

        # Safety guards
        assert balance >= -1e-2
        assert interest >= 0
        assert principal_component >= 0

        month += 1

    return schedule


def summarize_schedule(schedule: List[EMIRecord]) -> dict:
    total_interest = sum(r.interest for r in schedule)
    total_principal = sum(r.principal for r in schedule)
    total_paid = sum(r.emi for r in schedule)

    return {
        "months": len(schedule),
        "total_interest": round(total_interest, 2),
        "total_principal": round(total_principal, 2),
        "total_paid": round(total_paid, 2),
    }


# =========================
# TEST CASES
# =========================
if __name__ == "__main__":
    print("\n--- TEST 1: STANDARD HOME LOAN ---")
    schedule = generate_amortization_schedule(
        principal=10_00_000,
        annual_rate=9.0,
        tenure_months=240,
    )
    summary = summarize_schedule(schedule)
    print(summary)

    print("\n--- TEST 2: LOAN WITH MORATORIUM (6 MONTHS) ---")
    schedule = generate_amortization_schedule(
        principal=5_00_000,
        annual_rate=10.0,
        tenure_months=120,
        moratorium_months=6,
    )
    summary = summarize_schedule(schedule)
    print(summary)

    print("\n--- TEST 3: PREPAYMENT (TENURE REDUCTION) ---")
    schedule = generate_amortization_schedule(
        principal=8_00_000,
        annual_rate=8.5,
        tenure_months=180,
        prepayments={12: 1_00_000},
    )
    summary = summarize_schedule(schedule)
    print(summary)

    print("\n--- TEST 4: PREPAYMENT (EMI REDUCTION) ---")
    schedule = generate_amortization_schedule(
        principal=8_00_000,
        annual_rate=8.5,
        tenure_months=180,
        prepayments={12: 1_00_000},
        reduce="EMI",
    )
    summary = summarize_schedule(schedule)
    print(summary)
