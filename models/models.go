package models

import (
	"fmt"
	"git.sr.ht/~humaid/neatnote/modules/settings"
	_ "github.com/go-sql-driver/mysql" // MySQL driver support
	"github.com/hako/durafmt"
	_ "github.com/mattn/go-sqlite3" // SQLite driver support
	"log"
	"time"
	"xorm.io/core"
	"xorm.io/xorm"
)

var (
	engine *xorm.Engine
	tables []interface{}
)

func init() {
	tables = append(tables,
		new(User),
		new(Course),
		new(Post),
		new(Comment),
	)
}

// calcDuration calculates the duration between two timestamps.
// It will return a fancy formatted duration in text.
func calcDuration(unix int64) string {
	return durafmt.Parse(time.Now().Sub(time.Unix(unix, 0))).LimitFirstN(1).String()
}

// SetupEngine sets up an XORM engine according to the database configuration
// and syncs the schema.
func SetupEngine() *xorm.Engine {
	var err error
	dbConf := &settings.Config.DBConfig

	address := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8",
		dbConf.User, dbConf.Password, dbConf.Host, dbConf.Name)

	switch dbConf.Type {
	case settings.MySQL:
		engine, err = xorm.NewEngine("mysql", address)
	case settings.SQLite:
		engine, err = xorm.NewEngine("sqlite3", dbConf.Path)
	}

	if err != nil {
		log.Fatal("Unable to connect/load the database! ", err)
	}

	engine.SetMapper(core.GonicMapper{}) // So ID becomes 'id' instead of 'i_d'
	err = engine.Sync(tables...)         // Sync the schema of tables

	//cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
	//engine.SetDefaultCacher(cacher)

	if err != nil {
		log.Fatal("Unable to sync schema! ", err)
	}

	return engine
}
