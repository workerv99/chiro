FROM node:20-alpine AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.23-alpine AS build
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 go build -trimpath -o /bin/chiro ./cmd/server

FROM alpine:3.20
RUN addgroup -S chiro && adduser -S -G chiro chiro
USER chiro
COPY --from=build /bin/chiro /chiro
COPY --from=web /src/web/build /web
ENV WEB_DIST=/web
EXPOSE 8080
ENTRYPOINT ["/chiro"]
