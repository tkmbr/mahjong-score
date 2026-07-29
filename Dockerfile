FROM golang:1.24-bookworm AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/mahjong-score ./cmd/server \
    && mkdir /out/data

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/mahjong-score /mahjong-score
COPY --from=build --chown=65532:65532 /out/data /data
VOLUME ["/data"]
EXPOSE 8080
ENV ADDR=:8080
ENV DATABASE_PATH=/data/mahjong-score.db
ENTRYPOINT ["/mahjong-score"]
