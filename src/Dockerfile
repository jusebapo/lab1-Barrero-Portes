FROM golang:1.27-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api cmd/main.go

FROM scratch
COPY --from=build /out/api /api

EXPOSE 8000
ENTRYPOINT [ "/api" ]