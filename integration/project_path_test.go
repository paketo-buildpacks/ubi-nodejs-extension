package integration

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testProjectPath(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when built with custom project path set", func() {
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
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("builds and runs correctly", func() {
			var err error

			source, err = occam.Source(filepath.Join("testdata", "custom_project_path_app"))
			Expect(err).ToNot(HaveOccurred())

			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithExtensions(
					settings.Extensions.UbiNodejsExtension.Online,
				).
				WithBuildpacks(
					settings.Buildpacks.NodeEngine.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				WithEnv(map[string]string{
					"BP_NODE_PROJECT_PATH": "hello_world_server",
				}).
				WithPullPolicy("always").
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			versionPatternCandidate := regexp.MustCompile(`\.node-version -> "\d+\.\*"`)
			versionPatternSelectedNode := regexp.MustCompile(`Selected Node Engine Major version \d+`)
			versionPatternBuildNode := regexp.MustCompile(`nodejs:\d+`)
			logsReplaced := versionPatternCandidate.ReplaceAllString(logs.String(), `.node-version -> "x.x"`)
			logsReplaced = versionPatternSelectedNode.ReplaceAllString(logsReplaced, `Selected Node Engine Major version x`)
			logsReplaced = versionPatternBuildNode.ReplaceAllString(logsReplaced, `nodejs:x`)

			Expect(logsReplaced).To(ContainLines(
				fmt.Sprintf("%s 1.2.3", settings.Extension.Name),
				"  Resolving Node Engine version",
				"    Candidate version sources (in priority order):",
				"      .node-version -> \"x.x\"",
				"      <unknown>     -> \"\""))

			Expect(logsReplaced).To(ContainLines(
				"  Selected Node Engine Major version x"))
			Expect(logsReplaced).To(ContainLines("===> RESTORING"))
			Expect(logsReplaced).To(ContainLines("===> EXTENDING (BUILD)"))

			Expect(logsReplaced).To(ContainLines("[extender (build)] Enabling module streams:",
				"[extender (build)]     nodejs:x"))

			container, err = docker.Container.Run.
				WithCommand("node hello_world_server/server.js").
				WithPublish("8080").
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(container).Should(BeAvailable())

			response, err := http.Get(fmt.Sprintf("http://localhost:%s", container.HostPort("8080")))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			content, err := io.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("hello world"))
		})
	})
}
