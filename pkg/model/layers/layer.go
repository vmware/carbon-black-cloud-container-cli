package layers

type Layer struct {
	Digest  string           `json:"digest"`
	Command string           `json:"command"`
	Size    uint64           `json:"size"`
	Index   int              `json:"index"`
	Files   []ExecutableFile `json:"files"`
}
