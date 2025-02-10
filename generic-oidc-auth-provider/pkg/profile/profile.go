package profile

import (
	"context"
)

func FetchProfileIconURL(ctx context.Context, accessToken string) (string, error) {
	return "https://www.gravatar.com/avatar/?d=identicon", nil
}
