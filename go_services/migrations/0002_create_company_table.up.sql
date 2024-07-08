CREATE TABLE companies (
    company_name VARCHAR(255) PRIMARY KEY,
    docker_image_name VARCHAR(255) NOT NULL,
    docker_image_id VARCHAR(255) NOT NULL,
    pull_date TIMESTAMP NOT NULL
);

INSERT INTO companies (company_name, docker_image_name, docker_image_id, pull_date) VALUES
('amazon', 'harrisonll/jc_worker:test', '', NOW()),
('meta', 'harrisonll/jc_worker:test', '', NOW());