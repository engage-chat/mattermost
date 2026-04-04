// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package storetest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/request"
	"github.com/mattermost/mattermost/server/v8/channels/store"
)

func TestEngageChatStore(t *testing.T, rctx request.CTX, ss store.Store, s SqlStore) {
	t.Run("HasChannelMemberWithRoles", func(t *testing.T) {
		t.Run("EmptyOptions", func(t *testing.T) { testHasChannelMemberWithRolesEmptyOptions(t, rctx, ss) })
		t.Run("SystemRolesMatch", func(t *testing.T) { testHasChannelMemberWithRolesSystemRolesMatch(t, rctx, ss) })
		t.Run("SystemRolesNoMatch", func(t *testing.T) { testHasChannelMemberWithRolesSystemRolesNoMatch(t, rctx, ss) })
		t.Run("TeamRolesMatch", func(t *testing.T) { testHasChannelMemberWithRolesTeamRolesMatch(t, rctx, ss) })
		t.Run("TeamRolesNoMatch", func(t *testing.T) { testHasChannelMemberWithRolesTeamRolesNoMatch(t, rctx, ss) })
		t.Run("NonexistentChannel", func(t *testing.T) { testHasChannelMemberWithRolesNonexistentChannel(t, rctx, ss) })
		t.Run("DirectChannel", func(t *testing.T) { testHasChannelMemberWithRolesDirectChannel(t, rctx, ss) })
		t.Run("MultipleMembers", func(t *testing.T) { testHasChannelMemberWithRolesMultipleMembers(t, rctx, ss) })
	})
}

// engageChatCreateTeam creates a team for test use.
func engageChatCreateTeam(t *testing.T, ss store.Store) *model.Team {
	t.Helper()
	team, err := ss.Team().Save(&model.Team{
		DisplayName: "Test Team",
		Name:        NewTestID(),
		Email:       MakeEmail(),
		Type:        model.TeamOpen,
	})
	require.NoError(t, err)
	return team
}

// engageChatCreateUser creates a user with the given system roles.
func engageChatCreateUser(t *testing.T, rctx request.CTX, ss store.Store, roles string) *model.User {
	t.Helper()
	user := &model.User{
		Email:    MakeEmail(),
		Username: "u_" + model.NewId(),
		Roles:    roles,
	}
	user, err := ss.User().Save(rctx, user)
	require.NoError(t, err)
	return user
}

// engageChatCreateChannel creates a channel with the given type on the given team.
func engageChatCreateChannel(t *testing.T, rctx request.CTX, ss store.Store, teamID string, channelType model.ChannelType) *model.Channel {
	t.Helper()
	channel, err := ss.Channel().Save(rctx, &model.Channel{
		TeamId:      teamID,
		DisplayName: "Test Channel",
		Name:        NewTestID(),
		Type:        channelType,
	}, -1)
	require.NoError(t, err)
	return channel
}

// engageChatAddChannelMember adds the user as a member of the channel.
func engageChatAddChannelMember(t *testing.T, rctx request.CTX, ss store.Store, channelID, userID string) {
	t.Helper()
	_, err := ss.Channel().SaveMember(rctx, &model.ChannelMember{
		ChannelId:   channelID,
		UserId:      userID,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
		SchemeUser:  true,
	})
	require.NoError(t, err)
}

// engageChatAddTeamMember adds the user as a team member with the given explicit roles.
func engageChatAddTeamMember(t *testing.T, rctx request.CTX, ss store.Store, teamID, userID, explicitRoles string) {
	t.Helper()
	_, err := ss.Team().SaveMember(rctx, &model.TeamMember{
		TeamId:        teamID,
		UserId:        userID,
		SchemeUser:    true,
		ExplicitRoles: explicitRoles,
	}, -1)
	require.NoError(t, err)
}

// --- Test cases ---

func testHasChannelMemberWithRolesEmptyOptions(t *testing.T, rctx request.CTX, ss store.Store) {
	// Both SystemRoles and TeamRoles are empty → should return false immediately without DB query.
	result, err := ss.EngageChat().HasChannelMemberWithRoles(model.NewId(), &model.EngageChatRoleSearchOptions{})
	require.NoError(t, err)
	assert.False(t, result)
}

func testHasChannelMemberWithRolesSystemRolesMatch(t *testing.T, rctx request.CTX, ss store.Store) {
	team := engageChatCreateTeam(t, ss)

	// Create a user with system_engage_admin role.
	user := engageChatCreateUser(t, rctx, ss, model.SystemUserRoleId+" "+model.SystemEngageAdmin)
	defer func() { require.NoError(t, ss.User().PermanentDelete(rctx, user.Id)) }()

	// Create a private channel and add the user as a member.
	channel := engageChatCreateChannel(t, rctx, ss, team.Id, model.ChannelTypePrivate)
	engageChatAddTeamMember(t, rctx, ss, team.Id, user.Id, "")
	engageChatAddChannelMember(t, rctx, ss, channel.Id, user.Id)

	// Search for SystemEngageAdmin → should find the user.
	result, err := ss.EngageChat().HasChannelMemberWithRoles(channel.Id, &model.EngageChatRoleSearchOptions{
		SystemRoles: []string{model.SystemEngageAdmin},
	})
	require.NoError(t, err)
	assert.True(t, result)
}

func testHasChannelMemberWithRolesSystemRolesNoMatch(t *testing.T, rctx request.CTX, ss store.Store) {
	team := engageChatCreateTeam(t, ss)

	// Create a regular user WITHOUT system_engage_admin.
	user := engageChatCreateUser(t, rctx, ss, model.SystemUserRoleId)
	defer func() { require.NoError(t, ss.User().PermanentDelete(rctx, user.Id)) }()

	channel := engageChatCreateChannel(t, rctx, ss, team.Id, model.ChannelTypePrivate)
	engageChatAddTeamMember(t, rctx, ss, team.Id, user.Id, "")
	engageChatAddChannelMember(t, rctx, ss, channel.Id, user.Id)

	// Search for SystemEngageAdmin → should NOT find any match.
	result, err := ss.EngageChat().HasChannelMemberWithRoles(channel.Id, &model.EngageChatRoleSearchOptions{
		SystemRoles: []string{model.SystemEngageAdmin},
	})
	require.NoError(t, err)
	assert.False(t, result)
}

func testHasChannelMemberWithRolesTeamRolesMatch(t *testing.T, rctx request.CTX, ss store.Store) {
	team := engageChatCreateTeam(t, ss)

	user := engageChatCreateUser(t, rctx, ss, model.SystemUserRoleId)
	defer func() { require.NoError(t, ss.User().PermanentDelete(rctx, user.Id)) }()

	// Add the user as a team member with team_engage_admin explicit role.
	engageChatAddTeamMember(t, rctx, ss, team.Id, user.Id, model.TeamEngageAdmin)

	channel := engageChatCreateChannel(t, rctx, ss, team.Id, model.ChannelTypePrivate)
	engageChatAddChannelMember(t, rctx, ss, channel.Id, user.Id)

	// Search for TeamEngageAdmin → should find the user.
	result, err := ss.EngageChat().HasChannelMemberWithRoles(channel.Id, &model.EngageChatRoleSearchOptions{
		TeamRoles: []string{model.TeamEngageAdmin},
	})
	require.NoError(t, err)
	assert.True(t, result)
}

func testHasChannelMemberWithRolesTeamRolesNoMatch(t *testing.T, rctx request.CTX, ss store.Store) {
	team := engageChatCreateTeam(t, ss)

	user := engageChatCreateUser(t, rctx, ss, model.SystemUserRoleId)
	defer func() { require.NoError(t, ss.User().PermanentDelete(rctx, user.Id)) }()

	// Add as team member with NO custom roles.
	engageChatAddTeamMember(t, rctx, ss, team.Id, user.Id, "")

	channel := engageChatCreateChannel(t, rctx, ss, team.Id, model.ChannelTypePrivate)
	engageChatAddChannelMember(t, rctx, ss, channel.Id, user.Id)

	// Search for TeamEngageAdmin → should NOT find any match.
	result, err := ss.EngageChat().HasChannelMemberWithRoles(channel.Id, &model.EngageChatRoleSearchOptions{
		TeamRoles: []string{model.TeamEngageAdmin},
	})
	require.NoError(t, err)
	assert.False(t, result)
}

func testHasChannelMemberWithRolesNonexistentChannel(t *testing.T, rctx request.CTX, ss store.Store) {
	// A non-existent channel ID → should return false with no error.
	result, err := ss.EngageChat().HasChannelMemberWithRoles(model.NewId(), &model.EngageChatRoleSearchOptions{
		SystemRoles: []string{model.SystemEngageAdmin},
	})
	require.NoError(t, err)
	assert.False(t, result)
}

func testHasChannelMemberWithRolesDirectChannel(t *testing.T, rctx request.CTX, ss store.Store) {
	// Create two users — one with system_engage_admin, one without.
	user1 := engageChatCreateUser(t, rctx, ss, model.SystemUserRoleId+" "+model.SystemEngageAdmin)
	defer func() { require.NoError(t, ss.User().PermanentDelete(rctx, user1.Id)) }()

	user2 := engageChatCreateUser(t, rctx, ss, model.SystemUserRoleId)
	defer func() { require.NoError(t, ss.User().PermanentDelete(rctx, user2.Id)) }()

	// Create a DM channel between the two users.
	dm := model.Channel{
		DisplayName: "DM",
		Name:        model.GetDMNameFromIds(user1.Id, user2.Id),
		Type:        model.ChannelTypeDirect,
	}
	m1 := model.ChannelMember{UserId: user1.Id, NotifyProps: model.GetDefaultChannelNotifyProps()}
	m2 := model.ChannelMember{UserId: user2.Id, NotifyProps: model.GetDefaultChannelNotifyProps()}

	savedChannel, err := ss.Channel().SaveDirectChannel(rctx, &dm, &m1, &m2)
	require.NoError(t, err)

	// Search for SystemEngageAdmin on the DM channel → should find user1.
	result, hErr := ss.EngageChat().HasChannelMemberWithRoles(savedChannel.Id, &model.EngageChatRoleSearchOptions{
		SystemRoles: []string{model.SystemEngageAdmin},
	})
	require.NoError(t, hErr)
	assert.True(t, result)
}

func testHasChannelMemberWithRolesMultipleMembers(t *testing.T, rctx request.CTX, ss store.Store) {
	// Verify that if only one of multiple members has the role, it still returns true.
	team := engageChatCreateTeam(t, ss)

	userWithRole := engageChatCreateUser(t, rctx, ss, model.SystemUserRoleId+" "+model.SystemEngageAdmin)
	defer func() { require.NoError(t, ss.User().PermanentDelete(rctx, userWithRole.Id)) }()

	userWithoutRole := engageChatCreateUser(t, rctx, ss, model.SystemUserRoleId)
	defer func() { require.NoError(t, ss.User().PermanentDelete(rctx, userWithoutRole.Id)) }()

	channel := engageChatCreateChannel(t, rctx, ss, team.Id, model.ChannelTypePrivate)

	engageChatAddTeamMember(t, rctx, ss, team.Id, userWithRole.Id, "")
	engageChatAddTeamMember(t, rctx, ss, team.Id, userWithoutRole.Id, "")
	engageChatAddChannelMember(t, rctx, ss, channel.Id, userWithRole.Id)
	engageChatAddChannelMember(t, rctx, ss, channel.Id, userWithoutRole.Id)

	// Should find the member with the role.
	result, err := ss.EngageChat().HasChannelMemberWithRoles(channel.Id, &model.EngageChatRoleSearchOptions{
		SystemRoles: []string{model.SystemEngageAdmin},
	})
	require.NoError(t, err)
	assert.True(t, result)
}
