-- Add migration script here
CREATE TYPE user_status AS ENUM('ok', 'blocked', 'deleted');

CREATE FUNCTION update_now() RETURNS trigger AS $$
BEGIN
  NEW.updated_at := now();
  RETURN NEW;
END;
$$LANGUAGE plpgsql;
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
