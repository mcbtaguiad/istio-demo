FROM golang:1.26-alpine3.22

LABEL Maintainer="Mark Taguiad <marktaguiad@marktaguiad.dev>"

WORKDIR /app

# COPY go.mod go.sum ./
# RUN go mod download

RUN go mod init login-app
RUN go get github.com/redis/go-redis/v9
RUN go get golang.org/x/crypto/bcrypt
RUN go get github.com/google/uuid
RUN go get github.com/golang-jwt/jwt/v5
RUN go get github.com/rs/cors

COPY main.go /app/main.go

RUN go build -o app

EXPOSE 3000

CMD ["./app"]
