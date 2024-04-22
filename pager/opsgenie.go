package pager

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/firehydrant/signals-migrator/console"
	"github.com/firehydrant/signals-migrator/store"
	"github.com/gosimple/slug"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/opsgenie/opsgenie-go-sdk-v2/og"
	"github.com/opsgenie/opsgenie-go-sdk-v2/schedule"
	"github.com/opsgenie/opsgenie-go-sdk-v2/team"
	"github.com/opsgenie/opsgenie-go-sdk-v2/user"
)

type Opsgenie struct {
	userClient     *user.Client
	teamClient     *team.Client
	scheduleClient *schedule.Client
}

func NewOpsgenie(apiKey string) *Opsgenie {
	conf := &client.Config{ApiKey: apiKey}
	return NewOpsgenieWithConfig(conf)
}

func NewOpsgenieWithURL(apiKey, url string) *Opsgenie {
	conf := &client.Config{ApiKey: apiKey, OpsGenieAPIURL: client.ApiUrl(url)}
	return NewOpsgenieWithConfig(conf)
}

func NewOpsgenieWithConfig(conf *client.Config) *Opsgenie {
	userClient, err := user.NewClient(conf)
	if err != nil {
		panic(fmt.Sprintf("creating opsgenie user client: %v", err))
	}
	teamClient, err := team.NewClient(conf)
	if err != nil {
		panic(fmt.Sprintf("creating opsgenie team client: %v", err))
	}
	scheduleClient, err := schedule.NewClient(conf)
	if err != nil {
		panic(fmt.Sprintf("creating opsgenie schedule client: %v", err))
	}
	return &Opsgenie{
		userClient:     userClient,
		teamClient:     teamClient,
		scheduleClient: scheduleClient,
	}
}

func (p *Opsgenie) Kind() string {
	return "Opsgenie"
}

func (o *Opsgenie) TeamInterfaces() []string {
	return []string{"team"}
}

func (o *Opsgenie) UseTeamInterface(string) error {
	return nil
}

func (o *Opsgenie) Teams(ctx context.Context) ([]store.ExtTeam, error) {
	return store.UseQueries(ctx).ListExtTeams(ctx)
}

func (o *Opsgenie) LoadUsers(ctx context.Context) error {
	opts := user.ListRequest{}

	for {
		resp, err := o.userClient.List(ctx, &opts)
		if err != nil {
			return fmt.Errorf("listing users: %w", err)
		}

		for _, user := range resp.Users {
			if err := store.UseQueries(ctx).InsertExtUser(ctx, store.InsertExtUserParams{
				ID:    user.Id,
				Name:  user.FullName,
				Email: user.Username,
			}); err != nil {
				return fmt.Errorf("saving user to db: %w", err)
			}
		}

		// Results are paginated, so break if we're on the last page.
		if resp.Paging.Next == "" {
			break
		}
		opts.Offset += len(resp.Users)
	}
	return nil
}

func (o *Opsgenie) LoadTeams(ctx context.Context) error {
	opts := &team.ListTeamRequest{}

	resp, err := o.teamClient.List(ctx, opts)
	if err != nil {
		return fmt.Errorf("listing teams: %w", err)
	}

	for _, t := range resp.Teams {
		if err := store.UseQueries(ctx).InsertExtTeam(ctx, store.InsertExtTeamParams{
			ID:   t.Id,
			Name: t.Name,
			// Opsgenie does not expose a slug, so generate one.
			Slug: slug.Make(t.Name),
		}); err != nil {
			return fmt.Errorf("saving team to db: %w", err)
		}
	}

	return nil
}

func (o *Opsgenie) LoadTeamMembers(ctx context.Context) error {
	teams, err := store.UseQueries(ctx).ListTeams(ctx)
	if err != nil {
		return fmt.Errorf("listing teams: %w", err)
	}
	for _, t := range teams {
		resp, err := o.teamClient.Get(ctx, &team.GetTeamRequest{
			IdentifierType:  team.Id,
			IdentifierValue: t.ID,
		})
		if err != nil {
			return fmt.Errorf("getting team members: %w", err)
		}

		for _, m := range resp.Members {
			if err := store.UseQueries(ctx).InsertExtMembership(ctx, store.InsertExtMembershipParams{
				TeamID: t.ID,
				UserID: m.User.ID,
			}); err != nil {
				return fmt.Errorf("saving team member to db: %w", err)
			}
		}
	}
	return nil
}

func (o *Opsgenie) LoadSchedules(ctx context.Context) error {
	expandListRequest := true
	resp, err := o.scheduleClient.List(ctx, &schedule.ListRequest{
		Expand: &expandListRequest,
	})

	if err != nil {
		return err
	}

	for _, schedule := range resp.Schedule {
		// To decide: check enabled field and don't create if false?
		if err := o.saveScheduleToDB(ctx, schedule); err != nil {
			return fmt.Errorf("saving schedule to db: %w", err)
		}
	}

	return nil
}

func (o *Opsgenie) saveScheduleToDB(ctx context.Context, s schedule.Schedule) error {
	resp, err := o.scheduleClient.Get(ctx, &schedule.GetRequest{
		IdentifierType:  schedule.Id,
		IdentifierValue: s.Id,
	})

	if err != nil {
		return err
	}

	// each Opsgenie schedule can have multiple rotations, where each rotation has participants, start and handoff date/times, and restrictions.
	// as such, I think an opsgenie rotation best maps to a FH schedule, so let's try that.

	for _, rotation := range resp.Schedule.Rotations {
		if err := o.saveRotationToDB(ctx, s, rotation); err != nil {
			return fmt.Errorf("saving schedule to db: %w", err)
		}
	}
	return nil
}

func (o *Opsgenie) saveRotationToDB(ctx context.Context, s schedule.Schedule, r og.Rotation) error {
	// ExtSchedule
	desc := fmt.Sprintf("%s (%s)", s.Description, r.Name)
	desc = strings.TrimSpace(desc)

	ogsHandoffTime := r.StartDate.Format(time.TimeOnly)
	ogsHandoffDay := strings.ToLower(r.StartDate.Weekday().String())

	var ogsStrategy, ogsDuration string
	switch r.Type {
	case og.Daily:
		ogsStrategy = "daily"
		ogsDuration = ""
	case og.Weekly:
		ogsStrategy = "weekly"
		ogsDuration = ""
	case og.Hourly:
		ogsStrategy = "custom"
		ogsDuration = fmt.Sprintf("PT%dH", r.Length)
	default:
		return fmt.Errorf("unexpected schedule strategy %s.  skipping rotation %s of schedule %s", ogsStrategy, r.Id, s.Id)
	}

	ogsParams := store.InsertExtScheduleParams{
		ID:            s.Id + "-" + r.Id,
		Name:          s.Name + " - " + r.Name,
		Timezone:      s.Timezone,
		Description:   desc,
		HandoffTime:   ogsHandoffTime,
		HandoffDay:    ogsHandoffDay,
		Strategy:      ogsStrategy,
		ShiftDuration: ogsDuration,
	}

	q := store.UseQueries(ctx)
	if err := q.InsertExtSchedule(ctx, ogsParams); err != nil {
		return fmt.Errorf("saving schedule: %w", err)
	}

	// ExtScheduleTeam
	if s.OwnerTeam != nil {
		if err := q.InsertExtScheduleTeam(ctx, store.InsertExtScheduleTeamParams{
			ScheduleID: ogsParams.ID,
			TeamID:     s.OwnerTeam.Id,
		}); err != nil {
			if strings.Contains(err.Error(), "FOREIGN KEY constraint") {
				console.Warnf("Team %s not found for schedule %s, skipping...\n", s.OwnerTeam.Id, ogsParams.ID)
			} else {
				return fmt.Errorf("saving schedule team: %w", err)
			}
		}
	} else {
		console.Warnf("No owning team found for schedule %s, skipping...\n", ogsParams.ID)
	}

	// ExtScheduleMembers
	for _, p := range r.Participants {
		if err := q.InsertExtScheduleMember(ctx, store.InsertExtScheduleMemberParams{
			ScheduleID: ogsParams.ID,
			UserID:     p.Id,
		}); err != nil {
			if strings.Contains(err.Error(), "FOREIGN KEY constraint") {
				console.Warnf("User %s not found for schedule %s, skipping...\n", p.Id, ogsParams.ID)
			} else if strings.Contains(err.Error(), "UNIQUE constraint") {
				console.Warnf("User %s already exists for schedule %s, skipping duplicate...\n", p.Id, ogsParams.ID)
			} else {
				return fmt.Errorf("saving schedule user: %w", err)
			}
		}
	}

	// ExtScheduleRestriction
	// The opsgenie api, may in burn in the hell of a thousand suns, returns TimeRestriction.Restriction if the type is TimeOfDay
	// and TimeRestriction.RestrictionList if the type is WeekdayAndTimeOfDay because... I've got nothing.  There is no excuse that
	// works here, and I can't make it make sense for them.  At least they documented this... no wait, they didn't actually document
	// it at all and I had to guess at this behavior from testing before digging into the source to confirm which, as it turns out,
	// is exactly how I wanted to spend my morning so thanks for that.
	if r.TimeRestriction != nil {
		loc, err := time.LoadLocation(s.Timezone)
		if err != nil {
			console.Warnf("unable to parse location %s.  using UTC instead", s.Timezone)
			loc = time.UTC
		}
		switch r.TimeRestriction.Type {
		case og.WeekdayAndTimeOfDay:
			for i, tr := range r.TimeRestriction.RestrictionList {
				startTime := time.Date(0, time.January, 1, int(*tr.StartHour), int(*tr.StartMin), 0, 0, loc)
				endTime := time.Date(0, time.January, 1, int(*tr.EndHour), int(*tr.EndMin), 0, 0, loc)

				ogsRestrictionsParams := store.InsertExtScheduleRestrictionParams{
					ScheduleID:       ogsParams.ID,
					RestrictionIndex: strconv.Itoa(i),
					StartDay:         strings.ToLower(string(tr.StartDay)),
					StartTime:        startTime.Format(time.TimeOnly),
					EndDay:           strings.ToLower(string(tr.EndDay)),
					EndTime:          endTime.Format(time.TimeOnly),
				}
				if err := q.InsertExtScheduleRestriction(ctx, ogsRestrictionsParams); err != nil {
					return fmt.Errorf("saving time of day restriction: %w", err)
				}
			}
		case og.TimeOfDay:
			for i := range 7 {
				tr := r.TimeRestriction.Restriction
				startDayStr := strings.ToLower(time.Weekday(i).String())
				startTime := time.Date(0, time.January, 1, int(*tr.StartHour), int(*tr.StartMin), 0, 0, loc)
				endTime := time.Date(0, time.January, 1, int(*tr.EndHour), int(*tr.EndMin), 0, 0, loc)
				var endDayStr string
				if endTime.Before(startTime) {
					day := (i + 1) % 7
					endDayStr = strings.ToLower(time.Weekday(day).String())
				} else {
					endDayStr = startDayStr
				}

				ogsRestrictionsParams := store.InsertExtScheduleRestrictionParams{
					ScheduleID:       ogsParams.ID,
					RestrictionIndex: strconv.Itoa(i),
					StartDay:         startDayStr,
					StartTime:        startTime.Format(time.TimeOnly),
					EndDay:           endDayStr,
					EndTime:          endTime.Format(time.TimeOnly),
				}
				if err := q.InsertExtScheduleRestriction(ctx, ogsRestrictionsParams); err != nil {
					return fmt.Errorf("saving time of day restriction: %w", err)
				}
			}
		default:
			console.Warnf("Unknown schedule restriction type '%s' for schedule '%s', skipping...\n", r.TimeRestriction.Type, ogsParams.ID)
		}
	}

	return nil
}

func (o *Opsgenie) LoadEscalationPolicies(ctx context.Context) error {
	// TODO: implement
	console.Warnf("opsgenie.LoadEscalationPolicies is not currently supported.")
	return nil
}
