package validate

import (
	"github.com/sirupsen/logrus"
	"gitlab.bit9.local/octarine/cbctl/pkg/cberr"
)

// CheckValidBuildStep checks that a build step has a valid name.
func CheckValidBuildStep(buildStep string) error {
	if buildStep == "" {
		errMsg := "invalid build step name (empty)"
		err := cberr.NewError(cberr.ValidateFailedErr, errMsg, nil)
		logrus.Error(err)

		return err
	}

	return nil
}
