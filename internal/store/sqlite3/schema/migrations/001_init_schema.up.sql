begin immediate;

-- strftime('%Y-%m-%d %H:%M:%f000000+00:00', 'now')   <- current timestamp Go style

--
-- projects are the top level entity providing a namespace
--
create table if not exists projects (
  project_id    text not null,
  project_name  text not null default '',
  description   text not null default '',
  created_at    text not null,
  primary key (project_id)
);

--
-- smtp transports are used to send emails
--
create table if not exists smtp_transports (
  smtp_transport_id    text not null,
  project_id           text not null,
  transport_name       text not null,
  host                 text not null,
  port                 integer not null,
  username             text not null,
  encrypted_password   text not null,
  email_from           text not null,
  email_replyto        text not null,
  created_at           text not null,
  modified_at          text not null,
  primary key (smtp_transport_id, project_id),
  constraint transports_project_id_fkey foreign key (project_id) references projects (project_id)
);


--
-- Groups are a way to organise templates
--
create table if not exists groups (
  group_id     text not null,
  project_id   text not null,
  group_name   text not null,
  created_at   text not null,
  modified_at  text not null,
  primary key (group_id, project_id),
  constraint groups_project_id_fkey foreign key (project_id) references projects (project_id)
);

--
-- Templates are the actual email templates
--
create table if not exists templates (
  template_id    text not null,
  group_id       text not null,
  project_id     text not null,
  txt            text not null,
  html           text not null,
  created_at     text not null,
  modified_at    text not null,
  primary key (template_id, group_id, project_id),
  -- templates are unique within a group
  constraint templates_template_id_project_id_uindex unique (template_id, project_id),
  constraint templates_group_id_project_id_fkey
    foreign key (group_id, project_id)
    references groups (group_id, project_id)
);

commit;
