# Development

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- param::isNotitle::true:: -->
<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Source Build

### Go Build

To build SGPT from the source code using Go, follow these steps:

1. Ensure you have [Go](https://go.dev/dl/) installed on your system.
2. Clone the SGPT repository:

```shell
git clone https://github.com/tbckr/sgpt.git
```

3. Navigate to the cloned repository:

```shell
cd sgpt
```

4. Build the SGPT executable using Go:

```shell
go build -o sgpt cmd/sgpt/main.go
```

This will create an executable named sgpt in the current directory.

### Build Docker Image

To build a Docker image for SGPT, use the `docker-build.sh` script located in the `bin` folder:

1. Make sure you are in the root of the cloned SGPT repository.
2. Make the `docker-build.sh` script executable:

```shell
chmod +x bin/docker-build.sh
```

3. Run the `docker-build.sh` script to build the Docker image:

```shell
bin/docker-build.sh
```

The script will build the Docker image for the `linux/amd64` platform with the tag `sgpt:latest`. Change the platform,
build args or tag according to your needs.