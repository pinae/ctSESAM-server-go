# c't SESAM Storage Server

This is a rewrite of the [c't SESAM server](https://github.com/ola-ct/ctSESAM-server). It's now written in Go instead of PHP.

_c't SESAM Storage Server_ (the _SESAM server_) is a data store for [c't SESAM](https://github.com/ola-ct/Qt-SESAM). It's supposed to be installed on a Unix-like machine and used by only you or a small community (your family, close friends). It implements a simple REST API to read, write and delete entries in an SQLite 3 database.

## Installation

### Prerequisites

You need a Unix machine with Go 1.11 (or newer) and SQLite 3. It depends on your operating system (one the many Linux flavors, macOS etc.) how to install these components.

You also need some extra Go packages: "github.com/mattn/go-sqlite3", "github.com/abbot/go-http-auth" and "golang.org/x/crypto/bcrypt". You can easily `go get` them, e.g. `go get github.com/mattn/go-sqlite3`.

### Get the code

Clone the _SESAM server_ repository into a directory of your choice:

```
git clone https://github.com/ola-ct/ctSESAM-server-go.git
```

This creates the directory ctSESAM-server-go containing the &ast;.go files with the server's code.

### Add user

To restrict access _SESAM server_ authenticates users by means of [HTTP Basic authentication](https://en.wikipedia.org/wiki/Basic_access_authentication). The credentials are read from an Apache [.htpasswd](https://en.wikipedia.org/wiki/.htpasswd) file which must be placed in the top-level directory of the _SESAM server_, i.e. the directory you created in the previous step. To create a file for the user `demo` execute the following shell command:

```
htpasswd -B -c .htpasswd demo
```

You're then asked to type in the desired password (twice).

Option `-B` enables [bcrypt](https://en.wikipedia.org/wiki/Bcrypt) hashing of the password. _SESAM server_ doesn't support any other hash methods.

If your .htpasswd file has a different name or is stored in another location, you can change the value of the constant `credentialsFile` in main.go accordingly.

### Install SSL certificate

Save your SSL server certificate (as "server.crt") and the accompanying server key (as "server.key") into the subdirectory "cert".

If you want to use other file names please edit the call to `srv.ListenAndServeTLS("cert/server.crt", "cert/server.key")` in main.go.

The easiest and no-cost way to obtain these files is by using [acme.sh](https://github.com/Neilpang/acme.sh). You can also get a certificate from other [certificate authorities](https://en.wikipedia.org/wiki/Certificate_authority) (CA) than [Let's encrypt](https://letsencrypt.org/) like [Comodo](https://www.comodo.com/), [Godaddy](https://www.godaddy.com/web-security/ssl-certificate) or [GlobalSign](https://www.globalsign.com/en/ssl/). It's not possible to use the desktop version of c't SESAM ([Qt SESAM](https://github.com/ola-ct/Qt-SESAM)) with [self-signed certificates](https://en.wikipedia.org/wiki/Self-signed_certificate).

## Run _SESAM server_

You can now start the server with

```
./run.sh
```

_SESAM server_ should print something like this:

```
*** c't SESAM storage server 0.1.2 (go1.11.2)
Copyright (c) 2017-2018 Oliver Lau <ola@ct.de>
All rights reserved.

Opening log file SESAM.log ...
Parsing credentials in ./.htpasswd ...
Opening database ./ctsesam.sqlite.db ...
Starting database cleanup job ...
Starting secure web server on port 8443 ...
```

As you can see _SESAM server_ launches its HTTPS listener on port 8443. If that port is already occupied by another service, you can change the port by editing the value of the constant `port` in main.go.

## Configuration

A future release of _SESAM server_ will let you configure the port and the location of the .htpasswd and certificate files with the aid of a configuration file.

---

Copyright 2017-2018 [Oliver Lau](mailto:ola@ct.de), Heise Medien GmbH & Co. KG

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at [http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0).

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

---

__Diese Software wurde zu Lehr- und Demonstrationszwecken programmiert und ist nicht für den produktiven Einsatz vorgesehen. Der Autor und die Heise Medien GmbH & Co. KG haften nicht für eventuelle Schäden, die aus der Nutzung der Software entstehen, und übernehmen keine Gewähr für ihre Vollständigkeit, Fehlerfreiheit und Eignung für einen bestimmten Zweck.__
