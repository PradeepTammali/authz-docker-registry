FROM golang:1.13.0-alpine3.10 as builder

WORKDIR /docker-registry-serviceaccount-validation

COPY src .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/docker-registry-serviceaccount-validation

# Runtime image
FROM scratch AS base
COPY --from=builder /go/bin/docker-registry-serviceaccount-validation /bin/docker-registry-serviceaccount-validation
ENTRYPOINT ["/bin/docker-registry-serviceaccount-validation"]
