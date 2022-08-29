package cyclondx

// Provider implement the methods needed for creating cycloneDx xml.
type Provider interface {
	Title() string
	Footer() string
	CycloneDXDoc() ([]byte, error)
}
