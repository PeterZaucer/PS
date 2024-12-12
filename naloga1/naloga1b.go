package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"flag"
	"unicode"

	"github.com/laspp/PS-2024/vaje/naloga-1/koda/xkcd"
)

var wg sync.WaitGroup
var lock sync.Mutex

func ocisti(text string) string {
	var builder strings.Builder
	for _, c := range text{
		if unicode.IsLetter(c) || unicode.IsSpace(c){
			builder.WriteRune(unicode.ToLower(c))
		}else{
			builder.WriteRune(' ')
		}
	}
	return builder.String()

}

func stetje(goI int, goNum int, wordFreq *map[string]int) {
	defer wg.Done()

	wordFreqTemp := make(map[string]int)

	for i := goI; true; i+=goNum {
		c, _ := xkcd.FetchComic(i)

		if c.Id == 0 && c.Title == "" {
			i+=goNum
			c, _ := xkcd.FetchComic(i)
			if c.Id == 0 && c.Title == "" {
				break
			}

		}
		texts := []string{c.Title}

		if c.Transcript != "" {
			texts = append(texts, c.Transcript)
		}else{
			texts = append(texts, c.Tooltip)
		}

		for _, text := range texts {
			words := strings.Fields(ocisti(text))
			for _, word := range words{
				if len(word) >= 4 {
					wordFreqTemp[word]++
				}
			}
		}
	}

	lock.Lock()

	for w, c := range wordFreqTemp {
		(*wordFreq)[w] += c
	}

	lock.Unlock()
}


func main() {
	gPtr := flag.Int("g", 2, "# of goroutines")
	flag.Parse()

	wordFreq := make(map[string]int)

	wg.Add(*gPtr)

	for i := 0; i < *gPtr; i++ {
		go stetje(i, *gPtr, &wordFreq)
	}

	wg.Wait()

	type wordFreqSort struct {
		Word  string
		Count int
	}
	var freqList []wordFreqSort
	for word, count := range wordFreq {
		freqList = append(freqList, wordFreqSort{Word: word, Count: count})
	}

	sort.Slice(freqList, func(i, j int) bool {
		if freqList[i].Count == freqList[j].Count {
			return freqList[i].Word < freqList[j].Word
		}
		return freqList[i].Count > freqList[j].Count
	})

	for i := 0; i < 15 && i < len(freqList); i++ {
		fmt.Printf("%s: %d\n", freqList[i].Word, freqList[i].Count)
	}
}
