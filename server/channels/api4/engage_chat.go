// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
)

func (api *API) InitEngageChat() {
	api.BaseRoutes.EngageChat.Handle("/roles", api.APISessionRequired(enableCustomRoles)).Methods(http.MethodPost)
}

func enableCustomRoles(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionManageRoles) {
		c.SetPermissionError(model.PermissionManageRoles)
		return
	}

	roleNames, err := model.SortedArrayFromJSON(r.Body)
	if err != nil {
		c.SetInvalidParamWithErr("role_names", err)
		return
	}

	enabledRoles, appErr := c.App.EnableCustomRoles(c.AppContext, roleNames)
	if appErr != nil {
		c.Err = appErr
		return
	}

	js, err := json.Marshal(enabledRoles)
	if err != nil {
		c.Err = model.NewAppError("enableCustomRoles", "api.marshal_error", nil, "", http.StatusInternalServerError).Wrap(err)
		return
	}

	if _, err := w.Write(js); err != nil {
		c.Logger.Warn("Error while writing response", mlog.Err(err))
		return
	}
}
