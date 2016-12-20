select
  tag,
  value,
  tag_index
from tags
where tag in (%s);
