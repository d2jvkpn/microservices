#! /usr/bin/env bash
set -eu -o pipefail
_wd=$(pwd)
_path=$(dirname $0 | xargs -i readlink -f {})


#### 1. install grpc
apt install -y protobuf-compiler

go get -u google.golang.org/grpc
go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc

mkdir proto && cd proto
protoc --go_out=./  --go-grpc_out=./  proto/auth.proto
sed -i '/^\tmustEmbedUnimplemented/s#\t#\t// #' proto/auth_grpc.pb.go


#### 2. install postgres
cargo install --version=0.6.2 sqlx-cli --no-default-features --features native-tls,postgres

command -v sqlx

export DATABASE_URL=postgres://hello:world@127.0.0.1:5432/authentication
echo "export DATABASE_URL=$DATABASE_URL" >> .env

sqlx database create
# sqlx database drop

# psql --host 127.0.0.1 --username hello --port 5432 --password --dbname users -c 'SELECT current_database()'


#### 3. migrations
sqlx migrate add create_users_table
sql_file=$(ls migrations/*_create_users_table.sql | tail -n 1)
ls 

cat >> $sql_file <<EOF
CREATE TYPE user_status AS ENUM('ok', 'blocked', 'deleted');

CREATE FUNCTION update_now() RETURNS trigger AS \$\$
BEGIN
  NEW.updated_at := now();
  RETURN NEW;
END;
\$\$LANGUAGE plpgsql;
-- drop function update_now cascade;

CREATE TABLE users (
  id         UUID DEFAULT gen_random_uuid(),
  PRIMARY    KEY (id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),

  bah     VARCHAR(72)  NOT NULL,
  status  user_status  NOT NULL DEFAULT 'ok'
);

CREATE TRIGGER updated_at BEFORE INSERT OR UPDATE ON users
  FOR EACH ROW EXECUTE PROCEDURE update_now();
EOF

sqlx migrate run
