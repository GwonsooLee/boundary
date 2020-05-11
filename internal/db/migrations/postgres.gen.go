// Code generated by "make migrations"; DO NOT EDIT.
package migrations

var postgresMigrations = map[string]*fakeFile{
	"migrations": {
		name: "migrations",
	},
	"migrations/01_oplog.up.sql": {
		name: "01_oplog.up.sql",
		bytes: []byte(`
CREATE TABLE if not exists oplog_entry (
  id bigint generated always as identity primary key,
  create_time timestamp with time zone default current_timestamp,
  update_time timestamp with time zone default current_timestamp,
  version text NOT NULL,
  aggregate_name text NOT NULL,
  "data" bytea NOT NULL
);
CREATE TABLE if not exists oplog_ticket (
  id bigint generated always as identity primary key,
  create_time timestamp with time zone default current_timestamp,
  update_time timestamp with time zone default current_timestamp,
  "name" text NOT NULL UNIQUE,
  "version" bigint NOT NULL
);
CREATE TABLE if not exists oplog_metadata (
  id bigint generated always as identity primary key,
  create_time timestamp with time zone default current_timestamp,
  entry_id bigint NOT NULL REFERENCES oplog_entry(id) ON DELETE CASCADE ON UPDATE CASCADE,
  "key" text NOT NULL,
  value text NULL
);
create index if not exists idx_oplog_metatadata_key on oplog_metadata(key);
create index if not exists idx_oplog_metatadata_value on oplog_metadata(value);
INSERT INTO oplog_ticket (name, version)
values
  ('default', 1),
  ('iam_scope', 1),
  ('iam_user', 1),
  ('iam_auth_method', 1),
  ('iam_group', 1),
  ('iam_group_member_user', 1),
  ('iam_role', 1),
  ('iam_role_grant', 1),
  ('iam_role_group', 1),
  ('iam_role_user', 1),
  ('db_test_user', 1),
  ('db_test_car', 1),
  ('db_test_rental', 1);
`),
	},
	"migrations/02_domain_types.down.sql": {
		name: "02_domain_types.down.sql",
		bytes: []byte(`
begin;

drop domain wt_public_id;

commit;

`),
	},
	"migrations/02_domain_types.up.sql": {
		name: "02_domain_types.up.sql",
		bytes: []byte(`
begin;

create domain wt_public_id as text
check(
  length(trim(value)) > 10
);
comment on domain wt_public_id is
'Random ID generated with github.com/hashicorp/vault/sdk/helper/base62';

commit;

`),
	},
	"migrations/03_db.down.sql": {
		name: "03_db.down.sql",
		bytes: []byte(`
drop table if exists db_test_user;
drop table if exists db_test_car;
drop table if exists db_test_rental;
`),
	},
	"migrations/03_db.up.sql": {
		name: "03_db.up.sql",
		bytes: []byte(`
-- create test tables used in the unit tests for the internal/db package
-- these tables (db_test_user, db_test_car, db_test_rental) are not part
-- of the Watchtower domain model... they are simply used for testing the internal/db package
CREATE TABLE if not exists db_test_user (
  id bigint generated always as identity primary key,
  create_time timestamp with time zone default current_timestamp,
  update_time timestamp with time zone default current_timestamp,
  public_id wt_public_id NOT NULL UNIQUE,
  name text UNIQUE,
  phone_number text,
  email text
);
CREATE TABLE if not exists db_test_car (
  id bigint generated always as identity primary key,
  create_time timestamp with time zone default current_timestamp,
  update_time timestamp with time zone default current_timestamp,
  public_id wt_public_id NOT NULL UNIQUE,
  name text UNIQUE,
  model text,
  mpg smallint
);
CREATE TABLE if not exists db_test_rental (
  id bigint generated always as identity primary key,
  create_time timestamp with time zone default current_timestamp,
  update_time timestamp with time zone default current_timestamp,
  public_id wt_public_id NOT NULL UNIQUE,
  name text UNIQUE,
  user_id bigint not null REFERENCES db_test_user(id),
  car_id bigint not null REFERENCES db_test_car(id)
);
`),
	},
	"migrations/04_iam.down.sql": {
		name: "04_iam.down.sql",
		bytes: []byte(`
BEGIN;

drop table if exists iam_scope CASCADE;
drop trigger if exists iam_scope_insert;
drop function if exists iam_sub_scopes_func;
drop table if exists iam_auth_method cascade;
drop table if exists  iam_role cascade;
drop table if exists iam_group_member_type_enm cascade;
drop table if exists iam_group cascade cascade;
drop table if exists iam_group_member_user cascade;
drop view if exists iam_group_member_vw;
drop table if exists iam_auth_method_type_enm cascade;
drop table if exists iam_action_enm cascade;
drop table if exists iam_role_type_enm cascade;
drop table if exists iam_role_user cascade;
drop table if exists iam_role_group cascade;
drop table if exists iam_role_grant cascade;
drop view if exists iam_assigned_role_vw;


COMMIT;
`),
	},
	"migrations/04_iam.up.sql": {
		name: "04_iam.up.sql",
		bytes: []byte(`
BEGIN;

CREATE TABLE if not exists iam_scope_type_enm (
  string text NOT NULL primary key CHECK(string IN ('unknown', 'organization', 'project'))
);
INSERT INTO iam_scope_type_enm (string)
values
  ('unknown'),
  ('organization'),
  ('project');

 
CREATE TABLE if not exists iam_scope (
    public_id wt_public_id primary key,
    create_time timestamp with time zone default current_timestamp,
    update_time timestamp with time zone default current_timestamp,
    name text,
    type text NOT NULL REFERENCES iam_scope_type_enm(string) CHECK(
      (
        type = 'organization'
        and parent_id = NULL
      )
      or (
        type = 'project'
        and parent_id IS NOT NULL
      )
    ),
    description text,
    parent_id text REFERENCES iam_scope(public_id) ON DELETE CASCADE ON UPDATE CASCADE
  );
create table if not exists iam_scope_organization (
    scope_id wt_public_id NOT NULL UNIQUE REFERENCES iam_scope(public_id) ON DELETE CASCADE ON UPDATE CASCADE,
    name text UNIQUE,
    primary key(scope_id)
  );
create table if not exists iam_scope_project (
    scope_id wt_public_id NOT NULL REFERENCES iam_scope(public_id) ON DELETE CASCADE ON UPDATE CASCADE,
    parent_id wt_public_id NOT NULL REFERENCES iam_scope_organization(scope_id) ON DELETE CASCADE ON UPDATE CASCADE,
    name text,
    unique(parent_id, name),
    primary key(scope_id, parent_id)
  );


CREATE
  OR REPLACE FUNCTION iam_sub_scopes_func() RETURNS TRIGGER
SET SCHEMA
  'public' LANGUAGE plpgsql AS $$ DECLARE parent_type INT;
BEGIN IF new.type = 'organization' THEN
insert into iam_scope_organization (scope_id, name)
values
  (new.public_id, new.name);
return NEW;
END IF;
IF new.type = 'project' THEN
insert into iam_scope_project (scope_id, parent_id, name)
values
  (new.public_id, new.parent_id, new.name);
return NEW;
END IF;
RAISE EXCEPTION 'unknown scope type';
END;
$$;


CREATE TRIGGER iam_scope_insert
AFTER
insert ON iam_scope FOR EACH ROW EXECUTE PROCEDURE iam_sub_scopes_func();


CREATE
  OR REPLACE FUNCTION iam_immutable_scope_type_func() RETURNS TRIGGER
SET SCHEMA
  'public' LANGUAGE plpgsql AS $$ DECLARE parent_type INT;
BEGIN IF new.type != old.type THEN
RAISE EXCEPTION 'scope type cannot be updated';
END IF;
return NEW;
END;
$$;

CREATE TRIGGER iam_scope_update
BEFORE
update ON iam_scope FOR EACH ROW EXECUTE PROCEDURE iam_immutable_scope_type_func();

COMMIT;


CREATE TABLE if not exists iam_user (
    public_id wt_public_id not null primary key,
    create_time timestamp with time zone NOT NULL default current_timestamp,
    update_time timestamp with time zone NOT NULL default current_timestamp,
    name text UNIQUE,
    description text,
    external_name text NOT NULL,
    scope_id wt_public_id NOT NULL REFERENCES iam_scope_organization(scope_id),
    disabled BOOLEAN NOT NULL default FALSE
  );


CREATE TABLE if not exists iam_auth_method (
    public_id wt_public_id not null primary key, 
    create_time timestamp with time zone NOT NULL default current_timestamp,
    update_time timestamp with time zone NOT NULL default current_timestamp,
    name text UNIQUE,
    description text,
    scope_id wt_public_id NOT NULL REFERENCES iam_scope_organization(scope_id) ON DELETE CASCADE ON UPDATE CASCADE,
    disabled BOOLEAN NOT NULL default FALSE,
    type text NOT NULL
  );


CREATE TABLE if not exists iam_role (
    public_id wt_public_id not null primary key,
    create_time timestamp with time zone NOT NULL default current_timestamp,
    update_time timestamp with time zone NOT NULL default current_timestamp,
    name text UNIQUE,
    description text,
    scope_id wt_public_id NOT NULL REFERENCES iam_scope(public_id) ON DELETE CASCADE ON UPDATE CASCADE,
    disabled BOOLEAN NOT NULL default FALSE
  );

CREATE TABLE if not exists iam_group_member_type_enm (
    string text NOT NULL primary key CHECK(string IN ('unknown', 'user'))
  );
INSERT INTO iam_group_member_type_enm (string)
values
  ('unknown'),
  ('user');


CREATE TABLE if not exists iam_group (
    public_id wt_public_id not null primary key,
    create_time timestamp with time zone NOT NULL default current_timestamp,
    update_time timestamp with time zone NOT NULL default current_timestamp,
    name text UNIQUE,
    description text,
    scope_id wt_public_id NOT NULL REFERENCES iam_scope(public_id) ON DELETE CASCADE ON UPDATE CASCADE,
    disabled BOOLEAN NOT NULL default FALSE
  );


CREATE TABLE if not exists iam_group_member_user (
    create_time timestamp with time zone NOT NULL default current_timestamp,
    group_id wt_public_id NOT NULL REFERENCES iam_group(public_id) ON DELETE CASCADE ON UPDATE CASCADE,
    member_id wt_public_id NOT NULL REFERENCES iam_user(public_id) ON DELETE CASCADE ON UPDATE CASCADE,
    primary key (group_id, member_id)
  );


CREATE VIEW iam_group_member_vw AS
SELECT
  *, 'user' as type
FROM iam_group_member_user;


CREATE TABLE if not exists iam_auth_method_type_enm (
    string text NOT NULL primary key CHECK(string IN ('unknown', 'userpass', 'oidc'))
  );
INSERT INTO iam_auth_method_type_enm (string)
values
  ('unknown'),
  ('userpass'),
  ('oidc');
ALTER TABLE iam_auth_method
ADD
  FOREIGN KEY (type) REFERENCES iam_auth_method_type_enm(string);

CREATE TABLE if not exists iam_action_enm (
    string text NOT NULL primary key CHECK(
      string IN (
        'unknown',
        'list',
        'create',
        'update',
        'edit',
        'delete',
        'authen'
      )
    )
  );

INSERT INTO iam_action_enm (string)
values
  ('unknown'),
  ('list'),
  ('create'),
  ('update'),
  ('edit'),
  ('delete'),
  ('authen');

CREATE TABLE if not exists iam_role_user (
    create_time timestamp with time zone NOT NULL default current_timestamp,
    role_id wt_public_id NOT NULL REFERENCES iam_role(public_id) ON DELETE CASCADE ON UPDATE CASCADE,
    principal_id wt_public_id NOT NULL REFERENCES iam_user(public_id) ON DELETE CASCADE ON UPDATE CASCADE,
    primary key (role_id, principal_id)
  );

CREATE TABLE if not exists iam_role_group (
    create_time timestamp with time zone NOT NULL default current_timestamp,
    role_id wt_public_id NOT NULL REFERENCES iam_role(public_id) ON DELETE CASCADE ON UPDATE CASCADE,
    principal_id wt_public_id NOT NULL REFERENCES iam_group(public_id) ON DELETE CASCADE ON UPDATE CASCADE,
    primary key (role_id, principal_id)
  );

CREATE VIEW iam_assigned_role_vw AS
SELECT
  *, 'user' as type
FROM iam_role_user
UNION
select
  *, 'group' as type
from iam_role_group;



CREATE TABLE if not exists iam_role_grant (
    public_id wt_public_id not null primary key,
    create_time timestamp with time zone NOT NULL default current_timestamp,
    update_time timestamp with time zone NOT NULL default current_timestamp,
    name text UNIQUE,
    description text,
    role_id wt_public_id NOT NULL REFERENCES iam_role(public_id) ON DELETE CASCADE ON UPDATE CASCADE,
    "grant" text NOT NULL
  );

  COMMIT;

`),
	},
}
