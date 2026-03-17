// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
)

func (api *API) InitTunagCustom() {
	TunagCustom := api.BaseRoutes.APIRoot.PathPrefix("/tunag_custom").Subrouter()

	TunagCustom.Handle("/roles", api.APISessionRequired(getAllTunagCustomRoles)).Methods(http.MethodGet)
	TunagCustom.Handle("/roles/{role_group:[a-z0-9_]+}", api.APISessionRequired(getTunagCustomRoles)).Methods(http.MethodGet)
	TunagCustom.Handle("/roles/enable/{role_group:[a-z0-9_]+}", api.APISessionRequired(enableTunagCustomRoles)).Methods(http.MethodPost)
	TunagCustom.Handle("/roles/disable/{role_group:[a-z0-9_]+}", api.APISessionRequired(disableTunagCustomRoles)).Methods(http.MethodDelete)
}

func getAllTunagCustomRoles(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionManageRoles) {
		c.SetPermissionError(model.PermissionManageRoles)
		return
	}

	roles, appErr := c.App.GetCustomRolesForGroup(c.AppContext, "")
	if appErr != nil {
		c.Err = appErr
		return
	}

	js, err := json.Marshal(roles)
	if err != nil {
		c.Err = model.NewAppError("getAllTunagCustomRoles", "api.marshal_error", nil, "", http.StatusInternalServerError).Wrap(err)
		return
	}

	if _, err := w.Write(js); err != nil {
		c.Logger.Warn("Error while writing response", mlog.Err(err))
		return
	}
}

func getTunagCustomRoles(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionManageRoles) {
		c.SetPermissionError(model.PermissionManageRoles)
		return
	}

	params := mux.Vars(r)
	customRoleGroup, ok := params["role_group"]
	if !ok {
		c.SetInvalidParam("role_group")
		return
	}

	roles, appErr := c.App.GetCustomRolesForGroup(c.AppContext, customRoleGroup)
	if appErr != nil {
		c.Err = appErr
		return
	}

	js, err := json.Marshal(roles)
	if err != nil {
		c.Err = model.NewAppError("getTunagCustomRoles", "api.marshal_error", nil, "", http.StatusInternalServerError).Wrap(err)
		return
	}

	if _, err := w.Write(js); err != nil {
		c.Logger.Warn("Error while writing response", mlog.Err(err))
		return
	}
}

func enableTunagCustomRoles(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionManageRoles) {
		c.SetPermissionError(model.PermissionManageRoles)
		return
	}

	params := mux.Vars(r)
	customRoleGroup, ok := params["role_group"]
	if !ok {
		c.SetInvalidParam("role_group")
		return
	}

	enabledRoles, appErr := c.App.EnableCustomRoles(c.AppContext, customRoleGroup)
	if appErr != nil {
		c.Err = appErr
		return
	}

	js, err := json.Marshal(enabledRoles)
	if err != nil {
		c.Err = model.NewAppError("enableTunagCustomRoles", "api.marshal_error", nil, "", http.StatusInternalServerError).Wrap(err)
		return
	}

	if _, err := w.Write(js); err != nil {
		c.Logger.Warn("Error while writing response", mlog.Err(err))
		return
	}
}

func disableTunagCustomRoles(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionManageRoles) {
		c.SetPermissionError(model.PermissionManageRoles)
		return
	}

	params := mux.Vars(r)
	customRoleGroup, ok := params["role_group"]
	if !ok {
		c.SetInvalidParam("role_group")
		return
	}

	appErr := c.App.DisableCustomRoles(c.AppContext, customRoleGroup)
	if appErr != nil {
		c.Err = appErr
		return
	}

	ReturnStatusOK(w)
}
