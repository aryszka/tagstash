package sql

// generated code
const Cmd_insert_entry_pq = `

insert into tags
(tag, value, tag_index)
values ($1, $2, $3)
on conflict(tag, value) do
update set tag_index = $3;
`
