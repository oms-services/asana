FROM golang

RUN go get github.com/gorilla/mux

RUN go get github.com/orijtech/asana/v1

RUN go get github.com/odeke-em/asana/v1

RUN go get github.com/cloudevents/sdk-go

WORKDIR /go/src/github.com/oms-services/asana

ADD . /go/src/github.com/oms-services/asana

RUN go install github.com/oms-services/asana

ENTRYPOINT asana

EXPOSE 3000