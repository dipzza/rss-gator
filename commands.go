package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/dipzza/rss-gator/internal/database"
	"github.com/google/uuid"
)

func loginHandler(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("Missing argument. Usage: gator login <username>")
	}

	username := cmd.args[0]

	if _, err := s.db.GetUser(context.Background(), username); err != nil {
		return err
	}

	if err := s.cfg.SetUser(username); err != nil {
		return err
	}

	fmt.Println("User", username, "set")
	return nil
}

func registerHandler(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("Missing argument. Usage: gator register <username>")
	}
	username := cmd.args[0]

	if _, err := s.db.GetUser(context.Background(), username); err == nil {
		return fmt.Errorf("User already exists.")
	}

	newUser, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: username,
	})
	if err != nil { return err }

	if err := s.cfg.SetUser(username); err != nil { 
		return err 
	}

	fmt.Println("User", username, "created")
	fmt.Print(newUser)

	return nil
}

func resetHandler(s *state, cmd command) error {
	if err := s.db.DeleteAllUsers(context.Background()); err != nil {
		return err
	}
	fmt.Println("Reset success")
	return nil
}

func usersHandler(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Println("* " + user.Name + " (current)")
		} else {
			fmt.Println("* " + user.Name)
		}
	}

	fmt.Println("Reset success")
	return nil
}

func aggHandler(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("Missing argument. Usage: gator agg <time_between_reqs>")
	}
	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil { return err }

	fmt.Println("Collecting feeds every", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		err := scrapeFeeds(s)
		if err != nil { return err }
	}
}

func addfeedHandler(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("Missing arguments. Usage: gator addfeed <name> <url>")
	}
	name := cmd.args[0]
	url := cmd.args[1]

	ctx := context.Background()
	user, err := s.db.GetUser(ctx, s.cfg.CurrentUserName)
	if err != nil {
		return err
	}

	feed, err := s.db.CreateFeed(ctx, database.CreateFeedParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: name,
		Url: url,
		UserID: user.ID,
	})
	if err != nil {
		return err
	}

	fmt.Println(feed)

	return follow(s, feed.Url, user)
}

func feedsHandler(s *state, cmd command) error {
	ctx := context.Background()
	feeds, err := s.db.GetFeeds(ctx)
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		username, err := s.db.GetUserName(ctx, feed.UserID)
		if err != nil {
			return err
		}

		fmt.Println("Name:", feed.Name)
		fmt.Println("URL:", feed.Url)
		fmt.Println("User:", username)
	}

	return nil
}

func followHandler(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("Missing argument. Usage: gator follow <url>")
	}
	url := cmd.args[0]

	return follow(s, url, user)
}

func follow(s *state, feedUrl string, user database.User) error {
	ctx := context.Background()

	feed, err := s.db.GetFeed(ctx, feedUrl)
	if err != nil {
		return err
	}
	
	feedFollow, err := s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return err
	}

	fmt.Println("User", feedFollow.UserName, "followed feed", feedFollow.FeedName)

	return nil
}

func followingHandler(s *state, cmd command, user database.User) error {
	feedFollows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil { return err }

	fmt.Println("Following:")
	for _, feedFollow := range feedFollows {
		fmt.Println(" -", feedFollow.FeedName)
	}

	return nil
}

func browseHandler(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.args) >= 1 {
		parsedInt, err := strconv.Atoi(cmd.args[0])
		if err != nil { return err }
		limit = parsedInt
	}
	
	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		ID: user.ID,
		Limit: int32(limit),
	})
	if err != nil { return err }

	for _, post := range posts {
		fmt.Println("---", post.Title, "---", post.PublishedAt)
		fmt.Println(post.Description)
	}

	return nil;
}

func unfollowHandler(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("Missing argument. Usage: gator unfollow <url>")
	}
	url := cmd.args[0]

	ctx := context.Background()

	feed, err := s.db.GetFeed(ctx, url)
	if err != nil {
		return err
	}

	err = s.db.DeleteFeedFollow(ctx, database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return err
	}

	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, c command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return err
		}

		return handler(s, c, user)
	}
}

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil { return err }

	err = s.db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil { return err }

	rssFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil { return err }

	for _, item := range rssFeed.Channel.Item {
		pubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil { return err }

		_, err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID: uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Title: item.Title,
			Url: item.Link,
			Description: item.Description,
			PublishedAt: pubDate,
			FeedID: feed.ID,
		})
		if err.Error() == "pq: llave duplicada viola restricción de unicidad «posts_url_key»" {
			continue
		}
		if err != nil { return err }
	}
	return nil
}