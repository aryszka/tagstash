package sql

// generated code
const Cmd_get_tags = `

select tag from tags
where value = $1;
`
