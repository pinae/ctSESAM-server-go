FROM golang:onbuild
RUN mkdir /app 
ADD . /app/ 
WORKDIR /app
EXPOSE 8443
RUN yum update
RUN yum install -y sqlite3
RUN go build -o ctSESAM-storage-server .
CMD ["/app/ctSESAM-storage-server"]
