package sql

// generated code
const Cmd_insert_entry = `

insert or replace into tags
(tag, value, tag_index)
values ($1, $2, $3);
`
