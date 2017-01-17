/*
Package for scanning the corpus collections
*/
package corpus

import (
	"bufio"
	"bytes"
	"cnreader/config"
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"text/template"
	"time"
)

type CollectionEntry struct {
	CollectionFile, GlossFile, Title, Summary, Intro, DateUpdated, Corpus string
	CorpusEntries []CorpusEntry
	AnalysisFile, Format, Date, Genre string
}

const collectionsFile = "data/corpus/collections.csv"

type CorpusEntry struct {
	RawFile, GlossFile, Title, ColTitle string
}

// Index corpus entries by raw file name
var corpusEntryMap map[string]CorpusEntry

func init() {
	loadCorpusEntries()
}

// Gets the entry the collection
// Parameter
// collectionFile: The name of the file describing the collection
func GetCollectionEntry(collectionFile string) (CollectionEntry, error)  {
	log.Printf("corpus.GetCollectionEntry: collectionFile: '%s'.\n",
		collectionFile)
	collections := Collections()
	for _, entry := range collections {
		if strings.Compare(entry.CollectionFile, collectionFile) == 0 {
			return entry, nil
		}
	}
	return CollectionEntry{}, errors.New("could not find collection " +
		collectionFile)
}

// Gets the list of source and destination files for HTML conversion
func Collections() []CollectionEntry {
	collectionsFile := config.ProjectHome() + "/" + collectionsFile
	file, err := os.Open(collectionsFile)
	if err != nil {
		log.Fatal("Collections: Error opening collection file.", err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	reader.Comma = rune('\t')
	reader.Comment = rune('#')
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	collections := make([]CollectionEntry, 0)
	for i, row := range rawCSVdata {
		if len(row) < 9 {
			log.Fatal("Collections: not enough rows in file ", i, len(row),
					collectionsFile)
	  	}
		collectionFile := row[0]
		title := ""
		if row[2] != "\\N" {
			title = row[2]
		}
		summary := ""
		if row[3] != "\\N" {
			summary = row[3]
		}
		introFile := ""
		if row[4] != "\\N" {
			introFile = row[4]
		}
		corpus := ""
		if row[5] != "\\N" {
			corpus = row[5]
		}
		format := ""
		if row[6] != "\\N" {
			format = row[6]
		}
		date := ""
		if row[7] != "\\N" {
			date = row[7]
		}
		genre := ""
		if len(row) > 8 && row[8] != "\\N" {
			genre = row[8]
		}
		corpusEntries := make([]CorpusEntry, 0)
		//log.Printf("corpus.Collections: Read collection %s in corpus %s\n",
		//	collectionFile, corpus)
		collections = append(collections, CollectionEntry{collectionFile,
			row[1], title, summary, introFile, "", corpus, corpusEntries, "",
			format, date, genre})
	}
	return collections
}

// Get a list of files for a corpus
func CorpusEntries(collectionFile, colTitle string) []CorpusEntry {
	file, err := os.Open(collectionFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	reader.Comma = rune('\t')
	reader.Comment = rune('#')
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	corpusEntries := make([]CorpusEntry, 0)
	for _, row := range rawCSVdata {
		if len(row) != 3 {
			log.Fatal("corpus.CorpusEntries len(row) != 3 ", row)
		}
		corpusEntries = append(corpusEntries, CorpusEntry{row[0], row[1],
			row[2], colTitle})
	}
	return corpusEntries	
}

// Lookup a corpus entry by raw file name
func GetCorpusEntry(filename string) CorpusEntry {
	return corpusEntryMap[filename]
}

// Load all corpus entries and keep them in a hash map
func loadCorpusEntries() {
	corpusEntryMap = make(map[string]CorpusEntry)
	collections := Collections()
	for _, collectionEntry := range collections {
		corpusEntries := CorpusEntries(config.CorpusDataDir() + "/" +
		collectionEntry.CollectionFile, collectionEntry.Title)
		for _, entry := range corpusEntries {
			corpusEntryMap[entry.RawFile] = entry
		}
	}
}

// Constructor for an empty CollectionEntry
func NewCorpusEntry() *CorpusEntry {
	return &CorpusEntry{
		RawFile: "",
		GlossFile: "",
		Title: "",
	}
}

// Reads a text file introducing the collection. The file should be a plain
// text file. HTML breaks will be added for line breaks.
// Parameter
// introFile: The name of the file introducing the collection
func ReadIntroFile(introFile string) string {
	//log.Printf("ReadIntroFile: Reading introduction file.\n")
	infile, err := os.Open(config.ProjectHome() + "/corpus/" + introFile)
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(infile)
	var buffer bytes.Buffer
	eof := false
	for !eof {
		var line string
		line, err = reader.ReadString('\n')
		if err == io.EOF {
			err = nil
			eof = true
		} else if err != nil {
			break
		}
		if _, err = buffer.WriteString(line + "<br/>\n"); err != nil {
			break
		}
	}
	return buffer.String()
}

// Writes a HTML file describing the collection
// Parameter
// collectionFile: The name of the file describing the collection
func WriteCollectionFile(collectionFile, analysisFile string) {
	//log.Printf("WriteCollectionFile: Writing collection file.\n")
	collections := Collections()
	for _, entry := range collections {
		if entry.CollectionFile == collectionFile && entry.GlossFile != "\\N" {
			outputFile := config.ProjectHome() + "/data/corpus/" +collectionFile
			entry.CorpusEntries = CorpusEntries(outputFile, entry.Title)
			//log.Printf("WriteCollectionFile: Writing collection file %s\n",
			//	outputFile)
			entry.AnalysisFile = analysisFile

			// Write to file
			f, err := os.Create(config.ProjectHome() + "/web/" +
				entry.GlossFile)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			w := bufio.NewWriter(f)
			// Replace name of intro file with introduction text
			entry.Intro = ReadIntroFile(entry.Intro)
			entry.DateUpdated = time.Now().Format("2006-01-02")
			templFile := config.TemplateDir() + "/collection-template.html"
			//log.Println("Home: ", config.ProjectHome())
			tmpl:= template.Must(template.New(
					"collection-template.html").ParseFiles(templFile))
			err = tmpl.Execute(w, entry)
			if err != nil {
				log.Fatal(err)
			}
			w.Flush()
		}
	}
}
