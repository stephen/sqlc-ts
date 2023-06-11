/* name: GetAuthor :one */
SELECT * FROM authors
WHERE id = ? LIMIT 1;

/* name: ListAuthors :many */
SELECT * FROM authors
ORDER BY name;

/* name: CreateAuthor :execresult */
INSERT INTO authors (
  name, bio
) VALUES (
  ?, ?
);

/* name: DeleteAuthor :exec */
DELETE FROM authors
WHERE id = ?;

/* name: SearchAuthorsByName :many */
select id from authors where name like '%' || @text || '%';

/* name: SearchAuthorsByNameWithUnknown :many */
select cast(id as INTEGER) from authors where name like '%' || @text || '%';

/* name: CountAuthors :one */
select count(1) from authors;
