CREATE TABLE IF NOT EXISTS toys (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT now(),
    title text NOT NULL,
    description text NOT NULL,
    skills text[] NOT NULL,
    images text[] NOT NULL,
    categories text[] NOT NULL,
    recommended_age text,
    manufacturer text,
    value integer,
    is_available text
);