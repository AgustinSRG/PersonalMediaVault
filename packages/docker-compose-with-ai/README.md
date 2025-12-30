# Docker compose setup (With AI features)

This folder contains a setup for running the vault in a container, with all AI features enabled.

The [docker-compose.yml](./docker-compose.yml) contains the configuration for the containers that will be created.

## Configuring

Copy the file `.env.example` into `.env`:

```sh
cp .env.example .env
```

Edit the configuration variables, reading the documentation comments.

## Running

To run the setup, after it is configured, run:

```sh
docker compose up -d
```

## Upgrading

In order to upgrade the image, and re-create the containers run:

```sh
docker compose pull
docker compose up -d
```

## Stopping

If you wish to stop the containers, run:

```sh
docker compose down
```
