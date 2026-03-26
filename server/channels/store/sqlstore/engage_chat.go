// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"database/sql"

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
	if len(options.SystemRoles) == 0 && len(options.TeamRoles) == 0 {
		return false, nil
	}

	orConditions := sq.Or{}
	if len(options.SystemRoles) > 0 {
		for _, role := range options.SystemRoles {
			orConditions = append(orConditions, sq.Or{
				sq.Eq{"u.Roles": role},
				sq.Like{"u.Roles": role + " %"},
				sq.Like{"u.Roles": "% " + role},
				sq.Like{"u.Roles": "% " + role + " %"},
			})
		}
	}
	if len(options.TeamRoles) > 0 {
		for _, role := range options.TeamRoles {
			orConditions = append(orConditions, sq.Or{
				sq.Eq{"tm.Roles": role},
				sq.Like{"tm.Roles": role + " %"},
				sq.Like{"tm.Roles": "% " + role},
				sq.Like{"tm.Roles": "% " + role + " %"},
			})
		}
	}

	query := s.getQueryBuilder().Select("1").
		From("ChannelMembers cm").
		Join("Users u ON cm.UserId = u.Id").
		Join("Channels c ON cm.ChannelId = c.Id").
		LeftJoin("TeamMembers tm ON tm.TeamId = c.TeamId AND tm.UserId = cm.UserId").
		Where(sq.Eq{"cm.ChannelId": channelID}).
		Where(orConditions).
		Limit(1)

	queryString, args, err := query.ToSql()
	if err != nil {
		return false, errors.Wrap(err, "failed to build query")
	}

	var result int
	err = s.GetReplica().Get(&result, queryString, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, errors.Wrap(err, "failed to query for channel member with role")
	}

	return result == 1, nil
}
