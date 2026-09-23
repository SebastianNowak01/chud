package app

import (
	"github.com/sebnow/chud/features/activities"
	"github.com/sebnow/chud/features/entries"
	"github.com/sebnow/chud/features/plans"
	"github.com/sebnow/chud/features/stats"
	"github.com/sebnow/chud/features/users"
)

type Handlers struct {
	User     users.UserAPIController
	Activity activities.ActivityAPIController
	Plan     plans.PlanAPIController
	Entry    entries.EntryAPIController
	Stats    stats.StatsAPIController
}
