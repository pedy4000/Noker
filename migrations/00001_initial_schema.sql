-- migrations/00001_initial_schema.sql
-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE meetings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title TEXT NOT NULL,
    raw_notes TEXT NOT NULL,
    source TEXT NOT NULL CHECK (source IN ('notion', 'file', 'manual')),
    metadata JSONB,
    processing_status TEXT NOT NULL DEFAULT 'pending' CHECK (processing_status IN ('pending', 'processing', 'done', 'failed')),
    processing_error TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    processed_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE themes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE opportunities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_segment TEXT NOT NULL,
    struggle TEXT NOT NULL,
    why_it_matters TEXT,
    workaround TEXT,
    theme_id UUID REFERENCES themes(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE opportunity_evidence (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    opportunity_id UUID NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    meeting_id UUID NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
    quote TEXT NOT NULL,
    context TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_opportunities_theme ON opportunities(theme_id);
CREATE INDEX idx_evidence_opportunity ON opportunity_evidence(opportunity_id);
CREATE INDEX idx_evidence_meeting ON opportunity_evidence(meeting_id);
-- in case of providing api of meetings
-- CREATE INDEX idx_meetings_created ON meetings(created_at DESC);

-- +goose Down
DROP TABLE opportunity_evidence;
DROP TABLE opportunities;
DROP TABLE themes;
DROP TABLE meetings;