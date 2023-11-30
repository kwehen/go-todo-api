CREATE TABLE "tasks" (
    "id" int4 NOT NULL DEFAULT nextval('id_seq'::regclass),
    "task" varchar(256) NOT NULL,
    "urgency" varchar(10) NOT NULL DEFAULT 'Low'::character varying,
    "hours" numeric(2,1),
    "completed" bool NOT NULL DEFAULT false,
    PRIMARY KEY ("id")
);