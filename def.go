package main

import (
	// "encoding/json"
	"net/http"
	"fmt"
	//"io/ioutil"
	"os"
	"strings"
	"github.com/fatih/color"
	//"bufio"
	"golang.org/x/net/html"
)

type Glosbe struct {
	Result string `json:"result"`
	Found int `json:"found"`
	Examples []Example `json:"examples"`
}

type Example struct {
	Author int `json:"author"`
	First string `json:"first"`
	Second string `json:"second"`
}


type Definition struct {
	Lang string        // User defined, lang directions
	Words []string     // All word parellels
	Conj string        // Conjugation info
	Translation string // Primary parellel
	Examples Glosbe
}

func main() {
	//var search Define
	var err error
	args := os.Args[1:] // ignore the package name as arg
	language := args[0]  // f for french to english
	phrase := args[1]    // second args is the word to search for
	//reader := bufio.NewReader(os.Stdin)

	// Colorized outputs...
	//scaffColor := color.New(color.Bold, color.FgBlue).PrintlnFunc()
	fromColor := color.New(color.Bold, color.FgMagenta).SprintFunc()
	toColor := color.New(color.Bold, color.FgRed).SprintFunc()
	defColor := color.New(color.FgGreen).SprintFunc()

	var definition Definition
	if language == "f" {
		definition.Lang = "fren"
	} else if language == "e" {
		definition.Lang = "enfr"
	}

	err = definition.WordReference(phrase)
	if err != nil {
		fmt.Printf("[WR error: %s]", err)
	}

	// Print Words
	for _, w := range definition.Words {
		fmt.Printf("%s, ", defColor(w))
	}
	fmt.Println("\n")
	fmt.Printf("\n%s:      %s \n", definition.Lang, fromColor(phrase))
	fmt.Printf("Translate: %s \n", toColor(definition.Translation))

	// TODO: Examples sentences

}

// AppendIfMissing helper method for slices.
func AppendIfMissing(slice []string, i string) []string {
    for _, ele := range slice {
	    if ele == i {
		    return slice
	    }	    
    } // else append to slice
    return append(slice, i)
}

// WordReference scrapes wordreference.com
// for the word translations and, if exists, conjugations.
func (d *Definition) WordReference(phrase string) error {
	var words []string
	var conjugation string
	url := "http://www.wordreference.com/"+d.Lang+"/"+phrase
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	z := html.NewTokenizer(resp.Body)
	// Find all ToWrd values
LoopWords:
	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			break LoopWords
		case tt == html.StartTagToken:
			t := z.Token()
			isTd := t.Data == "td"
			isDl := t.Data == "dl"
			if isTd && len(t.Attr) > 0 {
				for _, a := range t.Attr {
					if a.Val == "ToWrd"{
						inner := z.Next()
						if inner == html.TextToken {
							text := (string)(z.Text())
							text = strings.Trim(text, " ")
							if text == "English" {
								continue
							}
							words = AppendIfMissing(words, text)
						}
					}
				}
			} else if isDl {
				// Get Conjugation of French Verbs

				//_ = z.Next()
				for {
					tagName, _ := z.TagName()
					if string(tagName) == "dl" {
						break
					}
					c := strings.Trim((string)(z.Text()), " ")
					if c == ": (" || c == "conjuguer" || len(c) == 0 {
						z.Next()
						continue
					} else if c == ")" {
						c = "->"
					} else if c == "est:" {
						c += "\n"
					}
					conjugation += c +" "
					_ = z.Next()
				}
			}
		}
	}
	d.Words = words
	d.Conj = conjugation
	d.Translation = words[0]
	return nil
}


func (d *Definition) GetGlosbeJson(phrase string) error {
	var search Glosbe
	url := "https://glosbe.com/gapi/tm?from=eng&dest=fra&format=json&phrase="+phrase+"&page=1&pretty=true"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(b, &search)
	if err != nil {
		return err
	}
	if len(search.Examples) == 0 {
		return errors.New("Desolé, no examples found...")
	} else {
		d.Glosbe = search
	}
}