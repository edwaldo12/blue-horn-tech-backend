-- +migrate Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE caregivers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name TEXT NOT NULL,
    email TEXT,
    phone TEXT,
    address TEXT,
    city TEXT,
    state TEXT,
    postal_code TEXT,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    caregiver_id UUID NOT NULL REFERENCES caregivers(id),
    client_id UUID NOT NULL REFERENCES clients(id),
    service_name TEXT NOT NULL,
    location_label TEXT,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('scheduled','in_progress','completed','cancelled','missed')),
    clock_in_at TIMESTAMPTZ,
    clock_in_lat DOUBLE PRECISION,
    clock_in_long DOUBLE PRECISION,
    clock_out_at TIMESTAMPTZ,
    clock_out_lat DOUBLE PRECISION,
    clock_out_long DOUBLE PRECISION,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE schedule_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL CHECK (status IN ('pending','completed','not_completed')),
    not_completed_reason TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE auth_clients (
    id TEXT PRIMARY KEY,
    secret_hash TEXT NOT NULL,
    description TEXT,
    caregiver_id UUID NOT NULL REFERENCES caregivers(id),
    scopes TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO caregivers (id, name, email)
VALUES 
    ('c2d1bb61-8d67-4db5-9e59-4c2c16f7d4f2', 'Louis Carewell', 'louis@careviah.com');

INSERT INTO clients (id, full_name, email, phone, address, city, state, postal_code, latitude, longitude)
VALUES
    ('4f1bbd73-df5e-4f3a-a59c-2d1fe15f0aaf', 'Melisa Adam', 'melisa@gmail.com', '+441232212333', '4333 Willison Street', 'Minneapolis', 'MN', '55415', 44.9727, -93.2354);

INSERT INTO schedules (
    id, caregiver_id, client_id, service_name, location_label, start_time, end_time, status,
    clock_in_at, clock_in_lat, clock_in_long, clock_out_at, clock_out_lat, clock_out_long, notes
) VALUES
    ('31b308fa-71d3-4f12-a6ec-cbb7d1a33c16', 'c2d1bb61-8d67-4db5-9e59-4c2c16f7d4f2', '4f1bbd73-df5e-4f3a-a59c-2d1fe15f0aaf',
     'Service Name A', 'Casa Grande Apartment', NOW() + INTERVAL '2 hour', NOW() + INTERVAL '3 hour',
     'scheduled', NULL, NULL, NULL, NULL, NULL, NULL, 'Remember to review medication plan.'),
    ('0f200db3-a712-4635-9cfa-8df4839f1adb', 'c2d1bb61-8d67-4db5-9e59-4c2c16f7d4f2', '4f1bbd73-df5e-4f3a-a59c-2d1fe15f0aaf',
     'Service Name A', 'Casa Grande Apartment', NOW() - INTERVAL '1 hour', NOW() + INTERVAL '0.5 hour',
     'in_progress', NOW() - INTERVAL '55 minute', 44.972700, -93.235400, NULL, NULL, NULL, 'Ongoing visit in progress.'),
    ('9b8c7d6e-5f4a-3210-9876-543210fedcba', 'c2d1bb61-8d67-4db5-9e59-4c2c16f7d4f2', '4f1bbd73-df5e-4f3a-a59c-2d1fe15f0aaf',
     'Home Care Service', 'Minnesota City Residence', NOW() - INTERVAL '30 minute', NOW() + INTERVAL '1.5 hour',
     'in_progress', NOW() - INTERVAL '25 minute', 44.1234, -93.5678, NULL, NULL, NULL, 'Currently providing home care services.'),
    ('a8a6d494-3b35-4536-b4f4-8023c2f13914', 'c2d1bb61-8d67-4db5-9e59-4c2c16f7d4f2', '4f1bbd73-df5e-4f3a-a59c-2d1fe15f0aaf',
     'Service Name A', 'Casa Grande Apartment', NOW() - INTERVAL '5 hour', NOW() - INTERVAL '4 hour',
     'completed', NOW() - INTERVAL '4 hour 45 minute', 44.972600, -93.235300, NOW() - INTERVAL '4 hour', 44.972800, -93.235100, 'All tasks completed on time.'),
    ('b704cb79-bcf4-4b52-a0b4-d5a2a0e12d62', 'c2d1bb61-8d67-4db5-9e59-4c2c16f7d4f2', '4f1bbd73-df5e-4f3a-a59c-2d1fe15f0aaf',
     'Service Name A', 'Casa Grande Apartment', NOW() - INTERVAL '7 hour', NOW() - INTERVAL '6 hour',
     'cancelled', NULL, NULL, NULL, NULL, NULL, NULL, 'Client rescheduled to next week.');

INSERT INTO schedule_tasks (id, schedule_id, title, description, status, sort_order) VALUES
    (gen_random_uuid(), '31b308fa-71d3-4f12-a6ec-cbb7d1a33c16', 'Activity Name A', 'Review medication checklist.', 'pending', 1),
    (gen_random_uuid(), '31b308fa-71d3-4f12-a6ec-cbb7d1a33c16', 'Activity Name B', 'Assist with morning routine.', 'pending', 2),
    (gen_random_uuid(), '31b308fa-71d3-4f12-a6ec-cbb7d1a33c16', 'Activity Name C', 'Prepare breakfast.', 'pending', 3),

    (gen_random_uuid(), '0f200db3-a712-4635-9cfa-8df4839f1adb', 'Activity Name A', 'Review medication checklist.', 'pending', 1),
    (gen_random_uuid(), '0f200db3-a712-4635-9cfa-8df4839f1adb', 'Activity Name B', 'Assist with morning routine.', 'pending', 2),
    (gen_random_uuid(), '0f200db3-a712-4635-9cfa-8df4839f1adb', 'Activity Name C', 'Prepare breakfast.', 'pending', 3),

    (gen_random_uuid(), '9b8c7d6e-5f4a-3210-9876-543210fedcba', 'Medication Administration', 'Administer morning medications as prescribed.', 'pending', 1),
    (gen_random_uuid(), '9b8c7d6e-5f4a-3210-9876-543210fedcba', 'Personal Care Assistance', 'Assist with personal hygiene and grooming.', 'pending', 2),
    (gen_random_uuid(), '9b8c7d6e-5f4a-3210-9876-543210fedcba', 'Meal Preparation', 'Prepare and serve lunch.', 'pending', 3),
    (gen_random_uuid(), '9b8c7d6e-5f4a-3210-9876-543210fedcba', 'Physical Therapy Exercises', 'Guide through prescribed exercises.', 'pending', 4),

    (gen_random_uuid(), 'a8a6d494-3b35-4536-b4f4-8023c2f13914', 'Activity Name A', 'Review medication checklist.', 'completed', 1),
    (gen_random_uuid(), 'a8a6d494-3b35-4536-b4f4-8023c2f13914', 'Activity Name B', 'Assist with morning routine.', 'completed', 2),
    (gen_random_uuid(), 'a8a6d494-3b35-4536-b4f4-8023c2f13914', 'Activity Name C', 'Prepare breakfast.', 'not_completed', 3),

    (gen_random_uuid(), 'b704cb79-bcf4-4b52-a0b4-d5a2a0e12d62', 'Activity Name A', 'Review medication checklist.', 'pending', 1);

UPDATE schedule_tasks
SET not_completed_reason = 'Client declined assistance.'
WHERE status = 'not_completed';

INSERT INTO auth_clients (id, secret_hash, description, caregiver_id, scopes)
VALUES
    ('caregiver-app', 'bcrypt$JDJhJDEwJDFXN3hpcWNONVlkbTlPTkMweFg1OE8vSW9nOWViT2hkU0hlclpNZ2htQ3ZFMzFOWGJFR3Uy', 'Default caregiver mobile app client', 'c2d1bb61-8d67-4db5-9e59-4c2c16f7d4f2', ARRAY['schedules.read','schedules.write','tasks.write']);

-- +migrate Down
DROP TABLE IF EXISTS auth_clients;
DROP TABLE IF EXISTS schedule_tasks;
DROP TABLE IF EXISTS schedules;
DROP TABLE IF EXISTS clients;
DROP TABLE IF EXISTS caregivers;
