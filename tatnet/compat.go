package tatnet

// Preserve subscription constant names exported before Kafka history introduced
// another enum containing 1h. Generated names are now qualified by their type.
const (
	N12m = V1VMSubscriptionUpdatePeriodTypeN12m
	N1d  = V1VMSubscriptionUpdatePeriodTypeN1d
	N1h  = V1VMSubscriptionUpdatePeriodTypeN1h
	N1m  = V1VMSubscriptionUpdatePeriodTypeN1m
	N3m  = V1VMSubscriptionUpdatePeriodTypeN3m
	N6m  = V1VMSubscriptionUpdatePeriodTypeN6m
)
