// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	sq "github.com/mattermost/squirrel"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/v8/channels/store"
)

type SqlEngageChatStore struct {
	*SqlStore
}

func newSqlEngageChatStore(sqlStore *SqlStore) store.EngageChatStore {
	s := &SqlEngageChatStore{
		SqlStore: sqlStore,
	}
	return s
}

func (s *SqlEngageChatStore) HasChannelMemberWithRoles(channelID string, options *model.EngageChatRoleSearchOptions) (bool, error) {
	if options == nil {
		return false, nil
	}

	hasSystemRoles := len(options.SystemRoles) > 0
	hasTeamRoles := len(options.TeamRoles) > 0

	if !hasSystemRoles && !hasTeamRoles {
		return false, nil
	}

	// Build role matching conditions using LIKE patterns.
	// Roles are stored as space-separated strings (e.g., "system_user system_engage_admin").
	// The pattern "% role %" with padding handles all positions:
	//   - beginning: " system_engage_admin " matches "system_engage_admin other"
	//   - middle:    " system_engage_admin " matches "other system_engage_admin another"
	//   - end:       " system_engage_admin " matches "other system_engage_admin"
	//   - sole:      " system_engage_admin " matches "system_engage_admin"
	// By prepending/appending a space to the column, we normalize all cases.
	orConditions := sq.Or{}
	if hasSystemRoles {
		for _, role := range options.SystemRoles {
			orConditions = append(orConditions, sq.Expr("CONCAT(' ', u.Roles, ' ') LIKE ?", "% "+role+" %"))
		}
	}
	if hasTeamRoles {
		for _, role := range options.TeamRoles {
			orConditions = append(orConditions, sq.Expr("CONCAT(' ', tm.Roles, ' ') LIKE ?", "% "+role+" %"))
		}
	}

	subQuery := s.getQueryBuilder().Select("1").
		From("ChannelMembers cm")

	// Only JOIN the tables that are actually needed for the search criteria.
	if hasSystemRoles {
		subQuery = subQuery.Join("Users u ON cm.UserId = u.Id AND u.DeleteAt = 0")
	}
	if hasTeamRoles {
		subQuery = subQuery.
			Join("Channels c ON cm.ChannelId = c.Id").
			Join("TeamMembers tm ON tm.TeamId = c.TeamId AND tm.UserId = cm.UserId AND tm.DeleteAt = 0")
	}

	subQuery = subQuery.
		Where(sq.Eq{"cm.ChannelId": channelID}).
		Where(orConditions).
		Limit(1)

	subQueryString, args, err := subQuery.ToSql()
	if err != nil {
		return false, errors.Wrap(err, "failed to build query")
	}

	// Wrap in EXISTS so we always get a single boolean result (no ErrNoRows handling needed).
	existsQuery := "SELECT EXISTS(" + subQueryString + ")"

	var exists bool
	err = s.GetReplica().Get(&exists, existsQuery, args...)
	if err != nil {
		return false, errors.Wrap(err, "failed to query for channel member with role")
	}

	return exists, nil
}
