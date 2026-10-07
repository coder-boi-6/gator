package main

import _ "github.com/lib/pq"

import (
	"log"
	"os"
	"example.com/gator/internal/database"
	"database/sql"
	"context"
)

func main() {
	data, err := Read()
	if err != nil {
		log.Fatal(err)
	}
	db,err:=sql.Open("postgres",data.DBURL)
	if err!=nil{
		log.Fatal(err)
	}
	dbQueries:=database.New(db)
	sate:=state{
		config:&data,
		db:dbQueries,
	}
	cmds:=commands{
		name_to_function: make(map[string]func(*state,command) error),
	}
	cmds.register("login",handlerLogin)
	cmds.register("register",handlerRegister)
	cmds.register("reset", DeyeetAll)
	cmds.register("users",GetAllUsers)
	cmds.register("agg",handlerAggregate)
	cmds.register("addfeed",middlewareLoggedIn(handlerAddfeed))
	cmds.register("feeds",handlerFEEDME)
	cmds.register("follow",middlewareLoggedIn(handlerFollowFeed))
	cmds.register("following",middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow",middlewareLoggedIn(handlerDelFeedFollow))
	cmds.register("browse",middlewareLoggedIn(handlerBrowse))
	if len(os.Args)<2{
		log.Fatal("i NEED more COOKIES man, ARG")
	}
	sus:=os.Args[2:]
	cmd:=command{
		name:os.Args[1],
		arguments:sus,
	}
	err=cmds.run(&sate, cmd)
	if err!=nil{
		log.Fatal("this person doesn't exist.")
	}
}
func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error{
	return func(s *state,cmd command) error{
		user, err := s.db.GetUser(context.Background(), s.config.CurrentUsername)
		if err != nil {
			return err
		}
		return handler(s,cmd,user)
	}
}
