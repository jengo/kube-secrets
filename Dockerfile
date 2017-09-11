FROM golang:1.8-alpine

WORKDIR /go/src/kube-secrets

ENV EDITOR=vim

RUN apk update \
	&& apk add git bash vim make \
	&& rm -rf /var/cache/apk/*
	
RUN go get -v github.com/mattn/goveralls \
	github.com/go-yaml/yaml \
	github.com/stretchr/testify \
	golang.org/x/tools/cmd/cover

# go-wrapper doesn't install this correctly
# I was only using go-wrapper to provide some type of output
RUN go get -u github.com/jstemmer/go-junit-report

ADD bash.sh /bash.sh

ENTRYPOINT /bash.sh
CMD ["/bash.sh"]
