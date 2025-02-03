package api

import (
	pb "bookserver_git/api/proto/v1"
	"bookserver_git/internal/domain"
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type Repository interface {
	SaveBookToDatabase(book domain.Book, ctx context.Context) (domain.Book, error)
	GetBookFromDatebase(id uint, ctx context.Context) (domain.Book, error)
	DeleteBookFromDatebase(id uint, ctx context.Context) error
	UpdateBookFromDatabase(book domain.Book, ctx context.Context) error
	AllBooksFromDatabase(ctx context.Context) ([]domain.Book, error)

	SaveUserToDatabase(ctx context.Context, user domain.User) (domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	SaveSessionToDatabase(ctx context.Context, session domain.Session) error
	GetUserByToken(ctx context.Context, token string) (int, error)
}

type Server struct {
	Database Repository
}

const authScheme = "Bearer"

func (s Server) AddBook(ctx context.Context, request *pb.AddBookRequest) (*pb.AddBookResponse, error) {
	token, err := auth.AuthFromMD(ctx, authScheme)
	if err != nil {
		return nil, fmt.Errorf("AuthFromMD: %w", err)
	}

	userID, err := s.Database.GetUserByToken(ctx, token)

	newBook := domain.Book{
		Title:  request.Title,
		Year:   int(request.Year),
		UserID: userID,
	}
	result, err := s.Database.SaveBookToDatabase(newBook, ctx)
	fmt.Println(result)
	if err != nil {
		return nil, err
	}
	// log.Info("Добавили книгу в bd")

	return &pb.AddBookResponse{Book: &pb.Book{
		Id:     int64(result.ID),
		Title:  result.Title,
		Year:   int64(result.Year),
		UserId: int64(result.UserID),
	}}, nil
}

func (s Server) GetBook(ctx context.Context, request *pb.GetBookRequest) (*pb.GetBookResponse, error) {
	idint := request.Id
	book, err := s.Database.GetBookFromDatebase(uint(idint), ctx)
	if err != nil {
		return nil, err
	}
	return &pb.GetBookResponse{Book: &pb.Book{
		Id:     int64(book.ID),
		Title:  book.Title,
		Year:   int64(book.Year),
		UserId: int64(book.UserID),
	}}, nil

}

func (s Server) DeleteBook(ctx context.Context, request *pb.DeleteBookRequest) (*pb.DeleteBookResponse, error) {

	idint := request.Id
	err := s.Database.DeleteBookFromDatebase(uint(idint), ctx)
	if err != nil {
		return nil, err
	}
	return nil, nil

}

func (s Server) UpdateBook(ctx context.Context, request *pb.UpdateBookRequest) (*pb.UpdateBookResponse, error) {
	newBook := domain.Book{
		ID:    int(request.Id),
		Title: request.Title,
		Year:  int(request.Year),
	}

	err := s.Database.UpdateBookFromDatabase(newBook, ctx)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s Server) AllBooks(ctx context.Context, request *pb.AllBooksRequests) (*pb.AllBooksResponse, error) {

	books, err := s.Database.AllBooksFromDatabase(ctx)
	if err != nil {
		return nil, err
	}
	pbBooks := make([]*pb.Book, len(books))
	for i := range books {
		pbBooks[i] = toBook(books[i])
	}

	return &pb.AllBooksResponse{Books: pbBooks}, nil
}

func (s Server) Registration(ctx context.Context, request *pb.RegistrationRequest) (*pb.RegistrationResponse, error) {
	newUser := domain.User{
		Email:    request.Email,
		Password: request.Password,
	}
	registeredUser, err := s.Database.SaveUserToDatabase(ctx, newUser)
	if err != nil {
		return nil, err
	}
	return &pb.RegistrationResponse{Id: int64(registeredUser.ID)}, nil

}

var (
	ErrInvalidPassword = errors.New("invalid password")
)

func (s Server) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := s.Database.GetUserByEmail(ctx, request.Email)
	if err != nil {
		return nil, err
	}

	if user.Password != request.Password {
		return nil, ErrInvalidPassword
	}
	ip, err := originFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	token := uuid.New().String()

	session := domain.Session{
		UserID:    user.ID,
		Token:     token,
		IP:        ip,
		UserAgent: "", // TODO
	}
	err = s.Database.SaveSessionToDatabase(ctx, session)
	if err != nil {
		return nil, err
	}

	err = grpc.SendHeader(ctx, metadata.MD{"authorization": {token}})

	return &pb.LoginResponse{User: &pb.User{
		Id:    int64(user.ID),
		Email: user.Email,
	}}, nil
}

func originFromCtx(ctx context.Context) (string, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("peer.FromContext: Error")
	}
	host, _, err := net.SplitHostPort(p.Addr.String())
	if err != nil {
		return "", fmt.Errorf("net.SplitHostPort: %w", err)
	}
	return host, nil
}

func toBook(u domain.Book) *pb.Book {
	return &pb.Book{
		Id:    int64(u.ID),
		Title: u.Title,
		Year:  int64(u.Year),
	}
}

// func (s Server) AddBook(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	log, found := logger.FromContext(ctx)
// 	if !found {
// 		handleError(w, http.StatusInternalServerError, errors.New("Help"))
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")

// 	jsong, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		handleError(w, http.StatusInternalServerError, err)
// 		return
// 	}

// 	var newBook domain.Book
// 	err = json.Unmarshal(jsong, &newBook)
// 	if err != nil {
// 		handleError(w, http.StatusBadRequest, err)
// 		return
// 	}
// 	if newBook.Title == "" {
// 		handleError(w, http.StatusBadRequest, fmt.Errorf("title not found"))
// 		return
// 	}
// if len(newBook.Authors) == 0 {
// 	handleError(w, http.StatusBadRequest, fmt.Errorf("authors not found"))
// 	return
// }
// 	if newBook.Year == 0 {
// 		handleError(w, http.StatusBadRequest, fmt.Errorf("year not found"))
// 		return
// 	}

// 	result, errRes := s.Database.SaveBookToDatabase(newBook, ctx)
// 	if errRes != nil {
// 		handleError(w, http.StatusInternalServerError, errRes)
// 		return
// 	}

// 	log.Info("Добавили книгу в bd")

// 	data, err := json.Marshal(result)
// 	if err != nil {
// 		handleError(w, http.StatusInternalServerError, err)
// 		return
// 	}
// 	w.Write(data)
// }

// func (s Server) GetBook(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	log, found := logger.FromContext(ctx)
// 	if !found {
// 		handleError(w, http.StatusInternalServerError, errors.New("Help"))
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	idstr := r.URL.Query().Get("id")
// 	idint, err := strconv.Atoi(idstr)
// 	if err != nil {
// 		handleError(w, http.StatusBadRequest, err)
// 		return
// 	}

// 	book, errRes := s.Database.GetBookFromDatebase(uint(idint), ctx)
// 	if errRes != nil {
// 		handleError(w, http.StatusInternalServerError, errRes)
// 		return
// 	}

// 	log.Info("Получили книгу из bd")

// 	data, err := json.Marshal(book)
// 	if err != nil {
// 		handleError(w, http.StatusInternalServerError, err)
// 		return
// 	}
// 	w.Write(data)
// }

// func (s Server) DeleteBook(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	log, found := logger.FromContext(ctx)
// 	if !found {
// 		handleError(w, http.StatusInternalServerError, errors.New("Help"))
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	idstr := r.URL.Query().Get("id")
// 	idint, err := strconv.Atoi(idstr)
// 	if err != nil {
// 		handleError(w, http.StatusBadRequest, err)
// 		return
// 	}
// 	err = s.Database.DeleteBookFromDatebase(uint(idint), ctx)
// 	if err != nil {
// 		handleError(w, http.StatusBadRequest, err)
// 		return
// 	}

// 	log.Info("Удалили книгу из bd")

// 	w.WriteHeader(http.StatusNoContent)
// }

// func (s Server) UpdateBook(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	log, found := logger.FromContext(ctx)
// 	if !found {
// 		handleError(w, http.StatusInternalServerError, errors.New("Help"))
// 		return
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	data, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		handleError(w, http.StatusInternalServerError, err)
// 		return
// 	}
// 	var book domain.Book
// 	err = json.Unmarshal(data, &book)
// 	if err != nil {
// 		handleError(w, http.StatusBadRequest, err)
// 		return
// 	}

// 	err = s.Database.UpdateBookFromDatabase(book, ctx)
// 	if err != nil {
// 		handleError(w, http.StatusBadRequest, err)
// 		return
// 	}

// 	log.Info("Обновили книгу в bd")

// 	w.WriteHeader(http.StatusNoContent)
// }

// func (s Server) AllBooks(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	w.Header().Set("Content-Type", "application/json")

// 	query := r.URL.Query()
// 	limit := query.Get("limit")

// 	var limitBooks []domain.Book
// 	tempAllBoks, errRes := s.Database.AllBooksFromDatabase(ctx)
// 	if errRes != nil {
// 		handleError(w, http.StatusInternalServerError, errRes)
// 		return
// 	}
// 	if limit != "" {
// 		limitNum, err := strconv.Atoi(limit)
// 		if err != nil {
// 			handleError(w, http.StatusBadRequest, errors.New("invalid limit parameter"))
// 			return
// 		}
// 		if limitNum > len(tempAllBoks) {
// 			limitNum = len(tempAllBoks)
// 		}
// 		limitBooks = tempAllBoks[:limitNum]
// 	} else {
// 		limitBooks = tempAllBoks
// 	}

// 	data, err := json.Marshal(limitBooks)
// 	if err != nil {
// 		handleError(w, http.StatusInternalServerError, err)
// 		return
// 	}
// 	w.Write(data)
// }
