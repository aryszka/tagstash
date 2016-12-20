delete from tags
where
  tag = $1 and
  value = $2;
