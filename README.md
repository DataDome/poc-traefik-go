# DataDome Traefik POC

## How to run it

```
export DATADOME_SERVER_SIDE_KEY=YOUR_SERVER_SIDE_KEY
docker-compose up --build
curl -v localhost
```

The Set-Cookie DataDome should be available on the response.