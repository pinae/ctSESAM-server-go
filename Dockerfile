FROM golang:bullseye
RUN mkdir /app
ADD . /app/
WORKDIR /app
EXPOSE 80
RUN apt-get update
RUN apt-get install -y sqlite3
RUN go get github.com/mattn/go-sqlite3 && go get github.com/abbot/go-http-auth && go get golang.org/x/crypto/bcrypt
RUN go build -o ctSESAM-storage-server .
CMD ["/app/ctSESAM-storage-server"]
