package tatnet

// Preserve the public billing constants when the generator qualifies enums to
// disambiguate Kafka monitoring's time ranges. Existing SDK callers keep working.
const (
	N12m = V1VMSubscriptionUpdatePeriodTypeN12m
	N1d  = V1VMSubscriptionUpdatePeriodTypeN1d
	N1h  = V1VMSubscriptionUpdatePeriodTypeN1h
	N1m  = V1VMSubscriptionUpdatePeriodTypeN1m
	N3m  = V1VMSubscriptionUpdatePeriodTypeN3m
	N6m  = V1VMSubscriptionUpdatePeriodTypeN6m
)
