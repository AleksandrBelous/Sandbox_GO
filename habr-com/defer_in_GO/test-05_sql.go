package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func executeTransaction(db *sql.DB) error {
	tx, err := db.Begin()

	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			fmt.Println("Транзакция откатилась")
		} else {
			tx.Commit()
			fmt.Println("Транзакция завершена")
		}
	}()

	_, err = tx.Exec("INSERT INTO users(name) VALUES('John')")

	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO accounts(user_id) VALUES(LAST_INSERT_ID())")

	return err
}

func main() {
	db, err := sql.Open("mysql", "user:password@/dbname")

	if err != nil {
		fmt.Println("Ошибка подключения к базе данных:", err)
		return
	}

	defer db.Close()

	err = executeTransaction(db)

	if err != nil {
		fmt.Println("Ошибка выполнения транзакции:", err)
	}
}
