BEGIN TRANSACTION;

INSERT INTO fh_users VALUES('49ef2cda-ab4f-4599-852c-8cc2c8884523','John Smith','jsmith@example.com');
INSERT INTO fh_users VALUES('90c17208-46b4-4e82-b9b8-b8f5d8215a05','FireHydrant Demo','fh-demo@example.com');
INSERT INTO fh_users VALUES('66506894-ecbc-4034-b8e6-30851dabf5f3','FireHydrant Eng','fh-eng@example.com');
INSERT INTO fh_users VALUES('a8fc03aa-8443-4c76-819c-8b7242fec459','FireHydrant Success','fh-success@example.com');

INSERT INTO ext_users VALUES('e0a51be7-3c7e-407f-8678-292ab421f55f','John Smith','jsmith@example.com','49ef2cda-ab4f-4599-852c-8cc2c8884523', '[Opsgenie] e0a51be7-3c7e-407f-8678-292ab421f55f jsmith@example.com');
INSERT INTO ext_users VALUES('1dc37638-ab52-44f3-848e-a16bcc584fb7','FireHydrant Demo','fh-demo@example.com','90c17208-46b4-4e82-b9b8-b8f5d8215a05', '[Opsgenie] 1dc37638-ab52-44f3-848e-a16bcc584fb7 fh-demo@example.com');
INSERT INTO ext_users VALUES('9253cf00-6195-4123-a9a6-f9f1e25718d8','FireHydrant Eng','fh-eng@example.com','66506894-ecbc-4034-b8e6-30851dabf5f3', '[Opsgenie] 9253cf00-6195-4123-a9a6-f9f1e25718d8 fh-eng@example.com');
INSERT INTO ext_users VALUES('e94e17aa-418c-44f7-8e47-1eaebf6b5343','FireHydrant Success','fh-success@example.com','a8fc03aa-8443-4c76-819c-8b7242fec459', '[Opsgenie] e94e17aa-418c-44f7-8e47-1eaebf6b5343 fh-success@example.com');

INSERT INTO fh_teams VALUES('8c465512-b0b4-47df-ba59-735574bc4dde','Alerting','alerting');
INSERT INTO fh_teams VALUES('d98aa7e2-9b38-41a8-b5de-49743c3b9ac2','Assign Product On-call','on-call-gameday');
INSERT INTO fh_teams VALUES('20c766e5-318a-4acc-a8f9-660e824e50f8','Customer Success and Support','customer-success-and-support');


INSERT INTO ext_teams VALUES('946bf740-0497-4d5d-b31f-23a6e55a2719','AJ Team','aj-team',NULL,0,1,'');

INSERT INTO ext_memberships VALUES('9253cf00-6195-4123-a9a6-f9f1e25718d8','946bf740-0497-4d5d-b31f-23a6e55a2719');
INSERT INTO ext_memberships VALUES('1dc37638-ab52-44f3-848e-a16bcc584fb7','946bf740-0497-4d5d-b31f-23a6e55a2719');
INSERT INTO ext_memberships VALUES('e0a51be7-3c7e-407f-8678-292ab421f55f','946bf740-0497-4d5d-b31f-23a6e55a2719');
INSERT INTO ext_memberships VALUES('e94e17aa-418c-44f7-8e47-1eaebf6b5343','946bf740-0497-4d5d-b31f-23a6e55a2719');

INSERT INTO ext_schedules VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-b1103233-600f-433c-bbdc-5269ad010255','AJ Team_schedule - Rota3','(Rota3)','America/Los_Angeles','custom','PT2H','','15:00:00','friday');
INSERT INTO ext_schedules VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-2b3ba1f8-5df9-4af7-a1ac-73f90bd30b2d','AJ Team_schedule - Daytime rotation','(Daytime rotation)','America/Los_Angeles','weekly','','','07:00:00','monday');
INSERT INTO ext_schedules VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-9b488cc6-efa0-44f3-a432-b913acab9147','AJ Team_schedule - Nighttime rotation','(Nighttime rotation)','America/Los_Angeles','daily','','','15:00:00','wednesday');

INSERT INTO ext_schedule_restrictions VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-2b3ba1f8-5df9-4af7-a1ac-73f90bd30b2d','0','08:00:00','monday','18:00:00','monday');
INSERT INTO ext_schedule_restrictions VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-2b3ba1f8-5df9-4af7-a1ac-73f90bd30b2d','1','08:00:00','tuesday','18:00:00','tuesday');
INSERT INTO ext_schedule_restrictions VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-2b3ba1f8-5df9-4af7-a1ac-73f90bd30b2d','2','08:00:00','wednesday','18:00:00','wednesday');
INSERT INTO ext_schedule_restrictions VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-2b3ba1f8-5df9-4af7-a1ac-73f90bd30b2d','3','08:00:00','thursday','18:00:00','thursday');
INSERT INTO ext_schedule_restrictions VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-2b3ba1f8-5df9-4af7-a1ac-73f90bd30b2d','4','08:00:00','friday','18:00:00','friday');
INSERT INTO ext_schedule_restrictions VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-9b488cc6-efa0-44f3-a432-b913acab9147','0','18:00:00','sunday','08:00:00','monday');
INSERT INTO ext_schedule_restrictions VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-9b488cc6-efa0-44f3-a432-b913acab9147','1','18:00:00','monday','08:00:00','tuesday');
INSERT INTO ext_schedule_restrictions VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-9b488cc6-efa0-44f3-a432-b913acab9147','2','18:00:00','tuesday','08:00:00','wednesday');
INSERT INTO ext_schedule_restrictions VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-9b488cc6-efa0-44f3-a432-b913acab9147','3','18:00:00','wednesday','08:00:00','thursday');
INSERT INTO ext_schedule_restrictions VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-9b488cc6-efa0-44f3-a432-b913acab9147','4','18:00:00','thursday','08:00:00','friday');
INSERT INTO ext_schedule_restrictions VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-9b488cc6-efa0-44f3-a432-b913acab9147','5','18:00:00','friday','08:00:00','saturday');
INSERT INTO ext_schedule_restrictions VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-9b488cc6-efa0-44f3-a432-b913acab9147','6','18:00:00','saturday','08:00:00','sunday');

INSERT INTO ext_schedule_teams VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-b1103233-600f-433c-bbdc-5269ad010255','946bf740-0497-4d5d-b31f-23a6e55a2719');
INSERT INTO ext_schedule_teams VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-2b3ba1f8-5df9-4af7-a1ac-73f90bd30b2d','946bf740-0497-4d5d-b31f-23a6e55a2719');
INSERT INTO ext_schedule_teams VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-9b488cc6-efa0-44f3-a432-b913acab9147','946bf740-0497-4d5d-b31f-23a6e55a2719');


INSERT INTO ext_schedule_members VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-b1103233-600f-433c-bbdc-5269ad010255','9253cf00-6195-4123-a9a6-f9f1e25718d8');
INSERT INTO ext_schedule_members VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-b1103233-600f-433c-bbdc-5269ad010255','1dc37638-ab52-44f3-848e-a16bcc584fb7');
INSERT INTO ext_schedule_members VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-b1103233-600f-433c-bbdc-5269ad010255','e94e17aa-418c-44f7-8e47-1eaebf6b5343');
INSERT INTO ext_schedule_members VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-2b3ba1f8-5df9-4af7-a1ac-73f90bd30b2d','9253cf00-6195-4123-a9a6-f9f1e25718d8');
INSERT INTO ext_schedule_members VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-2b3ba1f8-5df9-4af7-a1ac-73f90bd30b2d','1dc37638-ab52-44f3-848e-a16bcc584fb7');
INSERT INTO ext_schedule_members VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-2b3ba1f8-5df9-4af7-a1ac-73f90bd30b2d','e94e17aa-418c-44f7-8e47-1eaebf6b5343');
INSERT INTO ext_schedule_members VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-2b3ba1f8-5df9-4af7-a1ac-73f90bd30b2d','e0a51be7-3c7e-407f-8678-292ab421f55f');
INSERT INTO ext_schedule_members VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-9b488cc6-efa0-44f3-a432-b913acab9147','9253cf00-6195-4123-a9a6-f9f1e25718d8');
INSERT INTO ext_schedule_members VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-9b488cc6-efa0-44f3-a432-b913acab9147','1dc37638-ab52-44f3-848e-a16bcc584fb7');
INSERT INTO ext_schedule_members VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-9b488cc6-efa0-44f3-a432-b913acab9147','e94e17aa-418c-44f7-8e47-1eaebf6b5343');
INSERT INTO ext_schedule_members VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-9b488cc6-efa0-44f3-a432-b913acab9147','e0a51be7-3c7e-407f-8678-292ab421f55f');

INSERT INTO ext_escalation_policies VALUES('880ec24e-58db-441b-9681-2cb527bd24b2','AJ Team_escalation','','946bf740-0497-4d5d-b31f-23a6e55a2719',0,NULL,'','','',1);

INSERT INTO ext_escalation_policy_steps VALUES('880ec24e-58db-441b-9681-2cb527bd24b2-0','880ec24e-58db-441b-9681-2cb527bd24b2',0,'PT1M');

INSERT INTO ext_escalation_policy_step_targets VALUES('880ec24e-58db-441b-9681-2cb527bd24b2-0','OnCallSchedule','8ab5a183-8ef5-47db-9de0-56663cfbae7c-b1103233-600f-433c-bbdc-5269ad010255');
INSERT INTO ext_escalation_policy_step_targets VALUES('880ec24e-58db-441b-9681-2cb527bd24b2-0','OnCallSchedule','8ab5a183-8ef5-47db-9de0-56663cfbae7c-2b3ba1f8-5df9-4af7-a1ac-73f90bd30b2d');
INSERT INTO ext_escalation_policy_step_targets VALUES('880ec24e-58db-441b-9681-2cb527bd24b2-0','OnCallSchedule','8ab5a183-8ef5-47db-9de0-56663cfbae7c-9b488cc6-efa0-44f3-a432-b913acab9147');

COMMIT;
