package domain_test

import (
	"dreamkast-weaver/internal/dkui/domain"
	"dreamkast-weaver/internal/dkui/value"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newSlotID(v int32) value.SlotID {
	id, _ := value.NewSlotID(v)
	return id
}

func newTrackID(v int32) value.TrackID {
	id, _ := value.NewTrackID(v)
	return id
}

func newTalkID(v int32) value.TalkID {
	id, _ := value.NewTalkID(v)
	return id
}

var (
	svc = domain.DkUiService{}
)

func TestDkUiService_CreateOnlineWatchEvent(t *testing.T) {

	slotID := newSlotID(42)
	trackID := newTrackID(1)
	talkID := newTalkID(2)

	tests := []struct {
		name                      string
		given                     func() *domain.WatchEvents
		shouldStampChallengeAdded bool
	}{
		{
			name: "stamp condition fulfilled",
			given: func() *domain.WatchEvents {
				events := &domain.WatchEvents{}
				for i := 0; i < 9; i++ {
					ev := *domain.NewOnlineWatchEvent(newTrackID(11), newTalkID(22), slotID)
					ev.CreatedAt = ev.CreatedAt.Add(time.Duration(-1 * (value.GUARD_SECONDS + 1) * time.Second))
					events = events.AddImmutable(ev)
				}
				return events
			},
			shouldStampChallengeAdded: true,
		},
		{
			name: "stamp condition not fulfilled",
			given: func() *domain.WatchEvents {
				events := &domain.WatchEvents{}
				for i := 0; i < 8; i++ {
					ev := *domain.NewOnlineWatchEvent(newTrackID(11), newTalkID(22), slotID)
					ev.CreatedAt = ev.CreatedAt.Add(time.Duration(-1 * (value.GUARD_SECONDS + 1) * time.Second))
					events = events.AddImmutable(ev)
				}
				return events
			},
			shouldStampChallengeAdded: false,
		},
	}

	for _, tt := range tests {
		t.Run("ok:"+tt.name, func(t *testing.T) {
			stamps := &domain.StampChallenges{}
			events := tt.given()
			evLen := len(events.Items)

			got, err := svc.CreateOnlineWatchEvent(trackID, talkID, slotID, stamps, events)

			assert.Nil(t, err)
			assert.Equal(t, trackID, got.TrackID)
			assert.Equal(t, talkID, got.TalkID)
			assert.Equal(t, slotID, got.SlotID)
			assert.Equal(t, value.ViewingSeconds120, got.ViewingSeconds)
			assert.Equal(t, evLen, len(events.Items))
			if tt.shouldStampChallengeAdded {
				assert.Equal(t, 1, len(stamps.Items))
				stamp := stamps.Items[0]
				assert.Equal(t, value.StampReady, stamp.Condition)
			} else {
				assert.Equal(t, 0, len(stamps.Items))
			}
		})
	}

	errTests := []struct {
		name  string
		given func() *domain.WatchEvents
	}{
		{
			name: "too short request",
			given: func() *domain.WatchEvents {
				events := &domain.WatchEvents{}
				ev := *domain.NewOnlineWatchEvent(newTrackID(11), newTalkID(22), slotID)
				ev.CreatedAt = ev.CreatedAt.Add(time.Duration(-1 * (value.GUARD_SECONDS - 9) * time.Second))
				events = events.AddImmutable(ev)
				return events
			},
		},
	}

	for _, tt := range errTests {
		t.Run("err:"+tt.name, func(t *testing.T) {
			stamps := &domain.StampChallenges{}
			events := tt.given()

			_, err := svc.CreateOnlineWatchEvent(trackID, talkID, slotID, stamps, events)
			assert.Error(t, err)
		})
	}

}

func TestDkUiService_StampOnline(t *testing.T) {

	slotID := newSlotID(42)

	t.Run("ok", func(t *testing.T) {
		stamps := &domain.StampChallenges{[]domain.StampChallenge{
			*domain.NewStampChallenge(newSlotID(41)),
			*domain.NewStampChallenge(newSlotID(42)),
			*domain.NewStampChallenge(newSlotID(43)),
		}}

		err := svc.StampOnline(slotID, stamps)
		assert.Nil(t, err)

		for _, stamp := range stamps.Items {
			if stamp.SlotID == slotID {
				assert.Equal(t, value.StampStamped, stamp.Condition)
			} else {
				assert.Equal(t, value.StampSkipped, stamp.Condition)
			}
		}
	})

	errTests := []struct {
		name  string
		given func() *domain.StampChallenges
	}{
		{
			name: "ready stamp not found",
			given: func() *domain.StampChallenges {
				return &domain.StampChallenges{[]domain.StampChallenge{
					*domain.NewStampChallenge(newSlotID(41)),
					*domain.NewStampChallenge(newSlotID(43)),
				}}
			},
		},
	}

	for _, tt := range errTests {
		t.Run("err:"+tt.name, func(t *testing.T) {
			stamps := tt.given()
			err := svc.StampOnline(slotID, stamps)
			assert.Error(t, err)
		})
	}
}
