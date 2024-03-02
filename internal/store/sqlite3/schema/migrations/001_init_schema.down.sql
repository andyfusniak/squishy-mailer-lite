begin immediate;

drop table if exists mailqueue;
drop table if exists templates;
drop table if exists groups;
drop table if exists smtp_transports;
drop table if exists projects;

commit;
