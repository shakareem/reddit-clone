package storage

import (
	"encoding/json"
	"fmt"
	"slices"
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
	NOVOTE   UpDownVote = 0
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
	AddPost(rawPost RawPost, authorName string, authorID string) Post
	DeletePost(id string) error
	GetPosts() []Post
	GetPost(id string) (Post, error)
	UpvotePost(postID, userID string) (Post, error)
	DownvotePost(postID, userID string) (Post, error)
	UnvotePost(postID, userID string) (Post, error)
}

type PostInMemStorage struct {
	posts map[string]Post
	mu    *sync.RWMutex
}

func NewPostInMemStorage() *PostInMemStorage {
	return &PostInMemStorage{map[string]Post{}, &sync.RWMutex{}}
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

func (s *PostInMemStorage) AddPost(rawPost RawPost, authorName string, authorID string) Post {
	post := Post{}
	post.Type = rawPost.Type
	post.Category = rawPost.Category
	post.Title = rawPost.Title
	post.Content = rawPost.Content
	post.ID = uuid.NewString()
	post.Author = PostAuthor{
		Name: authorName,
		ID:   authorID,
	}
	post.Score = 1
	post.Views = 1
	post.CreatedTime = time.Now().Format(time.RFC3339)
	post.UpvotePercentage = 100
	post.Votes = []Vote{{UserID: authorID, Vote: UPVOTE}}
	post.Comments = []Comment{}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.posts[post.ID] = post

	return post
}

func (s *PostInMemStorage) DeletePost(id string) error {
	_, ok := s.posts[id]
	if !ok {
		return fmt.Errorf("post with id %s not found", id)
	}

	delete(s.posts, id)
	return nil
}

func (s *PostInMemStorage) GetPosts() []Post {
	posts := make([]Post, 0, len(s.posts))
	for _, p := range s.posts {
		posts = append(posts, p)
	}

	return posts
}

func (s *PostInMemStorage) GetPost(id string) (Post, error) {
	post, ok := s.posts[id]
	if !ok {
		return Post{}, fmt.Errorf("post with id %s not found", id)
	}

	return post, nil
}

func (s *PostInMemStorage) UpvotePost(postID, userID string) (Post, error) {
	post, ok := s.posts[postID]
	if !ok {
		return Post{}, fmt.Errorf("post with id %s not found", postID)
	}
	oldVote, found := updateVote(&post, userID, UPVOTE)
	if found && oldVote == UPVOTE {
		return post, nil
	}
	if found && oldVote == DOWNVOTE {
		post.Score += 2
	} else {
		post.Score += 1
	}

	post.UpvotePercentage = countUpvotePercentage(post.Votes)
	s.posts[postID] = post
	return post, nil
}

func (s *PostInMemStorage) DownvotePost(postID, userID string) (Post, error) {
	post, ok := s.posts[postID]
	if !ok {
		return Post{}, fmt.Errorf("post with id %s not found", postID)
	}
	oldVote, found := updateVote(&post, userID, DOWNVOTE)
	if found && oldVote == DOWNVOTE {
		return post, nil
	}
	if found && oldVote == UPVOTE {
		post.Score -= 2
	} else {
		post.Score -= 1
	}

	post.UpvotePercentage = countUpvotePercentage(post.Votes)
	s.posts[postID] = post
	return post, nil
}

func (s *PostInMemStorage) UnvotePost(postID, userID string) (Post, error) {
	post, ok := s.posts[postID]
	if !ok {
		return Post{}, fmt.Errorf("post with id %s not found", postID)
	}
	oldVote, found := updateVote(&post, userID, NOVOTE)
	if found && oldVote == UPVOTE {
		post.Score -= 1
	}
	if found && oldVote == DOWNVOTE {
		post.Score += 1
	}

	post.UpvotePercentage = countUpvotePercentage(post.Votes)
	s.posts[postID] = post
	return post, nil
}

func updateVote(post *Post, userID string, newVote UpDownVote) (oldVote UpDownVote, found bool) {
	for i, vote := range post.Votes {
		if vote.UserID == userID {
			oldVote = vote.Vote
			if i == len(post.Votes)-1 {
				post.Votes = post.Votes[:len(post.Votes)-1]
			} else {
				post.Votes = slices.Delete(post.Votes, i, i+1)
			}
			found = true
			break
		}
	}

	if newVote != 0 {
		post.Votes = append(post.Votes, Vote{userID, newVote})
	}
	return
}

func countUpvotePercentage(votes []Vote) int {
	if len(votes) == 0 {
		return 0
	}

	upvotes := 0
	for _, vote := range votes {
		if vote.Vote == UPVOTE {
			upvotes += 1
		}
	}

	return int(float64(upvotes) / float64(len(votes)) * 100)
}
