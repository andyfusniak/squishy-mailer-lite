begin immediate;

-- strftime('%Y-%m-%d %H:%M:%f000000+00:00', 'now')   <- current timestamp Go style

create table if not exists projects (
  id            text primary key,
  pname         varchar(1024) not null,
  description   text not null default '',
  created_at    text not null
);

create table if not exists transports (
  id                   text,
  project_id           text not null,
  trname               text not null,
  host                 text not null,
  port                 integer not null,
  username             text not null,
  encrypted_password   text not null,
  email_from           text not null,
  email_replyto        text not null,
  created_at           text not null,
  modified_at          text not null,
  primary key (id, project_id),
  constraint transports_project_id_fkey foreign key (project_id) references projects (id)
);

create table if not exists groups (
  id            text,
  project_id    text not null,
  gname         text not null,
  created_at    text not null,
  modified_at   text not null,
  primary key (id, project_id),
  constraint groups_project_id_fkey foreign key (project_id) references projects (id)
);

commit;
