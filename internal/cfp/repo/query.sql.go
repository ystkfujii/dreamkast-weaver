// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.2
// source: query.sql

package repo

import (
	"context"
	"database/sql"
)

const insertCfpVote = `-- name: InsertCfpVote :exec
INSERT INTO cfp_votes (
  conference_name,
  talk_id,
  global_ip,
  created_at
) VALUES ( 
  ?, ?, ?, now()
)
`

type InsertCfpVoteParams struct {
	ConferenceName string
	TalkID         int32
	GlobalIp       sql.NullString
}

func (q *Queries) InsertCfpVote(ctx context.Context, arg InsertCfpVoteParams) error {
	_, err := q.db.ExecContext(ctx, insertCfpVote, arg.ConferenceName, arg.TalkID, arg.GlobalIp)
	return err
}

const listCfpVotes = `-- name: ListCfpVotes :many
SELECT conference_name, talk_id, created_at, global_ip FROM cfp_votes
WHERE conference_name = ?
`

func (q *Queries) ListCfpVotes(ctx context.Context, conferenceName string) ([]CfpVote, error) {
	rows, err := q.db.QueryContext(ctx, listCfpVotes, conferenceName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []CfpVote
	for rows.Next() {
		var i CfpVote
		if err := rows.Scan(
			&i.ConferenceName,
			&i.TalkID,
			&i.CreatedAt,
			&i.GlobalIp,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
