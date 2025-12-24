package integration

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testVendored(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack().WithNoColor()
		docker = occam.NewDocker()
	})

	context("when the node_modules are vendored", func() {
		var (
			image     occam.Image
			container occam.Container

			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			source, err = occam.Source(filepath.Join("testdata", "vendored"))
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("builds a working OCI image for a simple app", func() {
			var (
				err  error
				logs fmt.Stringer
			)

			image, logs, err = pack.Build.
				WithExtensions(
					settings.Extensions.UbiNodejsExtension.Online,
				).
				WithBuildpacks(
					settings.Buildpacks.NodeEngine.Online,
					settings.Buildpacks.NPMInstall.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				WithPullPolicy("always").
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			container, err = docker.Container.Run.
				WithCommand("npm start").
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(container).Should(BeAvailable())

			response, err := http.Get(fmt.Sprintf("http://localhost:%s", container.HostPort("8080")))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			content, err := io.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("Hello, World!"))

			fmt.Println(logs)
			Expect(logs).To(ContainLines(
				"[extender (build)]   Resolving installation process",
				"[extender (build)]     Process inputs:",
				"[extender (build)]       node_modules      -> \"Found\"",
				"[extender (build)]       npm-cache         -> \"Not found\"",
				"[extender (build)]       package-lock.json -> \"Found\"",
				"[extender (build)] ",
				"[extender (build)]     Selected NPM build process: 'npm rebuild'"))
			Expect(logs).To(ContainLines("[extender (build)]   Executing launch environment install process"))
			Expect(logs).To(ContainLines("[extender (build)]     Running 'npm run-script preinstall --if-present'"))
			Expect(logs).To(ContainLines(MatchRegexp(`\[extender \(build\)\]     Running 'npm rebuild --nodedir='`)))
			Expect(logs).To(ContainLines("[extender (build)]     Running 'npm run-script postinstall --if-present'"))
			Expect(logs).To(ContainLines(
				"[extender (build)]   Configuring launch environment",
				"[extender (build)]     NODE_PROJECT_PATH   -> \"/workspace\"",
				"[extender (build)]     NPM_CONFIG_LOGLEVEL -> \"error\"",
			))
		})
	})
}
