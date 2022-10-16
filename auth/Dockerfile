FROM golang:1.19 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -o /app ./cmd

FROM alpine
WORKDIR /app
COPY --from=build /app /app/app
EXPOSE 8080
CMD ["./app"]
