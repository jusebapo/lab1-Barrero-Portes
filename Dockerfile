FROM golang:1.27-bookworm AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/api .

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=build /out/api /app/api

USER 65532:65532
EXPOSE 8000
CMD ["/app/api"]