// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetAllTunagCustomRoles returns a list of all the custom roles for tunag.
func (c *Client4) GetAllTunagCustomRoles(ctx context.Context) ([]*Role, *Response, error) {
	r, err := c.DoAPIGet(ctx, "/tunag_custom/roles", "")
	if err != nil {
		return nil, BuildResponse(r), err
	}
	defer closeBody(r)
	var roles []*Role
	if err := json.NewDecoder(r.Body).Decode(&roles); err != nil {
		return nil, nil, NewAppError("GetAllTunagCustomRoles", "api.unmarshal_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}
	return roles, BuildResponse(r), nil
}

// GetTunagCustomRoles returns a list of custom roles for a specific group for tunag.
func (c *Client4) GetTunagCustomRoles(ctx context.Context, group string) ([]*Role, *Response, error) {
	r, err := c.DoAPIGet(ctx, fmt.Sprintf("/tunag_custom/roles/%s", group), "")
	if err != nil {
		return nil, BuildResponse(r), err
	}
	defer closeBody(r)
	var roles []*Role
	if err := json.NewDecoder(r.Body).Decode(&roles); err != nil {
		return nil, nil, NewAppError("GetTunagCustomRoles", "api.unmarshal_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}
	return roles, BuildResponse(r), nil
}

// EnableTunagCustomRoles enables and returns a list of custom roles for a specific group for tunag.
func (c *Client4) EnableTunagCustomRoles(ctx context.Context, group string) ([]*Role, *Response, error) {
	r, err := c.DoAPIPost(ctx, fmt.Sprintf("/tunag_custom/roles/enable/%s", group), "")
	if err != nil {
		return nil, BuildResponse(r), err
	}
	defer closeBody(r)
	var roles []*Role
	if err := json.NewDecoder(r.Body).Decode(&roles); err != nil {
		return nil, nil, NewAppError("EnableTunagCustomRoles", "api.unmarshal_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}
	return roles, BuildResponse(r), nil
}

// DisableTunagCustomRoles disables custom roles for a specific group for tunag.
func (c *Client4) DisableTunagCustomRoles(ctx context.Context, group string) (*Response, error) {
	r, err := c.DoAPIDelete(ctx, fmt.Sprintf("/tunag_custom/roles/disable/%s", group))
	if err != nil {
		return BuildResponse(r), err
	}
	defer closeBody(r)
	return BuildResponse(r), nil
}
