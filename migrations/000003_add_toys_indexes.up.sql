CREATE INDEX IF NOT EXISTS toys_title_idx ON toys USING GIN (to_tsvector('simple', title));
CREATE INDEX IF NOT EXISTS toys_categories_idx ON toys USING GIN (categories);
CREATE INDEX IF NOT EXISTS toys_skills_genres_idx ON toys USING GIN (skills);


