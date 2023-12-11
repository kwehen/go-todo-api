CREATE TABLE completed (
    id SERIAL PRIMARY KEY,
    task varchar(256) NOT NULL,
    foreign key (task) references tasks (task)
);