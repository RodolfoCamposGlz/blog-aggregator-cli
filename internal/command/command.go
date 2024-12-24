package command

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/RodolfoCamposGlz/internal/database"
	"github.com/RodolfoCamposGlz/internal/feed"
	"github.com/RodolfoCamposGlz/internal/state"
	"github.com/google/uuid"
)

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	Handler map[string]func(*state.State, Command) error
}

func (c *Commands) Register(name string, f func(*state.State, Command) error) error {
	if c.Handler == nil {
		return fmt.Errorf("handler map not initialized")
	}

	if _, exists := c.Handler[name]; exists {
		return fmt.Errorf("command %q is already registered", name)
	}
	c.Handler[name] = f
	return nil
}

func (c *Commands) Run(s *state.State, cmd Command) error {
	handler, exists := c.Handler[cmd.Name]
	if !exists {
		return fmt.Errorf("command %q does not exist", cmd.Name)
	}
	return handler(s, cmd)
}

// Helper function to create a feed follow
func createFeedFollow(s *state.State, userID, feedID uuid.UUID) (*database.CreateFeedFollowRow, error) {
	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    userID,
		FeedID:    feedID,
	}

	feedFollow, err := s.DB.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return nil, fmt.Errorf("error creating feed follow: %w", err)
	}
	return &feedFollow, nil
}

// Command implementations

func (c *Commands) RegisterUser(s *state.State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("you must provide a name")
	}

	name := cmd.Args[0]

	_, err := s.DB.GetUserByName(context.Background(), name)
	if err == nil {
		return fmt.Errorf("a user with the name '%s' already exists", name)
	} else if err != sql.ErrNoRows {
		return fmt.Errorf("error checking if user exists: %w", err)
	}

	// Create new user
	userID := uuid.New()
	currentTime := time.Now()
	params := database.CreateUserParams{
		ID:        userID,
		Name:      name,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	user, err := s.DB.CreateUser(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	log.Printf("User '%s' created successfully with ID: %s", name, userID)

	err = s.Config.SetUser(name)
	if err != nil {
		return fmt.Errorf("error setting the username: %w", err)
	}

	fmt.Println("User has been set:", user)
	return nil
}

func (c *Commands) DeleteUsers(s *state.State, cmd Command) error {
	err := s.DB.DeleteAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error deleting users: %w", err)
	}
	log.Println("Users deleted successfully")
	return nil
}

func (c *Commands) ListUsers(s *state.State, cmd Command) error {
	users, err := s.DB.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error getting users: %w", err)
	}

	for _, user := range users {
		if user.Name == s.Config.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func (c *Commands) Aggregator(s *state.State, cmd Command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("usage: agg <time_between_reqs>")
	}

	// Parse the time_between_reqs argument
	timeBetweenRequests, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("invalid time duration: %w", err)
	}

	// Print the collecting message
	fmt.Printf("Collecting feeds every %s\n", timeBetweenRequests)

	// Initialize a ticker
	ticker := time.NewTicker(timeBetweenRequests)
	defer ticker.Stop()

	// Fetch feeds in a loop
	for {
		// Fetch immediately
		err := c.scrapeFeeds(s, user)
		if err != nil {
			log.Printf("Error scraping feeds: %v\n", err)
		}

		// Wait for the next tick
		<-ticker.C
	}
}

func (c *Commands) AddFeed(s *state.State, cmd Command, user database.User) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("you should provide both a name and a URL")
	}

	name := cmd.Args[0]
	url := cmd.Args[1]

	currentTime := time.Now()

	createFeedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
		Url:       url,
		UserID:    user.ID,
	}

	newFeed, err := s.DB.CreateFeed(context.Background(), createFeedParams)
	if err != nil {
		return fmt.Errorf("error creating feed: %w", err)
	}

	fmt.Println("New Feed:", newFeed)

	// Create feed follow
	_, err = createFeedFollow(s, user.ID, newFeed.ID)
	if err != nil {
		return err
	}

	fmt.Println("Feed followed successfully")
	return nil
}

func (c *Commands) GetFeeds(s *state.State, cmd Command) error {
	feeds, err := s.DB.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error getting feeds: %w", err)
	}
	fmt.Println("Feeds:", feeds)
	return nil
}

func (c *Commands) Follow(s *state.State, cmd Command, user database.User) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("you should provide a URL")
	}

	url := cmd.Args[0]

	feed, err := s.DB.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("error getting the feed: %w", err)
	}

	_, err = createFeedFollow(s, user.ID, feed.ID)
	if err != nil {
		return err
	}

	fmt.Println("Successfully followed the feed")
	return nil
}

func (c *Commands) Following(s *state.State, cmd Command, user database.User) error {
	following, err := s.DB.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("error getting the follows: %w", err)
	}

	fmt.Println("Following:", following)
	return nil
}

func (c *Commands) UnFollow(s *state.State, cmd Command, user database.User) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("you should provide a url")
	}
	url := cmd.Args[0]
	feed, err := s.DB.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("error getting the feed: %w", err)
	}

	params := database.UnFollowFeedParams {
		UserID: user.ID,
		FeedID: feed.ID,
	}
	err = s.DB.UnFollowFeed(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error getting the follows: %w", err)
	}

	fmt.Println("Successfully unfollwed!")
	return nil
}


func (c *Commands) scrapeFeeds(s *state.State, user database.User) error {
	nextFeedToFetch, err := s.DB.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Printf("Error getting the next feed to fetch: %v", err)
		os.Exit(1)
	}

	lastFetchedAt := sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	params := database.MarkFeedFetchedParams {
		UserID: user.ID,
		LastFetchedAt: lastFetchedAt,
	}
	err = s.DB.MarkFeedFetched(context.Background(), params)
	if err != nil {
		fmt.Printf("Error marking feed as fetched: %v", err)
		os.Exit(1)
	}

	resp, err := feed.FetchFeed(context.Background(), nextFeedToFetch.Url)
	if err != nil {
		return fmt.Errorf("error fetching the feed: %w", err)
	}


	// Iterate over the items in the feed and save them to the database
	for _, item := range resp.Channel.Item {
		var description sql.NullString
		if item.Description == "" {
			description = sql.NullString{Valid: false} // NULL if empty string
		} else {
			description = sql.NullString{String: item.Description, Valid: true}
		}
		var pubDate sql.NullString
		if item.PubDate == "" {
			pubDate = sql.NullString{Valid: false} // NULL if empty string
		} else {
			pubDate = sql.NullString{String: item.PubDate, Valid: true}
		}
		// Prepare the post data for insertion into the database
		post := database.CreatePostParams{
			ID: uuid.New(),
			Title:       item.Title,
			Url:         item.Link,
			Description: description,
			PublishedAt: pubDate,
			FeedID:      nextFeedToFetch.ID,
		}

		// Try inserting the post into the database
		err = s.DB.CreatePost(context.Background(), post)
		if err != nil {
			// Otherwise, log the error
			fmt.Printf("Error inserting post: %v\n", err)
			return err
		}

		// Successfully inserted the post
		fmt.Printf("Post saved: %s\n", item.Title)
	}

	return nil
}

func (c *Commands) Browse(s *state.State, cmd Command) error {
	// Parse the limit parameter
	limit := 2 // Default value
	if len(cmd.Args) > 0 {
		// If a limit is provided, parse it
		providedLimit, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("invalid limit value: %v", err)
		}
		limit = providedLimit
	}

	// Fetch the posts from the database
	posts, err := s.DB.GetPosts(context.Background())
	if err != nil {
		return fmt.Errorf("error fetching posts: %w", err)
	}

	// Display the posts
	if len(posts) == 0 {
		fmt.Println("No posts found.")
		return nil
	}

	// Print the posts to the console
	fmt.Printf("Displaying the latest %d posts:\n", limit)
	for _, post := range posts {
		fmt.Printf("Title: %s\n", post.Title)
		fmt.Printf("Description: %s\n", post.Description.String)
		fmt.Printf("URL: %s\n", post.Url)
		fmt.Printf("Published at: %s\n", post.PublishedAt.String)
		fmt.Println("------------")
	}

	return nil
}