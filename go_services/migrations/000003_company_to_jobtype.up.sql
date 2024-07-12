DROP TABLE IF EXISTS companies;
CREATE TABLE job_types (
    job_type_name VARCHAR(255) NOT NULL,
    company_name VARCHAR(255) NOT NULL,
    docker_image_name VARCHAR(255) NOT NULL,
    docker_image_id VARCHAR(255) NOT NULL,
    pull_date TIMESTAMP NOT NULL,
    PRIMARY KEY (job_type_name, company_name)
);

INSERT INTO job_types (job_type_name, company_name, docker_image_name, docker_image_id, pull_date)
VALUES 
('software engineer', 'amazon', 'harrisonll/jc_worker:test', '', CURRENT_TIMESTAMP),
('software engineer', 'meta', 'harrisonll/jc_worker:test', '', CURRENT_TIMESTAMP),
('machine learning engineer', 'amazon', 'harrisonll/jc_worker:test', '', CURRENT_TIMESTAMP),
('machine learning engineer', 'meta', 'harrisonll/jc_worker:test', '', CURRENT_TIMESTAMP),
('data scientist', 'amazon', 'harrisonll/jc_worker:test', '', CURRENT_TIMESTAMP),
('data scientist', 'meta', 'harrisonll/jc_worker:test', '', CURRENT_TIMESTAMP);