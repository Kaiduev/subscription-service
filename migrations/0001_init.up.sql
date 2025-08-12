CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS subscriptions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  service_name TEXT NOT NULL,
  price INTEGER NOT NULL CHECK (price >= 0),
  user_id UUID NOT NULL,
  start_date DATE NOT NULL CHECK (EXTRACT(DAY FROM start_date) = 1),
  end_date DATE NULL CHECK (end_date IS NULL OR (EXTRACT(DAY FROM end_date) = 1 AND end_date >= start_date)),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_service ON subscriptions(service_name);
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_service_start ON subscriptions(user_id, service_name, start_date);