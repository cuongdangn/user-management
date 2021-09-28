package webserver

import (
	"fmt"
	"net/http"
	"text/template"

	"user-management/config"
	"user-management/log"
	"user-management/tcp"
	"user-management/webserver/token"

	"github.com/gorilla/mux"
)

type WebServer struct {
	router      mux.Router
	tokenMaker  token.TokenMaker
	webtemplate template.Template
	tcpClient   *tcp.TCPClient
}

func NewWebServer(cfg *config.Config) (*WebServer, error) {
	tokenMaker, err := token.NewTokenMaker("EyfDgN3P7AovvXTvQewAsQV9dLREJLVOWlhNImyl")
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	tcpClient, err := tcp.NewTCPClient(cfg.WebServer.TCPClient.Host+":"+cfg.WebServer.TCPClient.Port, cfg.WebServer.TCPClient.NumConnection)
	if err != nil {
		return nil, err
	}
	server := &WebServer{
		tokenMaker:  *tokenMaker,
		webtemplate: *template.Must(template.ParseGlob("webserver/templates/*")),
		tcpClient:   tcpClient,
	}

	server.setupRouter()
	return server, nil
}

func (server *WebServer) setupRouter() {
	router := mux.NewRouter()
	router.HandleFunc("/login", server.getLoginPage).Methods("GET")
	router.HandleFunc("/login", server.userLogin).Methods("POST")
	router.HandleFunc("/edit/nickname", server.getEditNicknamePage).Methods("GET")
	router.HandleFunc("/edit/nickname", server.editNickName).Methods("POST")
	router.HandleFunc("/user/{username}", server.getUserInfo).Methods("GET")
	router.HandleFunc("/edit/pictureprofile", server.getEditProfilePicture).Methods("GET")
	router.HandleFunc("/edit/pictureprofile", server.editProfilePicture).Methods("POST")
	router.HandleFunc("/", server.getHomePage).Methods("GET")

	// images
	router.HandleFunc("/profilepicture/{username}", server.getProfilePicture).Methods("GET")
	server.router = *router
}

func (server *WebServer) Start(address string) error {
	log.Log.InfoLogger.Println("Webserver start listener: " + address)
	return http.ListenAndServe(address, &server.router)
}
