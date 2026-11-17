# ---- Stage 1: build ----
FROM golang:1-alpine AS build
WORKDIR /app
COPY go.mod ./
COPY *.go ./
# ดาวน์โหลดไลบรารี (SQLite) และสร้าง go.sum ให้อัตโนมัติ
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o server .

# ---- Stage 2: run ----
FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/server .
RUN mkdir -p /data
EXPOSE 8080
CMD ["./server"]
