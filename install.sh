#!/bin/bash
. tools/gen_tls_key.sh
sqlite3 ctsesam.sqlite.db < init/install.sql
