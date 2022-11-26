package cberr

// Code is the type of machine-readable error code.
type Code int

// ErrCode types.
const (
	UnclassifiedErr Code = iota
	ConfigErr
	HTTPConnectionErr
	HTTPUnsuccessfulResponseErr
	HTTPNotFoundErr
	HTTPNotAllowedErr
	ImageLoadErr
	SBOMGenerationErr
	LayersGenerationErr
	ScanFailedErr
	ValidateFailedErr
	TimeoutErr
	DisplayErr
	PolicyViolationErr
	EmptyResponse
)

//nolint:gomnd
func (c Code) exitCode() int {
	switch c {
	case UnclassifiedErr:
		return 0
	case ConfigErr:
		return 0
	case HTTPConnectionErr:
		return 1
	case HTTPUnsuccessfulResponseErr:
		return 1
	case HTTPNotFoundErr:
		return 1
	case HTTPNotAllowedErr:
		return 1
	case ImageLoadErr:
		return 1
	case SBOMGenerationErr:
		return 1
	case LayersGenerationErr:
		return 1
	case ScanFailedErr:
		return 1
	case ValidateFailedErr:
		return 1
	case TimeoutErr:
		return 1
	case PolicyViolationErr:
		return 127
	case DisplayErr:
		return 1
	case EmptyResponse:
		return 1
	default:
		return 0
	}
}
