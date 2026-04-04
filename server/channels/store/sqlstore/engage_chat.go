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
	hasSystemRoles := len(options.SystemRoles) > 0
	hasTeamRoles := len(options.TeamRoles) > 0

	if !hasSystemRoles && !hasTeamRoles {
		return false, nil
	}

	isPostgreSQL := s.DriverName() == model.DatabaseDriverPostgres

	orConditions := sq.Or{}
	if hasSystemRoles {
		for _, role := range options.SystemRoles {
			if isPostgreSQL {
				orConditions = append(orConditions, sq.Expr("? = ANY(string_to_array(u.Roles, ' '))", role))
			} else {
				orConditions = append(orConditions, sq.Expr("FIND_IN_SET(?, REPLACE(u.Roles, ' ', ','))", role))
			}
		}
	}
	if hasTeamRoles {
		for _, role := range options.TeamRoles {
			if isPostgreSQL {
				orConditions = append(orConditions, sq.Expr("? = ANY(string_to_array(tm.Roles, ' '))", role))
			} else {
				orConditions = append(orConditions, sq.Expr("FIND_IN_SET(?, REPLACE(tm.Roles, ' ', ','))", role))
			}
		}
	}

	query := s.getQueryBuilder().Select("1").
		From("ChannelMembers cm")

	// Only JOIN the tables that are actually needed for the search criteria.
	if hasSystemRoles {
		query = query.Join("Users u ON cm.UserId = u.Id")
	}
	if hasTeamRoles {
		query = query.
			Join("Channels c ON cm.ChannelId = c.Id").
			Join("TeamMembers tm ON tm.TeamId = c.TeamId AND tm.UserId = cm.UserId")
	}

	query = query.
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
