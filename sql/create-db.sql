create table tags (
  tag text not null,
  value text not null,
  tag_index int,
  primary key (tag, value)
);
