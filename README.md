# DataDome Traefik POC

## How to run it

### Install the dependencies

```sh
cd ./plugins/src/github.com/traefik/plugin_datadome
go mod tidy
go mod vendor
```

### Launch the docker container

```sh
export DATADOME_SERVER_SIDE_KEY=YOUR_SERVER_SIDE_KEY
docker-compose up --build
curl -v localhost
```

The Set-Cookie DataDome should be available on the response.