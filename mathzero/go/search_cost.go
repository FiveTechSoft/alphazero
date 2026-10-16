package main

// Strict relaxation, with no new numeric tolerance or altered edge objective.
func searchRelax(best map[string]float64, key string, cost float64) bool {
	old, exists := best[key]
	if exists && !(cost < old) {
		return false
	}
	best[key] = cost
	return true
}
func searchStale(best map[string]float64, key string, cost float64) bool {
	current, exists := best[key]
	return !exists || cost > current
}
