package controllers

import (
	"net/http"
	"strconv"

	"study-planner/utils"

	"github.com/gorilla/mux"
)

func taskID(w http.ResponseWriter, r *http.Request) (uint, bool) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil || id <= 0 {
		utils.JSONError(w, http.StatusBadRequest, "Invalid task ID")
		return 0, false
	}
	return uint(id), true
}
