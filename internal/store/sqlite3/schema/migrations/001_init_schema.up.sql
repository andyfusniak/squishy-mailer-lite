begin immediate;

-- strftime('%Y-%m-%d %H:%M:%f000000+00:00', 'now')   <- current timestamp Go style

create table if not exists projects (
  id            text primary key,
  pname         varchar(1024) not null,
  description   text not null default '',
  created_at    text not null
);

commit;
