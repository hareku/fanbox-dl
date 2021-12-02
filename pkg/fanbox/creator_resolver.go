package fanbox

import (
	"context"
	"fmt"
)

type CreatorID string

type CreatorResolverDoInput struct {
	InputCreatorID    CreatorID
	IncludeSupporting bool
	IncludeFollowing  bool
}

type CreatorResolver interface {
	Do(ctx context.Context, in *CreatorResolverDoInput) ([]CreatorID, error)
}

func NewCreatorResolver(api API) CreatorResolver {
	return &creatorResolver{api}
}

type creatorResolver struct {
	api API
}

func (c *creatorResolver) Do(ctx context.Context, in *CreatorResolverDoInput) ([]CreatorID, error) {
	if in.InputCreatorID != "" {
		return []CreatorID{in.InputCreatorID}, nil
	}

	ids := map[CreatorID]interface{}{}

	if in.IncludeSupporting {
		plans, err := c.api.ListPlans(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list plans: %w", err)
		}
		for _, p := range plans.Body {
			ids[CreatorID(p.CreatorID)] = nil
		}
	}

	if in.IncludeFollowing {
		following, err := c.api.ListFollowing(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list following: %w", err)
		}
		for _, f := range following.Body {
			ids[CreatorID(f.CreatorID)] = nil
		}
	}

	res := []CreatorID{}
	for id := range ids {
		res = append(res, id)
	}
	return res, nil
}
