package gui

// FormState holds the current form input state.
type FormState struct {
	MRURL             string
	Team              string
	Action            string
	Timezone          string
	MigrationsApplied bool
}
