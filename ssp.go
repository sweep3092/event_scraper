package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/go.text/encoding/japanese"
	"code.google.com/p/go.text/transform"
	"github.com/PuerkitoBio/goquery"
	"github.com/deckarep/golang-set"
)

func eucjp_to_utf8(str string) (string, error) {
	ret, err := ioutil.ReadAll(transform.NewReader(strings.NewReader(str), japanese.EUCJP.NewDecoder()))
	if err != nil {
		return "", err
	}
	return string(ret), err
}

func utf8_to_sjis(str string) string {
	iostr := strings.NewReader(str)
	rio := transform.NewReader(iostr, japanese.ShiftJIS.NewEncoder())
	ret, err := ioutil.ReadAll(rio)
	if err != nil {
		return ""
	}
	return string(ret)
}

func GetUrls(url string) (links []string) {
	doc, _ := goquery.NewDocument(url)
	doc.Find("div.btn_detail a").Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		// fmt.Println(url)
		links = append(links, href)
	})

	return links
}

func GetEventInfo(url string) (title string, date string, area string) {
	doc, _ := goquery.NewDocument(url)

	title, _ = eucjp_to_utf8(doc.Find("p.title").Text())
	arr := []string{title}
	// fmt.Println("Event:", title)
	doc.Find("dl.detail_box dd").Each(func(_ int, s *goquery.Selection) {
		text, _ := eucjp_to_utf8(s.Text())
		arr = append(arr, text)
		//fmt.Println(text)
	})

	return arr[0], arr[1], arr[4]
}

func parseDateAnd3DaysBefore(rawdate string) (string, string) {
	// 7.10(金)
	unbracket := strings.Split(rawdate, "(")
	monthday := strings.Split(unbracket[0], ".")
	month, _ := strconv.Atoi(monthday[0])
	day, _ := strconv.Atoi(monthday[1])
	now := time.Now()
	startdate := time.Date(now.Year(), time.Month(month), day, 0, 0, 0, 0, time.UTC)
	enddate := startdate.AddDate(0, 0, -3)

	return startdate.Format("2006-01-02"), enddate.Format("2006-01-02")
}

func main() {
	fmt.Println("--- Event Scraper for mynavi konkatsu ---")
	fmt.Println("(c) 2015 youk all rights reserved.\n")

	url := "http://event.konkatsu.mynavi.jp"
	urls := GetUrls(url)

	urls_set := mapset.NewSet()
	for _, v := range urls {
		urls_set.Add(v)
	}

	fmt.Printf("%d events are detected.\n", urls_set.Cardinality())
	os.Remove("result.csv")
	write_file, _ := os.OpenFile("result.csv", os.O_WRONLY|os.O_CREATE, 0600)
	writer := bufio.NewWriter(write_file)
	writer.Write([]byte(utf8_to_sjis("開催日,イベント名,エリア,掲載終了予定日\n")))

	urls_unique_slice := urls_set.ToSlice()
	for i, v := range urls_unique_slice {
		title, date, area := GetEventInfo(url + v.(string))
		startdate, closedate := parseDateAnd3DaysBefore(date)

		line := fmt.Sprintf("%s,%s,%s,%s\n", utf8_to_sjis(startdate), utf8_to_sjis(title), utf8_to_sjis(area), utf8_to_sjis(closedate))
		writer.Write([]byte(line))
		fmt.Println(i+1, "events are scraped.")
	}
	writer.Flush()

	fmt.Println("done.")
}
