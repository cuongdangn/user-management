package webserver

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"user-management/log"
	message "user-management/message"
	"user-management/webserver/token"

	"github.com/gorilla/mux"
	"google.golang.org/protobuf/proto"
)

func (server *WebServer) getUserClaimFromSessionKey(r *http.Request) (*token.UserClaims, error) {
	cookie, err := r.Cookie("session_key")
	if err != nil {
		return nil, http.ErrNoCookie
	}

	tokenStr := cookie.Value
	claims, err := server.tokenMaker.VerifyToken(tokenStr)

	if err != nil {
		return nil, http.ErrNoCookie
	}

	return claims, nil
}

func (server *WebServer) getLoginPage(w http.ResponseWriter, r *http.Request) {
	_, err := server.getUserClaimFromSessionKey(r)
	if err != nil {
		server.webtemplate.ExecuteTemplate(w, "Login", struct {
			DisplayWarning string
			Message        string
		}{
			DisplayWarning: "none",
			Message:        "none",
		})
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (server *WebServer) getHomePage(w http.ResponseWriter, r *http.Request) {
	claim, err := server.getUserClaimFromSessionKey(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		messageRequest := message.GetUserInfoReq{Username: claim.Username}
		content, err := proto.Marshal(&messageRequest)

		messageProto := message.MessageRequest{Type: message.MessageRequest_GET_USERINFO, Content: content}
		data, _ := proto.Marshal(&messageProto)
		connection, err := server.tcpClient.SendTCPData(data)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println(err.Error())
			return
		}

		bytesMessage, err := server.tcpClient.ReadTCPData(connection)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		messageOut := message.GetUserInfoRes{}
		err = proto.Unmarshal(bytesMessage, &messageOut)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		server.tcpClient.PutConnection(connection)

		server.webtemplate.ExecuteTemplate(w, "Profile", struct {
			Username string
			Nickname string
			Ava      string
		}{
			Username: claim.Username,
			Nickname: messageOut.Nickname,
			Ava:      "http://localhost:1234/profilepicture/" + claim.Username,
		})
	}
}

func (server *WebServer) userLogin(w http.ResponseWriter, r *http.Request) {
	_, err := server.getUserClaimFromSessionKey(r)
	if err == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther) // doublicate getUserNameFromSessionKey
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	log.Log.InfoLogger.Println("Login:" + username)
	messageLogin := message.UserLoginReq{Username: username, Password: password}
	content, err := proto.Marshal(&messageLogin)
	messageProto := message.MessageRequest{Type: message.MessageRequest_USERLOGIN, Content: content}
	data, err := proto.Marshal(&messageProto)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Log.ErrorLogger.Println(err.Error())
		return
	}

	connection, err := server.tcpClient.SendTCPData(data)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Log.ErrorLogger.Println(err.Error())
		return
	}

	bytesMessage, err := server.tcpClient.ReadTCPData(connection)

	messageOut := message.UserLoginRes{Code: 500}

	err = proto.Unmarshal(bytesMessage, &messageOut)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Log.ErrorLogger.Println(err.Error())
		return
	}
	server.tcpClient.PutConnection(connection)

	if messageOut.Code == 200 {
		expirationTime := time.Now().Add(time.Hour * 24) // todo: put to config
		tokenString, err := server.tokenMaker.CreateToken(username, messageOut.Nickname, time.Hour*24)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Log.ErrorLogger.Println(err.Error())
			return
		}

		http.SetCookie(w,
			&http.Cookie{
				Name:    "session_key",
				Value:   tokenString,
				Expires: expirationTime,
			})
		http.Redirect(w, r, "/", http.StatusSeeOther)
		log.Log.InfoLogger.Println("Login: " + username + " " + "success")
	} else {
		log.Log.ErrorLogger.Println(messageOut.Message)
		server.webtemplate.ExecuteTemplate(w, "Login", struct {
			DisplayWarning string
			Message        string
		}{
			DisplayWarning: "inline",
			Message:        "Invalid login or password. Please try again.",
		})

	}
}

func (server *WebServer) getEditNicknamePage(w http.ResponseWriter, r *http.Request) {
	server.webtemplate.ExecuteTemplate(w, "Editnickname", struct {
		DisplayAlert string
		Message      string
		TypeAlert    string
	}{
		DisplayAlert: "none",
		Message:      "none",
		TypeAlert:    "danger",
	})
}

func (server *WebServer) editNickName(w http.ResponseWriter, r *http.Request) {
	claims, err := server.getUserClaimFromSessionKey(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := claims.Username
	newNickname := r.FormValue("nickname")
	log.Log.InfoLogger.Println("Edict Nick Name:" + username + " " + newNickname)
	messageRequest := message.UpdateNickNameReq{Username: username, Nickname: newNickname}
	content, err := proto.Marshal(&messageRequest)

	if err != nil {
		log.Log.ErrorLogger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	messageProto := message.MessageRequest{Type: 1, Content: content}
	data, err := proto.Marshal(&messageProto)

	if err != nil {
		log.Log.ErrorLogger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	connection, err := server.tcpClient.SendTCPData(data)

	if err != nil {
		log.Log.ErrorLogger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	bytesMessage, err := server.tcpClient.ReadTCPData(connection)
	if err != nil {
		log.Log.ErrorLogger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	messageOut := message.UpdateNickNameRes{Code: 500}
	err = proto.Unmarshal(bytesMessage, &messageOut)
	if err != nil {
		log.Log.ErrorLogger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	server.tcpClient.PutConnection(connection)
	w.WriteHeader(int(messageOut.Code))
	if messageOut.Code == 200 {
		server.webtemplate.ExecuteTemplate(w, "Editnickname", struct {
			DisplayAlert string
			Message      string
			TypeAlert    string
		}{
			DisplayAlert: "inline",
			Message:      "Edit Success",
			TypeAlert:    "success",
		})
	} else {
		server.webtemplate.ExecuteTemplate(w, "Editnickname", struct {
			DisplayAlert string
			Message      string
			TypeAlert    string
		}{
			DisplayAlert: "inline",
			Message:      "Invalid nickname",
			TypeAlert:    "danger",
		})
	}

}

func (server *WebServer) getEditProfilePicture(w http.ResponseWriter, r *http.Request) {
	_, err := server.getUserClaimFromSessionKey(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	server.webtemplate.ExecuteTemplate(w, "EditProfilePicture", struct {
		DisplayAlert string
		Message      string
		TypeAlert    string
	}{
		DisplayAlert: "none",
		Message:      "Invalid nickname",
		TypeAlert:    "danger",
	})
}

func (server *WebServer) editProfilePicture(w http.ResponseWriter, r *http.Request) {
	claims, err := server.getUserClaimFromSessionKey(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := claims.Username
	log.Log.InfoLogger.Println("Edit Profile Picture: " + username)
	file, fileHeader, err := r.FormFile("file")

	if err != nil {
		log.Log.ErrorLogger.Println(err.Error())
		server.webtemplate.ExecuteTemplate(w, "EditProfilePicture", struct {
			DisplayAlert string
			Message      string
			TypeAlert    string
		}{
			DisplayAlert: "inline",
			Message:      err.Error(),
			TypeAlert:    "danger",
		})
		return
	}

	if fileHeader.Size > 1024*1024 { // 1 MB
		log.Log.InfoLogger.Println("The uploaded image is too big. Please use an image less than 1MB in size")
		server.webtemplate.ExecuteTemplate(w, "EditProfilePicture", struct {
			DisplayAlert string
			Message      string
			TypeAlert    string
		}{
			DisplayAlert: "inline",
			Message:      "The uploaded image is too big. Please use an image less than 1MB in size",
			TypeAlert:    "danger",
		})
		return
	}

	filetype, err := GetFileContentType(file)

	if err != nil || (filetype != "image/jpeg" && filetype != "image/png") {
		log.Log.InfoLogger.Println("The provided file format is not allowed. Please upload a JPEG or PNG image")
		server.webtemplate.ExecuteTemplate(w, "EditProfilePicture", struct {
			DisplayAlert string
			Message      string
			TypeAlert    string
		}{
			DisplayAlert: "inline",
			Message:      "The provided file format is not allowed. Please upload a JPEG or PNG image",
			TypeAlert:    "danger",
		})
		return
	}
	f, err := os.Create("./db/images/" + username)
	if err != nil {
		fmt.Println(err)
		return
	}

	server.webtemplate.ExecuteTemplate(w, "EditProfilePicture", struct {
		DisplayAlert string
		Message      string
		TypeAlert    string
	}{
		DisplayAlert: "inline",
		Message:      "Updated",
		TypeAlert:    "success",
	})
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.Copy(f, file)
	defer f.Close()
	defer file.Close()

}

func (server *WebServer) getProfilePicture(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	img, err := os.Open("./db/images/" + username)
	if err != nil {
		img, _ = os.Open("./db/images/default/1")
	}
	defer img.Close()
	w.Header().Set("Content-Type", "image/png")
	io.Copy(w, img)
}

func (server *WebServer) getUserInfo(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	w.Write([]byte(username))
}
