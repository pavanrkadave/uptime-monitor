CREATE TABLE IF NOT EXISTS monitors (
                                        id BIGSERIAL PRIMARY KEY,
                                        url VARCHAR(255) NOT NULL,
                                        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ping_results (
                                            id BIGSERIAL PRIMARY KEY,
                                            monitor_id BIGINT NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
                                            is_up BOOLEAN NOT NULL,
                                            status_code INT NOT NULL,
                                            duration_ms INT NOT NULL,
                                            error_message TEXT,
                                            checked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);