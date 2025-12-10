FROM golang:1.23-alpine

WORKDIR /app

# We will handle the build in the docker-compose context
COPY . .

RUN go build -o nexus-server ./nexus

EXPOSE 8080

CMD ["./nexus-server"]
