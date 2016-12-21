package sql

// generated code
const Cmd_create_db = `

create table tags (
  tag text not null,
  value text not null,
  tag_index int,
  primary key (tag, value)
);
`
