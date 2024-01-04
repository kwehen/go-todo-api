CREATE TABLE tasks (
    task_id SERIAL PRIMARY KEY,
    task varchar(256) NOT NULL,
    urgency varchar(10) NOT NULL DEFAULT 'Low',
    hours numeric(3,1), 
    completed boolean NOT NULL DEFAULT false,
    user_id varchar
);