package legacy

import (
	"context"
	"fmt"
	"time"

	"github.com/grafana/grafana/pkg/storage/legacysql"
	"github.com/grafana/grafana/pkg/storage/unified/sql/sqltemplate"
)

type DashboardStars struct {
	OrgID   int64
	UserUID string
	First   int64
	Last    int64

	Dashboards []string
}

// NOTE: this does not support paging
type LegacyStarSQL struct {
	DB legacysql.LegacyDatabaseProvider
}

func (s *LegacyStarSQL) GetStars(ctx context.Context, orgId int64, user string) ([]DashboardStars, int64, error) {
	rv := int64(0)

	sql, err := s.DB(ctx)
	if err != nil {
		return nil, rv, err
	}

	req := newQueryReq(sql, user, orgId)

	tmpl := sqlQueryStars
	rawQuery, err := sqltemplate.Execute(tmpl, req)
	if err != nil {
		return nil, 0, fmt.Errorf("execute template %q: %w", tmpl.Name(), err)
	}
	q := rawQuery
	// if false {
	// 	 pretty := sqltemplate.RemoveEmptyLines(rawQuery)
	//	 fmt.Printf("DASHBOARD QUERY: %s [%+v] // %+v\n", pretty, req.GetArgs(), query)
	// }

	rows, err := sql.DB.GetSqlxSession().Query(ctx, q, req.GetArgs()...)
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()

	stars := []DashboardStars{}

	current := &DashboardStars{}
	var orgID int64
	var userUID string
	var dashboardUID string
	var updated time.Time

	for rows.Next() {
		err := rows.Scan(&orgID, &userUID, &dashboardUID, &updated)
		if err != nil {
			return nil, 0, err
		}

		if orgID != current.OrgID || userUID != current.UserUID {
			if current.UserUID != "" {
				stars = append(stars, *current)
			}
			current = &DashboardStars{
				OrgID:   orgID,
				UserUID: userUID,
			}
		}
		ts := updated.UnixMilli()
		if ts > current.Last {
			current.Last = ts
		}
		if ts < current.First || current.First == 0 {
			current.First = ts
		}
		current.Dashboards = append(current.Dashboards, dashboardUID)
	}

	// Add the last value
	if current.UserUID != "" {
		stars = append(stars, *current)
	}

	return stars, rv, nil
}
