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

	// Create a new userClient
	var userClient, _ = user.NewClient(&client.Config{
		ApiKey: apiKey,
	})

	var teamClient, _ = team.NewClient(&client.Config{
		ApiKey: apiKey,
	})

	var scheduleClient, _ = schedule.NewClient(&client.Config{
		ApiKey: apiKey,
	})

	return &Opsgenie{
		userClient:     userClient,
		teamClient:     teamClient,
		scheduleClient: scheduleClient,
	}
}

func NewOpsgenieWithURL(apiKey, url string) *Opsgenie {

	// Create a new userClient
	var userClient, _ = user.NewClient(&client.Config{
		ApiKey:         apiKey,
		OpsGenieAPIURL: client.ApiUrl(url),
	})

	var teamClient, _ = team.NewClient(&client.Config{
		ApiKey:         apiKey,
		OpsGenieAPIURL: client.ApiUrl(url),
	})

	var scheduleClient, _ = schedule.NewClient(&client.Config{
		ApiKey:         apiKey,
		OpsGenieAPIURL: client.ApiUrl(url),
	})

	return &Opsgenie{
		userClient:     userClient,
		teamClient:     teamClient,
		scheduleClient: scheduleClient,
	}
}

func (p *Opsgenie) Kind() string {
	return "opsgenie"
}

func (o *Opsgenie) LoadSchedules(ctx context.Context) error {
	resp, err := o.scheduleClient.List(ctx, &schedule.ListRequest{
		// this is the cleanest way I found of making a *bool.  Happy to replace it with anything more readable.
		Expand: func() *bool { b := true; return &b }(),
	})

	if err != nil {
		return err
	}

	for _, schedule := range resp.Schedule {
		// To decide: check enabled field and don't create if false?
		if err := o.saveScheduleToDB(ctx, schedule); err != nil {
			return fmt.Errorf("error saving schedule to db: %w", err)
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
			return fmt.Errorf("error saving schedule to db: %w", err)
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

	var ogsStrategy string
	switch r.Type {
	case og.Daily:
		ogsStrategy = "daily"
	case og.Weekly:
		ogsStrategy = "weekly"
	default:
		ogsStrategy = "custom"
	}

	ogsParams := store.InsertExtScheduleParams{
		ID:            s.Id + "-" + r.Id,
		Name:          s.Name + " - " + r.Name,
		Timezone:      s.Timezone,
		Description:   desc,
		HandoffTime:   ogsHandoffTime,
		HandoffDay:    ogsHandoffDay,
		Strategy:      ogsStrategy,
		ShiftDuration: "",
	}

	q := store.UseQueries(ctx)
	if err := q.InsertExtSchedule(ctx, ogsParams); err != nil {
		return fmt.Errorf("error saving schedule: %w", err)
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
				return fmt.Errorf("error saving schedule team: %w", err)
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
				return fmt.Errorf("error saving schedule user: %w", err)
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
		switch r.TimeRestriction.Type {
		case og.WeekdayAndTimeOfDay:
			for i, tr := range r.TimeRestriction.RestrictionList {
				startTime, _ := time.Parse(time.TimeOnly, fmt.Sprintf("%d:%02d:00", *tr.StartHour, *tr.StartMin))
				endTime, _ := time.Parse(time.TimeOnly, fmt.Sprintf("%d:%02d:00", *tr.EndHour, *tr.EndMin))

				ogsRestrictionsParams := store.InsertExtScheduleRestrictionParams{
					ScheduleID:       ogsParams.ID,
					RestrictionIndex: strconv.Itoa(i),
					StartDay:         strings.ToLower(string(tr.StartDay)),
					StartTime:        startTime.Format(time.TimeOnly),
					EndDay:           strings.ToLower(string(tr.EndDay)),
					EndTime:          endTime.Format(time.TimeOnly),
				}
				if err := q.InsertExtScheduleRestriction(ctx, ogsRestrictionsParams); err != nil {
					return fmt.Errorf("error saving time of day restriction: %w", err)
				}
			}
		case og.TimeOfDay:
			for i := range 7 {
				tr := r.TimeRestriction.Restriction
				startDayStr := strings.ToLower(time.Weekday(i).String())
				startTime, _ := time.Parse(time.TimeOnly, fmt.Sprintf("%d:%02d:00", *tr.StartHour, *tr.StartMin))
				endTime, _ := time.Parse(time.TimeOnly, fmt.Sprintf("%d:%02d:00", *tr.EndHour, *tr.EndMin))
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
					return fmt.Errorf("error saving time of day restriction: %w", err)
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

func (p *Opsgenie) PopulateTeamMembers(ctx context.Context, t *Team) error {
	members := []*User{}

	resp, err := p.teamClient.Get(ctx, &team.GetTeamRequest{
		IdentifierType:  team.Name,
		IdentifierValue: t.Name,
	})

	if err != nil {
		return err
	}

	for _, member := range resp.Members {
		members = append(members, &User{Resource: Resource{ID: member.User.ID}})
	}

	t.Members = members

	return nil
}

func (p *Opsgenie) ListTeams(ctx context.Context) ([]*Team, error) {
	teams := []*Team{}
	opts := team.ListTeamRequest{}

	resp, err := p.teamClient.List(ctx, &opts)
	if err != nil {
		return nil, err
	}

	for _, team := range resp.Teams {
		teams = append(teams, p.toTeam(team))
	}

	return teams, nil
}

func (p *Opsgenie) toTeam(team team.ListedTeams) *Team {
	return &Team{
		// Opsgenie does not expose a slug, so generate one.
		Slug: slug.Make(team.Name),
		Resource: Resource{
			ID:   team.Id,
			Name: team.Name,
		},
	}
}

func (p *Opsgenie) ListUsers(ctx context.Context) ([]*User, error) {
	users := []*User{}
	opts := user.ListRequest{}

	for {
		resp, err := p.userClient.List(ctx, &opts)
		if err != nil {
			return nil, err
		}

		for _, user := range resp.Users {
			users = append(users, p.toUser(user))
		}

		// Results are paginated, so break if we're on the last page.
		if resp.Paging.Next == "" {
			break
		}
		opts.Offset += len(resp.Users)
	}
	return users, nil
}

func (p *Opsgenie) toUser(user user.User) *User {
	return &User{
		Email: user.Username,
		Resource: Resource{
			ID:   user.Id,
			Name: user.FullName,
		},
	}
}
