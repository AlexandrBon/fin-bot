package repo

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
	"tgbot/internal/app"
)

var (
	host     = os.Getenv("HOST")
	port     = os.Getenv("PORT")
	user     = os.Getenv("USER")
	password = os.Getenv("PASSWORD")
	dbname   = os.Getenv("DBNAME")
	sslMode  = os.Getenv("SSLMODE")

	dbInfo = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslMode)
)

type Repo struct {
	db *sql.DB
}

func New() (app.Repository, error) {
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return nil, err
	}
	return &Repo{db: db}, nil
}

func (r *Repo) AddUser(chatID int64) error {
	query := `
	INSERT INTO user_info(chat_id, balance) values ($1, 0)
	`

	if _, err := r.db.Exec(query, chatID); err != nil {
		return err
	}

	return nil
}

func (r *Repo) CreateUserHistoryTable() error {
	query := `
	CREATE TABLE user_history(
	    chat_id INT REFERENCES user_info(chat_id),
	    data    text[]
	);
	`

	if _, err := r.db.Exec(query); err != nil {
		return err
	}

	return nil
}

func (r *Repo) AddToHistory(chatID int64, data string) error {
	query := `
	UPDATE user_history
	SET data = array_append(data, $1)
	WHERE chat_id = $2;
	`

	if _, err := r.db.Exec(query, data, chatID); err != nil {
		return err
	}

	return nil
}

func (r *Repo) CreateUserInfoTable() error {
	query := `
	CREATE TABLE user_info(
	    chat_id SERIAL PRIMARY KEY, 
	    balance INT
	);
	`

	if _, err := r.db.Exec(query); err != nil {
		return err
	}

	return nil
}

func (r *Repo) UpdateBalance(chatID int64, updatedBalance int64) error {
	query := `
	UPDATE user_info SET balance = $1 WHERE chat_id = $2;
	`

	if _, err := r.db.Exec(query, updatedBalance, chatID); err != nil {
		return err
	}

	return nil
}

func (r *Repo) UserExists(chatID int64) (bool, error) {
	query := `
	SELECT chat_id FROM user_info WHERE chat_id = $1;
	`
	rows, err := r.db.Query(query, chatID)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

func (r *Repo) GetBalance(chatID int64) (int64, error) {
	query := `
	SELECT balance FROM user_info WHERE chat_id = $1;
	`

	rows, err := r.db.Query(query, chatID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var balance int64
	if rows.Next() {
		err = rows.Scan(&balance)
		if err != nil {
			log.Println(err)
		}
	} else {
		return 0, errors.New("chatID does not exist")
	}

	return balance, nil
}
