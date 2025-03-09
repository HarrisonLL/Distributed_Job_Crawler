# Main
The aim of the project is to simulate a distributed microservice that schedules crawling jobs and send notifications to registered users. 


# System Architecture

<img width="1504" alt="Screenshot 2025-03-02 at 2 26 00 PM" src="https://github.com/user-attachments/assets/b094fe86-9787-42dd-9a5f-458173e3f9ac" />


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
### worker container
```
docker run -v --env-file ./.env harrisonll/jc_worker:test --job_type "software engineer" --location "USA" --company "meta"
```


## Deployment
### docker-compose (linux)
```
# comment out gocron goemail in docker-compose
docker-compose up -d
# run CLI to add initial data to job type table
docker exec -it goweb sh
./main --service=CLI add "amazon" "software engineer" "harrisonll/jc_worker:v1.0.0-linux"
./main --service=CLI add "meta" "software engineer" "harrisonll/jc_worker:v1.0.0-linux"
# uncomment gocron and goemai
# note sometimes you may not be able to join the docker network immediately
# prune you container and do up and down a few times, it will show up in the jc_network
docker-compose down
docker-compose up -d
```
<img width="607" alt="Screenshot 2025-03-09 at 1 21 43 PM" src="https://github.com/user-attachments/assets/22879399-0580-48ad-99f7-3a8cef377c18" />

