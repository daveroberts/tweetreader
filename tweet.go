package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/ChimeraCoder/anaconda"
	"github.com/boltdb/bolt"
	"io/ioutil"
	"log"
	"strings"
	"unicode"
)

type config struct {
	ConsumerKey    string
	ConsumerSecret string
	ApiKey         string
	ApiKeySecret   string
}

func main() {
	conf, err := ioutil.ReadFile("config.toml")
	if err != nil {
		log.Fatal(err)
	}
	var config config
	_, err = toml.Decode(string(conf), &config)
	if err != nil {
		log.Fatal(err)
	}
	anaconda.SetConsumerKey(config.ConsumerKey)
	anaconda.SetConsumerSecret(config.ConsumerSecret)
	api := anaconda.NewTwitterApi(config.ApiKey, config.ApiKeySecret)

	db, err := bolt.Open("tweets.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("tweets"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	searchResult, _ := api.GetSearch("golang", nil)
	for _, tweet := range searchResult.Statuses {
		fmt.Println(tweet.Text)
		var tags = hashtags(tweet)
		for _, tag := range tags {
			fmt.Println("Found tag: ", tag)
			db.Update(func(tx *bolt.Tx) error {
				tweets := tx.Bucket([]byte("tweets"))
				err := tweets.Put([]byte(tag), []byte(tweet.Text))
				return err
			})
		}
	}

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("tweets"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Println("key=", string(k), " value=", string(v))
		}
		return nil
	})
}

func hashtags(tweet anaconda.Tweet) []string {
	tags := []string{}
	words := strings.Fields(tweet.Text)
	for _, word := range words {
		if word[:1] == "#" {
			tags = append(tags, strings.ToLower(removeEndingPunctuation(word)))
		}
	}
	return tags
}

func removeEndingPunctuation(s string) string {
	runes := []rune(s)
	var p int = len(runes)
	var i = p - 1
	for ; i >= 0; i-- {
		if !unicode.IsPunct(runes[i]) {
			break
		}
	}
	return s[0 : i+1]
}
