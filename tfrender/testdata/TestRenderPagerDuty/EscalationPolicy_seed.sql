BEGIN TRANSACTION;

INSERT INTO fh_users VALUES('35b5390f-d134-4bc6-966d-0b4048788b62','Alice Bob','alice.bob@example.com');

INSERT INTO ext_users VALUES('PRXEEQ8','Alice Bob','alice.bob@example.com','35b5390f-d134-4bc6-966d-0b4048788b62');
INSERT INTO ext_users VALUES('PXI6XNI','Engineering Shared Account','eng@example.com','35b5390f-d134-4bc6-966d-0b4048788b62');

INSERT INTO fh_teams VALUES('f159b173-1ffd-41ac-9254-ce8ec1142267','🐴 Cowboy Coders','cowboy-coders');

INSERT INTO ext_teams VALUES('PV9JOXL','team-rocket','team-rocket','f159b173-1ffd-41ac-9254-ce8ec1142267',0,1);

INSERT INTO ext_memberships VALUES('PRXEEQ8','PV9JOXL');
INSERT INTO ext_memberships VALUES('PXI6XNI','PV9JOXL');

INSERT INTO ext_schedules VALUES('PGR96WL-PR3J6XJ','🐴 @alice.bob is always on call - Layer 1','Always 😭 (Layer 1)','America/Los_Angeles','weekly','','','12:00:00','friday');

INSERT INTO ext_schedule_teams VALUES('PGR96WL-PR3J6XJ','PV9JOXL');

INSERT INTO ext_schedule_members VALUES('PGR96WL-PR3J6XJ','PRXEEQ8');

INSERT INTO ext_escalation_policies VALUES('P2D2WR1','🐴 @alice.bob Test Service-ep','',NULL,0,NULL,'','',1);
INSERT INTO ext_escalation_policies VALUES('PS6ITO0','🐴 Notify @alice.bob','',NULL,0,NULL,'','',1);

INSERT INTO ext_escalation_policy_steps VALUES('PKQDFZH','P2D2WR1',0,'PT30M');
INSERT INTO ext_escalation_policy_steps VALUES('P08T67P','PS6ITO0',0,'PT30M');

INSERT INTO ext_escalation_policy_step_targets VALUES('P08T67P','User','PXI6XNI');
INSERT INTO ext_escalation_policy_step_targets VALUES('PKQDFZH','OnCallSchedule','PGR96WL-PR3J6XJ');

COMMIT;
