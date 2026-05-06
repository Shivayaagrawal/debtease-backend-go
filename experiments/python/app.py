import streamlit as st
import pandas as pd

from loan_portfolio_engine import Loan
from strategy_engine_v2 import simulate_with_strategy, StrategyType
from metrics_service import (
    summarize_portfolio,
    debt_free_month,
    first_emi_relief_month,
    principal_delta_at,
    emi_drop_events,
)

st.set_page_config(layout="wide")
st.title("💸 Loan Repayment Strategy Simulator")

# =====================
# INPUTS
# =====================
st.sidebar.header("Loans")

loan_count = st.sidebar.number_input("Number of Loans", 1, 5, 2)

loans = []
for i in range(loan_count):
    st.sidebar.subheader(f"Loan {i+1}")
    loan_id = st.sidebar.text_input(f"Loan ID {i+1}", f"L{i+1}")
    principal = st.sidebar.number_input(
        f"Principal {i+1}", 1_000.0, value=5_00_000.0
    )
    rate = st.sidebar.number_input(f"Interest % {i+1}", 0.0, value=10.0)
    tenure = st.sidebar.number_input(f"Tenure (months) {i+1}", 1, value=60)

    loans.append(Loan(loan_id, principal, rate, tenure))

st.sidebar.header("Strategy")

strategy = st.sidebar.selectbox(
    "Strategy",
    [StrategyType.SNOWBALL, StrategyType.AVALANCHE],
)

monthly_extra = st.sidebar.number_input(
    "Monthly Extra Prepayment", 0.0, value=5_000.0
)

# =====================
# SIMULATIONS
# =====================
baseline = simulate_with_strategy(
    loans,
    strategy=StrategyType.BASELINE,
)

strategy_schedule = simulate_with_strategy(
    loans,
    strategy=strategy,
    monthly_extra=monthly_extra,
)

base_summary = summarize_portfolio(baseline)
strat_summary = summarize_portfolio(strategy_schedule)

# =====================
# PRIMARY METRICS
# =====================
st.header("📊 Key Outcomes")

interest_saved = base_summary["total_interest"] - strat_summary["total_interest"]

baseline_relief = first_emi_relief_month(baseline)
strategy_relief = first_emi_relief_month(strategy_schedule)

col1, col2, col3 = st.columns(3)

col1.metric(
    "💰 Interest Saved",
    f"₹{interest_saved:,.0f}",
)

col2.metric(
    "🟢 First EMI Relief Earlier By",
    f"{baseline_relief - strategy_relief} months"
    if baseline_relief and strategy_relief
    else "—",
)


col3.metric(
    "🏁 Debt-Free Month",
    f"Month {debt_free_month(strategy_schedule)}",
)

# =====================
# PROGRESS METRICS
# =====================
st.header("🚀 Progress Acceleration")

p1, p2, p3 = st.columns(3)

p1.metric(
    "Extra Principal @ Month 12",
    f"₹{principal_delta_at(baseline, strategy_schedule, 12):,.0f}",
)

p2.metric(
    "Extra Principal @ Month 24",
    f"₹{principal_delta_at(baseline, strategy_schedule, 24):,.0f}",
)

p3.metric(
    "Extra Principal @ Month 36",
    f"₹{principal_delta_at(baseline, strategy_schedule, 36):,.0f}",
)

# =====================
# EMI DROP EVENTS
# =====================
st.header("📉 EMI Reduction Events")

events = emi_drop_events(strategy_schedule)

if events:
    st.table(events)
else:
    st.info("No EMI reductions yet — EMIs remain constant until a loan closes.")

# =====================
# STRUCTURED SCHEDULE
# =====================
st.header("📆 Monthly Repayment Schedule")

rows = []
for month, records in strategy_schedule.items():
    for r in records:
        rows.append({
            "Month": month,
            "Loan": r.loan_id,
            "EMI": r.emi,
            "Interest": r.interest,
            "Principal": r.principal,
            "Closing Balance": r.closing_balance,
        })

df = pd.DataFrame(rows)

loan_filter = st.multiselect(
    "Filter by Loan",
    options=df["Loan"].unique(),
    default=list(df["Loan"].unique()),
)

month_limit = st.slider("Show months", 1, int(df["Month"].max()), 36)

filtered_df = df[
    (df["Loan"].isin(loan_filter)) &
    (df["Month"] <= month_limit)
]

st.dataframe(filtered_df, width='stretch')

# =====================
# DEBUG (OPTIONAL)
# =====================
with st.expander("🧪 Debug: Raw Schedule JSON"):
    st.json(strategy_schedule)
