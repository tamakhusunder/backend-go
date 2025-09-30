# backend-go

## Process when project was started
project setup
``go mod init backend-go``

install dependencies
```
go get github.com/golang-jwt/jwt/v5
go get github.com/gorilla/mux
go get go.mongodb.org/mongo-driver/mongo
go get github.com/joho/godotenv


```

## Running the project 
Start mongodb if run locally:
```sudo systemctl start mongod.service```

Run the program
```
    go run .
    go run main.go //use this main.go than . only
```

new update------------
now direct run `go run ./cmd` //to
    run all files in the cmd file as package name is main



Redis start
```
sudo systemctl enable redis-server
sudo systemctl start redis-server
```

Basically a mix of Clean Architecture / Hexagonal Architecture
