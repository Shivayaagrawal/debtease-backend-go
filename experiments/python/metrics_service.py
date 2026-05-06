from typing import Dict, List
from loan_portfolio_engine import EMIRecord


def summarize_portfolio(schedule: Dict[int, List[EMIRecord]]) -> dict:
    assert schedule is not None, "Schedule cannot be None"
    assert isinstance(schedule, dict), "Invalid schedule type"
    total_emi = 0
    total_interest = 0
    total_principal = 0

    last_month = max(schedule.keys())

    for records in schedule.values():
        for r in records:
            total_emi += r.emi
            total_interest += r.interest
            total_principal += r.principal

    return {
        "months": last_month,
        "total_paid": round(total_emi, 2),
        "total_interest": round(total_interest, 2),
        "total_principal": round(total_principal, 2),
    }

def cashflow_delta_timeline(
    baseline: Dict[int, List[EMIRecord]],
    strategy: Dict[int, List[EMIRecord]],
) -> List[dict]:

    timeline = []
    all_months = sorted(set(baseline.keys()) | set(strategy.keys()))

    for month in all_months:
        base_emi = sum(r.emi for r in baseline.get(month, []))
        strat_emi = sum(r.emi for r in strategy.get(month, []))

        timeline.append(
            {
                "month": month,
                "baseline_emi": round(base_emi, 2),
                "strategy_emi": round(strat_emi, 2),
                "delta": round(base_emi - strat_emi, 2),
            }
        )

    return timeline

def debt_free_month(schedule):
    for month in sorted(schedule.keys()):
        if sum(r.closing_balance for r in schedule[month]) == 0:
            return month
    return None

def first_emi_relief_month(schedule):
    prev_total_emi = None

    for month in sorted(schedule.keys()):
        current_total_emi = sum(r.emi for r in schedule[month])

        if prev_total_emi is not None and current_total_emi < prev_total_emi:
            return month

        prev_total_emi = current_total_emi

    return None


def principal_repaid_upto(schedule, upto_month):
    return sum(
        r.principal
        for m in schedule
        if m <= upto_month
        for r in schedule[m]
    )


def interest_repaid_upto(schedule, upto_month):
    return sum(
        r.interest
        for m in schedule
        if m <= upto_month
        for r in schedule[m]
    )


def emi_drop_events(schedule):
    events = []
    prev_emi = None

    for month in sorted(schedule.keys()):
        current_emi = sum(r.emi for r in schedule[month])

        if prev_emi is not None and current_emi < prev_emi:
            events.append({
                "month": month,
                "emi_reduced_by": round(prev_emi - current_emi, 2),
                "new_emi": round(current_emi, 2),
            })

        prev_emi = current_emi

    return events

def principal_delta_at(schedule_base, schedule_strategy, month):
    base = sum(
        r.principal
        for m in schedule_base if m <= month
        for r in schedule_base[m]
    )
    strat = sum(
        r.principal
        for m in schedule_strategy if m <= month
        for r in schedule_strategy[m]
    )
    return round(strat - base, 2)

