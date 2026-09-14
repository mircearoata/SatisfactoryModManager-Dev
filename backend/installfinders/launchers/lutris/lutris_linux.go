package lutris

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"

	"github.com/satisfactorymodding/SatisfactoryModManager/backend/installfinders/common"
	"github.com/satisfactorymodding/SatisfactoryModManager/backend/installfinders/launchers"
	"github.com/satisfactorymodding/SatisfactoryModManager/backend/installfinders/launchers/epic"
	"github.com/satisfactorymodding/SatisfactoryModManager/backend/utils"

	_ "modernc.org/sqlite"
)

type Game struct {
	ID        int    `json:"id"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Runner    string `json:"runner"`
	Directory string `json:"directory"`
}

func init() {
	launchers.Add("Lutris", func() ([]*common.Installation, []error) {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return nil, []error{fmt.Errorf("failed to get user config dir: %w", err)}
		}

		dataDir, err := utils.UserDataDir()
		if err != nil {
			return nil, []error{fmt.Errorf("failed to get user data dir: %w", err)}
		}

		return findInstallations(configDir, dataDir, []string{"lutris"}, "Lutris")
	})
	launchers.Add("Lutris-flatpak", func() ([]*common.Installation, []error) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, []error{fmt.Errorf("failed to get user home dir: %w", err)}
		}

		flatpakXdgConfigHome := filepath.Join(homeDir, ".var", "app", "net.lutris.Lutris", "config")
		flatpakXdgDataHome := filepath.Join(homeDir, ".var", "app", "net.lutris.Lutris", "data")

		return findInstallations(flatpakXdgConfigHome, flatpakXdgDataHome, []string{"flatpak", "run", "net.lutris.Lutris"}, "Lutris")
	})
}

func findInstallations(xdgConfigDir string, xdgDataDir string, lutrisCmd []string, launcher string) ([]*common.Installation, []error) {
	lutrisGames, err := getGames(xdgConfigDir, xdgDataDir)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to get lutris games: %w", err)}
	}

	installs := []*common.Installation{}
	findErrors := []error{}
	for _, lutrisGame := range lutrisGames {
		if lutrisGame.Runner != "wine" || lutrisGame.Directory == "" {
			continue
		}
		currentInstalls, errs := epic.FindInstallationsWine(lutrisGame.Directory, launcher+" - "+lutrisGame.Name, append(lutrisCmd, "lutris:rungame/"+lutrisGame.Slug))
		installs = append(installs, currentInstalls...)
		if errs != nil {
			findErrors = append(findErrors, errs...)
		}
	}
	return installs, findErrors
}

func getGames(xdgConfigDir string, xdgDataDir string) ([]Game, error) {
	dbPath, err := getLutrisDBPath(xdgConfigDir, xdgDataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get lutris database path: %w", err)
	}
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("lutris not installed")
	}

	return getGamesFromDB(dbPath)
}

func getLutrisDBPath(xdgConfigDir string, xdgDataDir string) (string, error) {
	configDir := filepath.Join(xdgConfigDir, "lutris")
	dataDir := filepath.Join(xdgDataDir, "lutris")

	dbPath := filepath.Join(dataDir, "pga.db")

	_, err := os.Stat(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			configDir = dataDir
		} else {
			return "", fmt.Errorf("failed to stat lutris config dir: %w", err)
		}
	}

	lutrisConfigPath := filepath.Join(configDir, "lutris.conf")
	if _, err := os.Stat(lutrisConfigPath); err != nil {
		if os.IsNotExist(err) {
			return dbPath, nil
		}
		return "", fmt.Errorf("failed to stat lutris config: %w", err)
	}

	lutrisConfig, err := ini.Load(lutrisConfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to load lutris config: %w", err)
	}

	lutrisSection := lutrisConfig.Section("lutris")
	configDbPath := lutrisSection.Key("pga_path").String()

	if configDbPath != "" {
		return configDbPath, nil
	}

	return dbPath, nil
}

func getGamesFromDB(dbPath string) ([]Game, error) {
	connectionString := fmt.Sprintf("file:%s?mode=ro", dbPath)
	db, err := sql.Open("sqlite", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open lutris database: %w", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		PRAGMA busy_timeout = 1000;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to setup connection pragmas: %w", err)
	}

	rows, err := db.Query("SELECT id, slug, name, runner, directory FROM games WHERE installed = 1")
	if err != nil {
		return nil, fmt.Errorf("failed to query lutris database: %w", err)
	}
	defer rows.Close()

	var games []Game

	for rows.Next() {
		var g Game
		var dir sql.NullString
		if err := rows.Scan(&g.ID, &g.Slug, &g.Name, &g.Runner, &dir); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		g.Directory = dir.String
		games = append(games, g)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read rows from database: %w", err)
	}

	return games, nil
}
