
CREATE TABLE calendar.events (
                                 id BINARY(16) NOT NULL,
                                 title VARCHAR(100) DEFAULT NULL,
                                 description VARCHAR(256) DEFAULT NULL,
                                 time TIMESTAMP NOT NULL,
                                 timezone VARCHAR(30) NOT NULL DEFAULT 'UTC',
                                 duration INT DEFAULT NULL,
                                 notes TEXT DEFAULT NULL,
                                 user_id BINARY(16) DEFAULT NULL,
                                 deleted_at TIMESTAMP NULL DEFAULT NULL,
                                 created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
                                 updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                                 PRIMARY KEY (id)
)
    ENGINE = INNODB,
CHARACTER SET utf8mb4,
COLLATE utf8mb4_0900_ai_ci;

ALTER TABLE calendar.events
    ADD CONSTRAINT events_ibfk_1 FOREIGN KEY (user_id)
        REFERENCES calendar.users(id) ON DELETE CASCADE;