CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slack_user_id TEXT NOT NULL UNIQUE,
    slack_team_id TEXT NOT NULL,
    email TEXT,
    display_name TEXT,
    neurotype TEXT CHECK (neurotype IN ('adhd', 'autism', 'anxiety', 'unspecified', 'ally')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE user_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    focus_mode_enabled BOOLEAN DEFAULT true,
    focus_threshold INTEGER DEFAULT 50,
    translator_enabled BOOLEAN DEFAULT true,
    digest_enabled BOOLEAN DEFAULT false,
    digest_hour INTEGER CHECK (digest_hour BETWEEN 0 AND 23) DEFAULT 16,
    deep_work_auto_detect BOOLEAN DEFAULT false,
    quiet_hours_start TIME DEFAULT '22:00',
    quiet_hours_end TIME DEFAULT '08:00',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slack_channel_id TEXT NOT NULL UNIQUE,
    slack_team_id TEXT NOT NULL,
    name TEXT,
    is_dm BOOLEAN DEFAULT false,
    focus_mode_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE channel_subscriptions (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    channel_id UUID REFERENCES channels(id) ON DELETE CASCADE,
    muted BOOLEAN DEFAULT false,
    PRIMARY KEY (user_id, channel_id)
);

CREATE TABLE digests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    sent_at TIMESTAMPTZ DEFAULT NOW(),
    mention_count INTEGER DEFAULT 0,
    thread_reply_count INTEGER DEFAULT 0,
    content JSONB,
    status TEXT CHECK (status IN ('pending', 'sent', 'read')) DEFAULT 'pending'
);

CREATE TABLE focus_summaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID REFERENCES channels(id) ON DELETE CASCADE,
    triggered_at TIMESTAMPTZ DEFAULT NOW(),
    message_count INTEGER,
    summary_text TEXT,
    ai_model TEXT DEFAULT 'gpt-4o-mini',
    raw_messages JSONB
);

CREATE TABLE translations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    original_message_ts TEXT,
    original_channel_id TEXT,
    original_text TEXT,
    translation_text TEXT,
    tone TEXT,
    intent TEXT,
    action TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_slack_id ON users(slack_user_id);
CREATE INDEX idx_channels_slack_id ON channels(slack_channel_id);
CREATE INDEX idx_digests_user_sent ON digests(user_id, sent_at);
CREATE INDEX idx_translations_user ON translations(user_id, created_at);
