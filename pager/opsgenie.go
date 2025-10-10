package pager

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/firehydrant/signals-migrator/console"
	"github.com/firehydrant/signals-migrator/store"
	"github.com/gosimple/slug"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/opsgenie/opsgenie-go-sdk-v2/escalation"
	"github.com/opsgenie/opsgenie-go-sdk-v2/og"
	"github.com/opsgenie/opsgenie-go-sdk-v2/schedule"
	"github.com/opsgenie/opsgenie-go-sdk-v2/team"
	"github.com/opsgenie/opsgenie-go-sdk-v2/user"
)

type Opsgenie struct {
	userClient       *user.Client
	teamClient       *team.Client
	scheduleClient   *schedule.Client
	escalationClient *escalation.Client
}

func NewOpsgenie(apiKey string) *Opsgenie {
	conf := &client.Config{
		ApiKey: apiKey,

		// This corresponds to logrus.ErrorLevel but avoids importing logrus,
		// since we don't use it but the SDK imports it.
		LogLevel: 2,
	}
	return NewOpsgenieWithConfig(conf)
}

func NewOpsgenieWithURL(apiKey, url string) *Opsgenie {
	conf := &client.Config{
		ApiKey:         apiKey,
		OpsGenieAPIURL: client.ApiUrl(url),

		// This corresponds to logrus.ErrorLevel but avoids importing logrus,
		// since we don't use it but the SDK imports it.
		LogLevel: 2,
	}
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
	escalationClient, err := escalation.NewClient(conf)
	if err != nil {
		panic(fmt.Sprintf("creating opsgenie escalation client: %v", err))
	}
	return &Opsgenie{
		userClient:       userClient,
		teamClient:       teamClient,
		scheduleClient:   scheduleClient,
		escalationClient: escalationClient,
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

				// Opsgenie "Username" is the user's email.
				Annotations: fmt.Sprintf("[Opsgenie] %s %s", user.Id, user.Username),
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
				if sqlErr, ok := store.AsSQLError(err); ok && sqlErr.IsForeignKeyConstraint() {
					console.Warnf("User %q (%s) isn't imported. Skipping...\n", m.User.Username, m.User.ID)
					return nil
				}
				return fmt.Errorf("saving user %q (%s) as member of %q (%s) to db: %w", m.User.Username, m.User.ID, t.Name, t.ID, err)
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
	// this maps directly to our structure with each rotation belonging to a schedule
	teamID := ""
	if s.OwnerTeam != nil {
		teamID = s.OwnerTeam.Id
	} else {
		console.Warnf("No owning team found for schedule %s, skipping...\n", s.Id)
		return nil
	}

	scheduleParams := store.InsertExtScheduleV2Params{
		ID:               s.Id,
		Name:             s.Name,
		Description:      s.Description,
		Timezone:         s.Timezone,
		TeamID:           teamID,
		SourceSystem:     "opsgenie",
		SourceScheduleID: s.Id,
	}

	q := store.UseQueries(ctx)

	if _, err := q.GetExtTeam(ctx, teamID); err != nil {
		console.Warnf("Schedule %q (%s) belongs to a team that isn't imported.  Skipping...\n", s.Name, s.Id)
		return nil
	}
	if err := q.InsertExtScheduleV2(ctx, scheduleParams); err != nil {
		return fmt.Errorf("saving schedule: %w", err)
	}

	// Create rotation records under the parent schedule
	for i, rotation := range resp.Schedule.Rotations {
		if err := o.saveRotationToDB(ctx, s.Id, rotation, i); err != nil {
			return fmt.Errorf("saving rotation to db: %w", err)
		}
	}
	return nil
}

func (o *Opsgenie) saveRotationToDB(ctx context.Context, scheduleID string, r og.Rotation, rotationOrder int) error {
	schedule, err := store.UseQueries(ctx).GetExtScheduleV2(ctx, scheduleID)
	if err != nil {
		return fmt.Errorf("getting schedule info: %w", err)
	}

	rotationName := r.Name
	if rotationName == "" {
		rotationName = fmt.Sprintf("%srotation%d", schedule.Name, rotationOrder+1)
	}

	desc := fmt.Sprintf("%s (%s)", schedule.Description, rotationName)
	desc = strings.TrimSpace(desc)

	ogsHandoffTime := r.StartDate.Format(time.TimeOnly)
	ogsHandoffDay := strings.ToLower(r.StartDate.Weekday().String())

	var ogsStrategy, ogsDuration string
	rotationLength := r.Length
	// Opsgenie's concept of rotation duration is different from FireHydrant's
	//   Opsgenie includes a rotation type and a rotation length for each rotation
	//   Type is the type of rotation (daily, weekly, hourly)
	//   Length is the number of times that interval repeats for a given shift
	//   For example, a daily rotation with a length of 2 would be a 2 day shift
	// Our daily shift is always exactly 24 hours, with no variable length, weekly always 7 days, etc.
	// Therefore, any Opsgenie rotation with a length > 1 is a custom rotation in FH, and we need to calculate the duration.

	// If length > 1, convert to custom strategy with calculated duration
	if rotationLength > 1 {
		ogsStrategy = "custom"
		switch r.Type {
		case og.Daily:
			ogsDuration = fmt.Sprintf("PT%dH", 24*rotationLength)
		case og.Weekly:
			ogsDuration = fmt.Sprintf("PT%dH", 168*rotationLength)
		case og.Hourly:
			ogsDuration = fmt.Sprintf("PT%dH", rotationLength)
		default:
			return fmt.Errorf("unexpected schedule strategy %s.  skipping rotation %s of schedule %s", r.Type, r.Id, scheduleID)
		}
	} else {
		// Length == 1, use standard strategies
		switch r.Type {
		case og.Daily:
			ogsStrategy = "daily"
			ogsDuration = ""
		case og.Weekly:
			ogsStrategy = "weekly"
			ogsDuration = ""
		case og.Hourly:
			ogsStrategy = "custom"
			ogsDuration = fmt.Sprintf("PT%dH", rotationLength)
		default:
			return fmt.Errorf("unexpected schedule strategy %s.  skipping rotation %s of schedule %s", r.Type, r.Id, scheduleID)
		}
	}

	rotationParams := store.InsertExtRotationParams{
		ID:            r.Id,
		ScheduleID:    scheduleID,
		Name:          rotationName,
		Description:   desc,
		Strategy:      ogsStrategy,
		ShiftDuration: ogsDuration,
		StartTime:     r.StartDate.Format(time.RFC3339),
		HandoffTime:   ogsHandoffTime,
		HandoffDay:    ogsHandoffDay,
		RotationOrder: int64(rotationOrder),
	}

	q := store.UseQueries(ctx)
	if err := q.InsertExtRotation(ctx, rotationParams); err != nil {
		return fmt.Errorf("saving rotation: %w", err)
	}

	// ExtRotationMembers
	for i, p := range r.Participants {
		// Store the order of participants as they appear in the API response
		// This preserves the exact order from OpsGenie
		if err := q.InsertExtRotationMember(ctx, store.InsertExtRotationMemberParams{
			RotationID:  r.Id,
			UserID:      p.Id,
			MemberOrder: int64(i),
		}); err != nil {
			if strings.Contains(err.Error(), "FOREIGN KEY constraint") {
				console.Warnf("User %s not found for rotation %s, skipping...\n", p.Id, r.Id)
			} else if strings.Contains(err.Error(), "UNIQUE constraint") {
				console.Warnf("User %s already exists for rotation %s, skipping duplicate...\n", p.Id, r.Id)
			} else {
				return fmt.Errorf("saving rotation user: %w", err)
			}
		}
	}

	// ExtRotationRestriction
	// The opsgenie api, may in burn in the hell of a thousand suns, returns TimeRestriction.Restriction if the type is TimeOfDay
	// and TimeRestriction.RestrictionList if the type is WeekdayAndTimeOfDay because... I've got nothing.  There is no excuse that
	// works here, and I can't make it make sense for them.  At least they documented this... no wait, they didn't actually document
	// it at all and I had to guess at this behavior from testing before digging into the source to confirm which, as it turns out,
	// is exactly how I wanted to spend my morning so thanks for that.
	if r.TimeRestriction != nil {
		loc, err := time.LoadLocation(schedule.Timezone)
		if err != nil {
			console.Warnf("unable to parse location %s.  using UTC instead", schedule.Timezone)
			loc = time.UTC
		}
		switch r.TimeRestriction.Type {
		case og.WeekdayAndTimeOfDay:
			for i, tr := range r.TimeRestriction.RestrictionList {
				startTime := time.Date(0, time.January, 1, int(*tr.StartHour), int(*tr.StartMin), 0, 0, loc)
				endTime := time.Date(0, time.January, 1, int(*tr.EndHour), int(*tr.EndMin), 0, 0, loc)

				rotationRestrictionsParams := store.InsertExtRotationRestrictionParams{
					RotationID:       r.Id,
					RestrictionIndex: strconv.Itoa(i),
					StartDay:         strings.ToLower(string(tr.StartDay)),
					StartTime:        startTime.Format(time.TimeOnly),
					EndDay:           strings.ToLower(string(tr.EndDay)),
					EndTime:          endTime.Format(time.TimeOnly),
				}
				if err := q.InsertExtRotationRestriction(ctx, rotationRestrictionsParams); err != nil {
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

				rotationRestrictionsParams := store.InsertExtRotationRestrictionParams{
					RotationID:       r.Id,
					RestrictionIndex: strconv.Itoa(i),
					StartDay:         startDayStr,
					StartTime:        startTime.Format(time.TimeOnly),
					EndDay:           endDayStr,
					EndTime:          endTime.Format(time.TimeOnly),
				}
				if err := q.InsertExtRotationRestriction(ctx, rotationRestrictionsParams); err != nil {
					return fmt.Errorf("saving time of day restriction: %w", err)
				}
			}
		default:
			console.Warnf("Unknown schedule restriction type '%s' for rotation '%s', skipping...\n", r.TimeRestriction.Type, r.Id)
		}
	}

	return nil
}

func (o *Opsgenie) LoadEscalationPolicies(ctx context.Context) error {
	resp, err := o.escalationClient.List(ctx)
	if err != nil {
		return err
	}

	for _, policy := range resp.Escalations {
		if err := o.saveEscalationPolicyToDB(ctx, policy); err != nil {
			return fmt.Errorf("saving escalation policy to db: %w", err)
		}
	}

	return nil
}

func (o *Opsgenie) saveEscalationPolicyToDB(ctx context.Context, policy escalation.Escalation) error {
	var repeatLimit int64
	repeatInterval := sql.NullString{Valid: false}
	if policy.Repeat != nil {
		repeatLimit = int64(policy.Repeat.Count)
		repeatInterval.Valid = true
		repeatInterval.String = fmt.Sprintf("PT%dM", policy.Repeat.WaitInterval)
	}
	teamID := sql.NullString{}
	if policy.OwnerTeam != nil {
		teamID.Valid = true
		teamID.String = policy.OwnerTeam.Id
	}
	ep := store.InsertExtEscalationPolicyParams{
		ID:             policy.Id,
		Name:           policy.Name,
		Description:    policy.Description,
		TeamID:         teamID,
		RepeatInterval: repeatInterval,
		RepeatLimit:    repeatLimit,
	}
	if err := store.UseQueries(ctx).InsertExtEscalationPolicy(ctx, ep); err != nil {
		if sqlErr, ok := store.AsSQLError(err); ok && sqlErr.IsForeignKeyConstraint() {
			console.Warnf("Escalation policy %q (%s) belongs to a team that isn't imported. Skipping...\n", ep.Name, ep.ID)
			return nil
		}

		return fmt.Errorf("saving escalation policy %q (%s): %w", ep.Name, ep.ID, err)
	}

	for i, rule := range policy.Rules {
		timeout := calculateTimeout(policy, i)
		if err := o.saveEscalationPolicyStepToDB(ctx, ep.ID, rule, int64(i), timeout); err != nil {
			return fmt.Errorf("saving escalation rule to db: %w", err)
		}
	}
	return nil
}

// regarding step.Timeout
// Opsgenie has a delay property on each rule, indicating the total time delay before executing that step (where time zero is the start time of the escalation
// policy).  On the FH side, we have a timeout value for each step, which represents the amount of time to wait *after* firing that step before moving to the
// next step, with a minimum timeout of 1 minutes and a max of 60 min.  Opsgenie could have multiple steps with the same delay value, in which case all those
// steps would fire simultaneously (this is not supported on the FH side).  Opsgenie also supports time intervals on the order of hours, days, weeks, or months
// none of which FH supports, so we're locking the max time between steps to 1 hour (and so are only interested in a time unit of minutes).

// To do a best approximation of this, we are assuming that the Opsgenie steps are delivered in order (they seem to be) and are setting the timeout to:
// Max(1, Min(60, step.next[delay minutes] - step.current[delay minutes]))
// with a special rule for the final step of:
// Max(1, Min(60, policy.Repeat.WaitInterval minutes))

func calculateTimeout(policy escalation.Escalation, position int) string {
	timeout := "PT1M"
	if position+1 == len(policy.Rules) {
		if policy.Repeat != nil {
			// WaitInterval is always in minutes, but delay amounts can be other units.
			timeout = fmt.Sprintf("PT%dM", int(math.Max(1, math.Min(float64(policy.Repeat.WaitInterval), 60))))
		}
		// if the policy doesn't repeat, then it shouldn't matter what this value is.  We'll go with a default of 1 min.
	} else {
		currentDelayMin := uint32(60)
		if policy.Rules[position].Delay.TimeUnit == og.Minutes {
			currentDelayMin = policy.Rules[position].Delay.TimeAmount
		}
		nextDelayMin := uint32(60)
		if policy.Rules[position+1].Delay.TimeUnit == og.Minutes {
			nextDelayMin = policy.Rules[position+1].Delay.TimeAmount
		}
		// Warn the user that we're locking to min/max and give the actual value
		if policy.Rules[position+1].Delay.TimeUnit != og.Minutes || policy.Rules[position+1].Delay.TimeAmount > 60 {
			console.Warnf("Actual delay time for step %d is %d %s.  Locking to a max of 60 minutes.\n",
				position+1,
				policy.Rules[position+1].Delay.TimeAmount,
				policy.Rules[position+1].Delay.TimeUnit)
		}
		if policy.Rules[position+1].Delay.TimeAmount == 0 {
			console.Warnf("Actual delay time for step %d is 0.  Locking to a min of 1 minute.\n", position+1)
		}

		timeout = fmt.Sprintf("PT%dM", int(math.Max(1, math.Min(float64(nextDelayMin-currentDelayMin), 60))))
	}
	return timeout
}

func (o *Opsgenie) saveEscalationPolicyStepToDB(ctx context.Context, policyID string, rule escalation.Rule, position int64, timeout string) error {
	stepID := fmt.Sprintf("%s-%d", policyID, position)
	step := store.InsertExtEscalationPolicyStepParams{
		ID:                 stepID,
		EscalationPolicyID: policyID,
		Position:           position,
		Timeout:            timeout,
	}

	t := store.InsertExtEscalationPolicyStepTargetParams{
		EscalationPolicyStepID: stepID,
		TargetID:               rule.Recipient.Id,
	}

	// The actual target for a rule is a combination of rule.Recipient.Type and rule.NotifyType, with only some combinations being valid.
	// A handy chart (because if I have to know this, so do you):
	// RecipientType    NotifyType    Notes
	// User             default       Just notify the user.
	// Schedule         default       Notify the currently on-call person for this schedule
	// Team             default       Notify the default escalation policy for a team.  We support this only as a handoff step, not in the middle of a policy
	// Team             users         Notify all non-admin members of a team
	// Team             admins        Notify all admin members of a team
	// Team             all           Notify all members of a team
	// Team             random        Notify a random?! member of the team (don't even get me started...)
	// Schedule         next          Notify the person who will be on-call next in the given schedule
	// Schedule         previous      Notify the person who was previously oncall in the given schedule
	// We only support recepients of User or Schedule and only the 'default' NotifyType.  Anything else we're just logging and skipping.

	if rule.NotifyType != og.Default {
		console.Errorf("Escalation policy step target is '%s' notify type '%s' for policy '%s' step %d.\nWe currently do not support this notify type, skipping...\n",
			rule.Recipient.Type,
			rule.NotifyType,
			policyID,
			position)
		return nil
	}

	switch rule.Recipient.Type {
	case og.User:
		t.TargetType = store.TARGET_TYPE_USER
	case og.Schedule:
		t.TargetType = store.TARGET_TYPE_SCHEDULE
	default:
		console.Warnf("Escalation policy step target is '%s' notify type '%s' for policy '%s' step %d, skipping...\n",
			rule.Recipient.Type,
			rule.NotifyType,
			policyID,
			position)
		return nil
	}

	if err := store.UseQueries(ctx).InsertExtEscalationPolicyStep(ctx, step); err != nil {
		return fmt.Errorf("saving escalation policy step: %w", err)
	}

	// For Opsgenie, the person(s) on-call for a schedule is the oncall person in all rotations for that schedule for the given time (rotations may overlap).
	// We've saved each rotation as a separate FH schedule.  So, since we're not sure exactly what was intended by escalating to a Opsgenie schedule, we're
	// adding all of the FH schedules corresponding to the Opsgenie rotations for that schedule as targets.  This should be viewed as an approximation and
	// review should certainly be required here.

	if t.TargetType == store.TARGET_TYPE_SCHEDULE {
		schedule, err := store.UseQueries(ctx).GetExtScheduleV2(ctx, rule.Recipient.Id)
		if err != nil {
			console.Errorf("Failed to resolve schedule target '%s' for escalation policy step '%s': %s\n", rule.Recipient.Id, stepID, err.Error())
			return fmt.Errorf("resolving schedule target '%s': %w", rule.Recipient.Id, err)
		}

		t.TargetID = schedule.ID
		if err := store.UseQueries(ctx).InsertExtEscalationPolicyStepTarget(ctx, t); err != nil {
			return fmt.Errorf("saving escalation policy step target: %w", err)
		}
	} else {
		if err := store.UseQueries(ctx).InsertExtEscalationPolicyStepTarget(ctx, t); err != nil {
			return fmt.Errorf("saving escalation policy step target: %w", err)
		}
	}

	return nil
}
