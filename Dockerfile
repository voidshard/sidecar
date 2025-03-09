FROM golang:1.22 as build

ARG USERNAME=app
ARG USER_UID=42000
ARG USER_GID=$USER_UID

RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME

WORKDIR /go/src/github.com/voidshard/sidecar
ADD go.mod go.sum ./
RUN go mod download
ADD *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /app *.go && chown -R $USER_UID:$USER_GID /app

FROM scratch

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /app /app
COPY --from=build /etc/ssl/certs/ /etc/ssl/certs/

USER $USERNAME
ENTRYPOINT ["/app"]
