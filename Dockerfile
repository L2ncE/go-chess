FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOPROXY https://goproxy.cn,direct

WORKDIR /build

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o /app/main ./main.go


FROM scratch

ENV TZ Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/main /app/main

EXPOSE 6666
CMD ["./main"]