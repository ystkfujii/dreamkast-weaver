// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.2

package db

import (
	"time"
)

type CfpVote struct {
	ConferenceName string
	TalkID         int32
	Dt             time.Time
}
