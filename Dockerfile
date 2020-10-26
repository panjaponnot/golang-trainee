# Stage 0 - Building server application
FROM golang:1.15 as builder
WORKDIR /app
ADD . .

# Disable CGO and compile server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w' -o golang_template .
RUN chmod +x golang_template

# Stage 1 - Server start
FROM alpine:3.8
WORKDIR /app
COPY --from=builder /app/golang_template /app/golang_template

# Install tzdata and set docker image timezone to bangkok timezone
RUN apk add --no-cache tzdata
ENV TZ=Asia/Bangkok

# Expose service port
EXPOSE 5000
CMD ["/app/golang_template", "start"]
