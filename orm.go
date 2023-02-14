package orm

import (
	"errors"
	"fmt"
	"github.com/ericnts/config"
	"github.com/ericnts/log"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
	"net/url"
	"strings"
	"time"
)

var DB *gorm.DB

func init() {
	options, err := config.Load[*Options]("db")
	if err != nil {
		log.WithError(err).Panicf("数据库初始化失败")
	}
	if options == nil {
		log.Panic("没有找到数据库配置")
	}
	err = initDB(options)
	if err != nil {
		log.WithError(err).Panicf("数据库初始化失败")
	}
}

func initDB(o *Options) error {
	log.Info("开始初始化数据库...")

	masterDialector, err := getDialector(o.Master)
	if err != nil {
		return err
	}
	DB, err = gorm.Open(masterDialector, &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
		Logger:                                   &Logger{LogLevel: logger.Info},
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: true, // AutoMigrate 会自动创建数据库外键约束，您可以在初始化时禁用此功能
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名，启用该选项后，`User` 表将是`user`
		},
	})
	if err != nil {
		return err
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	sqlDB.SetConnMaxIdleTime(time.Hour)
	sqlDB.SetConnMaxLifetime(time.Hour * 24)
	sqlDB.SetMaxOpenConns(max(o.Master.MaxOpen, 10))
	sqlDB.SetMaxIdleConns(max(o.Master.MaxIdle, 20))

	if o.Slave == nil {
		return nil
	}
	log.Info("开始初始化从数据库...")
	slaveDialector, err := getDialector(o.Slave)
	dbResolver := dbresolver.Register(dbresolver.Config{
		Replicas: []gorm.Dialector{slaveDialector},
	}).
		SetConnMaxIdleTime(time.Hour).
		SetConnMaxLifetime(time.Hour * 24).
		SetMaxOpenConns(max(o.Slave.MaxOpen, 10)).
		SetMaxIdleConns(max(o.Slave.MaxIdle, 20))
	if err := DB.Use(dbResolver); err != nil {
		return err
	}
	return nil
}

func getDialector(o *EntryOptions) (gorm.Dialector, error) {
	u := url.URL{
		Scheme:   o.Dialector,
		User:     url.UserPassword(o.Username, o.Password),
		Host:     o.Host,
		Path:     o.Path,
		RawQuery: o.RawQuery,
	}
	var dialector gorm.Dialector
	switch o.Dialector {
	case "sqlite":
		dialector = sqlite.Open(u.Path)
	case "mysql":
		if !strings.Contains(u.Host, "tcp") {
			u.Host = fmt.Sprintf("tcp(%s)", u.Host)
		}
		dialector = mysql.Open(u.String()[8:])
	case "sqlserver":
		dialector = sqlserver.Open(u.String())
	default:
		return nil, errors.New("未知的数据库配置模式")
	}
	log.Info(u.String())
	log.Infof("数据库模式：%s", o.Dialector)
	return dialector, nil
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
