package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testOpenSSL(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
	)

	var testCases = []struct {
		nodeVersion string
		expected    string
	}{
		{"20.*.*", "v20."},
		{"22.*.*", "v22."},
	}

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when the buildpack is run with pack build", func() {
		var (
			image     occam.Image
			container occam.Container
			name      string
			source    string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			source, err = occam.Source(filepath.Join("testdata", "simple_app"))
			Expect(err).ToNot(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		for _, tc := range testCases {
			nodeVersion := tc.nodeVersion
			context(fmt.Sprintf("when running Node %s", nodeVersion), func() {

				it("uses the OpenSSL CA store to verify certificates", func() {
					var (
						logs fmt.Stringer
						err  error
					)

					image, logs, err = pack.WithNoColor().Build.
						WithExtensions(
							settings.Buildpacks.NodeExtension.Online,
						).
						WithBuildpacks(
							settings.Buildpacks.NodeEngine.Online,
							settings.Buildpacks.BuildPlan.Online,
						).
						WithEnv(map[string]string{
							"BP_NODE_VERSION": nodeVersion,
						}).
						WithPullPolicy("always").
						Execute(name, source)
					Expect(err).NotTo(HaveOccurred(), logs.String)

					container, err = docker.Container.Run.
						WithPublish("8080").
						WithCommand("node server.js").
						Execute(image.ID)
					Expect(err).NotTo(HaveOccurred())

					Eventually(container).Should(Serve("hello world"))
					Expect(container).To(Serve(ContainSubstring(tc.expected)).WithEndpoint("/version"))
					Expect(container).To(Serve(ContainSubstring("301 Moved")).WithEndpoint("/test-openssl-ca"))

					Expect(logs).To(ContainLines(
						`[extender (build)]   Configuring launch environment`,
						`[extender (build)]     NODE_ENV     -> "production"`,
						`[extender (build)]     NODE_HOME    -> ""`,
						`[extender (build)]     NODE_OPTIONS -> "--use-openssl-ca"`,
						`[extender (build)]     NODE_VERBOSE -> "false"`,
					))
				})
			})

		}
	})
}
