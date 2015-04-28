FROM google/golang

ADD . /gopath/src/github.com/may215/go2aws
WORKDIR /gopath/src/github.com/may215/go2aws
RUN go get
RUN go install
RUN cp -r public /var/www

ENTRYPOINT ["/gopath/bin/go2aws"]
CMD ["-public", "/var/www"]