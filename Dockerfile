# USE GO 1.20 to build this image
# docker build -t golang:1.20 .

FROM golang:1.20

# Install go-taskfile
RUN curl -sL https://taskfile.dev/install.sh | sh

# Copy the local package files to the container's workspace.
ADD app /go/src/github.com/winkelchri/jira-export/app
ADD pkg /go/src/github.com/winkelchri/jira-export/pkg
ADD main.go /go/src/github.com/winkelchri/jira-export/main.go
ADD Taskfile.yaml /go/src/github.com/winkelchri/jira-export/Taskfile.yaml
ADD go.mod /go/src/github.com/winkelchri/jira-export/go.mod
ADD go.sum /go/src/github.com/winkelchri/jira-export/go.sum
WORKDIR /go/src/github.com/winkelchri/jira-export

# Build the jira-export command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go get -d -v ./...
RUN task build

# Copy the binary to the production image from the builder stage.
FROM alpine:latest
# Install Bash
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/winkelchri/jira-export/dist/jira-export .

CMD ["./jira-export"]
