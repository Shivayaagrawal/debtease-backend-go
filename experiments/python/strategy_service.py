from dataclasses import dataclass
from typing import Dict, List, Callable, Optional
from loan_portfolio_engine import Loan, EMIRecord, generate_loan_schedule , generate_portfolio_schedule

class StrategyType:
    SNOWBALL = "SNOWBALL"       # smallest balance first
    AVALANCHE = "AVALANCHE"     # highest interest rate first
    HYBRID = "HYBRID"           # snowball first loan, then avalanche

@dataclass
class StrategyPhase:
    start_month: int
    strategy: str

def select_target_loan(
    active_loans: Dict[str, float],
    loans: Dict[str, Loan],
    strategy: str,
    snowball_done: bool = False,
) -> Optional[str]:

    if not active_loans:
        return None

    if strategy == StrategyType.SNOWBALL:
        return min(active_loans, key=lambda lid: active_loans[lid])

    if strategy == StrategyType.AVALANCHE:
        return max(active_loans, key=lambda lid: loans[lid].annual_rate)

    if strategy == StrategyType.HYBRID:
        if not snowball_done:
            return min(active_loans, key=lambda lid: active_loans[lid])
        return max(active_loans, key=lambda lid: loans[lid].annual_rate)

    raise ValueError("Unknown strategy")

def apply_prepayment_strategy(
    loans: List[Loan],
    strategy_phases: List[StrategyPhase],
    monthly_extra: float,
) -> Dict[int, List[EMIRecord]]:

    # Start from baseline
    portfolio = generate_portfolio_schedule(loans)

    loans_map = {l.loan_id: l for l in loans}

    balances = {}
    for records in portfolio.values():
        for r in records:
            balances[r.loan_id] = r.closing_balance

    snowball_done = False
    phase_index = 0
    current_strategy = strategy_phases[0].strategy

    for month in sorted(portfolio.keys()):

        # Update strategy phase
        if (
            phase_index + 1 < len(strategy_phases)
            and month >= strategy_phases[phase_index + 1].start_month
        ):
            phase_index += 1
            current_strategy = strategy_phases[phase_index].strategy

        # Active loans
        active_loans = {
            r.loan_id: r.closing_balance
            for r in portfolio.get(month, [])
            if r.closing_balance > 0
        }

        if not active_loans or monthly_extra <= 0:
            continue

        target = select_target_loan(
            active_loans,
            loans_map,
            current_strategy,
            snowball_done,
        )

        if not target:
            continue

        extra = min(monthly_extra, active_loans[target])

        # Apply extra principal to THIS month’s record
        for r in portfolio[month]:
            if r.loan_id == target:
                r.principal += extra
                r.closing_balance = round(r.closing_balance - extra, 2)
                balances[target] = r.closing_balance

                if (
                    current_strategy == StrategyType.HYBRID
                    and not snowball_done
                    and r.closing_balance == 0
                ):
                    snowball_done = True

                break

    return portfolio


if __name__ == "__main__":

    loans = [
        Loan("HL", 20_00_000, 8.5, 240),
        Loan("PL", 5_00_000, 13.0, 60),
        Loan("CL", 6_00_000, 9.5, 84),
    ]

    phases = [
        StrategyPhase(start_month=1, strategy=StrategyType.SNOWBALL),
        StrategyPhase(start_month=12, strategy=StrategyType.AVALANCHE),
    ]

    result = apply_prepayment_strategy(
        loans=loans,
        strategy_phases=phases,
        monthly_extra=15_000,
    )

    print("Strategy applied successfully.")
