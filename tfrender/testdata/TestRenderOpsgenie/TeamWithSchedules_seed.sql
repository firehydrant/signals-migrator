BEGIN TRANSACTION;

INSERT INTO fh_users VALUES('0946be55-ea20-4483-b9ab-617d5f0969e2','Admin Account','local@example.io');
INSERT INTO fh_users VALUES('e6009411-0015-43e3-815e-ca9db72f4088','Mika','mika@example.com');
INSERT INTO fh_users VALUES('4c3f28fa-b402-453c-9652-f014ecbe65a9', 'Kiran','kiran@example.com');
INSERT INTO fh_users VALUES('6c08bff2-98f6-4ee9-8de1-12202186d084','Jack T.','jack@example.com');
INSERT INTO fh_users VALUES('032a1f07-987e-4f76-8273-136e08e50baa', 'Wong','wong@example.com');

INSERT INTO ext_users VALUES('1dc37638-ab52-44f3-848e-a16bcc584fb7','Admin Account','admin@example.io','0946be55-ea20-4483-b9ab-617d5f0969e2','[Opsgenie] 1dc37638-ab52-44f3-848e-a16bcc584fb7 admin@example.io');
INSERT INTO ext_users VALUES('9253cf00-6195-4123-a9a6-f9f1e25718d8','Mika','mika@example.io','e6009411-0015-43e3-815e-ca9db72f4088','[Opsgenie] 9253cf00-6195-4123-a9a6-f9f1e25718d8 mika@example.io');
INSERT INTO ext_users VALUES('e94e17aa-418c-44f7-8e47-1eaebf6b5343','Kiran','kiran@example.com','4c3f28fa-b402-453c-9652-f014ecbe65a9','[Opsgenie] e94e17aa-418c-44f7-8e47-1eaebf6b5343 kiran@example.com');
INSERT INTO ext_users VALUES('d68757c4-5eec-4560-8c5b-91c463f87dd8','Jack T.','jack@example.com','6c08bff2-98f6-4ee9-8de1-12202186d084','[Opsgenie] d68757c4-5eec-4560-8c5b-91c463f87dd8 jack@example.com');
INSERT INTO ext_users VALUES('a13020ca-cb08-48e3-9403-bed181a22e72','Wong','wong@example.io','032a1f07-987e-4f76-8273-136e08e50baa','[Opsgenie] a13020ca-cb08-48e3-9403-bed181a22e72 wong@example.io');

INSERT INTO ext_teams VALUES('017e4326-abdf-4cd9-8cad-5f23bd7f4753','Wong Squad','wong-squad',NULL,0,1,'');
INSERT INTO ext_teams VALUES('3dd6b50f-28ed-4660-b982-606bfa6c4cf2','Platform','platform',NULL,0,1,'');
INSERT INTO ext_teams VALUES('946bf740-0497-4d5d-b31f-23a6e55a2719','AJ Team','aj-team',NULL,0,1,'');
INSERT INTO ext_teams VALUES('b7acbc33-9853-4150-8a4b-10156d9408c8','Customer Success','customer-success',NULL,0,1,'');
INSERT INTO ext_teams VALUES('e3436ab1-7547-4a47-a02a-36fee3dc91f9','noodlebrigade','noodlebrigade',NULL,0,1,'');
INSERT INTO ext_teams VALUES('f5a99a73-cdfb-49a2-af34-aeb05c59d937','Christine Test Team','christine-test-team',NULL,0,1,'');

INSERT INTO ext_memberships VALUES('e94e17aa-418c-44f7-8e47-1eaebf6b5343','946bf740-0497-4d5d-b31f-23a6e55a2719');
INSERT INTO ext_memberships VALUES('9253cf00-6195-4123-a9a6-f9f1e25718d8','946bf740-0497-4d5d-b31f-23a6e55a2719');
INSERT INTO ext_memberships VALUES('1dc37638-ab52-44f3-848e-a16bcc584fb7','946bf740-0497-4d5d-b31f-23a6e55a2719');
INSERT INTO ext_memberships VALUES('d68757c4-5eec-4560-8c5b-91c463f87dd8','e3436ab1-7547-4a47-a02a-36fee3dc91f9');

INSERT INTO ext_schedules VALUES('3fee43f2-02da-49be-ab50-c88ed13aecc3-b1b5f7f6-728b-47bd-af02-c3e1c33bf219','Customer Success_schedule - Rot1','(Rot1)','America/Los_Angeles','weekly','','','04:45:32','tuesday');
INSERT INTO ext_schedules VALUES('57c134f4-adeb-41f2-990b-8eebe7bcfcf2-cf71a41b-3261-4f0b-93ed-5a1d1c583db8','how - Rota1','(Rota1)','America/Chicago','weekly','','','13:00:00','tuesday');
INSERT INTO ext_schedules VALUES('6520320e-d58b-4ed3-b910-a14439e84d07-55b9bf1d-7964-4eca-aa5e-33c642175b4e','Chris Test Team_schedule - Rot1','(Rot1)','America/Los_Angeles','weekly','','','07:00:00','monday');
INSERT INTO ext_schedules VALUES('791aa2c3-431b-4fb0-ae83-8d2814034098-0685bec4-5dd5-4a26-861d-3532166c2675','Wong Team_schedule - First','(First)','America/Los_Angeles','daily','','','08:00:00','monday');
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

INSERT INTO ext_schedule_teams VALUES('3fee43f2-02da-49be-ab50-c88ed13aecc3-b1b5f7f6-728b-47bd-af02-c3e1c33bf219','b7acbc33-9853-4150-8a4b-10156d9408c8');
INSERT INTO ext_schedule_teams VALUES('791aa2c3-431b-4fb0-ae83-8d2814034098-0685bec4-5dd5-4a26-861d-3532166c2675','017e4326-abdf-4cd9-8cad-5f23bd7f4753');
INSERT INTO ext_schedule_teams VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-2b3ba1f8-5df9-4af7-a1ac-73f90bd30b2d','946bf740-0497-4d5d-b31f-23a6e55a2719');
INSERT INTO ext_schedule_teams VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-9b488cc6-efa0-44f3-a432-b913acab9147','946bf740-0497-4d5d-b31f-23a6e55a2719');

INSERT INTO ext_schedule_members VALUES('57c134f4-adeb-41f2-990b-8eebe7bcfcf2-cf71a41b-3261-4f0b-93ed-5a1d1c583db8','9253cf00-6195-4123-a9a6-f9f1e25718d8',0);
INSERT INTO ext_schedule_members VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-2b3ba1f8-5df9-4af7-a1ac-73f90bd30b2d','9253cf00-6195-4123-a9a6-f9f1e25718d8',0);
INSERT INTO ext_schedule_members VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-2b3ba1f8-5df9-4af7-a1ac-73f90bd30b2d','1dc37638-ab52-44f3-848e-a16bcc584fb7',1);
INSERT INTO ext_schedule_members VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-2b3ba1f8-5df9-4af7-a1ac-73f90bd30b2d','e94e17aa-418c-44f7-8e47-1eaebf6b5343',2);
INSERT INTO ext_schedule_members VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-9b488cc6-efa0-44f3-a432-b913acab9147','9253cf00-6195-4123-a9a6-f9f1e25718d8',0);
INSERT INTO ext_schedule_members VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-9b488cc6-efa0-44f3-a432-b913acab9147','1dc37638-ab52-44f3-848e-a16bcc584fb7',1);
INSERT INTO ext_schedule_members VALUES('8ab5a183-8ef5-47db-9de0-56663cfbae7c-9b488cc6-efa0-44f3-a432-b913acab9147','e94e17aa-418c-44f7-8e47-1eaebf6b5343',2);

COMMIT;
