package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	dbw "./db"
)

//----------------- Users ----------------\\

// Data for json from front
type Data struct {
	Ids []string
}

// GetUser get girl from vk by screenname
func getUser(screenname string) (user dbw.User, err error) {
	cmd := exec.Command("python3", append([]string{SCRIPTSPATH + "get_girl_by_vkid.py"}, screenname)...)
	bytes, err := cmd.Output()

	if err == nil {
		err = json.Unmarshal(bytes, &user)
	}

	return user, err
}

func RandomUserHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	if sex, ok := r.Form["sex"]; ok {
		val, err := strconv.ParseBool(sex[0])

		if err == nil {
			users, _ := dbwrap.GetRandomUsers(3, val)
			b, _ := json.Marshal(users)

			fmt.Fprintln(w, string(b))
		}
	}
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		// TODO: add admin checking

		if vkid, ok := r.Form["vkid"]; ok {
			val, err := strconv.Atoi(vkid[0])

			if err == nil {
				dbwrap.DeleteUser(val)
			}
		}
	}
}

func AddUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		r.ParseForm()

		if url, ok := r.Form["url"]; ok {
			pieces := strings.Split(url[0], "/")
			scname := pieces[len(pieces)-1]

			user, err := getUser(scname)

			if err == nil {
				dbwrap.AddUser(user)
			} else {
				log.Println(err)
			}
		}
	}
}

func UpdateUserStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var d Data
		err := json.NewDecoder(r.Body).Decode(&d)
		if err != nil {
			panic(err)
		}
		var ids = []int{}
		for _, s := range d.Ids {
			i, _ := strconv.Atoi(s)
			ids = append(ids, i)
		}
		dbwrap.UpdateUserStats(ids)
		log.Println(ids)
	}
}

func UpdateUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		vkid, _ := strconv.Atoi(r.FormValue("vkid"))
		name := r.FormValue("name")
		gender := r.FormValue("sex")
		photo_url := r.FormValue("photo_url")

		var sex bool
		if gender == "male" {
			sex = true
		} else {
			sex = false
		}

		dbwrap.UpdateUserInfo(vkid, name, sex, photo_url)
		http.Redirect(w, r, "/", 302)
	}
}

//----------------- Keys ----------------\\

func checkAdminCookie(r *http.Request) bool {
	sess, _ := cookiestore.Get(r, SESSCODE)
	fmt.Println(sess.Values)

	if uid, ok := sess.Values["uid"]; ok {
		if val, err := strconv.Atoi(uid.(string)); err == nil {
			isAdmin, _ := dbwrap.IsUserAdmin(val)
			return isAdmin
		}
	}

	return false
}

func GenerateKeyHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	if creatorID, ok := r.Form["creator_id"]; ok && checkAdminCookie(r) {
		if val, err := strconv.Atoi(creatorID[0]); err == nil {
			key, err := dbwrap.GenerateKey(val)

			if err == nil {
				b, _ := json.Marshal(key)
				fmt.Fprintf(w, string(b))
			} else {
				fmt.Fprintf(w, "permission denied")
			}
		}
	}
}

func GetUsersKeysHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	if uid, ok := r.Form["uid"]; ok && checkAdminCookie(r) {
		if val, err := strconv.Atoi(uid[0]); err == nil {
			keys, err := dbwrap.GetUsersKeys(val)

			if err == nil {
				b, _ := json.Marshal(keys)

				fmt.Fprintf(w, string(b))
			}
		}
	}
}

func CheckKeyHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	if key, ok := r.Form["key"]; ok {
		valid, err := dbwrap.CheckKeyValidity(key[0])

		if err == nil {
			b, _ := json.Marshal(struct {
				IsValid bool `json:"is_valid"`
			}{valid})

			fmt.Fprintf(w, string(b))
		}
	}
}

//----------------- Admins ----------------\\
func CreateAdminHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	sess, _ := cookiestore.Get(r, SESSCODE)

	admin := dbwrap.CreateUser()
	b, _ := json.Marshal(admin)

	sess.Values["uid"] = strconv.Itoa(admin.UID)
	sess.Save(r, w)

	fmt.Fprintf(w, string(b))
}

func CheckAdminHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	if uid, ok := r.Form["uid"]; ok {
		if val, err := strconv.Atoi(uid[0]); err == nil {
			isAdmin, err := dbwrap.IsUserAdmin(val)

			if err == nil {
				b, _ := json.Marshal(struct {
					IsAdmin bool `json:"is_admin"`
				}{isAdmin})
				fmt.Fprintf(w, string(b))
			}
		}
	}
}

func GiveAdminHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	if uid, ok := r.Form["uid"]; ok {
		if key, ok := r.Form["key"]; ok {
			if val, err := strconv.Atoi(uid[0]); err == nil {
				err = dbwrap.GiveAdminPrivs(val, key[0])

				if err == nil {
					fmt.Fprintf(w, "ok")
				} else {
					fmt.Fprintf(w, "permission denied")
				}
			}
		}
	}
}
