-- +migrate Up
CREATE TABLE IF NOT EXISTS logs_caregivers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    caregiver_id UUID NOT NULL,
    log_type VARCHAR(10) NOT NULL CHECK (log_type IN ('clock_in', 'clock_out')),
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_logs_caregivers_caregiver_timestamp ON logs_caregivers (caregiver_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_logs_caregivers_type ON logs_caregivers (log_type);
CREATE INDEX IF NOT EXISTS idx_logs_caregivers_timestamp ON logs_caregivers (timestamp DESC);

-- +migrate Down
DROP TABLE IF EXISTS logs_caregivers;