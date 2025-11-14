package storage

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
)

type PostType string

const (
	TEXT PostType = "text"
	LINK PostType = "link"
)

type UpDownVote int

const (
	UPVOTE   UpDownVote = 1
	DOWNVOTE UpDownVote = -1
)

type RawPost struct {
	Type     PostType `json:"type"`
	Category string   `json:"category"`
	Title    string   `json:"title"`
	Content  string   `json:"-"`
}

type PostAuthor struct {
	Name string `json:"username"`
	ID   string `json:"id"`
}

type Comment struct {
	ID          string     `json:"id"`
	Body        string     `json:"body"`
	CreatedTime string     `json:"created"`
	Author      PostAuthor `json:"author"`
}

type Vote struct {
	UserID string     `json:"user"`
	Vote   UpDownVote `json:"vote"`
}

type Post struct {
	RawPost
	ID               string     `json:"id"`
	Author           PostAuthor `json:"author"`
	Score            int        `json:"score"`
	Views            int        `json:"views"`
	CreatedTime      string     `json:"created"`
	UpvotePercentage int        `json:"upvotePercentage"`
	Votes            []Vote     `json:"votes"`
	Comments         []Comment  `json:"comments"`
}

type PostStorage interface {
	AddPost(RawPost, User) Post
}

type PostInMemStorage struct {
	posts map[string]Post
	mu    *sync.RWMutex
}

func NewPostInMemStorage() *PostInMemStorage {
	return &PostInMemStorage{map[string]Post{}, &sync.RWMutex{}}
}

func (s *PostInMemStorage) AddPost(rawPost RawPost, author User) Post {
	post := Post{}
	post.Type = rawPost.Type
	post.Category = rawPost.Category
	post.Title = rawPost.Title
	post.Content = rawPost.Content
	post.ID = uuid.NewString()
	post.Author = PostAuthor{
		Name: author.Name,
		ID:   author.ID,
	}
	post.Score = 1
	post.Views = 1
	post.CreatedTime = time.Now().Format(time.RFC3339)
	post.UpvotePercentage = 100
	post.Votes = []Vote{{UserID: author.ID, Vote: UPVOTE}}
	post.Comments = []Comment{}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.posts[post.ID] = post

	return post
}

func (p Post) MarshalJSON() ([]byte, error) {
	type Alias Post
	aux := struct {
		*Alias
	}{
		Alias: (*Alias)(&p),
	}

	data, err := json.Marshal(aux)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	switch p.Type {
	case TEXT:
		result["text"] = p.Content
	case LINK:
		result["url"] = p.Content
	}

	return json.Marshal(result)
}

func (p *Post) UnmarshalJSON(data []byte) error {
	type Alias Post
	aux := &struct {
		*Alias
		Text string `json:"text"`
		URL  string `json:"url"`
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch p.Type {
	case TEXT:
		p.Content = aux.Text
	case LINK:
		p.Content = aux.URL
	}

	return nil
}
