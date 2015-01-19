package twitter

import (
	"github.com/ChimeraCoder/anaconda"
	"bufio"
	"os"
	"fmt"
)

var api *anaconda.TwitterApi

func init () {
	f, err := os.Open("CREDENTIALS")
	if err != nil {
		panic(err)
	}
	s := bufio.NewScanner(f)
	s.Scan()
	anaconda.SetConsumerKey(s.Text())
	s.Scan()
	anaconda.SetConsumerSecret(s.Text())
	s.Scan()
	tkn := s.Text()
	s.Scan()
	tknSkrt := s.Text()
	api = anaconda.NewTwitterApi(tkn, tknSkrt)
}

func Tweet (s string) error {
	_, err := api.PostTweet(fmt.Sprintf("%s on @radioxenu http://tunein.com/radio/Radio-Xenu-s118981/ \n", s), nil)

	return err
}
