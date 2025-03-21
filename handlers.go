package main

import (
	"fmt"
	"internal/database"
	"errors"
	"database/sql"
	"context"
	"github.com/google/uuid"
	"time"
	"strings"
	"strconv"
)

// command: log in
func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return errors.New("login requires a username")
	}
	_, err := s.database.GetUser(context.Background(), sql.NullString{String: cmd.arguments[0], Valid: true,})
	if err != nil {
		return errors.New("username does not exist")
	}
	err = s.config.SetUser(cmd.arguments[0])
	if err != nil {
		return err
	}
	fmt.Println("user successfully set")
	return nil
}

// command: register a user
func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return errors.New("register requires a username")
	}
	_, err := s.database.CreateUser(
		context.Background(), 
		database.CreateUserParams{
			ID: uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name: sql.NullString{String: cmd.arguments[0], Valid: true,},
		})
	if err != nil {
		return errors.New("username already exists")
	}
	err = s.config.SetUser(cmd.arguments[0])
	if err != nil {
		return err
	}
	fmt.Println("user successfully registered and set")
	//fmt.Println(fmt.Sprintf("%v %v %v %v", user.ID, user.CreatedAt, user.UpdatedAt, user.Name))
	return nil
}

// command: reset the database (maybe deregister this one later for "production")
func handlerReset(s *state, cmd command) error {
	err := s.database.DeleteUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error deleting the users: %v", err)
	}
	fmt.Println("users deleted successfully")
	return nil
}

// command: list all users
func handlerUsers(s *state, cmd command) error {
	users, err := s.database.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error getting all users: %v", err)
	}
	current_user := s.config.CurrentUserName
	for _, u := range users {
		n := u.String
		if n == current_user {
			fmt.Println("* " + n + " (current)")
		} else {
			fmt.Println("* " + n)
		}
	}
	return nil
}

// command: aggregator
func handlerAgg(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return errors.New("agg requires a time between requests")
	}
	timeBetweenRequests, err := time.ParseDuration(cmd.arguments[0])
	if err != nil {
		return err
	}
	onesec, _ := time.ParseDuration("1s")
	if timeBetweenRequests < onesec {
		fmt.Println("We're not looking to DOS anyone here >_<  You get every second, okay?")
		timeBetweenRequests = onesec
	}
	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		err = scrapeFeeds(s)
		if err != nil {
			return err
		}
	}
	return nil
}

// helper for the aggregator - prints the titles of the next fetched feed
func scrapeFeeds(s *state) error {
	fetch, err := s.database.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	err = s.database.MarkFeedFetched(
		context.Background(),
		database.MarkFeedFetchedParams{
			ID: fetch.ID,
			UpdatedAt: time.Now(),
		})
	if err != nil {
		return err
	}
	feed, err := fetchFeed(context.Background(), fetch.Url.String)
	if err != nil {
		return err
	}
	for _, item := range feed.Channel.Item {
		format := "Mon, 02 Jan 2006 15:04:05 -0700"
		published, err := time.Parse(format, item.PubDate)
		if err != nil {
			fmt.Println("Publishing time format is weird: " + item.PubDate)
		}
		post, err := s.database.CreatePost(
			context.Background(),
			database.CreatePostParams{
				ID: uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Title: item.Title,
				Url: item.Link,
				Description: sql.NullString{String: item.Description, Valid: true,},
				PublishedAt: published,
				FeedID: fetch.ID,
			})
		if err != nil {
			if !strings.Contains(fmt.Sprint(err), "posts_url_key") {
				fmt.Println(err)
			}
		} else {
			fmt.Println(fmt.Sprintf("%v: %v", post.Title, post.Url))
		}
	}
	return nil
}

// command: add a feed
func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 2 {
		return errors.New("addfeed requires a name and a url")
	}
	feed, err := s.database.CreateFeed(
		context.Background(),
		database.CreateFeedParams{
			ID: uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name: sql.NullString{String: cmd.arguments[0], Valid: true,},
			Url: sql.NullString{String: cmd.arguments[1], Valid: true,},
			UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
		})
	if err != nil {
		return err
	}
	fmt.Println("Feed successfully registered")
	//fmt.Println(fmt.Sprintf("%v %v %v %v %v %v", feed.ID, feed.CreatedAt, feed.UpdatedAt, feed.Name, feed.Url, feed.UserID))
	// and then we register it as followed too
	feedFollow, err := s.database.CreateFeedFollow(
		context.Background(),
		database.CreateFeedFollowParams {
			ID: uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
			FeedID: uuid.NullUUID{UUID: feed.ID, Valid: true},
		})
	if err != nil {
		return err
	}
	fmt.Println(fmt.Sprintf("%v is now following %v", feedFollow.UserName.String, feedFollow.FeedName.String))
	return nil
}

// command: list all feeds
func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.database.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error getting all users: %v", err)
	}
	for _, f := range feeds {
		fmt.Println(fmt.Sprintf("%v: %v (%v)", f.Creator.String, f.Name.String, f.Url.String))
	}
	return nil
}

// command: follow a feed based on its url
func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) == 0 {
		return errors.New("follow requires a url")
	}
	targetFeed, err := s.database.GetFeedByUrl(context.Background(), sql.NullString{String: cmd.arguments[0], Valid: true,})
	if err != nil {
		return err
	}
	feedFollow, err := s.database.CreateFeedFollow(
		context.Background(),
		database.CreateFeedFollowParams {
			ID: uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
			FeedID: uuid.NullUUID{UUID: targetFeed.ID, Valid: true},
		})
	if err != nil {
		return err
	}
	fmt.Println(fmt.Sprintf("%v is now following %v", feedFollow.UserName.String, feedFollow.FeedName.String))
	return nil;
}

// command: get all feeds followed by the current user
func handlerFollowing(s *state, cmd command, user database.User) error {
	followed, err := s.database.GetFeedFollowsForUser(context.Background(), user.Name)
	if err != nil {
		return err
	}
	for _, f := range followed {
		fmt.Println(f.FeedName.String)
	}
	return nil
}

// command: unfollow feed (by url) for the current user
func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) == 0 {
		return errors.New("unfollow requires a url")
	}
	targetFeed, err := s.database.GetFeedByUrl(context.Background(), sql.NullString{String: cmd.arguments[0], Valid: true,})
	if err != nil {
		return err
	}
	err = s.database.RemoveFeedFollow(
		context.Background(), 
		database.RemoveFeedFollowParams {
			UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
			FeedID: uuid.NullUUID{UUID: targetFeed.ID, Valid: true},
		})
	if err != nil {
		return err
	}
	fmt.Println("unfollowed successfully")
	return nil
}

// command: browse the user's feed's posts (takes a limit parameter, defaults to 2)
func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.arguments) > 0 {
		l, err := strconv.Atoi(cmd.arguments[0])
		if err != nil {
			fmt.Println("limit should be an integer; defaulting to 2")
		} else {
			limit = l
		}
	}
	posts, err := s.database.GetPostsForUser(
		context.Background(),
		database.GetPostsForUserParams {
			UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
			Limit: int32(limit),
		})
	if err != nil {
		return err
	}
	for _, p := range posts {
		fmt.Println(p.Title)
		fmt.Println(p.Url)
		fmt.Println(fmt.Sprintf("Published at: %v", p.PublishedAt))
		fmt.Println(p.Description.String)
		fmt.Println("*")
	}
	return nil
}

// register handlers
func registerHandlers(c commands) {
	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
	//c.register("reset", handlerReset)
	c.register("users", handlerUsers)
	c.register("agg", handlerAgg)
	c.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	c.register("feeds", handlerFeeds)
	c.register("follow", middlewareLoggedIn(handlerFollow))
	c.register("following", middlewareLoggedIn(handlerFollowing))
	c.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	c.register("browse", middlewareLoggedIn(handlerBrowse))
}