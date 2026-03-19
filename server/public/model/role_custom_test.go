// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMakeAllCustomRoleTemplates(t *testing.T) {
	templates := MakeAllCustomRoleTemplates()
	require.NotNil(t, templates)
	require.Len(t, templates, len(allCustomRoleNames))

	// Check for SystemEngageAdmin
	systemRole, ok := templates[SystemEngageAdmin]
	require.True(t, ok)
	require.Equal(t, SystemEngageAdmin, systemRole.Name)
	require.Equal(t, SystemEngageAdmin, systemRole.DisplayName)
	require.Equal(t, []string{PermissionCreatePrivateChannel.Id}, systemRole.Permissions)

	// Check for TeamEngageAdmin
	teamRole, ok := templates[TeamEngageAdmin]
	require.True(t, ok)
	require.Equal(t, TeamEngageAdmin, teamRole.Name)
	require.Equal(t, TeamEngageAdmin, teamRole.DisplayName)
	require.Equal(t, []string{PermissionCreateDirectChannel.Id, PermissionCreateGroupChannel.Id}, teamRole.Permissions)
}
