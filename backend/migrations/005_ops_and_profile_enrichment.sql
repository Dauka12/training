ALTER TABLE user_profiles
  ADD COLUMN IF NOT EXISTS preferred_training_style TEXT,
  ADD COLUMN IF NOT EXISTS preferred_meal_style TEXT,
  ADD COLUMN IF NOT EXISTS hydration_preference TEXT;

CREATE TABLE IF NOT EXISTS email_logs (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type TEXT NOT NULL,
  recipient_email TEXT NOT NULL,
  subject TEXT NOT NULL,
  status TEXT NOT NULL,
  error TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
