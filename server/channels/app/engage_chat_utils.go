// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"net/http"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/request"
)

// IsChannelAccessible checks if a user has permissions to post and react in a given channel
// based on the full validation logic including DM/GM restrictions and exceptions.
// It returns true if accessible, false if not.
func (a *App) IsChannelAccessible(c request.CTX, channelID string, userID string) (bool, *model.AppError) {
	channel, err := a.GetChannel(c, channelID)
	if err != nil {
		return false, err
	}

	user, err := a.GetUser(userID)
	if err != nil {
		return false, err
	}
	if user.IsBot {
		return true, nil
	}

	// 1. Check if the channel is offically created
	isOfficial, err := a.IsOfficialChannel(c, channel)
	if err != nil {
		return false, err
	}
	if isOfficial {
		return true, nil
	}

	// 2. Check if the situation is under restriction on unofficial channel (by checking the user themselves has the required permission)
	var requiredPermission *model.Permission
	var hasPermission bool
	switch channel.Type {
	case model.ChannelTypeDirect:
		requiredPermission = model.PermissionCreateDirectChannel
		hasPermission = a.SessionHasPermissionTo(*c.Session(), requiredPermission)
	case model.ChannelTypeGroup:
		requiredPermission = model.PermissionCreateGroupChannel
		hasPermission = a.SessionHasPermissionTo(*c.Session(), requiredPermission)
	case model.ChannelTypePrivate:
		requiredPermission = model.PermissionCreatePrivateChannel
		hasPermission = a.SessionHasPermissionToTeam(*c.Session(), channel.TeamId, requiredPermission)
	default:
		// For other channel types (e.g., public), access is not restricted by this logic.
		return true, nil
	}

	if hasPermission {
		return true, nil
	}

	// 3. Check for the exception in restriction (by checking any member in the channel has one of the exception roles)
	searchOpts := &model.EngageChatRoleSearchOptions{}
	switch channel.Type {
	case model.ChannelTypeDirect, model.ChannelTypeGroup:
		searchOpts.SystemRoles = []string{model.SystemEngageAdmin}
	case model.ChannelTypePrivate:
		searchOpts.TeamRoles = []string{model.TeamEngageAdmin}
	default:
		return false, nil
	}

	hasMemberWithRole, hasRoleErr := a.Srv().Store().EngageChat().HasChannelMemberWithRoles(channelID, searchOpts)
	if hasRoleErr != nil {
		return false, model.NewAppError("IsChannelAccessible", "app.channel.has_channel_member_with_role.app_error", nil, "", http.StatusInternalServerError).Wrap(hasRoleErr)
	}
	if hasMemberWithRole {
		return true, nil // Exception found, channel is accessible.
	}

	// No permissions and no exceptions found. The channel is not accessible.
	return false, nil
}
