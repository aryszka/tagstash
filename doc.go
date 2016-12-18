/*
Package tagstash provides tagging for arbitrary string values, typically URIs.

Tagstash implements many-to-many associations between values and tags. It is designed to return the best match
for a query with multiple tags. It prioritizes the entries based on how many tags they match, and if still
multiple values come out as the best, it takes the order of the querying tags into account.

It stores the value-tag associations in a persistent storage, and caches the most often queried tags in memory.
Both the persistence layer and the cache can be replaced with a custom implementation of a simple interface
(Storage). When evaluating a query, tagstash tries to find the best match first in the cache, and if any tags in
the query cannot be found there, only then fetches their associations from the persistent storage.
*/
package tagstash
