package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"errors"
	"fmt"
	"example.com/gator/internal/database"
	"time"
	"github.com/google/uuid"
	"context"
	"log"
	"strconv"
)

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUsername string `json:"current_user_name"`
}

func Read() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}
	path := filepath.Join(home, ".gatorconfig.json")
	bytes, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	err = json.Unmarshal(bytes, &cfg)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (cfg *Config) SetUser(username string) error {
	cfg.CurrentUsername = username
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".gatorconfig.json")
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

type state struct{
	config *Config
	db *database.Queries
}
type command struct{
	name string
	arguments []string
}
func handlerLogin(s *state,cmd command) error{
	if len(cmd.arguments)==0{
		return errors.New("where the arg(put da cookie down)s")
	}
	_,err:=s.db.GetUser(context.Background(),cmd.arguments[0])
	if err!=nil{
		return err
	}
	err=s.config.SetUser(cmd.arguments[0])
	if err!=nil{
		return err
	}
	fmt.Println("user hath been set cuh")
	return nil
}
func handlerRegister(s *state,cmd command) error{
	if len(cmd.arguments)!=1{
		return errors.New("stupid ass disapperaing line(too many or too few args)")
	}
	sus:=cmd.arguments[0]
	_,err:=s.db.CreateUser(context.Background(),database.CreateUserParams{
		ID:uuid.New(),
		CreatedAt:time.Now(),
		UpdatedAt:time.Now(),
		Name:sus,})
	if err!=nil{
		return err
	}
	err=s.config.SetUser(sus)
	if err!=nil{
		return err
	}
	fmt.Printf("%v has been set as a user\n",sus)
	return nil
}
func DeyeetAll(s *state,cmd command) error{
	err:=s.db.DeleteAllUsers(context.Background())
	if err!=nil{
		return err
	}
	return nil
}
func GetAllUsers(s *state,cmd command) error{
	da_ta,err:=s.db.GetUsers(context.Background())
	if err!=nil{
		return err
	}
	for _,i:=range da_ta{
		if i.Name==s.config.CurrentUsername{
			fmt.Printf("* %v (current)\n",i.Name)
		}else{
			fmt.Printf("* %v\n",i.Name)
		}
	}
	return nil
}
func handlerAggregate(s *state,cmd command) error{
	time_between_reqs:=cmd.arguments[0]
	tyme,err:=time.ParseDuration(time_between_reqs)
	if err!=nil{
		return err
	}
	ticker := time.NewTicker(tyme)
	fmt.Println("Collecting feeds every ",tyme)
	for ; ; <-ticker.C {
		err:=scrapeFeeds(s)
		if err!=nil{
			log.Println(err)
		}
	}
	return nil
}
func handlerAddfeed(s *state,cmd command,user database.User) error{
	if len(cmd.arguments)!=2{
		return errors.New("my args are cold.")
	}
	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
	ID:        uuid.New(),
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
	Name:      cmd.arguments[0],
	Url:       cmd.arguments[1],
	UserID:    user.ID,
	})
	if err != nil {
		return err
	}
        _,err= s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
        ID:        uuid.New(),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        UserID:    user.ID,
        FeedID:    feed.ID,
        })
        if err!=nil{
                return err
        }
	fmt.Printf("%+v\n", feed)
	return nil
}
func handlerFEEDME(s *state,cmd command) error{
	feeds,err:=s.db.ReadAllFeeds(context.Background())
	if err!=nil{
		return err
	}
	// coffeecoffeecoffee
	for _,i:=range feeds{
		fmt.Println(i.FeedName,i.Url,i.Name)
	}
	return nil
}
func handlerFollowFeed(s *state,cmd command,user database.User) error{
	feed,err:=s.db.FeedbyUrl(context.Background(),cmd.arguments[0])
	if err!=nil{
		return err
	}
	follow,err:= s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
        ID:        uuid.New(),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        UserID:    user.ID,
	FeedID:    feed.ID,
        })
	if err!=nil{
		return err
	}
	fmt.Println(follow.FeedName,follow.UserName)
	return nil
}
func handlerFollowing(s *state,cmd command,user database.User) error{
	following,err:=s.db.GetFeedFollows(context.Background(),user.ID)
	if err!=nil{
		return err
	}
	for _,i:=range following{
		fmt.Println(i.FeedName)
	}
	return nil
}
func handlerDelFeedFollow(s *state,cmd command,user database.User) error{
	feed,err:=s.db.FeedbyUrl(context.Background(),cmd.arguments[0])
        if err!=nil{
                return err
        }
	err=s.db.DeleteFeedFollow(context.Background(),database.DeleteFeedFollowParams{
	UserID: user.ID,
	FeedID: feed.ID,
	})
	if err!=nil{
		return err
	}
	return nil
}
func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.arguments) == 1 {
		if specifiedLimit, err := strconv.Atoi(cmd.arguments[0]); err == nil {
			limit = specifiedLimit
		} else {
			return fmt.Errorf("invalid limit: %w", err)
		}
	}
	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err!=nil{
		return err
	}
	for _,i:=range posts{
		fmt.Println(i)
	}
	return nil
}
type commands struct{
	name_to_function map[string]func(*state,command) error
}
func (c *commands) run (s *state, cmd command) error{
	if yippee,ok:=c.name_to_function[cmd.name];ok{
		return yippee(s,cmd)
	}
	return errors.New("doesn't exist pluh")
}
func (c *commands) register (name string,f func(*state,command) error){
	c.name_to_function[name]=f
}
