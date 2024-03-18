package fanbox

import (
	"context"
	"fmt"
	"net/http"
)

type CreatorIDListerDoInput struct {
	InputCreatorIDs   []string
	IncludeSupporting bool
	IncludeFollowing  bool
	IgnoreCreatorIDs  []string
}

type CreatorIDLister struct {
	OfficialAPIClient *OfficialAPIClient
	Logger            *Logger
}

func (c *CreatorIDLister) Do(ctx context.Context, in *CreatorIDListerDoInput) ([]string, error) {
	all, err := c.all(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("list all creator IDs: %w", err)
	}

	ignoreMap := map[string]interface{}{}
	for _, id := range in.IgnoreCreatorIDs {
		ignoreMap[id] = nil
	}
	res := make([]string, 0, len(all))
	for _, id := range all {
		if _, ok := ignoreMap[id]; ok {
			continue
		}
		res = append(res, id)
	}
	return res, nil
}

func (c *CreatorIDLister) all(ctx context.Context, in *CreatorIDListerDoInput) ([]string, error) {
	if len(in.InputCreatorIDs) > 0 {
		c.Logger.Infof("Use input creator IDs: %v", in.InputCreatorIDs)
		return in.InputCreatorIDs, nil
	}

	ids := map[string]interface{}{}

	if in.IncludeSupporting {
		plans := PlanListSupportingResponse{}
		err := c.OfficialAPIClient.RequestAndUnwrapJSON(ctx, http.MethodGet, "https://api.fanbox.cc/plan.listSupporting", &plans)
		if err != nil {
			return nil, fmt.Errorf("list supporintg plans: %w", err)
		}
		for _, p := range plans.Body {
			ids[p.CreatorID] = nil
		}
	}

	if in.IncludeFollowing {
		following := PlanListSupportingResponse{}
		err := c.OfficialAPIClient.RequestAndUnwrapJSON(ctx, http.MethodGet, "https://api.fanbox.cc/creator.listFollowing", &following)
		if err != nil {
			return nil, fmt.Errorf("list following creators: %w", err)
		}
		for _, f := range following.Body {
			ids[f.CreatorID] = nil
		}
	}

	res := make([]string, 0, len(ids))
	for id := range ids {
		res = append(res, id)
	}
	return res, nil
}
