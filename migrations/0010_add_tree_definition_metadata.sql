ALTER TABLE tgo_trees
ADD COLUMN seed_codes_json TEXT NOT NULL DEFAULT '[]';

ALTER TABLE tgo_trees
ADD COLUMN priority_skills_json TEXT NOT NULL DEFAULT '[]';

ALTER TABLE tree_tgos
ADD COLUMN prerequisites_json TEXT NOT NULL DEFAULT '[]';

ALTER TABLE tree_tgos
ADD COLUMN mastery_hint TEXT NOT NULL DEFAULT '';
