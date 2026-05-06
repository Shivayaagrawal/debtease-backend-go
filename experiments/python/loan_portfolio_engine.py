from dataclasses import dataclass
from typing import List, Dict


@dataclass
class EMIRecord:
    loan_id: str
    month: int
    opening_balance: float
    emi: float
    interest: float
    principal: float
    closing_balance: float


@dataclass
class Loan:
    loan_id: str
    principal: float
    annual_rate: float
    tenure_months: int
    moratorium_months: int = 0


# ---------------- CORE MATH ---------------- #

def calculate_emi(principal: float, annual_rate: float, tenure_months: int) -> float:
    r = annual_rate / 12 / 100
    if r == 0:
        return round(principal / tenure_months, 2)

    emi = principal * r * (1 + r) ** tenure_months / ((1 + r) ** tenure_months - 1)
    return round(emi, 2)


def generate_loan_schedule(loan: Loan) -> List[EMIRecord]:
    r = loan.annual_rate / 12 / 100
    emi = calculate_emi(loan.principal, loan.annual_rate, loan.tenure_months)
    balance = loan.principal

    schedule: List[EMIRecord] = []
    month = 1

    # ---- MORATORIUM ----
    for _ in range(loan.moratorium_months):
        interest = round(balance * r, 2)
        closing = round(balance + interest, 2)

        schedule.append(
            EMIRecord(
                loan_id=loan.loan_id,
                month=month,
                opening_balance=balance,
                emi=0.0,
                interest=interest,
                principal=0.0,
                closing_balance=closing,
            )
        )
        balance = closing
        month += 1

    # ---- REPAYMENT ----
    while balance > 0:
        opening = balance
        interest = round(balance * r, 2)

        effective_emi = min(emi, round(balance + interest, 2))
        principal_component = round(effective_emi - interest, 2)
        closing = round(balance - principal_component, 2)

        schedule.append(
            EMIRecord(
                loan_id=loan.loan_id,
                month=month,
                opening_balance=opening,
                emi=effective_emi,
                interest=interest,
                principal=principal_component,
                closing_balance=closing,
            )
        )

        assert principal_component >= 0
        assert interest >= 0

        balance = closing
        month += 1

    return schedule


# ---------------- PORTFOLIO ---------------- #

def generate_portfolio_schedule(loans: List[Loan]) -> Dict[int, List[EMIRecord]]:
    portfolio: Dict[int, List[EMIRecord]] = {}

    for loan in loans:
        schedule = generate_loan_schedule(loan)
        for record in schedule:
            portfolio.setdefault(record.month, []).append(record)

    return portfolio


def portfolio_summary(portfolio: Dict[int, List[EMIRecord]]) -> List[dict]:
    summary = []

    for month in sorted(portfolio.keys()):
        records = portfolio[month]
        summary.append(
            {
                "month": month,
                "total_emi": round(sum(r.emi for r in records), 2),
                "total_interest": round(sum(r.interest for r in records), 2),
                "total_principal": round(sum(r.principal for r in records), 2),
            }
        )

    return summary


# ---------------- DISPLAY ---------------- #

def print_exact_schedule(portfolio: Dict[int, List[EMIRecord]]):
    for month in sorted(portfolio.keys()):
        print(f"\nMONTH {month}")
        for r in portfolio[month]:
            print(
                f"  [{r.loan_id}] "
                f"OB={r.opening_balance:.2f} | "
                f"EMI={r.emi:.2f} | "
                f"I={r.interest:.2f} | "
                f"P={r.principal:.2f} | "
                f"CB={r.closing_balance:.2f}"
            )


# =======================
# TEST CASES
# =======================
if __name__ == "__main__":

    loans = [
        Loan(
            loan_id="HOME_LOAN",
            principal=30_00_000,
            annual_rate=8.5,
            tenure_months=240,
        ),
        Loan(
            loan_id="PERSONAL_LOAN",
            principal=5_00_000,
            annual_rate=12.0,
            tenure_months=60,
        ),
        Loan(
            loan_id="CAR_LOAN",
            principal=8_00_000,
            annual_rate=9.5,
            tenure_months=84,
        ),
    ]

    portfolio = generate_portfolio_schedule(loans)

    print("\n=========== EXACT LOAN SCHEDULE ===========")
    print_exact_schedule(portfolio)

    print("\n=========== MONTHLY PORTFOLIO SUMMARY ===========")
    for row in portfolio_summary(portfolio)[:12]:  # first 12 months
        print(row)

