CREATE TABLE IF NOT EXISTS trainer_assignments (
  user_id UUID NOT NULL,
  trainer_id UUID NOT NULL,
  assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, trainer_id)
);

CREATE TABLE IF NOT EXISTS user_goals (
  user_id UUID NOT NULL,
  goal_code TEXT NOT NULL,
  PRIMARY KEY (user_id, goal_code)
);

CREATE TABLE IF NOT EXISTS user_injuries (
  user_id UUID NOT NULL,
  injury_code TEXT NOT NULL,
  PRIMARY KEY (user_id, injury_code)
);

CREATE TABLE IF NOT EXISTS dietary_preferences (
  user_id UUID NOT NULL,
  preference_code TEXT NOT NULL,
  PRIMARY KEY (user_id, preference_code)
);

CREATE TABLE IF NOT EXISTS user_equipment (
  user_id UUID NOT NULL,
  equipment_id UUID NOT NULL,
  PRIMARY KEY (user_id, equipment_id)
);

CREATE TABLE IF NOT EXISTS availability_rules (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  weekday TEXT NOT NULL,
  duration_min INT NOT NULL
);

CREATE TABLE IF NOT EXISTS exercise_media (
  id UUID PRIMARY KEY,
  exercise_id UUID NOT NULL,
  media_url TEXT NOT NULL,
  media_type TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS workout_day_templates (
  id UUID PRIMARY KEY,
  plan_version_id UUID NOT NULL,
  week_index INT NOT NULL,
  weekday TEXT NOT NULL,
  session_name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS workout_day_exercises (
  id UUID PRIMARY KEY,
  template_id UUID NOT NULL,
  exercise_id UUID NOT NULL,
  exercise_order INT NOT NULL,
  sets_count INT NOT NULL,
  reps_text TEXT NOT NULL,
  rest_sec INT NOT NULL
);

CREATE TABLE IF NOT EXISTS hydration_targets (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  target_ml INT NOT NULL,
  override_ml INT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS weekly_checkins (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  weight_kg NUMERIC(6,2),
  energy_level INT,
  note TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS notification_preferences (
  user_id UUID PRIMARY KEY,
  workout_reminder BOOLEAN NOT NULL DEFAULT TRUE,
  weekly_checkin_reminder BOOLEAN NOT NULL DEFAULT TRUE,
  hydration_reminder BOOLEAN NOT NULL DEFAULT TRUE,
  email_enabled BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS ai_generation_logs (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  generation_type TEXT NOT NULL,
  request_payload JSONB NOT NULL,
  response_payload JSONB,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS support_threads (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  status TEXT NOT NULL,
  assigned_to UUID,
  title TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS support_messages (
  id UUID PRIMARY KEY,
  thread_id UUID NOT NULL,
  author_user_id UUID NOT NULL,
  body TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS public_discussion_threads (
  id UUID PRIMARY KEY,
  author_user_id UUID NOT NULL,
  category TEXT NOT NULL,
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  moderation_status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS public_discussion_replies (
  id UUID PRIMARY KEY,
  thread_id UUID NOT NULL,
  author_user_id UUID NOT NULL,
  body TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS trainer_notes (
  id UUID PRIMARY KEY,
  trainer_id UUID NOT NULL,
  user_id UUID NOT NULL,
  body TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS admin_audit_logs (
  id UUID PRIMARY KEY,
  admin_user_id UUID NOT NULL,
  action_type TEXT NOT NULL,
  resource_type TEXT NOT NULL,
  resource_id TEXT NOT NULL,
  details JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
