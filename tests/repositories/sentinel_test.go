package repositories_test

import (
	"errors"
	"testing"

	"my-go-app/internal/repositories"
)

func TestSentinelErrors_AreDistinct(t *testing.T) {
	sentinels := []struct {
		name string
		err  error
	}{
		{"ErrUserNotFound", repositories.ErrUserNotFound},
		{"ErrProjectNotFound", repositories.ErrProjectNotFound},
		{"ErrProjectUnauthorized", repositories.ErrProjectUnauthorized},
		{"ErrImageNotFound", repositories.ErrImageNotFound},
		{"ErrPageNotFound", repositories.ErrPageNotFound},
		{"ErrPageUnauthorized", repositories.ErrPageUnauthorized},
		{"ErrSnapshotNotFound", repositories.ErrSnapshotNotFound},
		{"ErrCommentNotFound", repositories.ErrCommentNotFound},
		{"ErrCommentUnauthorized", repositories.ErrCommentUnauthorized},
		{"ErrInvitationNotFound", repositories.ErrInvitationNotFound},
		{"ErrInvitationExpired", repositories.ErrInvitationExpired},
		{"ErrInvitationInvalid", repositories.ErrInvitationInvalid},
		{"ErrContentTypeNotFound", repositories.ErrContentTypeNotFound},
		{"ErrContentFieldNotFound", repositories.ErrContentFieldNotFound},
		{"ErrContentItemNotFound", repositories.ErrContentItemNotFound},
		{"ErrCustomElementNotFound", repositories.ErrCustomElementNotFound},
		{"ErrCustomElementTypeNotFound", repositories.ErrCustomElementTypeNotFound},
		{"ErrElementNotFound", repositories.ErrElementNotFound},

	}

	for i, a := range sentinels {
		for j, b := range sentinels {
			if i == j {
				continue
			}
			if errors.Is(a.err, b.err) {
				t.Errorf("%s and %s should be distinct sentinel errors", a.name, b.name)
			}
		}
	}
}

func TestSentinelErrors_NonNilAndHaveMessage(t *testing.T) {
	sentinels := []struct {
		name string
		err  error
	}{
		{"ErrUserNotFound", repositories.ErrUserNotFound},
		{"ErrProjectNotFound", repositories.ErrProjectNotFound},
		{"ErrProjectUnauthorized", repositories.ErrProjectUnauthorized},
		{"ErrImageNotFound", repositories.ErrImageNotFound},
		{"ErrPageNotFound", repositories.ErrPageNotFound},
		{"ErrPageUnauthorized", repositories.ErrPageUnauthorized},
		{"ErrSnapshotNotFound", repositories.ErrSnapshotNotFound},
		{"ErrCommentNotFound", repositories.ErrCommentNotFound},
		{"ErrCommentUnauthorized", repositories.ErrCommentUnauthorized},
		{"ErrInvitationNotFound", repositories.ErrInvitationNotFound},
		{"ErrInvitationExpired", repositories.ErrInvitationExpired},
		{"ErrInvitationInvalid", repositories.ErrInvitationInvalid},
		{"ErrContentTypeNotFound", repositories.ErrContentTypeNotFound},
		{"ErrContentFieldNotFound", repositories.ErrContentFieldNotFound},
		{"ErrContentItemNotFound", repositories.ErrContentItemNotFound},
		{"ErrCustomElementNotFound", repositories.ErrCustomElementNotFound},
		{"ErrCustomElementTypeNotFound", repositories.ErrCustomElementTypeNotFound},
		{"ErrElementNotFound", repositories.ErrElementNotFound},

	}

	for _, tc := range sentinels {
		if tc.err == nil {
			t.Errorf("%s must not be nil", tc.name)
		}
		if tc.err != nil && tc.err.Error() == "" {
			t.Errorf("%s must have a non-empty message", tc.name)
		}
	}
}

func TestSentinelErrors_ErrorsIsWorksForEachSentinel(t *testing.T) {
	cases := []struct {
		name     string
		sentinel error
	}{
		{"ErrUserNotFound", repositories.ErrUserNotFound},
		{"ErrProjectNotFound", repositories.ErrProjectNotFound},
		{"ErrProjectUnauthorized", repositories.ErrProjectUnauthorized},
		{"ErrImageNotFound", repositories.ErrImageNotFound},
		{"ErrPageNotFound", repositories.ErrPageNotFound},
		{"ErrPageUnauthorized", repositories.ErrPageUnauthorized},
		{"ErrSnapshotNotFound", repositories.ErrSnapshotNotFound},
		{"ErrCommentNotFound", repositories.ErrCommentNotFound},
		{"ErrInvitationNotFound", repositories.ErrInvitationNotFound},
		{"ErrContentTypeNotFound", repositories.ErrContentTypeNotFound},
		{"ErrContentItemNotFound", repositories.ErrContentItemNotFound},

	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if !errors.Is(tc.sentinel, tc.sentinel) {
				t.Errorf("%s: errors.Is(sentinel, sentinel) returned false", tc.name)
			}
			wrapped := errors.Join(errors.New("outer"), tc.sentinel)
			if !errors.Is(wrapped, tc.sentinel) {
				t.Errorf("%s: errors.Is(wrapped, sentinel) returned false", tc.name)
			}
		})
	}
}

func TestSentinelErrors_CommentSentinelsAreDistinct(t *testing.T) {
	if errors.Is(repositories.ErrCommentNotFound, repositories.ErrCommentUnauthorized) {
		t.Error("ErrCommentNotFound and ErrCommentUnauthorized must be distinct")
	}
}

func TestSentinelErrors_InvitationSentinelsAreDistinct(t *testing.T) {
	pairs := []struct{ a, b error }{
		{repositories.ErrInvitationNotFound, repositories.ErrInvitationExpired},
		{repositories.ErrInvitationNotFound, repositories.ErrInvitationInvalid},
		{repositories.ErrInvitationExpired, repositories.ErrInvitationInvalid},
	}
	for _, p := range pairs {
		if errors.Is(p.a, p.b) {
			t.Errorf("%v and %v should be distinct", p.a, p.b)
		}
	}
}