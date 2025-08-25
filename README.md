# backend-go

## starting process
project setup
``go mod init backend-go``

install dependencies
```
go get github.com/golang-jwt/jwt/v5
go get github.com/gorilla/mux
go get go.mongodb.org/mongo-driver/mongo
go get github.com/joho/godotenv


```
Start mongodb if run locally:
```sudo systemctl start mongod.service```

Run the program
```
    go run .
    go run main.go //use this main.go than . only
```
