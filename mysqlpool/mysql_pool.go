package mysqlpool

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlConnPool struct {
	username   string
	password   string
	hostname   string
	port       string
	database   string
	charset    string
	maxopen    int
	dbconnpool *sql.DB
}

func NewMysqlConnPool(username, password, hostname, port, database, charset string, maxopen int) (*MysqlConnPool, error) {
	db := &MysqlConnPool{}

	db.username = username
	db.password = password
	db.hostname = hostname
	if port != "" {
		db.port = port
	} else {
		db.port = "3306"
	}
	db.database = database

	if charset != "" {
		db.charset = charset
	} else {
		db.charset = "utf8"
	}
	if maxopen < 1 {
		db.maxopen = 1
	} else {
		db.maxopen = maxopen
	}

	var err error
	db.dbconnpool, err = sql.Open("mysql",
		db.username+":"+db.password+"@tcp("+db.hostname+":"+db.port+")/"+db.database+"?charset="+db.charset)

	db.dbconnpool.SetMaxOpenConns(db.maxopen)

	err = db.dbconnpool.Ping()
	if err != nil {
		return nil, err
	}
	return db, err
}

func (m *MysqlConnPool) GetDBConn() *sql.DB {
	return m.dbconnpool
}

func (m *MysqlConnPool) ClosePool() {
	m.dbconnpool.Close()
	m.dbconnpool = nil
}
