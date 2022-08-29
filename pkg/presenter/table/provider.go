package table

// Provider implement the methods needed for creating table.
type Provider interface {
	Title() string
	Footer() string
	Header() []string
	Rows() [][]string
}
