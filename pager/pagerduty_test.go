package pager_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/firehydrant/signals-migrator/pager"
	"github.com/firehydrant/signals-migrator/store"
	"github.com/firehydrant/signals-migrator/tfrender"
)

func TestPagerDuty(t *testing.T) {
	// What we're testing here is whether we manage to produce expected data in SQLite for the given API responses.
	// In other words, given responses within ./testdata/TestPagerDuty/apiserver, we expect the data transformed by
	// our implementation to be stored in the database to be consistent.
	// Asserting the content of database is a little tricky. As such, we encode the data into JSON and compare it
	// with the expected JSON in ./testdata/TestPagerDuty/[TestName].golden.json for each test case.

	// Rotation anchoring (and therefore member order) depends on how many
	// rotation cycles have elapsed between each fixture layer's virtual start
	// and "now", so pin the clock — otherwise order expectations flip as real
	// time advances past the fixtures. 2024-04-12 is shortly after the most
	// recent fixture virtual starts (2024-04-05 / 2024-04-08), so those layers
	// have completed zero cycles and keep their original member order.
	pinnedNow := time.Date(2024, 4, 12, 0, 0, 0, 0, time.UTC)

	// Avoid sharing setup code between tests to prevent test pollution in parallel execution.
	setup := func(t *testing.T) (context.Context, pager.Pager) {
		ctx := withTestDB(t)
		ts := pagerProviderHttpServer(t)
		pd := pager.NewPagerDutyWithURL("api-key-very-secret", ts.URL)
		pd.SetNow(func() time.Time { return pinnedNow })
		return ctx, pd
	}

	t.Run("LoadUsers", func(t *testing.T) {
		t.Parallel()
		ctx, pd := setup(t)

		if err := pd.LoadUsers(ctx); err != nil {
			t.Fatalf("error loading users: %s", err)
		}

		u, err := store.UseQueries(ctx).ListExtUsers(ctx)
		if err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		assertJSON(t, u)
	})

	// LoadTeams has 2 variants: one for literal teams and another for importing services as teams.
	// The "state" is maintained globally and as such should not be run in parallel.
	t.Run("LoadTeams", func(t *testing.T) {
		t.Run("loadTeams", func(t *testing.T) {
			ctx, pd := setup(t)

			if err := pd.UseTeamInterface("team"); err != nil {
				t.Fatalf("error setting team interface: %s", err)
			}
			if err := pd.LoadTeams(ctx); err != nil {
				t.Fatalf("error loading teams: %s", err)
			}
			teams, err := store.UseQueries(ctx).ListExtTeams(ctx)
			if err != nil {
				t.Fatalf("error loading teams: %s", err)
			}
			t.Logf("found %d teams", len(teams))
			assertJSON(t, teams)
		})

		t.Run("loadServices", func(t *testing.T) {
			ctx, pd := setup(t)

			if err := pd.UseTeamInterface("service"); err != nil {
				t.Fatalf("error setting team interface: %s", err)
			}
			if err := pd.LoadTeams(ctx); err != nil {
				t.Fatalf("error loading teams: %s", err)
			}
			teams, err := store.UseQueries(ctx).ListExtTeams(ctx)
			if err != nil {
				t.Fatalf("error loading teams: %s", err)
			}
			t.Logf("found %d teams, including services", len(teams))

			// We have 2 services:
			// - Endeavour, which has a team: Page Responder Team
			// - Server under Jack's desk, which has a team: Jack's team
			// We expect the membership association to link the users with the services, not their immediate team.
			assertJSON(t, teams)
		})
	})

	t.Run("LoadTeamMembers", func(t *testing.T) {
		t.Run("loadTeamMembers", func(t *testing.T) {
			ctx, pd := setup(t)

			if err := pd.UseTeamInterface("team"); err != nil {
				t.Fatalf("error setting team interface: %s", err)
			}
			if err := pd.LoadUsers(ctx); err != nil {
				t.Fatalf("error loading users: %s", err)
			}
			if err := pd.LoadTeams(ctx); err != nil {
				t.Fatalf("error loading teams: %s", err)
			}
			if err := pd.LoadTeamMembers(ctx); err != nil {
				t.Fatalf("error loading team members: %s", err)
			}
			members, err := store.UseQueries(ctx).ListExtTeamMemberships(ctx)
			if err != nil {
				t.Fatalf("error loading team members: %s", err)
			}
			t.Logf("found %d team members", len(members))
			assertJSON(t, members)
		})

		t.Run("loadTeamMembersPreservesOrder", func(t *testing.T) {
			ctx, pd := setup(t)

			if err := pd.UseTeamInterface("team"); err != nil {
				t.Fatalf("error setting team interface: %s", err)
			}
			if err := pd.LoadUsers(ctx); err != nil {
				t.Fatalf("error loading users: %s", err)
			}
			if err := pd.LoadTeams(ctx); err != nil {
				t.Fatalf("error loading teams: %s", err)
			}
			if err := pd.LoadTeamMembers(ctx); err != nil {
				t.Fatalf("error loading team members: %s", err)
			}
			members, err := store.UseQueries(ctx).ListExtTeamMemberships(ctx)
			if err != nil {
				t.Fatalf("error loading team members: %s", err)
			}

			// Verify that the order matches the PagerDuty API responses
			// For team PT54U20 (Jen): PXI6XNI, P2C9LBA
			// For team PV9JOXL (Service Catalog Team): PRXEEQ8, PXI6XNI

			// Find members for team PT54U20
			var pt54u20Members []string
			for _, member := range members {
				if member.ExtTeam.ID == "PT54U20" {
					pt54u20Members = append(pt54u20Members, member.ExtUser.ID)
				}
			}

			expectedPT54U20Order := []string{"PXI6XNI", "P2C9LBA"}
			if len(pt54u20Members) != len(expectedPT54U20Order) {
				t.Fatalf("team PT54U20: expected %d members, got %d", len(expectedPT54U20Order), len(pt54u20Members))
			}
			for i, userID := range pt54u20Members {
				if userID != expectedPT54U20Order[i] {
					t.Errorf("team PT54U20 member at position %d: expected %s, got %s", i, expectedPT54U20Order[i], userID)
				}
			}

			// Find members for team PV9JOXL
			var pv9joxlMembers []string
			for _, member := range members {
				if member.ExtTeam.ID == "PV9JOXL" {
					pv9joxlMembers = append(pv9joxlMembers, member.ExtUser.ID)
				}
			}

			expectedPV9JOXLOrder := []string{"PRXEEQ8", "PXI6XNI"}
			if len(pv9joxlMembers) != len(expectedPV9JOXLOrder) {
				t.Fatalf("team PV9JOXL: expected %d members, got %d", len(expectedPV9JOXLOrder), len(pv9joxlMembers))
			}
			for i, userID := range pv9joxlMembers {
				if userID != expectedPV9JOXLOrder[i] {
					t.Errorf("team PV9JOXL member at position %d: expected %s, got %s", i, expectedPV9JOXLOrder[i], userID)
				}
			}
		})

		t.Run("loadServiceMembers", func(t *testing.T) {
			ctx, pd := setup(t)

			if err := pd.UseTeamInterface("service"); err != nil {
				t.Fatalf("error setting team interface: %s", err)
			}
			if err := pd.LoadUsers(ctx); err != nil {
				t.Fatalf("error loading users: %s", err)
			}
			if err := pd.LoadTeams(ctx); err != nil {
				t.Fatalf("error loading teams: %s", err)
			}
			if err := pd.LoadTeamMembers(ctx); err != nil {
				t.Fatalf("error loading team members: %s", err)
			}
			members, err := store.UseQueries(ctx).ListGroupExtTeamMemberships(ctx)
			if err != nil {
				t.Fatalf("error loading team members: %s", err)
			}
			t.Logf("found %d team members, including services", len(members))
			assertJSON(t, members)
		})

		t.Run("loadServiceMembersPreservesOrder", func(t *testing.T) {
			ctx, pd := setup(t)

			if err := pd.UseTeamInterface("service"); err != nil {
				t.Fatalf("error setting team interface: %s", err)
			}
			if err := pd.LoadUsers(ctx); err != nil {
				t.Fatalf("error loading users: %s", err)
			}
			if err := pd.LoadTeams(ctx); err != nil {
				t.Fatalf("error loading teams: %s", err)
			}
			if err := pd.LoadTeamMembers(ctx); err != nil {
				t.Fatalf("error loading team members: %s", err)
			}
			members, err := store.UseQueries(ctx).ListGroupExtTeamMemberships(ctx)
			if err != nil {
				t.Fatalf("error loading team members: %s", err)
			}

			// Verify that the order matches the PagerDuty API responses
			// Services are loaded in order: P4XMRL3 (Endeavour), P3IIAF1 (Server under Jack's desk)

			// Extract service IDs in order
			var serviceIDs []string
			seen := make(map[string]bool)
			for _, member := range members {
				if !seen[member.ExtTeam.ID] {
					serviceIDs = append(serviceIDs, member.ExtTeam.ID)
					seen[member.ExtTeam.ID] = true
				}
			}

			expectedServiceOrder := []string{"P4XMRL3", "P3IIAF1"}
			if len(serviceIDs) != len(expectedServiceOrder) {
				t.Fatalf("expected %d services, got %d", len(expectedServiceOrder), len(serviceIDs))
			}
			for i, serviceID := range serviceIDs {
				if serviceID != expectedServiceOrder[i] {
					t.Errorf("service at position %d: expected %s, got %s", i, expectedServiceOrder[i], serviceID)
				}
			}
		})
	})

	t.Run("LoadSchedules", func(t *testing.T) {
		ctx, pd := setup(t)

		if err := pd.UseTeamInterface("team"); err != nil {
			t.Fatalf("error setting team interface: %s", err)
		}
		if err := pd.LoadUsers(ctx); err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		if err := pd.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}
		if err := pd.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}
		schedules, err := store.UseQueries(ctx).ListExtSchedulesV2(ctx)
		if err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}
		t.Logf("found %d schedules", len(schedules))
		assertJSON(t, schedules)
	})

	t.Run("LoadSchedulesRecordsMemberSkips", func(t *testing.T) {
		t.Parallel()
		ctx, pd := setup(t)

		if err := pd.UseTeamInterface("team"); err != nil {
			t.Fatalf("error setting team interface: %s", err)
		}
		// Load users — P1VTA5W appears in the schedule layer fixture but is absent from
		// users.json, so it won't be in ext_users and should produce a member skip.
		if err := pd.LoadUsers(ctx); err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		if err := pd.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}
		if err := pd.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}

		skips, err := store.UseQueries(ctx).ListRotationMemberSkips(ctx)
		if err != nil {
			t.Fatalf("error listing rotation member skips: %s", err)
		}

		if len(skips) != 1 {
			t.Fatalf("expected 1 skip record, got %d", len(skips))
		}
		skip := skips[0]
		if skip.UserID != "P1VTA5W" {
			t.Errorf("skip.UserID: expected %q, got %q", "P1VTA5W", skip.UserID)
		}
		if skip.RotationID != "P3BRVNT" {
			t.Errorf("skip.RotationID: expected %q, got %q", "P3BRVNT", skip.RotationID)
		}
		if skip.ScheduleName != "CS-on-call" {
			t.Errorf("skip.ScheduleName: expected %q, got %q", "CS-on-call", skip.ScheduleName)
		}
		if skip.Reason != "missing_fh_user" {
			t.Errorf("skip.Reason: expected %q, got %q", "missing_fh_user", skip.Reason)
		}
	})

	t.Run("LoadSchedulesDailyRestrictionCrossingMidnight", func(t *testing.T) {
		t.Parallel()
		ctx, pd := setup(t)

		if err := pd.UseTeamInterface("team"); err != nil {
			t.Fatalf("error setting team interface: %s", err)
		}
		if err := pd.LoadUsers(ctx); err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		if err := pd.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}
		if err := pd.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}

		// Layer PC1DX4O has a daily_restriction starting at 17:00:00 with duration 36000s (10h),
		// which means it ends at 03:00:00 the NEXT day. Each restriction's EndDay should be
		// the day after its StartDay.
		restrictions, err := store.UseQueries(ctx).ListExtRotationRestrictions(ctx, "PC1DX4O")
		if err != nil {
			t.Fatalf("error loading restrictions: %s", err)
		}

		if len(restrictions) != 7 {
			t.Fatalf("expected 7 daily restrictions (one per day), got %d", len(restrictions))
		}

		// Expected: each restriction's EndDay should be the next day
		expectedDays := []struct{ start, end string }{
			{"sunday", "monday"},
			{"monday", "tuesday"},
			{"tuesday", "wednesday"},
			{"wednesday", "thursday"},
			{"thursday", "friday"},
			{"friday", "saturday"},
			{"saturday", "sunday"},
		}

		for i, r := range restrictions {
			if r.StartDay != expectedDays[i].start {
				t.Errorf("restriction %d: expected StartDay %q, got %q", i, expectedDays[i].start, r.StartDay)
			}
			if r.EndDay != expectedDays[i].end {
				t.Errorf("restriction %d: expected EndDay %q, got %q", i, expectedDays[i].end, r.EndDay)
			}
			if r.StartTime != "17:00:00" {
				t.Errorf("restriction %d: expected StartTime 17:00:00, got %s", i, r.StartTime)
			}
			if r.EndTime != "03:00:00" {
				t.Errorf("restriction %d: expected EndTime 03:00:00, got %s", i, r.EndTime)
			}
		}
	})

	t.Run("LoadSchedulesSkipsEndedLayers", func(t *testing.T) {
		t.Parallel()
		ctx, pd := setup(t)

		if err := pd.UseTeamInterface("team"); err != nil {
			t.Fatalf("error setting team interface: %s", err)
		}
		if err := pd.LoadUsers(ctx); err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		if err := pd.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}
		if err := pd.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}

		// Jack's schedule (P85QTXZ) has two layers in the fixture: PDEADPL with a past
		// `end` timestamp and PE2BA4Y still active. Only the active one should become a rotation.
		rotations, err := store.UseQueries(ctx).ListExtRotationsByScheduleID(ctx, "P85QTXZ")
		if err != nil {
			t.Fatalf("error loading rotations: %s", err)
		}
		if len(rotations) != 1 {
			t.Fatalf("expected 1 rotation for P85QTXZ (ended layer should be filtered), got %d", len(rotations))
		}
		if rotations[0].ID != "PE2BA4Y" {
			t.Errorf("expected active rotation PE2BA4Y, got %s", rotations[0].ID)
		}

		if _, err := store.UseQueries(ctx).GetExtRotation(ctx, "PDEADPL"); err == nil {
			t.Error("ended layer PDEADPL should not be imported as a rotation")
		}
	})

	t.Run("LoadSchedulesPreservesOriginalVirtualStart", func(t *testing.T) {
		t.Parallel()
		ctx, pd := setup(t)

		if err := pd.UseTeamInterface("team"); err != nil {
			t.Fatalf("error setting team interface: %s", err)
		}
		if err := pd.LoadUsers(ctx); err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		if err := pd.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}
		if err := pd.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}

		r, err := store.UseQueries(ctx).GetExtRotation(ctx, "PE2BA4Y")
		if err != nil {
			t.Fatalf("error loading rotation PE2BA4Y: %s", err)
		}

		if want := "2023-06-02T14:00:00-07:00"; r.StartTime != want {
			t.Errorf("StartTime: got %q, want %q (original virtual_start should pass through unchanged)", r.StartTime, want)
		}
	})

	t.Run("LoadSchedulesGoldenPathPreservesPagerDutyOrdering", func(t *testing.T) {
		t.Parallel()
		ctx, pd := setup(t)

		if err := pd.UseTeamInterface("team"); err != nil {
			t.Fatalf("error setting team interface: %s", err)
		}
		if err := pd.LoadUsers(ctx); err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		if err := pd.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}
		if err := pd.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}

		r, err := store.UseQueries(ctx).GetExtRotation(ctx, "PROT01")
		if err != nil {
			t.Fatalf("error loading rotation PROT01: %s", err)
		}

		if want := "2026-06-02T12:00:00-07:00"; r.StartTime != want {
			t.Errorf("StartTime: got %q, want %q", r.StartTime, want)
		}
		if want := "weekly"; r.Strategy != want {
			t.Errorf("Strategy: got %q, want %q", r.Strategy, want)
		}
		if want := "tuesday"; r.HandoffDay != want {
			t.Errorf("HandoffDay: got %q, want %q", r.HandoffDay, want)
		}
		if want := "12:00:00"; r.HandoffTime != want {
			t.Errorf("HandoffTime: got %q, want %q", r.HandoffTime, want)
		}

		members, err := store.UseQueries(ctx).ListExtRotationMembers(ctx, "PROT01")
		if err != nil {
			t.Fatalf("error loading rotation members: %s", err)
		}

		expectedOrder := []string{"PU01A", "PU01B", "PU01C", "PU01D", "PU01E", "PU01F", "PU01G"}
		if len(members) != len(expectedOrder) {
			t.Fatalf("rotation PROT01: expected %d members, got %d", len(expectedOrder), len(members))
		}
		for i, member := range members {
			if member.UserID != expectedOrder[i] {
				t.Errorf("rotation PROT01 member at position %d: expected %s, got %s", i, expectedOrder[i], member.UserID)
			}
			if member.MemberOrder != int64(i) {
				t.Errorf("rotation PROT01 member at position %d: expected order %d, got %d", i, i, member.MemberOrder)
			}
		}
	})

	t.Run("LoadSchedulesSkipsScheduleWithMissingTeam", func(t *testing.T) {
		t.Parallel()
		ctx, pd := setup(t)

		if err := pd.UseTeamInterface("team"); err != nil {
			t.Fatalf("error setting team interface: %s", err)
		}
		if err := pd.LoadUsers(ctx); err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		if err := pd.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}

		// Simulate user not selecting any teams for import — delete all teams
		if err := store.UseQueries(ctx).DeleteExtTeamUnimported(ctx); err != nil {
			t.Fatalf("error deleting unimported teams: %s", err)
		}

		// LoadSchedules should not error — it should skip schedules whose teams are missing
		if err := pd.LoadSchedules(ctx); err != nil {
			t.Fatalf("expected no error loading schedules with missing teams, got: %s", err)
		}

		schedules, err := store.UseQueries(ctx).ListExtSchedulesV2(ctx)
		if err != nil {
			t.Fatalf("error listing schedules: %s", err)
		}
		if len(schedules) != 0 {
			t.Errorf("expected 0 schedules when all teams are missing, got %d", len(schedules))
		}
	})

	t.Run("LoadSchedulesPreservesMemberOrder", func(t *testing.T) {
		ctx, pd := setup(t)

		if err := pd.UseTeamInterface("team"); err != nil {
			t.Fatalf("error setting team interface: %s", err)
		}
		if err := pd.LoadUsers(ctx); err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		if err := pd.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}
		if err := pd.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}

		// Check that rotation members are ordered correctly
		// For rotation PSQ0VRL (which was a layer), we expect users in order: PXI6XNI, P2C9LBA
		members, err := store.UseQueries(ctx).ListExtRotationMembers(ctx, "PSQ0VRL")
		if err != nil {
			t.Fatalf("error loading rotation members: %s", err)
		}

		expectedOrder := []string{"PXI6XNI", "P2C9LBA"}
		if len(members) != len(expectedOrder) {
			t.Fatalf("rotation PSQ0VRL: expected %d members, got %d", len(expectedOrder), len(members))
		}
		for i, member := range members {
			if member.UserID != expectedOrder[i] {
				t.Errorf("rotation PSQ0VRL member at position %d: expected %s, got %s", i, expectedOrder[i], member.UserID)
			}
			if member.MemberOrder != int64(i) {
				t.Errorf("rotation PSQ0VRL member at position %d: expected order %d, got %d", i, i, member.MemberOrder)
			}
		}
	})

	t.Run("TerraformOutputMatchesScheduleOrder", func(t *testing.T) {
		ctx, pd := setup(t)

		if err := pd.UseTeamInterface("team"); err != nil {
			t.Fatalf("error setting team interface: %s", err)
		}
		if err := pd.LoadUsers(ctx); err != nil {
			t.Fatalf("error loading users: %s", err)
		}
		if err := pd.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}
		if err := pd.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}

		// First, verify the expected order from the database
		// For rotation PSQ0VRL (which was a layer), we expect users in order: PXI6XNI, P2C9LBA
		members, err := store.UseQueries(ctx).ListExtRotationMembers(ctx, "PSQ0VRL")
		if err != nil {
			t.Fatalf("error loading rotation members: %s", err)
		}

		expectedOrder := []string{"PXI6XNI", "P2C9LBA"}
		if len(members) != len(expectedOrder) {
			t.Fatalf("rotation PSQ0VRL: expected %d members, got %d", len(expectedOrder), len(members))
		}

		// Verify database order
		for i, member := range members {
			if member.UserID != expectedOrder[i] {
				t.Errorf("rotation PSQ0VRL member at position %d: expected %s, got %s", i, expectedOrder[i], member.UserID)
			}
			if member.MemberOrder != int64(i) {
				t.Errorf("rotation PSQ0VRL member at position %d: expected order %d, got %d", i, i, member.MemberOrder)
			}
		}

		// Mark all teams for import to ensure they appear in Terraform
		teams, err := store.UseQueries(ctx).ListExtTeams(ctx)
		if err != nil {
			t.Fatalf("error listing teams: %s", err)
		}
		for _, team := range teams {
			if err := store.UseQueries(ctx).MarkExtTeamToImport(ctx, team.ID); err != nil {
				t.Fatalf("error marking team for import: %s", err)
			}
		}

		// Create FireHydrant teams for linking
		for _, team := range teams {
			fhTeamID := fmt.Sprintf("fh-team-%s", team.ID)
			if err := store.UseQueries(ctx).InsertFhTeam(ctx, store.InsertFhTeamParams{
				ID:   fhTeamID,
				Name: team.Name,
				Slug: fmt.Sprintf("fh-team-%s", team.ID),
			}); err != nil {
				t.Fatalf("error creating FireHydrant team: %s", err)
			}
			if err := store.UseQueries(ctx).LinkExtTeam(ctx, store.LinkExtTeamParams{
				ID:       team.ID,
				FhTeamID: sql.NullString{String: fhTeamID, Valid: true},
			}); err != nil {
				t.Fatalf("error linking team: %s", err)
			}
		}

		// Mark all users for import to ensure they appear in Terraform
		users, err := store.UseQueries(ctx).ListExtUsers(ctx)
		if err != nil {
			t.Fatalf("error listing users: %s", err)
		}
		for _, user := range users {
			fhUserID := fmt.Sprintf("fh-user-%s", user.ID)
			if err := store.UseQueries(ctx).InsertFhUser(ctx, store.InsertFhUserParams{
				ID:    fhUserID,
				Email: user.Email,
				Name:  user.Name,
			}); err != nil {
				t.Fatalf("error creating FireHydrant user: %s", err)
			}
			if err := store.UseQueries(ctx).LinkExtUser(ctx, store.LinkExtUserParams{
				ID:       user.ID,
				FhUserID: sql.NullString{String: fhUserID, Valid: true},
			}); err != nil {
				t.Fatalf("error linking user: %s", err)
			}
		}

		// Generate Terraform file
		tfFile := filepath.Join(t.TempDir(), "schedule_order_verification.tf")
		tfr, err := tfrender.New(tfFile)
		if err != nil {
			t.Fatalf("error creating Terraform renderer: %s", err)
		}

		if err := tfr.Write(ctx); err != nil {
			t.Fatalf("error writing Terraform file: %s", err)
		}

		// Read the generated Terraform file
		tfContent, err := os.ReadFile(tfFile)
		if err != nil {
			t.Fatalf("error reading Terraform file: %s", err)
		}

		tfContentStr := string(tfContent)
		t.Logf("Generated Terraform file:\n%s", tfContentStr)

		// Extract the rotation resources for P3D7DLW-PSQ0VRL
		// Look for the rotation resource that contains the expected member order.
		// The regex matches from the resource declaration to a closing brace at the start of a line
		// (top-level block boundary), which correctly spans nested blocks like members { ... }.
		rotationRegex := regexp.MustCompile(`resource "firehydrant_rotation" "[^"]*" \{[\s\S]*?\n\}`)
		rotationMatches := rotationRegex.FindAllString(tfContentStr, -1)

		if len(rotationMatches) == 0 {
			t.Fatal("No firehydrant_rotation resources found in Terraform file")
		}

		// Find the rotation that contains our expected users
		// Looking for the rotation that contains both acme_success_eng (P2C9LBA) and acme_eng (PXI6XNI)
		var targetRotation string
		for _, rotation := range rotationMatches {
			if strings.Contains(rotation, "acme_success_eng") && strings.Contains(rotation, "acme_eng") {
				targetRotation = rotation
				break
			}
		}

		if targetRotation == "" {
			t.Fatal("No rotation found containing both acme_success_eng and acme_eng users")
		}

		t.Logf("Found target rotation:\n%s", targetRotation)

		// Extract user references from the nested members blocks:
		//   members {
		//     user_id = data.firehydrant_user.<slug>.id
		//   }
		userRefRegex := regexp.MustCompile(`data\.firehydrant_user\.([^\.]+)\.id`)
		userRefMatches := userRefRegex.FindAllStringSubmatch(targetRotation, -1)

		if len(userRefMatches) != len(expectedOrder) {
			t.Errorf("Expected %d user references in Terraform, got %d", len(expectedOrder), len(userRefMatches))
		}

		// Convert the Terraform user slugs back to the original PagerDuty user IDs
		var terraformUserOrder []string
		for _, userRefMatch := range userRefMatches {
			terraformUserSlug := userRefMatch[1]
			// Map the Terraform slugs to PagerDuty user IDs
			switch terraformUserSlug {
			case "acme_eng":
				terraformUserOrder = append(terraformUserOrder, "PXI6XNI")
			case "acme_success_eng":
				terraformUserOrder = append(terraformUserOrder, "P2C9LBA")
			default:
				t.Errorf("Unknown user slug in Terraform: %s", terraformUserSlug)
			}
		}

		t.Logf("Expected order from PagerDuty API: %v", expectedOrder)
		t.Logf("Actual order in Terraform file: %v", terraformUserOrder)

		// Verify that the order matches
		if len(terraformUserOrder) != len(expectedOrder) {
			t.Errorf("Order length mismatch: expected %d, got %d", len(expectedOrder), len(terraformUserOrder))
		} else {
			for i, userID := range terraformUserOrder {
				if userID != expectedOrder[i] {
					t.Errorf("Order mismatch at position %d: expected %s, got %s", i, expectedOrder[i], userID)
				}
			}
		}

		// Additional verification: Check that the rotation resource contains the correct structure
		if !strings.Contains(targetRotation, "firehydrant_rotation") {
			t.Error("Rotation resource should contain firehydrant_rotation")
		}

		if !strings.Contains(targetRotation, "members") {
			t.Error("Rotation resource should contain members")
		}

		t.Logf("✅ Terraform output order verification completed successfully")
	})

	t.Run("LoadEscalationPolicies", func(t *testing.T) {
		t.Parallel()
		ctx, pd := setup(t)
		data := map[string]any{}

		// Load teams first since schedules reference teams
		if err := pd.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}

		if err := pd.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}

		if err := pd.LoadEscalationPolicies(ctx); err != nil {
			t.Fatalf("error loading escalation policies: %s", err)
		}

		ep, err := store.UseQueries(ctx).ListExtEscalationPolicies(ctx)
		if err != nil {
			t.Fatalf("error loading escalation policies: %s", err)
		}
		data["escalation_policies"] = ep

		data["escalation_policy_steps"] = []any{}
		data["escalation_policy_step_targets"] = []any{}
		for _, p := range ep {
			steps, err := store.UseQueries(ctx).ListExtEscalationPolicySteps(ctx, p.ID)
			if err != nil {
				t.Fatalf("error loading escalation policy steps: %s", err)
			}
			for _, s := range steps {
				data["escalation_policy_steps"] = append(data["escalation_policy_steps"].([]any), s)
				targets, err := store.UseQueries(ctx).ListExtEscalationPolicyStepTargets(ctx, s.ID)
				if err != nil {
					t.Fatalf("error loading escalation policy step targets for '%s': %s", s.ID, err)
				}
				for _, t := range targets {
					data["escalation_policy_step_targets"] = append(data["escalation_policy_step_targets"].([]any), t)
				}
			}
		}

		assertJSON(t, data)
	})

	t.Run("LoadEscalationPoliciesWithTargets", func(t *testing.T) {
		ctx, pd := setup(t)

		if err := pd.UseTeamInterface("team"); err != nil {
			t.Fatalf("error setting team interface: %s", err)
		}

		if err := pd.LoadTeams(ctx); err != nil {
			t.Fatalf("error loading teams: %s", err)
		}

		if err := pd.LoadSchedules(ctx); err != nil {
			t.Fatalf("error loading schedules: %s", err)
		}

		if err := pd.LoadEscalationPolicies(ctx); err != nil {
			t.Fatalf("error loading escalation policies: %s", err)
		}

		policies, err := store.UseQueries(ctx).ListExtEscalationPolicies(ctx)
		if err != nil {
			t.Fatalf("error loading escalation policies: %s", err)
		}

		// Policy shoudl have user and schedule targets
		var targetPolicy *store.ExtEscalationPolicy
		for _, policy := range policies {
			if policy.Name == "Endeavour" {
				targetPolicy = &policy
				break
			}
		}

		if targetPolicy == nil {
			t.Fatal("Could not find 'Endeavour' policy")
		}

		steps, err := store.UseQueries(ctx).ListExtEscalationPolicySteps(ctx, targetPolicy.ID)
		if err != nil {
			t.Fatalf("error loading escalation policy steps: %s", err)
		}

		if len(steps) != 1 {
			t.Fatalf("Expected 1 step, got %d", len(steps))
		}

		step := steps[0]
		targets, err := store.UseQueries(ctx).ListExtEscalationPolicyStepTargets(ctx, step.ID)
		if err != nil {
			t.Fatalf("error loading targets for step: %s", err)
		}

		if len(targets) != 1 {
			t.Fatalf("Expected 1 target for step, got %d", len(targets))
		}

		target := targets[0]
		if target.TargetType != store.TARGET_TYPE_USER {
			t.Errorf("Expected target type USER, got %s", target.TargetType)
		}
		if target.TargetID != "P4CMCAU" {
			t.Errorf("Expected user target ID 'P4CMCAU', got %s", target.TargetID)
		}

		if step.Timeout != "PT1M" {
			t.Errorf("Expected timeout 'PT1M', got %s", step.Timeout)
		}

		t.Logf("✅ PagerDuty escalation policy targeting verification completed successfully")
	})
}
