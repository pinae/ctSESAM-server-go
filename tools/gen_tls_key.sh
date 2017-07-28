#!/bin/bash

openssl req -x509 -nodes \
  -newkey rsa:2048 \
  -keyout ./cert/server.rsa.key \
  -out ./cert/server.rsa.crt \
  -days 365 \
  -config ./cert/openssl.cnf \
  -batch
