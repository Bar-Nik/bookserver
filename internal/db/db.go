package db

import (
	"bookserver_git/internal/domain"
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(rawDB *sql.DB) *Repository {
	return &Repository{db: rawDB}
}

func (d Repository) SaveBookToDatabase(book domain.Book, ctx context.Context) (domain.Book, error) {
	query := "INSERT INTO books (title, year_book) VALUES($1,$2) RETURNING *"
	err := d.db.QueryRowContext(ctx, query, book.Title, book.Year).Scan(&book.ID, &book.Title, &book.Year)
	if err != nil {
		return domain.Book{}, err
	}
	return book, nil

}

func (d Repository) GetBookFromDatebase(id uint, ctx context.Context) (domain.Book, error) {
	var book domain.Book
	query := "SELECT * FROM books WHERE id = $1"
	err := d.db.QueryRowContext(ctx, query, id).Scan(&book.ID, &book.Title, &book.Year)
	if err != nil {
		return domain.Book{}, err
	}
	return book, nil
}

func (d Repository) DeleteBookFromDatebase(id uint, ctx context.Context) error {
	query := "DELETE FROM books WHERE id = $1"
	_, err := d.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}

func (d Repository) UpdateBookFromDatabase(book domain.Book, ctx context.Context) error {
	id := book.ID
	title := book.Title
	year := book.Year
	query := "UPDATE books SET title = $1, year_book = $2 WHERE id = $3"
	_, err := d.db.ExecContext(ctx, query, title, year, id)
	if err != nil {
		return err
	}
	return nil
}

func (d Repository) AllBooksFromDatabase(ctx context.Context) ([]domain.Book, error) {

	var books []domain.Book
	query := "SELECT * FROM books"
	row, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	for row.Next() {
		var book domain.Book
		row.Scan(&book.ID, &book.Title, &book.Year)
		books = append(books, book)
	}

	return books, nil
}
