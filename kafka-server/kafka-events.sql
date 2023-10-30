PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE events (
  Min INT NOT NULL,
  Player TEXT NOT NULL,
  Club TEXT NOT NULL,
  Score TEXT NOT NULL
);

INSERT INTO events (Min, Player, Club, Score) VALUES (28, 'Leroy Sane', 'FCB', '0-1');
INSERT INTO events (Min, Player, Club, Score) VALUES (49, 'Serge Gnabry', 'FCB', '0-2');
INSERT INTO events (Min, Player, Club, Score) VALUES (53, 'Rasmus Hojlund', 'MNU', '1-2');
INSERT INTO events (Min, Player, Club, Score) VALUES (88, 'Harry Kane', 'FCB', '1-3');
INSERT INTO events (Min, Player, Club, Score) VALUES (92, 'Casemiro', 'MNU', '2-4');
INSERT INTO events (Min, Player, Club, Score) VALUES (95, 'Mathys Tel', 'FCB', '2-4');
INSERT INTO events (Min, Player, Club, Score) VALUES (95, 'Casemiro', 'MNU', '3-4');

COMMIT;
