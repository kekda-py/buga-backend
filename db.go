package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gofiber/websocket/v2"
)

func Delete(id string) error {

	wait := time.Tick(time.Minute)

	<-wait

	var err error

	if _, err = db.Query(`DELETE FROM messages WHERE id = $1`, id); err != nil {
		return err
	}

	return nil

}

func parseArray(jsonBuffer []byte) ([]string, error) {
	var ids []string

	if err := json.Unmarshal(jsonBuffer, &ids); err != nil {
		return nil, err
	}

	return ids, nil
}

func getUUID() ([]string, error) {

	var (
		resp *http.Response
		err  error
	)

	if resp, err = http.Get("https://www.uuidtools.com/api/generate/v1"); err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var result []byte

	if result, err = io.ReadAll(resp.Body); err != nil {
		return nil, err
	}

	var ids []string

	if ids, err = parseArray(result); err != nil {
		return nil, err
	}

	var row *sql.Row
	var id string

	row = db.QueryRow(`SELECT id FROM messages WHERE id = $1`, ids[0])

	if err = row.Scan(&id); err == nil {
		ids, err = getUUID()
	}

	return ids, nil
}

func MakeMessage(content string, c *websocket.Conn) error {

	var (
		ids []string
		err error
	)

	if ids, err = getUUID(); err != nil {
		return err
	}

	query := fmt.Sprintf("INSERT INTO messages (id, content) VALUES ( '%s', '%s' )", ids[0], content)

	if _, err = db.Exec(query); err != nil {
		return err
	}

	broadcast <- message{
		content,
		c,
	}

	go Delete(ids[0])

	return nil

}

func GetMessages() ([]string, error) {
	var (
		rows *sql.Rows
		err  error
	)

	if rows, err = db.Query(`SELECT content FROM messages`); err != nil {
		return nil, err
	}

	var result []string

	for rows.Next() {
		var a string

		if err = rows.Scan(&a); err != nil {
			return nil, err
		}

		result = append(result, a)

	}

	return result, nil

}
