package utils_test

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/ubi-nodejs-extension/constants"
	testhelpers "github.com/paketo-buildpacks/ubi-nodejs-extension/internal/testhelpers"
	"github.com/paketo-buildpacks/ubi-nodejs-extension/internal/utils"
	"github.com/paketo-buildpacks/ubi-nodejs-extension/structs"
	"github.com/sclevine/spec"
)

func testGenerateConfigTomlContentFromImagesJson(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect = NewWithT(t).Expect
	)

	var imagesJsonDir string
	it.Before(func() {
		imagesJsonDir = t.TempDir()
	})

	it.After(func() {
		Expect(os.RemoveAll(imagesJsonDir)).To(Succeed())
	})
	context("When GenerateConfigTomlContentFromImagesJson is being called with a valid images.json file ", func() {

		it("successfully parses images.json file and returns the config.toml content", func() {

			imagesJsonContent := testhelpers.GenerateImagesJsonFile([]string{"16", "18", "20"}, []bool{false, false, true}, false, "9")
			imagesJsonTmpDir := t.TempDir()
			imagesJsonPath := filepath.Join(imagesJsonTmpDir, "images.json")
			Expect(os.WriteFile(imagesJsonPath, []byte(imagesJsonContent), 0644)).To(Succeed())

			configTomlContent, err := utils.GenerateConfigTomlContentFromImagesJson(imagesJsonPath, "io.buildpacks.stacks.ubi9")

			Expect(err).ToNot(HaveOccurred())
			Expect(string(configTomlContent)).To(ContainSubstring(`[metadata]
  [metadata.default-versions]
    node = "20.*.*"

  [[metadata.dependencies]]
    id = "node"
    source = "paketobuildpacks/run-nodejs-16-ubi9-base"
    stacks = ["io.buildpacks.stacks.ubi9"]
    version = "16.1000"

  [[metadata.dependencies]]
    id = "node"
    source = "paketobuildpacks/run-nodejs-18-ubi9-base"
    stacks = ["io.buildpacks.stacks.ubi9"]
    version = "18.1000"

  [[metadata.dependencies]]
    id = "node"
    source = "paketobuildpacks/run-nodejs-20-ubi9-base"
    stacks = ["io.buildpacks.stacks.ubi9"]
    version = "20.1000"`))
		})
	})

	context("When GenerateConfigTomlContentFromImagesJson is being called with an invalide images.json file ", func() {

		it("It should throw an error with a message", func() {

			_, err := utils.GenerateConfigTomlContentFromImagesJson("/path/to/invalid/images.json", "io.buildpacks.stacks.ubix")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no such file or directory"))
		})
	})

}

func testGetDefaultNodeVersion(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect = NewWithT(t).Expect
	)

	context("When passing an array of stacks with nodejs images", func() {

		context("and there is a default run image", func() {
			it("should find the default node version", func() {
				defaultNodeVersion, err := utils.GetDefaultNodeVersion([]utils.StackImages{
					{
						Name:              "nodejs-22",
						IsDefaultRunImage: true,
						NodeVersion:       "22",
					},
					{
						Name:              "nodejs-20",
						IsDefaultRunImage: false,
						NodeVersion:       "20",
					},
				})

				Expect(err).ToNot(HaveOccurred())
				Expect(defaultNodeVersion).To(Equal("22"))
			})
		})

		context("and there are no default run images", func() {
			it("should error with a message", func() {
				defaultNodeVersion, err := utils.GetDefaultNodeVersion([]utils.StackImages{
					{
						Name:              "nodejs-22",
						IsDefaultRunImage: false,
						NodeVersion:       "22",
					},
					{
						Name:              "nodejs-20",
						IsDefaultRunImage: false,
						NodeVersion:       "20",
					},
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("default node.js version not found"))
				Expect(defaultNodeVersion).To(Equal(""))
			})
		})

		context("and there are more than one default run images", func() {
			it("should error", func() {
				defaultNodeVersion, err := utils.GetDefaultNodeVersion([]utils.StackImages{
					{
						Name:              "nodejs-18",
						IsDefaultRunImage: true,
						NodeVersion:       "18",
					},
					{
						Name:              "nodejs-22",
						IsDefaultRunImage: false,
						NodeVersion:       "22",
					},
					{
						Name:              "nodejs-20",
						IsDefaultRunImage: true,
						NodeVersion:       "20",
					},
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("multiple default node.js versions found"))
				Expect(defaultNodeVersion).To(Equal(""))
			})
		})
	})
}

func testCreateConfigTomlFileContent(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect = NewWithT(t).Expect
	)

	context("When passing data properly to CreateConfigTomlFileContent function", func() {
		it("successfly create the content of config.toml", func() {
			configTomlFileContent, err := utils.CreateConfigTomlFileContent("22", []utils.StackImages{
				{
					Name:              "nodejs-22",
					IsDefaultRunImage: true,
					NodeVersion:       "22",
				},
				{
					Name:              "nodejs-20",
					IsDefaultRunImage: false,
					NodeVersion:       "24",
				},
			}, "io.buildpacks.stacks.ubi10")

			Expect(err).ToNot(HaveOccurred())
			Expect(configTomlFileContent.String()).To(ContainSubstring(`[metadata]
  [metadata.default-versions]
    node = "22.*.*"

  [[metadata.dependencies]]
    id = "node"
    source = "paketobuildpacks/ubi-10-run-nodejs-22-base"
    stacks = ["io.buildpacks.stacks.ubi10"]
    version = "22.1000"

  [[metadata.dependencies]]
    id = "node"
    source = "paketobuildpacks/ubi-10-run-nodejs-24-base"
    stacks = ["io.buildpacks.stacks.ubi10"]
    version = "24.1000"`))
		})
	})
}

func testParseImagesJsonFile(t *testing.T, _ spec.G, it spec.S) {

	var (
		Expect = NewWithT(t).Expect
	)

	var imagesJsonDir string
	it.Before(func() {
		imagesJsonDir = t.TempDir()
	})

	it.After(func() {
		Expect(os.RemoveAll(imagesJsonDir)).To(Succeed())
	})

	it("successfully parses images.json file", func() {

		imagesJsonContent := testhelpers.GenerateImagesJsonFile([]string{"16", "18", "20"}, []bool{false, false, true}, false, "8")
		imagesJsonTmpDir := t.TempDir()
		imagesJsonPath := filepath.Join(imagesJsonTmpDir, "images.json")
		Expect(os.WriteFile(imagesJsonPath, []byte(imagesJsonContent), 0644)).To(Succeed())

		imagesJsonData, err := utils.ParseImagesJsonFile(imagesJsonPath)
		Expect(err).ToNot(HaveOccurred())

		Expect(imagesJsonData).To(Equal(utils.ImagesJson{
			StackImages: []utils.StackImages{
				{
					Name:              "default",
					IsDefaultRunImage: false,
				},
				{
					Name:              "runtime-1",
					IsDefaultRunImage: false,
				},
				{
					Name:              "runtime-2",
					IsDefaultRunImage: false,
				},
				{
					Name:              "nodejs-16",
					IsDefaultRunImage: false,
				},
				{
					Name:              "nodejs-18",
					IsDefaultRunImage: false,
				},
				{
					Name:              "nodejs-20",
					IsDefaultRunImage: true,
				},
			},
		}))
	})

	it("erros when images.json file does not exist", func() {
		imagesJsonData, err := utils.ParseImagesJsonFile("/does/not/exist")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("no such file or directory"))
		Expect(imagesJsonData).To(Equal(utils.ImagesJson{}))
	})

	it("erros when images.json file is not a valid json", func() {

		imagesJsonContent := testhelpers.GenerateImagesJsonFile([]string{"16", "18", "20"}, []bool{false, false, true}, true, "8")
		imagesJsonTmpDir := t.TempDir()
		imagesJsonPath := filepath.Join(imagesJsonTmpDir, "images_not_valid.json")
		Expect(os.WriteFile(imagesJsonPath, []byte(imagesJsonContent), 0644)).To(Succeed())

		imagesJsonData, err := utils.ParseImagesJsonFile(imagesJsonPath)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("invalid character"))
		Expect(imagesJsonData).To(Equal(utils.ImagesJson{}))
	})
}

func testGetNodejsStackImages(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect = NewWithT(t).Expect
	)

	context("When passing the array with all the stacks", func() {

		it("should return only the nodejs stacks", func() {
			nodejsStacks, err := utils.GetNodejsStackImages(utils.ImagesJson{
				StackImages: []utils.StackImages{
					{
						Name:              "default",
						IsDefaultRunImage: false,
					},
					{
						Name:              "java-17",
						IsDefaultRunImage: false,
					},
					{
						Name:              "java-21",
						IsDefaultRunImage: false,
					},
					{
						Name:              "nodejs-16",
						IsDefaultRunImage: false,
					},
					{
						Name:              "nodejs-18",
						IsDefaultRunImage: false,
					},
					{
						Name:              "nodejs-20",
						IsDefaultRunImage: true,
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(nodejsStacks).To(Equal([]utils.StackImages{
				{
					Name:              "nodejs-16",
					IsDefaultRunImage: false,
					NodeVersion:       "16",
				},
				{
					Name:              "nodejs-18",
					IsDefaultRunImage: false,
					NodeVersion:       "18",
				},
				{
					Name:              "nodejs-20",
					IsDefaultRunImage: true,
					NodeVersion:       "20",
				},
			}))
		})
	})

	context("When passing a stack images array without any nodejs stacks in it", func() {

		it("should return an error with an appropriate message", func() {
			nodejsStacks, err := utils.GetNodejsStackImages(utils.ImagesJson{
				StackImages: []utils.StackImages{
					{
						Name:              "default",
						IsDefaultRunImage: false,
					},
					{
						Name:              "java-17",
						IsDefaultRunImage: false,
					},
					{
						Name:              "java-21",
						IsDefaultRunImage: false,
					},
				},
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("no nodejs stacks found"))
			Expect(nodejsStacks).To(Equal([]utils.StackImages{}))
		})
	})

	context("When node version is malformed or does not exist", func() {

		it("should error with a message", func() {

			imagesJsonTmpDir := t.TempDir()
			imagesJsonNodeVersionNotIntegerContent := testhelpers.GenerateImagesJsonFile([]string{"16", "18", "hello"}, []bool{false, false, true}, false, "8")
			imagesJsonNodeVersionNotIntegerPath := filepath.Join(imagesJsonTmpDir, "images_node_version_not_integer.json")
			Expect(os.WriteFile(imagesJsonNodeVersionNotIntegerPath, []byte(imagesJsonNodeVersionNotIntegerContent), 0600)).To(Succeed())

			imagesJsonNoNodeVersionContent := testhelpers.GenerateImagesJsonFile([]string{"16", "", "20"}, []bool{false, false, true}, false, "8")
			imagesJsonNoNodeVersionPath := filepath.Join(imagesJsonTmpDir, "images_no_node_version.json")
			Expect(os.WriteFile(imagesJsonNoNodeVersionPath, []byte(imagesJsonNoNodeVersionContent), 0600)).To(Succeed())

			for _, tt := range []struct {
				errorMessage   string
				imagesJsonPath string
			}{
				{
					errorMessage:   "extracted Node.js version [hello] for stack nodejs-hello is not an integer",
					imagesJsonPath: imagesJsonNodeVersionNotIntegerPath,
				},
				{
					errorMessage:   "extracted Node.js version [] for stack nodejs- is not an integer",
					imagesJsonPath: imagesJsonNoNodeVersionPath,
				},
			} {
				imagesJsonData, err := utils.ParseImagesJsonFile(filepath.Join(tt.imagesJsonPath))
				Expect(err).ToNot(HaveOccurred())

				nodejsStacks, err := utils.GetNodejsStackImages(imagesJsonData)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(tt.errorMessage))
				Expect(nodejsStacks).To(Equal([]utils.StackImages{}))
			}
		})
	})
}

func testGenerateBuildDockerfile(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect = NewWithT(t).Expect
	)

	context("Adding props on build.dockerfile template", func() {

		it("Should fill with properties the template/build.Dockerfile", func() {

			getInstalledPackages, err := utils.GetBuildPackages("io.buildpacks.stacks.ubi8", 16)
			Expect(err).NotTo(HaveOccurred())

			output, err := utils.GenerateBuildDockerfile(structs.BuildDockerfileProps{
				NODEJS_VERSION:       16,
				CNB_USER_ID:          1000,
				CNB_GROUP_ID:         1000,
				CNB_STACK_ID:         "io.buildpacks.stacks.ubi8",
				PACKAGES:             getInstalledPackages,
				ENABLE_NODEJS_MODULE: utils.ShouldEnableNodejsModule("io.buildpacks.stacks.ubi8"),
			})

			Expect(err).NotTo(HaveOccurred())
			expectedOutput := `ARG base_image
FROM ${base_image}

USER root

ARG build_id=0
RUN echo ${build_id}

RUN microdnf -y module enable nodejs:16 && \ microdnf --setopt=install_weak_deps=0 --setopt=tsflags=nodocs \
    install -y make gcc gcc-c++ libatomic_ops git openssl-devel nodejs npm nodejs-nodemon nss_wrapper which python3 && \
    microdnf clean all

RUN echo uid:gid "1000:1000"
USER 1000:1000

RUN echo "CNB_STACK_ID: io.buildpacks.stacks.ubi8"`
			Expect(output).To(Equal(expectedOutput))

		})
	})
}

func testGenerateRunDockerfile(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect = NewWithT(t).Expect
	)

	context("Adding props on build.dockerfile template", func() {

		it("Should fill with properties the template/run.Dockerfile", func() {

			RunDockerfileProps := structs.RunDockerfileProps{
				Source: "paketobuildpacks/run-nodejs-18-ubi8-base",
			}

			output, err := utils.GenerateRunDockerfile(RunDockerfileProps)

			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(`FROM paketobuildpacks/run-nodejs-18-ubi8-base`))

		})
	})
}

func testGetDuringBuildPermissions(t *testing.T, context spec.G, it spec.S) {

	var Expect = NewWithT(t).Expect

	context("/etc/passwd exists and has the cnb user", func() {

		it("It should return the permissions specified for the cnb user", func() {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "/passwd")
			Expect(os.WriteFile(path, []byte(`root:x:0:0:root:/root:/bin/bash
bin:x:1:1:bin:/bin:/sbin/nologin
daemon:x:2:2:daemon:/sbin:/sbin/nologin
adm:x:3:4:adm:/var/adm:/sbin/nologin
lp:x:4:7:lp:/var/spool/lpd:/sbin/nologin
sync:x:5:0:sync:/sbin:/bin/sync
shutdown:x:6:0:shutdown:/sbin:/sbin/shutdown
halt:x:7:0:halt:/sbin:/sbin/halt
mail:x:8:12:mail:/var/spool/mail:/sbin/nologin
operator:x:11:0:operator:/root:/sbin/nologin
games:x:12:100:games:/usr/games:/sbin/nologin
ftp:x:14:50:FTP User:/var/ftp:/sbin/nologin
cnb:x:1234:2345::/home/cnb:/bin/bash
nobody:x:65534:65534:Kernel Overflow User:/:/sbin/nologin
`), 0600)).To(Succeed())

			duringBuilderPermissions := utils.GetDuringBuildPermissions(path)

			Expect(duringBuilderPermissions).To(Equal(
				structs.DuringBuildPermissions{
					CNB_USER_ID:  1234,
					CNB_GROUP_ID: 2345,
				},
			))
		})
	})

	context("/etc/passwd exists and does NOT have the cnb user", func() {

		it("It should return the default permissions", func() {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "/passwd")

			Expect(os.WriteFile(path, []byte(`root:x:0:0:root:/root:/bin/bash
bin:x:1:1:bin:/bin:/sbin/nologin
daemon:x:2:2:daemon:/sbin:/sbin/nologin
adm:x:3:4:adm:/var/adm:/sbin/nologin
lp:x:4:7:lp:/var/spool/lpd:/sbin/nologin
sync:x:5:0:sync:/sbin:/bin/sync
shutdown:x:6:0:shutdown:/sbin:/sbin/shutdown
halt:x:7:0:halt:/sbin:/sbin/halt
mail:x:8:12:mail:/var/spool/mail:/sbin/nologin
operator:x:11:0:operator:/root:/sbin/nologin
games:x:12:100:games:/usr/games:/sbin/nologin
ftp:x:14:50:FTP User:/var/ftp:/sbin/nologin
nobody:x:65534:65534:Kernel Overflow User:/:/sbin/nologin
`), 0600)).To(Succeed())

			duringBuildPermissions := utils.GetDuringBuildPermissions(path)

			Expect(duringBuildPermissions).To(Equal(
				structs.DuringBuildPermissions{
					CNB_USER_ID:  constants.DEFAULT_USER_ID,
					CNB_GROUP_ID: constants.DEFAULT_GROUP_ID},
			))
		})
	})

	context("/etc/passwd does NOT exist", func() {
		it("It should return the default permissions", func() {
			tmpDir := t.TempDir()
			duringBuilderPermissions := utils.GetDuringBuildPermissions(tmpDir)

			Expect(duringBuilderPermissions).To(Equal(
				structs.DuringBuildPermissions{
					CNB_USER_ID:  constants.DEFAULT_USER_ID,
					CNB_GROUP_ID: constants.DEFAULT_GROUP_ID},
			))
		})
	})
}

func testGetBuildPackages(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect = NewWithT(t).Expect
	)

	context("Success cases", func() {
		it("should return the correct build packages for all supported combinations", func() {
			testCases := []struct {
				stackId          string
				nodeVersion      int
				expectedPackages string
				description      string
			}{
				// UBI8
				{
					stackId:          "io.buildpacks.stacks.ubi8",
					nodeVersion:      16,
					expectedPackages: "make gcc gcc-c++ libatomic_ops git openssl-devel nodejs npm nodejs-nodemon nss_wrapper which python3",
					description:      "UBI8 with Node.js 16",
				},
				{
					stackId:          "io.buildpacks.stacks.ubi8",
					nodeVersion:      18,
					expectedPackages: "make gcc gcc-c++ libatomic_ops git openssl-devel nodejs npm nodejs-nodemon nss_wrapper which python3",
					description:      "UBI8 with Node.js 18",
				},
				{
					stackId:          "io.buildpacks.stacks.ubi8",
					nodeVersion:      20,
					expectedPackages: "make gcc gcc-c++ libatomic_ops git openssl-devel nodejs npm nodejs-nodemon nss_wrapper which python3",
					description:      "UBI8 with Node.js 20",
				},
				{
					stackId:          "io.buildpacks.stacks.ubi8",
					nodeVersion:      22,
					expectedPackages: "make gcc-toolset-13-gcc gcc-toolset-13-gcc-c++ gcc-toolset-13-runtime libatomic_ops git openssl-devel python3.12 nodejs npm nodejs-nodemon nss_wrapper-libs which",
					description:      "UBI8 with Node.js 22 (uses GCC toolset 13)",
				},
			}

			for _, tt := range testCases {
				packages, err := utils.GetBuildPackages(tt.stackId, tt.nodeVersion)
				Expect(err).NotTo(HaveOccurred(), "Failed for: %s", tt.description)
				Expect(packages).To(Equal(tt.expectedPackages), "Package mismatch for: %s", tt.description)
			}
		})
	})

	context("Error cases", func() {
		it("should return errors for unsupported Node.js versions", func() {
			testCases := []struct {
				stackId       string
				nodeVersion   int
				expectedError string
				description   string
			}{
				{
					stackId:       "io.buildpacks.stacks.ubi8",
					nodeVersion:   14,
					expectedError: "unsupported Node.js version 14 for image io.buildpacks.stacks.ubi8",
					description:   "UBI8 with unsupported Node.js 14",
				},
				{
					stackId:       "io.buildpacks.stacks.ubi9",
					nodeVersion:   16,
					expectedError: "unsupported Node.js version 16 for image io.buildpacks.stacks.ubi9",
					description:   "UBI9 with unsupported Node.js 16",
				},
			}

			for _, tt := range testCases {
				packages, err := utils.GetBuildPackages(tt.stackId, tt.nodeVersion)
				Expect(err).To(HaveOccurred(), "Expected error for: %s", tt.description)
				Expect(err.Error()).To(Equal(tt.expectedError), "Error message mismatch for: %s", tt.description)
				Expect(packages).To(BeEmpty(), "Expected empty packages for: %s", tt.description)
			}
		})

		it("should return errors for unsupported stack IDs", func() {
			testCases := []struct {
				stackId       string
				nodeVersion   int
				expectedError string
				description   string
			}{
				{
					stackId:       "io.buildpacks.stacks.ubi7",
					nodeVersion:   20,
					expectedError: "unsupported image ID: io.buildpacks.stacks.ubi7",
					description:   "Unsupported UBI7 stack",
				},
				{
					stackId:       "invalid.stack.id",
					nodeVersion:   18,
					expectedError: "unsupported image ID: invalid.stack.id",
					description:   "Invalid stack ID",
				},
				{
					stackId:       "",
					nodeVersion:   20,
					expectedError: "unsupported image ID: ",
					description:   "Empty stack ID",
				},
			}

			for _, tt := range testCases {
				packages, err := utils.GetBuildPackages(tt.stackId, tt.nodeVersion)
				Expect(err).To(HaveOccurred(), "Expected error for: %s", tt.description)
				Expect(err.Error()).To(Equal(tt.expectedError), "Error message mismatch for: %s", tt.description)
				Expect(packages).To(BeEmpty(), "Expected empty packages for: %s", tt.description)
			}
		})
	})

}

func testGetOsCodenameFromStackId(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect = NewWithT(t).Expect
	)

	context("When passing a valid stack ID", func() {

		it("should return the correct OS codename", func() {

			testCases := []struct {
				stackId          string
				expectedCodename string
				description      string
			}{
				{
					stackId:          "io.buildpacks.stacks.ubi8",
					expectedCodename: "ubi8",
				},
				{
					stackId:          "io.buildpacks.stacks.ubi9",
					expectedCodename: "ubi9",
				},
			}

			for _, tc := range testCases {
				codename, err := utils.GetOsCodenameFromStackId(tc.stackId)
				Expect(err).NotTo(HaveOccurred())
				Expect(codename).To(Equal(tc.expectedCodename))
			}
		})
	})

	context("When os codename can not be extracted from the stack id", func() {

		it("should return an error", func() {
			testCases := []struct {
				stackId       string
				expectedError string
			}{
				{
					stackId:       "io.buildpacks.stacks.",
					expectedError: "failed to extract os codename from stack id 'io.buildpacks.stacks.': os codename length cannot be zero",
				},
				{
					stackId:       "invalid.stackid.ubi8",
					expectedError: "failed to extract os codename from stack id 'invalid.stackid.ubi8'. stack id is missing the required prefix 'io.buildpacks.stacks.'",
				},
				{
					stackId:       "",
					expectedError: "failed to extract os codename from stack id ''. stack id is missing the required prefix 'io.buildpacks.stacks.'",
				},
				{
					stackId:       "lskdjflksdfj s sdlkf sfd;alf ",
					expectedError: "failed to extract os codename from stack id 'lskdjflksdfj s sdlkf sfd;alf '. stack id is missing the required prefix 'io.buildpacks.stacks.'",
				},
			}

			for _, tt := range testCases {
				_, err := utils.GetOsCodenameFromStackId(tt.stackId)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(tt.expectedError))
			}
		})
	})
}
