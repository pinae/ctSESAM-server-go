/*
   Copyright (c) 2018 Oliver Lau <ola@ct.de>, Heise Medien GmbH & Co. KG

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

func main() {
	rawCert, err := ioutil.ReadFile("../cert/server.crt")
	if err != nil {
		log.Fatal(err)
	}
	block, _ := pem.Decode([]byte(rawCert))
	if block == nil {
		panic("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic("failed to parse certificate: " + err.Error())
	}

	now := time.Now()
	fmt.Printf("Valid? %t\n", now.After(cert.NotBefore) && now.Before(cert.NotAfter))
	fmt.Printf("E-Mail addresses: %s\n", cert.EmailAddresses)
	fmt.Printf("DNS names: %s\n", cert.DNSNames)
	fmt.Printf("Issuer: %s\n", cert.Issuer)
	fmt.Printf("Not before: %s\n", cert.NotBefore.String())
	fmt.Printf("Not after : %s\n", cert.NotAfter.String())
}
