CREATE TABLE `users` (
                         `id` varchar(36) NOT NULL,
                         `created_at` datetime(3) DEFAULT NULL,
                         `updated_at` datetime(3) DEFAULT NULL,
                         `deleted_at` datetime(3) DEFAULT NULL,
                         `login` varchar(100),
                         `email` varchar(100),
                         `password_hash` longtext,
                         `timezone` varchar(50),
                         PRIMARY KEY (`id`),
                         KEY `idx_users_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `events` (
                          `id` varchar(36) NOT NULL,
                          `created_at` datetime(3) DEFAULT NULL,
                          `updated_at` datetime(3) DEFAULT NULL,
                          `deleted_at` datetime(3) DEFAULT NULL,
                          `title`  varchar(100),
                          `description` longtext,
                          `time` datetime(3) DEFAULT NULL,
                          `timezone`  varchar(50),
                          `duration` varchar(64) DEFAULT NULL,
                          `notes` longtext,
                          `user_id` varchar(36),
                          PRIMARY KEY (`id`),
                          KEY `idx_events_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


ALTER TABLE calendar.events
    ADD CONSTRAINT events_ibfk_1 FOREIGN KEY (user_id)
        REFERENCES calendar.users(id) ON DELETE CASCADE;