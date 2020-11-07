package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const dataStore string = "./data"

//Answer - type fore recording user data
type Answer struct {
	ID       int
	FullName string
	Choice   string
	DateTime time.Time
}

func initDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		panic(err)
	}
	if db == nil {
		panic("db nil")
	}
	return db
}

func storeAnswer(db *sql.DB, answer Answer) error {
	sqlAddAnswer := `INSERT INTO answer( ad_id, full_name, choice, datetime) values(?, ?, ?, datetime('now'))`

	stmt, err := db.Prepare(sqlAddAnswer)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(answer.ID, answer.FullName, answer.Choice)
	if err2 != nil {
		return err2
	}

	return nil
}

func saveAnswerHTML(a Answer, u user) error {
	filename := fmt.Sprintf("%s/%d.html", dataStore, a.ID)
	//fmt.Println("about to store data into", filename)
	t := template.Must(template.ParseFiles("templates/answer.gohtml"))

	respData := struct {
		Answer                    Answer
		User                      user
		ChoiceX, ChoiceY, ChoiceN bool
	}{
		Answer:  a,
		User:    u,
		ChoiceX: a.Choice == "x",
		ChoiceY: a.Choice == "y",
		ChoiceN: a.Choice == "n",
	}

	var f *os.File
	var err error
	if f, err = os.Create(filename); err != nil {
		fmt.Println("Create file:", err)
		return err
	}
	defer f.Close()

	err = t.Execute(f, respData)
	if err != nil {
		fmt.Println("Execute:", err)
		return err
	}
	return nil

}

func retrieveAnswers(db *sql.DB, date string) ([]Answer, error) {
	sqlRetrieveAnswers := `
		SELECT ad_id, full_name, choice, datetime(answer.datetime, 'localtime')
		FROM answer
		ORDER BY datetime DESC
	`

	var result []Answer
	var rows *sql.Rows
	var err error
	if date != "" {
		sqlRetrieveAnswers = `			
			WITH latest AS (
				SELECT ad_id, full_name, max(datetime) as maxdt
				FROM answer
				WHERE date(datetime, 'localtime') = ?
				GROUP by ad_id, full_name
			)
			SELECT answer.ad_id, answer.full_name, answer.choice, datetime(answer.datetime, 'localtime')
			FROM latest
			JOIN answer ON latest.ad_id = answer.ad_id AND latest.maxdt = answer.datetime
			ORDER by answer.datetime DESC
		`
		//fmt.Println("***:", sqlRetrieveAnswers, date)
		rows, err = db.Query(sqlRetrieveAnswers, date)
	} else {
		rows, err = db.Query(sqlRetrieveAnswers)
	}

	if err != nil {
		return result, err
	}
	defer rows.Close()

	var dt string
	for rows.Next() {
		var a Answer
		err := rows.Scan(&a.ID, &a.FullName, &a.Choice, &dt)
		if err != nil {
			return result, err
		}
		a.DateTime, err = time.Parse("2006-01-02 15:04:05", dt)
		if err != nil {
			fmt.Println("Error parsing datetime from db: ", err)
		}
		result = append(result, a)
	}
	return result, nil
}
