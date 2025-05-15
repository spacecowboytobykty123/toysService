ALTER TABLE toys ADD CONSTRAINT toys_details_length_check CHECK (array_length(details, 1) BETWEEN 1 AND 5);
ALTER TABLE toys ADD CONSTRAINT toys_skills_length_check CHECK (array_length(skills, 1) BETWEEN 1 AND 5);
ALTER TABLE toys ADD CONSTRAINT toys_categories_length_check CHECK (array_length(categories, 1) BETWEEN 1 AND 5);
ALTER TABLE toys ADD CONSTRAINT toys_images_length_check CHECK (array_length(images, 1) BETWEEN 1 AND 5);

ALTER TABLE toys ADD CONSTRAINT toys_value_check CHECK (value >= 2000);