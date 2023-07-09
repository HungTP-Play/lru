# LRU (URL)

Simple URL shortener.

## Architecture

![Architecture](https://i.ibb.co/sFdm1s1/2023-07-07-14-08.png)

## Run

```bash
docker-compose build
docker-compose up
```

## Test

### Frontend

Open the following URL in your browser:

The simple frontend is available at:

```bash
http://localhost:3000
```

### Monitoring

The Grafana dashboard is available at:

```bash
http://localhost:3001
```

#### Dashboard

Go to `Dashboards` -> `Metrics`

#### Logs

Go to `Explore`, Select datasource `Loki`, and query `{jobs="log"}`

#### Trace

Go to `Explore`, Select datasource `Tempo`, select `Search` and run query.

### Send request via cRUL

```bash
curl  -X POST \
  'http://localhost:3333/shorten' \
  --header 'Accept: */*' \
  --header 'User-Agent: Thunder Client (https://www.thunderclient.com)' \
  --header 'Content-Type: application/json' \
  --data-raw '{
  "url":"https://google.com"
}'
```

```bash
curl  -X GET \
  'http://localhost:3333/redirect' \
  --header 'Accept: */*' \
  --header 'User-Agent: Thunder Client (https://www.thunderclient.com)' \
  --header 'Content-Type: application/json' \
  --data-raw '{
  "url": "http://localhost/x"
}'
```

## Crate fake traffic

```bash
docker compose up
```
