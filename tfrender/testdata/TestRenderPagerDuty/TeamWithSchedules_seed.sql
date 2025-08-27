INSERT INTO fh_users VALUES('0946be55-ea20-4483-b9ab-617d5f0969e2','Admin Account','local@example.io');
INSERT INTO fh_users VALUES('e6009411-0015-43e3-815e-ca9db72f4088','Mika','mika@example.com');
INSERT INTO fh_users VALUES('4c3f28fa-b402-453c-9652-f014ecbe65a9', 'Kiran','kiran@example.com');
INSERT INTO fh_users VALUES('35b5390f-d134-4bc6-966d-0b4048788b62','Horse','horse@example.com');
INSERT INTO fh_users VALUES('6c08bff2-98f6-4ee9-8de1-12202186d084','Jack T.','jack@example.com');
INSERT INTO fh_users VALUES('032a1f07-987e-4f76-8273-136e08e50baa', 'Wong','wong@example.com');

INSERT INTO ext_users VALUES('PXI6XNI','Admin','local@example.io','0946be55-ea20-4483-b9ab-617d5f0969e2','');
INSERT INTO ext_users VALUES('P5A1XH2','Mika','mika@example.io','e6009411-0015-43e3-815e-ca9db72f4088','');
INSERT INTO ext_users VALUES('P8ZZ1ZB','Kiran','kiran@example.io','4c3f28fa-b402-453c-9652-f014ecbe65a9','');
INSERT INTO ext_users VALUES('PRXEEQ8','Horse','horse@example.io','35b5390f-d134-4bc6-966d-0b4048788b62','');
INSERT INTO ext_users VALUES('P4CMCAU','Jack T.','jack@example.io','6c08bff2-98f6-4ee9-8de1-12202186d084','');
INSERT INTO ext_users VALUES('P2C9LBA','Wong','wong@example.io','032a1f07-987e-4f76-8273-136e08e50baa','');

INSERT INTO fh_teams VALUES('47016143-6547-483a-b68a-5220b21681fd','AAAA IPv6 migration strategy','aaaa-ipv6-migration-strategy');
INSERT INTO fh_teams VALUES('f159b173-1ffd-41ac-9254-ce8ec1142267','🐴 Cowboy Coders','cowboy-coders');
INSERT INTO fh_teams VALUES('97d539b0-47a5-44f6-81e6-b6fcd98f23ac','Dunder Mifflin Scranton','dunder-mifflin-scranton');

INSERT INTO ext_teams VALUES('PT54U20','Jen','jen','47016143-6547-483a-b68a-5220b21681fd',0,1,'[PagerDuty] Jen https://pdt-apidocs.pagerduty.com/service-directory/PT54U20');
INSERT INTO ext_teams VALUES('PD2F80U','Jack Team','jack-team','97d539b0-47a5-44f6-81e6-b6fcd98f23ac',0,1,'[PagerDuty] Jack Team https://pdt-apidocs.pagerduty.com/service-directory/PD2F80U');
INSERT INTO ext_teams VALUES('PV9JOXL','🐴 Cowboy Coders','cowboy-coders','f159b173-1ffd-41ac-9254-ce8ec1142267',0,1,'[PagerDuty] 🐴 Cowboy Coders https://pdt-apidocs.pagerduty.com/service-directory/PV9JOXL');

INSERT INTO ext_memberships VALUES('PXI6XNI','PT54U20');
INSERT INTO ext_memberships VALUES('P8ZZ1ZB','PT54U20');
INSERT INTO ext_memberships VALUES('P2C9LBA','PT54U20');
INSERT INTO ext_memberships VALUES('P4CMCAU','PD2F80U');

INSERT INTO ext_schedules_v2 VALUES('schedule-jen-primary','Jen - primary','Primary on-call schedule for Jen team','America/Los_Angeles','PT54U20','pagerduty','P3D7DLW');
INSERT INTO ext_schedules_v2 VALUES('schedule-jack-oncall','Jack On-Call Schedule','On-call schedule for Jack team','America/Los_Angeles','PD2F80U','pagerduty','P85QTXZ');
INSERT INTO ext_schedules_v2 VALUES('schedule-horse-always','🐴 is always on call','Always on call schedule','America/Los_Angeles','PV9JOXL','pagerduty','PGR96WL');

INSERT INTO ext_rotations VALUES('rotation-jen-layer-2','schedule-jen-primary','Layer 2','(Layer 2)','custom','PT93600S','2024-04-10T20:39:29-07:00','16:00:00','wednesday',0);
INSERT INTO ext_rotations VALUES('rotation-jen-layer-1','schedule-jen-primary','Layer 1','(Layer 1)','custom','PT7200S','2024-04-10T20:39:29-07:00','16:00:00','wednesday',1);
INSERT INTO ext_rotations VALUES('rotation-jack-layer-1','schedule-jack-oncall','Layer 1',' (Layer 1)','weekly','','','14:00:00','friday',0);
INSERT INTO ext_rotations VALUES('rotation-horse-layer-1','schedule-horse-always','Layer 1','(Layer 1)','weekly','','','12:00:00','friday',0);

INSERT INTO ext_rotation_members VALUES('rotation-jen-layer-2','P8ZZ1ZB',0);
INSERT INTO ext_rotation_members VALUES('rotation-jen-layer-1','PXI6XNI',0);
INSERT INTO ext_rotation_members VALUES('rotation-jen-layer-1','P2C9LBA',1);
INSERT INTO ext_rotation_members VALUES('rotation-jack-layer-1','P4CMCAU',0);
INSERT INTO ext_rotation_members VALUES('rotation-horse-layer-1','PRXEEQ8',0);

INSERT INTO ext_rotation_restrictions VALUES('rotation-jen-layer-2','0','09:00:00','monday','17:00:00','friday');
INSERT INTO ext_rotation_restrictions VALUES('rotation-jen-layer-1','0-0','09:00:00','sunday','17:00:00','sunday');
INSERT INTO ext_rotation_restrictions VALUES('rotation-jen-layer-1','0-1','09:00:00','monday','17:00:00','monday');
INSERT INTO ext_rotation_restrictions VALUES('rotation-jen-layer-1','0-2','09:00:00','tuesday','17:00:00','tuesday');
INSERT INTO ext_rotation_restrictions VALUES('rotation-jen-layer-1','0-3','09:00:00','wednesday','17:00:00','wednesday');
INSERT INTO ext_rotation_restrictions VALUES('rotation-jen-layer-1','0-4','09:00:00','thursday','17:00:00','thursday');
INSERT INTO ext_rotation_restrictions VALUES('rotation-jen-layer-1','0-5','09:00:00','friday','17:00:00','friday');
INSERT INTO ext_rotation_restrictions VALUES('rotation-jen-layer-1','0-6','09:00:00','saturday','17:00:00','saturday');
