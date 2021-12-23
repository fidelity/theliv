package router

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/fidelity/theliv/internal/rbac"
	"github.com/fidelity/theliv/pkg/auth/authmiddleware"
	"github.com/fidelity/theliv/pkg/database/etcd"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type FeebackData struct {
	User    *rbac.User `json:"user,omitempty"`
	Time    time.Time  `json:"time"`
	Message string     `json:"message"`
}
type Feedback struct {
	Message string `json:"message"`
}

func SubmitFeedback(r chi.Router) {
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var d Feedback
		err := json.NewDecoder(r.Body).Decode(&d)
		if err != nil {
			//TODO log & error handling
			log.Printf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}
		currentTime := time.Now()
		timestr := getTimeStr(currentTime)

		key := "/theliv/feedbacks/" + timestr

		user, err := authmiddleware.GetUser(r)
		if err != nil {
			//TODO log & error handling
			http.Error(w, err.Error(), 500)
			return
		}
		data := FeebackData{
			User:    user,
			Time:    currentTime,
			Message: d.Message,
		}

		err = etcd.Put(key, data)
		if err != nil {
			//TODO log & error handling
			http.Error(w, err.Error(), 500)
			return
		}

		s := "Feedback received."
		render.JSON(w, r, s)
	})

}

func getTimeStr(currentTime time.Time) string {
	year, month, day := currentTime.Date()
	hour, min, sec := currentTime.Clock()

	yyyy := strconv.Itoa(year)
	mm := fmt.Sprintf("%02d", int(month))
	dd := fmt.Sprintf("%02d", day)
	hr := fmt.Sprintf("%02d", hour)
	mn := fmt.Sprintf("%02d", min)
	ss := fmt.Sprintf("%02d", sec)
	return yyyy + mm + dd + hr + mn + ss
}
