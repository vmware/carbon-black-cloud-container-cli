package integration

import (
	"fmt"
	"os/exec"
	"testing"
	"time"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.
func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Integration Suite")
}

var _ = BeforeSuite(func() {
	os.Setenv("DOCKER_CONFIG", "./assets")
	session, err := Run("go", "build", "-o", "bin/auth", "../")
	Expect(err).ToNot(HaveOccurred())
	Eventually(session, 15*time.Second).Should(gexec.Exit(0))
})

func Run(command string, args ...string) (*gexec.Session, error) {
	fmt.Fprintf(GinkgoWriter, "run %s %v\n", command, args)
	com := exec.Command(command, args...)
	return gexec.Start(com, GinkgoWriter, GinkgoWriter)
}
