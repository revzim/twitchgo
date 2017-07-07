package main

import (
	"log"
	"net/http"
	// "io/ioutil"
	"time"
	"html/template"
	"encoding/json"
	"regexp"
	// "reflect"
	// "yt"
	// "google.golang.org/api/youtube/v3"
)

//template for page
var templates = template.Must(template.ParseFiles("twitchindex.html"))

/*
validPath - MustCompile will parse and compile the regexp
and return a regexp.
Regexep.MustCompile is distinct form Compile in that it will panic if the exp compiliation fails
*/
var validPath = regexp.MustCompile("^/(twitch|search|youtube)/([a-zA-Z0-9]+)") ///([a-zA-Z0-9]+)$

var CLIENT_ID = ""

type Twitch struct {
	Stream
	Channel
}

type Stream struct {
	ID          json.Number `json:"_id,number"`
	Game        string      `json:"game"`
	Viewers     uint        `json:"viewers"`
	VideoHeight uint        `json:"video_height"`
	AverageFps  float64     `json:"average_fps"`
	Delay       uint        `json:"delay"`
	CreatedAt   time.Time   `json:"created_at"`
	IsPlaylist  bool        `json:"is_playlist"`
	Preview     Preview     `json:"preview"`
	Channel     Channel     `json:"channel"`
}


type Preview struct {
	Large    string `json:"large"`
	Medium   string `json:"medium"`
	Small    string `json:"small"`
	Template string `json:"template"`
}

type StreamResponse struct {
	Stream Stream `json:"stream"`
}

// Channel Twitch Data
type Channel struct {
	Mature                       bool        `json:"mature"`
	Status                       string      `json:"status"`
	BroadcasterLanguage          string      `json:"broadcaster_language"`
	DisplayName                  string      `json:"display_name"`
	Game                         string      `json:"game"`
	Language                     string      `json:"language"`
	ID                           json.Number `json:"_id,number"`
	Name                         string      `json:"name"`
	CreatedAt                    time.Time   `json:"created_at"`
	UpdatedAt                    time.Time   `json:"updated_at"`
	Partner                      bool        `json:"partner"`
	Logo                         string      `json:"logo"`
	VideoBanner                  string      `json:"video_banner"`
	ProfileBanner                string      `json:"profile_banner"`
	ProfileBannerBackgroundColor string      `json:"profile_banner_background_color"`
	URL                          string      `json:"url"`
	Views                        uint        `json:"views"`
	Followers                    uint        `json:"followers"`
}

/*
load method
*/
func load(streamer string) (*Stream, error){
	s := "https://api.twitch.tv/kraken/streams/" + streamer + "?client_id=" + CLIENT_ID
	resp, err := http.Get(s)
	if err != nil {
		log.Println("Error: failed to fetch request.", err)
	}
	defer resp.Body.Close()
	// body, err := ioutil.ReadAll(resp.Body)
	var stream = &StreamResponse{}
	err = json.NewDecoder(resp.Body).Decode(&stream)
	if err != nil {
		log.Fatal(err)
	}
	return &stream.Stream, nil
}

/*
search method
*/
func searchHandler(w http.ResponseWriter, r *http.Request, streamer string) {
	s := r.FormValue("streamer-search")
	t, err := load(s)
	if t.Channel.URL == "" {
		http.Redirect(w, r, "/twitch/"+s, http.StatusFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/twitch/"+s, http.StatusFound)
}


/*
twitchHandler method --

*/
func twitchHandler(w http.ResponseWriter, r *http.Request, streamer string){
	t, err := load(streamer)
	if t.Channel.URL == "" {
		log.Printf("%s is not live.", streamer)
		renderTemplate(w, "twitchindex", &Stream{})
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	renderTemplate(w, "twitchindex", t)
}


/*
general renderTemplate func
http.Eerror sends specified internalservice error response code and err msg
*/
func renderTemplate(w http.ResponseWriter, tmpl string, s *Stream){
    err := templates.ExecuteTemplate(w, tmpl+".html", s)
    if err != nil{
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

/*
makeHandler method --
wrapper function takes handler functions and returns a function of type http.HandlerFunc
fn is enclosed by closure, fn will be one of the pages available
closure returned by makeHandler is a function that takes http.ResponseWriter and http.Request
then extracts title from request path, validates with TitleValidator regexp.
If title is invalid, error will be written, ResponseWriter, using http.NotFound
If title is valid, enclosed handler function fn will be called with the ResponseWriter, Request and title as args
*/
func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc{
    return func(w http.ResponseWriter, r *http.Request){
        //extract page title from Request
        //call provided handler 'fn'
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil{
            // http.NotFound(w, r)
            fn(w, r, "")
            return
        }
        fn(w, r, m[2])
    }
}
/*
func run(){
	client, err := yt.BuildOAuthHTTPClient(youtube.YoutubeForceSslScope)
	yt.HandleError(err, "Error building OAuth client")
	service, err := youtube.New(client)
	yt.HandleError(err, "Error creating YouTube client")
	yt.SearchListByKeyword(service, "snippet", 25, "league of legends", "")
}
*/

func main(){
	//run()
	http.HandleFunc("/twitch/", makeHandler(twitchHandler))
	http.HandleFunc("/search/", makeHandler(searchHandler))
	http.ListenAndServe(":8000", nil)

    
}
