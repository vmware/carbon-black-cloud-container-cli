package memorytool

import (
	"context"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

// ReadMemoryStats will print mem stat info in debug log.
func ReadMemoryStats(ctx context.Context) {
	var (
		maxRSSMB, totalRSSMB float64
		iterationsCount      int
		startedAt            = time.Now()
	)

	ticker := time.NewTicker(500 * time.Millisecond) // nolint: gomnd
	defer ticker.Stop()

	track := func() {
		var ms runtime.MemStats

		runtime.ReadMemStats(&ms)

		rssMB := float64(ms.Sys) / (1024 * 1024)
		logrus.Debugf("Current memory usage: %.1fMB", rssMB)

		if rssMB > maxRSSMB {
			maxRSSMB = rssMB
		}

		totalRSSMB += rssMB
		iterationsCount++
	}

trackLoop:
	for {
		select {
		case <-ctx.Done():
			logrus.Debug("Stopped resources tracking")
			break trackLoop
		case <-ticker.C:
			track()
		}
	}

	logrus.Infof("Summary: %d samples; Memory: avg is %.1fMB, max is %.1fMB; Time took: %s",
		iterationsCount, totalRSSMB/float64(iterationsCount), maxRSSMB, time.Since(startedAt))
}
