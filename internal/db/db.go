package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nv-copa/internal/copa"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

type User struct {
	ID           int64
	Name         string
	IsAdmin      bool
	GroupsLocked bool
}

type Match struct {
	ID        int64
	HomeTeam  string
	AwayTeam  string
	StartsAt  time.Time
	HomeScore sql.NullInt64
	AwayScore sql.NullInt64
}

type GroupResult struct {
	GroupName string
	TeamName  string
	Position  int
}

type UserFixtureHit struct {
	Round     int
	HomeTeam  string
	AwayTeam  string
	PredHome  int64
	PredAway  int64
	RealHome  int64
	RealAway  int64
	Points    int
	HitType   string
	HasResult bool
}

type UserFixtureRound struct {
	Round int
	Label string
	Hits  []UserFixtureHit
}

type UserGroupHit struct {
	GroupName string
	Points    int
	HitType   string
}

type UserRanking struct {
	Position      int
	UserName      string
	TotalPoints   int
	FixtureHits   []UserFixtureHit
	FixtureRounds []UserFixtureRound
	GroupHits     []UserGroupHit
	Podium        PodiumPrediction
}

type PodiumPrediction struct {
	UserID   int64
	Champion string
	RunnerUp string
	Third    string
}

type GroupPrediction struct {
	UserID    int64
	GroupName string
	TeamName  string
	Position  int
}

type Fixture struct {
	ID        int64
	Round     int
	GroupName string
	MatchDate time.Time
	HomeTeam  string
	AwayTeam  string
	HomeScore sql.NullInt64
	AwayScore sql.NullInt64
}

type FixturePrediction struct {
	Fixture
	PredHomeScore sql.NullInt64
	PredAwayScore sql.NullInt64
	PredCreatedAt sql.NullString
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1)

	if _, err := conn.Exec("PRAGMA foreign_keys = ON; PRAGMA journal_mode = WAL; PRAGMA busy_timeout = 5000;"); err != nil {
		_ = conn.Close()
		return nil, err
	}

	return &Store{db: conn}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
	token TEXT PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	expires_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS matches (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	home_team TEXT NOT NULL,
	away_team TEXT NOT NULL,
	starts_at TEXT NOT NULL,
	home_score INTEGER,
	away_score INTEGER,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS podium_predictions (
	user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
	champion TEXT NOT NULL,
	runner_up TEXT NOT NULL,
	third TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS group_predictions (
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	group_name TEXT NOT NULL,
	team_name TEXT NOT NULL,
	position INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (user_id, group_name, team_name),
	UNIQUE (user_id, group_name, position)
);

CREATE TABLE IF NOT EXISTS fixtures (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	stage TEXT NOT NULL,
	round_number INTEGER NOT NULL,
	group_name TEXT NOT NULL,
	match_date TEXT NOT NULL,
	home_team TEXT NOT NULL,
	away_team TEXT NOT NULL,
	home_score INTEGER,
	away_score INTEGER,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(stage, round_number, group_name, match_date, home_team, away_team)
);

CREATE TABLE IF NOT EXISTS fixture_predictions (
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	fixture_id INTEGER NOT NULL REFERENCES fixtures(id) ON DELETE CASCADE,
	home_score INTEGER NOT NULL,
	away_score INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (user_id, fixture_id)
);

CREATE TABLE IF NOT EXISTS group_results (
	group_name TEXT NOT NULL,
	team_name TEXT NOT NULL,
	position INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (group_name, team_name)
);
`)
	if err != nil {
		return err
	}
	if err := s.ensureUserColumns(); err != nil {
		return err
	}
	return nil
}

func (s *Store) CreateUser(name, passwordHash string) (User, error) {
	if name == "" || passwordHash == "" {
		return User{}, errors.New("nome e senha sao obrigatorios")
	}

	isAdmin := 0
	if strings.EqualFold(name, "ADMINISTRADOR") {
		isAdmin = 1
	}

	result, err := s.db.Exec("INSERT INTO users (name, password_hash, is_admin) VALUES (?, ?, ?)", name, passwordHash, isAdmin)
	if err != nil {
		return User{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return User{}, err
	}
	return User{ID: id, Name: name, IsAdmin: isAdmin == 1}, nil
}

func (s *Store) UserPasswordHash(name string) (User, string, error) {
	var user User
	var passwordHash string
	var isAdmin int
	var groupsLocked int
	err := s.db.QueryRow("SELECT id, name, password_hash, is_admin, CASE WHEN groups_locked_at IS NOT NULL THEN 1 ELSE 0 END FROM users WHERE name = ?", name).Scan(&user.ID, &user.Name, &passwordHash, &isAdmin, &groupsLocked)
	user.IsAdmin = isAdmin == 1
	user.GroupsLocked = groupsLocked == 1
	return user, passwordHash, err
}

func (s *Store) CreateSession(userID int64, token string, expiresAt time.Time) error {
	_, err := s.db.Exec(
		"INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)",
		token,
		userID,
		expiresAt.UTC().Format(time.RFC3339),
	)
	return err
}

func (s *Store) UserBySessionToken(token string) (User, error) {
	var user User
	var expiresAt string
	var isAdmin int
	var groupsLocked int
	err := s.db.QueryRow(`
SELECT u.id, u.name, u.is_admin, CASE WHEN u.groups_locked_at IS NOT NULL THEN 1 ELSE 0 END, s.expires_at
FROM sessions s
JOIN users u ON u.id = s.user_id
WHERE s.token = ?
`, token).Scan(&user.ID, &user.Name, &isAdmin, &groupsLocked, &expiresAt)
	if err != nil {
		return User{}, err
	}
	user.IsAdmin = isAdmin == 1
	user.GroupsLocked = groupsLocked == 1

	parsed, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return User{}, err
	}
	if time.Now().UTC().After(parsed) {
		_ = s.DeleteSession(token)
		return User{}, sql.ErrNoRows
	}

	return user, nil
}

func (s *Store) DeleteSession(token string) error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}

func (s *Store) PodiumPrediction(userID int64) (PodiumPrediction, error) {
	var prediction PodiumPrediction
	err := s.db.QueryRow(
		"SELECT user_id, champion, runner_up, third FROM podium_predictions WHERE user_id = ?",
		userID,
	).Scan(&prediction.UserID, &prediction.Champion, &prediction.RunnerUp, &prediction.Third)
	return prediction, err
}

func (s *Store) HasPodiumPrediction(userID int64) (bool, error) {
	_, err := s.PodiumPrediction(userID)
	if err == nil {
		return true, nil
	}
	if err == sql.ErrNoRows {
		return false, nil
	}
	return false, err
}

func (s *Store) SavePodiumPrediction(userID int64, champion, runnerUp, third string) error {
	if champion == "" || runnerUp == "" || third == "" {
		return errors.New("campeao, vice e terceiro sao obrigatorios")
	}
	if champion == runnerUp || champion == third || runnerUp == third {
		return errors.New("escolha tres selecoes diferentes para o podio")
	}

	_, err := s.db.Exec(`
INSERT INTO podium_predictions (user_id, champion, runner_up, third)
VALUES (?, ?, ?, ?)
ON CONFLICT(user_id) DO UPDATE SET
	champion = excluded.champion,
	runner_up = excluded.runner_up,
	third = excluded.third,
	updated_at = CURRENT_TIMESTAMP
`, userID, champion, runnerUp, third)
	return err
}

func (s *Store) GroupPredictions(userID int64) ([]GroupPrediction, error) {
	rows, err := s.db.Query(`
SELECT user_id, group_name, team_name, position
FROM group_predictions
WHERE user_id = ?
ORDER BY group_name, position
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var predictions []GroupPrediction
	for rows.Next() {
		var prediction GroupPrediction
		if err := rows.Scan(&prediction.UserID, &prediction.GroupName, &prediction.TeamName, &prediction.Position); err != nil {
			return nil, err
		}
		predictions = append(predictions, prediction)
	}
	return predictions, rows.Err()
}

func (s *Store) SaveGroupPredictions(userID int64, predictions []GroupPrediction) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM group_predictions WHERE user_id = ?", userID); err != nil {
		return err
	}
	for _, prediction := range predictions {
		if _, err := tx.Exec(`
INSERT INTO group_predictions (user_id, group_name, team_name, position)
VALUES (?, ?, ?, ?)
`, userID, prediction.GroupName, prediction.TeamName, prediction.Position); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) HasLockedGroupPredictions(userID int64) (bool, error) {
	var locked int
	if err := s.db.QueryRow("SELECT CASE WHEN groups_locked_at IS NOT NULL THEN 1 ELSE 0 END FROM users WHERE id = ?", userID).Scan(&locked); err != nil {
		return false, err
	}
	return locked == 1, nil
}

func (s *Store) SaveGroupPredictionsForGroup(userID int64, groupName string, predictions []GroupPrediction) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM group_predictions WHERE user_id = ? AND group_name = ?", userID, groupName); err != nil {
		return err
	}
	for _, prediction := range predictions {
		if _, err := tx.Exec(`
INSERT INTO group_predictions (user_id, group_name, team_name, position)
VALUES (?, ?, ?, ?)
`, userID, prediction.GroupName, prediction.TeamName, prediction.Position); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) FinalizeGroupPredictions(userID int64) error {
	predictions, err := s.GroupPredictions(userID)
	if err != nil {
		return err
	}
	if err := validateCompleteGroupPredictionSet(predictions); err != nil {
		return err
	}
	_, err = s.db.Exec("UPDATE users SET groups_locked_at = CURRENT_TIMESTAMP WHERE id = ?", userID)
	return err
}

func (s *Store) SetFixtureResult(fixtureID, homeScore, awayScore int64) error {
	_, err := s.db.Exec("UPDATE fixtures SET home_score = ?, away_score = ? WHERE id = ?", homeScore, awayScore, fixtureID)
	return err
}

func (s *Store) Seed() error {
	if err := s.seedFixtures(); err != nil {
		return err
	}

	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM matches").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now().UTC()
	matches := []struct {
		home string
		away string
		at   time.Time
	}{
		{"Brasil", "Mexico", now.Add(24 * time.Hour)},
		{"Argentina", "Canada", now.Add(48 * time.Hour)},
		{"Franca", "Japao", now.Add(72 * time.Hour)},
		{"Espanha", "Uruguai", now.Add(96 * time.Hour)},
	}

	for _, match := range matches {
		if _, err := s.db.Exec(
			"INSERT INTO matches (home_team, away_team, starts_at) VALUES (?, ?, ?)",
			match.home,
			match.away,
			match.at.Format(time.RFC3339),
		); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) seedFixtures() error {
	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM fixtures WHERE stage = 'groups'").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	for _, fixture := range copa.GroupStageFixtures {
		if _, err := s.db.Exec(`
INSERT INTO fixtures (stage, round_number, group_name, match_date, home_team, away_team)
VALUES ('groups', ?, ?, ?, ?, ?)
`, fixture.Round, fixture.GroupName, fixture.Date, fixture.HomeTeam, fixture.AwayTeam); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Matches() ([]Match, error) {
	rows, err := s.db.Query("SELECT id, home_team, away_team, starts_at, home_score, away_score FROM matches ORDER BY starts_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []Match
	for rows.Next() {
		var match Match
		var startsAt string
		if err := rows.Scan(&match.ID, &match.HomeTeam, &match.AwayTeam, &startsAt, &match.HomeScore, &match.AwayScore); err != nil {
			return nil, err
		}
		parsed, err := time.Parse(time.RFC3339, startsAt)
		if err != nil {
			return nil, fmt.Errorf("parse match time: %w", err)
		}
		match.StartsAt = parsed
		matches = append(matches, match)
	}
	return matches, rows.Err()
}

func (s *Store) CreateMatch(home, away, startsAt string) error {
	if home == "" || away == "" || startsAt == "" {
		return errors.New("times e horario sao obrigatorios")
	}
	parsed, err := time.Parse("2006-01-02T15:04", startsAt)
	if err != nil {
		return errors.New("horario invalido")
	}
	_, err = s.db.Exec(
		"INSERT INTO matches (home_team, away_team, starts_at) VALUES (?, ?, ?)",
		home,
		away,
		parsed.UTC().Format(time.RFC3339),
	)
	return err
}

func (s *Store) SetResult(matchID, homeScore, awayScore int64) error {
	_, err := s.db.Exec("UPDATE matches SET home_score = ?, away_score = ? WHERE id = ?", homeScore, awayScore, matchID)
	return err
}

func (s *Store) CurrentGroupRound() (int, error) {
	rows, err := s.db.Query(`
SELECT round_number,
	SUM(CASE WHEN home_score IS NOT NULL AND away_score IS NOT NULL THEN 1 ELSE 0 END) AS completed,
	COUNT(*) AS total
FROM fixtures
WHERE stage = 'groups'
GROUP BY round_number
ORDER BY round_number
`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	current := 1
	for rows.Next() {
		var round, completed, total int
		if err := rows.Scan(&round, &completed, &total); err != nil {
			return 0, err
		}
		current = round
		if completed < total {
			return round, nil
		}
	}
	return current, rows.Err()
}

func (s *Store) GroupFixtures(round int, groupName string) ([]Fixture, error) {
	rows, err := s.db.Query(`
SELECT id, round_number, group_name, match_date, home_team, away_team, home_score, away_score
FROM fixtures
WHERE stage = 'groups' AND round_number = ? AND group_name = ?
ORDER BY match_date, id
`, round, groupName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fixtures []Fixture
	for rows.Next() {
		fixture, err := scanFixture(rows)
		if err != nil {
			return nil, err
		}
		fixtures = append(fixtures, fixture)
	}
	return fixtures, rows.Err()
}

func (s *Store) AllGroupFixtures(groupName string) ([]Fixture, error) {
	rows, err := s.db.Query(`
SELECT id, round_number, group_name, match_date, home_team, away_team, home_score, away_score
FROM fixtures
WHERE stage = 'groups' AND group_name = ?
ORDER BY round_number, match_date, id
`, groupName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fixtures []Fixture
	for rows.Next() {
		fixture, err := scanFixture(rows)
		if err != nil {
			return nil, err
		}
		fixtures = append(fixtures, fixture)
	}
	return fixtures, rows.Err()
}

func (s *Store) GroupFixturePredictions(userID int64, round int, groupName string) ([]FixturePrediction, error) {
	rows, err := s.db.Query(`
SELECT
	f.id, f.round_number, f.group_name, f.match_date, f.home_team, f.away_team, f.home_score, f.away_score,
	p.home_score, p.away_score, p.created_at
FROM fixtures f
LEFT JOIN fixture_predictions p ON p.fixture_id = f.id AND p.user_id = ?
WHERE f.stage = 'groups' AND f.round_number = ? AND f.group_name = ?
ORDER BY f.match_date, f.id
`, userID, round, groupName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fixtures []FixturePrediction
	for rows.Next() {
		var fixture FixturePrediction
		var matchDate string
		if err := rows.Scan(
			&fixture.ID,
			&fixture.Round,
			&fixture.GroupName,
			&matchDate,
			&fixture.HomeTeam,
			&fixture.AwayTeam,
			&fixture.HomeScore,
			&fixture.AwayScore,
			&fixture.PredHomeScore,
			&fixture.PredAwayScore,
			&fixture.PredCreatedAt,
		); err != nil {
			return nil, err
		}
		parsed, err := time.Parse("2006-01-02", matchDate)
		if err != nil {
			return nil, err
		}
		fixture.MatchDate = parsed
		fixtures = append(fixtures, fixture)
	}
	return fixtures, rows.Err()
}

func (s *Store) SaveFixturePredictions(userID int64, predictions map[int64][2]int64, now time.Time) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for fixtureID, score := range predictions {
		fixture, err := s.fixtureByIDTx(tx, fixtureID)
		if err != nil {
			return err
		}
		if fixture.HomeScore.Valid && fixture.AwayScore.Valid {
			return errors.New("o resultado desse jogo ja foi registrado, palpite nao pode ser alterado")
		}
		var createdAt sql.NullString
		if err := tx.QueryRow("SELECT created_at FROM fixture_predictions WHERE user_id = ? AND fixture_id = ?", userID, fixtureID).Scan(&createdAt); err != nil && err != sql.ErrNoRows {
			return err
		}
		if createdAt.Valid {
			created, err := time.Parse("2006-01-02 15:04:05", createdAt.String)
			if err == nil && now.Sub(created) > 12*time.Hour {
				return errors.New("o palpite so pode ser alterado em ate 12 horas apos ser criado")
			}
		}
		if _, err := tx.Exec(`
INSERT INTO fixture_predictions (user_id, fixture_id, home_score, away_score)
VALUES (?, ?, ?, ?)
ON CONFLICT(user_id, fixture_id) DO UPDATE SET
	home_score = excluded.home_score,
	away_score = excluded.away_score,
	updated_at = CURRENT_TIMESTAMP
`, userID, fixtureID, score[0], score[1]); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) RoundPredictionProgress(userID int64, round int) (int, int, error) {
	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM fixtures WHERE stage = 'groups' AND round_number = ?", round).Scan(&total); err != nil {
		return 0, 0, err
	}
	var completed int
	if err := s.db.QueryRow(`
SELECT COUNT(*)
FROM fixture_predictions fp
JOIN fixtures f ON f.id = fp.fixture_id
WHERE fp.user_id = ? AND f.stage = 'groups' AND f.round_number = ?
`, userID, round).Scan(&completed); err != nil {
		return 0, 0, err
	}
	return completed, total, nil
}

func scanFixture(scanner interface {
	Scan(dest ...any) error
}) (Fixture, error) {
	var fixture Fixture
	var matchDate string
	if err := scanner.Scan(
		&fixture.ID,
		&fixture.Round,
		&fixture.GroupName,
		&matchDate,
		&fixture.HomeTeam,
		&fixture.AwayTeam,
		&fixture.HomeScore,
		&fixture.AwayScore,
	); err != nil {
		return Fixture{}, err
	}
	parsed, err := time.Parse("2006-01-02", matchDate)
	if err != nil {
		return Fixture{}, err
	}
	fixture.MatchDate = parsed
	return fixture, nil
}

func (s *Store) fixtureByIDTx(tx *sql.Tx, fixtureID int64) (Fixture, error) {
	row := tx.QueryRow(`
SELECT id, round_number, group_name, match_date, home_team, away_team, home_score, away_score
FROM fixtures
WHERE id = ?
`, fixtureID)
	return scanFixture(row)
}

func (s *Store) ensureUserColumns() error {
	alterStatements := []string{
		"ALTER TABLE users ADD COLUMN is_admin INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE users ADD COLUMN groups_locked_at TEXT",
	}
	for _, statement := range alterStatements {
		if _, err := s.db.Exec(statement); err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			return err
		}
	}
	return nil
}

func (s *Store) SaveGroupResults(groupName string, results []GroupResult) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM group_results WHERE group_name = ?", groupName); err != nil {
		return err
	}
	for _, r := range results {
		if _, err := tx.Exec(
			"INSERT INTO group_results (group_name, team_name, position) VALUES (?, ?, ?)",
			r.GroupName, r.TeamName, r.Position,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) GroupResults() ([]GroupResult, error) {
	rows, err := s.db.Query("SELECT group_name, team_name, position FROM group_results ORDER BY group_name, position")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []GroupResult
	for rows.Next() {
		var r GroupResult
		if err := rows.Scan(&r.GroupName, &r.TeamName, &r.Position); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (s *Store) AllUsers() ([]User, error) {
	rows, err := s.db.Query("SELECT id, name, is_admin, CASE WHEN groups_locked_at IS NOT NULL THEN 1 ELSE 0 END FROM users ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		var isAdmin, groupsLocked int
		if err := rows.Scan(&u.ID, &u.Name, &isAdmin, &groupsLocked); err != nil {
			return nil, err
		}
		u.IsAdmin = isAdmin == 1
		u.GroupsLocked = groupsLocked == 1
		users = append(users, u)
	}
	return users, rows.Err()
}

func (s *Store) AllFixturePredictionsForUser(userID int64) ([]FixturePrediction, error) {
	rows, err := s.db.Query(`
SELECT
	f.id, f.round_number, f.group_name, f.match_date, f.home_team, f.away_team, f.home_score, f.away_score,
	p.home_score, p.away_score
FROM fixtures f
LEFT JOIN fixture_predictions p ON p.fixture_id = f.id AND p.user_id = ?
WHERE f.stage = 'groups'
ORDER BY f.match_date, f.id
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fixtures []FixturePrediction
	for rows.Next() {
		var fp FixturePrediction
		var matchDate string
		if err := rows.Scan(
			&fp.ID, &fp.Round, &fp.GroupName, &matchDate,
			&fp.HomeTeam, &fp.AwayTeam, &fp.HomeScore, &fp.AwayScore,
			&fp.PredHomeScore, &fp.PredAwayScore,
		); err != nil {
			return nil, err
		}
		parsed, err := time.Parse("2006-01-02", matchDate)
		if err != nil {
			return nil, err
		}
		fp.MatchDate = parsed
		fixtures = append(fixtures, fp)
	}
	return fixtures, rows.Err()
}

func (s *Store) FullRanking() ([]UserRanking, error) {
	users, err := s.AllUsers()
	if err != nil {
		return nil, err
	}
	groupResults, err := s.GroupResults()
	if err != nil {
		return nil, err
	}

	actualByGroup := make(map[string]map[int]string)
	for _, r := range groupResults {
		if actualByGroup[r.GroupName] == nil {
			actualByGroup[r.GroupName] = make(map[int]string)
		}
		actualByGroup[r.GroupName][r.Position] = r.TeamName
	}

	var rankings []UserRanking
	for _, u := range users {
		ranking := UserRanking{UserName: u.Name}

		podium, err := s.PodiumPrediction(u.ID)
		if err == nil {
			ranking.Podium = podium
		}

		fixtures, err := s.AllFixturePredictionsForUser(u.ID)
		if err != nil {
			return nil, err
		}
		for _, f := range fixtures {
			if !f.PredHomeScore.Valid || !f.PredAwayScore.Valid {
				continue
			}
			ph, pa := f.PredHomeScore.Int64, f.PredAwayScore.Int64
			hit := UserFixtureHit{
				Round:    f.Round,
				HomeTeam: f.HomeTeam,
				AwayTeam: f.AwayTeam,
				PredHome: ph,
				PredAway: pa,
			}

			if f.HomeScore.Valid && f.AwayScore.Valid {
				rh, ra := f.HomeScore.Int64, f.AwayScore.Int64
				hit.HasResult = true
				hit.RealHome = rh
				hit.RealAway = ra
				if ph == rh && pa == ra {
					hit.Points = 3
					hit.HitType = "exact"
					ranking.TotalPoints += 3
				} else if outcome(ph, pa) == outcome(rh, ra) {
					hit.Points = 1
					hit.HitType = "outcome"
					ranking.TotalPoints += 1
				} else {
					hit.HitType = "miss"
				}
			} else {
				hit.HitType = "pending"
			}
			ranking.FixtureHits = append(ranking.FixtureHits, hit)
		}
		ranking.FixtureRounds = fixtureRounds(ranking.FixtureHits)

		predictions, err := s.GroupPredictions(u.ID)
		if err != nil {
			return nil, err
		}
		predByGroup := make(map[string]map[int]string)
		for _, p := range predictions {
			if predByGroup[p.GroupName] == nil {
				predByGroup[p.GroupName] = make(map[int]string)
			}
			predByGroup[p.GroupName][p.Position] = p.TeamName
		}

		for groupName, actual := range actualByGroup {
			pred := predByGroup[groupName]
			if pred == nil {
				continue
			}

			actual1, actual2 := actual[1], actual[2]
			pred1, pred2 := pred[1], pred[2]
			actual3, pred3 := actual[3], pred[3]

			if actual1 != "" && actual2 != "" && pred1 != "" && pred2 != "" {
				exactGroup := pred1 == actual1 && pred2 == actual2 && pred3 == actual3
				if exactGroup {
					ranking.GroupHits = append(ranking.GroupHits, UserGroupHit{
						GroupName: groupName, Points: 5, HitType: "exact_group",
					})
					ranking.TotalPoints += 5
				} else {
					actualSet := map[string]bool{actual1: true, actual2: true}
					if actual3 != "" {
						actualSet[actual3] = true
					}
					predSet := map[string]bool{pred1: true, pred2: true}
					if pred3 != "" {
						predSet[pred3] = true
					}
					matchCount := 0
					for team := range predSet {
						if actualSet[team] {
							matchCount++
						}
					}
					if matchCount > 0 {
						ranking.GroupHits = append(ranking.GroupHits, UserGroupHit{
							GroupName: groupName, Points: matchCount, HitType: "qualified_teams",
						})
						ranking.TotalPoints += matchCount
					}
				}
			}
		}

		rankings = append(rankings, ranking)
	}

	sortRankings(rankings)
	for i := range rankings {
		rankings[i].Position = i + 1
	}
	return rankings, nil
}

func outcome(home, away int64) int {
	if home > away {
		return 1
	}
	if home < away {
		return -1
	}
	return 0
}

func fixtureRounds(hits []UserFixtureHit) []UserFixtureRound {
	byRound := make(map[int][]UserFixtureHit)
	for _, hit := range hits {
		byRound[hit.Round] = append(byRound[hit.Round], hit)
	}

	var rounds []UserFixtureRound
	for round := 1; round <= 3; round++ {
		roundHits := byRound[round]
		if len(roundHits) == 0 {
			continue
		}
		rounds = append(rounds, UserFixtureRound{
			Round: round,
			Label: fmt.Sprintf("%da Rodada", round),
			Hits:  roundHits,
		})
	}
	return rounds
}

func sortRankings(rankings []UserRanking) {
	for i := 0; i < len(rankings); i++ {
		for j := i + 1; j < len(rankings); j++ {
			if rankings[j].TotalPoints > rankings[i].TotalPoints {
				rankings[i], rankings[j] = rankings[j], rankings[i]
			} else if rankings[j].TotalPoints == rankings[i].TotalPoints {
				jExact := countExact(rankings[j].FixtureHits)
				iExact := countExact(rankings[i].FixtureHits)
				if jExact > iExact {
					rankings[i], rankings[j] = rankings[j], rankings[i]
				} else if jExact == iExact && rankings[i].UserName > rankings[j].UserName {
					rankings[i], rankings[j] = rankings[j], rankings[i]
				}
			}
		}
	}
}

func countExact(hits []UserFixtureHit) int {
	count := 0
	for _, h := range hits {
		if h.HitType == "exact" {
			count++
		}
	}
	return count
}

func validateCompleteGroupPredictionSet(predictions []GroupPrediction) error {
	positionsByGroup := make(map[string]map[int]bool)
	thirdCount := 0
	for _, prediction := range predictions {
		if positionsByGroup[prediction.GroupName] == nil {
			positionsByGroup[prediction.GroupName] = make(map[int]bool)
		}
		positionsByGroup[prediction.GroupName][prediction.Position] = true
		if prediction.Position == 3 {
			thirdCount++
		}
	}
	for _, group := range copa.Groups {
		positions := positionsByGroup[group.Name]
		if !positions[1] || !positions[2] {
			return errors.New("preencha primeiro e segundo colocados de todos os grupos antes de finalizar")
		}
	}
	if thirdCount != 8 {
		return errors.New("escolha exatamente 8 terceiros classificados antes de finalizar")
	}
	return nil
}
