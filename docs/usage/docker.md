# Docker

## Running Queries

For users who prefer to use Docker, SGPT provides a Docker image:

1. Pull the latest Docker image:

```shell
docker pull ghcr.io/tbckr/sgpt:latest
```

1. Run queries using the Docker image:

```shell
$ docker run --rm -e OPENAI_API_KEY=${OPENAI_API_KEY} ghcr.io/tbckr/sgpt:latest txt "mass of sun"
The mass of the sun is approximately 1.989 x 10^30 kilograms.
```

## Saving Chat Sessions

When using SGPT within a Docker container, you can mount a local folder to the container's `/home/nonroot` path to save
and persist all active chat sessions. This allows you to maintain your chat history and resume previous conversations
across different container instances.

To mount a local folder and save chat sessions, follow these steps:

1. Pull the SGPT Docker image:

```shell
docker pull ghcr.io/tbckr/sgpt:latest
```

1. Create a local folder to store your chat sessions, e.g. `sgpt-chat-sessions`:

```shell
mkdir sgpt-chat-sessions
```

1. Change the permissions of the folder to the nonroot user of the Docker image:

```shell
sudo chown 65532:65532 sgpt-chat-sessions
```

1. Run the Docker container with the local folder mounted to `/home/nonroot`:

```shell
$ docker run --rm -e OPENAI_API_KEY=${OPENAI_API_KEY} -v $(pwd)/sgpt-chat-sessions:/home/nonroot ghcr.io/tbckr/sgpt:latest txt "mass of sun"
The mass of the sun is approximately 1.99 x 10^30 kilograms.
$ docker run --rm -e OPENAI_API_KEY=${OPENAI_API_KEY} -v $(pwd)/sgpt-chat-sessions:/home/nonroot ghcr.io/tbckr/sgpt:latest txt "convert to earth masses"
To convert the mass of the sun to earth masses, we need to divide it by the mass of the Earth:
1.99 x 10^30 kg / 5.97 x 10^24 kg = 333,000 Earth masses (rounded to the nearest thousand) 
So the mass of the sun is about 333,000 times greater than the mass of the Earth.
```
