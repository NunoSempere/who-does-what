package main

import (
	"log"
	"github.com/joho/godotenv"
	"os"
	"fmt"
	"encoding/json"
	openai "github.com/sashabaranov/go-openai"
	jsonschema "github.com/sashabaranov/go-openai/jsonschema"
)


type Actor struct {
	Name string `json:"name"`
	Goals string `json:"goals"`
	Powers string `json:"powers"`
}
type Actors struct {
	Actors []Actor
	Observations string `json:"observations"`
}

func GetActors(situation_description string, token string) (Actors, error){
	prompt := `Provide a list of the relevant actors and their goals as a JSON object \
	{
		actors: [
		{"name": "Name 1", "goals": "Description of goals", "powers": "Formal and informal powers"}
		{"name": "Name 2", "goals": "Description of goals 2", "powers": "Formal and informal powers"},
	  	...
		],
		observations: "any notes"
	}
	for the following situation: ` + situation_description

	var actors Actors
	schema, err := jsonschema.GenerateSchemaForType(actors)
	if err != nil {
		log.Printf("[GetActors] GenerateSchemaForType error: %v", err)
		return Actors{}, fmt.Errorf("schema generation failed: %v", err)
	}
	log.Printf("[GetActors] JSON schema generated successfully")
	openai_schema  := openai.ChatCompletionResponseFormatJSONSchema{
		Name:   "Actors",
		Schema: schema,
		Strict: true,
	}
	log.Printf("[GetActors] Making OpenAI API call with model: %s", GPT5_mini)
	openai_json, err := fetchOpenAIAnswerJSON(OpenAIRequest{prompt: prompt, model: GPT5_mini, token: token}, openai_schema)
	if err != nil {
		log.Printf("[GetActors] OpenAI API call failed: %v", err)
		return Actors{}, err
	}
	log.Printf("[GetActors] OpenAI API call successful - response length: %d chars", len(openai_json))
	
	log.Printf("[GetActors] Unmarshalling OpenAI response JSON")
	err = json.Unmarshal([]byte(openai_json), &actors)
	if err != nil {
		log.Printf("[GetActors] Error unmarshalling json: %v", err)
		log.Printf("[GetActors] Response JSON was: %v", openai_json)
		return Actors{}, err
	}
	log.Printf("[GetActors] JSON unmarshalled successfully")
	// log.Printf("[GetActors] Actors generation complete: %v", actors)
	return actors, nil
}

func main(){

	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}
  openaiToken := os.Getenv("OPENAI_API_KEY")

  situation_description := "The Bank of Japan is considering what to do about rates"
	if actors, err := GetActors(situation_description, openaiToken); err == nil {
		pretty_actors, err := json.MarshalIndent(actors, "", "  ")
		if err == nil {
			fmt.Printf("%v", pretty_actors)
		}
	}
}

/*

type SummaryBox struct {
	Summary string  `json:"summary"`
	Error   *string `json:"error"`
}

func Summarize(text string, token string) (string, error) {
	prompt := "The json API endpoint returns a {summary, error} object, like {summary: \"The article is about xyz\", error: null}. The summary contains, as a string, first a general summary of the contents of the article article in two paragraphs or less, and then an outline with the most salient, new and informative facts in an additional paragraph. The summary just states the contents of the article, and doesn't say \"The article says\" or similar introductions. For example, given the following article\n\n<INPUT>"
	prompt += text + "\n\n</INPUT>\n\nThe output is as follows (as a reminder, the json API endpoint returns a {summary, error} object, like {summary: \"The article is about xyz\", error: null}. The summary contains, as a string, first a general summary of the article in two paragraphs or less, and then an outline outlines the most salient, new and informative facts in an additional paragraph):"
	prompt += "<INPUT>" + text + "</INPUT>"

	var summary_box SummaryBox
	schema, err := jsonschema.GenerateSchemaForType(summary_box)
	if err != nil {
		log.Fatalf("GenerateSchemaForType error: %v", err)
	}
	openai_schema := openai.ChatCompletionResponseFormatJSONSchema{
		Name:   "Summary",
		Schema: schema,
		Strict: true,
	}
	summary_json, err := fetchOpenAIAnswerJSON(OpenAIRequest{prompt: prompt, model: GPT5_mini, token: token}, openai_schema)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal([]byte(summary_json), &summary_box)
	if err != nil {
		log.Printf("Error unmarshalling json: %v", err)
		log.Printf("String was: %v", summary_json)
		return "", err
	}
	if summary_box.Error != nil && (*summary_box.Error) != "" && (*summary_box.Error) != "null" {
		log.Printf("OpenAI json error field is not empty: %v", err)
		log.Printf("OpenAI answer: %v", summary_json)
		return "", nil
	}
	summary := summary_box.Summary
	return summary, nil
}


func main(){

	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}
  openaiToken := os.Getenv("OPENAI_API_KEY")

  situation_description := "The Bank of Japan is considering what to do about rates"
	if actors, err := Summarize(situation_description, openaiToken); err == nil {
		fmt.Printf("%v", actors)

	}
}
*/
