package utils

import (
	_ "embed"

	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/paketo-buildpacks/ubi-nodejs-extension/constants"
	"github.com/paketo-buildpacks/ubi-nodejs-extension/structs"

	"github.com/BurntSushi/toml"
)

//go:embed templates/build.Dockerfile
var buildDockerfileTemplate string

//go:embed templates/run.Dockerfile
var runDockerfileTemplate string

type StackImages struct {
	Name              string `json:"name"`
	IsDefaultRunImage bool   `json:"is_default_run_image,omitempty"`
	NodeVersion       string
}

type ImagesJson struct {
	StackImages []StackImages `json:"images"`
}

func GenerateConfigTomlContentFromImagesJson(imagesJsonPath string, stackId string) ([]byte, error) {
	imagesJsonData, err := ParseImagesJsonFile(imagesJsonPath)
	if err != nil {
		return []byte{}, err
	}

	nodejsStacks, err := GetNodejsStackImages(imagesJsonData)
	if err != nil {
		return []byte{}, err
	}

	defaultNodeVersion, err := GetDefaultNodeVersion(nodejsStacks)
	if err != nil {
		return []byte{}, err
	}

	osCodename, err := GetOsCodenameFromStackId(stackId)
	if err != nil {
		return []byte{}, err
	}

	configTomlContent, err := CreateConfigTomlFileContent(defaultNodeVersion, nodejsStacks, stackId, osCodename)
	if err != nil {
		return []byte{}, err
	}

	configTomlContentString := configTomlContent.Bytes()
	return configTomlContentString, nil
}

func GetDefaultNodeVersion(stacks []StackImages) (string, error) {
	var defaultNodeVersionsFound []string
	for _, stack := range stacks {
		if stack.IsDefaultRunImage {
			defaultNodeVersionsFound = append(defaultNodeVersionsFound, strings.Split(stack.Name, "-")[1])
		}
	}
	if len(defaultNodeVersionsFound) == 1 {
		return defaultNodeVersionsFound[0], nil
	} else if len(defaultNodeVersionsFound) > 1 {
		return "", errors.New("multiple default node.js versions found")
	} else {
		return "", errors.New("default node.js version not found")
	}
}

func CreateConfigTomlFileContent(defaultNodeVersion string, nodejsStacks []StackImages, stackId string, osCodename string) (bytes.Buffer, error) {

	var dependencies []map[string]interface{}

	for _, stack := range nodejsStacks {
		dependency := map[string]interface{}{
			"id":      "node",
			"stacks":  []string{stackId},
			"version": fmt.Sprintf("%s.1000", stack.NodeVersion),
			"source":  fmt.Sprintf("paketobuildpacks/run-nodejs-%s-%s-base", stack.NodeVersion, osCodename),
		}
		dependencies = append(dependencies, dependency)
	}

	config := map[string]interface{}{
		"metadata": map[string]interface{}{
			"default-versions": map[string]string{
				"node": fmt.Sprintf("%s.*.*", defaultNodeVersion),
			},
			"dependencies": dependencies,
		},
	}

	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(config); err != nil {
		return bytes.Buffer{}, err
	}

	return *buf, nil
}

func GetNodejsStackImages(imagesJsonData ImagesJson) ([]StackImages, error) {

	// Filter out the nodejs stacks based on the stack name
	nodejsRegex, _ := regexp.Compile("^nodejs")

	nodejsStacks := []StackImages{}
	for _, stack := range imagesJsonData.StackImages {

		if nodejsRegex.MatchString(stack.Name) {
			//Extract the node version from the stack name
			extractedNodeVersion := strings.Split(stack.Name, "-")[1]

			_, err := strconv.Atoi(extractedNodeVersion)
			if err != nil {
				return []StackImages{}, fmt.Errorf("extracted Node.js version [%s] for stack %s is not an integer", extractedNodeVersion, stack.Name)
			}

			stack.NodeVersion = extractedNodeVersion

			nodejsStacks = append(nodejsStacks, stack)
		}
	}
	if len(nodejsStacks) == 0 {
		return []StackImages{}, errors.New("no nodejs stacks found")
	}

	return nodejsStacks, nil
}

func ParseImagesJsonFile(imagesJsonPath string) (ImagesJson, error) {
	filepath, err := os.Open(imagesJsonPath)
	if err != nil {
		return ImagesJson{}, err
	}

	var imagesJsonData ImagesJson
	err = json.NewDecoder(filepath).Decode(&imagesJsonData)
	if err != nil {
		return ImagesJson{}, err
	}

	if err := filepath.Close(); err != nil {
		return ImagesJson{}, err
	}

	return imagesJsonData, nil
}

func GetDuringBuildPermissions(filepath string) structs.DuringBuildPermissions {

	defaultPermissions := structs.DuringBuildPermissions{
		CNB_USER_ID:  constants.DEFAULT_USER_ID,
		CNB_GROUP_ID: constants.DEFAULT_GROUP_ID,
	}
	re := regexp.MustCompile(`cnb:x:(\d+):(\d+)::`)

	etcPasswdFile, err := os.ReadFile(filepath)

	if err != nil {
		return defaultPermissions
	}
	etcPasswdContent := string(etcPasswdFile)

	matches := re.FindStringSubmatch(etcPasswdContent)

	if len(matches) != 3 {
		return defaultPermissions
	}

	CNB_USER_ID, err := strconv.Atoi(matches[1])

	if err != nil {
		return defaultPermissions
	}

	CNB_GROUP_ID, err := strconv.Atoi(matches[2])

	if err != nil {
		return defaultPermissions
	}

	return structs.DuringBuildPermissions{
		CNB_USER_ID:  CNB_USER_ID,
		CNB_GROUP_ID: CNB_GROUP_ID,
	}
}

func GenerateBuildDockerfile(buildProps structs.BuildDockerfileProps) (result string, Error error) {

	result, err := fillPropsToTemplate(buildProps, buildDockerfileTemplate)

	if err != nil {
		return "", err
	}

	return result, nil
}

func GenerateRunDockerfile(runProps structs.RunDockerfileProps) (result string, Error error) {

	result, err := fillPropsToTemplate(runProps, runDockerfileTemplate)

	if err != nil {
		return "", err
	}
	return result, nil
}

func fillPropsToTemplate(properties any, templateString string) (result string, Error error) {

	templ, err := template.New("template").Parse(templateString)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = templ.Execute(&buf, properties)
	if err != nil {
		panic(err)
	}

	return buf.String(), nil
}

func GetSymlinks(imageId string, nodeVersion int) string {
	if imageId == "io.buildpacks.stacks.ubi8" && (nodeVersion == 24 || nodeVersion == 22) {
		return `RUN ln -sf /opt/rh/gcc-toolset-13/root/usr/bin/gcc /usr/bin/gcc && \
    ln -sf /opt/rh/gcc-toolset-13/root/usr/bin/g++ /usr/bin/g++`
	} else if imageId == "io.buildpacks.stacks.ubi10" && (nodeVersion == 24) {
		return `RUN ln -s /usr/bin/node-24 /usr/bin/node && \
ln -s /usr/bin/npm-24 /usr/bin/npm && \
ln -s /usr/bin/npx-24 /usr/bin/npx`
	} else if imageId == "io.buildpacks.stacks.ubi10" && (nodeVersion == 22) {
		return `RUN rm /usr/bin/node && ln -s /usr/bin/node-22 /usr/bin/node && \
rm /usr/bin/npm && ln -s /usr/bin/npm-22 /usr/bin/npm && \
rm /usr/bin/npx && ln -s /usr/bin/npx-22 /usr/bin/npx`
	}
	return ""
}

func GetBuildPackages(imageId string, nodeVersion int) (string, error) {

	switch imageId {
	case "io.buildpacks.stacks.ubi8":
		switch nodeVersion {
		case 16, 18, 20:
			return "make gcc gcc-c++ libatomic_ops git openssl-devel nodejs npm nodejs-nodemon nss_wrapper which python3", nil
		case 22:
			return "make gcc-toolset-13-gcc gcc-toolset-13-gcc-c++ gcc-toolset-13-runtime libatomic_ops git openssl-devel python3.12 nodejs npm nodejs-nodemon nss_wrapper-libs which", nil
		case 24:
			return "make gcc-toolset-13-gcc gcc-toolset-13-gcc-c++ gcc-toolset-13-runtime libatomic_ops git openssl-devel python3.12 nodejs npm nodejs-nodemon nss_wrapper-libs which", nil
		default:
			return "", fmt.Errorf("unsupported Node.js version %d for image %s", nodeVersion, imageId)
		}
	case "io.buildpacks.stacks.ubi9":
		switch nodeVersion {
		case 18, 20, 22, 24:
			return "make gcc gcc-c++ git openssl-devel nodejs npm nodejs-nodemon nss_wrapper-libs python3", nil
		default:
			return "", fmt.Errorf("unsupported Node.js version %d for image %s", nodeVersion, imageId)
		}
	case "io.buildpacks.stacks.ubi10":
		switch nodeVersion {
		case 22:
			return "make gcc gcc-c++ git openssl-devel nodejs nodejs-nodemon nodejs-npm nss_wrapper-libs which", nil
		case 24:
			return "make gcc gcc-c++ git openssl-devel nodejs24 nodejs-nodemon nodejs24-npm nss_wrapper-libs which", nil
		default:
			return "", fmt.Errorf("unsupported Node.js version %d for image %s", nodeVersion, imageId)
		}
	}

	return "", fmt.Errorf("unsupported image ID: %s", imageId)
}

func GetOsCodenameFromStackId(stackId string) (string, error) {

	stackIdPrefix := "io.buildpacks.stacks."
	osCodename := strings.TrimPrefix(stackId, stackIdPrefix)

	if !strings.HasPrefix(stackId, stackIdPrefix) {
		return "", fmt.Errorf("failed to extract os codename from stack id '%s'. stack id is missing the required prefix '%s'", stackId, stackIdPrefix)
	}

	if len(osCodename) == 0 {
		return "", fmt.Errorf("failed to extract os codename from stack id '%s': os codename length cannot be zero", stackId)
	}

	return osCodename, nil
}

func ShouldEnableNodejsModule(stackId string) bool {
	switch stackId {
	case "io.buildpacks.stacks.ubi8", "io.buildpacks.stacks.ubi9":
		return true
	case "io.buildpacks.stacks.ubi10":
		return false
	default:
		return true
	}
}
