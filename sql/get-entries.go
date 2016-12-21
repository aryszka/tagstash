package sql

// generated code
const Cmd_get_entries = `

select
  tag,
  value,
  tag_index
from tags
where tag in (%s);
`
