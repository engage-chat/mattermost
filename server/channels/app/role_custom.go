// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/request"
)

func (a *App) GetCustomRolesForGroup(c request.CTX, group string) ([]*model.Role, *model.AppError) {
	// ダミーロールを返す
    if group == "" {
        return []*model.Role{
            {Id: model.NewId(), Name: "tunag_custom_role_1", DisplayName: "Custom Role 1", Permissions: []string{"read_channel"}},
            {Id: model.NewId(), Name: "tunag_custom_role_2", DisplayName: "Custom Role 2", Permissions: []string{"create_post"}},
        }, nil
    } else if group == "example_group" {
        return []*model.Role{
            {Id: model.NewId(), Name: "tunag_custom_role_3", DisplayName: "Custom Role 1", Permissions: []string{"read_channel"}},
            {Id: model.NewId(), Name: "tunag_custom_role_4", DisplayName: "Custom Role 2", Permissions: []string{"create_post"}},
        }, nil
	}
    return []*model.Role{}, nil
}

func (a *App) EnableCustomRoles(c request.CTX, group string) ([]*model.Role, *model.AppError) {
    // 成功したと仮定して、有効化されたロール情報を返す
    return a.GetCustomRolesForGroup(c, group)
}

func (a *App) DisableCustomRoles(c request.CTX, group string) *model.AppError {
    return nil
}
