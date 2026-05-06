package stress

// Classify returns treatment category, confidence, and flags from metrics (v1 policy).
func Classify(m StressMetrics) (classification, confidence string, flags []string) {
	if m.EMIRatio > 0.45 {
		flags = append(flags, FlagHighCashflowRigidity)
	}
	if m.FragmentationIndex >= 5 {
		flags = append(flags, FlagShortTermDebtStacking)
	}
	if m.ShockBufferMonths < 1 {
		flags = append(flags, FlagLowShockBuffer)
	}
	if m.UnsecuredRatio > MaxUnsecuredRatio {
		flags = append(flags, FlagUnsecuredOverexposure)
	}

	// Mapping logic (simplified v1)
	if m.EMIRatio <= MaxSafeEMIRatio && m.FragmentationIndex < MaxFragmentationIndex && m.ShockBufferMonths >= MinShockBufferMonths {
		classification = ClassificationSafeToHold
		confidence = "high"
		return
	}
	if m.FragmentationIndex >= 5 || m.ShockBufferMonths < 1 {
		classification = ClassificationAvoidAdding
		confidence = "medium"
		return
	}
	if m.UnsecuredRatio > MaxUnsecuredRatio && m.RigidityScore > 0.4 {
		classification = ClassificationActivelyReduce
		confidence = "medium"
		return
	}
	classification = ClassificationNeedsMonitoring
	confidence = "medium"
	return
}
