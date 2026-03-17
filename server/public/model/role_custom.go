// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

// カスタムロールのグループ
const (
	CustomRolesUnofficial = "unofficial"
	CustomRolesTest       = "test_group"
)

// カスタムロール
const (
	SystemTunagAdmin = "system_tunag_admin"
	TeamTunagAdmin   = "team_tunag_admin"

	TestTunagAdmin = "test_tunag_admin"
)

/*
グループ名をKeyとし、そのグループのカスタムロールマップを返す関数をValueとして保持する変数。
全てのカスタムロールグループ定義のマスターリストとなる。

	Key: グループ名
	Value: グループのカスタムロールマップを返す関数
*/
var customRoleGroupFactories = map[string]func() map[string]Role{
	CustomRolesUnofficial: makeTunagCustomRolesUnofficial,
	CustomRolesTest:       makeTunagCustomRolesTest,
}

func AllCustomRoleGroups() []string {
	groups := make([]string, 0, len(customRoleGroupFactories))
	for groupName := range customRoleGroupFactories {
		groups = append(groups, groupName)
	}
	return groups
}

func CustomRoleNamesForGroup(customRoleGroup string) []string {
	roles := MakeTunagCustomRoles(customRoleGroup)
	if roles == nil {
		return nil
	}

	names := make([]string, 0, len(roles))
	for roleKey := range roles {
		names = append(names, roleKey)
	}
	return names
}

/*
指定されたグループに所属するカスタムロールのMapを返す関数。
※varで宣言するとPermissionがinit()される前に初期化されてエラーになるため、関数で作成する
*/
func MakeTunagCustomRoles(customRoleGroup string) map[string]Role {
	if factory, ok := customRoleGroupFactories[customRoleGroup]; ok {
		return factory()
	}
	return nil
}

func makeTunagCustomRolesTest() map[string]Role {
	return map[string]Role{
		TestTunagAdmin: {
			Name:        TestTunagAdmin,
			DisplayName: TestTunagAdmin,
			Description: "",
			Permissions: []string{
				PermissionCreateBot.Id,
			},
		},
	}
}

func makeTunagCustomRolesUnofficial() map[string]Role {
	return map[string]Role{
		SystemTunagAdmin: {
			Name:        SystemTunagAdmin,
			DisplayName: SystemTunagAdmin,
			Description: "",
			Permissions: []string{
				PermissionCreatePrivateChannel.Id,
			},
		},
		TeamTunagAdmin: {
			Name:        TeamTunagAdmin,
			DisplayName: TeamTunagAdmin,
			Description: "",
			Permissions: []string{
				PermissionCreateDirectChannel.Id,
				PermissionCreateGroupChannel.Id,
			},
		},
	}
}
