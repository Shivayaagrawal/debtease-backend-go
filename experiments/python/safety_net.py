from loan_portfolio_engine import Loan, generate_portfolio_schedule
from strategy_service import (
    apply_prepayment_strategy,
    StrategyPhase,
    StrategyType,
)
from metrics_service import summarize_portfolio


EPS = 1e-2


def assert_conservation(summary):
    assert abs(
        summary["total_paid"]
        - (summary["total_principal"] + summary["total_interest"])
    ) < EPS, "Money conservation violated"


def test_baseline_closes_cleanly():
    loans = [
        Loan("HL", 10_00_000, 9.0, 120),
    ]
    baseline = generate_portfolio_schedule(loans)
    summary = summarize_portfolio(baseline)

    assert summary["months"] > 0
    assert summary["total_interest"] > 0
    assert_conservation(summary)


def test_avalanche_beats_baseline():
    loans = [
        Loan("HL", 20_00_000, 8.5, 240),
        Loan("PL", 5_00_000, 13.0, 60),
    ]

    baseline = generate_portfolio_schedule(loans)

    avalanche = apply_prepayment_strategy(
        loans,
        strategy_phases=[StrategyPhase(1, StrategyType.AVALANCHE)],
        monthly_extra=10_000,
    )

    base_summary = summarize_portfolio(baseline)
    ava_summary = summarize_portfolio(avalanche)

    assert ava_summary["total_interest"] < base_summary["total_interest"], \
        "Avalanche should never cost more interest"


def test_strategy_switch_does_not_rewrite_past():
    loans = [
        Loan("PL", 5_00_000, 12.0, 60),
    ]

    baseline = generate_portfolio_schedule(loans)

    switched = apply_prepayment_strategy(
        loans,
        strategy_phases=[
            StrategyPhase(1, StrategyType.SNOWBALL),
            StrategyPhase(12, StrategyType.AVALANCHE),
        ],
        monthly_extra=5_000,
    )

    # Compare first 11 months
    for month in range(1, 12):
        base_emi = sum(r.emi for r in baseline.get(month, []))
        switched_emi = sum(r.emi for r in switched.get(month, []))
        assert abs(base_emi - switched_emi) < EPS, \
            "Past EMIs should not change"


def test_no_negative_balances():
    loans = [
        Loan("CL", 3_00_000, 10.0, 36),
    ]

    strategy = apply_prepayment_strategy(
        loans,
        strategy_phases=[StrategyPhase(1, StrategyType.AVALANCHE)],
        monthly_extra=50_000,  # aggressive
    )

    for records in strategy.values():
        for r in records:
            assert r.closing_balance >= -EPS, "Negative balance detected"


if __name__ == "__main__":
    test_baseline_closes_cleanly()
    test_avalanche_beats_baseline()
    test_strategy_switch_does_not_rewrite_past()
    test_no_negative_balances()

    print("✅ ALL SAFETY NET CHECKS PASSED")
