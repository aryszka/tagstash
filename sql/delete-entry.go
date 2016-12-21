package sql

// generated code
const Cmd_delete_entry = `

delete from tags
where
  tag = $1 and
  value = $2;
`
