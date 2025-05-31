# Main
The project is a simulation of distributed microservice that schedules crawling jobs and send notifications to registered users.
The purpose is to learn multithreading in Golang and its Docker and K8S sdk.


# System Architecture

<img width="785" alt="image" src="https://github.com/user-attachments/assets/2a6d5e08-312a-4005-9929-8d5f5346eac7" />


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

