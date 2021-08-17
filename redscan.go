package main

import (
	"context"
	"fmt"
	"log"
	_ "encoding/json"
	"github.com/vartanbeno/go-reddit/v2/reddit"
	"github.com/mfdeux/pushshift/pushshift"
	"time"
)

var user string = "dirty_owl"
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
var last_reddit_comment_time int64

func main() {
	if err := scanPosts(); err != nil {
		log.Fatal(err)
	}
	if err := scanComments(); err != nil {
		log.Fatal(err)
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
			fmt.Println(node.Created)
			fmt.Println(count, total)
			//fmt.Println(node.Subreddit)
			//fmt.Println(node.Body)
			fmt.Println()
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
	client := pushshift.NewClient("redscan/0.0.1")
	q := &pushshift.CommentQuery{Author: user, Before: int(last_reddit_comment_time), Size: limit}
	comments, err := client.GetComments(q)

	if err != nil {
		return err
	}

	for _, comment := range comments {
		total++
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
		fmt.Println(total)
		//fmt.Println(node.Subreddit)
		//fmt.Println(node.Body)
		fmt.Println()
		entries[node.Created.String()] = node
	}
	return
}