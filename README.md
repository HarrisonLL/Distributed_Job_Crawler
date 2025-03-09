# Main
The point of the project is to simulate a distributed microservice that schedules jobs and send notifications to registered users. 


# System Architecture

<img width="1504" alt="Screenshot 2025-03-02 at 2 26 00 PM" src="https://github.com/user-attachments/assets/b094fe86-9787-42dd-9a5f-458173e3f9ac" />


## RUN crawler worker
```
docker run -v ./html_data/:/app/html_data --env-file ./.env harrisonll/jc_worker:test --job_type "software engineer" --location "USA" --company "meta"
```

## Development
### Golang Services
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

## Deployment
### docker-compose
```
# comment out gocron goemail in docker-compose
docker-compose up -d
docker exec -it goweb sh
./main --service=CLI add "amazon" "software engineer" "harrisonll/jc_worker:v1.0.0-linux"
./main --service=CLI add "meta" "software engineer" "harrisonll/jc_worker:v1.0.0-linux"
docker-compose down
docker-compose up -d
```
