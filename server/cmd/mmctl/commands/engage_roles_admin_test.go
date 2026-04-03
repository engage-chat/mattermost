// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost/server/v8/cmd/mmctl/printer"
)

func (s *MmctlUnitTestSuite) TestEngageAdminCmd() {
	s.Run("Enable engage admin role successfully", func() {
		printer.Clean()

		s.client.
			EXPECT().
			EnableCustomRoles(context.TODO(), []string{model.SystemEngageAdmin}).
			Return([]*model.Role{{Name: model.SystemEngageAdmin}}, &model.Response{StatusCode: http.StatusOK}, nil).
			Times(1)

		err := rolesEngageAdminCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().Nil(err)

		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Equal(fmt.Sprintf("Role %q enabled successfully.", model.SystemEngageAdmin), printer.GetLines()[0])
	})

	s.Run("Fail to enable engage admin role", func() {
		printer.Clean()

		s.client.
			EXPECT().
			EnableCustomRoles(context.TODO(), []string{model.SystemEngageAdmin}).
			Return(nil, &model.Response{StatusCode: http.StatusInternalServerError}, errors.New("mock error")).
			Times(1)

		err := rolesEngageAdminCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().ErrorContains(err, fmt.Sprintf("unable to enable %q role", model.SystemEngageAdmin))

		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

