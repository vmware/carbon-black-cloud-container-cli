package integration

import (
	"fmt"
	"os/exec"
	"time"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Integration Tests", func() {

	Context("Given a docker config json with matching registry credentials", func() {
		When("a registry hostname is provided", func() {
			It("can return the username", func() {

				session := cli(10, 0, "username", "dev.registry.tanzu.vmware.com")

				Expect(string(session.Out.Contents())).To(Equal("foo"))
			})
			It("can return the password", func() {

				session := cli(10, 0, "password", "dev.registry.tanzu.vmware.com")

				Expect(string(session.Out.Contents())).To(Equal("bar"))
			})
		})

		When("a image repository is provided", func() {
			It("can return the username", func() {

				session := cli(10, 0, "username", "dev.registry.tanzu.vmware.com/bizz/bazz")

				Expect(string(session.Out.Contents())).To(Equal("foo"))
			})
			It("can return the password", func() {

				session := cli(10, 0, "password", "dev.registry.tanzu.vmware.com/bizz/bazz")

				Expect(string(session.Out.Contents())).To(Equal("bar"))
			})
		})

		When("an invalid field is requested", func() {
			It("returns an error", func() {

				session := cli(10, 1, "other", "dev.registry.tanzu.vmware.com/bizz/bazz")

				Expect(string(session.Err.Contents())).Should(ContainSubstring("field must be 'username' or 'password'"))
			})
		})

		When("an invalid url is provided", func() {
			It("returns an error", func() {

				session := cli(10, 1, "username", ":\\")

				Expect(string(session.Err.Contents())).Should(ContainSubstring("could not parse reference"))
			})
		})
	})

	Context("Given a docker config json without matching registry credentials", func() {
		When("a registry hostname is provided", func() {
			It("returns empty username", func() {

				session := cli(10, 0, "username", "notaregistry.com")

				Expect(string(session.Out.Contents())).To(Equal(""))
			})
			It("returns empty password", func() {

				session := cli(10, 0, "password", "notaregistry.com")

				Expect(string(session.Out.Contents())).To(Equal(""))
			})
		})
	})

	Context("Given no docker config json", func() {
		BeforeEach(func() {
			os.Setenv("DOCKER_CONFIG", ".")
		})
		When("a registry hostname is provided", func() {
			It("returns empty username", func() {

				session := cli(10, 0, "username", "notaregistry.com")

				Expect(string(session.Out.Contents())).To(Equal(""))
			})
			It("returns empty password", func() {

				session := cli(10, 0, "password", "notaregistry.com")

				Expect(string(session.Out.Contents())).To(Equal(""))
			})
		})
	})
})

func cli(timeout time.Duration, expectedExitCode int, args ...string) *gexec.Session {
	session, err := run("bin/auth", args...)
	Expect(err).ToNot(HaveOccurred())
	Eventually(session, timeout*time.Second).Should(gexec.Exit(expectedExitCode))
	return session
}

func run(command string, args ...string) (*gexec.Session, error) {
	fmt.Fprintf(GinkgoWriter, "run %s %v\n", command, args)
	com := exec.Command(command, args...)
	return gexec.Start(com, GinkgoWriter, GinkgoWriter)
}
