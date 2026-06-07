# build a static binary, ship it on scratch (~10MB image)
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /tveitan .

FROM scratch
COPY --from=build /tveitan /tveitan
# content/ and theme/ are bind-mounted at runtime so they stay live-editable
# on the VPS without rebuilding the image (see docker-compose.yml).
ENV ADDR=":8080" CONTENT_DIR="/content" THEME_DIR="/theme"
EXPOSE 8080
ENTRYPOINT ["/tveitan"]
