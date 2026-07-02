package gui

// AutoTrigger tracks the last triggered values and detects changes.
type AutoTrigger struct {
	lastAction            string
	lastTimezone          string
	lastMigrationsApplied bool
}

// HasChanges returns true if any tracked value has changed since the last Record call.
func (a *AutoTrigger) HasChanges(action, tz string, migrations bool) bool {
	return a.lastAction != action || a.lastTimezone != tz || a.lastMigrationsApplied != migrations
}

// Record stores the current values for future comparison.
func (a *AutoTrigger) Record(action, tz string, migrations bool) {
	a.lastAction = action
	a.lastTimezone = tz
	a.lastMigrationsApplied = migrations
}
