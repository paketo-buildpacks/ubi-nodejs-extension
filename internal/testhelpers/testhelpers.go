package testhelpers

import (
	"fmt"
	"strings"
)

func GenerateImagesJsonFile(nodeVersions []string, isDefault []bool, isCorrupted bool, ubiVersion string) string {
	var images []string

	images = append(images, fmt.Sprintf(`{
      "name": "default",
      "config_dir": "stacks/stack",
      "output_dir": "builds/build",
      "build_image": "build",
      "run_image": "run",
      "create_build_image": true,
      "base_build_container_image": "docker://registry.access.redhat.com/ubi%s/ubi-minimal",
      "base_run_container_image": "docker://registry.access.redhat.com/ubi%s/ubi-minimal",
      "stack_type": "base",
      "pattern_image_registry_name": "os_name-os_codename-build_image_run_image-stack_type",
      "pattern_assets_prefix": "os_name-os_codename-build_image_run_image-stack_type-version-arch"
    }`, ubiVersion, ubiVersion))

	images = append(images, fmt.Sprintf(`  {
      "name": "runtime-1",
      "config_dir": "stacks/stack-runtime-1",
      "output_dir": "builds/build-runtime-1",
      "build_image": "build-runtime-1",
      "run_image": "run-runtime-1",
      "base_run_container_image": "docker://registry.access.redhat.com/ubi%s/platform-1-runtime",
      "stack_type": "base",
      "pattern_image_registry_name": "os_name-os_codename-build_image_run_image-stack_type",
      "pattern_assets_prefix": "os_name-os_codename-build_image_run_image-stack_type-version-arch"
    }`, ubiVersion))

	images = append(images, fmt.Sprintf(`  {
      "name": "runtime-2",
      "config_dir": "stacks/stack-runtime-2",
      "output_dir": "builds/build-runtime-2",
      "build_image": "build-runtime-2",
      "run_image": "run-runtime-2",
      "base_run_container_image": "docker://registry.access.redhat.com/ubi%s/platform-2-runtime",
      "stack_type": "base",
      "pattern_image_registry_name": "os_name-os_codename-build_image_run_image-stack_type",
      "pattern_assets_prefix": "os_name-os_codename-build_image_run_image-stack_type-version-arch"
    }`, ubiVersion))

	for i, nodeVersion := range nodeVersions {
		images = append(images, fmt.Sprintf(`  {
      "name": "nodejs-%s",
      "is_default_run_image": %t,
      "config_dir": "stacks/stack-nodejs-%s",
      "output_dir": "builds/build-nodejs-%s",
      "build_image": "build-nodejs-%s",
      "run_image": "run-nodejs-%s",
      "base_run_container_image": "docker://registry.access.redhat.com/ubi%s/nodejs-%s-runtime",
      "stack_type": "base",
      "pattern_image_registry_name": "os_name-os_codename-build_image_run_image-stack_type",
      "pattern_assets_prefix": "os_name-os_codename-build_image_run_image-stack_type-version-arch"
    }`, nodeVersion, isDefault[i], nodeVersion, nodeVersion, nodeVersion, nodeVersion, ubiVersion, nodeVersion))
	}

	if isCorrupted {
		images = append(images, `{
			"name": "nodejs-18",}
			not a valid json
		}`)
	}

	imagesJson := fmt.Sprintf(`{
  "support_usns": false,
  "update_on_new_image": true,
  "receipts_show_limit": 16,
  "images": [
    %s
  ]
}
`, strings.Join(images, ",\n  "))

	return imagesJson
}
