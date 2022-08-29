package json

// Provider implement the methods needed for creating json.
type Provider interface {
	Title() string
	Footer() string
}
