package schemas

import "time"

type suggester struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type Book struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Image       string      `json:"image"`
	Author      string      `json:"author"`
	Suggesters  []suggester `json:"suggesters"`
	CreatedAt   time.Time   `json:"created_at"`
}

// type DocumentRequest struct {
// 	Title   string `json:"title"`
// 	Content string `json:"content"`
// }

// type DocumentResponse struct {
// 	Title     string    `json:"title"`
// 	CreatedAt time.Time `json:"created_at"`
// 	Content   string    `json:"content"`
// }

type SearchResponse struct {
	Time      string `json:"time"`
	Hits      string `json:"hits"`
	Documents []Book `json:"books"`
}
