#! /usr/bin/env bash
set -eu -o pipefail
_wd=$(pwd)
_path=$(dirname $0 | xargs -i readlink -f {})

docker-compose pull

docker-compose up -d

exit

docker exec -it postgres_db psql --username postgres --dbname postgres --password

sql=```postgres
alter user postgres with password 'XXXXXXXX';

create user hello with password 'world';
create database authentication with owner = hello;
```

psql --host 127.0.0.1 --port 5432 --username hello --dbname authentication --password
