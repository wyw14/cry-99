FROM golang:1.26.2
ENV GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local CGO_ENABLED=0
WORKDIR /workspace
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .
RUN go build -mod=vendor -o /usr/local/bin/railvolt ./cmd/railvolt
EXPOSE 19699
CMD ["/usr/local/bin/railvolt", "-listen", "0.0.0.0:19699", "-data", "/var/lib/railvolt", "-web", "/workspace/web"]
