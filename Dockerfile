FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

RUN microdnf install tar gzip \
  && curl https://dl.google.com/go/go1.12.10.linux-amd64.tar.gz | tar zxC /usr/local \
  && mkdir -p /go/bin /go/src /go/pkg \
  && microdnf clean all

ENV GOPATH="/go"
ENV PATH=$PATH:/usr/local/go/bin

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

RUN go build -o ./out/fileops .

CMD ["./out/fileops"]