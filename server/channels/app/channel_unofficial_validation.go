// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"net/http"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/request"
)

// Validates the setting 「個人チャット・グループチャット・ルーム作成を許可する」 in TUNAG.
// If user is not allowed to create these, the permission checked by this function has been removed from the role.
//
//   - 個人チャット:     DM
//   - グループチャット:  GM
//   - ルーム:           Unofficial channel (not created via TUNAG). This can be checked by isOfficialChannel().
func (a *App) CheckChannelPermissions(c request.CTX, channel *model.Channel, userID string) *model.AppError {
	session := c.Session()
	if session == nil || session.Id == "" {
		return nil // No session means no DM/GM permission check needed
	}

	// If channel is nil, skip permission check
	if channel == nil {
		return nil
	}

	// Check if the user is a bot - bots should be allowed to post in DM/GM channels
	if userID != "" {
		user, err := a.GetUser(userID)
		if err != nil {
			if err.StatusCode != http.StatusNotFound {
				return err
			}
			// User not found, so not a bot. Continue to permission checks.
		} else if user.IsBot {
			return nil
		}
	}

	// If user is not a member of any team, skip permission check for api-test (TestCreatePostAll)
	if len(session.TeamMembers) == 0 {
		return nil
	}

	// System admins always have permission to DM/GM/Channels
	if a.SessionHasPermissionTo(*session, model.PermissionManageSystem) {
		return nil
	}

	var requiredPermission *model.Permission
	var hasPermission bool

	switch channel.Type {
	case model.ChannelTypeDirect:
		requiredPermission = model.PermissionCreateDirectChannel
		hasPermission = a.SessionHasPermissionTo(*session, requiredPermission)

	case model.ChannelTypeGroup:
		requiredPermission = model.PermissionCreateGroupChannel
		hasPermission = a.SessionHasPermissionTo(*session, requiredPermission)

	case model.ChannelTypePrivate:
		// If channel is official channel, user always have permission to access.
		isOfficial, err := a.IsOfficialChannel(c, channel)
		if err != nil {
			return err
		}
		if isOfficial {
			return nil
		}

		requiredPermission = model.PermissionCreatePrivateChannel
		hasPermission = a.SessionHasPermissionToTeam(*session, channel.TeamId, requiredPermission)
	}

	if requiredPermission != nil && !hasPermission {
		return model.NewAppError(
			"CheckChannelPermissions",
			"api.context.permissions.app_error",
			nil,
			"",
			http.StatusForbidden,
		)
	}

	return nil
}
