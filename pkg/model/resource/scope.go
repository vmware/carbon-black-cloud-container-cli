package resource

// Scope represents a resource scope/id.
type Scope struct {
	Cluster   string            `json:"-"`
	Namespace string            `json:"namespace"`
	Kind      string            `json:"kind"`
	Name      string            `json:"name"`
	Labels    map[string]string `json:"labels"`
}
