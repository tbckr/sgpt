# Development

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
<!-- param::isNotitle::true:: -->

- [Source Build](#source-build)
  - [Go Build](#go-build)
  - [Build Docker Image](#build-docker-image)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Preqrequisites

We use Task to manage our build and development tasks. To install Task, follow the instructions on
the [Task website](https://taskfile.dev/installation/). Please make sure you have task installed before you begin.

## Source Build

### Go Build

To build SGPT from the source code using Go, follow these steps:

1. Ensure you have [Go](https://go.dev/dl/) installed on your system.
2. Clone the SGPT repository:

  ```shell
  git clone https://github.com/tbckr/sgpt.git
  ```

1. Navigate to the cloned repository:

  ```shell
  cd sgpt
  ```

1. Build the SGPT executable using task:

  ```shell
  task build
  ```

This will create an executable named sgpt in the current directory.

### Build Docker Image

To build a Docker image for SGPT, use the task target `docker:build`:

1. Make sure you are in the root of the cloned SGPT repository.
2. Run the `docker:build` task:

  ```shell
  task docker:build
  ```

This will build the Docker image for the `linux/amd64` and `linux/arm64` platform with the tag `sgpt:latest`. To build a
docker image for specific platform either use `docker:build:linux-amd64` or `docker:build:linux-arm64` task targets.
