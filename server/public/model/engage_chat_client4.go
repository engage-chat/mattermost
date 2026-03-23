// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"context"
	"encoding/json"
	"net/http"
)

// EnableCustomRoles enables and returns a list of custom roles specified by name.
func (c *Client4) EnableCustomRoles(ctx context.Context, roleNames []string) ([]*Role, *Response, error) {
	jsonData, err := json.Marshal(roleNames)
	if err != nil {
		return nil, nil, NewAppError("EnableCustomRoles", "api.marshal_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	r, err := c.DoAPIPost(ctx, "/engage_chat/roles", string(jsonData))
	if err != nil {
		return nil, BuildResponse(r), err
	}
	defer closeBody(r)

	var roles []*Role
	if err := json.NewDecoder(r.Body).Decode(&roles); err != nil {
		return nil, nil, NewAppError("EnableCustomRoles", "api.unmarshal_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}
	return roles, BuildResponse(r), nil
}
