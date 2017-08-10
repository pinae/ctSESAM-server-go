FROM golang:onbuild
RUN mkdir /app 
ADD . /app/ 
WORKDIR /app
EXPOSE 8443
RUN apt-get update
RUN apt-get install -y sqlite3
RUN go build -o ctSESAM-storage-server .
RUN ./init/init.sh
CMD ["/app/ctSESAM-storage-server"]
