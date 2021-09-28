package coreserver

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"time"
	config "user-management/config"
	"user-management/coreserver/cache"
	"user-management/log"
	message "user-management/message"
	"user-management/tcp"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"

	"github.com/golang/protobuf/proto"
)

var (
	ErrSystem  = errors.New("System err")
	ErrTimeOut = errors.New("Time out")
)

type CoreServer struct {
	tcpServer     *tcp.TCPServer
	cache         *cache.UserCache
	sqlConnection *sql.DB
}

func dbConn(cfg *config.Config) (db *sql.DB) {
	dbDriver := cfg.CoreServer.Database.DbDriver
	dbUser := cfg.CoreServer.Database.DbUser
	dbPass := cfg.CoreServer.Database.DbPass
	dbName := cfg.CoreServer.Database.DbName
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		log.Log.ErrorLogger.Panic(err.Error())
	}
	return db
}

func NewServer(cfg *config.Config) (*CoreServer, error) {
	server := &CoreServer{}
	tcpServer, err := tcp.NewTCPServer(cfg.CoreServer.Host+":"+cfg.CoreServer.Port, server.handleConnection)
	if err != nil {
		panic(err)
	}
	server.tcpServer = tcpServer

	server.cache = cache.NewCache(cfg.CoreServer.Redis.Host+":"+cfg.CoreServer.Redis.Port, cfg.CoreServer.Redis.Pass, cfg.CoreServer.Redis.Index, time.Hour*time.Duration(cfg.CoreServer.Redis.ExpireTime))
	server.sqlConnection = dbConn(cfg)
	return server, nil
}

func (server *CoreServer) getUserInfo(username string) (*cache.UserInfo, error) {
	userInfo, err := server.cache.Get(username)
	if err == nil {
		return userInfo, nil
	}

	var databaseUsername string
	var databasePassword string
	var databaseNickname string
	rows := server.sqlConnection.QueryRow("SELECT username, nickname, password FROM users WHERE username=?", username)
	err = rows.Scan(&databaseUsername, &databaseNickname, &databasePassword)
	if err != nil {
		return nil, err
	}

	userInfo = &cache.UserInfo{
		Password: databasePassword,
		Nickname: databaseNickname,
		Username: databaseUsername,
	}

	err = server.cache.Set(username, userInfo)
	if err != nil {
		fmt.Println(err)
	}
	return userInfo, nil
}

func (server *CoreServer) handleLoginRequest(c net.Conn, messBytes []byte) {
	mess := message.UserLoginReq{}
	messageProto := message.UserLoginRes{}
	err := proto.Unmarshal(messBytes, &mess)
	if err != nil {
		messageProto.Code = 500
		messageProto.Message = err.Error()
	} else {
		userInfo, err := server.getUserInfo(mess.Username)

		if err != nil {
			messageProto.Code = 500
		} else {
			err = bcrypt.CompareHashAndPassword([]byte(userInfo.Password), []byte(mess.Password))

			if err != nil {
				messageProto.Code = 401
				messageProto.Message = err.Error()
			} else {
				messageProto.Code = 200
				messageProto.Username = userInfo.Username
				messageProto.Nickname = userInfo.Nickname
			}
		}
	}
	data, _ := proto.Marshal(&messageProto)
	server.tcpServer.SendTCPData(c, data)
}

func (server *CoreServer) handleEditNickname(c net.Conn, messBytes []byte) {

	messageProto := message.UpdateNickNameRes{}
	mess := message.UpdateNickNameReq{}
	err := proto.Unmarshal(messBytes, &mess)
	if err != nil {
		messageProto.Code = 500
		messageProto.Message = err.Error()
	} else {
		db := server.sqlConnection
		insForm, err := db.Prepare("UPDATE users SET nickname=? WHERE username=?")
		if err != nil {
			messageProto.Code = 500
			messageProto.Message = err.Error()
		} else {
			messageProto.Code = 200
			_, err := insForm.Exec(mess.Nickname, mess.Username)
			if err != nil {
				messageProto.Code = 401
				messageProto.Message = err.Error()
			} else {
				server.cache.Del(mess.Username)
			}
		}
	}

	data, _ := proto.Marshal(&messageProto)
	server.tcpServer.SendTCPData(c, data)
}

func (server *CoreServer) handleGetUserInfo(c net.Conn, messBytes []byte) {
	messageProto := message.GetUserInfoRes{}
	mess := message.GetUserInfoReq{}
	err := proto.Unmarshal(messBytes, &mess)
	if err != nil {
		messageProto.Code = 500
		messageProto.Message = err.Error()
	} else {

		userInfor, err := server.getUserInfo(mess.Username)
		if err != nil {
			messageProto.Code = 500
			messageProto.Message = err.Error()
		} else {
			messageProto.Code = 200
			messageProto.Username = userInfor.Username
			messageProto.Nickname = userInfor.Nickname
		}
	}

	data, _ := proto.Marshal(&messageProto)
	server.tcpServer.SendTCPData(c, data)
}

func (server *CoreServer) handleRequest(c net.Conn, mess *message.MessageRequest) {
	switch mess.Type {
	case 0:
		server.handleLoginRequest(c, mess.Content)
	case 1:
		server.handleEditNickname(c, mess.Content)
	case 2:
		server.handleGetUserInfo(c, mess.Content)
	default:
		c.Close()
	}
}

func (server *CoreServer) handleConnection(c net.Conn) error {
	for {
		messReq, err := server.tcpServer.ReadTCPData(c)
		if err == nil {
			server.handleRequest(c, messReq)
		} else {
			c.Close()
		}
	}
}

func (server *CoreServer) Start() {
	server.tcpServer.Start()
}
