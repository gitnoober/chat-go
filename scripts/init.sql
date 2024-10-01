-- Create the database if it doesn't exist
CREATE DATABASE IF NOT EXISTS test;

-- Create a new MySQL user and grant them access to the 'test' database
CREATE USER IF NOT EXISTS 'newuser'@'%' IDENTIFIED BY 'newpassword';
GRANT ALL PRIVILEGES ON test.* TO 'newuser'@'%';
FLUSH PRIVILEGES;

-- Switch to the 'test' database
USE test;

-- Create the 'users' table if it doesn't exist
CREATE TABLE IF NOT EXISTS users (
    id INT NOT NULL AUTO_INCREMENT,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    profile_url VARCHAR(255) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY idx_email (email)
);