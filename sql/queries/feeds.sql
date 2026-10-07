-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name,url,user_id)
VALUES (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6
)
RETURNING *;
-- name: ReadAllFeeds :many
select feeds.name as feed_name,url,users.name as name
from feeds
inner join users on feeds.user_id=users.id;
-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *
)
SELECT
inserted_feed_follow.*,
feeds.name AS feed_name,
users.name AS user_name
FROM inserted_feed_follow
INNER JOIN feeds ON inserted_feed_follow.feed_id = feeds.id
INNER JOIN users ON inserted_feed_follow.user_id = users.id;
-- name: DeleteFeedFollow :exec
delete from feed_follows
where feed_follows.user_id=$1 and feed_follows.feed_id=$2;
-- name: FeedbyUrl :one
SELECT * FROM feeds WHERE url = $1;
-- name: GetFeedFollows :many
SELECT
feed_follows.*,
feeds.name AS feed_name,
users.name AS user_name
FROM feed_follows
INNER JOIN feeds ON feed_follows.feed_id = feeds.id
INNER JOIN users ON feed_follows.user_id = users.id
where
feed_follows.user_id=$1;
-- name: MarkFeedFetched :exec
update feeds
set last_fetched_at=current_TIMEstamp,updated_at=current_TIMEstamp
where feeds.id=$1;
-- name: GetNextFeedToFetch :one
select *
from feeds
order by last_fetched_at nulls first
limit 1;
-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;
-- name: GetPostsForUser :many
SELECT posts.*, feeds.name AS feed_name FROM posts
JOIN feed_follows ON feed_follows.feed_id = posts.feed_id
JOIN feeds ON posts.feed_id = feeds.id
WHERE feed_follows.user_id = $1
ORDER BY posts.published_at DESC
LIMIT $2;
