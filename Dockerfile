# DESCRIPTION:  Run docker-alertd in a container
# COMMENTS:
#
#       This mounts the docker socket into a container and then run docker-alertd with
#       access to the host machine. This is a security risk because the container that it
#       is run in will have root access on the host.
#
# USAGE:
#
#       # Pull or build the docker image
#       # docker pull deltaskelta/docker-alertd
#
#       # If you don't already have a config file, generate one to standard out using the
#       # docker-alertd in the container, save the file as `.docker-alertd.yaml`
#       docker run --rm  \
#         deltaskelta/docker-alertd \
#         go-wrapper run initconfig --stdout
#
#       # Run docker-alertd with the created config file and mounted docker socket
#       docker run --rm \
#         -v /var/run/docker.sock:/var/run/docker.sock \
#         -v ~/.docker-aled.yaml:/root/.docker-alertd.yaml \
#         deltaskelta/docker-alertd
#
FROM golang:1.9

WORKDIR /go/src/app
COPY . .

RUN go-wrapper download   # "go get -d -v ./..."
RUN go-wrapper install    # "go install -v ./..."

CMD ["go-wrapper", "run" ] # ["app"]
