# Paketo Node.js Extension for ubi

The Node.js Extension for
[ubi](https://www.redhat.com/en/blog/introducing-red-hat-universal-base-image)
allows builders to be created which build Node.js applications on top of
Red Hat's Node.js ubi containers. For example
[ubi8/nodejs-16-minimal](https://catalog.redhat.com/software/containers/ubi8/nodejs-16-minimal/615aefd53f6014fa45ae1ae2).

## Integration

The ubi Node.js extension provides node and npm as dependencies.
Downstream buildpacks, like Yarn Install CNB or NPM CNB, can
require the node dependency by generating a Build Plan TOML
file that requires node and or npm.

The extension integrates with the existing Paketo buildpacks
so that building your application will have the same experience
as building with non ubi stacks. The main difference is that
node.js and npm will be provided by the extension intead of the
node-engine build pack.

## Usage

Support for extensions is currently experimental and the construction
of a ubi8 based builder is still a work in progress. Until that work
completes you have to complete a few steps to try out the Node.js
ubi extension.

1. Ensure you have either a local registry or a remote registry
   that you can push a builder to. If you want to run a local
   registry an easy way to do that is using
   `docker run -d -p 5000:5000 --restart=always --name registry registry:2`.
1. Get a current version of pack and ensure it is on your path.
   It should be at least version v0.30.0-pre2 or later. The releases are
   available from [here](https://github.com/buildpacks/pack/releases).
1. Clone the repository for this extension
1. cd into the directory into which you have cloned the repository
   and run scripts/build.sh. This builds the binaries for the extension.
1. Enable experimental features in pack by running
   `pack config experimental true`. This is needed because extensions
   are currently experimental.
1. Build the `detect` and `generate` binaries of the extension by running the `./scripts/build.sh` script.
1. Create a builder which includes the extension. The extension works together
   with the existing Paketo node.js buildpack so a minimal builder
   requires both the node.js buildpack and the extention as follows:

   ```
   description = "Sample builder that uses ubi Node.js extension to support Node.js apps"

   [[buildpacks]]
     uri = "docker://gcr.io/paketo-buildpacks/nodejs:1.4.0"
     version = "1.4.0"

   [lifecycle]
     version = "0.17.0"

   [[order]]
     [[order.group]]
       id = "paketo-buildpacks/nodejs"
       version = "1.4.0"

   [[extensions]]
     id = "paketo-community/ubi-nodejs-extension"
     version = "0.0.1"
     uri = "file:///home/user1/paketo/ubi-nodejs-extension"

   [[order-extensions]]
     [[order-extensions.group]]
       id = "paketo-community/ubi-nodejs-extension"
       version = "0.0.1"

   [stack]
     id = "io.buildpacks.stacks.ubi8"
     build-image = "paketocommunity/build-ubi-base"
     run-image = "paketocommunity/run-ubi-base"
   ```

   To create the builder:

   1. create a file called builder.toml with the minimal builder toml
      shown above. Modify the uri for the ubi-nodejs-extension so that
      if reflects the path where your clone of the ubi-nodejs-extesion
      repository exists.
   1. run `pack builder create localhost:5000/test-builder --config ./builder.toml`
      to create the builder. Replace localhost:5000 with your public
      registry if you are not running a local registry.
   1. run `docker push localhost:5000/test-builder` to push the builder to the
      local registry or push to a public registry if desired.
1. Build your Node.js application with the new builder:
   1. run `pack build test-app --path ./app-dir --builder localhost:5000/test-builder --network host -v`
      where test-app will be the name of the image built and app-dir is
      directory that contains your Node.js application. Replace
      `localhost:5000` with the host:port for the public repository
      if you are not using a local repostiory.
   1. run your application with `docker run -p 8080:8080 test-app` replacing
      `8080:8080` with the port on which your application listens.
   1. access your running application and enjoy :).

## Configurations

### Specifying a Node version

ubi only supports the latest version of each Node.js stream
currently available in the ubi version. At the time of writing
ubi8 supports the Node.js 16, and 18 streams. For example,
if the latest Node.js version for the 16.x stream in ubi8 is 16.10.1
then that is your only option when requesting the Node.js 16.x stream.
Therefore we suggest that you request the Node.js version such that it
will accept any version of the stream you want to use with something like
`~16`.

To specify the version of the Node that is installed, set the `$BP_NODE_VERSION`
environment variable at build time either directly (ex. `pack build my-app
--env BP_NODE_VERSION=~16`) or through a [`project.toml`
file](https://github.com/buildpacks/spec/blob/main/extensions/project-descriptor.md)

```shell
$BP_NODE_VERSION="~16"
```

You can also specify a node version via an `.nvmrc` or `.node-version` file,
also at the application directory root.

### Specifying a project path

To specify a project subdirectory to be used as the root of the app, please use
the `BP_NODE_PROJECT_PATH` environment variable at build time either directly
(ex. `pack build my-app --env BP_NODE_PROJECT_PATH=./src/my-app`) or through a
[`project.toml`
file](https://github.com/buildpacks/spec/blob/main/extensions/project-descriptor.md).
This could be useful if your app is a part of a monorepo.

## Run Tests

To run all unit tests, run:

```
./scripts/unit.sh
```

!!! Work in progress the integration.sh script does not yet exist.

To run all integration tests, run:

```
/scripts/integration.sh
```

## Package buildpack (Generate .tgx & .cnb files)

To generate `buildpackage.cnb` and `buildpack.tgz` files

```
./scripts/package.sh --version 0.0.1
```
