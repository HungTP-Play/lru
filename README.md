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
docker compose up -d
```

The wait about 10s , un-comment the following line in `docker-compose.yml`:

```yaml
  # k6-shorten:
  #   image: grafana/k6
  #   volumes:
  #     - ./k6:/scripts
  #   command:
  #     - 'run'
  #     - '/scripts/shorten.js'
  #   depends_on:
  #     - gateway
  # k6-redirect:
  #   image: grafana/k6
  #   volumes:
  #     - ./k6:/scripts
  #   command:
  #     - 'run'
  #     - '/scripts/redirect.js'
  #   depends_on:
  #     - gateway
```

Then run:

```bash
docker compose up -d
```
