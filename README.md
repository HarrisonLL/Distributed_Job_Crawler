<img width="878" alt="image" src="https://github.com/HarrisonLL/Distributed_Job_Crawler/assets/32115568/28876f49-7a23-4f76-bdb4-45a05f593fa2">
System Architecture
0. May not implement user authentication service
1. In K8S case, go service will spawn K8S pod
2. If users are lot, it's better to implement a queue system to send emails. Same for Crawler works


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
migrate -database 'postgres://admin:adminpass@172.17.0.1:5432/gs_db?sslmode=disable' -path ./migrations up
```
