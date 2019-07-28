FROM golang:1.12
WORKDIR /dlk8s

ENV GOPROXY=https://goproxy.io
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY backend backend
COPY bakers bakers
COPY cmd cmd
COPY events events
COPY logging logging
COPY models models
COPY stores stores
RUN go build -o server ./cmd/backend/

FROM golang:1.12
RUN groupadd -g 999 appuser && \
    useradd -r -u 999 -g appuser appuser
WORKDIR /home/appuser
COPY --from=0 /dlk8s/server /home/appuser
RUN chown -R appuser:appuser /home/appuser
USER appuser
ENTRYPOINT ["/home/appuser/server"]

