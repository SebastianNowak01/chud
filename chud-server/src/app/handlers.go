package app

import (
	"github.com/sebnow/chud/features/activities"
	"github.com/sebnow/chud/features/entries"
	"github.com/sebnow/chud/features/plans"
	"github.com/sebnow/chud/features/stats"
	"github.com/sebnow/chud/features/summaries"
	"github.com/sebnow/chud/features/users"
	"github.com/sebnow/chud/platform/auth"
)

type Handlers struct {
	User     users.UserAPIController
	Activity activities.ActivityAPIController
	Plan     plans.PlanAPIController
	Entry    entries.EntryAPIController
	Stats    stats.StatsAPIController
	Summary  summaries.SummaryAPIController
	Session  auth.SessionLookup
}
