FROM golang:1.12

WORKDIR $GOPATH/src/github.com/pakohler/jenkronize
COPY . .
RUN go get github.com/go-yaml/yaml github.com/cavaliercoder/grab
RUN go build .
RUN mkdir -p /opt/jenkronize/data
RUN mv jenkronize /opt/jenkronize/

# When running this docker image, the following should be performed:
#   - in config.yaml, use /opt/jenkronize/data as the base directory for caching artifacts
#   - mount your config.yaml as a volume to /opt/jenkronize/config.yaml
#   - mount the local directory you want to have artifacts cached to as a volume at /opt/jenkronize/data

ENTRYPOINT ["/opt/jenkronize/jenkronize"]
