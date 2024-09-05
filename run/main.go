package main

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/scribe"

	ubinodejsextension "github.com/paketo-community/ubi-nodejs-extension"
)

const IMAGES_JSON_PATH = "/etc/buildpacks/images.json"

func main() {
	dependencyManager := postal.NewService(cargo.NewTransport())
	logEmitter := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))
	duringBuildPermissions := ubinodejsextension.GetDuringBuildPermissions("/etc/passwd")
	imagesManager := ubinodejsextension.NewImagesManager(IMAGES_JSON_PATH)

	packit.RunExtension(
		ubinodejsextension.Detect(),
		ubinodejsextension.Generate(dependencyManager, logEmitter, duringBuildPermissions, imagesManager),
	)
}
