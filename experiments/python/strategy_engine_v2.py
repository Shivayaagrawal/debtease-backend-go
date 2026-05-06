from dataclasses import dataclass
from typing import Dict, List, Optional
from loan_portfolio_engine import Loan, EMIRecord, calculate_emi

class StrategyType:
    BASELINE = "BASELINE"
    SNOWBALL = "SNOWBALL"
    AVALANCHE = "AVALANCHE"

@dataclass
class LoanState:
    loan: Loan
    balance: float
    emi: float
    active: bool = True

def select_target(states: Dict[str, LoanState], strategy: str) -> Optional[str]:
    active = {k: v for k, v in states.items() if v.active}

    if not active:
        return None

    if strategy == StrategyType.SNOWBALL:
        return min(active, key=lambda k: active[k].balance)

    if strategy == StrategyType.AVALANCHE:
        return max(active, key=lambda k: active[k].loan.annual_rate)

    return None

def simulate_with_strategy(
    loans: List[Loan],
    strategy: str,
    monthly_extra: float = 0.0,
) -> Dict[int, List[EMIRecord]]:

    # Initialize live loan states
    states: Dict[str, LoanState] = {}
    for loan in loans:
        states[loan.loan_id] = LoanState(
            loan=loan,
            balance=loan.principal,
            emi=calculate_emi(loan.principal, loan.annual_rate, loan.tenure_months),
        )

    portfolio: Dict[int, List[EMIRecord]] = {}
    month = 1

    while any(s.active for s in states.values()):
        month_records: List[EMIRecord] = []

        # --------- NORMAL EMI PROCESS ---------
        for loan_id, state in states.items():
            if not state.active:
                continue

            r = state.loan.annual_rate / 12 / 100
            opening = state.balance
            interest = round(opening * r, 2)

            effective_emi = min(state.emi, round(opening + interest, 2))
            principal = round(effective_emi - interest, 2)
            closing = round(opening - principal, 2)

            if closing <= 0:
                closing = 0
                state.active = False

            state.balance = closing

            month_records.append(
                EMIRecord(
                    loan_id=loan_id,
                    month=month,
                    opening_balance=opening,
                    emi=effective_emi,
                    interest=interest,
                    principal=principal,
                    closing_balance=closing,
                )
            )

        # --------- PREPAYMENT STEP ---------
        if strategy != StrategyType.BASELINE and monthly_extra > 0:
            target_id = select_target(states, strategy)

            if target_id:
                target = states[target_id]
                extra = min(monthly_extra, target.balance)

                target.balance = round(target.balance - extra, 2)

                # reflect extra in this month's record
                for r in month_records:
                    if r.loan_id == target_id:
                        r.principal += extra
                        r.closing_balance = target.balance
                        break

                if target.balance <= 0:
                    target.active = False

        # --------- SAFETY NETS ---------
        for r in month_records:
            assert r.closing_balance >= 0
            assert r.interest >= 0
            assert r.principal >= 0

        portfolio[month] = month_records
        month += 1

    return portfolio
