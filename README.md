# Dev Notes
## RUN crawler worker
```
docker run -v ./html_data/:/app/html_data --env-file ./.env harrisonll/jc_worker:test --job_type "software engineer" --location "USA" --company "meta"
```

## Golang Service

- download dependencies
```
go mod tidy
```

- migrate db
```
migrate create -ext sql -dir ./migrations/ -seq init
migrate -database 'postgres://admin:adminpass@localhost:5432/gs_db?sslmode=disable' -path ./migrations up
```
