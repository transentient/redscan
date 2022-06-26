package main

import (
	"context"
	"fmt"
	"log"
	_ "encoding/json"
	"github.com/vartanbeno/go-reddit/v2/reddit"
	"github.com/mfdeux/pushshift/pushshift"
	"time"
	"flag"
)

var user string
var last_reddit_comment_time int64
var ctx = context.Background()
var limit = 100
var total = 0

type entry struct {
	Id 			string	`json:id`
	Created 	time.Time 	`json:created`
	Author 		string	`json:author`
	Subreddit 	string	`json:subreddit`
	Score		int 	`json:score`
	Link  		string	`json:string`
	Body  		string	`json:body`
} 

var entries = make(map[string]entry)

func main() {
	flag.StringVar(&user, "user", "dirty_owl", "reddit username")
	flag.Int64Var(&last_reddit_comment_time, "time", 0, "date to start scan")
	flag.Parse()
	fmt.Printf("Scanning for posts and comments for user %s, for time %d\n", user, last_reddit_comment_time)

	if (last_reddit_comment_time == 0) {
		if err := scanPosts(); err != nil {
			log.Fatal(err)
		}
		if err := scanComments(); err != nil {
			log.Fatal(err)
		}
	}	

	if err := scanPushShift(); err != nil {
		log.Fatal(err)
	}

	//entries_j, err := json.MarshalIndent(entries, "", "  ")
	//if err != nil {
	//	fmt.Printf("Error: %s", err.Error())
	//} else {
	//	fmt.Println(string(entries_j))
	//}
}

func scanPosts() (err error) {
	fmt.Println("Posts")
	afterval := ""
	for {
		posts, _, err := reddit.DefaultClient().User.PostsOf(ctx, user, &reddit.ListUserOverviewOptions{
			ListOptions: reddit.ListOptions{
				Limit: limit,
				After: afterval,
			},
		})

		if err != nil {
			return err
		}

		count := 0
		for _, post := range posts {
			count++
			afterval = post.FullID
			node := entry {
				Id: post.FullID,
				Created: post.Created.Time,
				Author: post.Author,
				Subreddit: post.SubredditName,
				Score: post.Score,
				Link: post.URL,
				Body: post.Body,
			}
			//fmt.Println(node.Created)
			//fmt.Println(node.Subreddit)
			//fmt.Println(node.Body)
			//fmt.Println()
			entries[node.Created.String()] = node
		}

		if count < limit {
			break
		}
	}
	return
}

func scanComments() (err error) {
	fmt.Println("Comments")
	afterval := ""

	for {
		comments, _, err := reddit.DefaultClient().User.CommentsOf(ctx, user, &reddit.ListUserOverviewOptions{
			ListOptions: reddit.ListOptions{
				Limit: limit,
				After: afterval,
			},
		})

		if err != nil {
			return err
		}

		count := 0
		for _, comment := range comments {
			count++
			total++
			last_reddit_comment_time = comment.Created.Time.Unix()
			afterval = comment.FullID
			node := entry {
				Id: comment.FullID,
				Created: comment.Created.Time,
				Author: comment.Author,
				Subreddit: comment.SubredditName,
				Score: comment.Score,
				Link: comment.Permalink,
				Body: comment.Body,
			}
			//fmt.Println(node.Created)
			//fmt.Println(count, total)
			//fmt.Println(node.Subreddit)
			//fmt.Println(node.Body)
			//fmt.Println()
			entries[node.Created.String()] = node
		}

		if count < limit {
			break
		}
	}
	return
}

func scanPushShift() (err error) {
	fmt.Println("Pushshift Comments")
	ps_size := 25
	client := pushshift.NewClient("redscan/0.0.1")
	
	for {
		q := &pushshift.CommentQuery{Author: user, Before: int(last_reddit_comment_time), Size: ps_size}
		comments, err := client.GetComments(q)

		if err != nil {
			//return err
			fmt.Println(err)
		}

		count := 0
		for _, comment := range comments {
			count++
			total++
			last_reddit_comment_time = int64(comment.CreatedUtc)
			node := entry {
				Id: comment.ID,
				Created: time.Unix(int64(comment.CreatedUtc), 0),
				Author: comment.Author,
				Subreddit: comment.Subreddit,
				Score: comment.Score,
				Link: comment.Permalink,
				Body: comment.Body,
			}
			fmt.Println(node.Created)
			fmt.Println(last_reddit_comment_time)
			fmt.Println(total)
			fmt.Println(node.Subreddit)
			fmt.Println(node.Body)
			fmt.Println()
			entries[node.Created.String()] = node
		}

		if count < ps_size {
			break
		}
		//time.Sleep(1 * time.Second)
	}
	return
}