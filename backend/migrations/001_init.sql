CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  email_verified_at TIMESTAMPTZ,
  locale TEXT NOT NULL DEFAULT 'ru',
  theme TEXT NOT NULL DEFAULT 'light',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS roles (
  id UUID PRIMARY KEY,
  code TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS user_roles (
  user_id UUID NOT NULL,
  role_id UUID NOT NULL,
  PRIMARY KEY (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS sessions (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  token_hash TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS email_verification_tokens (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  token_hash TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  token_hash TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS user_profiles (
  user_id UUID PRIMARY KEY,
  age INT,
  biological_sex TEXT,
  height_cm INT,
  current_weight_kg NUMERIC(6,2),
  target_weight_kg NUMERIC(6,2),
  primary_goal TEXT,
  program_duration_weeks INT,
  experience_level TEXT,
  activity_level TEXT,
  training_location TEXT,
  timezone TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS equipment_catalog (
  id UUID PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  category TEXT NOT NULL,
  location_type TEXT NOT NULL,
  media_url TEXT,
  active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS equipment_catalog_translations (
  equipment_id UUID NOT NULL,
  locale TEXT NOT NULL,
  name TEXT NOT NULL,
  description TEXT NOT NULL,
  PRIMARY KEY (equipment_id, locale)
);

CREATE TABLE IF NOT EXISTS exercise_catalog (
  id UUID PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  movement_pattern TEXT NOT NULL,
  difficulty TEXT NOT NULL,
  location_type TEXT NOT NULL,
  media_url TEXT,
  active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS exercise_catalog_translations (
  exercise_id UUID NOT NULL,
  locale TEXT NOT NULL,
  name TEXT NOT NULL,
  description TEXT NOT NULL,
  technique_steps TEXT NOT NULL,
  PRIMARY KEY (exercise_id, locale)
);

CREATE TABLE IF NOT EXISTS generated_plan_versions (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  parent_plan_version_id UUID,
  version_no INT NOT NULL,
  status TEXT NOT NULL,
  regeneration_reason TEXT,
  superseded_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS nutrition_plan_versions (
  id UUID PRIMARY KEY,
  plan_version_id UUID NOT NULL,
  daily_calories INT NOT NULL,
  protein_g INT NOT NULL,
  carbs_g INT NOT NULL,
  fat_g INT NOT NULL,
  daily_water_ml INT NOT NULL
);

CREATE TABLE IF NOT EXISTS workout_schedule_instances (
  id UUID PRIMARY KEY,
  plan_version_id UUID NOT NULL,
  scheduled_for TIMESTAMPTZ NOT NULL,
  status TEXT NOT NULL,
  rescheduled_from_id UUID
);

CREATE TABLE IF NOT EXISTS workout_logs (
  id UUID PRIMARY KEY,
  plan_version_id UUID NOT NULL,
  schedule_instance_id UUID NOT NULL,
  status TEXT NOT NULL,
  discomfort_flag BOOLEAN NOT NULL DEFAULT FALSE,
  difficulty SMALLINT,
  note TEXT,
  completed_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS meal_logs (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  status TEXT NOT NULL,
  note TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS water_logs (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  amount_ml INT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS notifications (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  type TEXT NOT NULL,
  title TEXT NOT NULL,
  body TEXT,
  target_url TEXT,
  is_read BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
