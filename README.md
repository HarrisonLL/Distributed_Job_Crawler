The point of the project is to simulate a distributed microservice that schedules jobs and send notifications to registered users. 


# System Architecture
<img width="767" alt="Screen Shot 2024-09-15 at 9 17 52 PM" src="https://github.com/user-attachments/assets/297e466e-32c6-4af1-8f10-160ebdff634f">


# Dev Notes
## RUN crawler worker
```
docker run -v ./html_data/:/app/html_data --env-file ./.env harrisonll/jc_worker:test --job_type "software engineer" --location "USA" --company "meta"
```

## Golang Service

### dev env
- download dependencies
```
go mod tidy
```

- migrate db
```
migrate create -ext sql -dir ./migrations/ -seq init
migrate -database 'postgres://admin:adminpass@localhost:5432/gs_db?sslmode=disable' -path ./migrations up
```

- start service
```
go run main.go -service web
go run main.go -service scheduler
go run main.go -service emailConsumer
```
