package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"server/internal/db"
	"server/internal/domain"
	"server/internal/logger"
	"strconv"
)

type Server struct {
	Database *db.Repository
}

func (s Server) AddBook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log, found := logger.FromContext(ctx)
	if !found {
		handleError(w, http.StatusInternalServerError, errors.New("Help"))
		return
	}

	w.Header().Set("Content-Type", "application/json")

	jsong, err := io.ReadAll(r.Body)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err)
		return
	}

	var newBook domain.Book
	err = json.Unmarshal(jsong, &newBook)
	if err != nil {
		handleError(w, http.StatusBadRequest, err)
		return
	}
	if newBook.Title == "" {
		handleError(w, http.StatusBadRequest, fmt.Errorf("title not found"))
		return
	}
	// if len(newBook.Authors) == 0 {
	// 	handleError(w, http.StatusBadRequest, fmt.Errorf("authors not found"))
	// 	return
	// }
	if newBook.Year == 0 {
		handleError(w, http.StatusBadRequest, fmt.Errorf("year not found"))
		return
	}

	result, errRes := s.Database.SaveBookToDatabase(newBook, ctx)
	if errRes != nil {
		handleError(w, http.StatusInternalServerError, errRes)
		return
	}

	log.Info("Добавили книгу в bd")

	data, err := json.Marshal(result)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err)
		return
	}
	w.Write(data)
}

func (s Server) GetBook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log, found := logger.FromContext(ctx)
	if !found {
		handleError(w, http.StatusInternalServerError, errors.New("Help"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	idstr := r.URL.Query().Get("id")
	idint, err := strconv.Atoi(idstr)
	if err != nil {
		handleError(w, http.StatusBadRequest, err)
		return
	}

	book, errRes := s.Database.GetBookFromDatebase(uint(idint), ctx)
	if errRes != nil {
		handleError(w, http.StatusInternalServerError, errRes)
		return
	}

	log.Info("Получили книгу из bd")

	data, err := json.Marshal(book)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err)
		return
	}
	w.Write(data)
}

func (s Server) DeleteBook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log, found := logger.FromContext(ctx)
	if !found {
		handleError(w, http.StatusInternalServerError, errors.New("Help"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	idstr := r.URL.Query().Get("id")
	idint, err := strconv.Atoi(idstr)
	if err != nil {
		handleError(w, http.StatusBadRequest, err)
		return
	}
	err = s.Database.DeleteBookFromDatebase(uint(idint), ctx)
	if err != nil {
		handleError(w, http.StatusBadRequest, err)
		return
	}

	log.Info("Удалили книгу из bd")

	w.WriteHeader(http.StatusNoContent)
}

func (s Server) UpdateBook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log, found := logger.FromContext(ctx)
	if !found {
		handleError(w, http.StatusInternalServerError, errors.New("Help"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	data, err := io.ReadAll(r.Body)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err)
		return
	}
	var book domain.Book
	err = json.Unmarshal(data, &book)
	if err != nil {
		handleError(w, http.StatusBadRequest, err)
		return
	}

	err = s.Database.UpdateBookFromDatabase(book, ctx)
	if err != nil {
		handleError(w, http.StatusBadRequest, err)
		return
	}

	log.Info("Обновили книгу в bd")

	w.WriteHeader(http.StatusNoContent)
}

func (s Server) AllBooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()
	limit := query.Get("limit")

	var limitBooks []domain.Book
	tempAllBoks, errRes := s.Database.AllBooksFromDatabase(ctx)
	if errRes != nil {
		handleError(w, http.StatusInternalServerError, errRes)
		return
	}
	if limit != "" {
		limitNum, err := strconv.Atoi(limit)
		if err != nil {
			handleError(w, http.StatusBadRequest, errors.New("invalid limit parameter"))
			return
		}
		if limitNum > len(tempAllBoks) {
			limitNum = len(tempAllBoks)
		}
		limitBooks = tempAllBoks[:limitNum]
	} else {
		limitBooks = tempAllBoks
	}

	data, err := json.Marshal(limitBooks)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err)
		return
	}
	w.Write(data)
}
