FROM golang:latest as build

WORKDIR /src/
COPY . /src/
RUN go mod download
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -buildmode=plugin -o /usr/lib/olric-consul-plugin.so

FROM olricio/olricd:v0.5.0-beta.2
COPY --from=build /usr/lib/olric-consul-plugin.so /usr/lib/olric-consul-plugin.so
