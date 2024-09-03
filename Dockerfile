FROM golang:1.23.0 AS build

WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app

# copy app binary from building stage to production image
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/app .

EXPOSE 50051

CMD [ "./app" ]
