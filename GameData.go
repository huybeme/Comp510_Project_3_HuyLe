package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

func Create_tables(database *sql.DB) {
	create_game_table := "CREATE TABLE IF NOT EXISTS players(" +
		"player_num INTEGER PRIMARY KEY," +
		"player_name TEXT NOT NULL," +
		"player_score INTEGER DEFAULT 0);"
	database.Exec(create_game_table)
}

func OpenDatabase(dbfile string) *sql.DB {
	database, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Fatal("Open Database Error", err)
	}
	return database
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatal(msg+" ", err)
	}
}

func DbExists(path string) bool {
	//_, err := os.Stat(path)	// another way of writing it
	//if err != nil {
	//	return true
	//}
	//if _, err := os.Stat(path); err == nil{
	//	fmt.Println("path exists")
	//	return true
	//}

	_, err := os.Stat(path)
	if os.IsExist(err) {
		fmt.Println("file not exist")
		return true
	}
	return false
}

func printPlayers(players []Players) {
	for i := range players {
		fmt.Println(players[i])
	}
}

type Players struct {
	num   int
	name  string
	score int
}

var (
	path          = "./GameDatabase.db"
	gameDB        = OpenDatabase(path)
	SortedPlayers = SortScores(gameDB)

	TopFive       []string
	emptyTable    bool
	LastHighScore int
)

//func main(){
//
//	defer gameDB.Close()
//
//	if DbExists(path) == false {
//		Create_tables(gameDB)
//	}
//
//	//AddPlayerName(gameDB, "Khai")
//
//	SortedPlayers = SortScores(gameDB)
//	printPlayers(SortedPlayers)
//	fmt.Println(len(SortedPlayers))
//	TopFive = GetTopFive(SortedPlayers)
//	for i := range TopFive{
//		fmt.Println(TopFive[i])
//	}
//}

func GetPlayerNum() int {
	var num int
	statement := "SELECT IFNULL(MAX(player_num), 0) FROM players"
	err := gameDB.QueryRow(statement).Scan(&num)
	checkErr(err, "Get Last Entry from Player Num error")
	return num
}

func UpdateScore(score int, id int) {
	statement := "UPDATE players SET player_score = ? WHERE player_num = ?"
	update_score, err := gameDB.Prepare(statement)
	checkErr(err, "update score error")
	execUpdate, err := update_score.Exec(score, id)
	check, err := execUpdate.RowsAffected()
	if check == 0 {
		log.Println("Score did not update")
	} else {
		log.Println("Score update Success")
	}
}

func GetTopFive(players []Players) ([]string, int) {
	var playerScore []string
	tempStr := fmt.Sprintf("|%-16s  %-6s ", "player", "score")
	lastScore := 0
	playerScore = append(playerScore, tempStr)
	if len(players) <= 5 {
		for i := range players {
			tempStr = fmt.Sprintf("|%-16s  %-6d ", players[i].name, players[i].score)
			playerScore = append(playerScore, tempStr)
			lastScore = players[i].score
		}
	} else {
		for i := 0; i < 5; i++ {
			tempStr = fmt.Sprintf("|%-16s  %-6d", players[i].name, players[i].score)
			playerScore = append(playerScore, tempStr)
			lastScore = players[i].score
		}
	}
	return playerScore, lastScore
}

func AddPlayerName(playername string) {
	statement := "INSERT INTO players (player_name, player_score) VALUES (?, ?)"
	prepped_statement, err := gameDB.Prepare(statement)
	checkErr(err, "add player prep statement error")
	//playerKey := 1001

	_, err = prepped_statement.Exec(playername, 0)
	checkErr(err, "add player exec error")
}

func SortScores(db *sql.DB) []Players {
	var data []Players
	rows, err := db.Query("SELECT * FROM players ORDER BY player_score desc")
	checkErr(err, "sort scores query error")
	for rows.Next() {
		var num int
		var name string
		var score int
		err2 := rows.Scan(&num, &name, &score)
		checkErr(err2, "sort scores query scan error")
		dataElement := Players{
			num:   num,
			name:  name,
			score: score,
		}
		data = append(data, dataElement)
	}
	return data
}

func isTableEmpty(db *sql.DB) bool {
	_, table_check := db.Query("select * from players;")

	if table_check == nil {
		return false
	}
	return true
}
