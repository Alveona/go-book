package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	"github.com/teris-io/shortid"

	"go-book/schemas"
)

const (
	elasticIndexName = "books"
	elasticTypeName  = "book"
)

var (
	elasticClient *elastic.Client
)

func main() {
	var err error
	for {
		elasticClient, err = elastic.NewClient(
			elastic.SetURL("http://127.0.0.1:9200"),
			elastic.SetSniff(false),
		)
		if err != nil {
			log.Println(err)
			// Retry every 3 seconds
			time.Sleep(3 * time.Second)
		} else {
			break
		}
	}
	r := gin.Default()
	r.POST("/books", bulkCreateBooks)
	r.GET("/search", searchEndpoint)
	if err = r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func bulkCreateBooks(c *gin.Context) {
	// Parse request
	var docs []schemas.Book
	if err := c.BindJSON(&docs); err != nil {
		errorResponse(c, http.StatusBadRequest, "Malformed request body")
		return
	}
	// Insert documents in bulk
	bulk := elasticClient.
		Bulk().
		Index(elasticIndexName).
		Type(elasticTypeName)
	for _, d := range docs {
		doc := schemas.Book{
			ID:          shortid.MustGenerate(),
			Title:       d.Title,
			Description: d.Description,
			Image:       d.Image,
			Author:      d.Author,
			Suggesters:  d.Suggesters,
			CreatedAt:   time.Now().UTC(),
		}
		bulk.Add(elastic.NewBulkIndexRequest().Id(doc.ID).Doc(doc))
	}
	if _, err := bulk.Do(c.Request.Context()); err != nil {
		log.Println(err)
		errorResponse(c, http.StatusInternalServerError, "Failed to create documents")
		return
	}
	c.Status(http.StatusOK)
}

func searchEndpoint(c *gin.Context) {
	// Parse request
	query := c.Query("query")
	if query == "" {
		errorResponse(c, http.StatusBadRequest, "Query not specified")
		return
	}
	offset := 0
	limit := 10
	if i, err := strconv.Atoi(c.Query("offset")); err == nil {
		offset = i
	}
	if i, err := strconv.Atoi(c.Query("limit")); err == nil {
		limit = i
	}
	// Perform search
	esQuery := elastic.NewMultiMatchQuery(query, "suggesters.firstName").
		// Fuzziness("2").
		MinimumShouldMatch("2")
	result, err := elasticClient.Search().
		Index(elasticIndexName).
		Query(esQuery).
		From(offset).Size(limit).
		Do(c.Request.Context())
	if err != nil {
		log.Println(err)
		errorResponse(c, http.StatusInternalServerError, "Something went wrong")
		return
	}
	res := schemas.SearchResponse{
		Time: fmt.Sprintf("%d", result.TookInMillis),
		Hits: fmt.Sprintf("%d", result.Hits.TotalHits),
	}
	// Transform search results before returning them
	docs := make([]schemas.Book, 0)
	for _, hit := range result.Hits.Hits {
		var doc schemas.Book
		json.Unmarshal(hit.Source, &doc)
		// if err := json.Unmarshal(hit.Source, &r); err != nil {
		// 	log.Errorf("ERROR UNMARSHALLING ES SUGGESTION RESPONSE: %v", err)
		// 	continue
		// }
		docs = append(docs, doc)
	}
	res.Documents = docs
	c.JSON(http.StatusOK, res)
}

func errorResponse(c *gin.Context, code int, err string) {
	c.JSON(code, gin.H{
		"error": err,
	})
}
