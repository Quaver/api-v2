CREATE UNIQUE INDEX `UNIQUE`
    ON scores (id);

CREATE INDEX personal_best
    ON scores (personal_best);

CREATE INDEX mods
    ON scores (mods);
